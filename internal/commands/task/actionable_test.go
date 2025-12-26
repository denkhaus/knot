package task

import (
	"context"
	"flag"
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

// setupCLIContextWithDI creates a CLI context with proper DI container setup
func setupCLIContextWithDI(t *testing.T, projectID string) (*cli.Context, *cli.App) {
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

func TestActionableAction(t *testing.T) {
	var project *types.Project
	var app *cli.App

	t.Run("no actionable tasks in empty project", func(t *testing.T) {
		cliCtx, _ := setupCLIContextWithDI(t, "")

		// Create a project in the test's DI container
		diContainer := shared.GetContainerFromContext(cliCtx)
		injector := diContainer.GetInjector()
		projectManager := do.MustInvoke[manager.ProjectManager](injector)

		project := testutil.CreateTestProject(t, projectManager)
		err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
		require.NoError(t, err)

		action := ActionableAction()
		err = action(cliCtx)
		assert.NoError(t, err)
	})

	t.Run("single actionable task", func(t *testing.T) {
		cliCtx, _ := setupCLIContextWithDI(t, "")

		// Create a project in the test's DI container
		diContainer := shared.GetContainerFromContext(cliCtx)
		injector := diContainer.GetInjector()
		projectManager := do.MustInvoke[manager.ProjectManager](injector)

		project := testutil.CreateTestProject(t, projectManager)
		err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
		require.NoError(t, err)

		// Create a pending task
		task, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Actionable Task", "A task to work on", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		action := ActionableAction()
		err = action(cliCtx)
		assert.NoError(t, err)

		// Verify task is still there
		retrievedTask, err := projectManager.GetTask(context.TODO(), task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.Title, retrievedTask.Title)
	})

	t.Run("actionable with multiple tasks", func(t *testing.T) {
		cliCtx, _ := setupCLIContextWithDI(t, "")

		// Create a project in the test's DI container
		diContainer := shared.GetContainerFromContext(cliCtx)
		injector := diContainer.GetInjector()
		projectManager := do.MustInvoke[manager.ProjectManager](injector)

		project := testutil.CreateTestProject(t, projectManager)
		err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
		require.NoError(t, err)

		// Create multiple tasks
		_, err = projectManager.CreateTask(context.TODO(), project.ID, nil, "High Priority Task", "Important task", 2, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		_, err = projectManager.CreateTask(context.TODO(), project.ID, nil, "Low Priority Task", "Less important task", 1, types.TaskPriorityLow, "test-user")
		require.NoError(t, err)

		action := ActionableAction()
		err = action(cliCtx)
		assert.NoError(t, err)
	})

	t.Run("actionable with JSON output", func(t *testing.T) {
		cliCtx, _ := setupCLIContextWithDI(t, "")

		// Create a project in the test's DI container
		diContainer := shared.GetContainerFromContext(cliCtx)
		injector := diContainer.GetInjector()
		projectManager := do.MustInvoke[manager.ProjectManager](injector)

		project = testutil.CreateTestProject(t, projectManager)
		err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
		require.NoError(t, err)

		_, app = setupCLIContextWithDI(t, project.ID.String())

		// Create a new flag set with JSON flag
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", true, "")
		_ = flagSet.Set("json", "true")

		ctx := cli.NewContext(app, flagSet, nil)

		action := ActionableAction()
		err = action(ctx)
		assert.NoError(t, err)
	})
}
