package selection

import (
	"fmt"
	"strings"
	"time"

	"github.com/denkhaus/knot/internal/types"
)

// TaskSelectionUtils provides utility functions for task selection
type TaskSelectionUtils struct{}

// FormatSelectionResult formats a selection result for display
func (u *TaskSelectionUtils) FormatSelectionResult(result *SelectionResult, verbose bool) string {
	if result == nil {
		return "No selection result available"
	}

	var output strings.Builder

	// Basic info
	output.WriteString(fmt.Sprintf("Selected: %s\n", result.SelectedTask.Title))
	output.WriteString(fmt.Sprintf("Strategy: %s\n", result.Strategy.String()))
	output.WriteString(fmt.Sprintf("Reason: %s\n", result.Reason))

	if verbose {
		output.WriteString(fmt.Sprintf("Score: %.2f\n", result.Score.Score))
		output.WriteString(fmt.Sprintf("Selection time: %v\n", result.ExecutionTime))

		// Score breakdown
		output.WriteString("\nScore Details:\n")
		output.WriteString(fmt.Sprintf("  Priority: %d\n", result.Score.Priority))
		output.WriteString(fmt.Sprintf("  Dependent count: %d\n", result.Score.DependentCount))
		output.WriteString(fmt.Sprintf("  Unblocked count: %d\n", result.Score.UnblockedTaskCount))
		output.WriteString(fmt.Sprintf("  Hierarchy depth: %d\n", result.Score.HierarchyDepth))
		output.WriteString(fmt.Sprintf("  Critical path length: %d\n", result.Score.CriticalPathLength))

		// Alternatives
		if len(result.Alternatives) > 0 {
			output.WriteString(fmt.Sprintf("\nAlternatives (%d):\n", len(result.Alternatives)))
			for i, alt := range result.Alternatives {
				if i >= 5 { // Limit display to top 5 alternatives
					output.WriteString("  ...\n")
					break
				}
				output.WriteString(fmt.Sprintf("  %d. %s (score: %.2f)\n", i+1, alt.Task.Title, alt.Score))
			}
		}
	}

	return output.String()
}

// FormatDependencyGraph formats a dependency graph for display
func (u *TaskSelectionUtils) FormatDependencyGraph(graph *DependencyGraph, showDetails bool) string {
	if graph == nil {
		return "No dependency graph available"
	}

	var output strings.Builder

	output.WriteString("Dependency Graph Analysis:\n")
	output.WriteString(fmt.Sprintf("  Total tasks: %d\n", graph.TaskCount))
	output.WriteString(fmt.Sprintf("  Actionable tasks: %d\n", graph.ActionableCount))
	output.WriteString(fmt.Sprintf("  Root tasks: %d\n", len(graph.RootTasks)))
	output.WriteString(fmt.Sprintf("  Leaf tasks: %d\n", len(graph.LeafTasks)))

	if graph.HasCycles {
		output.WriteString(fmt.Sprintf("  ⚠️  Cycles detected in %d tasks\n", len(graph.CyclicTasks)))
	}

	if len(graph.CriticalPath) > 0 {
		output.WriteString(fmt.Sprintf("  Critical path length: %d\n", len(graph.CriticalPath)))
	}

	if showDetails {
		output.WriteString("\nTask Details:\n")
		for _, node := range graph.Nodes {
			status := "✓"
			if !node.IsActionable {
				status = "✗"
			}

			output.WriteString(fmt.Sprintf("  %s %s\n", status, node.Task.Title))
			if len(node.Dependencies) > 0 {
				output.WriteString(fmt.Sprintf("    Dependencies: %d\n", len(node.Dependencies)))
			}
			if len(node.Dependents) > 0 {
				output.WriteString(fmt.Sprintf("    Dependents: %d\n", len(node.Dependents)))
			}
			if len(node.BlockingReasons) > 0 {
				output.WriteString(fmt.Sprintf("    Blocked: %s\n", strings.Join(node.BlockingReasons, ", ")))
			}
		}
	}

	return output.String()
}

// FormatTaskScore formats a task score for display
func (u *TaskSelectionUtils) FormatTaskScore(score *TaskScore) string {
	return fmt.Sprintf("%s (score: %.2f, priority: %d, dependents: %d)",
		score.Task.Title, score.Score, score.Priority, score.DependentCount)
}

