package selection

import (
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
)

// TaskMap provides efficient task lookup by ID
type TaskMap struct {
	tasks map[uuid.UUID]*types.Task
}

// NewTaskMap creates a new task map from a slice of tasks
func NewTaskMap(tasks []*types.Task) *TaskMap {
	taskMap := make(map[uuid.UUID]*types.Task, len(tasks))
	for _, task := range tasks {
		taskMap[task.ID] = task
	}
	return &TaskMap{tasks: taskMap}
}

// Get retrieves a task by ID
func (tm *TaskMap) Get(id uuid.UUID) (*types.Task, bool) {
	task, exists := tm.tasks[id]
	return task, exists
}

// Exists checks if a task exists in the map
func (tm *TaskMap) Exists(id uuid.UUID) bool {
	_, exists := tm.tasks[id]
	return exists
}

// GetAll returns all tasks as a slice
func (tm *TaskMap) GetAll() []*types.Task {
	tasks := make([]*types.Task, 0, len(tm.tasks))
	for _, task := range tm.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// Size returns the number of tasks in the map
func (tm *TaskMap) Size() int {
	return len(tm.tasks)
}

// ForEach iterates over all tasks in the map
func (tm *TaskMap) ForEach(fn func(id uuid.UUID, task *types.Task)) {
	for id, task := range tm.tasks {
		fn(id, task)
	}
}