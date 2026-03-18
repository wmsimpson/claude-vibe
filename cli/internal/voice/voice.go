// Package voice provides microphone speech-to-text capture using macOS native dictation.
// It spawns a Swift helper binary (vibe-dictation) that uses SFSpeechRecognizer for
// on-device real-time transcription, and streams results back as JSON events.
package voice

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

// TranscriptEvent represents a JSON message from the dictation helper.
type TranscriptEvent struct {
	Type string `json:"type"` // "ready", "partial", "final", "error"
	Text string `json:"text"`
}

// Listener manages the dictation helper subprocess and parses its JSON output.
type Listener struct {
	cmd     *exec.Cmd
	scanner *bufio.Scanner
}

// NewListener spawns the vibe-dictation helper with the given silence timeout
// and returns a Listener that streams transcript events.
func NewListener(silenceTimeout float64) (*Listener, error) {
	helperPath, err := FindHelper()
	if err != nil {
		return nil, err
	}

	args := []string{
		"--timeout", fmt.Sprintf("%.1f", silenceTimeout),
	}
	cmd := exec.Command(helperPath, args...)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting dictation helper: %w", err)
	}

	return &Listener{
		cmd:     cmd,
		scanner: bufio.NewScanner(stdout),
	}, nil
}

// Next blocks until the next transcript event is available from the helper.
func (l *Listener) Next() (*TranscriptEvent, error) {
	if !l.scanner.Scan() {
		if err := l.scanner.Err(); err != nil {
			return nil, fmt.Errorf("reading from dictation helper: %w", err)
		}
		// EOF - helper exited
		return nil, fmt.Errorf("dictation helper exited")
	}

	var evt TranscriptEvent
	if err := json.Unmarshal(l.scanner.Bytes(), &evt); err != nil {
		return nil, fmt.Errorf("parsing dictation event: %w", err)
	}
	return &evt, nil
}

// Stop sends SIGTERM to the helper and waits for it to exit.
func (l *Listener) Stop() error {
	if l.cmd.Process != nil {
		_ = l.cmd.Process.Signal(syscall.SIGTERM)
	}
	return l.cmd.Wait()
}

// FindHelper locates the vibe-dictation binary by checking:
// 1. The same directory as the running vibe binary
// 2. The system PATH
func FindHelper() (string, error) {
	if runtime.GOOS != "darwin" {
		return "", fmt.Errorf("voice dictation requires macOS (current OS: %s)", runtime.GOOS)
	}

	const name = "vibe-dictation"

	// Check next to the vibe binary
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	// Check PATH
	if p, err := exec.LookPath(name); err == nil {
		return p, nil
	}

	return "", fmt.Errorf("%s not found; build it with 'make build-dictation'", name)
}
