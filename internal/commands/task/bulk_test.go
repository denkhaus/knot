package task

import (
	"context"
	"flag"
	"testing"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestBulkCommands(t *testing.T) {
	commands := BulkCommands()

	assert.NotEmpty(t, commands)

	// Check for expected bulk commands
	commandNames := make(map[string]bool)
	for _, cmd := range commands {
		commandNames[cmd.Name] = true
	}

	assert.True(t, commandNames["duplicate"])
	assert.True(t, commandNames["list-by-state"])
	assert.True(t, commandNames["bulk-create"])
	assert.True(t, commandNames["bulk-delete"])

	// Ensure bulk-update is not present
	assert.False(t, commandNames["bulk-update"], "bulk-update command should have been removed")
}

func TestListByStateAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
testInjector := config.SetupTestInjector(t)
	// Get project manager from DI
	projectManager := do.MustInvoke[manager.ProjectManager](testInjector)
	project := testutil.CreateTestProject(t, projectManager)

	// Set project context
	err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
	require.NoError(t, err)

	t.Run("list pending tasks", func(t *testing.T) {
		// Create tasks with different states
		pendingTask, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Pending Task", "A pending task", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		inProgressTask, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "In Progress Task", "An in-progress task", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)
		_, err = projectManager.UpdateTaskState(context.TODO(), inProgressTask.ID, types.TaskStateInProgress, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("state", "", "")
		flagSet.Bool("json", false, "")
		_ = flagSet.Set("state", "pending")

		ctx := cli.NewContext(app, flagSet, nil)

		// Use testInjector instead of AppContext

		action := ListByStateAction()
		err = action(ctx)
		assert.NoError(t, err)

		// Verify tasks still exist
		_, err = projectManager.GetTask(context.TODO(), pendingTask.ID)
		assert.NoError(t, err)
		_, err = projectManager.GetTask(context.TODO(), inProgressTask.ID)
		assert.NoError(t, err)
	})

	t.Run("invalid state", func(t *testing.T) {
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("state", "", "")
		flagSet.Bool("json", false, "")
		_ = flagSet.Set("state", "invalid-state")

		ctx := cli.NewContext(app, flagSet, nil)

		// Use testInjector instead of AppContext

		action := ListByStateAction()
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

		// Use testInjector instead of AppContext

		action := ListByStateAction()
		err := action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state is required")
	})
}

func TestDuplicateAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
testInjector := config.SetupTestInjector(t)
	// Get project manager from DI
	projectManager := do.MustInvoke[manager.ProjectManager](testInjector)
	project := testutil.CreateTestProject(t, projectManager)

	// Set project context
	err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
	require.NoError(t, err)

	t.Run("duplicate existing task", func(t *testing.T) {
		// Create a task to duplicate
		originalTask, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Original Task", "Original description", 4, types.TaskPriorityHigh, "test-user")
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

		// Use testInjector instead of AppContext

		action := DuplicateAction()
		err = action(ctx)
		assert.NoError(t, err)

		// Verify original task still exists
		_, err = projectManager.GetTask(context.TODO(), originalTask.ID)
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

		// Use testInjector instead of AppContext

		action := DuplicateAction()
		err := action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task-id is required")
	})
}
