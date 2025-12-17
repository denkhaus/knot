package commands

import (
	"testing"

	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestNewDependencyCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	cmd := NewDependencyCommand(testInjector)

	assert.NotNil(t, cmd)
	assert.Equal(t, "dependency", cmd.Name)
	assert.Equal(t, []string{"dep"}, cmd.Aliases)
	assert.Equal(t, "Task dependency management", cmd.Usage)
	assert.NotNil(t, cmd.Subcommands)
	assert.Empty(t, cmd.Flags) // No flags expected
}

func TestNewTemplateCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	cmd := NewTemplateCommand(testInjector)

	assert.NotNil(t, cmd)
	assert.Equal(t, "template", cmd.Name)
	assert.Equal(t, []string{"tmpl"}, cmd.Aliases)
	assert.Equal(t, "Task template management commands", cmd.Usage)
	assert.NotNil(t, cmd.Subcommands)
	assert.Empty(t, cmd.Flags) // No flags expected
}

func TestNewHealthCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	cmd := NewHealthCommand(testInjector)

	assert.NotNil(t, cmd)
	assert.Equal(t, "health", cmd.Name)
	assert.Equal(t, "Database health and connectivity checks", cmd.Usage)
	assert.NotNil(t, cmd.Subcommands)
	assert.Empty(t, cmd.Flags) // No flags expected
}

func TestNewConfigCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	cmd := NewConfigCommand(testInjector)

	assert.NotNil(t, cmd)
	assert.Equal(t, "config", cmd.Name)
	assert.Equal(t, []string{"cfg"}, cmd.Aliases)
	assert.Equal(t, "Configuration management", cmd.Usage)
	assert.NotNil(t, cmd.Subcommands)
	assert.Empty(t, cmd.Flags) // Main config command has no flags, subcommands have them
}

func TestNewCompletionCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	cmd := NewCompletionCommand(testInjector)

	assert.NotNil(t, cmd)
	assert.Equal(t, "completion", cmd.Name)
	assert.Equal(t, "Generate shell completion scripts", cmd.Usage)
	assert.NotNil(t, cmd.Action)
	assert.NotEmpty(t, cmd.Description)
	assert.Empty(t, cmd.Flags)       // No flags expected
	assert.Empty(t, cmd.Subcommands) // No subcommands expected
}

func TestNewTaskCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	cmd := NewTaskCommand(testInjector)

	assert.NotNil(t, cmd)
	assert.Equal(t, "task", cmd.Name)
	assert.Equal(t, []string{"t"}, cmd.Aliases)
	assert.Equal(t, "Task management commands", cmd.Usage)
	assert.NotNil(t, cmd.Subcommands)
	assert.Empty(t, cmd.Flags) // No flags expected
}

func TestNewGetStartedCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	cmd := NewGetStartedCommand(testInjector)

	assert.NotNil(t, cmd)
	assert.Equal(t, "get-started", cmd.Name)
	assert.Equal(t, "Get started guide for LLM agents with available commands and usage", cmd.Usage)
	assert.NotNil(t, cmd.Action)
	assert.Empty(t, cmd.Flags)       // No flags expected
	assert.Empty(t, cmd.Subcommands) // No subcommands expected
}

func TestNewValidateCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	cmd := NewValidateCommand(testInjector)

	assert.NotNil(t, cmd)
	assert.Equal(t, "validate", cmd.Name)
	assert.Equal(t, "Task state validation and transition checks", cmd.Usage)
	assert.NotNil(t, cmd.Subcommands)
	assert.Empty(t, cmd.Flags) // No flags expected
}

func TestNewProjectCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	cmd := NewProjectCommand(testInjector)

	assert.NotNil(t, cmd)
	assert.Equal(t, "project", cmd.Name)
	assert.Equal(t, []string{"p"}, cmd.Aliases)
	assert.Equal(t, "Project management commands", cmd.Usage)
	assert.NotNil(t, cmd.Subcommands)
	assert.NotEmpty(t, cmd.Flags)

	// Check for expected flags
	flagNames := make(map[string]bool)
	for _, flag := range cmd.Flags {
		flagNames[flag.Names()[0]] = true
	}
	assert.True(t, flagNames["json"])
}

func TestNewStatusCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	cmd := NewStatusCommand(testInjector)

	assert.NotNil(t, cmd)
	assert.Equal(t, "status", cmd.Name)
	assert.Equal(t, []string{"st"}, cmd.Aliases)
	assert.NotNil(t, cmd.Subcommands)
	assert.NotEmpty(t, cmd.Usage)
	assert.NotEmpty(t, cmd.Description)

	// Check that all expected subcommands exist
	subcommandNames := make(map[string]bool)
	for _, subcmd := range cmd.Subcommands {
		subcommandNames[subcmd.Name] = true
	}

	assert.True(t, subcommandNames["actionable"])
	assert.True(t, subcommandNames["ready"])
	assert.True(t, subcommandNames["blocked"])
	assert.True(t, subcommandNames["breakdown"])
}

// Integration tests for command behavior and edge cases

func TestCommandsWithNilInjector(t *testing.T) {
	// Test that all command constructors create valid command structures
	// Using a test injector since some commands need DI for their subcommands
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	assert.NotNil(t, NewDependencyCommand(testInjector))
	assert.NotNil(t, NewTemplateCommand(testInjector))
	assert.NotNil(t, NewHealthCommand(testInjector))
	assert.NotNil(t, NewConfigCommand(testInjector))
	assert.NotNil(t, NewCompletionCommand(testInjector))
	assert.NotNil(t, NewTaskCommand(testInjector))
	assert.NotNil(t, NewStatusCommand(testInjector))
	assert.NotNil(t, NewGetStartedCommand(testInjector))
	assert.NotNil(t, NewValidateCommand(testInjector))
	assert.NotNil(t, NewProjectCommand(testInjector))
}

func TestCommandNamesUniqueness(t *testing.T) {
	// Test that all command names are unique within the CLI context
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	commands := []*cli.Command{
		NewDependencyCommand(testInjector),
		NewTemplateCommand(testInjector),
		NewHealthCommand(testInjector),
		NewConfigCommand(testInjector),
		NewCompletionCommand(testInjector),
		NewTaskCommand(testInjector),
		NewStatusCommand(testInjector),
		NewGetStartedCommand(testInjector),
		NewValidateCommand(testInjector),
		NewProjectCommand(testInjector),
	}

	// Check for duplicate names
	nameSet := make(map[string]bool)
	aliasSet := make(map[string]bool)

	for _, cmd := range commands {
		// Check main command name
		assert.False(t, nameSet[cmd.Name], "Duplicate command name found: %s", cmd.Name)
		nameSet[cmd.Name] = true

		// Check aliases
		for _, alias := range cmd.Aliases {
			assert.False(t, nameSet[alias], "Command alias conflicts with existing name: %s", alias)
			assert.False(t, aliasSet[alias], "Duplicate command alias found: %s", alias)
			aliasSet[alias] = true
		}

		// Validate command structure
		assert.NotEmpty(t, cmd.Name, "Command name should not be empty")
		assert.NotEmpty(t, cmd.Usage, "Command usage should not be empty")
	}
}

func TestCommandsWithValidAppContext(t *testing.T) {
	// Test that all commands work with a valid test injector
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	// All commands should be created successfully
	cmds := []func(do.Injector) *cli.Command{
		NewDependencyCommand,
		NewTemplateCommand,
		NewHealthCommand,
		NewConfigCommand,
		NewCompletionCommand,
		NewTaskCommand,
		NewStatusCommand,
		NewGetStartedCommand,
		NewValidateCommand,
		NewProjectCommand,
	}

	for i, createCmd := range cmds {
		cmd := createCmd(testInjector)
		assert.NotNil(t, cmd, "Command %d should not be nil", i)
		assert.NotEmpty(t, cmd.Name, "Command %d should have a name", i)
		assert.NotEmpty(t, cmd.Usage, "Command %d should have usage text", i)
	}
}

