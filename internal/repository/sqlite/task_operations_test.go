package sqlite

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetTasksWithDependencies tests the GetTasksWithDependencies function comprehensively
func TestGetTasksWithDependencies(t *testing.T) {
	repo, cleanup := setupTestRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test project
	project := &types.Project{
		ID:          uuid.New(),
		Title:       "Dependency Test Project",
		Description: "Project for testing task dependencies",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.CreateProject(ctx, project)
	require.NoError(t, err)

	t.Run("empty task list", func(t *testing.T) {
		// Test with empty list of task IDs
		tasks, err := repo.GetTasksWithDependencies(ctx, []uuid.UUID{})
		assert.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("single task without dependencies", func(t *testing.T) {
		// Create a task without dependencies
		task := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Simple Task",
			Description: "A task without dependencies",
			State:       types.TaskStatePending,
			Priority:    types.TaskPriorityMedium,
			Complexity:  2,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.CreateTask(ctx, task)
		require.NoError(t, err)

		// Get the task using GetTasksWithDependencies
		tasks, err := repo.GetTasksWithDependencies(ctx, []uuid.UUID{task.ID})
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		
		retrieved := tasks[0]
		assert.Equal(t, task.ID, retrieved.ID)
		assert.Equal(t, task.Title, retrieved.Title)
		assert.Equal(t, task.Description, retrieved.Description)
		assert.Equal(t, task.State, retrieved.State)
		assert.Equal(t, task.Priority, retrieved.Priority)
		assert.Equal(t, task.Complexity, retrieved.Complexity)
		assert.Empty(t, retrieved.Dependencies)
		assert.Empty(t, retrieved.Dependents)
	})

	t.Run("multiple tasks without dependencies", func(t *testing.T) {
		// Create multiple tasks without dependencies
		task1 := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Task 1",
			Description: "First task",
			State:       types.TaskStatePending,
			Priority:    types.TaskPriorityLow,
			Complexity:  1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		task2 := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Task 2",
			Description: "Second task",
			State:       types.TaskStateInProgress,
			Priority:    types.TaskPriorityHigh,
			Complexity:  3,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.CreateTask(ctx, task1)
		require.NoError(t, err)
		err = repo.CreateTask(ctx, task2)
		require.NoError(t, err)

		// Get both tasks
		tasks, err := repo.GetTasksWithDependencies(ctx, []uuid.UUID{task1.ID, task2.ID})
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)

		// Verify both tasks are present (order may vary)
		taskMap := make(map[uuid.UUID]*types.Task)
		for _, task := range tasks {
			taskMap[task.ID] = task
		}

		assert.Contains(t, taskMap, task1.ID)
		assert.Contains(t, taskMap, task2.ID)
		assert.Equal(t, task1.Title, taskMap[task1.ID].Title)
		assert.Equal(t, task2.Title, taskMap[task2.ID].Title)
	})

	t.Run("tasks with dependencies", func(t *testing.T) {
		// Create three tasks: A depends on B, B depends on C
		taskC := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Task C (Base)",
			Description: "Base task with no dependencies",
			State:       types.TaskStatePending,
			Priority:    types.TaskPriorityMedium,
			Complexity:  1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		taskB := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Task B (Middle)",
			Description: "Middle task that depends on C",
			State:       types.TaskStatePending,
			Priority:    types.TaskPriorityMedium,
			Complexity:  2,
			Dependencies: []uuid.UUID{taskC.ID},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		taskA := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Task A (Top)",
			Description: "Top task that depends on B",
			State:       types.TaskStatePending,
			Priority:    types.TaskPriorityMedium,
			Complexity:  3,
			Dependencies: []uuid.UUID{taskB.ID},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Create tasks in dependency order
		err := repo.CreateTask(ctx, taskC)
		require.NoError(t, err)
		err = repo.CreateTask(ctx, taskB)
		require.NoError(t, err)
		err = repo.CreateTask(ctx, taskA)
		require.NoError(t, err)

		// Get all tasks with dependencies
		tasks, err := repo.GetTasksWithDependencies(ctx, []uuid.UUID{taskA.ID, taskB.ID, taskC.ID})
		assert.NoError(t, err)
		assert.Len(t, tasks, 3)

		// Create a map for easier lookup
		taskMap := make(map[uuid.UUID]*types.Task)
		for _, task := range tasks {
			taskMap[task.ID] = task
		}

		// Verify task A
		retrievedA := taskMap[taskA.ID]
		assert.NotNil(t, retrievedA)
		assert.Equal(t, taskA.Title, retrievedA.Title)
		assert.Len(t, retrievedA.Dependencies, 1)
		assert.Contains(t, retrievedA.Dependencies, taskB.ID)
		assert.Empty(t, retrievedA.Dependents)

		// Verify task B
		retrievedB := taskMap[taskB.ID]
		assert.NotNil(t, retrievedB)
		assert.Equal(t, taskB.Title, retrievedB.Title)
		assert.Len(t, retrievedB.Dependencies, 1)
		assert.Contains(t, retrievedB.Dependencies, taskC.ID)
		assert.Len(t, retrievedB.Dependents, 1)
		assert.Contains(t, retrievedB.Dependents, taskA.ID)

		// Verify task C
		retrievedC := taskMap[taskC.ID]
		assert.NotNil(t, retrievedC)
		assert.Equal(t, taskC.Title, retrievedC.Title)
		assert.Empty(t, retrievedC.Dependencies)
		assert.Len(t, retrievedC.Dependents, 1)
		assert.Contains(t, retrievedC.Dependents, taskB.ID)
	})

	t.Run("tasks with multiple dependencies", func(t *testing.T) {
		// Create a task that depends on multiple other tasks
		dep1 := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Dependency 1",
			Description: "First dependency",
			State:       types.TaskStateCompleted,
			Priority:    types.TaskPriorityMedium,
			Complexity:  1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		dep2 := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Dependency 2",
			Description: "Second dependency",
			State:       types.TaskStateCompleted,
			Priority:    types.TaskPriorityMedium,
			Complexity:  1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mainTask := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Main Task",
			Description: "Task with multiple dependencies",
			State:       types.TaskStatePending,
			Priority:    types.TaskPriorityHigh,
			Complexity:  4,
			Dependencies: []uuid.UUID{dep1.ID, dep2.ID},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Create tasks
		err := repo.CreateTask(ctx, dep1)
		require.NoError(t, err)
		err = repo.CreateTask(ctx, dep2)
		require.NoError(t, err)
		err = repo.CreateTask(ctx, mainTask)
		require.NoError(t, err)

		// Get all tasks
		tasks, err := repo.GetTasksWithDependencies(ctx, []uuid.UUID{mainTask.ID, dep1.ID, dep2.ID})
		assert.NoError(t, err)
		assert.Len(t, tasks, 3)

		// Find main task
		var retrievedMain *types.Task
		for _, task := range tasks {
			if task.ID == mainTask.ID {
				retrievedMain = task
				break
			}
		}

		assert.NotNil(t, retrievedMain)
		assert.Len(t, retrievedMain.Dependencies, 2)
		assert.Contains(t, retrievedMain.Dependencies, dep1.ID)
		assert.Contains(t, retrievedMain.Dependencies, dep2.ID)

		// Find dependency tasks and verify dependents
		depMap := make(map[uuid.UUID]*types.Task)
		for _, task := range tasks {
			if task.ID == dep1.ID || task.ID == dep2.ID {
				depMap[task.ID] = task
			}
		}

		assert.Len(t, depMap, 2)
		for _, depTask := range depMap {
			assert.Len(t, depTask.Dependents, 1)
			assert.Contains(t, depTask.Dependents, mainTask.ID)
		}
	})

	t.Run("non-existent task", func(t *testing.T) {
		// Test with a task ID that doesn't exist
		nonExistentID := uuid.New()
		_, err := repo.GetTasksWithDependencies(ctx, []uuid.UUID{nonExistentID})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected 1 tasks, got 0")
	})

	t.Run("partial non-existent tasks", func(t *testing.T) {
		// Create one valid task
		validTask := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Valid Task",
			Description: "A valid task",
			State:       types.TaskStatePending,
			Priority:    types.TaskPriorityMedium,
			Complexity:  2,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.CreateTask(ctx, validTask)
		require.NoError(t, err)

		// Test with mix of valid and invalid task IDs
		nonExistentID := uuid.New()
		_, err = repo.GetTasksWithDependencies(ctx, []uuid.UUID{validTask.ID, nonExistentID})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected 2 tasks, got 1")
	})

	t.Run("tasks with hierarchical relationships", func(t *testing.T) {
		// Create parent task
		parentTask := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Parent Task",
			Description: "A parent task",
			State:       types.TaskStatePending,
			Priority:    types.TaskPriorityMedium,
			Complexity:  1,
			Depth:       0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Create child task
		childTask := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			ParentID:    &parentTask.ID,
			Title:       "Child Task",
			Description: "A child task",
			State:       types.TaskStatePending,
			Priority:    types.TaskPriorityMedium,
			Complexity:  2,
			Depth:       1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.CreateTask(ctx, parentTask)
		require.NoError(t, err)
		err = repo.CreateTask(ctx, childTask)
		require.NoError(t, err)

		// Get both tasks
		tasks, err := repo.GetTasksWithDependencies(ctx, []uuid.UUID{parentTask.ID, childTask.ID})
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)

		// Verify parent task
		var retrievedParent, retrievedChild *types.Task
		for _, task := range tasks {
			if task.ID == parentTask.ID {
				retrievedParent = task
			} else if task.ID == childTask.ID {
				retrievedChild = task
			}
		}

		assert.NotNil(t, retrievedParent)
		assert.NotNil(t, retrievedChild)
		assert.Equal(t, 0, retrievedParent.Depth)
		assert.Equal(t, 1, retrievedChild.Depth)
		assert.Equal(t, parentTask.ID, *retrievedChild.ParentID)
		assert.Nil(t, retrievedParent.ParentID)
	})

	t.Run("performance test with many tasks", func(t *testing.T) {
		// Create multiple tasks to test batch loading performance
		numTasks := 10
		taskIDs := make([]uuid.UUID, numTasks)
		
		for i := 0; i < numTasks; i++ {
			task := &types.Task{
				ID:          uuid.New(),
				ProjectID:   project.ID,
				Title:       fmt.Sprintf("Perf Task %d", i),
				Description: fmt.Sprintf("Performance test task %d", i),
				State:       types.TaskStatePending,
				Priority:    types.TaskPriorityMedium,
				Complexity:  1,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			
			err := repo.CreateTask(ctx, task)
			require.NoError(t, err)
			taskIDs[i] = task.ID
		}

		// Measure time for batch loading
		start := time.Now()
		tasks, err := repo.GetTasksWithDependencies(ctx, taskIDs)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Len(t, tasks, numTasks)

		// This should be much faster than calling GetTask multiple times
		// The exact timing depends on hardware, but it should be sub-second
		t.Logf("Loaded %d tasks in %v", numTasks, duration)
		assert.Less(t, duration, time.Second)

		// Verify all tasks were loaded correctly
		taskMap := make(map[uuid.UUID]*types.Task)
		for _, task := range tasks {
			taskMap[task.ID] = task
		}

		for i, id := range taskIDs {
			task := taskMap[id]
			assert.NotNil(t, task)
			assert.Equal(t, fmt.Sprintf("Perf Task %d", i), task.Title)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		// Test with cancelled context
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := repo.GetTasksWithDependencies(cancelledCtx, []uuid.UUID{uuid.New()})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

// TestGetTasksWithDependenciesEdgeCases tests edge cases and error conditions
func TestGetTasksWithDependenciesEdgeCases(t *testing.T) {
	repo, cleanup := setupTestRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test project
	project := &types.Project{
		ID:          uuid.New(),
		Title:       "Edge Case Test Project",
		Description: "Project for testing edge cases",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.CreateProject(ctx, project)
	require.NoError(t, err)

	t.Run("large number of task IDs", func(t *testing.T) {
		// Test with a large number of task IDs (some exist, some don't)
		const numTasks = 50
		taskIDs := make([]uuid.UUID, numTasks)
		
		// Create half the tasks
		for i := 0; i < numTasks/2; i++ {
			task := &types.Task{
				ID:          uuid.New(),
				ProjectID:   project.ID,
				Title:       fmt.Sprintf("Large Test Task %d", i),
				Description: fmt.Sprintf("Task %d for large test", i),
				State:       types.TaskStatePending,
				Priority:    types.TaskPriorityMedium,
				Complexity:  1,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			
			err := repo.CreateTask(ctx, task)
			require.NoError(t, err)
			taskIDs[i] = task.ID
		}

		// Fill the rest with non-existent IDs
		for i := numTasks / 2; i < numTasks; i++ {
			taskIDs[i] = uuid.New()
		}

		// This should fail because not all tasks exist
		_, err := repo.GetTasksWithDependencies(ctx, taskIDs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("expected %d tasks", numTasks))
	})

	t.Run("duplicate task IDs", func(t *testing.T) {
		// Create a single task
		task := &types.Task{
			ID:          uuid.New(),
			ProjectID:   project.ID,
			Title:       "Duplicate Test Task",
			Description: "Task for testing duplicates",
			State:       types.TaskStatePending,
			Priority:    types.TaskPriorityMedium,
			Complexity:  1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.CreateTask(ctx, task)
		require.NoError(t, err)

		// Request the same task multiple times
		duplicateIDs := []uuid.UUID{task.ID, task.ID, task.ID}
		
		// This should fail because the duplicate IDs result in fewer tasks than expected
		_, err = repo.GetTasksWithDependencies(ctx, duplicateIDs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected 3 tasks, got 1")
	})

	t.Run("timeout context", func(t *testing.T) {
		// Test with timeout context
		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		
		// Wait for timeout to trigger
		time.Sleep(time.Millisecond)
		
		_, err := repo.GetTasksWithDependencies(timeoutCtx, []uuid.UUID{uuid.New()})
		assert.Error(t, err)
		// Should contain either "context deadline exceeded" or "context canceled"
		assert.True(t, 
			err.Error() == "context deadline exceeded" || 
			err.Error() == "context canceled" ||
			strings.Contains(err.Error(), "context"),
		)
	})
}