// GenerateSelectionSummary creates a summary of selection statistics
func (u *TaskSelectionUtils) GenerateSelectionSummary(tasks []*types.Task, graph *DependencyGraph) string {
	var output strings.Builder

	// Count tasks by state
	pendingCount := 0
	inProgressCount := 0
	completedCount := 0

	for _, task := range tasks {
		switch task.State {
		case types.TaskStatePending:
			pendingCount++
		case types.TaskStateInProgress:
			inProgressCount++
		case types.TaskStateCompleted:
			completedCount++
		}
	}

	// Calculate completion percentage
	total := len(tasks)
	completionPct := float64(completedCount) / float64(total) * 100

	output.WriteString("Project Summary:\n")
	output.WriteString(fmt.Sprintf("  Total tasks: %d\n", total))
	output.WriteString(fmt.Sprintf("  Completed: %d (%.1f%%)\n", completedCount, completionPct))
	output.WriteString(fmt.Sprintf("  In progress: %d\n", inProgressCount))
	output.WriteString(fmt.Sprintf("  Pending: %d\n", pendingCount))

	if graph != nil {
		output.WriteString(fmt.Sprintf("  Actionable: %d\n", graph.ActionableCount))

		if graph.ActionableCount > 0 {
			actionablePct := float64(graph.ActionableCount) / float64(pendingCount+inProgressCount) * 100
			output.WriteString(fmt.Sprintf("  Actionable rate: %.1f%%\n", actionablePct))
		}

		if graph.HasCycles {
			output.WriteString("  ⚠️  Circular dependencies detected\n")
		}
	}

	return output.String()
}

// ProjectAnalyzer analyzes project characteristics
type ProjectAnalyzer struct{}

// AnalyzeProject analyzes project characteristics to recommend optimal configuration
func (pa *ProjectAnalyzer) AnalyzeProject(tasks []*types.Task) *ProjectCharacteristics {
	characteristics := &ProjectCharacteristics{
		TaskCount:  len(tasks),
		AnalyzedAt: time.Now(),
	}

	if len(tasks) == 0 {
		return characteristics
	}

	// Analyze task dependencies
	totalDependencies := 0
	maxDependencies := 0
	tasksWithDependencies := 0

	// Analyze hierarchy
	maxHierarchyDepth := 0
	tasksWithParents := 0

	// Analyze priorities
	prioritySum := 0
	highPriorityCount := 0

	for _, task := range tasks {
		// Dependencies
		depCount := len(task.Dependencies)
		totalDependencies += depCount
		if depCount > 0 {
			tasksWithDependencies++
		}
		if depCount > maxDependencies {
			maxDependencies = depCount
		}

		// Hierarchy
		if task.ParentID != nil {
			tasksWithParents++
		}
		depth := pa.calculateTaskHierarchyDepth(task, tasks)
		if depth > maxHierarchyDepth {
			maxHierarchyDepth = depth
		}

		// Priority
		prioritySum += int(task.Priority)
		if priorityToScore(task.Priority) >= 2 {
			highPriorityCount++
		}
	}

	// Calculate averages
	characteristics.AverageDependencies = float64(totalDependencies) / float64(len(tasks))
	characteristics.MaxDependencies = maxDependencies
	characteristics.DependencyRatio = float64(tasksWithDependencies) / float64(len(tasks))

	characteristics.MaxHierarchyDepth = maxHierarchyDepth
	characteristics.HierarchyRatio = float64(tasksWithParents) / float64(len(tasks))
	characteristics.HasHierarchy = maxHierarchyDepth > 0

	characteristics.AveragePriority = float64(prioritySum) / float64(len(tasks))
	characteristics.HighPriorityRatio = float64(highPriorityCount) / float64(len(tasks))

	// Determine complexity
	characteristics.Complexity = pa.determineComplexity(characteristics)

	return characteristics
}

// calculateTaskHierarchyDepth calculates the hierarchy depth for a single task
func (pa *ProjectAnalyzer) calculateTaskHierarchyDepth(task *types.Task, allTasks []*types.Task) int {
	depth := 0
	current := task
	visited := make(map[string]bool)

	taskMap := make(map[string]*types.Task)
	for _, t := range allTasks {
		taskMap[t.ID.String()] = t
	}

	for current.ParentID != nil && !visited[current.ID.String()] {
		visited[current.ID.String()] = true
		depth++
		if parent, exists := taskMap[current.ParentID.String()]; exists {
			current = parent
		} else {
			break
		}
	}

	return depth
}

// determineComplexity determines project complexity based on characteristics
func (pa *ProjectAnalyzer) determineComplexity(char *ProjectCharacteristics) ProjectComplexity {
	if char.TaskCount < 10 {
		return ComplexitySimple
	}

	if char.TaskCount > 50 || char.AverageDependencies > 2.0 || char.MaxHierarchyDepth > 3 {
		return ComplexityComplex
	}

	return ComplexityMedium
}

