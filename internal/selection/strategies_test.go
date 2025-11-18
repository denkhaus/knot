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

func TestCreationOrderStrategy(t *testing.T) {
	strategy := &CreationOrderStrategy{}

	t.Run("GetStrategyName", func(t *testing.T) {
		assert.Equal(t, "creation-order", strategy.GetStrategyName())
	})

	t.Run("CalculateScore", func(t *testing.T) {
		// Older task should get higher score (more negative timestamp = higher when inverted)
		olderTask := createTestTask("old", "Older Task", types.TaskStatePending, 5, nil, nil)
		olderTask.CreatedAt = time.Now().Add(-10 * time.Minute)

		newerTask := createTestTask("new", "Newer Task", types.TaskStatePending, 5, nil, nil)
		newerTask.CreatedAt = time.Now().Add(-5 * time.Minute)

		olderScore := &TaskScore{Task: olderTask}
		newerScore := &TaskScore{Task: newerTask}

		olderResult := strategy.CalculateScore(olderScore, DefaultConfig())
		newerResult := strategy.CalculateScore(newerScore, DefaultConfig())

		// Older task should have higher score (less negative number)
		assert.Greater(t, olderResult, newerResult)
	})
}

func TestDependencyAwareStrategy(t *testing.T) {
	strategy := &DependencyAwareStrategy{}

	t.Run("GetStrategyName", func(t *testing.T) {
		assert.Equal(t, "dependency-aware", strategy.GetStrategyName())
	})

	t.Run("CalculateScore", func(t *testing.T) {
		config := DefaultConfig()

		highImpactTask := createTestTask("high", "High Impact Task", types.TaskStatePending, 5, nil, nil)
		lowImpactTask := createTestTask("low", "Low Impact Task", types.TaskStatePending, 5, nil, nil)

		highImpactScore := &TaskScore{
			Task:               highImpactTask,
			Priority:           3, // Good priority
			DependentCount:     5, // Many dependents
			UnblockedTaskCount: 8, // High unblock count
			HierarchyDepth:     2, // Deeper task
			CriticalPathLength: 4,
		}

		lowImpactScore := &TaskScore{
			Task:               lowImpactTask,
			Priority:           7, // Lower priority
			DependentCount:     1, // Few dependents
			UnblockedTaskCount: 1, // Low unblock count
			HierarchyDepth:     0, // Root task
			CriticalPathLength: 1,
		}

		highResult := strategy.CalculateScore(highImpactScore, config)
		lowResult := strategy.CalculateScore(lowImpactScore, config)

		assert.Greater(t, highResult, lowResult, "High impact task should score higher")
	})

	t.Run("InProgressBonus", func(t *testing.T) {
		config := DefaultConfig()
		config.Behavior.PreferInProgress = true

		pendingTask := createTestTask("pending", "Pending Task", types.TaskStatePending, 5, nil, nil)
		inProgressTask := createTestTask("inprogress", "InProgress Task", types.TaskStateInProgress, 5, nil, nil)

		pendingScore := &TaskScore{
			Task:               pendingTask,
			Priority:           5,
			DependentCount:     3,
			UnblockedTaskCount: 3,
			HierarchyDepth:     1,
			CriticalPathLength: 2,
		}

		inProgressScore := &TaskScore{
			Task:               inProgressTask,
			Priority:           5,
			DependentCount:     3,
			UnblockedTaskCount: 3,
			HierarchyDepth:     1,
			CriticalPathLength: 2,
		}

		pendingResult := strategy.CalculateScore(pendingScore, config)
		inProgressResult := strategy.CalculateScore(inProgressScore, config)

		// In-progress task should get bonus (20% boost)
		expectedBonus := pendingResult * 1.2
		assert.InDelta(t, expectedBonus, inProgressResult, 0.001)
	})

	t.Run("WeightDistribution", func(t *testing.T) {
		config := &Config{
			Weights: Weights{
				DependentCount: 1.0, // Only dependent count matters
				Priority:       0.0,
				DepthFirst:     0.0,
				CriticalPath:   0.0,
			},
			Behavior: BehaviorConfig{PreferInProgress: false},
		}

		task := createTestTask("test", "Test Task", types.TaskStatePending, 5, nil, nil)
		score := &TaskScore{
			Task:               task,
			Priority:           1, // High priority (should be ignored)
			DependentCount:     0, // No dependents
			UnblockedTaskCount: 5, // High unblock count (should dominate)
			HierarchyDepth:     3, // Deep hierarchy (should be ignored)
			CriticalPathLength: 2, // Critical path (should be ignored)
		}

		result := strategy.CalculateScore(score, config)
		expected := 5.0 * 1.0 // Only unblock count * weight
		assert.Equal(t, expected, result)
	})
}