func TestCommandStructureConsistency(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	// Test that subcommands have consistent structure
	cmd := NewTaskCommand(testInjector)
	assert.NotNil(t, cmd.Subcommands, "Task command should have subcommands")

	// Test that commands with actions don't have subcommands
	actionCmds := []*cli.Command{
		NewGetStartedCommand(testInjector),
		NewCompletionCommand(testInjector),
	}

	for _, cmd := range actionCmds {
		assert.NotNil(t, cmd.Action, "Command %s should have an action", cmd.Name)
		assert.Empty(t, cmd.Subcommands, "Command %s with action should not have subcommands", cmd.Name)
	}

	// Test that commands with subcommands don't have actions
	subcommandCmds := []*cli.Command{
		NewDependencyCommand(testInjector),
		NewTemplateCommand(testInjector),
		NewHealthCommand(testInjector),
		NewConfigCommand(testInjector),
		NewTaskCommand(testInjector),
		NewStatusCommand(testInjector),
		NewValidateCommand(testInjector),
		NewProjectCommand(testInjector),
	}

	for _, cmd := range subcommandCmds {
		assert.Empty(t, cmd.Action, "Command %s with subcommands should not have action", cmd.Name)
		assert.NotNil(t, cmd.Subcommands, "Command %s should have subcommands", cmd.Name)
	}
}

func TestCommandFlagTypes(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	// Test status command subcommands have correct flag types
	statusCmd := NewStatusCommand(testInjector)
	assert.NotEmpty(t, statusCmd.Subcommands, "Status command should have subcommands")

	// Find actionable subcommand and test its flags
	var actionableCmd *cli.Command
	for _, subcmd := range statusCmd.Subcommands {
		if subcmd.Name == "actionable" {
			actionableCmd = subcmd
			break
		}
	}
	assert.NotNil(t, actionableCmd, "Should have actionable subcommand")
	assert.Len(t, actionableCmd.Flags, 5, "Actionable subcommand should have 5 flags")

	var strategyFlag *cli.StringFlag
	var jsonFlag *cli.BoolFlag
	var verboseFlag *cli.BoolFlag

	for _, flag := range actionableCmd.Flags {
		switch f := flag.(type) {
		case *cli.StringFlag:
			if f.Name == "strategy" {
				strategyFlag = f
			}
		case *cli.BoolFlag:
			switch f.Name {
			case "json":
				jsonFlag = f
			case "verbose":
				verboseFlag = f
			}
		}
	}

	assert.NotNil(t, strategyFlag, "Should have a StringFlag for strategy")
	assert.NotNil(t, jsonFlag, "Should have a BoolFlag for json")
	assert.NotNil(t, verboseFlag, "Should have a BoolFlag for verbose")
	assert.Equal(t, []string{"v"}, verboseFlag.Aliases, "Verbose flag should have 'v' alias")

	// Find breakdown subcommand and test its flags
	var breakdownCmd *cli.Command
	for _, subcmd := range statusCmd.Subcommands {
		if subcmd.Name == "breakdown" {
			breakdownCmd = subcmd
			break
		}
	}
	assert.NotNil(t, breakdownCmd, "Should have breakdown subcommand")
	assert.Len(t, breakdownCmd.Flags, 4, "Breakdown subcommand should have 4 flags")

	var thresholdFlag *cli.IntFlag
	for _, flag := range breakdownCmd.Flags {
		if intFlag, ok := flag.(*cli.IntFlag); ok && intFlag.Name == "threshold" {
			thresholdFlag = intFlag
			break
		}
	}
	assert.NotNil(t, thresholdFlag, "Should have an IntFlag for threshold")
}
