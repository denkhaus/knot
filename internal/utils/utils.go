// Package utils provides common utility functions for the KNOT project management system.
//
// This package contains helper functions for common operations across the application,
// including mathematical utilities, task state checking, JSON output formatting,
// and UUID conversion helpers.
//
// Key Functions:
//   - Min: Returns the minimum of two integers
//   - IsTaskReady: Checks if a task has all its dependencies completed
//
// # Package utils provides utility functions for task management and data processing
//
// Key utilities:
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

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// IsTaskReady checks if a task has all its dependencies completed
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

// OutputTasksAsJSON outputs tasks in JSON format
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

// ConvertUUIDsToStrings converts UUID slice to string slice for logging
func ConvertUUIDsToStrings(uuids []uuid.UUID) []string {
	result := make([]string, len(uuids))
	for i, u := range uuids {
		result[i] = u.String()
	}
	return result
}

// GetParentIDAsString safely converts a parent UUID pointer to string
func GetParentIDAsString(parentID *uuid.UUID) *string {
	if parentID == nil {
		return nil
	}
	str := parentID.String()
	return &str
}

// BuildTaskTreeFromTasks builds a hierarchical tree structure from a flat list of tasks
// This is a utility function that can be used across the application
func BuildTaskTreeFromTasks(tasks []*types.Task) []TaskTreeNode {
	// Create a map of tasks for quick lookup
	taskMap := make(map[uuid.UUID]*types.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	// Find root tasks (no parent)
	var rootTasks []*types.Task
	for _, task := range tasks {
		if task.ParentID == nil {
			rootTasks = append(rootTasks, task)
		}
	}

	// Build tree recursively
	var tree []TaskTreeNode
	for _, rootTask := range rootTasks {
		node := buildTaskTreeNode(rootTask, taskMap, 0)
		tree = append(tree, node)
	}

	return tree
}

// TaskTreeNode represents a task node in the tree structure
// This is a generic structure that can be used across the application
type TaskTreeNode struct {
	Task     *types.Task     `json:"task"`
	Children []TaskTreeNode  `json:"children,omitempty"`
	Level    int             `json:"level"`
}

// buildTaskTreeNode recursively builds a tree node and its children
func buildTaskTreeNode(task *types.Task, taskMap map[uuid.UUID]*types.Task, level int) TaskTreeNode {
	node := TaskTreeNode{
		Task:  task,
		Level: level,
	}

	// Find children
	for _, potentialChild := range taskMap {
		if potentialChild.ParentID != nil && *potentialChild.ParentID == task.ID {
			childNode := buildTaskTreeNode(potentialChild, taskMap, level+1)
			node.Children = append(node.Children, childNode)
		}
	}

	return node
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
