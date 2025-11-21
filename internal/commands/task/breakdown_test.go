package task

import (
	"flag"
	"fmt"
	"testing"

	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestBreakdownAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Set project context
	err := mgr.SetSelectedProject(nil, project.ID, "test-user")
	require.NoError(t, err)

	t.Run("no tasks need breakdown in empty project", func(t *testing.T) {
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 8, "")
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := BreakdownAction(appCtx)
		err := action(ctx)
		assert.NoError(t, err)
	})

	t.Run("task with high complexity needs breakdown", func(t *testing.T) {
		// Create a high complexity task
		task, err := mgr.CreateTask(nil, project.ID, nil, "Complex Task", "A very complex task", 9, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 8, "")
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := BreakdownAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify task still exists
		retrievedTask, err := mgr.GetTask(nil, task.ID)
		require.NoError(t, err)
		assert.Equal(t, 9, retrievedTask.Complexity)
	})

	t.Run("task with subtasks should not need breakdown", func(t *testing.T) {
		// Create a parent task with high complexity
		parentTask, err := mgr.CreateTask(nil, project.ID, nil, "Parent Task", "Complex parent task", 9, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		// Create a subtask
		_, err = mgr.CreateTask(nil, project.ID, &parentTask.ID, "Subtask", "A subtask", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 8, "")
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := BreakdownAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)
	})

	t.Run("custom threshold", func(t *testing.T) {
		// Create a task with complexity 6
		task, err := mgr.CreateTask(nil, project.ID, nil, "Medium Complex Task", "Moderately complex", 6, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 5, "") // Lower threshold
		flagSet.Int("limit", 0, "")
		_ = flagSet.Set("threshold", "5")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := BreakdownAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify task still exists
		retrievedTask, err := mgr.GetTask(nil, task.ID)
		require.NoError(t, err)
		assert.Equal(t, 6, retrievedTask.Complexity)
	})

	t.Run("limit results", func(t *testing.T) {
		// Create multiple high complexity tasks
		for i := 0; i < 5; i++ {
			_, err := mgr.CreateTask(nil, project.ID, nil, 
				fmt.Sprintf("Complex Task %d", i), 
				fmt.Sprintf("Complex task number %d", i), 
				9, types.TaskPriorityHigh, "test-user")
			require.NoError(t, err)
		}

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 8, "")
		flagSet.Int("limit", 2, "")
		_ = flagSet.Set("limit", "2")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := BreakdownAction(appCtx)
		err := action(ctx)
		assert.NoError(t, err)
	})

	t.Run("tasks at different depths", func(t *testing.T) {
		// Create a root task
		rootTask, err := mgr.CreateTask(nil, project.ID, nil, "Root Complex Task", "Complex root", 10, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		// Create a child task with high complexity (but no children of its own)
		childTask, err := mgr.CreateTask(nil, project.ID, &rootTask.ID, "Child Complex Task", "Complex child", 9, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 8, "")
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := BreakdownAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify both tasks still exist
		_, err = mgr.GetTask(nil, rootTask.ID)
		assert.NoError(t, err)
		_, err = mgr.GetTask(nil, childTask.ID)
		assert.NoError(t, err)
	})

	t.Run("mixed complexity tasks", func(t *testing.T) {
		// Create tasks with various complexities
		lowComplexTask, err := mgr.CreateTask(nil, project.ID, nil, "Low Complex", "Simple task", 3, types.TaskPriorityLow, "test-user")
		require.NoError(t, err)

		mediumComplexTask, err := mgr.CreateTask(nil, project.ID, nil, "Medium Complex", "Medium task", 7, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		highComplexTask, err := mgr.CreateTask(nil, project.ID, nil, "High Complex", "Complex task", 10, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 8, "")
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := BreakdownAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify all tasks still exist
		_, err = mgr.GetTask(nil, lowComplexTask.ID)
		assert.NoError(t, err)
		_, err = mgr.GetTask(nil, mediumComplexTask.ID)
		assert.NoError(t, err)
		_, err = mgr.GetTask(nil, highComplexTask.ID)
		assert.NoError(t, err)
	})
}

func TestBreakdownActionErrorHandling(t *testing.T) {
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
		flagSet.Int("threshold", 8, "")
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		action := BreakdownAction(appCtx)
		err := action(ctx)
		assert.Error(t, err) // Should fail because no project is selected
	})
}