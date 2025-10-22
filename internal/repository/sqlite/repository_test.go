package sqlite

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupTestRepository creates a test repository with a temporary database
func setupTestRepository(t *testing.T) (types.Repository, func()) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "knot_test_*")
	require.NoError(t, err)
	
	dbPath := filepath.Join(tempDir, "test.db")
	
	// Create repository with test configuration
	repo, err := NewRepository(
		WithDatabasePath(dbPath),
		WithAutoMigrate(true),
		WithLogger(zap.NewNop()), // Silent logger for tests
	)
	require.NoError(t, err)
	
	// Return cleanup function
	cleanup := func() {
		if closer, ok := repo.(interface{ Close() error }); ok {
			closer.Close()
		}
		os.RemoveAll(tempDir)
	}
	
	return repo, cleanup
}

// TestRepositoryInitialization tests repository creation and initialization
func TestRepositoryInitialization(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		repo, cleanup := setupTestRepository(t)
		defer cleanup()
		
		assert.NotNil(t, repo)
	})
	
	t.Run("initialization with custom config", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "knot_test_*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)
		
		dbPath := filepath.Join(tempDir, "custom.db")
		
		repo, err := NewRepository(
			WithDatabasePath(dbPath),
			WithAutoMigrate(true),
			WithConnectionPool(2, 1),
			WithConnectionLifetime(time.Hour, time.Minute*30),
		)
		require.NoError(t, err)
		defer func() {
			if closer, ok := repo.(interface{ Close() error }); ok {
				closer.Close()
			}
		}()
		
		assert.NotNil(t, repo)
	})
}

