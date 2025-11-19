package selection

import (
	"fmt"

	"github.com/denkhaus/knot/internal/types"
)

// ActionabilityValidator handles task actionability validation
type ActionabilityValidator struct {
	config *Config
}

// NewActionabilityValidator creates a new actionability validator
func NewActionabilityValidator(config *Config) *ActionabilityValidator {
	return &ActionabilityValidator{
		config: config,
	}
}

// ValidateActionability checks if a task can be worked on right now
func (av *ActionabilityValidator) ValidateActionability(task *types.Task, allTasks []*types.Task) bool {
	// Only pending or in-progress tasks can be actionable
	if task.State != types.TaskStatePending && task.State != types.TaskStateInProgress {
		return false
	}

	// Check dependencies
	if !av.areDependenciesMet(task, allTasks) {
		return false
	}

	// Check subtasks (if configured to disallow parent with subtasks)
	if !av.config.Behavior.AllowParentWithSubtasks && av.hasActiveSubtasks(task, allTasks) {
		return false
	}

	return true
}

// ValidateAndExplain checks if a task can be worked on and provides detailed reasoning
func (av *ActionabilityValidator) ValidateAndExplain(task *types.Task, taskMap *TaskMap) (bool, []string) {
	reasons := make([]string, 0)

	// Check task state
	if task.State != types.TaskStatePending && task.State != types.TaskStateInProgress {
		reasons = append(reasons, fmt.Sprintf("task state is %v", task.State))
		return false, reasons
	}

	// Check dependencies
	for _, depID := range task.Dependencies {
		if depTask, exists := taskMap.Get(depID); exists {
			if depTask.State != types.TaskStateCompleted {
				reasons = append(reasons, fmt.Sprintf("dependency %s is not completed", depTask.Title))
			}
		} else {
			reasons = append(reasons, fmt.Sprintf("dependency %s not found", depID))
		}
	}

	// Check subtasks constraint
	if !av.config.Behavior.AllowParentWithSubtasks {
		for _, t := range taskMap.GetAll() {
			if t.ParentID != nil && *t.ParentID == task.ID {
				if t.State == types.TaskStatePending || t.State == types.TaskStateInProgress {
					reasons = append(reasons, "has active subtasks")
					break
				}
			}
		}
	}

	return len(reasons) == 0, reasons
}

// areDependenciesMet checks if all task dependencies are completed
func (av *ActionabilityValidator) areDependenciesMet(task *types.Task, allTasks []*types.Task) bool {
	taskMap := NewTaskMap(allTasks)

	for _, depID := range task.Dependencies {
		depTask, exists := taskMap.Get(depID)
		if !exists {
			if av.config.Behavior.StrictDependencies {
				return false // Missing dependency
			}
			continue // Ignore missing dependencies in non-strict mode
		}

		if depTask.State != types.TaskStateCompleted {
			return false
		}
	}

	return true
}

// hasActiveSubtasks checks if task has pending or in-progress children
func (av *ActionabilityValidator) hasActiveSubtasks(task *types.Task, allTasks []*types.Task) bool {
	for _, t := range allTasks {
		if t.ParentID != nil && *t.ParentID == task.ID {
			if t.State == types.TaskStatePending || t.State == types.TaskStateInProgress {
				return true
			}
		}
	}
	return false
}

// CountActionableTasks determines how many tasks are currently actionable
func (av *ActionabilityValidator) CountActionableTasks(graph *DependencyGraph, tasks []*types.Task) int {
	taskMap := NewTaskMap(tasks)
	count := 0

	for _, node := range graph.Nodes {
		isActionable, _ := av.ValidateAndExplain(node.Task, taskMap)
		node.IsActionable = isActionable
		if isActionable {
			count++
		}
	}

	graph.ActionableCount = count
	return count
}

// AddBlockingReasons determines why a task is not actionable and adds to node
func (av *ActionabilityValidator) AddBlockingReasons(node *DependencyNode, taskMap *TaskMap) {
	isActionable, reasons := av.ValidateAndExplain(node.Task, taskMap)
	if !isActionable {
		node.BlockingReasons = append(node.BlockingReasons, reasons...)
	}
}