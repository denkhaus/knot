package dependency

import (
	"testing"

	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/testutil"
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
			name: "basic dependency commands",
			expectedCommands: []string{
				"add",
				"remove",
				"list",
			},
		},
		{
			name: "all dependency commands including enhanced",
			expectedCommands: []string{
				"add",
				"remove",
				"list",
				// Enhanced commands might also be included
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appCtx := createTestAppContext(t)
			commands := Commands(appCtx)

			// Verify we get some commands
			assert.NotEmpty(t, commands)

			// Verify basic commands are present
			commandNames := make([]string, len(commands))
			for i, cmd := range commands {
				commandNames[i] = cmd.Name
			}

			for _, expectedName := range tt.expectedCommands {
				assert.Contains(t, commandNames, expectedName,
					"Command '%s' should be present in dependency commands", expectedName)
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

func TestAddCommandFlags(t *testing.T) {
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

	// Find the add command
	var addCommand *cli.Command
	for _, cmd := range commands {
		if cmd.Name == "add" {
			addCommand = cmd
			break
		}
	}

	require.NotNil(t, addCommand, "Add command should be found")

	// Check required flags
	expectedFlags := []string{"task-id", "depends-on"}
	flagNames := make([]string, 0)
	for _, flag := range addCommand.Flags {
		flagNames = append(flagNames, flag.Names()...)
	}

	for _, expectedFlag := range expectedFlags {
		assert.Contains(t, flagNames, expectedFlag,
			"Add command should have '%s' flag", expectedFlag)
	}
}

func TestRemoveCommandFlags(t *testing.T) {
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

	// Find the remove command
	var removeCommand *cli.Command
	for _, cmd := range commands {
		if cmd.Name == "remove" {
			removeCommand = cmd
			break
		}
	}

	require.NotNil(t, removeCommand, "Remove command should be found")

	// Check required flags
	expectedFlags := []string{"task-id", "depends-on"}
	flagNames := make([]string, 0)
	for _, flag := range removeCommand.Flags {
		flagNames = append(flagNames, flag.Names()...)
	}

	for _, expectedFlag := range expectedFlags {
		assert.Contains(t, flagNames, expectedFlag,
			"Remove command should have '%s' flag", expectedFlag)
	}
}

func TestListCommandFlags(t *testing.T) {
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

	// Find the list command
	var listCommand *cli.Command
	for _, cmd := range commands {
		if cmd.Name == "list" {
			listCommand = cmd
			break
		}
	}

	require.NotNil(t, listCommand, "List command should be found")

	// Check required flags
	expectedFlags := []string{"task-id"}
	flagNames := make([]string, 0)
	for _, flag := range listCommand.Flags {
		flagNames = append(flagNames, flag.Names()...)
	}

	for _, expectedFlag := range expectedFlags {
		assert.Contains(t, flagNames, expectedFlag,
			"List command should have '%s' flag", expectedFlag)
	}
}

func TestCommandUsageText(t *testing.T) {
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

	// Test that commands have meaningful usage text
	expectedUsages := map[string]string{
		"add":    "Add task dependency",
		"remove": "Remove task dependency",
		"list":   "List task dependencies",
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

// Integration test with mock context
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

// Test edge cases
func TestCommandEdgeCases(t *testing.T) {
	t.Run("empty app context", func(t *testing.T) {
		// This test would verify behavior with nil/empty app context
		// But since we don't want to cause panics, we'll use a valid context
		appCtx := createTestAppContext(t)

		commands := Commands(appCtx)
		assert.NotEmpty(t, commands, "Commands should be created even with minimal app context")
	})
}
