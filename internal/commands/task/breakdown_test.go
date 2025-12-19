package task

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/denkhaus/knot/v2/internal/di"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

// setupCLIContextWithDIBreakdown creates a CLI context with proper DI container setup for breakdown tests
func setupCLIContextWithDIBreakdown(t *testing.T, projectID string) (*cli.Context, *cli.App) {
	diContainer := di.NewContainer()

	// Create CLI context with proper flags for DI
	app := &cli.App{}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("log-level", "info", "")
	flagSet.Int("complexity-threshold", 5, "")
	flagSet.Int("max-depth", 10, "")
	flagSet.Int("max-tasks-per-depth", 50, "")
	flagSet.Int("max-description-length", 500, "")
	flagSet.Bool("auto-reduce-complexity", true, "")

	cliCtx := cli.NewContext(app, flagSet, nil)
	_ = diContainer.RegisterAllServices(cliCtx)

	// Store container in CLI context metadata like BeforeCommand does
	if app.Metadata == nil {
		app.Metadata = make(map[string]interface{})
	}
	app.Metadata["container"] = diContainer

	// Set project in the injector if projectID is provided
	if projectID != "" {
		injector := diContainer.GetInjector()
		projectManager := do.MustInvoke[manager.ProjectManager](injector)
		// Parse the project ID
		pid, err := uuid.Parse(projectID)
		require.NoError(t, err)
		err = projectManager.SetSelectedProject(cliCtx.Context, pid, "test-user")
		require.NoError(t, err)
	}

	return cliCtx, app
}

