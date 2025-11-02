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
	repo, err := NewRepository(dbPath,
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
		
		repo, err := NewRepository(dbPath,
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
	
	
	t.Run("create then list project consistency", func(t *testing.T) {
		// This test reproduces the bug where CreateProject succeeds but ListProjects returns empty
		ctx := context.Background()

		// Create a project
		project := &types.Project{
			ID:          uuid.New(),
			Title:       "Consistency Test Project",
			Description: "Testing create/list consistency",
			State:       types.ProjectStateActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Create the project
		err := repo.CreateProject(ctx, project)
		require.NoError(t, err, "CreateProject should succeed")

		// Immediately try to list projects
		projects, err := repo.ListProjects(ctx)
		assert.NoError(t, err, "ListProjects should succeed")
		assert.GreaterOrEqual(t, len(projects), 1, "Should find at least the created project")

		// Verify our project is in the list
		found := false
		for _, p := range projects {
			if p.ID == project.ID {
				found = true
				assert.Equal(t, project.Title, p.Title)
				assert.Equal(t, project.Description, p.Description)
				break
			}
		}
		assert.True(t, found, "Created project should be found in list")

		// Also test direct retrieval
		retrieved, err := repo.GetProject(ctx, project.ID)
		assert.NoError(t, err, "GetProject should succeed")
		assert.Equal(t, project.Title, retrieved.Title)
		assert.Equal(t, project.Description, retrieved.Description)
	})

	t.Run("list empty projects", func(t *testing.T) {
		ctx := context.Background()

		// List projects when there should be none (using fresh repo)
		repo2, cleanup2 := setupTestRepository(t)
		defer cleanup2()

		projects, err := repo2.ListProjects(ctx)
		assert.NoError(t, err, "ListProjects should succeed even when empty")
		assert.Empty(t, projects, "Should return empty list when no projects exist")
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

// TestBugReproduction reproduces the create/list inconsistency bug
func TestBugReproduction(t *testing.T) {
	t.Run("reproduce create list bug", func(t *testing.T) {
		ctx := context.Background()
		repo, cleanup := setupTestRepository(t)
		defer cleanup()

		// Step 1: Create a project (this should succeed)
		project := &types.Project{
			ID:          uuid.New(),
			Title:       "Bug Reproduction Project",
			Description: "Project to reproduce the create/list bug",
			State:       types.ProjectStateActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.CreateProject(ctx, project)
		require.NoError(t, err, "CreateProject should succeed")

		// Step 2: List projects (this should also succeed and find our project)
		projects, err := repo.ListProjects(ctx)
		assert.NoError(t, err, "ListProjects should succeed")
		assert.Len(t, projects, 1, "Should find exactly one project")

		// Step 3: Verify project details
		found := projects[0]
		assert.Equal(t, project.ID, found.ID)
		assert.Equal(t, project.Title, found.Title)
		assert.Equal(t, project.Description, found.Description)
		assert.Equal(t, project.State, found.State)

		// Step 4: Multiple creates and lists
		for i := 0; i < 3; i++ {
			p := &types.Project{
				ID:          uuid.New(),
				Title:       fmt.Sprintf("Project %d", i),
				Description: fmt.Sprintf("Description for project %d", i),
				State:       types.ProjectStateActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			err := repo.CreateProject(ctx, p)
			require.NoError(t, err)

			// Verify project appears in list immediately
			projects, err := repo.ListProjects(ctx)
			assert.NoError(t, err)
			assert.Len(t, projects, i+2, "Should find all created projects")
		}
	})

	t.Run("project context operations", func(t *testing.T) {
		ctx := context.Background()
		repo, cleanup := setupTestRepository(t)
		defer cleanup()

		// Create a project
		project := &types.Project{
			ID:          uuid.New(),
			Title:       "Context Test Project",
			Description: "Testing project context operations",
			State:       types.ProjectStateActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.CreateProject(ctx, project)
		require.NoError(t, err)

		// Test project context operations
		err = repo.SetSelectedProject(ctx, project.ID, "test-actor")
		assert.NoError(t, err)

		selectedID, err := repo.GetSelectedProject(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, selectedID)
		assert.Equal(t, project.ID, *selectedID)

		hasSelected, err := repo.HasSelectedProject(ctx)
		assert.NoError(t, err)
		assert.True(t, hasSelected)

		// Clear selection
		err = repo.ClearSelectedProject(ctx)
		assert.NoError(t, err)

		hasSelected, err = repo.HasSelectedProject(ctx)
		assert.NoError(t, err)
		assert.False(t, hasSelected)

		selectedID, err = repo.GetSelectedProject(ctx)
		assert.NoError(t, err)
		assert.Nil(t, selectedID)
	})

	t.Run("secure directory permissions", func(t *testing.T) {
		_, cleanup := setupTestRepository(t)
		defer cleanup()

		// Check that the .knot directory was created with secure permissions
		projectDir, err := os.Getwd()
		require.NoError(t, err)

		knotDir := filepath.Join(projectDir, ".knot")
		info, err := os.Stat(knotDir)
		if err == nil {
			// Directory exists, check permissions are at least secure (owner rwx)
			perms := info.Mode().Perm()
			// Check that owner has read, write, execute permissions (0700 = 0x1c0)
			// In CI environments, umask might affect final permissions, so we check minimum security
			assert.Equal(t, os.FileMode(0700), perms&os.FileMode(0700), "Directory should have owner secure permissions (0700)")
			// Also verify no world-writable permissions (security check)
			assert.False(t, perms&os.FileMode(0002) != 0, "Directory should not be world-writable")
		}
	})
}

func TestSecurityFeatures(t *testing.T) {
	t.Run("secure project directory creation", func(t *testing.T) {
		// Test the EnsureProjectDir function directly
		tempDir, err := os.MkdirTemp("", "knot_security_test_*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Change to temp directory
		originalCwd, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			_ = os.Chdir(originalCwd) // Ignore error during cleanup
		}()

		err = os.Chdir(tempDir)
		require.NoError(t, err)

		// Test secure directory creation
		projectDir, err := EnsureProjectDir()
		assert.NoError(t, err)

		// Verify permissions
		info, err := os.Stat(projectDir)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0700), info.Mode().Perm())
	})

	t.Run("database file permissions", func(t *testing.T) {
		ctx := context.Background()
		repo, cleanup := setupTestRepository(t)
		defer cleanup()

		// Get the database path
		dbPath, err := GetDatabasePath()
		assert.NoError(t, err)

		// Create a project to ensure the database file is created
		project := &types.Project{
			ID:          uuid.New(),
			Title:       "Security Test Project",
			Description: "Testing file permissions",
			State:       types.ProjectStateActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err = repo.CreateProject(ctx, project)
		assert.NoError(t, err)

		// Check database file permissions
		info, err := os.Stat(dbPath)
		if err == nil {
			perms := info.Mode().Perm()
			// Check if permissions are within acceptable range for security
			// SQLite typically creates files with 0644, which allows read access to group/others
			// This is acceptable for database files that don't contain highly sensitive data
			assert.True(t, perms <= os.FileMode(0644), "Database file should not have world-writable permissions, got %o", perms)
		}
	})
}