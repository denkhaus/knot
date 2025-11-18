package selection

import (
	"fmt"
	"testing"
	"time"

	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPackageIntegration tests the full package integration
func TestPackageIntegration(t *testing.T) {
	t.Run("EndToEndUserScenario", func(t *testing.T) {
		// This test specifically addresses the user's reported issue:
		// "actionable task is not the first subtask of the first root task but the second root task"

		tasks := createUserScenarioTasks()

		// Test with current-style logic (creation order)
		currentTask, err := SelectActionableTask(tasks, StrategyCreationOrder)
		require.NoError(t, err)

		// Test with improved logic (dependency-aware)
		improvedTask, err := SelectActionableTask(tasks, StrategyDependencyAware)
		require.NoError(t, err)

		// Test with depth-first logic
		depthFirstTask, err := SelectActionableTask(tasks, StrategyDepthFirst)
		require.NoError(t, err)

		t.Logf("Current logic selected: %s", currentTask.Title)
		t.Logf("Dependency-aware selected: %s", improvedTask.Title)
		t.Logf("Depth-first selected: %s", depthFirstTask.Title)

		// Depth-first should definitely select the subtask
		assert.Equal(t, "First Subtask of First Root", depthFirstTask.Title,
			"Depth-first strategy should prioritize completing subtasks first")

		// The improved logic should NOT select "Second Root Task"
		assert.NotEqual(t, "Second Root Task", improvedTask.Title,
			"Dependency-aware strategy should not select second root task before first root's subtasks")
	})

	t.Run("CompleteProjectWorkflow", func(t *testing.T) {
		// Simulate working through a complete project
		tasks := createComplexDependencyGraph()

		config := DefaultConfig()
		selector, err := NewTaskSelector(StrategyDependencyAware, config)
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
		taskMap := make(map[string]*types.Task)
		for _, task := range tasks {
			taskMap[task.Title] = task
		}

		for i, taskName := range completionOrder {
			task := taskMap[taskName]

			// Check that all dependencies were completed before this task
			for _, depID := range task.Dependencies {
				var depTask *types.Task
				for _, t := range tasks {
					if t.ID == depID {
						depTask = t
						break
					}
				}
				if depTask != nil {
					depIndex := findTaskIndex(depTask.Title, completionOrder)
					assert.Less(t, depIndex, i,
						"Dependency %s should be completed before %s", depTask.Title, taskName)
				}
			}
		}
	})

	t.Run("MultipleStrategyComparison", func(t *testing.T) {
		tasks := createLargeProjectTasks(20)

		strategies := []Strategy{
			StrategyCreationOrder,
			StrategyDependencyAware,
			StrategyPriority,
			StrategyDepthFirst,
			StrategyCriticalPath,
		}

		results := make(map[Strategy]*types.Task)
		executionTimes := make(map[Strategy]time.Duration)

		for _, strategy := range strategies {
			selector, err := NewTaskSelector(strategy, DefaultConfig())
			require.NoError(t, err)

			start := time.Now()
			selectedTask, err := selector.SelectNextActionableTask(tasks)
			duration := time.Since(start)

			require.NoError(t, err)
			require.NotNil(t, selectedTask)

			results[strategy] = selectedTask
			executionTimes[strategy] = duration

			t.Logf("Strategy %s selected: %s (took %v)",
				strategy.String(), selectedTask.Title, duration)
		}

		// All strategies should be reasonably fast
		for strategy, duration := range executionTimes {
			assert.Less(t, duration, 100*time.Millisecond,
				"Strategy %s took too long: %v", strategy.String(), duration)
		}

		// Strategies should potentially select different tasks (though not guaranteed)
		uniqueSelections := make(map[string]bool)
		for _, task := range results {
			uniqueSelections[task.Title] = true
		}

		t.Logf("Number of unique task selections: %d", len(uniqueSelections))
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// Test various error conditions

		// Circular dependencies
		circularTasks := createCircularDependencyTasks()
		_, err := SelectActionableTask(circularTasks, StrategyDependencyAware)
		assert.Error(t, err)
		selErr, ok := err.(*SelectionError)
		require.True(t, ok)
		assert.Equal(t, ErrorTypeCircularDep, selErr.Type)

		// No actionable tasks
		blockedTasks := []*types.Task{
			createTestTask("blocked", "Blocked Task", types.TaskStatePending, 5, nil,
				[]uuid.UUID{testTaskID("nonexistent")}),
		}
		_, err = SelectActionableTask(blockedTasks, StrategyDependencyAware)
		assert.Error(t, err)

		// Empty task list
		_, err = SelectActionableTask([]*types.Task{}, StrategyDependencyAware)
		assert.Error(t, err)
	})

	t.Run("ProjectAnalysisAndRecommendations", func(t *testing.T) {
		// Test project analysis and strategy recommendation
		tasks := createComplexDependencyGraph()

		strategy, reason, err := AnalyzeProjectAndRecommendStrategy(tasks)
		require.NoError(t, err)

		t.Logf("Recommended strategy: %s", strategy.String())
		t.Logf("Reason: %s", reason)

		assert.NotEmpty(t, reason)
		assert.Contains(t, []Strategy{
			StrategyCreationOrder,
			StrategyDependencyAware,
			StrategyPriority,
			StrategyDepthFirst,
			StrategyCriticalPath,
		}, strategy)

		// Test that the recommended strategy actually works
		selectedTask, err := SelectActionableTask(tasks, strategy)
		require.NoError(t, err)
		require.NotNil(t, selectedTask)
	})

	t.Run("DependencyValidation", func(t *testing.T) {
		// Test dependency validation
		validTasks := createComplexDependencyGraph()
		validationErrors, err := ValidateTaskDependencies(validTasks)
		require.NoError(t, err)
		assert.Empty(t, validationErrors, "Valid tasks should have no validation errors")

		// Test with circular dependencies
		circularTasks := createCircularDependencyTasks()
		validationErrors, err = ValidateTaskDependencies(circularTasks)
		require.NoError(t, err)
		assert.NotEmpty(t, validationErrors, "Circular dependencies should be detected")

		for _, validationError := range validationErrors {
			assert.Equal(t, "circular_dependency", validationError.Type)
		}

		// Test with missing dependencies
		tasksWithMissingDeps := []*types.Task{
			createTestTask("task1", "Task with missing dependency", types.TaskStatePending, 5, nil,
				[]uuid.UUID{testTaskID("nonexistent")}),
		}
		validationErrors, err = ValidateTaskDependencies(tasksWithMissingDeps)
		require.NoError(t, err)
		assert.NotEmpty(t, validationErrors, "Missing dependencies should be detected")

		found := false
		for _, validationError := range validationErrors {
			if validationError.Type == "missing_dependency" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should detect missing dependency validation error")
	})

	t.Run("PerformanceCharacteristics", func(t *testing.T) {
		// Test performance with different project sizes
		sizes := []int{10, 50, 100}

		for _, size := range sizes {
			t.Run(fmt.Sprintf("Size_%d", size), func(t *testing.T) {
				tasks := createLargeProjectTasks(size)

				start := time.Now()
				selectedTask, err := SelectActionableTask(tasks, StrategyDependencyAware)
				duration := time.Since(start)

				require.NoError(t, err)
				require.NotNil(t, selectedTask)

				t.Logf("Size %d: Selected %s in %v", size, selectedTask.Title, duration)

				// Performance should scale reasonably
				maxDuration := time.Duration(size) * time.Millisecond
				assert.Less(t, duration, maxDuration,
					"Performance should scale reasonably with project size")
			})
		}
	})

	t.Run("RealWorldScenarios", func(t *testing.T) {
		// Software development project
		t.Run("SoftwareProject", func(t *testing.T) {
			frontendID := testTaskID("frontend")
			backendID := testTaskID("backend")

			tasks := []*types.Task{
				// Frontend
				createTestTask("frontend", "Frontend Feature", types.TaskStatePending, 3, nil, nil),
				createTestTask("ui-mockup", "UI Mockup", types.TaskStatePending, 5, &frontendID, nil),
				createTestTask("ui-impl", "UI Implementation", types.TaskStatePending, 4, &frontendID,
					[]uuid.UUID{testTaskID("ui-mockup"), testTaskID("api")}),

				// Backend
				createTestTask("backend", "Backend Feature", types.TaskStatePending, 3, nil, nil),
				createTestTask("api", "API Design", types.TaskStatePending, 4, &backendID, nil),
				createTestTask("api-impl", "API Implementation", types.TaskStatePending, 4, &backendID,
					[]uuid.UUID{testTaskID("api"), testTaskID("database")}),

				// Infrastructure
				createTestTask("database", "Database Schema", types.TaskStatePending, 2, nil, nil),

				// Testing
				createTestTask("tests", "Integration Tests", types.TaskStatePending, 5, nil,
					[]uuid.UUID{testTaskID("ui-impl"), testTaskID("api-impl")}),

				// Deployment
				createTestTask("deploy", "Production Deployment", types.TaskStatePending, 1, nil,
					[]uuid.UUID{testTaskID("tests")}),
			}

			selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
			require.NoError(t, err)

			selectedTask, err := selector.SelectNextActionableTask(tasks)
			require.NoError(t, err)

			// Should select a foundational task
			foundationalTasks := []string{"UI Mockup", "API Design", "Database Schema"}
			assert.Contains(t, foundationalTasks, selectedTask.Title,
				"Should select a foundational task that enables other work")

			result := selector.GetLastResult()
			require.NotNil(t, result)
			t.Logf("Selected: %s", selectedTask.Title)
			t.Logf("Reason: %s", result.Reason)
			t.Logf("Score: %.2f", result.Score.Score)

			if len(result.Alternatives) > 0 {
				t.Logf("Alternatives: %d tasks", len(result.Alternatives))
				for i, alt := range result.Alternatives[:min(3, len(result.Alternatives))] {
					t.Logf("  %d. %s (score: %.2f)", i+1, alt.Task.Title, alt.Score)
				}
			}
		})
	})
}

// Helper functions
func findTaskIndex(taskName string, completionOrder []string) int {
	for i, name := range completionOrder {
		if name == taskName {
			return i
		}
	}
	return -1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestPackageAPICompatibility ensures the package API is stable and usable
func TestPackageAPICompatibility(t *testing.T) {
	// Test that all main functions work as expected
	tasks := createUserScenarioTasks()

	t.Run("BasicSelection", func(t *testing.T) {
		task, err := SelectActionableTask(tasks, StrategyDependencyAware)
		assert.NoError(t, err)
		assert.NotNil(t, task)
	})

	t.Run("ConfiguredSelection", func(t *testing.T) {
		config := DefaultConfig()
		config.Behavior.PreferInProgress = false

		task, err := SelectActionableTaskWithConfig(tasks, config)
		assert.NoError(t, err)
		assert.NotNil(t, task)
	})

	t.Run("ProjectAnalysis", func(t *testing.T) {
		strategy, reason, err := AnalyzeProjectAndRecommendStrategy(tasks)
		assert.NoError(t, err)
		assert.NotEmpty(t, reason)
		assert.NotEqual(t, Strategy(999), strategy)
	})

	t.Run("DependencyValidation", func(t *testing.T) {
		errors, err := ValidateTaskDependencies(tasks)
		assert.NoError(t, err)
		assert.Empty(t, errors) // Should be no errors in the test scenario
	})

	t.Run("TaskScoring", func(t *testing.T) {
		scores, err := ScoreTasks(tasks, StrategyDependencyAware, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, scores)

		for _, score := range scores {
			assert.NotNil(t, score.Task)
			assert.GreaterOrEqual(t, score.Score, 0.0)
		}
	})

	t.Run("GraphCreation", func(t *testing.T) {
		graph, err := CreateDependencyGraph(tasks)
		assert.NoError(t, err)
		assert.NotNil(t, graph)
		assert.Equal(t, len(tasks), graph.TaskCount)
	})
}
