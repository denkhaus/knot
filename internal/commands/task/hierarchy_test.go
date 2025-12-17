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

func TestHierarchyCommands(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	commands := HierarchyCommands(testInjector)

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
	testInjector := config.SetupTestInjector(t)

	// Get project manager from DI
	projectManager := do.MustInvoke[manager.ProjectManager](testInjector)
	project := testutil.CreateTestProject(t, projectManager)

	// Set project context
	err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
	require.NoError(t, err)

	t.Run("task with children", func(t *testing.T) {
		// Create a parent task
		parentTask, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Parent Task", "A parent task", 4, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		// Create child tasks
		child1, err := projectManager.CreateTask(context.TODO(), project.ID, &parentTask.ID, "Child 1", "First child", 2, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		child2, err := projectManager.CreateTask(context.TODO(), project.ID, &parentTask.ID, "Child 2", "Second child", 3, types.TaskPriorityLow, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("task-id", "", "")
		flagSet.Bool("json", false, "")
		_ = flagSet.Set("task-id", parentTask.ID.String())

		ctx := cli.NewContext(app, flagSet, nil)

		// Use testInjector instead of AppContext

		action := ChildrenAction(testInjector)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify children still exist
		_, err = projectManager.GetTask(context.TODO(), child1.ID)
		assert.NoError(t, err)
		_, err = projectManager.GetTask(context.TODO(), child2.ID)
		assert.NoError(t, err)
	})
}

func TestRootsAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	// Get project manager from DI
	projectManager := do.MustInvoke[manager.ProjectManager](testInjector)
	project := testutil.CreateTestProject(t, projectManager)

	// Set project context
	err := projectManager.SetSelectedProject(context.TODO(), project.ID, "test-user")
	require.NoError(t, err)

	t.Run("project with root tasks", func(t *testing.T) {
		// Create root tasks
		root1, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Root 1", "First root task", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		root2, err := projectManager.CreateTask(context.TODO(), project.ID, nil, "Root 2", "Second root task", 4, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		// Create a child task (should not appear in roots)
		_, err = projectManager.CreateTask(context.TODO(), project.ID, &root1.ID, "Child", "Child task", 2, types.TaskPriorityLow, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.Bool("json", false, "")

		ctx := cli.NewContext(app, flagSet, nil)

		// Use testInjector instead of AppContext

		action := RootsAction(testInjector)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify root tasks still exist
		_, err = projectManager.GetTask(context.TODO(), root1.ID)
		assert.NoError(t, err)
		_, err = projectManager.GetTask(context.TODO(), root2.ID)
		assert.NoError(t, err)
	})
}
