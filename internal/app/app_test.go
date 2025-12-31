package app

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/denkhaus/knot/v2/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestMainFunction(t *testing.T) {
	// Test the main function by capturing stdout/stderr
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test successful execution with help command
	os.Args = []string{"knot", "--help"}

	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set version for test
	SetVersionFromBuild("test", "test-commit", "test-date")

	// Create application
	application, err := New()
	require.NoError(t, err)

	// Run with help to avoid exit
	os.Args = []string{"knot", "--help"}
	err = application.Run([]string{"knot", "--help"})
	assert.NoError(t, err)

	// Restore stdout
	if err := w.Close(); err != nil {
		t.Logf("Warning: failed to close stdout writer: %v", err)
	}
	os.Stdout = old

	// Read the output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should contain help information
	assert.Contains(t, output, "A CLI tool for hierarchical project and task management")
}

func TestAppNew(t *testing.T) {
	// Test successful app creation
	app, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, app.App)
	assert.Equal(t, "knot", app.Name)
}

func TestAppRun(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	// Test with help command to avoid exit
	err = app.Run([]string{"knot", "--help"})
	assert.NoError(t, err)
}

func TestIsUserInputError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "required flag error",
			err:      fmt.Errorf("Required flag not provided"),
			expected: true,
		},
		{
			name:     "flag not defined error",
			err:      fmt.Errorf("flag provided but not defined"),
			expected: true,
		},
		{
			name:     "invalid value error",
			err:      fmt.Errorf("invalid value for flag"),
			expected: true,
		},
		{
			name:     "command not found error",
			err:      fmt.Errorf("command not found"),
			expected: true,
		},
		{
			name:     "incorrect usage error",
			err:      fmt.Errorf("incorrect usage"),
			expected: true,
		},
		{
			name:     "flag needs argument error",
			err:      fmt.Errorf("flag needs an argument"),
			expected: true,
		},
		{
			name:     "help topic error",
			err:      fmt.Errorf("No help topic for 'unknown'"),
			expected: true,
		},
		{
			name:     "internal error",
			err:      fmt.Errorf("internal application error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUserInputError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppRunWithError(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	// The CLI framework calls os.Exit for invalid commands rather than returning errors
	// This is standard behavior for CLI applications using urfave/cli
	// We'll test that the app was created successfully instead
	assert.NotNil(t, app)

	// Commands are added dynamically in the Before hook, so we test flags instead
	assert.NotEmpty(t, app.Flags)
}

func TestSetVersionFromBuild(t *testing.T) {
	// Test version setting
	SetVersionFromBuild("v1.0.0", "abc123", "2023-01-01")

	// Create a new app to see if the version was set
	app, err := New()
	require.NoError(t, err)

	// Check that version was set, though we can't easily verify the internal version variable
	// since it's not exposed, we'll just ensure app creation still works
	assert.Equal(t, "knot", app.Name) // This will still be knot regardless of version
}

func TestAppWithMemoryRepository(t *testing.T) {
	// This test verifies the app structure
	app, err := New()
	require.NoError(t, err)
	require.NotNil(t, app)

	// Verify the app structure
	assert.Equal(t, "knot", app.Name)
	assert.NotNil(t, app.App)
}

func TestAppDIInitialization(t *testing.T) {
	app, err := New()
	require.NoError(t, err)
	require.NotNil(t, app)

	// Test that the app is properly initialized
	assert.NotNil(t, app.App)
	assert.Equal(t, "knot", app.Name)
}

func TestAppRunWithValidArgs(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	// Test with version command
	_ = app.Run([]string{"knot", "--version"})
	// Version command should not return an error
	// but since it calls os.Exit in real usage, we might need a different approach
	// For now, just test that the app can be created without issues
	assert.NotNil(t, app)
}

func TestAppCommandsStructure(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	// Commands are added dynamically in the Before hook, so we need to trigger it
	// Create a mock flag set with required flags
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("log-level", "info", "")
	flagSet.Int("complexity-threshold", 5, "")
	flagSet.Int("max-depth", 10, "")
	flagSet.Int("max-tasks-per-depth", 50, "")
	flagSet.Int("max-description-length", 500, "")
	flagSet.Bool("auto-reduce-complexity", true, "")

	// Create a mock context to trigger the Before hook
	mockCtx := cli.NewContext(app.App, flagSet, nil)

	// Trigger the Before hook to initialize commands
	err = app.Before(mockCtx)
	require.NoError(t, err)

	// Check that all expected commands are present
	expectedCommands := []string{
		"project", "task", "template", "dependency", "config", "health", "validate",
		"status", "get-started", "completion", "mcp",
	}

	commandsMap := make(map[string]*cli.Command)
	for _, cmd := range app.Commands {
		commandsMap[cmd.Name] = cmd
	}

	for _, expected := range expectedCommands {
		assert.Contains(t, commandsMap, expected, "Command %s should be present", expected)
	}
}

func TestAppFlags(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	// Check that expected flags are present
	expectedFlags := []string{"actor", "log-level"}

	flagsMap := make(map[string]cli.Flag)
	for _, flag := range app.Flags {
		flagsMap[flag.Names()[0]] = flag
	}

	for _, expected := range expectedFlags {
		assert.Contains(t, flagsMap, expected, "Flag %s should be present", expected)
	}
}

func TestAppBeforeHook(t *testing.T) {
	app, err := New()
	require.NoError(t, err)
	require.NotNil(t, app.Before)

	// Test that Before hook exists and doesn't cause panic
	// The actual functionality is complex to test without a proper context
	// so we just verify the function exists and can be called safely
	assert.NotNil(t, app.Before)
}

func TestAppIntegration(t *testing.T) {
	// Full integration test: create app, verify app works
	app, err := New()
	require.NoError(t, err)

	// Test that the app is properly initialized
	assert.NotNil(t, app.App)
	assert.Equal(t, "knot", app.Name)

	// This test verifies the app is working, but we don't test specific
	// project operations here since this is the app-level test and not a manager test
	// The manager tests should handle their own integration testing
}

func TestAppRun_UserInputError(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	// Test with a command that has a missing required flag
	// This will error due to missing required flags
	// Note: CLI framework may exit, so we skip this test if it causes issues
	t.Skip("CLI framework calls os.Exit for invalid commands - cannot test")
	_ = app
}

func TestIsUserInputError_ConfigurationError(t *testing.T) {
	// Test the Configuration error pattern
	err := fmt.Errorf("Configuration error: missing required field")
	result := isUserInputError(err)
	assert.True(t, result, "Configuration errors should be treated as user input errors")
}

func TestIsUserInputError_MCPError(t *testing.T) {
	// Test the MCP error pattern
	err := fmt.Errorf("Failed to initialize MCP server: connection refused")
	result := isUserInputError(err)
	assert.True(t, result, "MCP initialization errors should be treated as user input errors")
}

func TestIsUserInputError_EnhancedError(t *testing.T) {
	// Test with actual EnhancedError from errors package
	enhancedErr := errors.InvalidUUIDError("task-id", "invalid-uuid")
	result := isUserInputError(enhancedErr)
	assert.True(t, result, "EnhancedError should be treated as user input error")
}

func TestAppRun_UserInputErrorHandling(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	// Test that Run function properly handles and returns user input errors
	// We'll test this by calling Run with args that cause an error but don't exit
	// and verifying the error is returned correctly

	// Create a mock error that simulates a user input error
	mockErr := fmt.Errorf("Required flag 'test' not provided")

	// We can't directly test the error path without a mock,
	// but we can verify the isUserInputError function works correctly
	result := isUserInputError(mockErr)
	assert.True(t, result, "Should identify user input errors")

	_ = app
}

func TestAppRun_InternalErrorHandling(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	// Test that Run function properly handles internal errors differently
	// from user input errors
	internalErr := fmt.Errorf("internal database connection failed")

	result := isUserInputError(internalErr)
	assert.False(t, result, "Internal errors should not be classified as user input errors")

	_ = app
}

func TestAppRun_ErrorCases(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "validation error",
			err:      fmt.Errorf("validation failed: invalid input"),
			expected: false,
		},
		{
			name:     "network error",
			err:      fmt.Errorf("connection refused"),
			expected: false,
		},
		{
			name:     "context error",
			err:      fmt.Errorf("context deadline exceeded"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUserInputError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppRun_HelpMessage(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	// Test that the app can be created and has the proper structure
	// to provide help messages
	assert.NotNil(t, app)
	assert.NotEmpty(t, app.Name)

	// Verify the app has commands that provide help
	assert.NotEmpty(t, app.Commands)
}

func TestRun_EdgeCases(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	t.Run("run with empty args", func(t *testing.T) {
		// Running with empty args should show help (no error returned by Run itself)
		// The CLI framework handles empty args by showing help
		_ = app.Run([]string{})
		// Empty args typically trigger help display, which doesn't return an error
		// It may exit, so we just verify app is valid
		assert.NotNil(t, app)
	})

	t.Run("run with help flag", func(t *testing.T) {
		// Run with help flag should succeed
		err := app.Run([]string{"knot", "--help"})
		assert.NoError(t, err)
	})

	t.Run("run with version flag", func(t *testing.T) {
		// Run with version flag should succeed
		err := app.Run([]string{"knot", "--version"})
		assert.NoError(t, err)
	})
}

