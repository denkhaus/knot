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

func TestHierarchyCommands(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	commands := HierarchyCommands(appCtx)

	assert.NotEmpty(t, commands)
	
	// Check for expected hierarchy commands
	commandNames := make(map[string]bool)
	for _, cmd := range commands {
		commandNames[cmd.Name] = true
	}

	assert.True(t, commandNames["children"])
	assert.True(t, commandNames["parent"])
	assert.True(t, commandNames["roots"])
}

func TestChildrenAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Set project context
	err := mgr.SetSelectedProject(nil, project.ID, "test-user")
	require.NoError(t, err)

	t.Run("task with children", func(t *testing.T) {
		// Create a parent task
		parentTask, err := mgr.CreateTask(nil, project.ID, nil, "Parent Task", "A parent task", 4, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		// Create child tasks
		child1, err := mgr.CreateTask(nil, project.ID, &parentTask.ID, "Child 1", "First child", 2, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		child2, err := mgr.CreateTask(nil, project.ID, &parentTask.ID, "Child 2", "Second child", 3, types.TaskPriorityLow, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("task-id", "", "")
		flagSet.Bool("json", false, "")
		_ = flagSet.Set("task-id", parentTask.ID.String())

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := ChildrenAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify children still exist
		_, err = mgr.GetTask(nil, child1.ID)
		assert.NoError(t, err)
		_, err = mgr.GetTask(nil, child2.ID)
		assert.NoError(t, err)
	})
}

func TestRootsAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Set project context
	err := mgr.SetSelectedProject(nil, project.ID, "test-user")
	require.NoError(t, err)

	t.Run("project with root tasks", func(t *testing.T) {
		// Create root tasks
		root1, err := mgr.CreateTask(nil, project.ID, nil, "Root 1", "First root task", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		root2, err := mgr.CreateTask(nil, project.ID, nil, "Root 2", "Second root task", 4, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		// Create a child task (should not appear in roots)
		_, err = mgr.CreateTask(nil, project.ID, &root1.ID, "Child", "Child task", 2, types.TaskPriorityLow, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := RootsAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify root tasks still exist
		_, err = mgr.GetTask(nil, root1.ID)
		assert.NoError(t, err)
		_, err = mgr.GetTask(nil, root2.ID)
		assert.NoError(t, err)
	})
}