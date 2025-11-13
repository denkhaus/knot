package internal

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/denkhaus/knot/internal/app"
	"github.com/stretchr/testify/require"
)

// CLITestHelper provides utilities for testing CLI commands
type CLITestHelper struct {
	tempDir     string
	originalDir string
	app         *app.App
	t           *testing.T
}

// NewCLITestHelper creates a new CLI test helper with isolated environment
func NewCLITestHelper(t *testing.T) *CLITestHelper {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "knot_cli_test_*")
	require.NoError(t, err)

	// Save original directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	// Change to temp directory
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Create .knot directory
	knotDir := filepath.Join(tempDir, ".knot")
	err = os.MkdirAll(knotDir, 0755)
	require.NoError(t, err)

	// Create application
	application, err := app.New()
	require.NoError(t, err)

	return &CLITestHelper{
		tempDir:     tempDir,
		originalDir: originalDir,
		app:         application,
		t:           t,
	}
}

// RunCommand executes a CLI command and returns output and error
func (h *CLITestHelper) RunCommand(args ...string) (string, string, error) {
	// Prepare full args with program name
	fullArgs := append([]string{"knot"}, args...)

	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	// Run the command
	err := h.app.Run(fullArgs)

	// Close writers and restore
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read captured output
	var stdout, stderr bytes.Buffer
	_, _ = io.Copy(&stdout, rOut)
	_, _ = io.Copy(&stderr, rErr)

	return stdout.String(), stderr.String(), err
}

// RunCommandExpectSuccess runs a command and expects it to succeed
func (h *CLITestHelper) RunCommandExpectSuccess(args ...string) (string, string) {
	stdout, stderr, err := h.RunCommand(args...)
	require.NoError(h.t, err, "Command should succeed: %v", args)
	return stdout, stderr
}

// RunCommandExpectError runs a command and expects it to fail
func (h *CLITestHelper) RunCommandExpectError(args ...string) (string, string, error) {
	stdout, stderr, err := h.RunCommand(args...)
	require.Error(h.t, err, "Command should fail: %v", args)
	return stdout, stderr, err
}

// ExtractProjectID extracts project ID from create project output
func (h *CLITestHelper) ExtractProjectID(output string) string {
	// Look for pattern like "ID: uuid"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "ID:") {
			parts := strings.Split(line, "ID:")
			if len(parts) > 1 {
				id := strings.TrimSpace(strings.Split(parts[1], ")")[0])
				return strings.Trim(id, "()")
			}
		}
	}
	return ""
}

// ExtractTaskID extracts task ID from create task output
func (h *CLITestHelper) ExtractTaskID(output string) string {
	return h.ExtractProjectID(output) // Same pattern
}

// Cleanup cleans up the test environment
func (h *CLITestHelper) Cleanup() {
	_ = os.Chdir(h.originalDir)
	_ = os.RemoveAll(h.tempDir)
}

// CreateTestProject creates a test project and returns its ID
func (h *CLITestHelper) CreateTestProject(title, description string) string {
	stdout, _ := h.RunCommandExpectSuccess("project", "create", "--title", title, "--description", description)
	projectID := h.ExtractProjectID(stdout)
	require.NotEmpty(h.t, projectID, "Should extract project ID from output")
	return projectID
}

// CreateTestTask creates a test task and returns its ID
func (h *CLITestHelper) CreateTestTask(projectID, title, description string, complexity int) string {
	// First select the project
	h.RunCommandExpectSuccess("project", "select", "--id", projectID)

	// Then create the task
	stdout, _ := h.RunCommandExpectSuccess("task", "create",
		"--title", title,
		"--description", description,
		"--complexity", strconv.Itoa(complexity),
		"--priority", "medium")
	taskID := h.ExtractTaskID(stdout)
	require.NotEmpty(h.t, taskID, "Should extract task ID from output")
	return taskID
}