// RecommendStrategy recommends the best selection strategy based on project characteristics
func (pa *ProjectAnalyzer) RecommendStrategy(char *ProjectCharacteristics) (Strategy, string) {
	if char.TaskCount < 5 {
		return StrategyCreationOrder, "Simple project - creation order is sufficient"
	}

	if char.DependencyRatio > 0.7 || char.AverageDependencies > 2.0 {
		return StrategyDependencyAware, "High dependency complexity - focus on unblocking tasks"
	}

	if char.HasHierarchy && char.HierarchyRatio > 0.5 {
		return StrategyDepthFirst, "Hierarchical structure - complete branches systematically"
	}

	if char.HighPriorityRatio > 0.3 {
		return StrategyPriority, "Many high-priority tasks - focus on urgent work"
	}

	return StrategyDependencyAware, "Balanced approach for general project management"
}

// ProjectCharacteristics holds analyzed project characteristics
type ProjectCharacteristics struct {
	TaskCount           int               `json:"task_count"`
	AverageDependencies float64           `json:"average_dependencies"`
	MaxDependencies     int               `json:"max_dependencies"`
	DependencyRatio     float64           `json:"dependency_ratio"` // Ratio of tasks with dependencies
	HasHierarchy        bool              `json:"has_hierarchy"`
	MaxHierarchyDepth   int               `json:"max_hierarchy_depth"`
	HierarchyRatio      float64           `json:"hierarchy_ratio"` // Ratio of tasks with parents
	AveragePriority     float64           `json:"average_priority"`
	HighPriorityRatio   float64           `json:"high_priority_ratio"` // Ratio of high-priority tasks
	Complexity          ProjectComplexity `json:"complexity"`
	AnalyzedAt          time.Time         `json:"analyzed_at"`
}

// ProjectComplexity represents the complexity level of a project
type ProjectComplexity int

const (
	ComplexitySimple ProjectComplexity = iota
	ComplexityMedium
	ComplexityComplex
)

// String returns the string representation of project complexity
func (c ProjectComplexity) String() string {
	switch c {
	case ComplexitySimple:
		return "simple"
	case ComplexityMedium:
		return "medium"
	case ComplexityComplex:
		return "complex"
	default:
		return "unknown"
	}
}

// TaskStateCounter counts tasks by their states
type TaskStateCounter struct{}

// CountByState counts tasks by their current state
func (tsc *TaskStateCounter) CountByState(tasks []*types.Task) map[types.TaskState]int {
	counts := make(map[types.TaskState]int)

	for _, task := range tasks {
		counts[task.State]++
	}

	return counts
}

// CountActionable counts how many tasks are currently actionable
func (tsc *TaskStateCounter) CountActionable(tasks []*types.Task, analyzer DependencyAnalyzer) int {
	count := 0
	for _, task := range tasks {
		if analyzer.ValidateActionability(task, tasks) {
			count++
		}
	}
	return count
}

// PerformanceMonitor tracks selection performance metrics
type PerformanceMonitor struct {
	selections []SelectionMetrics
}

// SelectionMetrics holds metrics for a single selection operation
type SelectionMetrics struct {
	TaskCount       int           `json:"task_count"`
	ActionableCount int           `json:"actionable_count"`
	Strategy        Strategy      `json:"strategy"`
	ExecutionTime   time.Duration `json:"execution_time"`
	MemoryUsage     int64         `json:"memory_usage"`
	Timestamp       time.Time     `json:"timestamp"`
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		selections: make([]SelectionMetrics, 0),
	}
}

// RecordSelection records metrics for a selection operation
func (pm *PerformanceMonitor) RecordSelection(result *SelectionResult, taskCount, actionableCount int, memoryUsage int64) {
	metrics := SelectionMetrics{
		TaskCount:       taskCount,
		ActionableCount: actionableCount,
		Strategy:        result.Strategy,
		ExecutionTime:   result.ExecutionTime,
		MemoryUsage:     memoryUsage,
		Timestamp:       result.SelectedAt,
	}

	pm.selections = append(pm.selections, metrics)

	// Keep only last 100 selections
	if len(pm.selections) > 100 {
		pm.selections = pm.selections[1:]
	}
}

// GetAverageExecutionTime returns average execution time for a strategy
func (pm *PerformanceMonitor) GetAverageExecutionTime(strategy Strategy) time.Duration {
	var total time.Duration
	count := 0

	for _, selection := range pm.selections {
		if selection.Strategy == strategy {
			total += selection.ExecutionTime
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return total / time.Duration(count)
}

// GetMetrics returns all recorded metrics
func (pm *PerformanceMonitor) GetMetrics() []SelectionMetrics {
	return pm.selections
}
