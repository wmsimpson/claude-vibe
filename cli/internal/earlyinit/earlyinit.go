// Package earlyinit provides early initialization that runs before other packages.
// This must be imported FIRST in main.go to suppress spurious SDK init warnings.
package earlyinit

import (
	"os"
	"strings"
	"sync"
	"time"
)

var (
	realStderr *os.File
	pipeWriter *os.File
	pipeReader *os.File
	once       sync.Once
	buffer     strings.Builder
	bufferMu   sync.Mutex
)

func init() {
	once.Do(func() {
		// Save the real stderr
		realStderr = os.Stderr

		// Create a pipe to capture stderr
		r, w, err := os.Pipe()
		if err != nil {
			return
		}
		pipeReader = r
		pipeWriter = w

		// Replace stderr with our pipe
		os.Stderr = w

		// Start a goroutine to read from the pipe and buffer it
		go func() {
			buf := make([]byte, 4096)
			for {
				n, err := pipeReader.Read(buf)
				if n > 0 {
					bufferMu.Lock()
					buffer.Write(buf[:n])
					bufferMu.Unlock()
				}
				if err != nil {
					break
				}
			}
		}()
	})
}

// FlushFiltered flushes buffered stderr, filtering out spurious SDK messages.
// Call this at the start of main() after all init() functions have run.
func FlushFiltered() {
	if realStderr == nil {
		return
	}

	// Close the pipe writer to signal EOF to the reader
	if pipeWriter != nil {
		pipeWriter.Close()
	}

	// Give the goroutine time to finish reading all buffered data
	time.Sleep(10 * time.Millisecond)

	// Restore real stderr
	os.Stderr = realStderr

	// Get the buffered content
	bufferMu.Lock()
	content := buffer.String()
	bufferMu.Unlock()

	// Check if this contains any zerobus-related messages
	hasZerobusMessage := strings.Contains(content, "ERROR: Rust FFI library not found") ||
		strings.Contains(content, "The Zerobus Go SDK requires a one-time build step") ||
		strings.Contains(content, "zerobus-sdk-go")

	if hasZerobusMessage {
		// Find and remove the entire error block(s)
		lines := strings.Split(content, "\n")
		var filtered []string
		inBorderedBlock := false
		inTextBlock := false
		blockEndCount := 0

		for _, line := range lines {
			// Handle bordered error blocks (═══════════════════════════════════════════)
			if strings.Contains(line, "═══════════════════════════════════════════") {
				if !inBorderedBlock {
					inBorderedBlock = true
					blockEndCount = 0
				} else {
					blockEndCount++
					if blockEndCount >= 1 {
						inBorderedBlock = false
						continue
					}
				}
				continue
			}

			// Handle plain text "go generate" instruction block
			if strings.Contains(line, "The Zerobus Go SDK requires a one-time build step") {
				inTextBlock = true
				continue
			}
			if inTextBlock {
				// End the text block after "go build' will work normally."
				if strings.Contains(line, "will work normally") {
					inTextBlock = false
					continue
				}
				continue
			}

			if !inBorderedBlock && !inTextBlock {
				filtered = append(filtered, line)
			}
		}

		content = strings.Join(filtered, "\n")
		// Remove any leading/trailing newlines that might be left
		content = strings.TrimPrefix(content, "\n")
		content = strings.TrimSuffix(content, "\n")
	}

	// Write any remaining content to real stderr
	if content != "" {
		realStderr.WriteString(content)
	}
}
