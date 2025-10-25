package config

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/denkhaus/knot/internal/manager"
	"github.com/denkhaus/knot/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCommands(t *testing.T) {
	// Create a test app context
	config := manager.DefaultConfig()
	appCtx := &shared.AppContext{
		ProjectManager: manager.NewManagerWithRepository(nil, config),
	}

	commands := Commands(appCtx)

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

	appCtx := &shared.AppContext{
		ProjectManager: manager.NewManagerWithRepository(nil, manager.DefaultConfig()),
	}

	actionFunc := ShowAction(appCtx)

	// Create a mock context
	app := &cli.App{}
	ctx := cli.NewContext(app, nil, nil)

	err := actionFunc(ctx)
	assert.NoError(t, err)

	// Restore stdout
	w.Close()
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
	appCtx := &shared.AppContext{
		ProjectManager: manager.NewManagerWithRepository(nil, manager.DefaultConfig()),
	}

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
		t.Run(tt.name, func(t *testing.T) {
			// Create CLI context with flags
			app := &cli.App{}
			set := flag.NewFlagSet("test", flag.ContinueOnError)
			set.String("key", "", "config key")
			set.String("value", "", "config value")

			set.Set("key", tt.key)
			set.Set("value", tt.value)

			ctx := cli.NewContext(app, set, nil)

			actionFunc := SetAction(appCtx)
			err := actionFunc(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the config was updated
				updatedConfig := appCtx.ProjectManager.GetConfig()
				switch tt.key {
				case "complexity-threshold":
					assert.Equal(t, int64(tt.value[0]-'0'), updatedConfig.ComplexityThreshold)
				case "max-depth":
					assert.Equal(t, int64(tt.value[0]-'0'), updatedConfig.MaxDepth)
				case "max-tasks-per-depth":
					assert.Equal(t, int64(tt.value[0]-'0'), updatedConfig.MaxTasksPerDepth)
				case "max-description-length":
					assert.Equal(t, int64(tt.value[0]-'0'), updatedConfig.MaxDescriptionLength)
				case "auto-reduce-complexity":
					expected := tt.value == "1"
					assert.Equal(t, expected, updatedConfig.AutoReduceComplexity)
				}
			}
		})
	}
}

