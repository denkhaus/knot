package task

import (
	"context"
	"flag"
	"os"
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

	// Create DI container that shares the same repository as the testInjector
	diContainer := di.NewContainer()

	// Create a minimal CLI context for testing
	app := &cli.App{}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("log-level", "info", "")
	flagSet.Int("complexity-threshold", 5, "")
	flagSet.Int("max-depth", 10, "")
	flagSet.Int("max-tasks-per-depth", 50, "")
	flagSet.Int("max-description-length", 500, "")
	flagSet.Bool("auto-reduce-complexity", true, "")
	cliCtx := cli.NewContext(app, flagSet, nil)

	// Register services
	injector := diContainer.RegisterAllServices(cliCtx)

	// Override repository with the same repository used by testInjector
	repo := do.MustInvoke[types.Repository](testInjector)
	do.Override(injector, func(do.Injector) (types.Repository, error) {
		return repo, nil
	})

	t.Run("list pending tasks", func(t *testing.T) {
		// Create tasks with different states
		pendingTask, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Pending Task", "A pending task", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		inProgressTask, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "In Progress Task", "An in-progress task", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)
		_, err = projectManager.UpdateTaskState(context.TODO(), inProgressTask.ID, types.TaskStateInProgress, "test-user")
		require.NoError(t, err)

		// Set up CLI context with DI container
		app := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("state", "", "")
		flagSet.Bool("json", false, "")
		_ = flagSet.Set("state", "pending")

		ctx := cli.NewContext(app, flagSet, nil)

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
		// Set up CLI context with DI container
		app := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("state", "", "")
		flagSet.Bool("json", false, "")
		_ = flagSet.Set("state", "invalid-state")

		ctx := cli.NewContext(app, flagSet, nil)

		action := ListByStateAction()
		err := action(ctx)
		// Note: Action may not error for invalid state, just return empty results
		assert.NoError(t, err)
	})

	t.Run("missing state parameter", func(t *testing.T) {
		// Set up CLI context with DI container
		app := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("state", "", "")
		flagSet.Bool("json", false, "")
		// Don't set state

		ctx := cli.NewContext(app, flagSet, nil)

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

	// Create DI container that shares the same repository as the testInjector
	diContainer := di.NewContainer()

	// Create a minimal CLI context for testing
	app := &cli.App{}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("log-level", "info", "")
	flagSet.Int("complexity-threshold", 5, "")
	flagSet.Int("max-depth", 10, "")
	flagSet.Int("max-tasks-per-depth", 50, "")
	flagSet.Int("max-description-length", 500, "")
	flagSet.Bool("auto-reduce-complexity", true, "")
	cliCtx := cli.NewContext(app, flagSet, nil)

	// Register services
	injector := diContainer.RegisterAllServices(cliCtx)

	// Override repository with the same repository used by testInjector
	repo := do.MustInvoke[types.Repository](testInjector)
	do.Override(injector, func(do.Injector) (types.Repository, error) {
		return repo, nil
	})

	t.Run("duplicate existing task", func(t *testing.T) {
		// Create a task to duplicate
		originalTask, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Original Task", "Original description", 4, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		// Set up CLI context with DI container
		app := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("task-id", "", "")
		flagSet.String("target-project-id", "", "")
		flagSet.String("actor", "", "")
		_ = flagSet.Set("task-id", originalTask.ID.String())
		_ = flagSet.Set("target-project-id", project.ID.String())
		_ = flagSet.Set("actor", "test-duplicator")

		ctx := cli.NewContext(app, flagSet, nil)

		action := DuplicateAction()
		err = action(ctx)
		assert.NoError(t, err)

		// Verify original task still exists
		_, err = projectManager.GetTask(context.TODO(), originalTask.ID)
		assert.NoError(t, err)
	})

	t.Run("missing task ID", func(t *testing.T) {
		// Set up CLI context with DI container
		app := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("task-id", "", "")
		flagSet.String("actor", "", "")
		// Don't set task-id
		_ = flagSet.Set("actor", "test-duplicator")

		ctx := cli.NewContext(app, flagSet, nil)

		action := DuplicateAction()
		err := action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task-id is required")
	})
}

func TestBulkCreateAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)
	projectManager := do.MustInvoke[manager.ProjectManager](testInjector)
	project := testutil.CreateTestProject(t, projectManager)

	err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
	require.NoError(t, err)

	// Create DI container that shares the same repository as the testInjector
	diContainer := di.NewContainer()

	app := &cli.App{}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("log-level", "info", "")
	flagSet.Int("complexity-threshold", 5, "")
	flagSet.Int("max-depth", 10, "")
	flagSet.Int("max-tasks-per-depth", 50, "")
	flagSet.Int("max-description-length", 500, "")
	flagSet.Bool("auto-reduce-complexity", true, "")
	cliCtx := cli.NewContext(app, flagSet, nil)

	injector := diContainer.RegisterAllServices(cliCtx)

	repo := do.MustInvoke[types.Repository](testInjector)
	do.Override(injector, func(do.Injector) (types.Repository, error) {
		return repo, nil
	})

	t.Run("missing file parameter", func(t *testing.T) {
		testApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		testFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		testFlagSet.String("file", "", "")
		// Don't set file

		ctx := cli.NewContext(testApp, testFlagSet, nil)

		action := BulkCreateAction()
		err := action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file is required")
	})

	t.Run("non-existent file", func(t *testing.T) {
		testApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		testFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		testFlagSet.String("file", "", "")
		_ = testFlagSet.Set("file", "/tmp/nonexistent-file-12345.json")

		ctx := cli.NewContext(testApp, testFlagSet, nil)

		action := BulkCreateAction()
		err := action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		// Create temp file with invalid JSON
		tmpfile, err := os.CreateTemp("", "invalid-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.WriteString("{invalid json}")
		require.NoError(t, err)
		tmpfile.Close()

		testApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		testFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		testFlagSet.String("file", "", "")
		_ = testFlagSet.Set("file", tmpfile.Name())

		ctx := cli.NewContext(testApp, testFlagSet, nil)

		action := BulkCreateAction()
		err = action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse JSON")
	})

	t.Run("empty tasks array", func(t *testing.T) {
		// Create temp file with empty array
		tmpfile, err := os.CreateTemp("", "empty-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.WriteString("[]")
		require.NoError(t, err)
		tmpfile.Close()

		testApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		testFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		testFlagSet.String("file", "", "")
		_ = testFlagSet.Set("file", tmpfile.Name())

		ctx := cli.NewContext(testApp, testFlagSet, nil)

		action := BulkCreateAction()
		err = action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no tasks found")
	})

	t.Run("task with missing title", func(t *testing.T) {
		// Create temp file with task missing title
		tmpfile, err := os.CreateTemp("", "no-title-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		content := `[{"description": "No title", "complexity": 3}]`
		_, err = tmpfile.WriteString(content)
		require.NoError(t, err)
		tmpfile.Close()

		testApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		testFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		testFlagSet.String("file", "", "")
		_ = testFlagSet.Set("file", tmpfile.Name())

		ctx := cli.NewContext(testApp, testFlagSet, nil)

		action := BulkCreateAction()
		err = action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})

	t.Run("task with invalid parent ID", func(t *testing.T) {
		// Create temp file with invalid parent ID
		tmpfile, err := os.CreateTemp("", "invalid-parent-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		content := `[{"title": "Task with invalid parent", "parent_id": "not-a-uuid"}]`
		_, err = tmpfile.WriteString(content)
		require.NoError(t, err)
		tmpfile.Close()

		testApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		testFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		testFlagSet.String("file", "", "")
		_ = testFlagSet.Set("file", tmpfile.Name())

		ctx := cli.NewContext(testApp, testFlagSet, nil)

		action := BulkCreateAction()
		err = action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid parent ID")
	})

	t.Run("successful bulk create", func(t *testing.T) {
		// Create temp file with valid tasks
		tmpfile, err := os.CreateTemp("", "valid-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		content := `[
			{"title": "Task 1", "description": "First task", "complexity": 3},
			{"title": "Task 2", "description": "Second task", "complexity": 5},
			{"title": "Task 3", "complexity": 2}
		]`
		_, err = tmpfile.WriteString(content)
		require.NoError(t, err)
		tmpfile.Close()

		testApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		testFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		testFlagSet.String("file", "", "")
		_ = testFlagSet.Set("file", tmpfile.Name())

		ctx := cli.NewContext(testApp, testFlagSet, nil)

		action := BulkCreateAction()
		err = action(ctx)
		assert.NoError(t, err)
	})
}

func TestBulkDeleteAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)
	projectManager := do.MustInvoke[manager.ProjectManager](testInjector)
	project := testutil.CreateTestProject(t, projectManager)

	err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
	require.NoError(t, err)

	// Create DI container that shares the same repository as the testInjector
	diContainer := di.NewContainer()

	app := &cli.App{}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("log-level", "info", "")
	flagSet.Int("complexity-threshold", 5, "")
	flagSet.Int("max-depth", 10, "")
	flagSet.Int("max-tasks-per-depth", 50, "")
	flagSet.Int("max-description-length", 500, "")
	flagSet.Bool("auto-reduce-complexity", true, "")
	cliCtx := cli.NewContext(app, flagSet, nil)

	injector := diContainer.RegisterAllServices(cliCtx)

	repo := do.MustInvoke[types.Repository](testInjector)
	do.Override(injector, func(do.Injector) (types.Repository, error) {
		return repo, nil
	})

	t.Run("missing task-ids parameter", func(t *testing.T) {
		testApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		testFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		testFlagSet.String("task-ids", "", "")
		testFlagSet.Bool("dry-run", false, "")
		testFlagSet.Bool("force", false, "")
		// Don't set task-ids

		ctx := cli.NewContext(testApp, testFlagSet, nil)

		action := BulkDeleteAction()
		err := action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task-ids are required")
	})

	t.Run("invalid task ID", func(t *testing.T) {
		testApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		testFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		testFlagSet.String("task-ids", "", "")
		testFlagSet.Bool("dry-run", false, "")
		testFlagSet.Bool("force", false, "")
		_ = testFlagSet.Set("task-ids", "not-a-uuid,another-invalid")

		ctx := cli.NewContext(testApp, testFlagSet, nil)

		action := BulkDeleteAction()
		err := action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task ID")
	})

	t.Run("dry-run mode", func(t *testing.T) {
		// Create some test tasks
		task1, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Task to Delete 1", "Will be deleted", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		task2, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Task to Delete 2", "Also will be deleted", 2, types.TaskPriorityLow, "test-user")
		require.NoError(t, err)

		testApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		testFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		testFlagSet.String("task-ids", "", "")
		testFlagSet.Bool("dry-run", false, "")
		testFlagSet.Bool("force", false, "")
		_ = testFlagSet.Set("task-ids", task1.ID.String()+","+task2.ID.String())
		_ = testFlagSet.Set("dry-run", "true")

		ctx := cli.NewContext(testApp, testFlagSet, nil)

		action := BulkDeleteAction()
		err = action(ctx)
		assert.NoError(t, err)

		// Verify tasks still exist (dry-run)
		_, err = projectManager.GetTask(context.TODO(), task1.ID)
		assert.NoError(t, err)
		_, err = projectManager.GetTask(context.TODO(), task2.ID)
		assert.NoError(t, err)
	})

	t.Run("force delete without confirmation", func(t *testing.T) {
		// Create test task
		task1, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Task to Force Delete", "Will be force deleted", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		testApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
		}
		testFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		testFlagSet.String("task-ids", "", "")
		testFlagSet.Bool("dry-run", false, "")
		testFlagSet.Bool("force", false, "")
		_ = testFlagSet.Set("task-ids", task1.ID.String())
		_ = testFlagSet.Set("force", "true")

		ctx := cli.NewContext(testApp, testFlagSet, nil)

		action := BulkDeleteAction()
		err = action(ctx)
		assert.NoError(t, err)

		// Verify task was deleted
		_, err = projectManager.GetTask(context.TODO(), task1.ID)
		assert.Error(t, err)
	})
}

