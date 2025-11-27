package task

import (
	"context"
	"flag"
	"testing"

	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestGetStartedAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Set project context
	err := mgr.SetSelectedProject(context.TODO(), project.ID, "test-user")
	require.NoError(t, err)

	t.Run("get started in empty project", func(t *testing.T) {
		// Create a new project for clean test
		project2 := testutil.CreateTestProject(t, mgr)
		err := mgr.SetSelectedProject(context.TODO(), project2.ID, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := GetStartedAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)
	})

	t.Run("get started with existing tasks", func(t *testing.T) {
		// Create some tasks in various states
		pendingTask, err := mgr.CreateTask(context.TODO(), project.ID, nil, "Pending Task", "A pending task", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		inProgressTask, err := mgr.CreateTask(context.TODO(), project.ID, nil, "In Progress Task", "An in-progress task", 4, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)
		_, err = mgr.UpdateTaskState(context.TODO(), inProgressTask.ID, types.TaskStateInProgress, "test-user")
		require.NoError(t, err)

		blockedTask, err := mgr.CreateTask(context.TODO(), project.ID, nil, "Blocked Task", "A blocked task", 2, types.TaskPriorityLow, "test-user")
		require.NoError(t, err)
		_, err = mgr.UpdateTaskState(context.TODO(), blockedTask.ID, types.TaskStateBlocked, "test-user")
		require.NoError(t, err)

		completedTask, err := mgr.CreateTask(context.TODO(), project.ID, nil, "Completed Task", "A completed task", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)
		_, err = mgr.UpdateTaskState(context.TODO(), completedTask.ID, types.TaskStateInProgress, "test-user")
		require.NoError(t, err)
		_, err = mgr.UpdateTaskState(context.TODO(), completedTask.ID, types.TaskStateCompleted, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := GetStartedAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify all tasks still exist
		_, err = mgr.GetTask(context.TODO(), pendingTask.ID)
		assert.NoError(t, err)
		_, err = mgr.GetTask(context.TODO(), inProgressTask.ID)
		assert.NoError(t, err)
		_, err = mgr.GetTask(context.TODO(), blockedTask.ID)
		assert.NoError(t, err)
		_, err = mgr.GetTask(context.TODO(), completedTask.ID)
		assert.NoError(t, err)
	})

	t.Run("get started with hierarchical tasks", func(t *testing.T) {
		// Create parent and child tasks
		parentTask, err := mgr.CreateTask(context.TODO(), project.ID, nil, "Parent Task", "A parent task", 6, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		childTask1, err := mgr.CreateTask(context.TODO(), project.ID, &parentTask.ID, "Child Task 1", "First child", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		childTask2, err := mgr.CreateTask(context.TODO(), project.ID, &parentTask.ID, "Child Task 2", "Second child", 2, types.TaskPriorityLow, "test-user")
		require.NoError(t, err)

		// Complete one child task
		_, err = mgr.UpdateTaskState(context.TODO(), childTask2.ID, types.TaskStateInProgress, "test-user")
		require.NoError(t, err)
		_, err = mgr.UpdateTaskState(context.TODO(), childTask2.ID, types.TaskStateCompleted, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := GetStartedAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify tasks still exist
		_, err = mgr.GetTask(context.TODO(), parentTask.ID)
		assert.NoError(t, err)
		_, err = mgr.GetTask(context.TODO(), childTask1.ID)
		assert.NoError(t, err)
		_, err = mgr.GetTask(context.TODO(), childTask2.ID)
		assert.NoError(t, err)
	})

	t.Run("get started with high complexity tasks", func(t *testing.T) {
		// Create tasks with different complexity levels
		simpleTask, err := mgr.CreateTask(context.TODO(), project.ID, nil, "Simple Task", "Easy task", 2, types.TaskPriorityLow, "test-user")
		require.NoError(t, err)

		complexTask, err := mgr.CreateTask(context.TODO(), project.ID, nil, "Complex Task", "Difficult task", 9, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := GetStartedAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify tasks still exist
		_, err = mgr.GetTask(context.TODO(), simpleTask.ID)
		assert.NoError(t, err)
		_, err = mgr.GetTask(context.TODO(), complexTask.ID)
		assert.NoError(t, err)
	})

	t.Run("get started with mixed priorities", func(t *testing.T) {
		// Create tasks with different priorities
		lowPriorityTask, err := mgr.CreateTask(context.TODO(), project.ID, nil, "Low Priority", "Can wait", 3, types.TaskPriorityLow, "test-user")
		require.NoError(t, err)

		mediumPriorityTask, err := mgr.CreateTask(context.TODO(), project.ID, nil, "Medium Priority", "Normal importance", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		highPriorityTask, err := mgr.CreateTask(context.TODO(), project.ID, nil, "High Priority", "Urgent task", 3, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := GetStartedAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify tasks still exist
		_, err = mgr.GetTask(context.TODO(), lowPriorityTask.ID)
		assert.NoError(t, err)
		_, err = mgr.GetTask(context.TODO(), mediumPriorityTask.ID)
		assert.NoError(t, err)
		_, err = mgr.GetTask(context.TODO(), highPriorityTask.ID)
		assert.NoError(t, err)
	})

	t.Run("get started shows project context", func(t *testing.T) {
		// This test verifies that the action runs without error
		// and presumably shows project context information
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := GetStartedAction(appCtx)
		err := action(ctx)
		assert.NoError(t, err)
	})
}

func TestGetStartedErrorHandling(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	// Don't set a project context to test error handling
	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	t.Run("no project context", func(t *testing.T) {
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

		ctx := cli.NewContext(app, flagSet, nil)

		action := GetStartedAction(appCtx)
		err := action(ctx)
		assert.Error(t, err) // Should fail because no project is selected
	})
}
