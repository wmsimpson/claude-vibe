package install

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SSHHostKeysStep pre-populates ~/.ssh/known_hosts with GitHub's official SSH
// host keys. This prevents an interactive fingerprint verification prompt when
// SSH connects to github.com for the first time (e.g., during gh repo clone).
//
// The prompt reads from /dev/tty, so it cannot be auto-accepted via stdin
// piping, and blocks indefinitely when run in a non-interactive context.
//
// Keys are pinned from GitHub's published fingerprints:
// https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/githubs-ssh-key-fingerprints
type SSHHostKeysStep struct{}

func (s *SSHHostKeysStep) ID() string          { return "ssh_host_keys" }
func (s *SSHHostKeysStep) Name() string        { return "SSH Host Keys" }
func (s *SSHHostKeysStep) Description() string { return "Pre-populate GitHub SSH host keys" }
func (s *SSHHostKeysStep) ActiveForm() string  { return "Configuring SSH host keys" }
func (s *SSHHostKeysStep) CanSkip(opts *Options) bool { return false }
func (s *SSHHostKeysStep) NeedsSudo() bool     { return false }

// GitHub's official SSH host keys, pinned from:
// https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/githubs-ssh-key-fingerprints
var githubHostKeys = []struct {
	keyType string // e.g. "ssh-ed25519"
	entry   string // full known_hosts line
}{
	{
		keyType: "ssh-ed25519",
		entry:   "github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl",
	},
	{
		keyType: "ecdsa-sha2-nistp256",
		entry:   "github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=",
	},
	{
		keyType: "ssh-rsa",
		entry:   "github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=",
	},
}

func (s *SSHHostKeysStep) Check(ctx *Context) (bool, error) {
	knownHostsPath := filepath.Join(ctx.HomeDir, ".ssh", "known_hosts")

	content, err := os.ReadFile(knownHostsPath)
	if err != nil {
		return false, nil
	}

	// Check that all three key types are present
	for _, key := range githubHostKeys {
		if !strings.Contains(string(content), "github.com "+key.keyType) {
			return false, nil
		}
	}

	return true, nil
}

func (s *SSHHostKeysStep) Run(ctx *Context) StepResult {
	sshDir := filepath.Join(ctx.HomeDir, ".ssh")
	knownHostsPath := filepath.Join(sshDir, "known_hosts")

	// Ensure ~/.ssh exists with correct permissions
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return Failure("Failed to create ~/.ssh directory", err)
	}

	// Read existing known_hosts content
	existingContent := ""
	if data, err := os.ReadFile(knownHostsPath); err == nil {
		existingContent = string(data)
	}

	// Determine which keys need to be added
	var keysToAdd []string
	for _, key := range githubHostKeys {
		if !strings.Contains(existingContent, "github.com "+key.keyType) {
			keysToAdd = append(keysToAdd, key.entry)
		}
	}

	if len(keysToAdd) == 0 {
		return Success("GitHub SSH host keys already present")
	}

	// Open known_hosts for appending (create if needed)
	f, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return Failure("Failed to open known_hosts", err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	for _, entry := range keysToAdd {
		writer.WriteString(entry + "\n")
	}
	if err := writer.Flush(); err != nil {
		return Failure("Failed to write to known_hosts", err)
	}

	// Verify the keys were written by re-reading the file
	verifyContent, err := os.ReadFile(knownHostsPath)
	if err != nil {
		return Failure("Failed to verify known_hosts", err)
	}
	for _, key := range githubHostKeys {
		if !strings.Contains(string(verifyContent), "github.com "+key.keyType) {
			return Failure(
				fmt.Sprintf("Verification failed: %s key not found after write", key.keyType),
				nil,
			)
		}
	}

	return SuccessWithDetails(
		"GitHub SSH host keys added",
		fmt.Sprintf("  Added %d key(s) to %s", len(keysToAdd), knownHostsPath),
	)
}
