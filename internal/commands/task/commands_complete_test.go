package task

import (
	"context"
	"flag"
	"testing"

	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestTaskCreateAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	tests := []struct {
		name        string
		title       string
		description string
		complexity  string
		parentID    string
		actor       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid task creation",
			title:       "Test Task",
			description: "Test Description",
			complexity:  "5",
			actor:       "test-user",
			expectError: false,
		},
		{
			name:        "task with parent",
			title:       "Subtask",
			description: "Subtask Description",
			complexity:  "3",
			actor:       "test-user",
			expectError: false,
		},
		{
			name:        "empty title should fail",
			title:       "",
			description: "Test Description",
			complexity:  "5",
			actor:       "test-user",
			expectError: true,
			errorMsg:    "title cannot be empty",
		},
		{
			name:        "invalid complexity should fail",
			title:       "Test Task",
			description: "Test Description",
			complexity:  "15",
			actor:       "test-user",
			expectError: true,
			errorMsg:    "complexity must be between 1 and 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CLI context
			app := &cli.App{}
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			flagSet.String("title", "", "")
			flagSet.String("description", "", "")
			flagSet.String("complexity", "", "")
			flagSet.String("priority", "", "")
			flagSet.String("parent-id", "", "")
			flagSet.String("actor", "", "")

			_ = flagSet.Set("title", tt.title)
			_ = flagSet.Set("description", tt.description)
			_ = flagSet.Set("complexity", tt.complexity)
			_ = flagSet.Set("priority", "medium")
			if tt.parentID != "" {
				_ = flagSet.Set("parent-id", tt.parentID)
			}
			_ = flagSet.Set("actor", tt.actor)

			ctx := cli.NewContext(app, flagSet, nil)

			// Set project context for the test
			err := mgr.SetSelectedProject(ctx.Context, project.ID, "test-user")
			require.NoError(t, err)

			// Execute action
			action := createAction(appCtx)
			err = action(ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskListAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	t.Run("empty task list", func(t *testing.T) {
		// Create CLI context
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

		ctx := cli.NewContext(app, flagSet, nil)

		// Set project context for the test
		err := mgr.SetSelectedProject(ctx.Context, project.ID, "test-user")
		require.NoError(t, err)

		// Execute action
		action := listAction(appCtx)
		err = action(ctx)

		// Should succeed but show no tasks
		require.NoError(t, err)
	})

	t.Run("list with tasks", func(t *testing.T) {
		// Create test tasks
		testutil.CreateTestTask(t, mgr, project.ID)
		testutil.CreateTestTask(t, mgr, project.ID)

		// Create CLI context
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

		ctx := cli.NewContext(app, flagSet, nil)

		// Set project context for the test
		err := mgr.SetSelectedProject(ctx.Context, project.ID, "test-user")
		require.NoError(t, err)

		// Execute action
		action := listAction(appCtx)
		err = action(ctx)

		// Should succeed
		assert.NoError(t, err)
	})
}

func TestTaskUpdateStateAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)
	task := testutil.CreateTestTask(t, mgr, project.ID)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	// Set selected project context for the CLI
	err := mgr.SetSelectedProject(context.Background(), project.ID, "test-user")
	require.NoError(t, err)

	tests := []struct {
		name        string
		taskID      string
		state       string
		actor       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid state transition",
			taskID:      task.ID.String(),
			state:       "in-progress",
			actor:       "test-user",
			expectError: false,
		},
		{
			name:        "invalid task ID",
			taskID:      "invalid-uuid",
			state:       "in-progress",
			actor:       "test-user",
			expectError: true,
			errorMsg:    "invalid task-id format",
		},
		{
			name:        "non-existent task",
			taskID:      "123e4567-e89b-12d3-a456-426614174000",
			state:       "in-progress",
			actor:       "test-user",
			expectError: true,
			errorMsg:    "task not found",
		},
		{
			name:        "invalid state",
			taskID:      task.ID.String(),
			state:       "invalid-state",
			actor:       "test-user",
			expectError: true,
			errorMsg:    "invalid task state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CLI context
			app := &cli.App{}
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			flagSet.String("id", "", "")
			flagSet.String("state", "", "")
			flagSet.String("actor", "", "")

			_ = flagSet.Set("id", tt.taskID)
			_ = flagSet.Set("state", tt.state)
			_ = flagSet.Set("actor", tt.actor)

			ctx := cli.NewContext(app, flagSet, nil)

			// Execute action
			action := updateStateAction(appCtx)
			err := action(ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)

				// Verify state was actually updated
				testutil.AssertTaskState(t, mgr, task.ID, types.TaskState(tt.state))
			}
		})
	}
}

func TestTaskUpdateTitleAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)
	task := testutil.CreateTestTask(t, mgr, project.ID)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	tests := []struct {
		name        string
		taskID      string
		title       string
		actor       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid title update",
			taskID:      task.ID.String(),
			title:       "Updated Task Title",
			actor:       "test-user",
			expectError: false,
		},
		{
			name:        "empty title should fail",
			taskID:      task.ID.String(),
			title:       "",
			actor:       "test-user",
			expectError: true,
			errorMsg:    "title cannot be empty",
		},
		{
			name:        "invalid task ID",
			taskID:      "invalid-uuid",
			title:       "Updated Title",
			actor:       "test-user",
			expectError: true,
			errorMsg:    "invalid task-id format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CLI context
			app := &cli.App{}
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			flagSet.String("id", "", "")
			flagSet.String("title", "", "")
			flagSet.String("actor", "", "")

			_ = flagSet.Set("id", tt.taskID)
			_ = flagSet.Set("title", tt.title)
			_ = flagSet.Set("actor", tt.actor)

			ctx := cli.NewContext(app, flagSet, nil)

			// Execute action
			action := updateTitleAction(appCtx)
			err := action(ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskWorkflow(t *testing.T) {
	// Integration test for complete task workflow
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	t.Run("complete task workflow", func(t *testing.T) {
		// 1. Create task
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("title", "", "")
		flagSet.String("description", "", "")
		flagSet.String("complexity", "", "")
		flagSet.String("priority", "", "")
		flagSet.String("actor", "", "")

		_ = flagSet.Set("title", "Workflow Test Task")
		_ = flagSet.Set("description", "Created during workflow test")
		_ = flagSet.Set("complexity", "5")
		_ = flagSet.Set("priority", "medium")
		_ = flagSet.Set("actor", "workflow-test")

		ctx := cli.NewContext(app, flagSet, nil)

		// Set project context for the test
		err := mgr.SetSelectedProject(ctx.Context, project.ID, "test-user")
		require.NoError(t, err)

		createActionFunc := createAction(appCtx)
		err = createActionFunc(ctx)
		require.NoError(t, err)

		// 2. List tasks (should now have at least one)
		listFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)

		listCtx := cli.NewContext(app, listFlagSet, nil)

		listActionFunc := listAction(appCtx)
		err = listActionFunc(listCtx)
		assert.NoError(t, err)

		// 3. Get tasks and update state
		tasks, err := mgr.ListTasksForProject(ctx.Context, project.ID)
		require.NoError(t, err)
		require.NotEmpty(t, tasks)

		updateFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		updateFlagSet.String("id", "", "")
		updateFlagSet.String("state", "", "")
		updateFlagSet.String("actor", "", "")
		_ = updateFlagSet.Set("id", tasks[0].ID.String())
		_ = updateFlagSet.Set("state", "in-progress")
		_ = updateFlagSet.Set("actor", "workflow-test")

		updateCtx := cli.NewContext(app, updateFlagSet, nil)

		updateActionFunc := updateStateAction(appCtx)
		err = updateActionFunc(updateCtx)
		assert.NoError(t, err)

		// 4. Verify state was updated
		testutil.AssertTaskState(t, mgr, tasks[0].ID, types.TaskStateInProgress)
	})
}
