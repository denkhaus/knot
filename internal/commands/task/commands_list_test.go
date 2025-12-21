package task

import (
	"context"
	"flag"
	"testing"

	"github.com/denkhaus/knot/v2/internal/di"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

// TestSimplifiedTaskListFlags tests the simplified list command flags
// Knot Task 75010287-8330-4001-8f31-a824ce6c5d09: Simplified task list flags
func TestSimplifiedTaskListFlags(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	// Get project manager from DI
	projectManager := do.MustInvoke[manager.ProjectManager](testInjector)

	// Create and configure DI container for test
	diContainer := di.NewContainer()

	// Create a minimal CLI context for testing
	app := &cli.App{
		Metadata: map[string]interface{}{
			"container": diContainer,
		},
	}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("log-level", "info", "")
	flagSet.Int("complexity-threshold", 5, "")
	flagSet.Int("max-depth", 10, "")
	flagSet.Int("max-tasks-per-depth", 50, "")
	flagSet.Int("max-description-length", 500, "")
	flagSet.Bool("auto-reduce-complexity", true, "")
	cliCtx := cli.NewContext(app, flagSet, nil)

	// Register services in container
	containerInjector := diContainer.RegisterAllServices(cliCtx)

	// Override repository with test repository (in-memory) from the testInjector
	testRepo := do.MustInvoke[types.Repository](testInjector)
	do.Override(containerInjector, func(do.Injector) (types.Repository, error) {
		return testRepo, nil
	})

	// Override project manager with the one from testInjector
	do.Override(containerInjector, func(do.Injector) (manager.ProjectManager, error) {
		return projectManager, nil
	})

	project := testutil.CreateTestProject(t, projectManager)

	// Set selected project context
	err := projectManager.SetSelectedProject(context.Background(), project.ID, "test-user")
	require.NoError(t, err)

	// Create test tasks with different properties
	tasks := []*types.Task{
		{Title: "Low complexity task", Description: "Simple task", Complexity: 2, Priority: types.TaskPriorityLow, State: types.TaskStatePending},
		{Title: "Medium complexity task", Description: "Moderate task", Complexity: 5, Priority: types.TaskPriorityMedium, State: types.TaskStateInProgress},
		{Title: "High complexity task", Description: "Complex task", Complexity: 8, Priority: types.TaskPriorityHigh, State: types.TaskStateCompleted},
	}

	// Create the tasks
	for _, task := range tasks {
		_, err := projectManager.CreateTask(context.Background(), project.ID, nil, task.Title, task.Description, task.Complexity, task.Priority, "test-user")
		require.NoError(t, err)
	}

	t.Run("default limit is 20", func(t *testing.T) {
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Bool("quiet", false, "")
		flagSet.String("state", "", "")
		flagSet.String("priority", "", "")
		flagSet.Int("complexity", 0, "")
		flagSet.String("search", "", "")
		flagSet.Int("limit", 20, "")

		ctx := cli.NewContext(app, flagSet, nil)

		action := listAction()
		err = action(ctx)
		assert.NoError(t, err)
	})

	t.Run("complexity filter as minimum", func(t *testing.T) {
		// Use app from outer scope with DI container metadata
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Bool("quiet", false, "")
		flagSet.String("state", "", "")
		flagSet.String("priority", "", "")
		flagSet.Int("complexity", 5, "") // Should show tasks with complexity >= 5
		flagSet.String("search", "", "")
		flagSet.Int("limit", 20, "")

		ctx := cli.NewContext(app, flagSet, nil)

		action := listAction()
		err = action(ctx)
		assert.NoError(t, err)
	})

	t.Run("state filter", func(t *testing.T) {
		// Use app from outer scope with DI container metadata
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Bool("quiet", false, "")
		flagSet.String("state", "completed", "")
		flagSet.String("priority", "", "")
		flagSet.Int("complexity", 0, "")
		flagSet.String("search", "", "")
		flagSet.Int("limit", 20, "")

		ctx := cli.NewContext(app, flagSet, nil)

		action := listAction()
		err = action(ctx)
		assert.NoError(t, err)
	})

	t.Run("priority filter", func(t *testing.T) {
		// Use app from outer scope with DI container metadata
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Bool("quiet", false, "")
		flagSet.String("state", "", "")
		flagSet.String("priority", "high", "")
		flagSet.Int("complexity", 0, "")
		flagSet.String("search", "", "")
		flagSet.Int("limit", 20, "")

		ctx := cli.NewContext(app, flagSet, nil)

		action := listAction()
		err = action(ctx)
		assert.NoError(t, err)
	})

	t.Run("search filter", func(t *testing.T) {
		// Use app from outer scope with DI container metadata
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Bool("quiet", false, "")
		flagSet.String("state", "", "")
		flagSet.String("priority", "", "")
		flagSet.Int("complexity", 0, "")
		flagSet.String("search", "Complex", "")
		flagSet.Int("limit", 20, "")

		ctx := cli.NewContext(app, flagSet, nil)

		action := listAction()
		err = action(ctx)
		assert.NoError(t, err)
	})

	t.Run("custom limit", func(t *testing.T) {
		// Use app from outer scope with DI container metadata
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")
		flagSet.Bool("quiet", false, "")
		flagSet.String("state", "", "")
		flagSet.String("priority", "", "")
		flagSet.Int("complexity", 0, "")
		flagSet.String("search", "", "")
		flagSet.Int("limit", 2, "")

		ctx := cli.NewContext(app, flagSet, nil)

		action := listAction()
		err = action(ctx)
		assert.NoError(t, err)
	})
}
