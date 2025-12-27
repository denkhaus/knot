package task

import (
	"context"
	"flag"
	"testing"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestReadyAction(t *testing.T) {
	t.Run("pending task is ready", func(t *testing.T) {
		cliCtx, _ := setupCLIContextWithDI(t, "")

		// Create a project in the test's DI container
		diContainer := shared.GetContainerFromContext(cliCtx)
		injector := diContainer.GetInjector()
		projectManager := do.MustInvoke[manager.ProjectManager](injector)

		project := testutil.CreateTestProject(t, projectManager)
		err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
		require.NoError(t, err)

		// Create a pending task
		task, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Ready Task", "A task ready to work on", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		// Create a new flag set for the ReadyAction
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Int("limit", 0, "")
		flagSet.String("log-level", "info", "")
		flagSet.Int("complexity-threshold", 5, "")
		flagSet.Int("max-depth", 10, "")
		flagSet.Int("max-tasks-per-depth", 50, "")
		flagSet.Int("max-description-length", 500, "")
		flagSet.Bool("auto-reduce-complexity", true, "")

		ctx := cli.NewContext(cliCtx.App, flagSet, nil)

		action := ReadyAction()
		err = action(ctx)
		assert.NoError(t, err)

		// Verify task is still there
		retrievedTask, err := projectManager.GetTask(context.TODO(), task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.Title, retrievedTask.Title)
		assert.Equal(t, types.TaskStatePending, retrievedTask.State)
	})

	t.Run("in-progress task is ready", func(t *testing.T) {
		cliCtx, _ := setupCLIContextWithDI(t, "")

		// Create a project in the test's DI container
		diContainer := shared.GetContainerFromContext(cliCtx)
		injector := diContainer.GetInjector()
		projectManager := do.MustInvoke[manager.ProjectManager](injector)

		project := testutil.CreateTestProject(t, projectManager)
		err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
		require.NoError(t, err)

		// Create an in-progress task
		task, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "In Progress Task", "A task in progress", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		// Update to in-progress
		_, err = projectManager.UpdateTaskState(context.TODO(), task.ID, types.TaskStateInProgress, "test-user")
		require.NoError(t, err)

		// Create a new flag set for the ReadyAction
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Int("limit", 0, "")
		flagSet.String("log-level", "info", "")
		flagSet.Int("complexity-threshold", 5, "")
		flagSet.Int("max-depth", 10, "")
		flagSet.Int("max-tasks-per-depth", 50, "")
		flagSet.Int("max-description-length", 500, "")
		flagSet.Bool("auto-reduce-complexity", true, "")

		ctx := cli.NewContext(cliCtx.App, flagSet, nil)

		action := ReadyAction()
		err = action(ctx)
		assert.NoError(t, err)

		// Verify task is still there
		retrievedTask, err := projectManager.GetTask(context.TODO(), task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.Title, retrievedTask.Title)
		assert.Equal(t, types.TaskStateInProgress, retrievedTask.State)
	})

	t.Run("completed task is not ready", func(t *testing.T) {
		cliCtx, _ := setupCLIContextWithDI(t, "")

		// Create a project in the test's DI container
		diContainer := shared.GetContainerFromContext(cliCtx)
		injector := diContainer.GetInjector()
		projectManager := do.MustInvoke[manager.ProjectManager](injector)

		project := testutil.CreateTestProject(t, projectManager)
		err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
		require.NoError(t, err)

		// Create and complete a task
		task, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Completed Task", "A completed task", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		// First transition to in-progress, then to completed (following valid transition rules)
		_, err = projectManager.UpdateTaskState(context.TODO(), task.ID, types.TaskStateInProgress, "test-user")
		require.NoError(t, err)
		_, err = projectManager.UpdateTaskState(context.TODO(), task.ID, types.TaskStateCompleted, "test-user")
		require.NoError(t, err)

		// Create a new flag set for the ReadyAction
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Int("limit", 0, "")
		flagSet.String("log-level", "info", "")
		flagSet.Int("complexity-threshold", 5, "")
		flagSet.Int("max-depth", 10, "")
		flagSet.Int("max-tasks-per-depth", 50, "")
		flagSet.Int("max-description-length", 500, "")
		flagSet.Bool("auto-reduce-complexity", true, "")

		ctx := cli.NewContext(cliCtx.App, flagSet, nil)

		action := ReadyAction()
		err = action(ctx)
		assert.NoError(t, err)
	})
}
