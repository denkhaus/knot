package selection

import (
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestValidationError(t *testing.T) {
	err := ValidationError{
		TaskID:  testTaskID("task1"),
		Message: "test error message",
		Type:    "test_type",
	}

	assert.Equal(t, "test error message", err.Error())
	assert.Equal(t, "test_type", err.Type)
	assert.Equal(t, testTaskID("task1"), err.TaskID)
}

func TestSelectionError(t *testing.T) {
	taskID := testTaskID("task1")
	err := SelectionError{
		Type:    ErrorTypeNoActionable,
		Message: "no actionable tasks",
		TaskID:  &taskID,
		ValidationErrs: []ValidationError{
			{Message: "validation error 1", Type: "validation"},
			{Message: "validation error 2", Type: "validation"},
		},
	}

	assert.Equal(t, "no actionable tasks", err.Error())
	assert.Equal(t, ErrorTypeNoActionable, err.Type)
	assert.Equal(t, &taskID, err.TaskID)
	assert.Len(t, err.ValidationErrs, 2)
}

func TestProjectComplexity_String(t *testing.T) {
	tests := []struct {
		complexity ProjectComplexity
		expected   string
	}{
		{ComplexitySimple, "simple"},
		{ComplexityMedium, "medium"},
		{ComplexityComplex, "complex"},
		{ProjectComplexity(999), "unknown"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			assert.Equal(t, test.expected, test.complexity.String())
		})
	}
}

func TestErrorTypes(t *testing.T) {
	// Test that error type constants are defined
	assert.Equal(t, "no_tasks", ErrorTypeNoTasks)
	assert.Equal(t, "no_actionable", ErrorTypeNoActionable)
	assert.Equal(t, "deadlock", ErrorTypeDeadlock)
	assert.Equal(t, "invalid_config", ErrorTypeInvalidConfig)
	assert.Equal(t, "circular_dependency", ErrorTypeCircularDep)
	assert.Equal(t, "validation", ErrorTypeValidation)
}

func TestTaskScore_DefaultValues(t *testing.T) {
	task := createTestTask("task1", "Test Task", types.TaskStatePending, 5, nil, nil)
	score := &TaskScore{
		Task:               task,
		DependentCount:     0,
		UnblockedTaskCount: 0,
		DependencyDepth:    0,
		CriticalPathLength: 0,
		HierarchyDepth:     0,
		Priority:           task.Priority,
		Score:              0.0,
		SelectionReason:    "",
		CalculatedAt:       time.Now(),
	}

	assert.Equal(t, task, score.Task)
	assert.Equal(t, 0, score.DependentCount)
	assert.Equal(t, 0, score.UnblockedTaskCount)
	assert.Equal(t, task.Priority, score.Priority)
	assert.NotZero(t, score.CalculatedAt)
}

func TestDependencyNode_DefaultValues(t *testing.T) {
	task := createTestTask("task1", "Test Task", types.TaskStatePending, 5, nil, nil)
	node := &DependencyNode{
		TaskID:             task.ID,
		Task:               task,
		Dependencies:       make([]uuid.UUID, 0),
		Dependents:         make([]uuid.UUID, 0),
		Children:           make([]uuid.UUID, 0),
		Parent:             nil,
		DependentCount:     0,
		ChildCount:         0,
		DependencyDepth:    0,
		CriticalPathLength: 0,
		UnblockedCount:     0,
		IsActionable:       false,
		BlockingReasons:    make([]string, 0),
	}

	assert.Equal(t, task.ID, node.TaskID)
	assert.Equal(t, task, node.Task)
	assert.Empty(t, node.Dependencies)
	assert.Empty(t, node.Dependents)
	assert.Empty(t, node.Children)
	assert.Nil(t, node.Parent)
	assert.False(t, node.IsActionable)
	assert.Empty(t, node.BlockingReasons)
}

func TestDependencyGraph_DefaultValues(t *testing.T) {
	graph := &DependencyGraph{
		Nodes:           make(map[uuid.UUID]*DependencyNode),
		RootTasks:       make([]uuid.UUID, 0),
		LeafTasks:       make([]uuid.UUID, 0),
		CriticalPath:    make([]uuid.UUID, 0),
		HasCycles:       false,
		CyclicTasks:     make([]uuid.UUID, 0),
		AnalyzedAt:      time.Now(),
		TaskCount:       0,
		ActionableCount: 0,
	}

	assert.Empty(t, graph.Nodes)
	assert.Empty(t, graph.RootTasks)
	assert.Empty(t, graph.LeafTasks)
	assert.Empty(t, graph.CriticalPath)
	assert.False(t, graph.HasCycles)
	assert.Empty(t, graph.CyclicTasks)
	assert.NotZero(t, graph.AnalyzedAt)
	assert.Equal(t, 0, graph.TaskCount)
	assert.Equal(t, 0, graph.ActionableCount)
}
