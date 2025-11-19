package sqlite

import (
	"context"
	"fmt"
	"testing"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEntRepositoryInterface tests that the ent-based repository implements all Repository interface methods correctly
func TestEntRepositoryInterface(t *testing.T) {
	// Skip if no database URL is provided (for CI/CD environments without database)
	databaseURL := getDatabaseURL()
	if databaseURL == "" {
		t.Skip("Skipping ent repository test - no database URL provided")
	}

	ctx := context.Background()

	// Create repository
	repo, err := NewRepository(databaseURL, WithAutoMigrate(true))
	require.NoError(t, err, "Failed to create ent repository")
	defer func() {
		if closer, ok := repo.(interface{ Close() error }); ok {
			closer.Close()
		}
	}()

	// Test Project Operations
	t.Run("Project Operations", func(t *testing.T) {
		testProject := &types.Project{
			ID:          uuid.New(),
			Title:       "Test Project",
			Description: "A test project for ent repository",
		}

		// Test CreateProject
		err := repo.CreateProject(ctx, testProject)
		require.NoError(t, err, "CreateProject should succeed")

		// Test GetProject
		retrievedProject, err := repo.GetProject(ctx, testProject.ID)
		require.NoError(t, err, "GetProject should succeed")
		assert.Equal(t, testProject.Title, retrievedProject.Title)
		assert.Equal(t, testProject.Description, retrievedProject.Description)

		// Test UpdateProject
		retrievedProject.Description = "Updated description"
		err = repo.UpdateProject(ctx, retrievedProject)
		require.NoError(t, err, "UpdateProject should succeed")

		// Verify update
		updatedProject, err := repo.GetProject(ctx, testProject.ID)
		require.NoError(t, err, "GetProject after update should succeed")
		assert.Equal(t, "Updated description", updatedProject.Description)

		// Test ListProjects
		projects, err := repo.ListProjects(ctx)
		require.NoError(t, err, "ListProjects should succeed")
		assert.GreaterOrEqual(t, len(projects), 1, "Should have at least one project")

		// Clean up for subsequent tests
		defer func() {
			_ = repo.DeleteProject(ctx, testProject.ID)
		}()
	})

	// Test Task Operations
	t.Run("Task Operations", func(t *testing.T) {
		// Create test project first
		testProject := &types.Project{
			ID:          uuid.New(),
			Title:       "Task Test Project",
			Description: "Project for testing tasks",
		}
		err := repo.CreateProject(ctx, testProject)
		require.NoError(t, err)
		defer func() { _ = repo.DeleteProject(ctx, testProject.ID) }()

		// Test CreateTask
		testTask := &types.Task{
			ID:          uuid.New(),
			ProjectID:   testProject.ID,
			Title:       "Test Task",
			Description: "A test task",
			State:       types.TaskStatePending,
			Complexity:  5,
		}

		err = repo.CreateTask(ctx, testTask)
		require.NoError(t, err, "CreateTask should succeed")

		// Test GetTask
		retrievedTask, err := repo.GetTask(ctx, testTask.ID)
		require.NoError(t, err, "GetTask should succeed")
		assert.Equal(t, testTask.Title, retrievedTask.Title)
		assert.Equal(t, testTask.State, retrievedTask.State)
		assert.Equal(t, testTask.Complexity, retrievedTask.Complexity)

		// Test UpdateTask
		retrievedTask.State = types.TaskStateInProgress
		retrievedTask.Description = "Updated task description"
		err = repo.UpdateTask(ctx, retrievedTask)
		require.NoError(t, err, "UpdateTask should succeed")

		// Verify update
		updatedTask, err := repo.GetTask(ctx, testTask.ID)
		require.NoError(t, err)
		assert.Equal(t, types.TaskStateInProgress, updatedTask.State)
		assert.Equal(t, "Updated task description", updatedTask.Description)
	})

	// Test Task Hierarchy
	t.Run("Task Hierarchy", func(t *testing.T) {
		// Create test project
		testProject := &types.Project{
			ID:    uuid.New(),
			Title: "Hierarchy Test Project",
		}
		err := repo.CreateProject(ctx, testProject)
		require.NoError(t, err)
		defer func() { _ = repo.DeleteProject(ctx, testProject.ID) }()

		// Create parent task
		parentTask := &types.Task{
			ID:         uuid.New(),
			ProjectID:  testProject.ID,
			Title:      "Parent Task",
			State:      types.TaskStatePending,
			Complexity: 3,
		}
		err = repo.CreateTask(ctx, parentTask)
		require.NoError(t, err)

		// Create child task
		childTask := &types.Task{
			ID:         uuid.New(),
			ProjectID:  testProject.ID,
			ParentID:   &parentTask.ID,
			Title:      "Child Task",
			State:      types.TaskStatePending,
			Complexity: 2,
		}
		err = repo.CreateTask(ctx, childTask)
		require.NoError(t, err)

		// Test GetTasksByParent
		children, err := repo.GetTasksByParent(ctx, parentTask.ID)
		require.NoError(t, err)
		assert.Len(t, children, 1)
		assert.Equal(t, childTask.ID, children[0].ID)

		// Test GetParentTask
		parent, err := repo.GetParentTask(ctx, childTask.ID)
		require.NoError(t, err)
		assert.Equal(t, parentTask.ID, parent.ID)

		// Test GetRootTasks
		rootTasks, err := repo.GetRootTasks(ctx, testProject.ID)
		require.NoError(t, err)
		assert.Len(t, rootTasks, 1)
		assert.Equal(t, parentTask.ID, rootTasks[0].ID)

		// Test GetTasksByProject
		allTasks, err := repo.GetTasksByProject(ctx, testProject.ID)
		require.NoError(t, err)
		assert.Len(t, allTasks, 2)
	})

	// Test Task Dependencies
	t.Run("Task Dependencies", func(t *testing.T) {
		// Create test project
		testProject := &types.Project{
			ID:    uuid.New(),
			Title: "Dependency Test Project",
		}
		err := repo.CreateProject(ctx, testProject)
		require.NoError(t, err)
		defer func() { _ = repo.DeleteProject(ctx, testProject.ID) }()

		// Create two tasks
		task1 := &types.Task{
			ID:         uuid.New(),
			ProjectID:  testProject.ID,
			Title:      "Task 1",
			State:      types.TaskStatePending,
			Complexity: 3,
		}
		task2 := &types.Task{
			ID:         uuid.New(),
			ProjectID:  testProject.ID,
			Title:      "Task 2",
			State:      types.TaskStatePending,
			Complexity: 4,
		}

		err = repo.CreateTask(ctx, task1)
		require.NoError(t, err)
		err = repo.CreateTask(ctx, task2)
		require.NoError(t, err)

		// Test AddTaskDependency
		updatedTask, err := repo.AddTaskDependency(ctx, task2.ID, task1.ID)
		require.NoError(t, err)
		assert.Contains(t, updatedTask.Dependencies, task1.ID)

		// Test GetTaskDependencies
		dependencies, err := repo.GetTaskDependencies(ctx, task2.ID)
		require.NoError(t, err)
		assert.Len(t, dependencies, 1)
		assert.Equal(t, task1.ID, dependencies[0].ID)

		// Test GetDependentTasks
		dependents, err := repo.GetDependentTasks(ctx, task1.ID)
		require.NoError(t, err)
		assert.Len(t, dependents, 1)
		assert.Equal(t, task2.ID, dependents[0].ID)

		// Test RemoveTaskDependency
		updatedTask, err = repo.RemoveTaskDependency(ctx, task2.ID, task1.ID)
		require.NoError(t, err)
		assert.NotContains(t, updatedTask.Dependencies, task1.ID)
	})

	// Test Project Progress
	t.Run("Project Progress", func(t *testing.T) {
		// Create test project
		testProject := &types.Project{
			ID:    uuid.New(),
			Title: "Progress Test Project",
		}
		err := repo.CreateProject(ctx, testProject)
		require.NoError(t, err)
		defer func() { _ = repo.DeleteProject(ctx, testProject.ID) }()

		// Create tasks with different states
		tasks := []*types.Task{
			{
				ID:         uuid.New(),
				ProjectID:  testProject.ID,
				Title:      "Completed Task",
				State:      types.TaskStateCompleted,
				Complexity: 3,
			},
			{
				ID:         uuid.New(),
				ProjectID:  testProject.ID,
				Title:      "In Progress Task",
				State:      types.TaskStateInProgress,
				Complexity: 4,
			},
			{
				ID:         uuid.New(),
				ProjectID:  testProject.ID,
				Title:      "Pending Task",
				State:      types.TaskStatePending,
				Complexity: 2,
			},
		}

		for _, task := range tasks {
			err = repo.CreateTask(ctx, task)
			require.NoError(t, err)
		}

		// Test GetProjectProgress
		progress, err := repo.GetProjectProgress(ctx, testProject.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, progress.TotalTasks)
		assert.Equal(t, 1, progress.CompletedTasks)
		assert.Equal(t, 1, progress.InProgressTasks)
		assert.Equal(t, 1, progress.PendingTasks)
		assert.InDelta(t, 33.33, progress.OverallProgress, 0.1)
	})

	// Test Error Cases
	t.Run("Error Cases", func(t *testing.T) {
		// Test getting non-existent project
		_, err := repo.GetProject(ctx, uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Test getting non-existent task
		_, err = repo.GetTask(ctx, uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Test creating task with non-existent project
		invalidTask := &types.Task{
			ID:         uuid.New(),
			ProjectID:  uuid.New(),
			Title:      "Invalid Task",
			State:      types.TaskStatePending,
			Complexity: 1,
		}
		err = repo.CreateTask(ctx, invalidTask)
		assert.Error(t, err)
	})
}

// getDatabaseURL returns a test database URL or empty string if not available
func getDatabaseURL() string {
	// In a real test environment, this would come from environment variables
	// For now, return empty to skip tests when no database is available
	return ""
}

// BenchmarkEntRepository benchmarks the ent repository performance
func BenchmarkEntRepository(b *testing.B) {
	databaseURL := getDatabaseURL()
	if databaseURL == "" {
		b.Skip("Skipping benchmark - no database URL provided")
	}

	ctx := context.Background()
	repo, err := NewRepository(databaseURL, WithAutoMigrate(true))
	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}
	defer func() {
		if closer, ok := repo.(interface{ Close() error }); ok {
			closer.Close()
		}
	}()

	// Create test project
	testProject := &types.Project{
		ID:    uuid.New(),
		Title: "Benchmark Project",
	}
	err = repo.CreateProject(ctx, testProject)
	if err != nil {
		b.Fatalf("Failed to create test project: %v", err)
	}
	defer func() { _ = repo.DeleteProject(ctx, testProject.ID) }()

	b.Run("CreateTask", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			task := &types.Task{
				ID:         uuid.New(),
				ProjectID:  testProject.ID,
				Title:      fmt.Sprintf("Benchmark Task %d", i),
				State:      types.TaskStatePending,
				Complexity: 1,
			}
			_ = repo.CreateTask(ctx, task)
		}
	})

	b.Run("GetTask", func(b *testing.B) {
		// Create a test task first
		testTask := &types.Task{
			ID:         uuid.New(),
			ProjectID:  testProject.ID,
			Title:      "Get Task Test",
			State:      types.TaskStatePending,
			Complexity: 1,
		}
		_ = repo.CreateTask(ctx, testTask)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = repo.GetTask(ctx, testTask.ID)
		}
	})
}
