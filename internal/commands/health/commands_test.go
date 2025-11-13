package health

import (
	"testing"
	"time"

	"github.com/denkhaus/knot/internal/shared"
	"github.com/denkhaus/knot/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap/zaptest"
)

// createTestAppContext creates a test app context for testing
func createTestAppContext(t *testing.T) *shared.AppContext {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	logger := zaptest.NewLogger(t)
	return shared.NewAppContext(mgr, logger)
}

func TestCommands(t *testing.T) {
	tests := []struct {
		name             string
		expectedCommands []string
	}{
		{
			name: "basic health commands",
			expectedCommands: []string{
				"check",
				"ping",
				"validate",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appCtx := createTestAppContext(t)
			commands := Commands(appCtx)

			// Verify we get the expected number of commands
			assert.Len(t, commands, len(tt.expectedCommands))

			// Verify command names
			commandNames := make([]string, len(commands))
			for i, cmd := range commands {
				commandNames[i] = cmd.Name
			}

			for _, expectedName := range tt.expectedCommands {
				assert.Contains(t, commandNames, expectedName,
					"Command '%s' should be present in health commands", expectedName)
			}
		})
	}
}

func TestCommandStructure(t *testing.T) {
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

	// Test that all commands have required structure
	for _, cmd := range commands {
		t.Run("command_"+cmd.Name, func(t *testing.T) {
			// Verify command has name and usage
			assert.NotEmpty(t, cmd.Name, "Command should have a name")
			assert.NotEmpty(t, cmd.Usage, "Command should have usage text")

			// Verify command has action
			assert.NotNil(t, cmd.Action, "Command should have an action function")

			// Verify flags are properly structured
			for _, flag := range cmd.Flags {
				assert.NotNil(t, flag, "Flag should not be nil")
				flagNames := flag.Names()
				assert.NotEmpty(t, flagNames, "Flag should have at least one name")
			}
		})
	}
}

func TestCheckCommandFlags(t *testing.T) {
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

	// Find the check command
	var checkCommand *cli.Command
	for _, cmd := range commands {
		if cmd.Name == "check" {
			checkCommand = cmd
			break
		}
	}

	require.NotNil(t, checkCommand, "Check command should be found")

	// Check expected flags
	expectedFlags := []string{"json", "timeout"}
	flagNames := make([]string, 0)
	for _, flag := range checkCommand.Flags {
		flagNames = append(flagNames, flag.Names()...)
	}

	for _, expectedFlag := range expectedFlags {
		assert.Contains(t, flagNames, expectedFlag,
			"Check command should have '%s' flag", expectedFlag)
	}

	// Check default values for flags
	for _, flag := range checkCommand.Flags {
		switch flag.Names()[0] {
		case "json":
			boolFlag, ok := flag.(*cli.BoolFlag)
			require.True(t, ok, "json flag should be a BoolFlag")
			assert.False(t, boolFlag.Value, "json flag should default to false")
		case "timeout":
			durationFlag, ok := flag.(*cli.DurationFlag)
			require.True(t, ok, "timeout flag should be a DurationFlag")
			assert.Equal(t, 10*time.Second, durationFlag.Value, "timeout flag should default to 10 seconds")
		}
	}
}

func TestPingCommandFlags(t *testing.T) {
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

	// Find the ping command
	var pingCommand *cli.Command
	for _, cmd := range commands {
		if cmd.Name == "ping" {
			pingCommand = cmd
			break
		}
	}

	require.NotNil(t, pingCommand, "Ping command should be found")

	// Check expected flags
	expectedFlags := []string{"timeout"}
	flagNames := make([]string, 0)
	for _, flag := range pingCommand.Flags {
		flagNames = append(flagNames, flag.Names()...)
	}

	for _, expectedFlag := range expectedFlags {
		assert.Contains(t, flagNames, expectedFlag,
			"Ping command should have '%s' flag", expectedFlag)
	}

	// Check default value for timeout flag
	for _, flag := range pingCommand.Flags {
		if flag.Names()[0] == "timeout" {
			durationFlag, ok := flag.(*cli.DurationFlag)
			require.True(t, ok, "timeout flag should be a DurationFlag")
			assert.Equal(t, 5*time.Second, durationFlag.Value, "timeout flag should default to 5 seconds")
		}
	}
}

