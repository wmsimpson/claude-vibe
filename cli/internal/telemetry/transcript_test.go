package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseTranscript(t *testing.T) {
	// Create a temporary transcript file
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "test.jsonl")

	transcriptContent := `{"type":"file-history-snapshot","messageId":"e3a8aeb7","timestamp":"2026-01-28T20:02:45.681Z"}
{"type":"user","sessionId":"test-session-123","version":"2.1.25","cwd":"/Users/test/project","gitBranch":"main","permissionMode":"default","timestamp":"2026-01-28T20:02:45.623Z","message":{"role":"user","content":"hello"}}
{"type":"assistant","sessionId":"test-session-123","version":"2.1.25","timestamp":"2026-01-28T20:02:48.546Z","message":{"model":"claude-opus-4-5-20251101","role":"assistant","content":[{"type":"text","text":"Hello!"}],"usage":{"input_tokens":100,"output_tokens":50,"cache_creation_input_tokens":1000,"cache_read_input_tokens":500}}}
{"type":"assistant","sessionId":"test-session-123","timestamp":"2026-01-28T20:02:50.000Z","message":{"model":"claude-opus-4-5-20251101","content":[{"type":"tool_use","name":"Bash","input":{"command":"ls"}}],"usage":{"input_tokens":50,"output_tokens":25,"cache_creation_input_tokens":0,"cache_read_input_tokens":1000}}}
{"type":"assistant","sessionId":"test-session-123","timestamp":"2026-01-28T20:02:52.000Z","message":{"model":"claude-opus-4-5-20251101","content":[{"type":"tool_use","name":"Skill","input":{"skill":"databricks-tools:databricks-query"}}],"usage":{"input_tokens":30,"output_tokens":15,"cache_creation_input_tokens":0,"cache_read_input_tokens":500}}}
{"type":"assistant","sessionId":"test-session-123","timestamp":"2026-01-28T20:02:54.000Z","message":{"model":"claude-opus-4-5-20251101","content":[{"type":"tool_use","name":"Task","input":{"subagent_type":"fe-internal-tools:field-data-analyst","prompt":"query data"}}],"usage":{"input_tokens":20,"output_tokens":10,"cache_creation_input_tokens":0,"cache_read_input_tokens":200}}}
{"type":"assistant","sessionId":"test-session-123","timestamp":"2026-01-28T20:02:56.000Z","message":{"model":"claude-opus-4-5-20251101","content":[{"type":"tool_use","name":"Bash","input":{"command":"pwd"}}],"usage":{"input_tokens":10,"output_tokens":5,"cache_creation_input_tokens":0,"cache_read_input_tokens":100}}}
{"type":"assistant","sessionId":"test-session-123","timestamp":"2026-01-28T20:02:57.000Z","message":{"model":"claude-opus-4-5-20251101","content":[{"type":"tool_use","name":"mcp__slack__slack_read_api_call","input":{"endpoint":"conversations.list"}}],"usage":{"input_tokens":5,"output_tokens":3,"cache_creation_input_tokens":0,"cache_read_input_tokens":50}}}
{"type":"assistant","sessionId":"test-session-123","timestamp":"2026-01-28T20:02:57.500Z","message":{"model":"claude-opus-4-5-20251101","content":[{"type":"tool_use","name":"mcp__slack__slack_read_api_call","input":{"endpoint":"conversations.history"}}],"usage":{"input_tokens":5,"output_tokens":3,"cache_creation_input_tokens":0,"cache_read_input_tokens":50}}}
{"type":"system","subtype":"turn_duration","durationMs":15000,"timestamp":"2026-01-28T20:02:58.000Z"}
{"type":"summary","summary":"Test conversation about data queries"}
`
	if err := os.WriteFile(transcriptPath, []byte(transcriptContent), 0644); err != nil {
		t.Fatalf("Failed to write test transcript: %v", err)
	}

	// Parse the transcript
	stats, err := ParseTranscript(transcriptPath)
	if err != nil {
		t.Fatalf("ParseTranscript failed: %v", err)
	}

	// Verify session metadata
	if stats.SessionID != "test-session-123" {
		t.Errorf("Expected session ID 'test-session-123', got '%s'", stats.SessionID)
	}
	if stats.Version != "2.1.25" {
		t.Errorf("Expected version '2.1.25', got '%s'", stats.Version)
	}
	if stats.CWD != "/Users/test/project" {
		t.Errorf("Expected cwd '/Users/test/project', got '%s'", stats.CWD)
	}
	if stats.GitBranch != "main" {
		t.Errorf("Expected git branch 'main', got '%s'", stats.GitBranch)
	}
	if stats.Model != "claude-opus-4-5-20251101" {
		t.Errorf("Expected model 'claude-opus-4-5-20251101', got '%s'", stats.Model)
	}
	if stats.Summary != "Test conversation about data queries" {
		t.Errorf("Expected summary 'Test conversation about data queries', got '%s'", stats.Summary)
	}

	// Verify message counts
	if stats.UserMessages != 1 {
		t.Errorf("Expected 1 user message, got %d", stats.UserMessages)
	}
	if stats.AssistMessages != 7 {
		t.Errorf("Expected 7 assistant messages, got %d", stats.AssistMessages)
	}

	// Verify token usage
	expectedInput := 100 + 50 + 30 + 20 + 10 + 5 + 5
	if stats.TokenUsage.Input != expectedInput {
		t.Errorf("Expected input tokens %d, got %d", expectedInput, stats.TokenUsage.Input)
	}
	expectedOutput := 50 + 25 + 15 + 10 + 5 + 3 + 3
	if stats.TokenUsage.Output != expectedOutput {
		t.Errorf("Expected output tokens %d, got %d", expectedOutput, stats.TokenUsage.Output)
	}

	// Verify tool usage
	if stats.ToolsInvoked["Bash"] != 2 {
		t.Errorf("Expected 2 Bash invocations, got %d", stats.ToolsInvoked["Bash"])
	}
	if stats.ToolsInvoked["Skill"] != 1 {
		t.Errorf("Expected 1 Skill invocation, got %d", stats.ToolsInvoked["Skill"])
	}
	if stats.ToolsInvoked["Task"] != 1 {
		t.Errorf("Expected 1 Task invocation, got %d", stats.ToolsInvoked["Task"])
	}

	// Verify skills invoked
	if stats.SkillsInvoked["databricks-tools:databricks-query"] != 1 {
		t.Errorf("Expected 1 databricks-query skill invocation, got %d", stats.SkillsInvoked["databricks-tools:databricks-query"])
	}

	// Verify agents spawned
	if stats.AgentsSpawned["fe-internal-tools:field-data-analyst"] != 1 {
		t.Errorf("Expected 1 field-data-analyst agent spawn, got %d", stats.AgentsSpawned["fe-internal-tools:field-data-analyst"])
	}

	// Verify MCP tools invoked
	if stats.MCPToolsInvoked["mcp__slack__slack_read_api_call"] != 2 {
		t.Errorf("Expected 2 mcp__slack__slack_read_api_call invocations, got %d", stats.MCPToolsInvoked["mcp__slack__slack_read_api_call"])
	}

	// Verify duration
	if stats.DurationMs != 15000 {
		t.Errorf("Expected duration 15000ms, got %d", stats.DurationMs)
	}
}

