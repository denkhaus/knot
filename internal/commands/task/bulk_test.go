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

func TestBulkCommands(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	commands := BulkCommands(appCtx)

	assert.NotEmpty(t, commands)
	
	// Check for expected bulk commands
	commandNames := make(map[string]bool)
	for _, cmd := range commands {
		commandNames[cmd.Name] = true
	}

	assert.True(t, commandNames["bulk-update"])
	assert.True(t, commandNames["duplicate"])
	assert.True(t, commandNames["list-by-state"])
	assert.True(t, commandNames["bulk-create"])
	assert.True(t, commandNames["bulk-delete"])
}

func TestListByStateAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Set project context
	err := mgr.SetSelectedProject(nil, project.ID, "test-user")
	require.NoError(t, err)

	t.Run("list pending tasks", func(t *testing.T) {
		// Create tasks with different states
		pendingTask, err := mgr.CreateTask(nil, project.ID, nil, "Pending Task", "A pending task", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		inProgressTask, err := mgr.CreateTask(nil, project.ID, nil, "In Progress Task", "An in-progress task", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)
		_, err = mgr.UpdateTaskState(nil, inProgressTask.ID, types.TaskStateInProgress, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("state", "", "")
		flagSet.Bool("json", false, "")
		_ = flagSet.Set("state", "pending")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := ListByStateAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify tasks still exist
		_, err = mgr.GetTask(nil, pendingTask.ID)
		assert.NoError(t, err)
		_, err = mgr.GetTask(nil, inProgressTask.ID)
		assert.NoError(t, err)
	})

	t.Run("invalid state", func(t *testing.T) {
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("state", "", "")
		flagSet.Bool("json", false, "")
		_ = flagSet.Set("state", "invalid-state")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := ListByStateAction(appCtx)
		err := action(ctx)
		// Note: Action may not error for invalid state, just return empty results
		assert.NoError(t, err)
	})

	t.Run("missing state parameter", func(t *testing.T) {
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("state", "", "")
		flagSet.Bool("json", false, "")
		// Don't set state

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := ListByStateAction(appCtx)
		err := action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state is required")
	})
}

func TestDuplicateAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Set project context
	err := mgr.SetSelectedProject(nil, project.ID, "test-user")
	require.NoError(t, err)

	t.Run("duplicate existing task", func(t *testing.T) {
		// Create a task to duplicate
		originalTask, err := mgr.CreateTask(nil, project.ID, nil, "Original Task", "Original description", 4, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("task-id", "", "")
		flagSet.String("target-project-id", "", "")
		flagSet.String("actor", "", "")
		_ = flagSet.Set("task-id", originalTask.ID.String())
		_ = flagSet.Set("target-project-id", project.ID.String())
		_ = flagSet.Set("actor", "test-duplicator")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := DuplicateAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify original task still exists
		_, err = mgr.GetTask(nil, originalTask.ID)
		assert.NoError(t, err)
	})

	t.Run("missing task ID", func(t *testing.T) {
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("task-id", "", "")
		flagSet.String("actor", "", "")
		// Don't set task-id
		_ = flagSet.Set("actor", "test-duplicator")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := DuplicateAction(appCtx)
		err := action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task-id is required")
	})
}