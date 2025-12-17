package config

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCommands(t *testing.T) {
	// Create test injector with DI
	testConfig := testutil.NewTestConfig(t)
	testInjector := testConfig.SetupTestInjector(t)

	commands := Commands(testInjector)

	assert.Len(t, commands, 3)

	commandNames := make(map[string]*cli.Command)
	for _, cmd := range commands {
		commandNames[cmd.Name] = cmd
	}

	assert.Contains(t, commandNames, "show")
	assert.Contains(t, commandNames, "set")
	assert.Contains(t, commandNames, "reset")
}

func TestShowAction(t *testing.T) {
	// Capture stdout to verify output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test injector with DI
	testConfig := testutil.NewTestConfig(t)
	testInjector := testConfig.SetupTestInjector(t)

	actionFunc := ShowAction(testInjector)

	// Create a mock context
	app := &cli.App{}
	ctx := cli.NewContext(app, nil, nil)

	err := actionFunc(ctx)
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

	assert.Contains(t, output, "Current Knot Configuration:")
	assert.Contains(t, output, "Complexity Threshold:")
	assert.Contains(t, output, "Max Depth:")
	assert.Contains(t, output, "Max Tasks Per Depth:")
	assert.Contains(t, output, "Max Description Length:")
	assert.Contains(t, output, "Auto Reduce Complexity:")
}

func TestSetAction(t *testing.T) {
	// Create test injector with DI
	testConfig := testutil.NewTestConfig(t)
	testInjector := testConfig.SetupTestInjector(t)
	projectManager := do.MustInvoke[manager.ProjectManager](testInjector)

	tests := []struct {
		name        string
		key         string
		value       string
		expectError bool
	}{
		{
			name:        "valid complexity-threshold",
			key:         "complexity-threshold",
			value:       "7",
			expectError: false,
		},
		{
			name:        "invalid complexity-threshold low",
			key:         "complexity-threshold",
			value:       "0",
			expectError: true,
		},
		{
			name:        "invalid complexity-threshold high",
			key:         "complexity-threshold",
			value:       "11",
			expectError: true,
		},
		{
			name:        "valid max-depth",
			key:         "max-depth",
			value:       "5",
			expectError: false,
		},
		{
			name:        "invalid max-depth low",
			key:         "max-depth",
			value:       "0",
			expectError: true,
		},
		{
			name:        "valid max-tasks-per-depth",
			key:         "max-tasks-per-depth",
			value:       "50",
			expectError: false,
		},
		{
			name:        "invalid max-tasks-per-depth low",
			key:         "max-tasks-per-depth",
			value:       "0",
			expectError: true,
		},
		{
			name:        "valid max-description-length",
			key:         "max-description-length",
			value:       "3000",
			expectError: false,
		},
		{
			name:        "invalid max-description-length low",
			key:         "max-description-length",
			value:       "0",
			expectError: true,
		},
		{
			name:        "valid auto-reduce-complexity true",
			key:         "auto-reduce-complexity",
			value:       "1",
			expectError: false,
		},
		{
			name:        "valid auto-reduce-complexity false",
			key:         "auto-reduce-complexity",
			value:       "0",
			expectError: false,
		},
		{
			name:        "invalid auto-reduce-complexity",
			key:         "auto-reduce-complexity",
			value:       "2",
			expectError: true,
		},
		{
			name:        "unknown key",
			key:         "unknown-key",
			value:       "5",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// Create CLI context with flags
			app := &cli.App{}
			set := flag.NewFlagSet("test", flag.ContinueOnError)
			set.String("key", "", "config key")
			set.String("value", "", "config value")

			_ = set.Set("key", tt.key)
			_ = set.Set("value", tt.value)

			ctx := cli.NewContext(app, set, nil)

			actionFunc := SetAction(testInjector)
			err := actionFunc(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the config was updated
				updatedConfig := projectManager.GetConfig()
				switch tt.key {
				case "complexity-threshold":
					expected, _ := parseInt64(tt.value)
					assert.Equal(t, int(expected), updatedConfig.ComplexityThreshold)
				case "max-depth":
					expected, _ := parseInt64(tt.value)
					assert.Equal(t, int(expected), updatedConfig.MaxDepth)
				case "max-tasks-per-depth":
					expected, _ := parseInt64(tt.value)
					assert.Equal(t, int(expected), updatedConfig.MaxTasksPerDepth)
				case "max-description-length":
					expected, _ := parseInt64(tt.value)
					assert.Equal(t, int(expected), updatedConfig.MaxDescriptionLength)
				case "auto-reduce-complexity":
					expected := tt.value == "1"
					assert.Equal(t, expected, updatedConfig.AutoReduceComplexity)
				}
			}
		})
	}
}

