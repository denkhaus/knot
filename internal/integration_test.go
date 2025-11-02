package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/denkhaus/knot/internal/manager"
	"github.com/denkhaus/knot/internal/repository/sqlite"
	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupIntegrationTest creates a complete test environment with repository and manager
func setupIntegrationTest(t *testing.T) (manager.ProjectManager, func()) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "knot_integration_test_*")
	require.NoError(t, err)
	
	dbPath := filepath.Join(tempDir, "integration_test.db")
	
	// Create repository
	repo, err := sqlite.NewRepository(dbPath,
		sqlite.WithAutoMigrate(true),
		sqlite.WithLogger(zap.NewNop()), // Silent logger for tests
	)
	require.NoError(t, err)
	
	// Create manager with repository
	mgr := manager.NewManagerWithRepository(repo, manager.DefaultConfig())
	
	// Return cleanup function
	cleanup := func() {
		if closer, ok := repo.(interface{ Close() error }); ok {
			closer.Close()
		}
		os.RemoveAll(tempDir)
	}
	
	return mgr, cleanup
}

// TestCompleteProjectWorkflow tests the entire project lifecycle
func TestCompleteProjectWorkflow(t *testing.T) {
	mgr, cleanup := setupIntegrationTest(t)
	defer cleanup()
	
	ctx := context.Background()
	actor := "integration-test-user"
	
	t.Run("complete project lifecycle", func(t *testing.T) {
		// Step 1: Create a project
		project, err := mgr.CreateProject(ctx, "Integration Test Project", "A project for testing complete workflows", actor)
		require.NoError(t, err)
		projectID := project.ID
		
		// Step 2: Verify project was created
		retrievedProject, err := mgr.GetProject(ctx, projectID)
		require.NoError(t, err)
		assert.Equal(t, "Integration Test Project", retrievedProject.Title)
		assert.Equal(t, "A project for testing complete workflows", retrievedProject.Description)
		
		// Step 3: Create root tasks
		rootTask1, err := mgr.CreateTask(ctx, projectID, nil, "Root Task 1: Setup", "Initial setup task", 3, types.TaskPriorityMedium, actor)
		require.NoError(t, err)
		rootTask1ID := rootTask1.ID
		
		rootTask2, err := mgr.CreateTask(ctx, projectID, nil, "Root Task 2: Implementation", "Main implementation task", 5, types.TaskPriorityMedium, actor)
		require.NoError(t, err)
		rootTask2ID := rootTask2.ID
		
		// Step 4: Create subtasks
		subtask1, err := mgr.CreateTask(ctx, projectID, &rootTask1ID, "Subtask 1.1: Configuration", "Configure the system", 2, types.TaskPriorityMedium, actor)
		require.NoError(t, err)
		subtask1ID := subtask1.ID
		
		subtask2, err := mgr.CreateTask(ctx, projectID, &rootTask2ID, "Subtask 2.1: Core Logic", "Implement core business logic", 4, types.TaskPriorityMedium, actor)
		require.NoError(t, err)
		subtask2ID := subtask2.ID
		
		// Step 5: Add task dependencies
		_, err = mgr.AddTaskDependency(ctx, rootTask2ID, rootTask1ID, actor)
		require.NoError(t, err, "Root Task 2 should depend on Root Task 1")
		
		_, err = mgr.AddTaskDependency(ctx, subtask2ID, subtask1ID, actor)
		require.NoError(t, err, "Subtask 2.1 should depend on Subtask 1.1")
		
		// Step 6: Verify task hierarchy
		rootTasks, err := mgr.GetRootTasks(ctx, projectID)
		require.NoError(t, err)
		assert.Len(t, rootTasks, 2, "Should have 2 root tasks")
		
		// Step 7: Test task state transitions
		// Start with subtask1 (no dependencies)
		_, err = mgr.UpdateTaskState(ctx, subtask1ID, types.TaskStateInProgress, actor)
		require.NoError(t, err)
		
		_, err = mgr.UpdateTaskState(ctx, subtask1ID, types.TaskStateCompleted, actor)
		require.NoError(t, err)
		
		// Now start rootTask1
		_, err = mgr.UpdateTaskState(ctx, rootTask1ID, types.TaskStateInProgress, actor)
		require.NoError(t, err)
		
		_, err = mgr.UpdateTaskState(ctx, rootTask1ID, types.TaskStateCompleted, actor)
		require.NoError(t, err)
		
		// Step 8: Verify project progress
		progress, err := mgr.GetProjectProgress(ctx, projectID)
		require.NoError(t, err)
		assert.Equal(t, 4, progress.TotalTasks)
		assert.Equal(t, 2, progress.CompletedTasks)
		assert.Equal(t, 50.0, progress.OverallProgress)
		
		// Step 9: Complete remaining tasks
		_, err = mgr.UpdateTaskState(ctx, subtask2ID, types.TaskStateInProgress, actor)
		require.NoError(t, err)
		
		_, err = mgr.UpdateTaskState(ctx, subtask2ID, types.TaskStateCompleted, actor)
		require.NoError(t, err)
		
		_, err = mgr.UpdateTaskState(ctx, rootTask2ID, types.TaskStateInProgress, actor)
		require.NoError(t, err)
		
		_, err = mgr.UpdateTaskState(ctx, rootTask2ID, types.TaskStateCompleted, actor)
		require.NoError(t, err)
		
		// Step 10: Verify final project state
		finalProgress, err := mgr.GetProjectProgress(ctx, projectID)
		require.NoError(t, err)
		assert.Equal(t, 4, finalProgress.TotalTasks)
		assert.Equal(t, 4, finalProgress.CompletedTasks)
		assert.Equal(t, 100.0, finalProgress.OverallProgress)
		
		// Step 11: Update project to completed
		_, err = mgr.UpdateProjectState(ctx, projectID, types.ProjectStateCompleted, actor)
		require.NoError(t, err)
		
		// Step 12: Verify final project state
		finalProject, err := mgr.GetProject(ctx, projectID)
		require.NoError(t, err)
		assert.Equal(t, types.ProjectStateCompleted, finalProject.State)
	})
}

