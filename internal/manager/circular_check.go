package manager

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// CircularDependencyError is returned when adding a dependency would create a circular relationship
type CircularDependencyError struct {
	TaskID    uuid.UUID
	DependsOn uuid.UUID
	Path      []uuid.UUID
}

func (e *CircularDependencyError) Error() string {
	if len(e.Path) > 0 {
		return fmt.Sprintf("circular dependency detected: %s", formatCircularPath(e.Path))
	}
	return fmt.Sprintf("circular dependency detected: task %s depends on %s", e.TaskID, e.DependsOn)
}

// CheckCircularDependency checks if adding a dependency would create a circular relationship
// Returns (isCircular, circularPath)
func CheckCircularDependency(ctx context.Context, projectManager ProjectManager, taskID, dependsOnTaskID uuid.UUID) (bool, []uuid.UUID) {
	// Check if tasks are the same (self-dependency)
	if taskID == dependsOnTaskID {
		return true, []uuid.UUID{taskID, taskID}
	}

	// Get current dependencies of the task that would depend on the other
	taskDependencies, err := projectManager.GetTaskDependencies(ctx, taskID)
	if err != nil {
		// If we can't get dependencies, assume not circular to avoid blocking valid operations
		return false, nil
	}

	// Check if dependsOnTask is already in the task's dependencies (inverse check)
	for _, dep := range taskDependencies {
		if dep.ID == dependsOnTaskID {
			// Already depends on it, so adding would be a duplicate, not circular
			return false, nil
		}
	}

	// Check if dependsOnTask depends on current task (would create A -> B -> A cycle)
	dependsOnDependencies, err := projectManager.GetTaskDependencies(ctx, dependsOnTaskID)
	if err != nil {
		return false, nil
	}

	for _, dep := range dependsOnDependencies {
		if dep.ID == taskID {
			return true, []uuid.UUID{taskID, dependsOnTaskID, taskID}
		}
	}

	// Check for deeper cycles: A -> B -> C -> A
	// We need to traverse the dependency graph of dependsOnTask to see if it leads back to taskID
	visited := make(map[uuid.UUID]bool)
	path := []uuid.UUID{dependsOnTaskID}
	if hasPathTo(ctx, projectManager, dependsOnTaskID, taskID, visited, &path) {
		// Prepend taskID to show full cycle
		fullPath := append([]uuid.UUID{taskID}, path...)
		return true, fullPath
	}

	return false, nil
}

// hasPathTo recursively checks if there's a dependency path from currentTaskID to targetTaskID
func hasPathTo(ctx context.Context, projectManager ProjectManager, currentTaskID, targetTaskID uuid.UUID, visited map[uuid.UUID]bool, path *[]uuid.UUID) bool {
	// Prevent infinite recursion
	if visited[currentTaskID] {
		return false
	}
	visited[currentTaskID] = true

	// Get dependencies of current task
	dependencies, err := projectManager.GetTaskDependencies(ctx, currentTaskID)
	if err != nil {
		return false
	}

	for _, dep := range dependencies {
		if dep.ID == targetTaskID {
			*path = append(*path, dep.ID)
			return true
		}

		// Recursively check
		if hasPathTo(ctx, projectManager, dep.ID, targetTaskID, visited, path) {
			*path = append(*path, dep.ID)
			return true
		}
	}

	return false
}

// formatCircularPath formats a circular dependency path for display
func formatCircularPath(path []uuid.UUID) string {
	if len(path) == 0 {
		return "unknown cycle"
	}

	result := ""
	for i, id := range path {
		if i > 0 {
			result += " -> "
		}
		result += id.String()[:8] // Show first 8 characters of UUID for readability
	}
	return result
}
