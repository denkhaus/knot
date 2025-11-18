package selection

import (
	"fmt"

	"github.com/denkhaus/knot/internal/types"
)

// DefaultTaskFilter implements TaskFilter interface
type DefaultTaskFilter struct {
	analyzer DependencyAnalyzer
	config   *Config
}

// NewTaskFilter creates a new task filter
func NewTaskFilter(analyzer DependencyAnalyzer, config *Config) *DefaultTaskFilter {
	return &DefaultTaskFilter{
		analyzer: analyzer,
		config:   config,
	}
}

// FilterActionableTasks returns tasks that are eligible for selection
func (tf *DefaultTaskFilter) FilterActionableTasks(tasks []*types.Task) ([]*types.Task, error) {
	if len(tasks) == 0 {
		return []*types.Task{}, nil
	}

	candidates := tf.getCandidateTasks(tasks)
	if len(candidates) == 0 {
		return []*types.Task{}, fmt.Errorf("no pending or in-progress tasks available")
	}

	actionable := make([]*types.Task, 0, len(candidates))
	for _, task := range candidates {
		if tf.IsTaskActionable(task, tasks) {
			actionable = append(actionable, task)
		}
	}

	if len(actionable) == 0 {
		if tf.hasPendingTasks(tasks) {
			return []*types.Task{}, &SelectionError{
				Type:    ErrorTypeDeadlock,
				Message: "no actionable tasks found: all pending tasks have unmet dependencies (possible deadlock scenario)",
			}
		}
		return []*types.Task{}, &SelectionError{
			Type:    ErrorTypeNoActionable,
			Message: "no actionable tasks available",
		}
	}

	return actionable, nil
}

// IsTaskActionable checks if a specific task can be worked on right now
func (tf *DefaultTaskFilter) IsTaskActionable(task *types.Task, allTasks []*types.Task) bool {
	return tf.analyzer.ValidateActionability(task, allTasks)
}

// getCandidateTasks returns tasks that are in a workable state
func (tf *DefaultTaskFilter) getCandidateTasks(allTasks []*types.Task) []*types.Task {
	candidates := make([]*types.Task, 0)

	for _, task := range allTasks {
		if task.State == types.TaskStatePending || task.State == types.TaskStateInProgress {
			candidates = append(candidates, task)
		}
	}

	return candidates
}

// hasPendingTasks checks if there are any pending tasks (for deadlock detection)
func (tf *DefaultTaskFilter) hasPendingTasks(allTasks []*types.Task) bool {
	for _, task := range allTasks {
		if task.State == types.TaskStatePending {
			return true
		}
	}
	return false
}

// SeparateInProgressTasks separates tasks by state for prioritization
func (tf *DefaultTaskFilter) SeparateInProgressTasks(tasks []*types.Task) (inProgress, pending []*types.Task) {
	inProgress = make([]*types.Task, 0)
	pending = make([]*types.Task, 0)

	for _, task := range tasks {
		if task.State == types.TaskStateInProgress {
			inProgress = append(inProgress, task)
		} else if task.State == types.TaskStatePending {
			pending = append(pending, task)
		}
	}

	return inProgress, pending
}

// FilterByHierarchyDepth filters tasks by hierarchy depth
func (tf *DefaultTaskFilter) FilterByHierarchyDepth(tasks []*types.Task, maxDepth int) []*types.Task {
	filtered := make([]*types.Task, 0)

	for _, task := range tasks {
		depth := tf.calculateTaskHierarchyDepth(task, tasks)
		if depth <= maxDepth {
			filtered = append(filtered, task)
		}
	}

	return filtered
}

// FilterByPriority filters tasks by priority threshold
func (tf *DefaultTaskFilter) FilterByPriority(tasks []*types.Task, maxPriority types.TaskPriority) []*types.Task {
	filtered := make([]*types.Task, 0)

	for _, task := range tasks {
		if task.Priority <= maxPriority {
			filtered = append(filtered, task)
		}
	}

	return filtered
}

// calculateTaskHierarchyDepth calculates hierarchy depth for a single task
func (tf *DefaultTaskFilter) calculateTaskHierarchyDepth(task *types.Task, allTasks []*types.Task) int {
	depth := 0
	current := task
	visited := make(map[string]bool) // Use task ID as string to avoid circular references

	taskMap := make(map[string]*types.Task)
	for _, t := range allTasks {
		taskMap[t.ID.String()] = t
	}

	for current.ParentID != nil && !visited[current.ID.String()] {
		visited[current.ID.String()] = true
		depth++
		if parent, exists := taskMap[current.ParentID.String()]; exists {
			current = parent
		} else {
			break
		}
	}

	return depth
}