// TestDependencyWorkflow tests complex dependency scenarios
func TestDependencyWorkflow(t *testing.T) {
	mgr, cleanup := setupIntegrationTest(t)
	defer cleanup()
	
	ctx := context.Background()
	actor := "dependency-test-user"
	
	t.Run("complex dependency chain", func(t *testing.T) {
		// Create project
		project, err := mgr.CreateProject(ctx, "Dependency Test Project", "Testing complex dependency chains", actor)
		require.NoError(t, err)
		projectID := project.ID
		
		// Create tasks: A -> B -> C (linear dependency chain)
		taskA, err := mgr.CreateTask(ctx, projectID, nil, "Task A: Foundation", "Foundation task", 2, types.TaskPriorityMedium, actor)
		require.NoError(t, err)
		taskAID := taskA.ID
		
		taskB, err := mgr.CreateTask(ctx, projectID, nil, "Task B: Building", "Building on foundation", 3, types.TaskPriorityMedium, actor)
		require.NoError(t, err)
		taskBID := taskB.ID
		
		taskC, err := mgr.CreateTask(ctx, projectID, nil, "Task C: Completion", "Final completion task", 2, types.TaskPriorityMedium, actor)
		require.NoError(t, err)
		taskCID := taskC.ID
		
		// Set up dependency chain: C depends on B, B depends on A
		_, err = mgr.AddTaskDependency(ctx, taskBID, taskAID, actor)
		require.NoError(t, err)
		
		_, err = mgr.AddTaskDependency(ctx, taskCID, taskBID, actor)
		require.NoError(t, err)
		
		// Verify dependencies
		dependencies, err := mgr.GetTaskDependencies(ctx, taskCID)
		require.NoError(t, err)
		assert.Len(t, dependencies, 1)
		assert.Equal(t, taskBID, dependencies[0].ID)
		
		// Test actionable task logic
		// Initially, only Task A should be actionable
		actionableTask, err := mgr.FindNextActionableTask(ctx, projectID)
		require.NoError(t, err)
		assert.Equal(t, taskAID, actionableTask.ID)
		
		// Complete Task A
		_, err = mgr.UpdateTaskState(ctx, taskAID, types.TaskStateInProgress, actor)
		require.NoError(t, err)
		
		_, err = mgr.UpdateTaskState(ctx, taskAID, types.TaskStateCompleted, actor)
		require.NoError(t, err)
		
		// Now Task B should be actionable
		actionableTask, err = mgr.FindNextActionableTask(ctx, projectID)
		require.NoError(t, err)
		assert.Equal(t, taskBID, actionableTask.ID)
		
		// Complete Task B
		_, err = mgr.UpdateTaskState(ctx, taskBID, types.TaskStateInProgress, actor)
		require.NoError(t, err)
		
		_, err = mgr.UpdateTaskState(ctx, taskBID, types.TaskStateCompleted, actor)
		require.NoError(t, err)
		
		// Now Task C should be actionable
		actionableTask, err = mgr.FindNextActionableTask(ctx, projectID)
		require.NoError(t, err)
		assert.Equal(t, taskCID, actionableTask.ID)
		
		// Remove a dependency and verify
		_, err = mgr.RemoveTaskDependency(ctx, taskCID, taskBID, actor)
		require.NoError(t, err)
		
		dependencies, err = mgr.GetTaskDependencies(ctx, taskCID)
		require.NoError(t, err)
		assert.Len(t, dependencies, 0)
	})
}

