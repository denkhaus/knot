package selection

import (
	"testing"
	"github.com/denkhaus/knot/v2/internal/types"
)

func TestBasicDependencyAnalyzer(t *testing.T) {
	analyzer := NewDependencyAnalyzer(DefaultConfig())

	t.Run("EmptyTaskList", func(t *testing.T) {
		graph, err := analyzer.BuildDependencyGraph([]*types.Task{})
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if graph == nil {
			t.Error("Expected graph, got nil")
		}
		if len(graph.Nodes) != 0 {
			t.Errorf("Expected 0 nodes, got %d", len(graph.Nodes))
		}
		if graph.TaskCount != 0 {
			t.Errorf("Expected 0 task count, got %d", graph.TaskCount)
		}
	})

	t.Run("SingleTask", func(t *testing.T) {
		task := createTestTask("task1", "Single Task", types.TaskStatePending, 2, nil, nil)
		tasks := []*types.Task{task}

		graph, err := analyzer.BuildDependencyGraph(tasks)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(graph.Nodes) != 1 {
			t.Errorf("Expected 1 node, got %d", len(graph.Nodes))
		}
		if graph.TaskCount != 1 {
			t.Errorf("Expected 1 task count, got %d", graph.TaskCount)
		}
		
		node := graph.Nodes[testTaskID("task1")]
		if node == nil {
			t.Error("Expected node to exist")
		}
	})

	t.Run("ValidateActionability", func(t *testing.T) {
		readyTask := createTestTask("ready", "Ready Task", types.TaskStatePending, 2, nil, nil)
		completedTask := createTestTask("completed", "Completed Task", types.TaskStateCompleted, 2, nil, nil)
		tasks := []*types.Task{readyTask, completedTask}

		if !analyzer.ValidateActionability(readyTask, tasks) {
			t.Error("Expected ready task to be actionable")
		}
		if analyzer.ValidateActionability(completedTask, tasks) {
			t.Error("Expected completed task to not be actionable")
		}
	})
}

func TestTaskSelectorBasic(t *testing.T) {
	selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create selector: %v", err)
	}

	t.Run("SingleTaskSelection", func(t *testing.T) {
		task := createTestTask("only", "Only Task", types.TaskStatePending, 2, nil, nil)
		tasks := []*types.Task{task}

		selected, err := selector.SelectNextActionableTask(tasks)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if selected == nil {
			t.Error("Expected a task to be selected")
		}
		if selected != nil && selected.Title != "Only Task" {
			t.Errorf("Expected 'Only Task', got '%s'", selected.Title)
		}
	})

	t.Run("EmptyTaskList", func(t *testing.T) {
		selected, err := selector.SelectNextActionableTask([]*types.Task{})
		if err == nil {
			t.Error("Expected an error for empty task list")
		}
		if selected != nil {
			t.Error("Expected no task to be selected")
		}
	})
}