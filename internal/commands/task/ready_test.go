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

func TestReadyAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Set project context
	err := mgr.SetSelectedProject(nil, project.ID, "test-user")
	require.NoError(t, err)

	t.Run("pending task is ready", func(t *testing.T) {
		// Create a pending task
		task, err := mgr.CreateTask(nil, project.ID, nil, "Ready Task", "A task ready to work on", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := ReadyAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify task is still there
		retrievedTask, err := mgr.GetTask(nil, task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.Title, retrievedTask.Title)
		assert.Equal(t, types.TaskStatePending, retrievedTask.State)
	})

	t.Run("in-progress task is ready", func(t *testing.T) {
		// Create an in-progress task
		task, err := mgr.CreateTask(nil, project.ID, nil, "In Progress Task", "A task in progress", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		// Update to in-progress
		_, err = mgr.UpdateTaskState(nil, task.ID, types.TaskStateInProgress, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := ReadyAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify task is still there
		retrievedTask, err := mgr.GetTask(nil, task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.Title, retrievedTask.Title)
		assert.Equal(t, types.TaskStateInProgress, retrievedTask.State)
	})

	t.Run("completed task is not ready", func(t *testing.T) {
		// Create and complete a task
		task, err := mgr.CreateTask(nil, project.ID, nil, "Completed Task", "A completed task", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		// First transition to in-progress, then to completed (following valid transition rules)
		_, err = mgr.UpdateTaskState(nil, task.ID, types.TaskStateInProgress, "test-user")
		require.NoError(t, err)
		_, err = mgr.UpdateTaskState(nil, task.ID, types.TaskStateCompleted, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := ReadyAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)
	})
}