// TestErrorScenarios tests error handling in integration scenarios
func TestErrorScenarios(t *testing.T) {
	mgr, cleanup := setupIntegrationTest(t)
	defer cleanup()
	
	ctx := context.Background()
	actor := "error-test-user"
	
	t.Run("error handling scenarios", func(t *testing.T) {
		// Test creating task with non-existent project
		nonExistentProjectID := uuid.New()
		_, err := mgr.CreateTask(ctx, nonExistentProjectID, nil, "Invalid Task", "Task with non-existent project", 1, types.TaskPriorityMedium, actor)
		assert.Error(t, err, "Should fail to create task with non-existent project")
		
		// Create valid project and task for further tests
		project, err := mgr.CreateProject(ctx, "Error Test Project", "Project for error testing", actor)
		require.NoError(t, err)
		projectID := project.ID
		
		task, err := mgr.CreateTask(ctx, projectID, nil, "Valid Task", "A valid task for testing", 2, types.TaskPriorityMedium, actor)
		require.NoError(t, err)
		taskID := task.ID
		
		// Test circular dependency
		_, err = mgr.AddTaskDependency(ctx, taskID, taskID, actor)
		assert.Error(t, err, "Should prevent self-dependency")
		
		// Test invalid state transition
		_, err = mgr.UpdateTaskState(ctx, taskID, types.TaskStateCompleted, actor)
		assert.Error(t, err, "Should prevent invalid state transition from pending to completed")
		
		// Test operations on non-existent task
		nonExistentTaskID := uuid.New()
		_, err = mgr.UpdateTaskState(ctx, nonExistentTaskID, types.TaskStateInProgress, actor)
		assert.Error(t, err, "Should fail to update non-existent task")
		
		_, err = mgr.GetTask(ctx, nonExistentTaskID)
		assert.Error(t, err, "Should fail to get non-existent task")
	})
}

// TestConcurrentOperations tests concurrent access patterns
func TestConcurrentOperations(t *testing.T) {
	mgr, cleanup := setupIntegrationTest(t)
	defer cleanup()
	
	ctx := context.Background()
	actor := "concurrent-test-user"
	
	t.Run("concurrent project and task operations", func(t *testing.T) {
		// Create a project
		project, err := mgr.CreateProject(ctx, "Concurrent Test Project", "Project for concurrent testing", actor)
		require.NoError(t, err)
		projectID := project.ID
		
		// Create multiple tasks concurrently
		const numTasks = 5
		errChan := make(chan error, numTasks)
		
		for i := 0; i < numTasks; i++ {
			go func(index int) {
				_, err := mgr.CreateTask(ctx, projectID, nil, fmt.Sprintf("Concurrent Task %d", index), fmt.Sprintf("Task created concurrently %d", index), 1, types.TaskPriorityMedium, actor)
				errChan <- err
			}(i)
		}
		
		// Check results
		successCount := 0
		for i := 0; i < numTasks; i++ {
			err := <-errChan
			if err == nil {
				successCount++
			}
		}
		
		// Verify at least some tasks were created successfully
		assert.Greater(t, successCount, 0, "At least some concurrent task creations should succeed")
		
		// Verify tasks exist
		tasks, err := mgr.ListTasksForProject(ctx, projectID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(tasks), successCount)
	})
}