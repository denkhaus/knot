package commands

import (
	"testing"

	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestNewDependencyCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	cmd := NewDependencyCommand(appCtx)

	assert.NotNil(t, cmd)
	assert.Equal(t, "dependency", cmd.Name)
	assert.Equal(t, []string{"dep"}, cmd.Aliases)
	assert.Equal(t, "Task dependency management", cmd.Usage)
	assert.NotNil(t, cmd.Subcommands)
	assert.Empty(t, cmd.Flags) // No flags expected
}

func TestNewTemplateCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	cmd := NewTemplateCommand(appCtx)

	assert.NotNil(t, cmd)
	assert.Equal(t, "template", cmd.Name)
	assert.Equal(t, []string{"tmpl"}, cmd.Aliases)
	assert.Equal(t, "Task template management commands", cmd.Usage)
	assert.NotNil(t, cmd.Subcommands)
	assert.Empty(t, cmd.Flags) // No flags expected
}

func TestNewHealthCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	cmd := NewHealthCommand(appCtx)

	assert.NotNil(t, cmd)
	assert.Equal(t, "health", cmd.Name)
	assert.Equal(t, "Database health and connectivity checks", cmd.Usage)
	assert.NotNil(t, cmd.Subcommands)
	assert.Empty(t, cmd.Flags) // No flags expected
}

func TestNewConfigCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	cmd := NewConfigCommand(appCtx)

	assert.NotNil(t, cmd)
	assert.Equal(t, "config", cmd.Name)
	assert.Equal(t, []string{"cfg"}, cmd.Aliases)
	assert.Equal(t, "Configuration management", cmd.Usage)
	assert.NotNil(t, cmd.Subcommands)
	assert.Empty(t, cmd.Flags) // Main config command has no flags, subcommands have them
}

func TestNewCompletionCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	cmd := NewCompletionCommand(appCtx)

	assert.NotNil(t, cmd)
	assert.Equal(t, "completion", cmd.Name)
	assert.Equal(t, "Generate shell completion scripts", cmd.Usage)
	assert.NotNil(t, cmd.Action)
	assert.NotEmpty(t, cmd.Description)
	assert.Empty(t, cmd.Flags) // No flags expected
	assert.Empty(t, cmd.Subcommands) // No subcommands expected
}

func TestNewTaskCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	cmd := NewTaskCommand(appCtx)

	assert.NotNil(t, cmd)
	assert.Equal(t, "task", cmd.Name)
	assert.Equal(t, []string{"t"}, cmd.Aliases)
	assert.Equal(t, "Task management commands", cmd.Usage)
	assert.NotNil(t, cmd.Subcommands)
	assert.Empty(t, cmd.Flags) // No flags expected
}

func TestNewBreakdownCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	cmd := NewBreakdownCommand(appCtx)

	assert.NotNil(t, cmd)
	assert.Equal(t, "breakdown", cmd.Name)
	assert.Equal(t, "Find tasks that need breakdown based on complexity", cmd.Usage)
	assert.NotNil(t, cmd.Action)
	assert.NotEmpty(t, cmd.Flags)
	assert.Empty(t, cmd.Subcommands) // No subcommands expected

	// Check for expected flags
	flagNames := make(map[string]bool)
	for _, flag := range cmd.Flags {
		flagNames[flag.Names()[0]] = true
	}

	assert.True(t, flagNames["limit"])
	assert.True(t, flagNames["json"])
	assert.True(t, flagNames["quiet"])
	assert.True(t, flagNames["threshold"])

	// Check specific flag properties
	var thresholdFlag *cli.IntFlag
	for _, flag := range cmd.Flags {
		if intFlag, ok := flag.(*cli.IntFlag); ok && intFlag.Name == "threshold" {
			thresholdFlag = intFlag
			break
		}
	}
	assert.NotNil(t, thresholdFlag)
	assert.Equal(t, 8, thresholdFlag.Value)
	assert.Equal(t, []string{"t"}, thresholdFlag.Aliases)
	assert.Equal(t, "Complexity threshold for breakdown (default: 8)", thresholdFlag.Usage)
	assert.Contains(t, thresholdFlag.EnvVars, "KNOT_COMPLEXITY_THRESHOLD")
}

func TestNewBlockedCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	cmd := NewBlockedCommand(appCtx)

	assert.NotNil(t, cmd)
	assert.Equal(t, "blocked", cmd.Name)
	assert.Equal(t, "Show tasks blocked by dependencies", cmd.Usage)
	assert.NotNil(t, cmd.Action)
	assert.NotEmpty(t, cmd.Flags)
	assert.Empty(t, cmd.Subcommands) // No subcommands expected

	// Check for expected flags
	flagNames := make(map[string]bool)
	for _, flag := range cmd.Flags {
		flagNames[flag.Names()[0]] = true
	}

	assert.True(t, flagNames["limit"])
	assert.True(t, flagNames["json"])
}

func TestNewGetStartedCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	cmd := NewGetStartedCommand(appCtx)

	assert.NotNil(t, cmd)
	assert.Equal(t, "get-started", cmd.Name)
	assert.Equal(t, "Get started guide for LLM agents with available commands and usage", cmd.Usage)
	assert.NotNil(t, cmd.Action)
	assert.Empty(t, cmd.Flags) // No flags expected
	assert.Empty(t, cmd.Subcommands) // No subcommands expected
}

func TestNewReadyCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	cmd := NewReadyCommand(appCtx)

	assert.NotNil(t, cmd)
	assert.Equal(t, "ready", cmd.Name)
	assert.Equal(t, "Show tasks with no blockers (ready to work on)", cmd.Usage)
	assert.NotNil(t, cmd.Action)
	assert.NotEmpty(t, cmd.Flags)
	assert.Empty(t, cmd.Subcommands) // No subcommands expected

	// Check for expected flags
	flagNames := make(map[string]bool)
	for _, flag := range cmd.Flags {
		flagNames[flag.Names()[0]] = true
	}

	assert.True(t, flagNames["limit"])
	assert.True(t, flagNames["json"])
}

func TestNewValidateCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	cmd := NewValidateCommand(appCtx)

	assert.NotNil(t, cmd)
	assert.Equal(t, "validate", cmd.Name)
	assert.Equal(t, "Task state validation and transition checks", cmd.Usage)
	assert.NotNil(t, cmd.Subcommands)
	assert.Empty(t, cmd.Flags) // No flags expected
}

func TestNewProjectCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	cmd := NewProjectCommand(appCtx)

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

func TestNewActionableCommand(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	cmd := NewActionableCommand(appCtx)

	assert.NotNil(t, cmd)
	assert.Equal(t, "actionable", cmd.Name)
	assert.Equal(t, []string{"next"}, cmd.Aliases)
	assert.NotNil(t, cmd.Action)
	assert.NotEmpty(t, cmd.Flags)
	assert.NotEmpty(t, cmd.Usage)
	assert.NotEmpty(t, cmd.Description)
	assert.Empty(t, cmd.Subcommands) // No subcommands expected

	// Test that all expected flags exist
	flagNames := make(map[string]bool)
	for _, flag := range cmd.Flags {
		flagNames[flag.Names()[0]] = true
	}

	assert.True(t, flagNames["strategy"])
	assert.True(t, flagNames["allow-parent-with-subtasks"])
	assert.True(t, flagNames["prefer-pending"])
	assert.True(t, flagNames["verbose"])
	assert.True(t, flagNames["json"])

	// Check specific flag properties
	var strategyFlag *cli.StringFlag
	for _, flag := range cmd.Flags {
		if stringFlag, ok := flag.(*cli.StringFlag); ok && stringFlag.Name == "strategy" {
			strategyFlag = stringFlag
			break
		}
	}
	assert.NotNil(t, strategyFlag)
	assert.Equal(t, []string{"s"}, strategyFlag.Aliases)
	assert.Contains(t, strategyFlag.Usage, "dependency-aware")
	assert.Contains(t, strategyFlag.Usage, "depth-first")
	assert.Contains(t, strategyFlag.Usage, "priority")
	assert.Contains(t, strategyFlag.Usage, "creation-order")
	assert.Contains(t, strategyFlag.Usage, "critical-path")

	// Check boolean flags
	for _, flag := range cmd.Flags {
		if boolFlag, ok := flag.(*cli.BoolFlag); ok {
			switch boolFlag.Name {
			case "allow-parent-with-subtasks":
				assert.Equal(t, "Allow selection of parent tasks even when subtasks exist", boolFlag.Usage)
			case "prefer-pending":
				assert.Equal(t, "Prefer pending tasks over in-progress tasks", boolFlag.Usage)
			case "verbose":
				assert.Equal(t, []string{"v"}, boolFlag.Aliases)
				assert.Contains(t, boolFlag.Usage, "detailed selection reasoning")
			case "json":
				assert.Equal(t, "Output result as JSON", boolFlag.Usage)
			}
		}
	}
}

// Integration tests for command behavior and edge cases

func TestCommandsWithNilAppContext(t *testing.T) {
	// Test that all command constructors handle nil AppContext gracefully
	assert.NotNil(t, NewDependencyCommand(nil))
	assert.NotNil(t, NewTemplateCommand(nil))
	assert.NotNil(t, NewHealthCommand(nil))
	assert.NotNil(t, NewConfigCommand(nil))
	assert.NotNil(t, NewCompletionCommand(nil))
	assert.NotNil(t, NewTaskCommand(nil))
	assert.NotNil(t, NewBreakdownCommand(nil))
	assert.NotNil(t, NewBlockedCommand(nil))
	assert.NotNil(t, NewGetStartedCommand(nil))
	assert.NotNil(t, NewReadyCommand(nil))
	assert.NotNil(t, NewValidateCommand(nil))
	assert.NotNil(t, NewProjectCommand(nil))
	assert.NotNil(t, NewActionableCommand(nil))
}

