package utils

import (
	"encoding/json"
	"fmt"

	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
)

// Helper function for min
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isTaskReady checks if a task has all its dependencies completed
func IsTaskReady(task *types.Task, taskMap map[uuid.UUID]*types.Task) bool {
	// If task has no dependencies, it's ready
	if len(task.Dependencies) == 0 {
		return true
	}

	// Check if all dependencies are completed
	for _, depID := range task.Dependencies {
		depTask, exists := taskMap[depID]
		if !exists || depTask.State != types.TaskStateCompleted {
			return false
		}
	}

	return true
}

// outputTasksAsJSON outputs tasks in JSON format
func OutputTasksAsJSON(tasks []*types.Task) error {
	jsonData, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks to JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// OutputTaskAsJSON outputs a single task in JSON format
func OutputTaskAsJSON(task *types.Task) error {
	jsonData, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal task to JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// ParsePriority converts string priority to TaskPriority int
func ParsePriority(priority string) types.TaskPriority {
	switch priority {
	case "high":
		return 1
	case "medium":
		return 2
	case "low":
		return 3
	default:
		return 2 // Default to medium
	}
}
