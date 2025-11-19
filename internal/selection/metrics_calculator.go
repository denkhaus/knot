package selection

import (
	"fmt"
	"time"

	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
)

// MetricsCalculator handles the calculation of various task and graph metrics
type MetricsCalculator struct {
	config *Config
}

// NewMetricsCalculator creates a new metrics calculator
func NewMetricsCalculator(config *Config) *MetricsCalculator {
	return &MetricsCalculator{
		config: config,
	}
}

// CalculateAll computes metrics for all nodes in the graph
func (mc *MetricsCalculator) CalculateAll(graph *DependencyGraph) error {
	for taskID := range graph.Nodes {
		node := graph.Nodes[taskID]

		// Basic counts
		node.DependentCount = len(node.Dependents)
		node.ChildCount = len(node.Children)

		// Complex metrics
		node.DependencyDepth = mc.calculateDependencyDepth(graph, taskID, make(map[uuid.UUID]bool))
		node.CriticalPathLength = mc.calculateCriticalPathLength(graph, taskID, make(map[uuid.UUID]bool))
		node.UnblockedCount = mc.calculateUnblockedCount(graph, taskID, make(map[uuid.UUID]bool))
	}

	return nil
}

// calculateDependencyDepth computes how deep a task is in dependency chains
func (mc *MetricsCalculator) calculateDependencyDepth(graph *DependencyGraph, taskID uuid.UUID, visited map[uuid.UUID]bool) int {
	if visited[taskID] {
		return 0 // Circular dependency protection
	}

	if mc.config.Advanced.MaxDependencyDepth > 0 && len(visited) >= mc.config.Advanced.MaxDependencyDepth {
		return len(visited) // Max depth reached
	}

	node := graph.Nodes[taskID]
	if len(node.Dependencies) == 0 {
		return 0
	}

	visited[taskID] = true
	defer delete(visited, taskID)

	maxDepth := 0
	for _, depID := range node.Dependencies {
		if _, exists := graph.Nodes[depID]; exists {
			depth := 1 + mc.calculateDependencyDepth(graph, depID, visited)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	}

	return maxDepth
}

// calculateCriticalPathLength finds the longest dependency chain through this task
func (mc *MetricsCalculator) calculateCriticalPathLength(graph *DependencyGraph, taskID uuid.UUID, visited map[uuid.UUID]bool) int {
	if visited[taskID] {
		return 0
	}

	visited[taskID] = true
	defer delete(visited, taskID)

	node := graph.Nodes[taskID]
	maxPath := 1 // This task itself

	for _, depID := range node.Dependents {
		if _, exists := graph.Nodes[depID]; exists {
			path := 1 + mc.calculateCriticalPathLength(graph, depID, visited)
			if path > maxPath {
				maxPath = path
			}
		}
	}

	return maxPath
}

// calculateUnblockedCount counts how many tasks would become actionable if this task completes
func (mc *MetricsCalculator) calculateUnblockedCount(graph *DependencyGraph, taskID uuid.UUID, visited map[uuid.UUID]bool) int {
	if visited[taskID] {
		return 0
	}

	visited[taskID] = true
	defer delete(visited, taskID)

	count := 0
	node := graph.Nodes[taskID]

	for _, depTaskID := range node.Dependents {
		depNode := graph.Nodes[depTaskID]
		depTask := depNode.Task

		if depTask.State == types.TaskStatePending {
			// Check if this task completion would make the dependent actionable
			if mc.wouldBecomeActionable(depNode, graph, taskID) {
				count++ // This dependent would become actionable
				// Recursively count what this dependent would unblock
				count += mc.calculateUnblockedCount(graph, depTaskID, visited)
			}
		}
	}

	return count
}

// wouldBecomeActionable checks if a task would become actionable when a dependency completes
func (mc *MetricsCalculator) wouldBecomeActionable(node *DependencyNode, graph *DependencyGraph, completedDepID uuid.UUID) bool {
	// Check all other dependencies
	for _, depID := range node.Dependencies {
		if depID != completedDepID {
			if depNode, exists := graph.Nodes[depID]; exists {
				if depNode.Task.State != types.TaskStateCompleted {
					return false // Still has other incomplete dependencies
				}
			}
		}
	}

	// Check subtasks constraint
	if !mc.config.Behavior.AllowParentWithSubtasks {
		for _, childID := range node.Children {
			if childNode, exists := graph.Nodes[childID]; exists {
				if childNode.Task.State == types.TaskStatePending || childNode.Task.State == types.TaskStateInProgress {
					return false // Has active subtasks
				}
			}
		}
	}

	return true
}

// FindCriticalPath identifies the longest dependency chain in the graph
func (mc *MetricsCalculator) FindCriticalPath(graph *DependencyGraph) {
	maxLength := 0
	var criticalPath []uuid.UUID

	// Start from root tasks and find longest path
	for _, rootID := range graph.RootTasks {
		path := mc.findLongestPath(graph, rootID, make(map[uuid.UUID]bool))
		if len(path) > maxLength {
			maxLength = len(path)
			criticalPath = path
		}
	}

	graph.CriticalPath = criticalPath
}

// findLongestPath finds the longest dependency path starting from a task
func (mc *MetricsCalculator) findLongestPath(graph *DependencyGraph, taskID uuid.UUID, visited map[uuid.UUID]bool) []uuid.UUID {
	if visited[taskID] {
		return []uuid.UUID{}
	}

	visited[taskID] = true
	defer delete(visited, taskID)

	node := graph.Nodes[taskID]
	longestPath := []uuid.UUID{taskID}

	for _, depID := range node.Dependents {
		if _, exists := graph.Nodes[depID]; exists {
			path := mc.findLongestPath(graph, depID, visited)
			if len(path)+1 > len(longestPath) {
				longestPath = append([]uuid.UUID{taskID}, path...)
			}
		}
	}

	return longestPath
}

// CalculateTaskScore computes a detailed score for a task
func (mc *MetricsCalculator) CalculateTaskScore(task *types.Task, graph *DependencyGraph) (*TaskScore, error) {
	node, exists := graph.Nodes[task.ID]
	if !exists {
		return nil, fmt.Errorf("task %s not found in dependency graph", task.ID)
	}

	hierarchyDepth := mc.calculateHierarchyDepth(task, graph)

	score := &TaskScore{
		Task:               task,
		DependentCount:     node.DependentCount,
		UnblockedTaskCount: node.UnblockedCount,
		DependencyDepth:    node.DependencyDepth,
		CriticalPathLength: node.CriticalPathLength,
		HierarchyDepth:     hierarchyDepth,
		Priority:           task.Priority,
		CalculatedAt:       time.Now(),
	}

	return score, nil
}

// calculateHierarchyDepth determines how deep a task is in the parent-child hierarchy
func (mc *MetricsCalculator) calculateHierarchyDepth(task *types.Task, graph *DependencyGraph) int {
	depth := 0
	current := task
	visited := make(map[uuid.UUID]bool)

	for current.ParentID != nil && !visited[current.ID] {
		visited[current.ID] = true
		depth++
		if parentNode, exists := graph.Nodes[*current.ParentID]; exists {
			current = parentNode.Task
		} else {
			break
		}
	}

	return depth
}