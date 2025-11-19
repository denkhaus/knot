package selection

import (
	"fmt"
	"time"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
)

// DependencyGraphBuilder handles the construction of dependency graphs
type DependencyGraphBuilder struct {
	config *Config
}

// NewDependencyGraphBuilder creates a new graph builder
func NewDependencyGraphBuilder(config *Config) *DependencyGraphBuilder {
	return &DependencyGraphBuilder{
		config: config,
	}
}

// Build creates a comprehensive dependency graph from tasks
func (dgb *DependencyGraphBuilder) Build(tasks []*types.Task) (*DependencyGraph, error) {
	if len(tasks) == 0 {
		return dgb.createEmptyGraph(), nil
	}

	graph := dgb.initializeGraph(tasks)

	if err := dgb.buildNodes(graph, tasks); err != nil {
		return nil, fmt.Errorf("failed to build nodes: %w", err)
	}

	if err := dgb.buildRelationships(graph); err != nil {
		return nil, fmt.Errorf("failed to build relationships: %w", err)
	}

	return graph, nil
}

// createEmptyGraph creates an empty dependency graph
func (dgb *DependencyGraphBuilder) createEmptyGraph() *DependencyGraph {
	return &DependencyGraph{
		Nodes:           make(map[uuid.UUID]*DependencyNode),
		RootTasks:       make([]uuid.UUID, 0),
		LeafTasks:       make([]uuid.UUID, 0),
		CriticalPath:    make([]uuid.UUID, 0),
		CyclicTasks:     make([]uuid.UUID, 0),
		AnalyzedAt:      time.Now(),
		TaskCount:       0,
		ActionableCount: 0,
	}
}

// initializeGraph creates the initial graph structure
func (dgb *DependencyGraphBuilder) initializeGraph(tasks []*types.Task) *DependencyGraph {
	return &DependencyGraph{
		Nodes:           make(map[uuid.UUID]*DependencyNode, len(tasks)),
		RootTasks:       make([]uuid.UUID, 0),
		LeafTasks:       make([]uuid.UUID, 0),
		CriticalPath:    make([]uuid.UUID, 0),
		CyclicTasks:     make([]uuid.UUID, 0),
		AnalyzedAt:      time.Now(),
		TaskCount:       len(tasks),
		ActionableCount: 0,
	}
}

// buildNodes creates dependency nodes for all tasks
func (dgb *DependencyGraphBuilder) buildNodes(graph *DependencyGraph, tasks []*types.Task) error {
	for _, task := range tasks {
		node := &DependencyNode{
			TaskID:          task.ID,
			Task:            task,
			Dependencies:    make([]uuid.UUID, len(task.Dependencies)),
			Dependents:      make([]uuid.UUID, 0),
			Children:        make([]uuid.UUID, 0),
			Parent:          task.ParentID,
			BlockingReasons: make([]string, 0),
		}

		// Copy dependencies to avoid modifying the original
		copy(node.Dependencies, task.Dependencies)
		graph.Nodes[task.ID] = node
	}

	return nil
}

// buildRelationships establishes dependency and hierarchy relationships
func (dgb *DependencyGraphBuilder) buildRelationships(graph *DependencyGraph) error {
	// Build dependency and parent-child relationships
	for _, node := range graph.Nodes {
		// Process dependencies
		for _, depID := range node.Dependencies {
			if depNode, exists := graph.Nodes[depID]; exists {
				depNode.Dependents = append(depNode.Dependents, node.TaskID)
			} else {
				// Dependency not found - record as blocking reason
				node.BlockingReasons = append(node.BlockingReasons,
					fmt.Sprintf("dependency %s not found", depID))
			}
		}

		// Process parent-child relationships
		if node.Parent != nil {
			if parentNode, exists := graph.Nodes[*node.Parent]; exists {
				parentNode.Children = append(parentNode.Children, node.TaskID)
			}
		}
	}

	// Identify root and leaf tasks
	dgb.identifyRootAndLeafTasks(graph)

	return nil
}

// identifyRootAndLeafTasks finds tasks with no dependencies and no dependents
func (dgb *DependencyGraphBuilder) identifyRootAndLeafTasks(graph *DependencyGraph) {
	for taskID, node := range graph.Nodes {
		if len(node.Dependencies) == 0 {
			graph.RootTasks = append(graph.RootTasks, taskID)
		}
		if len(node.Dependents) == 0 {
			graph.LeafTasks = append(graph.LeafTasks, taskID)
		}
	}
}