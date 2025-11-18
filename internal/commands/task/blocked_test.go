package task

import (
	"context"
	"testing"
	"time"

	"github.com/denkhaus/knot/internal/manager"
	"github.com/denkhaus/knot/internal/repository/inmemory"
	"github.com/denkhaus/knot/internal/types"
	"github.com/denkhaus/knot/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBlockedCommandWithDependencies tests the blocked command functionality
func TestBlockedCommandWithDependencies(t *testing.T) {
	// Setup test environment
	repo := inmemory.NewMemoryRepository()
	config := manager.DefaultConfig()
	projectManager := manager.NewManagerWithRepository(repo, config)

	ctx := context.Background()

	// Create a test project
	project, err := projectManager.CreateProject(ctx, "Test Project", "Project for testing blocked command", "TestUser")
	require.NoError(t, err)

	// Create test tasks with dependencies
	// Task A (completed) - no dependencies
	taskA, err := projectManager.CreateTask(ctx, project.ID, nil, "Task A - Foundation", "Completed foundation task", 3, types.TaskPriorityMedium, "TestUser")
	require.NoError(t, err)
	_, err = projectManager.UpdateTaskState(ctx, taskA.ID, types.TaskStateInProgress, "TestUser")
	require.NoError(t, err)
	_, err = projectManager.UpdateTaskState(ctx, taskA.ID, types.TaskStateCompleted, "TestUser")
	require.NoError(t, err)

	// Task B (in-progress) - depends on A (should not be blocked)
	taskB, err := projectManager.CreateTask(ctx, project.ID, nil, "Task B - Build on A", "Task that builds on A", 4, types.TaskPriorityMedium, "TestUser")
	require.NoError(t, err)
	_, err = projectManager.AddTaskDependency(ctx, taskB.ID, taskA.ID, "TestUser")
	require.NoError(t, err)
	_, err = projectManager.UpdateTaskState(ctx, taskB.ID, types.TaskStateInProgress, "TestUser")
	require.NoError(t, err)

	// Task C (pending) - depends on B (should be blocked)
	taskC, err := projectManager.CreateTask(ctx, project.ID, nil, "Task C - Depends on B", "Task blocked by B", 5, types.TaskPriorityMedium, "TestUser")
	require.NoError(t, err)
	_, err = projectManager.AddTaskDependency(ctx, taskC.ID, taskB.ID, "TestUser")
	require.NoError(t, err)

	// Task D (pending) - depends on B and A (should be blocked because B is not completed)
	taskD, err := projectManager.CreateTask(ctx, project.ID, nil, "Task D - Multiple Dependencies", "Task with multiple dependencies", 6, types.TaskPriorityMedium, "TestUser")
	require.NoError(t, err)
	_, err = projectManager.AddTaskDependency(ctx, taskD.ID, taskA.ID, "TestUser")
	require.NoError(t, err)
	_, err = projectManager.AddTaskDependency(ctx, taskD.ID, taskB.ID, "TestUser")
	require.NoError(t, err)

	// Task E (pending) - no dependencies (should not be blocked)
	taskE, err := projectManager.CreateTask(ctx, project.ID, nil, "Task E - Independent", "Independent task", 2, types.TaskPriorityMedium, "TestUser")
	require.NoError(t, err)

	// Get all tasks to create task map
	allTasks, err := projectManager.ListTasksForProject(ctx, project.ID)
	require.NoError(t, err)

	taskMap := make(map[uuid.UUID]*types.Task)
	for _, task := range allTasks {
		taskMap[task.ID] = task
	}

	// Test isTaskReady function
	t.Run("Test isTaskReady function", func(t *testing.T) {
		// Task A should be ready (completed, no dependencies)
		assert.True(t, utils.IsTaskReady(taskMap[taskA.ID], taskMap), "Task A should be ready")

		// Task B should be ready (depends on completed A)
		assert.True(t, utils.IsTaskReady(taskMap[taskB.ID], taskMap), "Task B should be ready")

		// Task C should NOT be ready (depends on in-progress B)
		assert.False(t, utils.IsTaskReady(taskMap[taskC.ID], taskMap), "Task C should be blocked")

		// Task D should NOT be ready (depends on in-progress B)
		assert.False(t, utils.IsTaskReady(taskMap[taskD.ID], taskMap), "Task D should be blocked")

		// Task E should be ready (no dependencies)
		assert.True(t, utils.IsTaskReady(taskMap[taskE.ID], taskMap), "Task E should be ready")
	})

	// Test blocked task identification
	t.Run("Test blocked task identification", func(t *testing.T) {
		var blockedTasks []*types.Task
		for _, task := range allTasks {
			if task.State == types.TaskStatePending || task.State == types.TaskStateInProgress {
				if !utils.IsTaskReady(task, taskMap) && len(task.Dependencies) > 0 {
					blockedTasks = append(blockedTasks, task)
				}
			}
		}

		// Should find exactly 2 blocked tasks (C and D)
		assert.Len(t, blockedTasks, 2, "Should find exactly 2 blocked tasks")

		// Verify the blocked tasks are C and D
		blockedIDs := make(map[uuid.UUID]bool)
		for _, task := range blockedTasks {
			blockedIDs[task.ID] = true
		}

		assert.True(t, blockedIDs[taskC.ID], "Task C should be in blocked list")
		assert.True(t, blockedIDs[taskD.ID], "Task D should be in blocked list")
		assert.False(t, blockedIDs[taskA.ID], "Task A should not be in blocked list")
		assert.False(t, blockedIDs[taskB.ID], "Task B should not be in blocked list")
		assert.False(t, blockedIDs[taskE.ID], "Task E should not be in blocked list")
	})

	// Test blocking reasons
	t.Run("Test blocking reasons", func(t *testing.T) {
		// Task C is blocked by Task B (in-progress)
		taskCData := taskMap[taskC.ID]
		assert.Len(t, taskCData.Dependencies, 1, "Task C should have 1 dependency")

		blockingTask := taskMap[taskCData.Dependencies[0]]
		assert.Equal(t, taskB.ID, blockingTask.ID, "Task C should be blocked by Task B")
		assert.Equal(t, types.TaskStateInProgress, blockingTask.State, "Blocking task should be in-progress")

		// Task D is blocked by Task B (in-progress), even though A is completed
		taskDData := taskMap[taskD.ID]
		assert.Len(t, taskDData.Dependencies, 2, "Task D should have 2 dependencies")

		// Check that at least one dependency is not completed
		hasIncompleteDepencency := false
		for _, depID := range taskDData.Dependencies {
			depTask := taskMap[depID]
			if depTask.State != types.TaskStateCompleted {
				hasIncompleteDepencency = true
				break
			}
		}
		assert.True(t, hasIncompleteDepencency, "Task D should have at least one incomplete dependency")
	})
}

// TestBlockedCommandEdgeCases tests edge cases for the blocked command
func TestBlockedCommandEdgeCases(t *testing.T) {
	// Setup test environment
	repo := inmemory.NewMemoryRepository()
	config := manager.DefaultConfig()
	projectManager := manager.NewManagerWithRepository(repo, config)

	ctx := context.Background()

	// Create a test project
	project, err := projectManager.CreateProject(ctx, "Edge Case Project", "Project for testing edge cases", "TestUser")
	require.NoError(t, err)

	t.Run("No tasks in project", func(t *testing.T) {
		// Test with empty project
		allTasks, err := projectManager.ListTasksForProject(ctx, project.ID)
		require.NoError(t, err)
		assert.Len(t, allTasks, 0, "Project should have no tasks")

		// Should handle empty task list gracefully
		taskMap := make(map[uuid.UUID]*types.Task)
		var blockedTasks []*types.Task

		// This should not panic
		for _, task := range allTasks {
			if task.State == types.TaskStatePending || task.State == types.TaskStateInProgress {
				if !utils.IsTaskReady(task, taskMap) && len(task.Dependencies) > 0 {
					blockedTasks = append(blockedTasks, task)
				}
			}
		}

		assert.Len(t, blockedTasks, 0, "Should find no blocked tasks in empty project")
	})

	// Create some test tasks for further edge case testing
	taskA, err := projectManager.CreateTask(ctx, project.ID, nil, "Task A", "Test task A", 3, types.TaskPriorityMedium, "TestUser")
	require.NoError(t, err)

	taskB, err := projectManager.CreateTask(ctx, project.ID, nil, "Task B", "Test task B", 4, types.TaskPriorityMedium, "TestUser")
	require.NoError(t, err)

	t.Run("Task with missing dependency", func(t *testing.T) {
		// Create a task with a dependency that doesn't exist
		taskWithMissingDep := &types.Task{
			ID:           uuid.New(),
			ProjectID:    project.ID,
			Title:        "Task with missing dep",
			State:        types.TaskStatePending,
			Dependencies: []uuid.UUID{uuid.New()}, // Non-existent dependency
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Create task map with our tasks
		taskMap := map[uuid.UUID]*types.Task{
			taskA.ID:              taskA,
			taskB.ID:              taskB,
			taskWithMissingDep.ID: taskWithMissingDep,
		}

		// Task with missing dependency should not be ready
		assert.False(t, utils.IsTaskReady(taskWithMissingDep, taskMap), "Task with missing dependency should not be ready")
	})

	t.Run("Task with self-dependency", func(t *testing.T) {
		// Create a task that depends on itself (should not happen in practice, but test robustness)
		selfDepTask := &types.Task{
			ID:           uuid.New(),
			ProjectID:    project.ID,
			Title:        "Self-dependent task",
			State:        types.TaskStatePending,
			Dependencies: []uuid.UUID{}, // Will add self after creation
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		selfDepTask.Dependencies = []uuid.UUID{selfDepTask.ID}

		taskMap := map[uuid.UUID]*types.Task{
			selfDepTask.ID: selfDepTask,
		}

		// Self-dependent task should not be ready (depends on itself which is not completed)
		assert.False(t, utils.IsTaskReady(selfDepTask, taskMap), "Self-dependent task should not be ready")
	})

	t.Run("Completed tasks should not be considered blocked", func(t *testing.T) {
		// Complete task A
		_, err = projectManager.UpdateTaskState(ctx, taskA.ID, types.TaskStateInProgress, "TestUser")
		require.NoError(t, err)
		_, err = projectManager.UpdateTaskState(ctx, taskA.ID, types.TaskStateCompleted, "TestUser")
		require.NoError(t, err)

		// Add dependency from B to A
		_, err = projectManager.AddTaskDependency(ctx, taskB.ID, taskA.ID, "TestUser")
		require.NoError(t, err)

		// Complete task B as well
		_, err = projectManager.UpdateTaskState(ctx, taskB.ID, types.TaskStateInProgress, "TestUser")
		require.NoError(t, err)
		_, err = projectManager.UpdateTaskState(ctx, taskB.ID, types.TaskStateCompleted, "TestUser")
		require.NoError(t, err)

		// Get updated tasks
		allTasks, err := projectManager.ListTasksForProject(ctx, project.ID)
		require.NoError(t, err)

		taskMap := make(map[uuid.UUID]*types.Task)
		for _, task := range allTasks {
			taskMap[task.ID] = task
		}

		// Find blocked tasks
		var blockedTasks []*types.Task
		for _, task := range allTasks {
			if task.State == types.TaskStatePending || task.State == types.TaskStateInProgress {
				if !utils.IsTaskReady(task, taskMap) && len(task.Dependencies) > 0 {
					blockedTasks = append(blockedTasks, task)
				}
			}
		}

		// Should find no blocked tasks since all are completed
		assert.Len(t, blockedTasks, 0, "Should find no blocked tasks when all tasks are completed")
	})
}

// TestBlockedCommandDependencyChains tests complex dependency chains
func TestBlockedCommandDependencyChains(t *testing.T) {
	// Setup test environment
	repo := inmemory.NewMemoryRepository()
	config := manager.DefaultConfig()
	projectManager := manager.NewManagerWithRepository(repo, config)

	ctx := context.Background()

	// Create a test project
	project, err := projectManager.CreateProject(ctx, "Dependency Chain Project", "Project for testing dependency chains", "TestUser")
	require.NoError(t, err)

	// Create a chain: A -> B -> C -> D
	// Where A is completed, B is in-progress, C and D are pending
	taskA, err := projectManager.CreateTask(ctx, project.ID, nil, "Task A - Root", "Root task", 2, types.TaskPriorityMedium, "TestUser")
	require.NoError(t, err)
	_, err = projectManager.UpdateTaskState(ctx, taskA.ID, types.TaskStateInProgress, "TestUser")
	require.NoError(t, err)
	_, err = projectManager.UpdateTaskState(ctx, taskA.ID, types.TaskStateCompleted, "TestUser")
	require.NoError(t, err)

	taskB, err := projectManager.CreateTask(ctx, project.ID, nil, "Task B - Level 1", "Depends on A", 3, types.TaskPriorityMedium, "TestUser")
	require.NoError(t, err)
	_, err = projectManager.AddTaskDependency(ctx, taskB.ID, taskA.ID, "TestUser")
	require.NoError(t, err)
	_, err = projectManager.UpdateTaskState(ctx, taskB.ID, types.TaskStateInProgress, "TestUser")
	require.NoError(t, err)

	taskC, err := projectManager.CreateTask(ctx, project.ID, nil, "Task C - Level 2", "Depends on B", 4, types.TaskPriorityMedium, "TestUser")
	require.NoError(t, err)
	_, err = projectManager.AddTaskDependency(ctx, taskC.ID, taskB.ID, "TestUser")
	require.NoError(t, err)

	taskD, err := projectManager.CreateTask(ctx, project.ID, nil, "Task D - Level 3", "Depends on C", 5, types.TaskPriorityMedium, "TestUser")
	require.NoError(t, err)
	_, err = projectManager.AddTaskDependency(ctx, taskD.ID, taskC.ID, "TestUser")
	require.NoError(t, err)

	// Get all tasks
	allTasks, err := projectManager.ListTasksForProject(ctx, project.ID)
	require.NoError(t, err)

	taskMap := make(map[uuid.UUID]*types.Task)
	for _, task := range allTasks {
		taskMap[task.ID] = task
	}

	t.Run("Test dependency chain blocking", func(t *testing.T) {
		// Task A: completed, no dependencies -> ready
		assert.True(t, utils.IsTaskReady(taskMap[taskA.ID], taskMap), "Task A should be ready")

		// Task B: in-progress, depends on completed A -> ready
		assert.True(t, utils.IsTaskReady(taskMap[taskB.ID], taskMap), "Task B should be ready")

		// Task C: pending, depends on in-progress B -> blocked
		assert.False(t, utils.IsTaskReady(taskMap[taskC.ID], taskMap), "Task C should be blocked")

		// Task D: pending, depends on pending C -> blocked
		assert.False(t, utils.IsTaskReady(taskMap[taskD.ID], taskMap), "Task D should be blocked")
	})

	t.Run("Test chain unblocking", func(t *testing.T) {
		// Complete task B
		_, err = projectManager.UpdateTaskState(ctx, taskB.ID, types.TaskStateInProgress, "TestUser")
		require.NoError(t, err)
		_, err = projectManager.UpdateTaskState(ctx, taskB.ID, types.TaskStateCompleted, "TestUser")
		require.NoError(t, err)

		// Refresh task map
		allTasks, err = projectManager.ListTasksForProject(ctx, project.ID)
		require.NoError(t, err)
		for _, task := range allTasks {
			taskMap[task.ID] = task
		}

		// Now Task C should be ready
		assert.True(t, utils.IsTaskReady(taskMap[taskC.ID], taskMap), "Task C should be ready after B is completed")

		// But Task D should still be blocked (depends on pending C)
		assert.False(t, utils.IsTaskReady(taskMap[taskD.ID], taskMap), "Task D should still be blocked")

		// Complete task C
		_, err = projectManager.UpdateTaskState(ctx, taskC.ID, types.TaskStateInProgress, "TestUser")
		require.NoError(t, err)
		_, err = projectManager.UpdateTaskState(ctx, taskC.ID, types.TaskStateCompleted, "TestUser")
		require.NoError(t, err)

		// Refresh task map
		allTasks, err = projectManager.ListTasksForProject(ctx, project.ID)
		require.NoError(t, err)
		for _, task := range allTasks {
			taskMap[task.ID] = task
		}

		// Now Task D should be ready
		assert.True(t, utils.IsTaskReady(taskMap[taskD.ID], taskMap), "Task D should be ready after C is completed")
	})
}
