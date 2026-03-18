package telemetry

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// commandNameRe matches <command-name>/skill-name</command-name> tags in user messages
var commandNameRe = regexp.MustCompile(`<command-name>/([^<]+)</command-name>`)

// TranscriptRecord represents a single line in the transcript JSONL file
type TranscriptRecord struct {
	Type           string          `json:"type"`
	Subtype        string          `json:"subtype,omitempty"`
	SessionID      string          `json:"sessionId,omitempty"`
	Version        string          `json:"version,omitempty"`
	CWD            string          `json:"cwd,omitempty"`
	GitBranch      string          `json:"gitBranch,omitempty"`
	PermissionMode string          `json:"permissionMode,omitempty"`
	Timestamp      string          `json:"timestamp,omitempty"`
	DurationMs     int64           `json:"durationMs,omitempty"`
	Message        *MessageContent `json:"message,omitempty"`
	Summary        string          `json:"summary,omitempty"`
}

// MessageContent represents the message field in assistant/user records
type MessageContent struct {
	Role    string          `json:"role,omitempty"`
	Model   string          `json:"model,omitempty"`
	Content json.RawMessage `json:"content,omitempty"` // Can be string (user) or []ContentItem (assistant)
	Usage   *UsageStats     `json:"usage,omitempty"`
}

