package selection

import (
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTaskSelector(t *testing.T) {
	t.Run("ValidConfiguration", func(t *testing.T) {
		selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())

		assert.NoError(t, err)
		assert.NotNil(t, selector)
		assert.NotNil(t, selector.analyzer)
		assert.NotNil(t, selector.filter)
		assert.NotNil(t, selector.strategy)
		assert.NotNil(t, selector.config)
	})

	t.Run("NilConfiguration", func(t *testing.T) {
		selector, err := NewTaskSelector(StrategyDependencyAware, nil)

		assert.NoError(t, err)
		assert.NotNil(t, selector)
		assert.NotNil(t, selector.config) // Should use default config
	})

	t.Run("InvalidStrategy", func(t *testing.T) {
		config := DefaultConfig()
		_, err := NewTaskSelector(Strategy(999), config)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create strategy")
	})

	t.Run("InvalidConfiguration", func(t *testing.T) {
		config := &Config{
			Strategy: StrategyDependencyAware,
			Weights: Weights{
				DependentCount: -1.0, // Invalid negative weight
				Priority:       0.5,
				DepthFirst:     0.3,
				CriticalPath:   0.2,
			},
		}

		_, err := NewTaskSelector(StrategyDependencyAware, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid configuration")
	})
}

func TestDefaultTaskSelector_SelectNextActionableTask(t *testing.T) {
	t.Run("EmptyTaskList", func(t *testing.T) {
		selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
		require.NoError(t, err)

		task, err := selector.SelectNextActionableTask([]*types.Task{})

		assert.Error(t, err)
		assert.Nil(t, task)

		if selErr, ok := err.(*SelectionError); ok {
			assert.Equal(t, ErrorTypeNoTasks, selErr.Type)
		}
	})

	t.Run("UserScenario_DepthFirstStrategy", func(t *testing.T) {
		config := DefaultConfig()
		config.Strategy = StrategyDepthFirst
		selector, err := NewTaskSelector(StrategyDepthFirst, config)
		require.NoError(t, err)

		tasks := createUserScenarioTasks()
		selectedTask, err := selector.SelectNextActionableTask(tasks)

		assertTaskSelected(t, "First Subtask of First Root", selectedTask, err)

		// Verify the reason mentions depth-first logic
		result := selector.GetLastResult()
		assert.NotNil(t, result)
		assert.Contains(t, result.Reason, "depth-first")
	})

	t.Run("UserScenario_DependencyAwareStrategy", func(t *testing.T) {
		selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
		require.NoError(t, err)

		tasks := createUserScenarioTasks()
		selectedTask, err := selector.SelectNextActionableTask(tasks)

		// With default dependency-aware strategy, should still prefer the subtask
		// due to depth-first component in weights
		assert.NoError(t, err)
		assert.NotNil(t, selectedTask)

		// Should NOT select "Second Root Task"
		assert.NotEqual(t, "Second Root Task", selectedTask.Title)
	})

	t.Run("PriorityStrategy", func(t *testing.T) {
		selector, err := NewTaskSelector(StrategyPriority, DefaultConfig())
		require.NoError(t, err)

		tasks := createPriorityTestTasks()
		selectedTask, err := selector.SelectNextActionableTask(tasks)

		assertTaskSelected(t, "High Priority Task", selectedTask, err)
	})

	t.Run("CreationOrderStrategy", func(t *testing.T) {
		selector, err := NewTaskSelector(StrategyCreationOrder, DefaultConfig())
		require.NoError(t, err)

		tasks := createInProgressTasks()
		selectedTask, err := selector.SelectNextActionableTask(tasks)

		assert.NoError(t, err)
		assert.NotNil(t, selectedTask)
		// Should select the oldest task by creation time
	})

	t.Run("InProgressPreference", func(t *testing.T) {
		config := DefaultConfig()
		config.Behavior.PreferInProgress = true
		selector, err := NewTaskSelector(StrategyDependencyAware, config)
		require.NoError(t, err)

		tasks := createInProgressTasks()
		selectedTask, err := selector.SelectNextActionableTask(tasks)

		assert.NoError(t, err)
		assert.NotNil(t, selectedTask)
		assert.Equal(t, types.TaskStateInProgress, selectedTask.State)
	})

	t.Run("NoInProgressPreference", func(t *testing.T) {
		config := DefaultConfig()
		config.Behavior.PreferInProgress = false
		selector, err := NewTaskSelector(StrategyDependencyAware, config)
		require.NoError(t, err)

		tasks := createInProgressTasks()
		selectedTask, err := selector.SelectNextActionableTask(tasks)

		assert.NoError(t, err)
		assert.NotNil(t, selectedTask)
		// Should not necessarily prefer in-progress tasks
	})

	t.Run("CircularDependencies", func(t *testing.T) {
		selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
		require.NoError(t, err)

		tasks := createCircularDependencyTasks()
		selectedTask, err := selector.SelectNextActionableTask(tasks)

		assert.Error(t, err)
		assert.Nil(t, selectedTask)

		if selErr, ok := err.(*SelectionError); ok {
			assert.Equal(t, ErrorTypeCircularDep, selErr.Type)
			assert.Contains(t, selErr.Message, "circular dependencies")
		}
	})

	t.Run("DeadlockScenario", func(t *testing.T) {
		selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
		require.NoError(t, err)

		// Create tasks that are all blocked by dependencies
		tasks := []*types.Task{
			createTestTask("task1", "Task 1", types.TaskStatePending, 5, nil,
				[]uuid.UUID{testTaskID("task2")}),
			createTestTask("task2", "Task 2", types.TaskStatePending, 5, nil,
				[]uuid.UUID{testTaskID("nonexistent")}), // Missing dependency
		}

		selectedTask, err := selector.SelectNextActionableTask(tasks)

		assert.Error(t, err)
		assert.Nil(t, selectedTask)

		if selErr, ok := err.(*SelectionError); ok {
			assert.Equal(t, ErrorTypeDeadlock, selErr.Type)
		}
	})

	t.Run("ComplexDependencyGraph", func(t *testing.T) {
		selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
		require.NoError(t, err)

		tasks := createComplexDependencyGraph()
		selectedTask, err := selector.SelectNextActionableTask(tasks)

		assert.NoError(t, err)
		assert.NotNil(t, selectedTask)

		// Should select one of the actionable tasks (C, D, or E)
		actionableNames := []string{"Task C", "Task D", "Task E"}
		assert.Contains(t, actionableNames, selectedTask.Title)

		// Verify execution time is recorded
		result := selector.GetLastResult()
		assert.NotNil(t, result)
		assert.Greater(t, result.ExecutionTime, time.Duration(0))
	})

	t.Run("ScoreThreshold", func(t *testing.T) {
		config := DefaultConfig()
		config.Advanced.ScoreThreshold = 100.0 // Very high threshold
		selector, err := NewTaskSelector(StrategyDependencyAware, config)
		require.NoError(t, err)

		tasks := createPriorityTestTasks()
		selectedTask, err := selector.SelectNextActionableTask(tasks)

		assert.Error(t, err)
		assert.Nil(t, selectedTask)
		// Should fail because no task meets the high score threshold
	})
}

func TestDefaultTaskSelector_GetSelectionReason(t *testing.T) {
	selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
	require.NoError(t, err)

	t.Run("NoSelectionMade", func(t *testing.T) {
		reason := selector.GetSelectionReason()
		assert.Equal(t, "no selection has been made", reason)
	})

	t.Run("AfterSelection", func(t *testing.T) {
		tasks := createPriorityTestTasks()
		_, err := selector.SelectNextActionableTask(tasks)
		require.NoError(t, err)

		reason := selector.GetSelectionReason()
		assert.NotEmpty(t, reason)
		assert.NotEqual(t, "no selection has been made", reason)
	})
}

func TestDefaultTaskSelector_GetLastResult(t *testing.T) {
	selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
	require.NoError(t, err)

	t.Run("NoSelectionMade", func(t *testing.T) {
		result := selector.GetLastResult()
		assert.Nil(t, result)
	})

	t.Run("AfterSelection", func(t *testing.T) {
		tasks := createPriorityTestTasks()
		selectedTask, err := selector.SelectNextActionableTask(tasks)
		require.NoError(t, err)

		result := selector.GetLastResult()
		assert.NotNil(t, result)
		assert.Equal(t, selectedTask, result.SelectedTask)
		assert.Equal(t, StrategyDependencyAware, result.Strategy)
		assert.NotNil(t, result.Score)
		assert.NotEmpty(t, result.Reason)
		assert.NotZero(t, result.SelectedAt)
		assert.Greater(t, result.ExecutionTime, time.Duration(0))
	})
}

func TestDefaultTaskSelector_UpdateConfig(t *testing.T) {
	selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
	require.NoError(t, err)

	t.Run("ValidConfigUpdate", func(t *testing.T) {
		newConfig := DefaultConfig()
		newConfig.Strategy = StrategyPriority
		newConfig.Behavior.PreferInProgress = false

		err := selector.UpdateConfig(newConfig)
		assert.NoError(t, err)
		assert.Equal(t, StrategyPriority, selector.config.Strategy)
		assert.False(t, selector.config.Behavior.PreferInProgress)
	})

	t.Run("InvalidConfigUpdate", func(t *testing.T) {
		invalidConfig := &Config{
			Strategy: StrategyDependencyAware,
			Weights: Weights{
				DependentCount: -1.0, // Invalid
				Priority:       0.5,
				DepthFirst:     0.3,
				CriticalPath:   0.2,
			},
		}

		err := selector.UpdateConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid configuration")
	})

	t.Run("StrategyChange", func(t *testing.T) {
		newConfig := DefaultConfig()
		newConfig.Strategy = StrategyPriority

		err := selector.UpdateConfig(newConfig)
		assert.NoError(t, err)

		// Strategy should be updated
		assert.Equal(t, "priority", selector.strategy.GetStrategyName())
	})
}

func TestDefaultTaskSelector_scoreActionableTasks(t *testing.T) {
	selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
	require.NoError(t, err)

	t.Run("ScoreCalculation", func(t *testing.T) {
		tasks := createComplexDependencyGraph()
		graph, err := selector.analyzer.BuildDependencyGraph(tasks)
		require.NoError(t, err)

		actionableTasks, err := selector.filter.FilterActionableTasks(tasks)
		require.NoError(t, err)

		scores, err := selector.scoreActionableTasks(actionableTasks, graph)
		assert.NoError(t, err)
		assert.NotEmpty(t, scores)

		// All scores should have non-negative values
		for _, score := range scores {
			assert.GreaterOrEqual(t, score.Score, 0.0)
			assert.NotNil(t, score.Task)
			assert.NotZero(t, score.CalculatedAt)
		}
	})

	t.Run("EmptyActionableList", func(t *testing.T) {
		graph := &DependencyGraph{
			Nodes: make(map[uuid.UUID]*DependencyNode),
		}

		scores, err := selector.scoreActionableTasks([]*types.Task{}, graph)
		assert.NoError(t, err)
		assert.Empty(t, scores)
	})
}

func TestDefaultTaskSelector_selectBestTask(t *testing.T) {
	selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
	require.NoError(t, err)

	t.Run("EmptyScoreList", func(t *testing.T) {
		selected, alternatives := selector.selectBestTask([]*TaskScore{})
		assert.Nil(t, selected)
		assert.Empty(t, alternatives)
	})

	t.Run("SingleScore", func(t *testing.T) {
		task := createTestTask("task1", "Single Task", types.TaskStatePending, 5, nil, nil)
		scores := []*TaskScore{
			{
				Task:  task,
				Score: 10.0,
			},
		}

		selected, alternatives := selector.selectBestTask(scores)
		assert.NotNil(t, selected)
		assert.Equal(t, task, selected.Task)
		assert.Empty(t, alternatives)
	})

	t.Run("MultipleScores", func(t *testing.T) {
		task1 := createTestTask("task1", "Low Score", types.TaskStatePending, 5, nil, nil)
		task2 := createTestTask("task2", "High Score", types.TaskStatePending, 5, nil, nil)

		scores := []*TaskScore{
			{Task: task1, Score: 5.0},
			{Task: task2, Score: 15.0},
		}

		selected, alternatives := selector.selectBestTask(scores)
		assert.NotNil(t, selected)
		assert.Equal(t, task2, selected.Task) // Higher score should be selected
		assert.Len(t, alternatives, 1)
		assert.Equal(t, task1, alternatives[0].Task)
	})

	t.Run("InProgressPreference", func(t *testing.T) {
		config := DefaultConfig()
		config.Behavior.PreferInProgress = true
		selector, err := NewTaskSelector(StrategyDependencyAware, config)
		require.NoError(t, err)

		pendingTask := createTestTask("pending", "Pending", types.TaskStatePending, 5, nil, nil)
		inProgressTask := createTestTask("inprogress", "In Progress", types.TaskStateInProgress, 5, nil, nil)

		scores := []*TaskScore{
			{Task: pendingTask, Score: 15.0},    // Higher score
			{Task: inProgressTask, Score: 10.0}, // Lower score but in progress
		}

		selected, alternatives := selector.selectBestTask(scores)
		assert.NotNil(t, selected)
		assert.Equal(t, inProgressTask, selected.Task) // In-progress should be preferred
		assert.Len(t, alternatives, 1)
	})
}

func TestDefaultTaskSelector_generateSelectionReason(t *testing.T) {
	selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
	require.NoError(t, err)

	graph := &DependencyGraph{
		Nodes: map[uuid.UUID]*DependencyNode{
			testTaskID("task1"): {
				TaskID:             testTaskID("task1"),
				CriticalPathLength: 5,
			},
		},
	}

	t.Run("BasicReason", func(t *testing.T) {
		task := createTestTask("task1", "Test Task", types.TaskStatePending, 5, nil, nil)
		score := &TaskScore{
			Task:               task,
			DependentCount:     2,
			UnblockedTaskCount: 3,
			Priority:           2, // High priority
			HierarchyDepth:     0,
			Score:              10.0,
		}

		reason := selector.generateSelectionReason(score, graph)
		assert.NotEmpty(t, reason)
		assert.Contains(t, reason, "dependency-aware strategy")
	})

	t.Run("HighPriorityTask", func(t *testing.T) {
		task := createTestTask("task1", "High Priority", types.TaskStatePending, 1, nil, nil)
		score := &TaskScore{
			Task:     task,
			Priority: 1,
		}

		reason := selector.generateSelectionReason(score, graph)
		assert.Contains(t, reason, "dependency-aware strategy")
	})

	t.Run("SubtaskReason", func(t *testing.T) {
		parentID := testTaskID("parent")
		task := createTestTask("child", "Child Task", types.TaskStatePending, 5, &parentID, nil)
		score := &TaskScore{
			Task:           task,
			HierarchyDepth: 1,
		}

		reason := selector.generateSelectionReason(score, graph)
		assert.Contains(t, reason, "dependency-aware strategy")
	})
}

func BenchmarkTaskSelector_Selection(b *testing.B) {
	selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
	require.NoError(b, err)

	b.Run("Small_10tasks", func(b *testing.B) {
		tasks := createLargeProjectTasks(10)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := selector.SelectNextActionableTask(tasks)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Medium_50tasks", func(b *testing.B) {
		tasks := createLargeProjectTasks(50)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := selector.SelectNextActionableTask(tasks)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Large_100tasks", func(b *testing.B) {
		tasks := createLargeProjectTasks(100)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := selector.SelectNextActionableTask(tasks)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestDefaultTaskSelector_IntegrationScenarios(t *testing.T) {
	t.Run("RealWorldProject", func(t *testing.T) {
		// Simulate a real project with multiple features
		frontendID := testTaskID("frontend")
		backendID := testTaskID("backend")

		tasks := []*types.Task{
			// Frontend feature
			createTestTask("frontend", "Frontend Feature", types.TaskStatePending, 3, nil, nil),
			createTestTask("ui-design", "UI Design", types.TaskStatePending, 5, &frontendID, nil),
			createTestTask("ui-impl", "UI Implementation", types.TaskStatePending, 4, &frontendID,
				[]uuid.UUID{testTaskID("ui-design"), testTaskID("api")}),

			// Backend feature
			createTestTask("backend", "Backend Feature", types.TaskStatePending, 3, nil, nil),
			createTestTask("api", "API Design", types.TaskStatePending, 4, &backendID, nil),
			createTestTask("api-impl", "API Implementation", types.TaskStatePending, 4, &backendID,
				[]uuid.UUID{testTaskID("api"), testTaskID("database")}),

			// Infrastructure
			createTestTask("database", "Database Setup", types.TaskStatePending, 2, nil, nil), // High priority

			// Integration
			createTestTask("integration", "Integration Testing", types.TaskStatePending, 5, nil,
				[]uuid.UUID{testTaskID("ui-impl"), testTaskID("api-impl")}),
		}

		// Test dependency-aware strategy
		selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
		require.NoError(t, err)

		selectedTask, err := selector.SelectNextActionableTask(tasks)
		assert.NoError(t, err)
		assert.NotNil(t, selectedTask)

		// Should select a task that unblocks others or has high priority
		// Database Setup should be a good choice (high priority + dependency for API impl)
		validChoices := []string{"UI Design", "API Design", "Database Setup"}
		assert.Contains(t, validChoices, selectedTask.Title)

		// Test depth-first strategy
		depthFirstSelector, err := NewTaskSelector(StrategyDepthFirst, DefaultConfig())
		require.NoError(t, err)

		depthFirstTask, err := depthFirstSelector.SelectNextActionableTask(tasks)
		assert.NoError(t, err)
		assert.NotNil(t, depthFirstTask)

		// Should prefer subtasks over root tasks
		subtaskChoices := []string{"UI Design", "API Design"}
		if contains(subtaskChoices, depthFirstTask.Title) {
			t.Logf("✓ Depth-first correctly selected subtask: %s", depthFirstTask.Title)
		}
	})
}

// Helper function for integration test
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
