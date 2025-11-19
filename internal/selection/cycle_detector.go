package selection

import (
	"github.com/google/uuid"
)

// CycleDetector handles circular dependency detection in dependency graphs
type CycleDetector struct{}

// NewCycleDetector creates a new cycle detector
func NewCycleDetector() *CycleDetector {
	return &CycleDetector{}
}

// Detect analyzes a dependency graph for cycles
func (cd *CycleDetector) Detect(graph *DependencyGraph) {
	visited := make(map[uuid.UUID]bool)
	recStack := make(map[uuid.UUID]bool)

	for taskID := range graph.Nodes {
		if !visited[taskID] {
			if cd.detectCyclesDFS(graph, taskID, visited, recStack) {
				graph.HasCycles = true
			}
		}
	}
}

// detectCyclesDFS performs depth-first search to detect cycles
func (cd *CycleDetector) detectCyclesDFS(graph *DependencyGraph, taskID uuid.UUID, visited, recStack map[uuid.UUID]bool) bool {
	visited[taskID] = true
	recStack[taskID] = true

	node := graph.Nodes[taskID]
	for _, depID := range node.Dependencies {
		if _, exists := graph.Nodes[depID]; exists {
			if !visited[depID] {
				if cd.detectCyclesDFS(graph, depID, visited, recStack) {
					graph.CyclicTasks = append(graph.CyclicTasks, taskID)
					return true
				}
			} else if recStack[depID] {
				graph.CyclicTasks = append(graph.CyclicTasks, taskID)
				return true
			}
		}
	}

	recStack[taskID] = false
	return false
}

// HasCycles returns true if the graph contains cycles
func (cd *CycleDetector) HasCycles(graph *DependencyGraph) bool {
	return graph.HasCycles
}

// GetCyclicTasks returns the list of tasks involved in cycles
func (cd *CycleDetector) GetCyclicTasks(graph *DependencyGraph) []uuid.UUID {
	return graph.CyclicTasks
}