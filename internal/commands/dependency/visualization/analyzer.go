package visualization

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/internal/manager"
	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
)

// Analyzer handles dependency analysis logic
type Analyzer struct {
	projectManager manager.ProjectManager
	taskMap        map[uuid.UUID]*types.Task
}

// NewAnalyzer creates a new dependency analyzer
func NewAnalyzer(projectManager manager.ProjectManager, tasks []*types.Task) *Analyzer {
	taskMap := make(map[uuid.UUID]*types.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	return &Analyzer{
		projectManager: projectManager,
		taskMap:        taskMap,
	}
}

// AnalyzeTask performs comprehensive task analysis
func (a *Analyzer) AnalyzeTask(taskID uuid.UUID) (*TaskAnalysisResult, error) {
	task, exists := a.taskMap[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	result := &TaskAnalysisResult{
		Task: task,
	}

	// Get upstream dependencies
	if dependencies, err := a.projectManager.GetTaskDependencies(context.Background(), taskID); err == nil {
		result.UpstreamTasks = dependencies
		result.Dependencies = a.buildRelationships(task, dependencies, RelationshipDependency)
	}

	// Get downstream dependents
	if dependents, err := a.projectManager.GetDependentTasks(context.Background(), taskID); err == nil {
		result.DownstreamTasks = dependents
		result.Dependents = a.buildRelationships(task, dependents, RelationshipBlocks)
	}

	// Determine if task is blocked
	result.IsBlocked = a.isTaskBlocked(task)

	// Find blocking tasks
	result.BlockingTasks = a.findBlockingTasks(task)

	// Check if in cycle
	result.InCycle = a.isTaskInCycle(taskID)

	return result, nil
}

// AnalyzeProject performs project-wide analysis
func (a *Analyzer) AnalyzeProject() (*ProjectAnalysisResult, error) {
	// Convert taskMap to slice
	allTasks := make([]*types.Task, 0, len(a.taskMap))
	for _, task := range a.taskMap {
		allTasks = append(allTasks, task)
	}

	result := &ProjectAnalysisResult{
		TotalTasks: len(a.taskMap),
		AllTasks:   allTasks,
	}

	// Analyze each task
	for _, task := range a.taskMap {
		// Count by state
		switch task.State {
		case "completed":
			result.CompletedTasks++
		case "in-progress":
			result.InProgressTasks++
		case "pending":
			result.PendingTasks++
		}

		// Count tasks with dependencies
		if len(task.Dependencies) > 0 {
			result.TasksWithDeps++
		}

		// Count blocked tasks
		if a.isTaskBlocked(task) {
			result.BlockedTasks++
		}

		// Find root tasks (no dependencies)
		if len(task.Dependencies) == 0 {
			result.RootTasks = append(result.RootTasks, task)
		}
	}

	// Build all relationships
	result.AllRelationships = a.buildAllRelationships()

	// Find cycles
	result.Cycles = a.findCycles()

	return result, nil
}

// Helper methods

func (a *Analyzer) buildRelationships(fromTask *types.Task, toTasks []*types.Task, relType RelationshipType) []TaskRelationship {
	relationships := make([]TaskRelationship, 0, len(toTasks))

	for _, toTask := range toTasks {
		relationship := TaskRelationship{
			FromTask:   fromTask,
			ToTask:     toTask,
			Type:       relType,
			IsCircular: a.areInCycle(fromTask.ID, toTask.ID),
		}
		relationships = append(relationships, relationship)
	}

	return relationships
}

func (a *Analyzer) buildAllRelationships() []TaskRelationship {
	var relationships []TaskRelationship

	for _, task := range a.taskMap {
		// Dependency relationships
		for _, depID := range task.Dependencies {
			if depTask, exists := a.taskMap[depID]; exists {
				rel := TaskRelationship{
					FromTask:   depTask,
					ToTask:     task,
					Type:       RelationshipDependency,
					IsCircular: a.areInCycle(depID, task.ID),
				}
				relationships = append(relationships, rel)
			}
		}
	}

	return relationships
}

func (a *Analyzer) isTaskBlocked(task *types.Task) bool {
	for _, depID := range task.Dependencies {
		if depTask, exists := a.taskMap[depID]; exists {
			if depTask.State != "completed" {
				return true
			}
		}
	}
	return false
}

func (a *Analyzer) findBlockingTasks(task *types.Task) []*types.Task {
	var blockingTasks []*types.Task

	for _, depID := range task.Dependencies {
		if depTask, exists := a.taskMap[depID]; exists {
			if depTask.State != "completed" {
				blockingTasks = append(blockingTasks, depTask)
			}
		}
	}

	return blockingTasks
}

func (a *Analyzer) isTaskInCycle(taskID uuid.UUID) bool {
	for _, cycle := range a.detectCycles() {
		for _, id := range cycle {
			if id == taskID {
				return true
			}
		}
	}
	return false
}

func (a *Analyzer) areInCycle(fromID, toID uuid.UUID) bool {
	for _, cycle := range a.detectCycles() {
		hasFrom := false
		hasTo := false

		for _, id := range cycle {
			if id == fromID {
				hasFrom = true
			}
			if id == toID {
				hasTo = true
			}
		}

		if hasFrom && hasTo {
			return true
		}
	}
	return false
}

func (a *Analyzer) findCycles() [][]string {
	cycles := a.detectCycles()
	result := make([][]string, 0, len(cycles))

	for _, cycle := range cycles {
		titles := make([]string, 0, len(cycle))
		for _, id := range cycle {
			if task, exists := a.taskMap[id]; exists {
				titles = append(titles, task.Title)
			} else {
				titles = append(titles, "Unknown Task")
			}
		}
		result = append(result, titles)
	}

	return result
}

func (a *Analyzer) detectCycles() [][]uuid.UUID {
	return a.findCyclesInTasks()
}

func (a *Analyzer) findCyclesInTasks() [][]uuid.UUID {
	tasks := make([]*types.Task, 0, len(a.taskMap))
	for _, task := range a.taskMap {
		tasks = append(tasks, task)
	}

	// Build adjacency list
	graph := make(map[uuid.UUID][]uuid.UUID)
	for _, task := range tasks {
		graph[task.ID] = task.Dependencies
	}

	var cycles [][]uuid.UUID
	visited := make(map[uuid.UUID]bool)
	recStack := make(map[uuid.UUID]bool)
	path := make([]uuid.UUID, 0)

	var dfs func(uuid.UUID) bool
	dfs = func(taskID uuid.UUID) bool {
		visited[taskID] = true
		recStack[taskID] = true
		path = append(path, taskID)

		for _, depID := range graph[taskID] {
			if !visited[depID] {
				if dfs(depID) {
					return true
				}
			} else if recStack[depID] {
				// Found a cycle - extract it from path
				cycleStart := -1
				for i, id := range path {
					if id == depID {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := make([]uuid.UUID, len(path)-cycleStart)
					copy(cycle, path[cycleStart:])
					cycles = append(cycles, cycle)
				}
				return true
			}
		}

		recStack[taskID] = false
		path = path[:len(path)-1]
		return false
	}

	// Check all nodes
	for _, task := range tasks {
		if !visited[task.ID] {
			dfs(task.ID)
		}

	}

	return cycles
}