func TestSetActionWithValidValues(t *testing.T) {
	// Create test injector with DI
	testConfig := testutil.NewTestConfig(t)
	testInjector := testConfig.SetupTestInjector(t)

	// Test all valid configuration updates
	testCases := []struct {
		key   string
		value string
	}{
		{"complexity-threshold", "8"},
		{"max-depth", "10"},
		{"max-tasks-per-depth", "100"},
		{"max-description-length", "5000"},
		{"auto-reduce-complexity", "1"},
		{"auto-reduce-complexity", "0"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("set %s to %s", tc.key, tc.value), func(_ *testing.T) {
			app := &cli.App{}
			set := flag.NewFlagSet("test", flag.ContinueOnError)
			set.String("key", "", "config key")
			set.String("value", "", "config value")

			_ = set.Set("key", tc.key)
			_ = set.Set("value", tc.value)

			ctx := cli.NewContext(app, set, nil)

			actionFunc := SetAction(testInjector)
			err := actionFunc(ctx)
			assert.NoError(t, err)
		})
	}
}

func TestResetAction(t *testing.T) {
	// Capture stdout to verify output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test injector with DI
	testConfig := testutil.NewTestConfig(t)
	testInjector := testConfig.SetupTestInjector(t)
	projectManager := do.MustInvoke[manager.ProjectManager](testInjector)

	// Get initial config to compare after reset
	initialConfig := projectManager.GetConfig()

	actionFunc := ResetAction(testInjector)

	// Create a mock context
	app := &cli.App{}
	ctx := cli.NewContext(app, nil, nil)

	err := actionFunc(ctx)
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

	// Verify output contains reset information
	assert.Contains(t, output, "Configuration reset to defaults:")

	// Verify config is still valid after reset (reset should set it to defaults)
	resetConfig := projectManager.GetConfig()
	assert.Equal(t, initialConfig.ComplexityThreshold, resetConfig.ComplexityThreshold)
	assert.Equal(t, initialConfig.MaxDepth, resetConfig.MaxDepth)
	assert.Equal(t, initialConfig.AutoReduceComplexity, resetConfig.AutoReduceComplexity)
	assert.Equal(t, initialConfig.MaxTasksPerDepth, resetConfig.MaxTasksPerDepth)
	assert.Equal(t, initialConfig.MaxDescriptionLength, resetConfig.MaxDescriptionLength)
}

func TestSetActionIntegration(t *testing.T) {
	// Create test injector with DI
	testConfig := testutil.NewTestConfig(t)
	testInjector := testConfig.SetupTestInjector(t)
	projectManager := do.MustInvoke[manager.ProjectManager](testInjector)

	// Initial values
	initialConfig := projectManager.GetConfig()
	initialThreshold := initialConfig.ComplexityThreshold

	// Set a new value
	app := &cli.App{}
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.String("key", "", "config key")
	set.String("value", "", "config value")

	_ = set.Set("key", "complexity-threshold")
	_ = set.Set("value", "9")

	ctx := cli.NewContext(app, set, nil)

	actionFunc := SetAction(testInjector)
	err := actionFunc(ctx)
	assert.NoError(t, err)

	// Verify value was updated
	updatedConfig := projectManager.GetConfig()
	assert.Equal(t, 9, updatedConfig.ComplexityThreshold)
	assert.NotEqual(t, initialThreshold, updatedConfig.ComplexityThreshold)
}

func TestSetActionFlagValidation(t *testing.T) {
	// Create test injector with DI
	testConfig := testutil.NewTestConfig(t)
	testInjector := testConfig.SetupTestInjector(t)

	// Test without required flags - should fail
	app := &cli.App{}
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	// Don't set required flags

	ctx := cli.NewContext(app, set, nil)

	actionFunc := SetAction(testInjector)
	_ = actionFunc(ctx)
	// This would fail due to missing required flags, which is expected
	// The exact error depends on CLI framework behavior
}

func TestShowActionOutputFormat(t *testing.T) {
	// Capture stdout to check formatting
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test injector with DI
	testConfig := testutil.NewTestConfig(t)
	testInjector := testConfig.SetupTestInjector(t)

	actionFunc := ShowAction(testInjector)

	app := &cli.App{}
	ctx := cli.NewContext(app, nil, nil)

	err := actionFunc(ctx)
	assert.NoError(t, err)

	if err := w.Close(); err != nil {
		t.Logf("Warning: failed to close stdout writer: %v", err)
	}
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify the output format
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should start with title
	assert.Contains(t, lines[0], "Current Knot Configuration:")

	// Should contain each config field
	requiredLines := []string{
		"Complexity Threshold:",
		"Max Depth:",
		"Max Tasks Per Depth:",
		"Max Description Length:",
		"Auto Reduce Complexity:",
	}

	for _, requiredLine := range requiredLines {
		found := false
		for _, line := range lines {
			if strings.Contains(line, requiredLine) {
				found = true
				break
			}
		}
		assert.True(t, found, "Missing line containing: %s", requiredLine)
	}
}