func TestParseTranscriptSlashCommandSkills(t *testing.T) {
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "test.jsonl")

	transcriptContent := `{"type":"user","sessionId":"slash-cmd-session","version":"2.1.25","cwd":"/Users/test","gitBranch":"main","permissionMode":"default","timestamp":"2026-01-28T20:00:00.000Z","message":{"role":"user","content":"<command-message>databricks-tools:databricks-lakeview-dashboard</command-message>\n<command-name>/databricks-tools:databricks-lakeview-dashboard</command-name>\n<command-args>create a dashboard</command-args>"}}
{"type":"user","sessionId":"slash-cmd-session","timestamp":"2026-01-28T20:00:01.000Z","message":{"role":"user","content":[{"type":"text","text":"Base directory for this skill: /tmp\n\n<command-name>/google-tools:google-docs</command-name>\n\n# Google Docs Skill"}]}}
{"type":"assistant","sessionId":"slash-cmd-session","timestamp":"2026-01-28T20:00:02.000Z","message":{"model":"claude-opus-4-5-20251101","content":[{"type":"tool_use","name":"Skill","input":{"skill":"fe-internal-tools:logfood-querier"}}],"usage":{"input_tokens":100,"output_tokens":50,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}
{"type":"assistant","sessionId":"slash-cmd-session","timestamp":"2026-01-28T20:00:03.000Z","message":{"model":"claude-opus-4-5-20251101","content":[{"type":"text","text":"Done!"}],"usage":{"input_tokens":10,"output_tokens":5,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}
`
	if err := os.WriteFile(transcriptPath, []byte(transcriptContent), 0644); err != nil {
		t.Fatalf("Failed to write test transcript: %v", err)
	}

	stats, err := ParseTranscript(transcriptPath)
	if err != nil {
		t.Fatalf("ParseTranscript failed: %v", err)
	}

	// Slash command skill from string content
	if stats.SkillsInvoked["databricks-tools:databricks-lakeview-dashboard"] != 1 {
		t.Errorf("Expected 1 databricks-lakeview-dashboard slash command skill, got %d", stats.SkillsInvoked["databricks-tools:databricks-lakeview-dashboard"])
	}

	// Slash command skill from array content
	if stats.SkillsInvoked["google-tools:google-docs"] != 1 {
		t.Errorf("Expected 1 google-docs slash command skill, got %d", stats.SkillsInvoked["google-tools:google-docs"])
	}

	// Skill tool invocation (assistant-initiated)
	if stats.SkillsInvoked["fe-internal-tools:logfood-querier"] != 1 {
		t.Errorf("Expected 1 logfood-querier Skill tool invocation, got %d", stats.SkillsInvoked["fe-internal-tools:logfood-querier"])
	}

	// Total unique skills should be 3
	if len(stats.SkillsInvoked) != 3 {
		t.Errorf("Expected 3 unique skills, got %d: %v", len(stats.SkillsInvoked), stats.SkillsInvoked)
	}

	// Verify user message count
	if stats.UserMessages != 2 {
		t.Errorf("Expected 2 user messages, got %d", stats.UserMessages)
	}
}

