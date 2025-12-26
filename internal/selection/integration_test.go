package selection

import (
	"fmt"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPackageIntegration tests the full package integration
func TestPackageIntegration(t *testing.T) {
	t.Run("CompleteProjectWorkflow", func(t *testing.T) {
		// Simulate working through a complete project
		tasks := createComplexDependencyGraphIntegration()

		config := DefaultConfig()
		selector, err := NewTaskSelector(config)
		require.NoError(t, err)

		completionOrder := make([]string, 0)

		// Simulate working through all tasks
		for len(completionOrder) < len(tasks) {
			selectedTask, err := selector.SelectNextActionableTask(tasks)
			if err != nil {
				break // No more actionable tasks
			}

			// Mark task as completed
			for _, task := range tasks {
				if task.ID == selectedTask.ID {
					task.State = types.TaskStateCompleted
					completionOrder = append(completionOrder, task.Title)
					break
				}
			}
		}

		t.Logf("Task completion order: %v", completionOrder)

		// Verify that dependencies were respected
		// Create maps by both ID and Title for easy lookup
		taskMapByID := make(map[string]*types.Task)
		taskMapByTitle := make(map[string]*types.Task)
		for _, task := range tasks {
			taskMapByID[task.ID.String()] = task
			taskMapByTitle[task.Title] = task
		}

		for i, title := range completionOrder {
			completedTask := taskMapByTitle[title]
			for _, depID := range completedTask.Dependencies {
				depTask := taskMapByID[depID.String()]
				if depTask == nil {
					t.Errorf("Task %s has dependency %s that is not in the task list", title, depID)
					continue
				}
				// Find when this dependency was completed
				depCompletedAt := -1
				for j, completedTitle := range completionOrder {
					if completedTitle == depTask.Title {
						depCompletedAt = j
						break
					}
				}
				if depCompletedAt >= i {
					t.Errorf("Task %s was completed before its dependency %s", title, depTask.Title)
				}
			}
		}

		// Verify no cycles were detected
		assert.True(t, len(completionOrder) > 0, "Should complete at least some tasks")
	})

	t.Run("PerformanceMetrics", func(t *testing.T) {
		tasks := createLargeTaskSet(100)

		config := DefaultConfig()
		selector, err := NewTaskSelector(config)
		require.NoError(t, err)

		startTime := time.Now()
		task, err := selector.SelectNextActionableTask(tasks)
		elapsed := time.Since(startTime)

		require.NoError(t, err)
		assert.NotNil(t, task)

		t.Logf("Selected task from 100 tasks in %v", elapsed)

		// Performance assertion - should complete within reasonable time
		assert.Less(t, elapsed.Milliseconds(), int64(100),
			"Task selection from 100 tasks should complete in less than 100ms")
	})
}

// createUserScenarioTasksIntegration creates tasks matching the user's reported scenario
func createUserScenarioTasksIntegration() []*types.Task {
	now := time.Now()

	firstRoot := &types.Task{
		ID:          uuid.New(),
		Title:       "First Root Task",
		Description: "First root task in the project",
		State:       types.TaskStatePending,
		Priority:    2,
		Depth:       0,
		CreatedAt:   now,
	}

	firstSubtask := &types.Task{
		ID:          uuid.New(),
		Title:       "First Subtask of First Root",
		Description: "First subtask that needs to complete before parent",
		State:       types.TaskStatePending,
		Priority:    2,
		Depth:       1,
		ParentID:    &firstRoot.ID,
		CreatedAt:   now.Add(time.Millisecond),
	}

	secondRoot := &types.Task{
		ID:          uuid.New(),
		Title:       "Second Root Task",
		Description: "Second root task (should not be selected before first root's subtasks)",
		State:       types.TaskStatePending,
		Priority:    2,
		Depth:       0,
		CreatedAt:   now.Add(2 * time.Millisecond),
	}

	return []*types.Task{firstRoot, firstSubtask, secondRoot}
}

// createComplexDependencyGraphIntegration creates a more complex task graph for testing
func createComplexDependencyGraphIntegration() []*types.Task {
	now := time.Now()

	// Create a diamond dependency pattern
	//     A
	//    / \
	//   B   C
	//    \ /
	//     D

	taskA := &types.Task{
		ID:        uuid.New(),
		Title:     "Task A",
		State:     types.TaskStatePending,
		Priority:  2,
		Depth:     0,
		CreatedAt: now,
	}

	taskB := &types.Task{
		ID:           uuid.New(),
		Title:        "Task B",
		State:        types.TaskStatePending,
		Priority:     2,
		Depth:        1,
		Dependencies: []uuid.UUID{taskA.ID},
		CreatedAt:    now.Add(time.Millisecond),
	}

	taskC := &types.Task{
		ID:           uuid.New(),
		Title:        "Task C",
		State:        types.TaskStatePending,
		Priority:     2,
		Depth:        1,
		Dependencies: []uuid.UUID{taskA.ID},
		CreatedAt:    now.Add(2 * time.Millisecond),
	}

	taskD := &types.Task{
		ID:           uuid.New(),
		Title:        "Task D",
		State:        types.TaskStatePending,
		Priority:     2,
		Depth:        2,
		Dependencies: []uuid.UUID{taskB.ID, taskC.ID},
		CreatedAt:    now.Add(3 * time.Millisecond),
	}

	return []*types.Task{taskA, taskB, taskC, taskD}
}

// createLargeTaskSet creates a large set of tasks for performance testing
func createLargeTaskSet(count int) []*types.Task {
	tasks := make([]*types.Task, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		tasks[i] = &types.Task{
			ID:        uuid.New(),
			Title:     fmt.Sprintf("Task %d", i),
			State:     types.TaskStatePending,
			Priority:  types.TaskPriorityMedium,
			Depth:     0,
			CreatedAt: now.Add(time.Duration(i) * time.Millisecond),
		}
	}

	return tasks
}