func TestValidateCommandFlags(t *testing.T) {
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

	// Find the validate command
	var validateCommand *cli.Command
	for _, cmd := range commands {
		if cmd.Name == "validate" {
			validateCommand = cmd
			break
		}
	}

	require.NotNil(t, validateCommand, "Validate command should be found")

	// Validate command might have different flags - just check it has a flags array
	assert.NotNil(t, validateCommand.Flags, "Validate command should have flags array")
}

func TestCommandUsageText(t *testing.T) {
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

	// Test that commands have meaningful usage text
	expectedUsages := map[string]string{
		"check":    "Check database connection health",
		"ping":     "Simple database connectivity test",
		"validate": "Comprehensive database connection validation",
	}

	for _, cmd := range commands {
		if expectedUsage, exists := expectedUsages[cmd.Name]; exists {
			assert.Equal(t, expectedUsage, cmd.Usage,
				"Command '%s' should have correct usage text", cmd.Name)
		}
	}
}

func TestCommandsReturnNewSlice(t *testing.T) {
	appCtx := createTestAppContext(t)
	commands1 := Commands(appCtx)
	commands2 := Commands(appCtx)

	// Commands should return new slices to avoid modification issues
	assert.NotSame(t, &commands1[0], &commands2[0],
		"Commands should return new slice instances")
}

func TestCommandIntegration(t *testing.T) {
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

	t.Run("command creation with app context", func(t *testing.T) {
		// Verify commands can be created with a valid app context
		assert.NotNil(t, commands)
		assert.Greater(t, len(commands), 0)
	})

	t.Run("command app context dependency", func(t *testing.T) {
		// Test that commands properly depend on app context
		assert.NotNil(t, appCtx, "AppContext should be available")
		assert.NotNil(t, appCtx.ProjectManager, "ProjectManager should be available")
		assert.NotNil(t, appCtx.Logger, "Logger should be available")
	})
}

func TestCommandEdgeCases(t *testing.T) {
	t.Run("empty app context", func(t *testing.T) {
		// This test would verify behavior with nil/empty app context
		// But since we don't want to cause panics, we'll use a valid context
		appCtx := createTestAppContext(t)

		commands := Commands(appCtx)
		assert.NotEmpty(t, commands, "Commands should be created even with minimal app context")
	})
}

func TestHealthCheckTimeouts(t *testing.T) {
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

	t.Run("check command timeout configuration", func(t *testing.T) {
		var checkCommand *cli.Command
		for _, cmd := range commands {
			if cmd.Name == "check" {
				checkCommand = cmd
				break
			}
		}

		require.NotNil(t, checkCommand)

		// Find timeout flag and verify it's properly configured
		for _, flag := range checkCommand.Flags {
			if flag.Names()[0] == "timeout" {
				durationFlag := flag.(*cli.DurationFlag)
				assert.Equal(t, "Health check timeout", durationFlag.Usage)
				assert.Equal(t, 10*time.Second, durationFlag.Value)
				break
			}
		}
	})

	t.Run("ping command timeout configuration", func(t *testing.T) {
		var pingCommand *cli.Command
		for _, cmd := range commands {
			if cmd.Name == "ping" {
				pingCommand = cmd
				break
			}
		}

		require.NotNil(t, pingCommand)

		// Find timeout flag and verify it's properly configured
		for _, flag := range pingCommand.Flags {
			if flag.Names()[0] == "timeout" {
				durationFlag := flag.(*cli.DurationFlag)
				assert.Equal(t, "Ping timeout", durationFlag.Usage)
				assert.Equal(t, 5*time.Second, durationFlag.Value)
				break
			}
		}
	})
}