// ContentItem represents items in the content array
type ContentItem struct {
	Type  string                 `json:"type"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
}

// UsageStats represents token usage statistics
type UsageStats struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// SessionStats contains aggregated statistics from a transcript
type SessionStats struct {
	SessionID       string         `json:"session_id"`
	Version         string         `json:"version"`
	CWD             string         `json:"cwd"`
	GitBranch       string         `json:"git_branch"`
	PermissionMode  string         `json:"permission_mode"`
	Summary         string         `json:"summary,omitempty"`
	Model           string         `json:"model,omitempty"`
	StartTime       time.Time      `json:"start_time"`
	EndTime         time.Time      `json:"end_time"`
	DurationMs      int64          `json:"duration_ms"`
	UserMessages    int            `json:"user_messages"`
	AssistMessages  int            `json:"assistant_messages"`
	TokenUsage      TokenUsage     `json:"token_usage"`
	ToolsInvoked    map[string]int `json:"tools_invoked"`
	SkillsInvoked   map[string]int `json:"skills_invoked"`
	AgentsSpawned   map[string]int `json:"agents_spawned"`
	MCPToolsInvoked map[string]int `json:"mcp_tools_invoked"`
}

// TokenUsage contains aggregated token counts
type TokenUsage struct {
	Input         int `json:"input"`
	Output        int `json:"output"`
	CacheCreation int `json:"cache_creation"`
	CacheRead     int `json:"cache_read"`
}

// ParseTranscript reads a transcript JSONL file and extracts session statistics
func ParseTranscript(path string) (*SessionStats, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open transcript: %w", err)
	}
	defer file.Close()

	stats := &SessionStats{
		ToolsInvoked:    make(map[string]int),
		SkillsInvoked:   make(map[string]int),
		AgentsSpawned:   make(map[string]int),
		MCPToolsInvoked: make(map[string]int),
	}

	var firstTimestamp, lastTimestamp time.Time
	scanner := bufio.NewScanner(file)

	// Increase buffer size for large lines
	buf := make([]byte, 0, 1024*1024) // 1MB buffer
	scanner.Buffer(buf, 10*1024*1024) // 10MB max line size

	for scanner.Scan() {
		var record TranscriptRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			// Skip malformed lines
			continue
		}

		// Parse timestamp
		if record.Timestamp != "" {
			if ts, err := time.Parse(time.RFC3339, record.Timestamp); err == nil {
				if firstTimestamp.IsZero() || ts.Before(firstTimestamp) {
					firstTimestamp = ts
				}
				if ts.After(lastTimestamp) {
					lastTimestamp = ts
				}
			}
		}

		switch record.Type {
		case "user":
			stats.UserMessages++
			// Capture session metadata from first user message
			if stats.SessionID == "" {
				stats.SessionID = record.SessionID
				stats.Version = record.Version
				stats.CWD = record.CWD
				stats.GitBranch = record.GitBranch
				stats.PermissionMode = record.PermissionMode
			}

			// Check for slash command skill invocations via <command-name> tags
			if record.Message != nil && len(record.Message.Content) > 0 {
				var textToSearch []string
				// Try as string first (most common for slash command messages)
				var contentStr string
				if err := json.Unmarshal(record.Message.Content, &contentStr); err == nil {
					textToSearch = append(textToSearch, contentStr)
				} else {
					// Try as array of content items with {"type":"text","text":"..."}
					var items []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					}
					if err := json.Unmarshal(record.Message.Content, &items); err == nil {
						for _, item := range items {
							if item.Type == "text" && item.Text != "" {
								textToSearch = append(textToSearch, item.Text)
							}
						}
					}
				}
				for _, text := range textToSearch {
					for _, match := range commandNameRe.FindAllStringSubmatch(text, -1) {
						if len(match) > 1 {
							stats.SkillsInvoked[match[1]]++
						}
					}
				}
			}

		case "assistant":
			stats.AssistMessages++
			if record.Message != nil {
				// Capture model
				if stats.Model == "" && record.Message.Model != "" {
					stats.Model = record.Message.Model
				}

				// Aggregate token usage
				if record.Message.Usage != nil {
					stats.TokenUsage.Input += record.Message.Usage.InputTokens
					stats.TokenUsage.Output += record.Message.Usage.OutputTokens
					stats.TokenUsage.CacheCreation += record.Message.Usage.CacheCreationInputTokens
					stats.TokenUsage.CacheRead += record.Message.Usage.CacheReadInputTokens
				}

				// Count tool usage - parse content as array of ContentItem
				var contentItems []ContentItem
				if len(record.Message.Content) > 0 {
					// Try to parse as array of content items
					if err := json.Unmarshal(record.Message.Content, &contentItems); err == nil {
						for _, content := range contentItems {
							if content.Type == "tool_use" && content.Name != "" {
								switch {
								case content.Name == "Skill":
									// Extract skill name from input
									if skillName, ok := content.Input["skill"].(string); ok {
										stats.SkillsInvoked[skillName]++
									}
									// Also count Skill as a tool
									stats.ToolsInvoked["Skill"]++

								case content.Name == "Task":
									// Extract subagent type from input
									if agentType, ok := content.Input["subagent_type"].(string); ok {
										stats.AgentsSpawned[agentType]++
									}
									// Also count Task as a tool
									stats.ToolsInvoked["Task"]++

								case strings.HasPrefix(content.Name, "mcp__"):
									// MCP tool - count by full name
									stats.MCPToolsInvoked[content.Name]++
									// Also count in general tools
									stats.ToolsInvoked[content.Name]++

								default:
									stats.ToolsInvoked[content.Name]++
								}
							}
						}
					}
				}
			}

		case "system":
			if record.Subtype == "turn_duration" {
				stats.DurationMs += record.DurationMs
			}

		case "summary":
			stats.Summary = record.Summary
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading transcript: %w", err)
	}

	stats.StartTime = firstTimestamp
	stats.EndTime = lastTimestamp

	// If no turn duration recorded, calculate from timestamps
	if stats.DurationMs == 0 && !firstTimestamp.IsZero() && !lastTimestamp.IsZero() {
		stats.DurationMs = lastTimestamp.Sub(firstTimestamp).Milliseconds()
	}

	return stats, nil
}

// ToEvent converts SessionStats to a telemetry Event
func (s *SessionStats) ToEvent(eventType string) *Event {
	payload := map[string]interface{}{
		"session_id":         s.SessionID,
		"version":            s.Version,
		"cwd":                s.CWD,
		"git_branch":         s.GitBranch,
		"permission_mode":    s.PermissionMode,
		"model":              s.Model,
		"duration_ms":        s.DurationMs,
		"user_messages":      s.UserMessages,
		"assistant_messages": s.AssistMessages,
		"token_usage":        s.TokenUsage,
		"tools_invoked":      s.ToolsInvoked,
		"skills_invoked":     s.SkillsInvoked,
		"agents_spawned":     s.AgentsSpawned,
		"mcp_tools_invoked":  s.MCPToolsInvoked,
	}

	if s.Summary != "" {
		payload["summary"] = s.Summary
	}

	// Serialize payload to JSON string for zerobus compatibility
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		// Fallback to empty object if serialization fails
		payloadJSON = []byte("{}")
	}

	return &Event{
		EventType: eventType,
		Timestamp: s.EndTime,
		Payload:   string(payloadJSON),
	}
}
