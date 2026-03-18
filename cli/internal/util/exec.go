package util

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCommand executes a command and returns the stdout, stderr, and any error.
// Stdout and stderr are returned as trimmed strings.
func RunCommand(name string, args ...string) (stdout, stderr string, err error) {
	cmd := exec.Command(name, args...)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()

	stdout = strings.TrimSpace(outBuf.String())
	stderr = strings.TrimSpace(errBuf.String())

	return stdout, stderr, err
}

// RunCommandInDir executes a command in the specified directory and returns
// the stdout, stderr, and any error. Returns an error if the directory
// doesn't exist.
func RunCommandInDir(dir, name string, args ...string) (stdout, stderr string, err error) {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return "", "", fmt.Errorf("directory does not exist: %s", dir)
	}
	if !info.IsDir() {
		return "", "", fmt.Errorf("not a directory: %s", dir)
	}

	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()

	stdout = strings.TrimSpace(outBuf.String())
	stderr = strings.TrimSpace(errBuf.String())

	return stdout, stderr, err
}

// CommandExists checks if a command is available in the system PATH.
func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