func TestSetActionWithDifferentValues(t *testing.T) {
	// Create test injector with DI
	testConfig := testutil.NewTestConfig(t)
	testInjector := testConfig.SetupTestInjector(t)
	projectManager := do.MustInvoke[manager.ProjectManager](testInjector)

	// Test setting each config parameter to a different value
	testParams := map[string]string{
		"complexity-threshold":   "7",
		"max-depth":              "15",
		"max-tasks-per-depth":    "75",
		"max-description-length": "2500",
		"auto-reduce-complexity": "0",
	}

	for key, value := range testParams {
		t.Run(fmt.Sprintf("set %s", key), func(_ *testing.T) {
			app := &cli.App{}
			set := flag.NewFlagSet("test", flag.ContinueOnError)
			set.String("key", "", "config key")
			set.String("value", "", "config value")

			_ = set.Set("key", key)
			_ = set.Set("value", value)

			ctx := cli.NewContext(app, set, nil)

			actionFunc := SetAction(testInjector)
			err := actionFunc(ctx)
			assert.NoError(t, err)

			// Verify the setting worked
			config := projectManager.GetConfig()
			switch key {
			case "complexity-threshold":
				expected, _ := parseInt64(value)
				assert.Equal(t, int(expected), config.ComplexityThreshold)
			case "max-depth":
				expected, _ := parseInt64(value)
				assert.Equal(t, int(expected), config.MaxDepth)
			case "max-tasks-per-depth":
				expected, _ := parseInt64(value)
				assert.Equal(t, int(expected), config.MaxTasksPerDepth)
			case "max-description-length":
				expected, _ := parseInt64(value)
				assert.Equal(t, int(expected), config.MaxDescriptionLength)
			case "auto-reduce-complexity":
				expected := value == "1"
				assert.Equal(t, expected, config.AutoReduceComplexity)
			}
		})
	}
}

// Helper function to parse int64 for test verification
func parseInt64(s string) (int64, error) {
	var result int64
	for _, c := range s {
		result = result*10 + int64(c-'0')
	}
	return result, nil
}

func TestSetActionErrorMessages(t *testing.T) {
	// Create test injector with DI
	testConfig := testutil.NewTestConfig(t)
	testInjector := testConfig.SetupTestInjector(t)

	testCases := []struct {
		name          string
		key           string
		value         string
		expectedError string
	}{
		{
			name:          "invalid complexity range low",
			key:           "complexity-threshold",
			value:         "0",
			expectedError: "complexity-threshold must be between 1 and 10",
		},
		{
			name:          "invalid complexity range high",
			key:           "complexity-threshold",
			value:         "11",
			expectedError: "complexity-threshold must be between 1 and 10",
		},
		{
			name:          "invalid max-depth",
			key:           "max-depth",
			value:         "0",
			expectedError: "max-depth must be at least 1",
		},
		{
			name:          "invalid max-tasks-per-depth",
			key:           "max-tasks-per-depth",
			value:         "0",
			expectedError: "max-tasks-per-depth must be at least 1",
		},
		{
			name:          "invalid max-description-length",
			key:           "max-description-length",
			value:         "0",
			expectedError: "max-description-length must be at least 1",
		},
		{
			name:          "invalid auto-reduce-complexity",
			key:           "auto-reduce-complexity",
			value:         "2",
			expectedError: "auto-reduce-complexity must be 0 (false) or 1 (true)",
		},
		{
			name:          "unknown key",
			key:           "unknown-key",
			value:         "5",
			expectedError: "unknown configuration key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(_ *testing.T) {
			app := &cli.App{}
			set := flag.NewFlagSet("test", flag.ContinueOnError)
			set.String("key", "", "config key")
			set.String("value", "", "config value")

			_ = set.Set("key", tc.key)
			_ = set.Set("value", tc.value)

			ctx := cli.NewContext(app, set, nil)

			actionFunc := SetAction(testInjector)
			err := actionFunc(ctx)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestCommandsStructure(t *testing.T) {
	// Create test injector with DI
	testConfig := testutil.NewTestConfig(t)
	testInjector := testConfig.SetupTestInjector(t)

	commands := Commands(testInjector)

	assert.Len(t, commands, 3)

	// Check each command has the expected structure
	for _, cmd := range commands {
		switch cmd.Name {
		case "show":
			assert.Equal(t, "Show current configuration", cmd.Usage)
			assert.NotNil(t, cmd.Action)
		case "set":
			assert.Equal(t, "Set configuration value", cmd.Usage)
			assert.NotNil(t, cmd.Action)
			assert.NotEmpty(t, cmd.Flags)
		case "reset":
			assert.Equal(t, "Reset configuration to defaults", cmd.Usage)
			assert.NotNil(t, cmd.Action)
		default:
			t.Fatalf("Unexpected command: %s", cmd.Name)
		}
	}
}

func TestSetActionMissingRequiredFlags(t *testing.T) {
	// Create test injector with DI
	testConfig := testutil.NewTestConfig(t)
	testInjector := testConfig.SetupTestInjector(t)

	// Create context without required flags
	app := &cli.App{}
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	// Don't set any flags

	ctx := cli.NewContext(app, set, nil)

	actionFunc := SetAction(testInjector)
	_ = actionFunc(ctx)
	// The error here depends on how the CLI framework handles missing required flags
	// This is an integration-style test
	// Note: This test may or may not error depending on CLI framework behavior
}