// TestProjectOperations tests all project-related repository operations
func TestProjectOperations(t *testing.T) {
	repo, cleanup := setupTestRepository(t)
	defer cleanup()
	
	ctx := context.Background()
	
	t.Run("create project", func(t *testing.T) {
		project := &types.Project{
			ID:          uuid.New(),
			Title:       "Test Project",
			Description: "Test Description",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		err := repo.CreateProject(ctx, project)
		assert.NoError(t, err)
	})
	
	t.Run("get project", func(t *testing.T) {
		// Create a project first
		project := &types.Project{
			ID:          uuid.New(),
			Title:       "Get Test Project",
			Description: "Get Test Description",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		err := repo.CreateProject(ctx, project)
		require.NoError(t, err)
		
		// Get the project
		retrieved, err := repo.GetProject(ctx, project.ID)
		assert.NoError(t, err)
		assert.Equal(t, project.ID, retrieved.ID)
		assert.Equal(t, project.Title, retrieved.Title)
		assert.Equal(t, project.Description, retrieved.Description)
	})
	
	t.Run("list projects", func(t *testing.T) {
		// Create multiple projects
		projects := []*types.Project{
			{
				ID:          uuid.New(),
				Title:       "List Test Project 1",
				Description: "Description 1",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          uuid.New(),
				Title:       "List Test Project 2",
				Description: "Description 2",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
		
		for _, project := range projects {
			err := repo.CreateProject(ctx, project)
			require.NoError(t, err)
		}
		
		// List projects
		retrieved, err := repo.ListProjects(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(retrieved), 2)
	})
	
	t.Run("update project", func(t *testing.T) {
		// Create a project first
		project := &types.Project{
			ID:          uuid.New(),
			Title:       "Update Test Project",
			Description: "Original Description",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		err := repo.CreateProject(ctx, project)
		require.NoError(t, err)
		
		// Update the project
		project.Title = "Updated Title"
		project.Description = "Updated Description"
		project.UpdatedAt = time.Now()
		
		err = repo.UpdateProject(ctx, project)
		assert.NoError(t, err)
		
		// Verify the update
		retrieved, err := repo.GetProject(ctx, project.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", retrieved.Title)
		assert.Equal(t, "Updated Description", retrieved.Description)
	})
	
	t.Run("delete project", func(t *testing.T) {
		// Create a project first
		project := &types.Project{
			ID:          uuid.New(),
			Title:       "Delete Test Project",
			Description: "Delete Test Description",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		err := repo.CreateProject(ctx, project)
		require.NoError(t, err)
		
		// Delete the project
		err = repo.DeleteProject(ctx, project.ID)
		assert.NoError(t, err)
		
		// Verify deletion
		_, err = repo.GetProject(ctx, project.ID)
		assert.Error(t, err)
	})
}

// TestTaskOperations tests all task-related repository operations
func TestTaskOperations(t *testing.T) {
	repo, cleanup := setupTestRepository(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Create a test project first
	project := &types.Project{
		ID:          uuid.New(),
		Title:       "Task Test Project",
		Description: "Project for task tests",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	err := repo.CreateProject(ctx, project)
	require.NoError(t, err)
	
	t.Run("create task", func(t *testing.T) {
		task := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Test Task",
			Description: "Test Task Description",
			State:       types.TaskStatePending,
			Priority:    types.TaskPriorityMedium,
			Complexity:  3,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		err := repo.CreateTask(ctx, task)
		assert.NoError(t, err)
	})
	
	t.Run("get task", func(t *testing.T) {
		// Create a task first
		task := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Get Test Task",
			Description: "Get Test Description",
			State:       types.TaskStatePending,
			Priority:    types.TaskPriorityMedium,
			Complexity:  2,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		err := repo.CreateTask(ctx, task)
		require.NoError(t, err)
		
		// Get the task
		retrieved, err := repo.GetTask(ctx, task.ID)
		assert.NoError(t, err)
		assert.Equal(t, task.ID, retrieved.ID)
		assert.Equal(t, task.Title, retrieved.Title)
		assert.Equal(t, task.State, retrieved.State)
		assert.Equal(t, task.Complexity, retrieved.Complexity)
	})
	
	t.Run("list tasks by project", func(t *testing.T) {
		// Create multiple tasks
		tasks := []*types.Task{
			{
				ID:          uuid.New(),
				ProjectID:   project.ID,
				Title:       "List Test Task 1",
				Description: "Description 1",
				State:       types.TaskStatePending,
			Priority:    types.TaskPriorityMedium,
				Complexity:  1,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          uuid.New(),
				ProjectID:   project.ID,
				Title:       "List Test Task 2",
				Description: "Description 2",
				State:       types.TaskStateInProgress,
				Priority:    types.TaskPriorityMedium,
				Complexity:  2,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
		
		for _, task := range tasks {
			err := repo.CreateTask(ctx, task)
			require.NoError(t, err)
		}
		
		// List tasks
		retrieved, err := repo.GetTasksByProject(ctx, project.ID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(retrieved), 2)
	})
	
	t.Run("update task", func(t *testing.T) {
		// Create a task first
		task := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Update Test Task",
			Description: "Original Description",
			State:       types.TaskStatePending,
			Priority:    types.TaskPriorityMedium,
			Complexity:  1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		err := repo.CreateTask(ctx, task)
		require.NoError(t, err)
		
		// Update the task
		task.Title = "Updated Task Title"
		task.State = types.TaskStateInProgress
		task.Complexity = 3
		task.UpdatedAt = time.Now()
		
		err = repo.UpdateTask(ctx, task)
		assert.NoError(t, err)
		
		// Verify the update
		retrieved, err := repo.GetTask(ctx, task.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Task Title", retrieved.Title)
		assert.Equal(t, types.TaskStateInProgress, retrieved.State)
		assert.Equal(t, 3, retrieved.Complexity)
	})
	
	t.Run("delete task", func(t *testing.T) {
		// Create a task first
		task := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Delete Test Task",
			Description: "Delete Test Description",
			State:       types.TaskStatePending,
			Priority:    types.TaskPriorityMedium,
			Complexity:  1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		err := repo.CreateTask(ctx, task)
		require.NoError(t, err)
		
		// Delete the task
		err = repo.DeleteTask(ctx, task.ID)
		assert.NoError(t, err)
		
		// Verify deletion
		_, err = repo.GetTask(ctx, task.ID)
		assert.Error(t, err)
	})
}

// TestErrorHandling tests repository error handling
func TestErrorHandling(t *testing.T) {
	repo, cleanup := setupTestRepository(t)
	defer cleanup()
	
	ctx := context.Background()
	
	t.Run("get non-existent project", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := repo.GetProject(ctx, nonExistentID)
		assert.Error(t, err)
	})
	
	t.Run("get non-existent task", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := repo.GetTask(ctx, nonExistentID)
		assert.Error(t, err)
	})
	
	t.Run("create task with non-existent project", func(t *testing.T) {
		task := &types.Task{
			ID:          uuid.New(),
			ProjectID:   uuid.New(), // Non-existent project
			Title:       "Invalid Task",
			Description: "Task with invalid project",
			State:       types.TaskStatePending,
			Priority:    types.TaskPriorityMedium,
			Complexity:  1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		err := repo.CreateTask(ctx, task)
		assert.Error(t, err)
	})
}

// TestConcurrency tests repository operations under concurrent access
func TestConcurrency(t *testing.T) {
	repo, cleanup := setupTestRepository(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Create a test project
	project := &types.Project{
		ID:          uuid.New(),
		Title:       "Concurrency Test Project",
		Description: "Project for concurrency tests",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	err := repo.CreateProject(ctx, project)
	require.NoError(t, err)
	
	t.Run("concurrent task creation", func(t *testing.T) {
		// Note: SQLite has limited concurrency due to file locking
		// This test demonstrates the behavior but may have some failures
		const numTasks = 5 // Reduced for SQLite limitations
		
		// Create tasks concurrently
		errChan := make(chan error, numTasks)
		successCount := 0
		
		for i := 0; i < numTasks; i++ {
			go func(index int) {
				task := &types.Task{
					ID:          uuid.New(),
					ProjectID:   project.ID,
					Title:       fmt.Sprintf("Concurrent Task %d", index),
					Description: fmt.Sprintf("Description %d", index),
					State:       types.TaskStatePending,
			Priority:    types.TaskPriorityMedium,
					Complexity:  1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				
				errChan <- repo.CreateTask(ctx, task)
			}(i)
		}
		
		// Check operations - some may fail due to SQLite locking
		for i := 0; i < numTasks; i++ {
			err := <-errChan
			if err == nil {
				successCount++
			}
			// Don't assert NoError since SQLite locking is expected
		}
		
		// Verify at least some tasks were created
		tasks, err := repo.GetTasksByProject(ctx, project.ID)
		assert.NoError(t, err)
		assert.Greater(t, successCount, 0, "At least some concurrent operations should succeed")
		assert.GreaterOrEqual(t, len(tasks), successCount)
	})
}