func TestSessionStatsToEvent(t *testing.T) {
	stats := &SessionStats{
		SessionID:      "test-123",
		Version:        "2.1.25",
		CWD:            "/test",
		GitBranch:      "main",
		Model:          "claude-opus-4-5-20251101",
		DurationMs:     5000,
		UserMessages:   2,
		AssistMessages: 3,
		TokenUsage: TokenUsage{
			Input:         100,
			Output:        50,
			CacheCreation: 200,
			CacheRead:     300,
		},
		ToolsInvoked:  map[string]int{"Bash": 2, "Read": 1},
		SkillsInvoked: map[string]int{"my-skill": 1},
		AgentsSpawned: map[string]int{"my-agent": 2},
	}

	event := stats.ToEvent("claude.session.stop")

	if event.EventType != "claude.session.stop" {
		t.Errorf("Expected event type 'claude.session.stop', got '%s'", event.EventType)
	}

	// Payload is now a JSON string, parse it to verify contents
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
		t.Fatalf("Failed to parse payload JSON: %v", err)
	}

	if payload["session_id"] != "test-123" {
		t.Errorf("Expected session_id 'test-123', got '%v'", payload["session_id"])
	}

	toolsInvoked, ok := payload["tools_invoked"].(map[string]interface{})
	if !ok {
		t.Errorf("Expected tools_invoked to be map[string]interface{}")
	} else if toolsInvoked["Bash"].(float64) != 2 {
		t.Errorf("Expected Bash count 2, got %v", toolsInvoked["Bash"])
	}
}
