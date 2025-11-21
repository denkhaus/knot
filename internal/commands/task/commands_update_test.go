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

func TestUpdateDescriptionAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Set project context
	err := mgr.SetSelectedProject(nil, project.ID, "test-user")
	require.NoError(t, err)

	t.Run("successful description update", func(t *testing.T) {
		// Create a task
		task, err := mgr.CreateTask(nil, project.ID, nil, "Test Task", "Original description", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("id", "", "")
		flagSet.String("description", "", "")
		flagSet.String("actor", "", "")

		_ = flagSet.Set("id", task.ID.String())
		_ = flagSet.Set("description", "Updated description")
		_ = flagSet.Set("actor", "test-updater")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := updateDescriptionAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify the update
		updatedTask, err := mgr.GetTask(nil, task.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated description", updatedTask.Description)
	})

	t.Run("invalid task ID", func(t *testing.T) {
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("id", "", "")
		flagSet.String("description", "", "")
		flagSet.String("actor", "", "")

		_ = flagSet.Set("id", "invalid-uuid")
		_ = flagSet.Set("description", "New description")
		_ = flagSet.Set("actor", "test-updater")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := updateDescriptionAction(appCtx)
		err := action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task-id format")
	})
}

func TestUpdatePriorityAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Set project context
	err := mgr.SetSelectedProject(nil, project.ID, "test-user")
	require.NoError(t, err)

	t.Run("successful priority update", func(t *testing.T) {
		// Create a task
		task, err := mgr.CreateTask(nil, project.ID, nil, "Test Task", "Description", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("id", "", "")
		flagSet.String("priority", "", "")
		flagSet.String("actor", "", "")

		_ = flagSet.Set("id", task.ID.String())
		_ = flagSet.Set("priority", "high")
		_ = flagSet.Set("actor", "test-updater")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := updatePriorityAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify the update
		updatedTask, err := mgr.GetTask(nil, task.ID)
		require.NoError(t, err)
		assert.Equal(t, types.TaskPriorityHigh, updatedTask.Priority)
	})
}

func TestGetAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Set project context
	err := mgr.SetSelectedProject(nil, project.ID, "test-user")
	require.NoError(t, err)

	t.Run("get existing task", func(t *testing.T) {
		// Create a task
		task, err := mgr.CreateTask(nil, project.ID, nil, "Test Task", "Test description", 5, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("id", "", "")
		flagSet.Bool("json", false, "")
		flagSet.Bool("quiet", false, "")

		_ = flagSet.Set("id", task.ID.String())

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := getAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)
	})

	t.Run("invalid task ID", func(t *testing.T) {
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("id", "", "")
		flagSet.Bool("json", false, "")
		flagSet.Bool("quiet", false, "")

		_ = flagSet.Set("id", "invalid-uuid")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := getAction(appCtx)
		err := action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task-id format")
	})
}