func TestBreakdownAction(t *testing.T) {
	var project *types.Project

	t.Run("no tasks need breakdown in empty project", func(t *testing.T) {
		cliCtx, _ := setupCLIContextWithDIBreakdown(t, "")

		// Create a project in the test's DI container
		diContainer := shared.GetContainerFromCLIContext(cliCtx)
		injector := diContainer.GetInjector()
		projectManager := do.MustInvoke[manager.ProjectManager](injector)

		project = testutil.CreateTestProject(t, projectManager)
		err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
		require.NoError(t, err)

		action := BreakdownAction()
		err = action(cliCtx)
		assert.NoError(t, err)
	})

	t.Run("task with high complexity needs breakdown", func(t *testing.T) {
		cliCtx, _ := setupCLIContextWithDIBreakdown(t, "")

		// Create a project in the test's DI container
		diContainer := shared.GetContainerFromCLIContext(cliCtx)
		injector := diContainer.GetInjector()
		projectManager := do.MustInvoke[manager.ProjectManager](injector)

		project = testutil.CreateTestProject(t, projectManager)
		err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
		require.NoError(t, err)

		// Create a high complexity task
		task, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Complex Task", "A very complex task", 9, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		action := BreakdownAction()
		err = action(cliCtx)
		assert.NoError(t, err)

		// Verify task still exists
		retrievedTask, err := projectManager.GetTask(context.TODO(), task.ID)
		require.NoError(t, err)
		assert.Equal(t, 9, retrievedTask.Complexity)
	})

	t.Run("task with subtasks should not need breakdown", func(t *testing.T) {
		cliCtx, _ := setupCLIContextWithDIBreakdown(t, "")

		// Create a project in the test's DI container
		diContainer := shared.GetContainerFromCLIContext(cliCtx)
		injector := diContainer.GetInjector()
		projectManager := do.MustInvoke[manager.ProjectManager](injector)

		project = testutil.CreateTestProject(t, projectManager)
		err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
		require.NoError(t, err)

		// Create a parent task with high complexity
		parentTask, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Parent Task", "Complex parent task", 9, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		// Create a subtask
		_, err = projectManager.CreateTask(context.TODO(), project.ID, &parentTask.ID, "Subtask", "A subtask", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		action := BreakdownAction()
		err = action(cliCtx)
		assert.NoError(t, err)
	})

	t.Run("custom threshold", func(t *testing.T) {
		cliCtx, app := setupCLIContextWithDIBreakdown(t, "")

		// Create a project in the test's DI container
		diContainer := shared.GetContainerFromCLIContext(cliCtx)
		injector := diContainer.GetInjector()
		projectManager := do.MustInvoke[manager.ProjectManager](injector)

		project = testutil.CreateTestProject(t, projectManager)
		err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
		require.NoError(t, err)

		// Create a task with complexity 6
		task, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Medium Complex Task", "Moderately complex", 6, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		// Set threshold flag
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 5, "Complexity threshold")
		flagSet.Int("limit", 0, "Result limit")
		flagSet.String("log-level", "info", "")
		flagSet.Int("complexity-threshold", 5, "")
		flagSet.Int("max-depth", 10, "")
		flagSet.Int("max-tasks-per-depth", 50, "")
		flagSet.Int("max-description-length", 500, "")
		flagSet.Bool("auto-reduce-complexity", true, "")
		_ = flagSet.Set("threshold", "5")

		ctx := cli.NewContext(app, flagSet, nil)

		action := BreakdownAction()
		err = action(ctx)
		assert.NoError(t, err)

		// Verify task still exists
		retrievedTask, err := projectManager.GetTask(context.TODO(), task.ID)
		require.NoError(t, err)
		assert.Equal(t, 6, retrievedTask.Complexity)
	})

	t.Run("limit results", func(t *testing.T) {
		cliCtx, app := setupCLIContextWithDIBreakdown(t, "")

		// Create a project in the test's DI container
		diContainer := shared.GetContainerFromCLIContext(cliCtx)
		injector := diContainer.GetInjector()
		projectManager := do.MustInvoke[manager.ProjectManager](injector)

		project = testutil.CreateTestProject(t, projectManager)
		err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
		require.NoError(t, err)

		// Create multiple high complexity tasks
		for i := 0; i < 5; i++ {
			_, err := projectManager.CreateTask(context.TODO(), project.ID, nil,
				fmt.Sprintf("Complex Task %d", i),
				fmt.Sprintf("Complex task number %d", i),
				9, types.TaskPriorityHigh, "test-user")
			require.NoError(t, err)
		}

		// Set limit flag
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 8, "Complexity threshold")
		flagSet.Int("limit", 2, "Result limit")
		flagSet.String("log-level", "info", "")
		flagSet.Int("complexity-threshold", 5, "")
		flagSet.Int("max-depth", 10, "")
		flagSet.Int("max-tasks-per-depth", 50, "")
		flagSet.Int("max-description-length", 500, "")
		flagSet.Bool("auto-reduce-complexity", true, "")
		_ = flagSet.Set("limit", "2")

		ctx := cli.NewContext(app, flagSet, nil)

		action := BreakdownAction()
		err = action(ctx)
		assert.NoError(t, err)
	})
}

func TestBreakdownActionErrorHandling(t *testing.T) {

	t.Run("no project context", func(t *testing.T) {
		// Set up DI container for this test
		diContainer := di.NewContainer()

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Int("threshold", 8, "")
		flagSet.Int("limit", 0, "")
		// Add DI flags
		flagSet.String("log-level", "info", "")
		flagSet.Int("complexity-threshold", 5, "")
		flagSet.Int("max-depth", 10, "")
		flagSet.Int("max-tasks-per-depth", 50, "")
		flagSet.Int("max-description-length", 500, "")
		flagSet.Bool("auto-reduce-complexity", true, "")

		ctx := cli.NewContext(app, flagSet, nil)
		_ = diContainer.RegisterAllServices(ctx)

		// Store container in CLI context metadata like BeforeCommand does
		if app.Metadata == nil {
			app.Metadata = make(map[string]interface{})
		}
		app.Metadata["container"] = diContainer

		action := BreakdownAction()
		err := action(ctx)
		assert.Error(t, err) // Should fail because no project is selected
	})
}