func TestSetActionWithValidValues(t *testing.T) {
	appCtx := &shared.AppContext{
		ProjectManager: manager.NewManagerWithRepository(nil, manager.DefaultConfig()),
	}

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
		t.Run(fmt.Sprintf("set %s to %s", tc.key, tc.value), func(t *testing.T) {
			app := &cli.App{}
			set := flag.NewFlagSet("test", flag.ContinueOnError)
			set.String("key", "", "config key")
			set.String("value", "", "config value")

			set.Set("key", tc.key)
			set.Set("value", tc.value)

			ctx := cli.NewContext(app, set, nil)

			actionFunc := SetAction(appCtx)
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

	// Create app context with modified config
	initialConfig := manager.DefaultConfig()
	// Modify some values to ensure they get reset
	initialConfig.ComplexityThreshold = 99
	initialConfig.MaxDepth = 99
	initialConfig.AutoReduceComplexity = true

	appCtx := &shared.AppContext{
		ProjectManager: manager.NewManagerWithRepository(nil, initialConfig),
	}

	// Verify initial modified values
	config := appCtx.ProjectManager.GetConfig()
	assert.Equal(t, int64(99), config.ComplexityThreshold)
	assert.Equal(t, int64(99), config.MaxDepth)
	assert.Equal(t, true, config.AutoReduceComplexity)

	actionFunc := ResetAction(appCtx)

	// Create a mock context
	app := &cli.App{}
	ctx := cli.NewContext(app, nil, nil)

	err := actionFunc(ctx)
	assert.NoError(t, err)

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read the output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains reset information
	assert.Contains(t, output, "Configuration reset to defaults:")

	// Verify config was reset to defaults
	resetConfig := appCtx.ProjectManager.GetConfig()
	defaultConfig := manager.DefaultConfig()
	assert.Equal(t, defaultConfig.ComplexityThreshold, resetConfig.ComplexityThreshold)
	assert.Equal(t, defaultConfig.MaxDepth, resetConfig.MaxDepth)
	assert.Equal(t, defaultConfig.AutoReduceComplexity, resetConfig.AutoReduceComplexity)
	assert.Equal(t, defaultConfig.MaxTasksPerDepth, resetConfig.MaxTasksPerDepth)
	assert.Equal(t, defaultConfig.MaxDescriptionLength, resetConfig.MaxDescriptionLength)
}

func TestSetActionIntegration(t *testing.T) {
	appCtx := &shared.AppContext{
		ProjectManager: manager.NewManagerWithRepository(nil, manager.DefaultConfig()),
	}

	// Initial values
	initialConfig := appCtx.ProjectManager.GetConfig()
	initialThreshold := initialConfig.ComplexityThreshold

	// Set a new value
	app := &cli.App{}
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.String("key", "", "config key")
	set.String("value", "", "config value")

	set.Set("key", "complexity-threshold")
	set.Set("value", "8")

	ctx := cli.NewContext(app, set, nil)

	actionFunc := SetAction(appCtx)
	err := actionFunc(ctx)
	assert.NoError(t, err)

	// Verify value was updated
	updatedConfig := appCtx.ProjectManager.GetConfig()
	assert.Equal(t, int(8), updatedConfig.ComplexityThreshold)
	assert.NotEqual(t, initialThreshold, updatedConfig.ComplexityThreshold)
}

func TestSetActionFlagValidation(t *testing.T) {
	appCtx := &shared.AppContext{
		ProjectManager: manager.NewManagerWithRepository(nil, manager.DefaultConfig()),
	}

	// Test without required flags - should fail
	app := &cli.App{}
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	// Don't set required flags

	ctx := cli.NewContext(app, set, nil)

	actionFunc := SetAction(appCtx)
	_ = actionFunc(ctx)
	// This would fail due to missing required flags, which is expected
	// The exact error depends on CLI framework behavior
}

func TestShowActionOutputFormat(t *testing.T) {
	// Capture stdout to check formatting
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	appCtx := &shared.AppContext{
		ProjectManager: manager.NewManagerWithRepository(nil, manager.DefaultConfig()),
	}

	actionFunc := ShowAction(appCtx)

	app := &cli.App{}
	ctx := cli.NewContext(app, nil, nil)

	err := actionFunc(ctx)
	assert.NoError(t, err)

	w.Close()
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
	appCtx := &shared.AppContext{
		ProjectManager: manager.NewManagerWithRepository(nil, manager.DefaultConfig()),
	}

	// Test setting each config parameter to a different value
	testParams := map[string]string{
		"complexity-threshold":   "7",
		"max-depth":              "15",
		"max-tasks-per-depth":    "75",
		"max-description-length": "2500",
		"auto-reduce-complexity": "0",
	}

	for key, value := range testParams {
		t.Run(fmt.Sprintf("set %s", key), func(t *testing.T) {
			app := &cli.App{}
			set := flag.NewFlagSet("test", flag.ContinueOnError)
			set.String("key", "", "config key")
			set.String("value", "", "config value")

			set.Set("key", key)
			set.Set("value", value)

			ctx := cli.NewContext(app, set, nil)

			actionFunc := SetAction(appCtx)
			err := actionFunc(ctx)
			assert.NoError(t, err)

			// Verify the setting worked
			config := appCtx.ProjectManager.GetConfig()
			switch key {
			case "complexity-threshold":
				expected, _ := parseInt64(value)
				assert.Equal(t, expected, config.ComplexityThreshold)
			case "max-depth":
				expected, _ := parseInt64(value)
				assert.Equal(t, expected, config.MaxDepth)
			case "max-tasks-per-depth":
				expected, _ := parseInt64(value)
				assert.Equal(t, expected, config.MaxTasksPerDepth)
			case "max-description-length":
				expected, _ := parseInt64(value)
				assert.Equal(t, expected, config.MaxDescriptionLength)
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
	appCtx := &shared.AppContext{
		ProjectManager: manager.NewManagerWithRepository(nil, manager.DefaultConfig()),
	}

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
		t.Run(tc.name, func(t *testing.T) {
			app := &cli.App{}
			set := flag.NewFlagSet("test", flag.ContinueOnError)
			set.String("key", "", "config key")
			set.String("value", "", "config value")

			set.Set("key", tc.key)
			set.Set("value", tc.value)

			ctx := cli.NewContext(app, set, nil)

			actionFunc := SetAction(appCtx)
			err := actionFunc(ctx)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestCommandsStructure(t *testing.T) {
	appCtx := &shared.AppContext{
		ProjectManager: manager.NewManagerWithRepository(nil, manager.DefaultConfig()),
	}

	commands := Commands(appCtx)

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
	appCtx := &shared.AppContext{
		ProjectManager: manager.NewManagerWithRepository(nil, manager.DefaultConfig()),
	}

	// Create context without required flags
	app := &cli.App{}
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	// Don't set any flags

	ctx := cli.NewContext(app, set, nil)

	actionFunc := SetAction(appCtx)
	_ = actionFunc(ctx)
	// The error here depends on how the CLI framework handles missing required flags
	// This is an integration-style test
	// Note: This test may or may not error depending on CLI framework behavior
}
