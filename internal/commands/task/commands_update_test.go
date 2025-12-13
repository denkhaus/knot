package task

import (
	"context"
	"flag"
	"testing"

	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestUpdateDescriptionAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Set project context
	err := mgr.SetSelectedProject(context.TODO(), project.ID, "test-user")
	require.NoError(t, err)

	t.Run("successful description update", func(t *testing.T) {
		// Create a task
		task, err := mgr.CreateTask(context.TODO(), project.ID, nil, "Test Task", "Original description", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("id", "", "")
		flagSet.String("description", "", "")
		flagSet.String("actor", "", "")

		_ = flagSet.Set("id", task.ID.String())
		_ = flagSet.Set("description", "Updated description")
		_ = flagSet.Set("actor", "test-updater")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := updateDescriptionSubAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify the update
		updatedTask, err := mgr.GetTask(context.TODO(), task.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated description", updatedTask.Description)
	})

	t.Run("invalid task ID", func(t *testing.T) {
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("id", "", "")
		flagSet.String("description", "", "")
		flagSet.String("actor", "", "")

		_ = flagSet.Set("id", "invalid-uuid")
		_ = flagSet.Set("description", "New description")
		_ = flagSet.Set("actor", "test-updater")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := updateDescriptionSubAction(appCtx)
		err := action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task-id format")
	})
}

func TestUpdatePriorityAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Set project context
	err := mgr.SetSelectedProject(context.TODO(), project.ID, "test-user")
	require.NoError(t, err)

	t.Run("successful priority update", func(t *testing.T) {
		// Create a task
		task, err := mgr.CreateTask(context.TODO(), project.ID, nil, "Test Task", "Description", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("id", "", "")
		flagSet.String("priority", "", "")
		flagSet.String("actor", "", "")

		_ = flagSet.Set("id", task.ID.String())
		_ = flagSet.Set("priority", "high")
		_ = flagSet.Set("actor", "test-updater")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := updatePrioritySubAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify the update
		updatedTask, err := mgr.GetTask(context.TODO(), task.ID)
		require.NoError(t, err)
		assert.Equal(t, types.TaskPriorityHigh, updatedTask.Priority)
	})
}

func TestUpdateComplexityAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Set project context
	err := mgr.SetSelectedProject(context.TODO(), project.ID, "test-user")
	require.NoError(t, err)

	t.Run("successful complexity update", func(t *testing.T) {
		// Create a task
		task, err := mgr.CreateTask(context.TODO(), project.ID, nil, "Test Task", "Description", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("id", "", "")
		flagSet.String("complexity", "", "")
		flagSet.String("actor", "", "")

		_ = flagSet.Set("id", task.ID.String())
		_ = flagSet.Set("complexity", "7")
		_ = flagSet.Set("actor", "test-updater")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := updateComplexitySubAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)

		// Verify the update
		updatedTask, err := mgr.GetTask(context.TODO(), task.ID)
		require.NoError(t, err)
		assert.Equal(t, 7, updatedTask.Complexity)
	})

	t.Run("invalid complexity value", func(t *testing.T) {
		// Create a task
		task, err := mgr.CreateTask(context.TODO(), project.ID, nil, "Test Task", "Description", 3, types.TaskPriorityMedium, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("id", "", "")
		flagSet.String("complexity", "", "")
		flagSet.String("actor", "", "")

		_ = flagSet.Set("id", task.ID.String())
		_ = flagSet.Set("complexity", "15") // Invalid: complexity must be 1-10
		_ = flagSet.Set("actor", "test-updater")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := updateComplexitySubAction(appCtx)
		err = action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "complexity 15 is out of range")
	})
}

func TestGetAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Set project context
	err := mgr.SetSelectedProject(context.TODO(), project.ID, "test-user")
	require.NoError(t, err)

	t.Run("get existing task", func(t *testing.T) {
		// Create a task
		task, err := mgr.CreateTask(context.TODO(), project.ID, nil, "Test Task", "Test description", 5, types.TaskPriorityHigh, "test-user")
		require.NoError(t, err)

		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("id", "", "")
		flagSet.Bool("json", false, "")
		flagSet.Bool("quiet", false, "")

		_ = flagSet.Set("id", task.ID.String())

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := getAction(appCtx)
		err = action(ctx)
		assert.NoError(t, err)
	})

	t.Run("invalid task ID", func(t *testing.T) {
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("id", "", "")
		flagSet.Bool("json", false, "")
		flagSet.Bool("quiet", false, "")

		_ = flagSet.Set("id", "invalid-uuid")

		ctx := cli.NewContext(app, flagSet, nil)

		appCtx := &shared.AppContext{
			ProjectManager: mgr,
			Logger:         config.Logger,
		}

		action := getAction(appCtx)
		err := action(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task-id format")
	})
}

func TestUpdateAction(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		expectedError string
	}{
		{
			name:        "no fields specified",
			args:        []string{"update", "--id", "123e4567-e89b-12d3-a456-426614174000"},
			expectError: true,
		},
		{
			name:        "state update",
			args:        []string{"update", "--id", "123e4567-e89b-12d3-a456-426614174000", "--state", "in-progress"},
			expectError: false,
		},
		{
			name:        "title update",
			args:        []string{"update", "--id", "123e4567-e89b-12d3-a456-426614174000", "--title", "New Title"},
			expectError: false,
		},
		{
			name:        "multiple fields update",
			args:        []string{"update", "--id", "123e4567-e89b-12d3-a456-426614174000", "--state", "in-progress", "--title", "Updated Task", "--priority", "high"},
			expectError: false,
		},
		{
			name:        "complexity update",
			args:        []string{"update", "--id", "123e4567-e89b-12d3-a456-426614174000", "--complexity", "7"},
			expectError: false,
		},
		{
			name:        "description update",
			args:        []string{"update", "--id", "123e4567-e89b-12d3-a456-426614174000", "--description", "New Description"},
			expectError: false,
		},
		{
			name:        "priority update",
			args:        []string{"update", "--id", "123e4567-e89b-12d3-a456-426614174000", "--priority", "low"},
			expectError: false,
		},
		{
			name:        "all fields update",
			args:        []string{"update", "--id", "123e4567-e89b-12d3-a456-426614174000", "--title", "Complete Update", "--description", "Full description", "--state", "in-progress", "--priority", "high", "--complexity", "8"},
			expectError: false,
		},
		{
			name:          "invalid complexity",
			args:          []string{"update", "--id", "123e4567-e89b-12d3-a456-426614174000", "--complexity", "15"},
			expectError:   true,
			expectedError: "complexity 15 is out of range",
		},
		{
			name:        "using aliases",
			args:        []string{"update", "--id", "123e4567-e89b-12d3-a456-426614174000", "-t", "Alias Title", "-d", "Alias Description", "-s", "in-progress", "-p", "medium", "-c", "6"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appCtx, projectID := createTestAppContextWithProject(t)

			// Create a test task to update
			task := testutil.CreateTestTask(t, appCtx.ProjectManager, projectID)
			// Replace the placeholder UUID with the actual task ID
			for i, arg := range tt.args {
				if arg == "--id" && i+1 < len(tt.args) && tt.args[i+1] == "123e4567-e89b-12d3-a456-426614174000" {
					tt.args[i+1] = task.ID.String()
					break
				}
			}

			// Create CLI app with our update command
			app := &cli.App{
				Name: "test",
				Commands: []*cli.Command{
					{
						Name:   "update",
						Usage:  "Update task fields",
						Action: updateAction(appCtx),
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Usage:    "Task ID",
								Required: true,
							},
							&cli.StringFlag{
								Name:    "title",
								Aliases: []string{"t"},
								Usage:   "New task title",
							},
							&cli.StringFlag{
								Name:    "description",
								Aliases: []string{"d"},
								Usage:   "New task description",
							},
							&cli.StringFlag{
								Name:    "state",
								Aliases: []string{"s"},
								Usage:   "New state",
							},
							&cli.StringFlag{
								Name:    "priority",
								Aliases: []string{"p"},
								Usage:   "New task priority",
							},
							&cli.IntFlag{
								Name:    "complexity",
								Aliases: []string{"c"},
								Usage:   "New task complexity",
							},
						},
					},
				},
			}

			// Set project context
			err := appCtx.ProjectManager.SetSelectedProject(context.Background(), projectID, "test-user")
			require.NoError(t, err)

			// Run the command
			err = app.Run(append([]string{"test"}, tt.args...))

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			} else {
				assert.NoError(t, err)

				// Verify the updates were actually applied
				updatedTask, err := appCtx.ProjectManager.GetTask(context.Background(), task.ID)
				require.NoError(t, err)

				// Check specific fields based on what was updated
				switch tt.name {
				case "state update":
					assert.Equal(t, types.TaskStateInProgress, updatedTask.State)
				case "title update":
					assert.Equal(t, "New Title", updatedTask.Title)
				case "description update":
					assert.Equal(t, "New Description", updatedTask.Description)
				case "priority update":
					assert.Equal(t, types.TaskPriorityLow, updatedTask.Priority)
				case "complexity update":
					assert.Equal(t, 7, updatedTask.Complexity)
				case "multiple fields update":
					assert.Equal(t, types.TaskStateInProgress, updatedTask.State)
					assert.Equal(t, "Updated Task", updatedTask.Title)
					assert.Equal(t, types.TaskPriorityHigh, updatedTask.Priority)
				case "all fields update":
					assert.Equal(t, "Complete Update", updatedTask.Title)
					assert.Equal(t, "Full description", updatedTask.Description)
					assert.Equal(t, types.TaskStateInProgress, updatedTask.State)
					assert.Equal(t, types.TaskPriorityHigh, updatedTask.Priority)
					assert.Equal(t, 8, updatedTask.Complexity)
				case "using aliases":
					assert.Equal(t, "Alias Title", updatedTask.Title)
					assert.Equal(t, "Alias Description", updatedTask.Description)
					assert.Equal(t, types.TaskStateInProgress, updatedTask.State)
					assert.Equal(t, types.TaskPriorityMedium, updatedTask.Priority)
					assert.Equal(t, 6, updatedTask.Complexity)
				}
			}
		})
	}
}

// createTestAppContextWithProject creates a test app context and returns the project ID
func createTestAppContextWithProject(t *testing.T) (*shared.AppContext, uuid.UUID) {
	t.Helper()

	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	return appCtx, project.ID
}