func TestCommandNamesUniqueness(t *testing.T) {
	// Test that all command names are unique within the CLI context
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	commands := []*cli.Command{
		NewDependencyCommand(appCtx),
		NewTemplateCommand(appCtx),
		NewHealthCommand(appCtx),
		NewConfigCommand(appCtx),
		NewCompletionCommand(appCtx),
		NewTaskCommand(appCtx),
		NewBreakdownCommand(appCtx),
		NewBlockedCommand(appCtx),
		NewGetStartedCommand(appCtx),
		NewReadyCommand(appCtx),
		NewValidateCommand(appCtx),
		NewProjectCommand(appCtx),
		NewActionableCommand(appCtx),
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
	// Test that all commands work with a valid AppContext
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
		Actor:          "test-user",
	}

	// All commands should be created successfully
	cmds := []func(*shared.AppContext) *cli.Command{
		NewDependencyCommand,
		NewTemplateCommand,
		NewHealthCommand,
		NewConfigCommand,
		NewCompletionCommand,
		NewTaskCommand,
		NewBreakdownCommand,
		NewBlockedCommand,
		NewGetStartedCommand,
		NewReadyCommand,
		NewValidateCommand,
		NewProjectCommand,
		NewActionableCommand,
	}

	for i, createCmd := range cmds {
		cmd := createCmd(appCtx)
		assert.NotNil(t, cmd, "Command %d should not be nil", i)
		assert.NotEmpty(t, cmd.Name, "Command %d should have a name", i)
		assert.NotEmpty(t, cmd.Usage, "Command %d should have usage text", i)
	}
}

func TestCommandStructureConsistency(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	// Test that subcommands have consistent structure
	cmd := NewTaskCommand(appCtx)
	assert.NotNil(t, cmd.Subcommands, "Task command should have subcommands")

	// Test that commands with actions don't have subcommands
	actionCmds := []*cli.Command{
		NewBreakdownCommand(appCtx),
		NewBlockedCommand(appCtx),
		NewGetStartedCommand(appCtx),
		NewReadyCommand(appCtx),
		NewCompletionCommand(appCtx),
		NewActionableCommand(appCtx),
	}

	for _, cmd := range actionCmds {
		assert.NotNil(t, cmd.Action, "Command %s should have an action", cmd.Name)
		assert.Empty(t, cmd.Subcommands, "Command %s with action should not have subcommands", cmd.Name)
	}

	// Test that commands with subcommands don't have actions
	subcommandCmds := []*cli.Command{
		NewDependencyCommand(appCtx),
		NewTemplateCommand(appCtx),
		NewHealthCommand(appCtx),
		NewConfigCommand(appCtx),
		NewTaskCommand(appCtx),
		NewValidateCommand(appCtx),
		NewProjectCommand(appCtx),
	}

	for _, cmd := range subcommandCmds {
		assert.Empty(t, cmd.Action, "Command %s with subcommands should not have action", cmd.Name)
		assert.NotNil(t, cmd.Subcommands, "Command %s should have subcommands", cmd.Name)
	}
}

func TestCommandFlagTypes(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	// Test breakdown command has correct flag types
	breakdownCmd := NewBreakdownCommand(appCtx)
	assert.Len(t, breakdownCmd.Flags, 4, "Breakdown command should have 4 flags")

	var thresholdFlag *cli.IntFlag
	for _, flag := range breakdownCmd.Flags {
		if intFlag, ok := flag.(*cli.IntFlag); ok {
			thresholdFlag = intFlag
			break
		}
	}
	assert.NotNil(t, thresholdFlag, "Should have an IntFlag for threshold")

	// Test actionable command has correct flag types
	actionableCmd := NewActionableCommand(appCtx)
	assert.Len(t, actionableCmd.Flags, 5, "Actionable command should have 5 flags")

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
			if f.Name == "json" {
				jsonFlag = f
			} else if f.Name == "verbose" {
				verboseFlag = f
			}
		}
	}

	assert.NotNil(t, strategyFlag, "Should have a StringFlag for strategy")
	assert.NotNil(t, jsonFlag, "Should have a BoolFlag for json")
	assert.NotNil(t, verboseFlag, "Should have a BoolFlag for verbose")
	assert.Equal(t, []string{"v"}, verboseFlag.Aliases, "Verbose flag should have 'v' alias")
}
