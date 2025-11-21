package task

import (
	"flag"
	"testing"

	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestActionableAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Set project context
	err := mgr.SetSelectedProject(nil, project.ID, "test-user")
	require.NoError(t, err)

	t.Run("no actionable tasks in empty project", func(t *testing.T) {
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Bool("verbose", false, "")
		flagSet.String("strategy", "", "")
		flagSet.Bool("allow-parent-with-subtasks", false, "")
		flagSet.Bool("prefer-pending", false, "")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := ActionableAction(appCtx)
		err := action(ctx)
		assert.NoError(t, err)
	})

	t.Run("single actionable task", func(t *testing.T) {
		// Create a pending task
		task, err := mgr.CreateTask(nil, project.ID, nil, "Actionable Task", "A task to work on", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Bool("verbose", false, "")
		flagSet.String("strategy", "", "")
		flagSet.Bool("allow-parent-with-subtasks", false, "")
		flagSet.Bool("prefer-pending", false, "")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := ActionableAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify task is still there
		retrievedTask, err := mgr.GetTask(nil, task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.Title, retrievedTask.Title)
	})

	t.Run("actionable with specific strategy", func(t *testing.T) {
		// Create multiple tasks
		task1, err := mgr.CreateTask(nil, project.ID, nil, "High Priority Task", "Important task", 2, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		task2, err := mgr.CreateTask(nil, project.ID, nil, "Low Priority Task", "Less important task", 1, types.TaskPriorityLow, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Bool("verbose", false, "")
		flagSet.String("strategy", "priority", "")
		flagSet.Bool("allow-parent-with-subtasks", false, "")
		flagSet.Bool("prefer-pending", false, "")
		_ = flagSet.Set("strategy", "priority")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := ActionableAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify tasks still exist
		_, err = mgr.GetTask(nil, task1.ID)
		assert.NoError(t, err)
		_, err = mgr.GetTask(nil, task2.ID)
		assert.NoError(t, err)
	})

	t.Run("actionable with verbose output", func(t *testing.T) {
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Bool("verbose", true, "")
		flagSet.String("strategy", "", "")
		flagSet.Bool("allow-parent-with-subtasks", false, "")
		flagSet.Bool("prefer-pending", false, "")
		_ = flagSet.Set("verbose", "true")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := ActionableAction(appCtx)
		err := action(ctx)
		assert.NoError(t, err)
	})

	t.Run("actionable with JSON output", func(t *testing.T) {
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", true, "")
		flagSet.Bool("verbose", false, "")
		flagSet.String("strategy", "", "")
		flagSet.Bool("allow-parent-with-subtasks", false, "")
		flagSet.Bool("prefer-pending", false, "")
		_ = flagSet.Set("json", "true")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := ActionableAction(appCtx)
		err := action(ctx)
		assert.NoError(t, err)
	})
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
	assert.Contains(t, cmd.Aliases, "next")
	assert.NotNil(t, cmd.Action)
	assert.NotEmpty(t, cmd.Flags)
	assert.NotEmpty(t, cmd.Usage)
	assert.NotEmpty(t, cmd.Description)

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
}