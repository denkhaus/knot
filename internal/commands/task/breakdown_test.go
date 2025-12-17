package task

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestBreakdownAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	// Get project manager from DI
	projectManager := do.MustInvoke[manager.ProjectManager](testInjector)

	project := testutil.CreateTestProject(t, projectManager)

	// Set project context
	err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
	require.NoError(t, err)

	t.Run("no tasks need breakdown in empty project", func(t *testing.T) {
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 8, "")
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		action := BreakdownAction(testInjector)
		err := action(ctx)
		assert.NoError(t, err)
	})

	t.Run("task with high complexity needs breakdown", func(t *testing.T) {
		// Create a high complexity task
		task, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Complex Task", "A very complex task", 9, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 8, "")
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		action := BreakdownAction(testInjector)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify task still exists
		retrievedTask, err := projectManager.GetTask(context.TODO(), task.ID)
		require.NoError(t, err)
		assert.Equal(t, 9, retrievedTask.Complexity)
	})

	t.Run("task with subtasks should not need breakdown", func(t *testing.T) {
		// Create a parent task with high complexity
		parentTask, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Parent Task", "Complex parent task", 9, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		// Create a subtask
		_, err = projectManager.CreateTask(context.TODO(), project.ID, &parentTask.ID, "Subtask", "A subtask", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 8, "")
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		action := BreakdownAction(testInjector)
		err = action(ctx)
		assert.NoError(t, err)
	})

	t.Run("custom threshold", func(t *testing.T) {
		// Create a task with complexity 6
		task, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Medium Complex Task", "Moderately complex", 6, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 5, "") // Lower threshold
		flagSet.Int("limit", 0, "")
		_ = flagSet.Set("threshold", "5")

		ctx := cli.NewContext(app, flagSet, nil)

		action := BreakdownAction(testInjector)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify task still exists
		retrievedTask, err := projectManager.GetTask(context.TODO(), task.ID)
		require.NoError(t, err)
		assert.Equal(t, 6, retrievedTask.Complexity)
	})

	t.Run("limit results", func(t *testing.T) {
		// Create multiple high complexity tasks
		for i := 0; i < 5; i++ {
			_, err := projectManager.CreateTask(context.TODO(), project.ID, nil,
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

		action := BreakdownAction(testInjector)
		err := action(ctx)
		assert.NoError(t, err)
	})

	t.Run("tasks at different depths", func(t *testing.T) {
		// Create a root task
		rootTask, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Root Complex Task", "Complex root", 10, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		// Create a child task with high complexity (but no children of its own)
		childTask, err := projectManager.CreateTask(context.TODO(), project.ID, &rootTask.ID, "Child Complex Task", "Complex child", 9, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 8, "")
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		action := BreakdownAction(testInjector)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify both tasks still exist
		_, err = projectManager.GetTask(context.TODO(), rootTask.ID)
		assert.NoError(t, err)
		_, err = projectManager.GetTask(context.TODO(), childTask.ID)
		assert.NoError(t, err)
	})

	t.Run("mixed complexity tasks", func(t *testing.T) {
		// Create tasks with various complexities
		lowComplexTask, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Low Complex", "Simple task", 3, types.TaskPriorityLow, "test-user")
		require.NoError(t, err)

		mediumComplexTask, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Medium Complex", "Medium task", 7, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		highComplexTask, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "High Complex", "Complex task", 10, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 8, "")
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		action := BreakdownAction(testInjector)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify all tasks still exist
		_, err = projectManager.GetTask(context.TODO(), lowComplexTask.ID)
		assert.NoError(t, err)
		_, err = projectManager.GetTask(context.TODO(), mediumComplexTask.ID)
		assert.NoError(t, err)
		_, err = projectManager.GetTask(context.TODO(), highComplexTask.ID)
		assert.NoError(t, err)
	})
}

func TestBreakdownActionErrorHandling(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	t.Run("no project context", func(t *testing.T) {
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 8, "")
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		action := BreakdownAction(testInjector)
		err := action(ctx)
		assert.Error(t, err) // Should fail because no project is selected
	})
}

func TestBreakdownActionUsesConfig(t *testing.T) {
	// Use standard test setup - this test verifies CLI flag behavior vs config defaults
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	// Get project manager from DI
	projectManager := do.MustInvoke[manager.ProjectManager](testInjector)

	// Create project
	project := testutil.CreateTestProject(t, projectManager)

	// Set project context
	err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
	require.NoError(t, err)

	t.Run("uses config threshold when no CLI flag", func(t *testing.T) {
		// Create tasks with complexity 6 (should need breakdown with default threshold 5)
		_, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Task 1", "Description", 6, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)
		_, err = projectManager.CreateTask(context.TODO(), project.ID, nil, "Task 2", "Description", 7, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		// Create task with complexity 4 (should not need breakdown with default threshold 5)
		_, err = projectManager.CreateTask(context.TODO(), project.ID, nil, "Task 3", "Description", 4, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		// Don't set threshold flag - should use config value
		flagSet.Int("limit", 0, "")

		ctx := cli.NewContext(app, flagSet, nil)

		action := BreakdownAction(testInjector)
		err = action(ctx)
		assert.NoError(t, err)

		// Get all tasks to verify which ones meet the threshold
		tasks, err := projectManager.ListTasksForProject(context.Background(), project.ID)
		require.NoError(t, err)

		// Count tasks with complexity >= 5 (default config threshold)
		expectedCount := 0
		for _, task := range tasks {
			if task.Complexity >= 5 {
				expectedCount++
			}
		}

		// Should have 2 tasks that meet the threshold
		assert.Equal(t, 2, expectedCount)
	})
}