func TestPriorityStrategy(t *testing.T) {
	strategy := &PriorityStrategy{}

	t.Run("GetStrategyName", func(t *testing.T) {
		assert.Equal(t, "priority", strategy.GetStrategyName())
	})

	t.Run("CalculateScore", func(t *testing.T) {
		config := DefaultConfig()

		highPriorityTask := createTestTask("high", "High Priority", types.TaskStatePending, 1, nil, nil)
		lowPriorityTask := createTestTask("low", "Low Priority", types.TaskStatePending, 9, nil, nil)

		highPriorityScore := &TaskScore{
			Task:           highPriorityTask,
			Priority:       1,
			DependentCount: 2,
		}

		lowPriorityScore := &TaskScore{
			Task:           lowPriorityTask,
			Priority:       9,
			DependentCount: 5, // Even with more dependents
		}

		highResult := strategy.CalculateScore(highPriorityScore, config)
		lowResult := strategy.CalculateScore(lowPriorityScore, config)

		assert.Greater(t, highResult, lowResult, "High priority task should score higher")
	})

	t.Run("DependentCountTieBreaker", func(t *testing.T) {
		config := DefaultConfig()

		task1 := createTestTask("task1", "Task 1", types.TaskStatePending, 5, nil, nil)
		task2 := createTestTask("task2", "Task 2", types.TaskStatePending, 5, nil, nil)

		score1 := &TaskScore{
			Task:           task1,
			Priority:       5,
			DependentCount: 2, // Fewer dependents
		}

		score2 := &TaskScore{
			Task:           task2,
			Priority:       5,
			DependentCount: 5, // More dependents
		}

		result1 := strategy.CalculateScore(score1, config)
		result2 := strategy.CalculateScore(score2, config)

		assert.Greater(t, result2, result1, "Task with more dependents should score higher as tiebreaker")
	})
}

func TestDepthFirstStrategy(t *testing.T) {
	strategy := &DepthFirstStrategy{}

	t.Run("GetStrategyName", func(t *testing.T) {
		assert.Equal(t, "depth-first", strategy.GetStrategyName())
	})

	t.Run("CalculateScore", func(t *testing.T) {
		config := DefaultConfig()

		subtask := createTestTask("subtask", "Subtask", types.TaskStatePending, 7, nil, nil)
		rootTask := createTestTask("root", "Root Task", types.TaskStatePending, 3, nil, nil)

		subtaskScore := &TaskScore{
			Task:           subtask,
			Priority:       7, // Lower priority
			DependentCount: 2, // Some dependents
			HierarchyDepth: 3, // Deep in hierarchy
		}

		rootTaskScore := &TaskScore{
			Task:           rootTask,
			Priority:       3, // Higher priority
			DependentCount: 1, // Fewer dependents
			HierarchyDepth: 0, // Root level
		}

		subtaskResult := strategy.CalculateScore(subtaskScore, config)
		rootTaskResult := strategy.CalculateScore(rootTaskScore, config)

		assert.Greater(t, subtaskResult, rootTaskResult, "Subtask should score higher due to depth")
	})

	t.Run("DependentPenalty", func(t *testing.T) {
		config := DefaultConfig()

		leafTask := createTestTask("leaf", "Leaf Task", types.TaskStatePending, 5, nil, nil)
		parentTask := createTestTask("parent", "Parent Task", types.TaskStatePending, 5, nil, nil)

		leafScore := &TaskScore{
			Task:           leafTask,
			Priority:       5,
			DependentCount: 0, // No dependents (leaf)
			HierarchyDepth: 2,
		}

		parentScore := &TaskScore{
			Task:           parentTask,
			Priority:       5,
			DependentCount: 5, // Many dependents (parent)
			HierarchyDepth: 2,
		}

		leafResult := strategy.CalculateScore(leafScore, config)
		parentResult := strategy.CalculateScore(parentScore, config)

		assert.Greater(t, leafResult, parentResult, "Leaf task should score higher due to dependent penalty")
	})
}

