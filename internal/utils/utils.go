// Package utils provides common utility functions for the KNOT project management system.
//
// This package contains helper functions for common operations across the application,
// including mathematical utilities, task state checking, JSON output formatting,
// and UUID conversion helpers.
//
// Key Functions:
//   - Min: Returns the minimum of two integers
//   - IsTaskReady: Checks if a task has all its dependencies completed
//   - OutputTasksAsJSON: Formats tasks as JSON for output
//   - ConvertUUIDsToStrings: Converts UUID slices to string slices for logging
package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/denkhaus/knot/v2/internal/types"
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

// Helper function to convert UUID slice to strings for logging
func ConvertUUIDsToStrings(uuids []uuid.UUID) []string {
	result := make([]string, len(uuids))
	for i, u := range uuids {
		result[i] = u.String()
	}
	return result
}

// WrapText wraps text at the specified width and returns a slice of lines
// It respects word boundaries to avoid breaking words mid-sentence
func WrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	for len(text) > width {
		// Find the last space before width to avoid breaking words
		breakPoint := width
		for i := width; i > 0; i-- {
			if text[i] == ' ' {
				breakPoint = i
				break
			}
		}

		if breakPoint == 0 {
			// No space found, break at width
			breakPoint = width
		}

		lines = append(lines, text[:breakPoint])
		text = strings.TrimSpace(text[breakPoint:])
	}

	if len(text) > 0 {
		lines = append(lines, text)
	}

	return lines
}
