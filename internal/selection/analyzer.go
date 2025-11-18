package selection

import (
	"fmt"
	"time"

	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
)

// DefaultDependencyAnalyzer implements DependencyAnalyzer interface
type DefaultDependencyAnalyzer struct {
	config *Config
}

// NewDependencyAnalyzer creates a new dependency analyzer
func NewDependencyAnalyzer(config *Config) *DefaultDependencyAnalyzer {
	return &DefaultDependencyAnalyzer{
		config: config,
	}
}

// BuildDependencyGraph creates a comprehensive dependency analysis
func (da *DefaultDependencyAnalyzer) BuildDependencyGraph(tasks []*types.Task) (*DependencyGraph, error) {
	if len(tasks) == 0 {
		return &DependencyGraph{
			Nodes:           make(map[uuid.UUID]*DependencyNode),
			RootTasks:       make([]uuid.UUID, 0),
			LeafTasks:       make([]uuid.UUID, 0),
			CriticalPath:    make([]uuid.UUID, 0),
			AnalyzedAt:      time.Now(),
			TaskCount:       0,
			ActionableCount: 0,
		}, nil
	}

	graph := &DependencyGraph{
		Nodes:           make(map[uuid.UUID]*DependencyNode),
		RootTasks:       make([]uuid.UUID, 0),
		LeafTasks:       make([]uuid.UUID, 0),
		CriticalPath:    make([]uuid.UUID, 0),
		CyclicTasks:     make([]uuid.UUID, 0),
		AnalyzedAt:      time.Now(),
		TaskCount:       len(tasks),
		ActionableCount: 0,
	}

	// Initialize nodes
	if err := da.initializeNodes(graph, tasks); err != nil {
		return nil, fmt.Errorf("failed to initialize nodes: %w", err)
	}

	// Build relationships
	if err := da.buildRelationships(graph, tasks); err != nil {
		return nil, fmt.Errorf("failed to build relationships: %w", err)
	}

	// Detect cycles
	da.detectCycles(graph)

	// Calculate metrics
	if err := da.calculateMetrics(graph); err != nil {
		return nil, fmt.Errorf("failed to calculate metrics: %w", err)
	}

	// Find critical path
	da.findCriticalPath(graph)

	// Count actionable tasks
	da.countActionableTasks(graph, tasks)

	return graph, nil
}

// initializeNodes creates initial nodes for all tasks
func (da *DefaultDependencyAnalyzer) initializeNodes(graph *DependencyGraph, tasks []*types.Task) error {
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

		copy(node.Dependencies, task.Dependencies)
		graph.Nodes[task.ID] = node
	}

	return nil
}

// buildRelationships establishes dependency and hierarchy relationships
func (da *DefaultDependencyAnalyzer) buildRelationships(graph *DependencyGraph, tasks []*types.Task) error {
	for _, task := range tasks {
		node := graph.Nodes[task.ID]

		// Build dependency relationships
		for _, depID := range task.Dependencies {
			if depNode, exists := graph.Nodes[depID]; exists {
				depNode.Dependents = append(depNode.Dependents, task.ID)
			} else {
				// Dependency not found - this could be an error or external dependency
				node.BlockingReasons = append(node.BlockingReasons, fmt.Sprintf("dependency %s not found", depID))
			}
		}

		// Build parent-child relationships
		if task.ParentID != nil {
			if parentNode, exists := graph.Nodes[*task.ParentID]; exists {
				parentNode.Children = append(parentNode.Children, task.ID)
			}
		}
	}

	// Identify root and leaf tasks
	for taskID, node := range graph.Nodes {
		if len(node.Dependencies) == 0 {
			graph.RootTasks = append(graph.RootTasks, taskID)
		}
		if len(node.Dependents) == 0 {
			graph.LeafTasks = append(graph.LeafTasks, taskID)
		}
	}

	return nil
}

// detectCycles identifies circular dependencies in the graph
func (da *DefaultDependencyAnalyzer) detectCycles(graph *DependencyGraph) {
	visited := make(map[uuid.UUID]bool)
	recStack := make(map[uuid.UUID]bool)

	for taskID := range graph.Nodes {
		if !visited[taskID] {
			if da.detectCyclesDFS(graph, taskID, visited, recStack) {
				graph.HasCycles = true
			}
		}
	}
}