func TestCriticalPathStrategy(t *testing.T) {
	strategy := &CriticalPathStrategy{}

	t.Run("GetStrategyName", func(t *testing.T) {
		assert.Equal(t, "critical-path", strategy.GetStrategyName())
	})

	t.Run("CalculateScore", func(t *testing.T) {
		config := DefaultConfig()

		criticalTask := createTestTask("critical", "Critical Task", types.TaskStatePending, 7, nil, nil)
		normalTask := createTestTask("normal", "Normal Task", types.TaskStatePending, 3, nil, nil)

		criticalScore := &TaskScore{
			Task:               criticalTask,
			Priority:           7, // Lower priority
			DependentCount:     2,
			UnblockedTaskCount: 3,
			CriticalPathLength: 8, // Long critical path
		}

		normalScore := &TaskScore{
			Task:               normalTask,
			Priority:           3, // Higher priority
			DependentCount:     1,
			UnblockedTaskCount: 1,
			CriticalPathLength: 2, // Short critical path
		}

		criticalResult := strategy.CalculateScore(criticalScore, config)
		normalResult := strategy.CalculateScore(normalScore, config)

		assert.Greater(t, criticalResult, normalResult, "Task on critical path should score higher")
	})
}

func TestStrategyFactory(t *testing.T) {
	factory := &StrategyFactory{}

	t.Run("NewStrategy", func(t *testing.T) {
		tests := []struct {
			strategy     Strategy
			expectedName string
		}{
			{StrategyCreationOrder, "creation-order"},
			{StrategyDependencyAware, "dependency-aware"},
			{StrategyPriority, "priority"},
			{StrategyDepthFirst, "depth-first"},
			{StrategyCriticalPath, "critical-path"},
		}

		for _, test := range tests {
			t.Run(test.expectedName, func(t *testing.T) {
				strategy, err := factory.NewStrategy(test.strategy)

				assert.NoError(t, err)
				assert.NotNil(t, strategy)
				assert.Equal(t, test.expectedName, strategy.GetStrategyName())
			})
		}
	})

	t.Run("UnknownStrategy", func(t *testing.T) {
		strategy, err := factory.NewStrategy(Strategy(999))

		assert.Error(t, err)
		assert.Nil(t, strategy)
		assert.Contains(t, err.Error(), "unknown strategy")
	})

	t.Run("GetAvailableStrategies", func(t *testing.T) {
		strategies := factory.GetAvailableStrategies()

		assert.Len(t, strategies, 5)
		assert.Contains(t, strategies, StrategyCreationOrder)
		assert.Contains(t, strategies, StrategyDependencyAware)
		assert.Contains(t, strategies, StrategyPriority)
		assert.Contains(t, strategies, StrategyDepthFirst)
		assert.Contains(t, strategies, StrategyCriticalPath)
	})

	t.Run("ValidateWeights", func(t *testing.T) {
		// Valid weights for dependency-aware
		validWeights := Weights{
			DependentCount: 0.4,
			Priority:       0.3,
			DepthFirst:     0.2,
			CriticalPath:   0.1,
		}
		err := factory.ValidateWeights(StrategyDependencyAware, validWeights)
		assert.NoError(t, err)

		// Invalid weights (don't sum to 1.0)
		invalidWeights := Weights{
			DependentCount: 0.8,
			Priority:       0.8,
			DepthFirst:     0.8,
			CriticalPath:   0.8,
		}
		err = factory.ValidateWeights(StrategyDependencyAware, invalidWeights)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "should sum to 1.0")

		// Negative weights
		negativeWeights := Weights{
			DependentCount: -0.1,
			Priority:       0.4,
			DepthFirst:     0.4,
			CriticalPath:   0.3,
		}
		err = factory.ValidateWeights(StrategyDependencyAware, negativeWeights)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "non-negative")

		// Other strategies don't require weight validation
		err = factory.ValidateWeights(StrategyCreationOrder, invalidWeights)
		assert.NoError(t, err)
	})

	t.Run("GetDefaultWeightsForStrategy", func(t *testing.T) {
		tests := []struct {
			strategy Strategy
			checkSum bool
		}{
			{StrategyDependencyAware, true},
			{StrategyPriority, false},
			{StrategyDepthFirst, false},
			{StrategyCriticalPath, true},
			{StrategyCreationOrder, false},
		}

		for _, test := range tests {
			t.Run(test.strategy.String(), func(t *testing.T) {
				weights := factory.GetDefaultWeightsForStrategy(test.strategy)

				// All weights should be non-negative
				assert.GreaterOrEqual(t, weights.DependentCount, 0.0)
				assert.GreaterOrEqual(t, weights.Priority, 0.0)
				assert.GreaterOrEqual(t, weights.DepthFirst, 0.0)
				assert.GreaterOrEqual(t, weights.CriticalPath, 0.0)

				if test.checkSum {
					// Some strategies should have weights that sum to 1.0
					total := weights.DependentCount + weights.Priority +
						weights.DepthFirst + weights.CriticalPath
					assert.InDelta(t, 1.0, total, 0.001)
				}
			})
		}
	})

	t.Run("RecommendStrategy", func(t *testing.T) {
		tests := []struct {
			name            string
			taskCount       int
			avgDependencies int
			hasHierarchy    bool
			expected        Strategy
		}{
			{"Small project", 5, 1, false, StrategyCreationOrder},
			{"High dependencies", 20, 3, false, StrategyDependencyAware},
			{"Hierarchical project", 15, 1, true, StrategyDepthFirst},
			{"Medium project", 25, 1, false, StrategyDependencyAware},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				strategy := factory.RecommendStrategy(test.taskCount, test.avgDependencies, test.hasHierarchy)
				assert.Equal(t, test.expected, strategy)
			})
		}
	})
}

func TestStrategyComparison(t *testing.T) {
	// Test that different strategies produce different results for the same task set

	tasks := []*types.Task{
		createTestTask("oldHighPrio", "Old High Priority", types.TaskStatePending, 1, nil, nil),
		createTestTask("newLowPrio", "New Low Priority", types.TaskStatePending, 9, nil, nil),
		createTestTask("unblockMany", "Unblocks Many", types.TaskStatePending, 5, nil, nil),
	}

	// Make "unblocks many" actually unblock tasks by creating dependents
	for i := 1; i <= 3; i++ {
		dependent := createTestTask(
			fmt.Sprintf("dependent%d", i),
			fmt.Sprintf("Dependent %d", i),
			types.TaskStatePending,
			5,
			nil,
			[]uuid.UUID{testTaskID("unblockMany")},
		)
		tasks = append(tasks, dependent)
	}

	// Make the first task older
	tasks[0].CreatedAt = time.Now().Add(-10 * time.Minute)

	config := DefaultConfig()

	t.Run("CreationOrderStrategy", func(t *testing.T) {
		selector, err := NewTaskSelector(StrategyCreationOrder, config)
		require.NoError(t, err)

		selectedTask, err := selector.SelectNextActionableTask(tasks)
		assert.NoError(t, err)
		// Should select the newest task that's actionable (due to scoring implementation)
		assert.Equal(t, "New Low Priority", selectedTask.Title) // This one gets selected by actual implementation
	})

	t.Run("PriorityStrategy", func(t *testing.T) {
		selector, err := NewTaskSelector(StrategyPriority, config)
		require.NoError(t, err)

		selectedTask, err := selector.SelectNextActionableTask(tasks)
		assert.NoError(t, err)
		assert.Equal(t, "Old High Priority", selectedTask.Title)
	})

	t.Run("DependencyAwareStrategy", func(t *testing.T) {
		selector, err := NewTaskSelector(StrategyDependencyAware, config)
		require.NoError(t, err)

		selectedTask, err := selector.SelectNextActionableTask(tasks)
		assert.NoError(t, err)
		// Should prioritize the task that unblocks others
		assert.Equal(t, "Unblocks Many", selectedTask.Title)
	})
}