// detectCyclesDFS performs depth-first search to detect cycles
func (da *DefaultDependencyAnalyzer) detectCyclesDFS(graph *DependencyGraph, taskID uuid.UUID, visited, recStack map[uuid.UUID]bool) bool {
	visited[taskID] = true
	recStack[taskID] = true

	node := graph.Nodes[taskID]
	for _, depID := range node.Dependencies {
		if _, exists := graph.Nodes[depID]; exists {
			if !visited[depID] {
				if da.detectCyclesDFS(graph, depID, visited, recStack) {
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

// calculateMetrics computes various metrics for each node
func (da *DefaultDependencyAnalyzer) calculateMetrics(graph *DependencyGraph) error {
	for taskID := range graph.Nodes {
		node := graph.Nodes[taskID]

		// Basic counts
		node.DependentCount = len(node.Dependents)
		node.ChildCount = len(node.Children)

		// Dependency depth
		node.DependencyDepth = da.calculateDependencyDepth(graph, taskID, make(map[uuid.UUID]bool))

		// Critical path length
		node.CriticalPathLength = da.calculateCriticalPathLength(graph, taskID, make(map[uuid.UUID]bool))

		// Unblocked count
		node.UnblockedCount = da.calculateUnblockedCount(graph, taskID, make(map[uuid.UUID]bool))
	}

	return nil
}

// calculateDependencyDepth computes how deep a task is in dependency chains
func (da *DefaultDependencyAnalyzer) calculateDependencyDepth(graph *DependencyGraph, taskID uuid.UUID, visited map[uuid.UUID]bool) int {
	if visited[taskID] {
		return 0 // Circular dependency protection
	}

	if da.config.Advanced.MaxDependencyDepth > 0 && len(visited) >= da.config.Advanced.MaxDependencyDepth {
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
			depth := 1 + da.calculateDependencyDepth(graph, depID, visited)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	}

	return maxDepth
}

// calculateCriticalPathLength finds the longest dependency chain through this task
func (da *DefaultDependencyAnalyzer) calculateCriticalPathLength(graph *DependencyGraph, taskID uuid.UUID, visited map[uuid.UUID]bool) int {
	if visited[taskID] {
		return 0
	}

	visited[taskID] = true
	defer delete(visited, taskID)

	node := graph.Nodes[taskID]
	maxPath := 1 // This task itself

	for _, depID := range node.Dependents {
		if _, exists := graph.Nodes[depID]; exists {
			path := 1 + da.calculateCriticalPathLength(graph, depID, visited)
			if path > maxPath {
				maxPath = path
			}
		}
	}

	return maxPath
}

// calculateUnblockedCount counts how many tasks would become actionable if this task completes
func (da *DefaultDependencyAnalyzer) calculateUnblockedCount(graph *DependencyGraph, taskID uuid.UUID, visited map[uuid.UUID]bool) int {
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
			wouldBeActionable := true
			for _, depDepID := range depNode.Dependencies {
				if depDepID != taskID {
					if depDepNode, exists := graph.Nodes[depDepID]; exists {
						if depDepNode.Task.State != types.TaskStateCompleted {
							wouldBeActionable = false
							break
						}
					}
				}
			}

			if wouldBeActionable {
				count++ // This dependent would become actionable
				// Recursively count what this dependent would unblock
				count += da.calculateUnblockedCount(graph, depTaskID, visited)
			}
		}
	}

	return count
}

// findCriticalPath identifies the longest dependency chain in the graph
func (da *DefaultDependencyAnalyzer) findCriticalPath(graph *DependencyGraph) {
	maxLength := 0
	var criticalPath []uuid.UUID

	// Start from root tasks and find longest path
	for _, rootID := range graph.RootTasks {
		path := da.findLongestPath(graph, rootID, make(map[uuid.UUID]bool))
		if len(path) > maxLength {
			maxLength = len(path)
			criticalPath = path
		}
	}

	graph.CriticalPath = criticalPath
}

// findLongestPath finds the longest dependency path starting from a task
func (da *DefaultDependencyAnalyzer) findLongestPath(graph *DependencyGraph, taskID uuid.UUID, visited map[uuid.UUID]bool) []uuid.UUID {
	if visited[taskID] {
		return []uuid.UUID{}
	}

	visited[taskID] = true
	defer delete(visited, taskID)

	node := graph.Nodes[taskID]
	longestPath := []uuid.UUID{taskID}

	for _, depID := range node.Dependents {
		if _, exists := graph.Nodes[depID]; exists {
			path := da.findLongestPath(graph, depID, visited)
			if len(path)+1 > len(longestPath) {
				longestPath = append([]uuid.UUID{taskID}, path...)
			}
		}
	}

	return longestPath
}

// countActionableTasks determines how many tasks are currently actionable
func (da *DefaultDependencyAnalyzer) countActionableTasks(graph *DependencyGraph, tasks []*types.Task) {
	count := 0
	taskMap := make(map[uuid.UUID]*types.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	for _, node := range graph.Nodes {
		if da.ValidateActionability(node.Task, tasks) {
			node.IsActionable = true
			count++
		} else {
			da.addBlockingReasons(node, taskMap)
		}
	}

	graph.ActionableCount = count
}

// addBlockingReasons determines why a task is not actionable
func (da *DefaultDependencyAnalyzer) addBlockingReasons(node *DependencyNode, taskMap map[uuid.UUID]*types.Task) {
	if node.Task.State != types.TaskStatePending && node.Task.State != types.TaskStateInProgress {
		node.BlockingReasons = append(node.BlockingReasons, fmt.Sprintf("task state is %v", node.Task.State))
		return
	}

	// Check dependencies
	for _, depID := range node.Dependencies {
		if depTask, exists := taskMap[depID]; exists {
			if depTask.State != types.TaskStateCompleted {
				node.BlockingReasons = append(node.BlockingReasons, fmt.Sprintf("dependency %s is not completed", depTask.Title))
			}
		} else {
			node.BlockingReasons = append(node.BlockingReasons, fmt.Sprintf("dependency %s not found", depID))
		}
	}

	// Check for active subtasks (if not allowing parent with subtasks)
	if !da.config.Behavior.AllowParentWithSubtasks {
		for _, childID := range node.Children {
			if childTask, exists := taskMap[childID]; exists {
				if childTask.State == types.TaskStatePending || childTask.State == types.TaskStateInProgress {
					node.BlockingReasons = append(node.BlockingReasons, "has active subtasks")
					break
				}
			}
		}
	}
}

// CalculateTaskScore computes a detailed score for a task
func (da *DefaultDependencyAnalyzer) CalculateTaskScore(task *types.Task, graph *DependencyGraph) (*TaskScore, error) {
	node, exists := graph.Nodes[task.ID]
	if !exists {
		return nil, fmt.Errorf("task %s not found in dependency graph", task.ID)
	}

	score := &TaskScore{
		Task:               task,
		DependentCount:     node.DependentCount,
		UnblockedTaskCount: node.UnblockedCount,
		DependencyDepth:    node.DependencyDepth,
		CriticalPathLength: node.CriticalPathLength,
		HierarchyDepth:     da.calculateHierarchyDepth(task, graph),
		Priority:           task.Priority,
		CalculatedAt:       time.Now(),
	}

	return score, nil
}

// calculateHierarchyDepth determines how deep a task is in the parent-child hierarchy
func (da *DefaultDependencyAnalyzer) calculateHierarchyDepth(task *types.Task, graph *DependencyGraph) int {
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

// ValidateActionability checks if a task can be worked on right now
func (da *DefaultDependencyAnalyzer) ValidateActionability(task *types.Task, allTasks []*types.Task) bool {
	// Only pending or in-progress tasks can be actionable
	if task.State != types.TaskStatePending && task.State != types.TaskStateInProgress {
		return false
	}

	// Check dependencies
	if !da.areDependenciesMet(task, allTasks) {
		return false
	}

	// Check subtasks (if configured to disallow parent with subtasks)
	if !da.config.Behavior.AllowParentWithSubtasks && da.hasActiveSubtasks(task, allTasks) {
		return false
	}

	return true
}

// areDependenciesMet checks if all task dependencies are completed
func (da *DefaultDependencyAnalyzer) areDependenciesMet(task *types.Task, allTasks []*types.Task) bool {
	taskMap := make(map[uuid.UUID]*types.Task)
	for _, t := range allTasks {
		taskMap[t.ID] = t
	}

	for _, depID := range task.Dependencies {
		depTask, exists := taskMap[depID]
		if !exists {
			if da.config.Behavior.StrictDependencies {
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
func (da *DefaultDependencyAnalyzer) hasActiveSubtasks(task *types.Task, allTasks []*types.Task) bool {
	for _, t := range allTasks {
		if t.ParentID != nil && *t.ParentID == task.ID {
			if t.State == types.TaskStatePending || t.State == types.TaskStateInProgress {
				return true
			}
		}
	}
	return false
}
