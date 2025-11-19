package visualization

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/denkhaus/knot/v2/internal/types"
)

// Renderer handles output rendering for different formats
type Renderer struct {
	config *VisualizationConfig
	output []string
}

// NewRenderer creates a new renderer
func NewRenderer(config *VisualizationConfig) *Renderer {
	return &Renderer{
		config: config,
		output: make([]string, 0),
	}
}

// RenderProjectOverview renders project overview
func (r *Renderer) RenderProjectOverview(result *ProjectAnalysisResult) error {
	r.addLine("[LINK] PROJECT DEPENDENCY OVERVIEW")
	r.addLine(strings.Repeat("-", 55))

	// Statistics section
	r.renderStatistics(result)

	// Cycles section
	r.renderCycles(result.Cycles)

	// Root tasks section
	if len(result.RootTasks) > 0 {
		r.renderRootTasks(result.RootTasks)
	}

	// Blocked tasks section
	if result.BlockedTasks > 0 {
		r.renderBlockedTasks(result)
	}

	return nil
}

// RenderTaskAnalysis renders task-specific analysis
func (r *Renderer) RenderTaskAnalysis(result *TaskAnalysisResult) error {
	r.addLine(fmt.Sprintf("[LINK] DEPENDENCY ANALYSIS: %s", result.Task.Title))
	r.addLine(strings.Repeat("-", 65))
	r.addLine(fmt.Sprintf("Task ID: %s | State: %s | Priority: %d",
		result.Task.ID, result.Task.State, result.Task.Priority))
	r.addLine("")

	// Status indicators
	r.renderTaskStatus(result)

	// Upstream dependencies
	r.renderUpstreamDependencies(result)

	// Downstream dependents
	r.renderDownstreamDependents(result)

	// Blocking relationships
	if len(result.BlockingTasks) > 0 {
		r.renderBlockingTasks(result.BlockingTasks)
	}

	return nil
}

// RenderTree renders dependency tree
func (r *Renderer) RenderTree(result *ProjectAnalysisResult) error {
	r.addLine("[TREE] DEPENDENCY TREE")
	r.addLine(strings.Repeat("-", 45))

	if len(result.RootTasks) == 0 {
		r.addLine("  No root tasks found (possible circular dependencies)")
		return nil
	}

	r.addLine("[ROOT] Root tasks (no dependencies):")
	for _, root := range result.RootTasks {
		r.addLine(fmt.Sprintf("  %s %s (ID: %s)", IconFolder, root.Title, root.ID))
		r.renderTaskTree(root, result.AllRelationships, 1)
	}

	return nil
}

// RenderGraph renders dependency graph
func (r *Renderer) RenderGraph(result *ProjectAnalysisResult) error {
	r.addLine("[GRAPH] DEPENDENCY GRAPH")
	r.addLine(strings.Repeat("-", 45))

	r.addLine("Task relationships:")
	for _, rel := range result.AllRelationships {
		icon := IconDependency
		if rel.IsCircular {
			icon = IconCycle
		}
		r.addLine(fmt.Sprintf("  %s %s %s %s",
			rel.FromTask.Title, icon, icon, rel.ToTask.Title))
	}

	return nil
}

// RenderJSON renders analysis as JSON
func (r *Renderer) RenderJSON(taskResult *TaskAnalysisResult, projectResult *ProjectAnalysisResult) error {
	var data interface{}

	if taskResult != nil {
		data = map[string]interface{}{
			"type":  "task_analysis",
			"task":  r.taskToMap(taskResult.Task),
			"upstream": r.tasksToMap(taskResult.UpstreamTasks),
			"downstream": r.tasksToMap(taskResult.DownstreamTasks),
			"blocked": taskResult.IsBlocked,
			"blocking": r.tasksToMap(taskResult.BlockingTasks),
			"in_cycle": taskResult.InCycle,
		}
	} else {
		data = map[string]interface{}{
			"type":           "project_analysis",
			"total_tasks":    projectResult.TotalTasks,
			"tasks_with_deps": projectResult.TasksWithDeps,
			"blocked_tasks":  projectResult.BlockedTasks,
			"completed":      projectResult.CompletedTasks,
			"in_progress":    projectResult.InProgressTasks,
			"pending":        projectResult.PendingTasks,
			"cycles":         projectResult.Cycles,
			"root_tasks":     r.tasksToMap(projectResult.RootTasks),
		}
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// Private helper methods

func (r *Renderer) renderStatistics(result *ProjectAnalysisResult) {
	r.addLine("[STATS] Statistics:")
	r.addLine(fmt.Sprintf("  * Total tasks: %d", result.TotalTasks))
	r.addLine(fmt.Sprintf("  * Tasks with dependencies: %d", result.TasksWithDeps))
	r.addLine(fmt.Sprintf("  * Blocked tasks: %d", result.BlockedTasks))
	r.addLine(fmt.Sprintf("  * Completed: %d", result.CompletedTasks))
	r.addLine(fmt.Sprintf("  * In Progress: %d", result.InProgressTasks))
	r.addLine(fmt.Sprintf("  * Pending: %d", result.PendingTasks))
	r.addLine(fmt.Sprintf("  * Circular dependencies: %d", len(result.Cycles)))
	r.addLine("")
}

func (r *Renderer) renderCycles(cycles [][]string) {
	if len(cycles) > 0 {
		r.addLine("[!] Circular Dependencies:")
		for i, cycle := range cycles {
			r.addLine(fmt.Sprintf("  Cycle %d: %s", i+1, strings.Join(cycle, " -> ")))
		}
		r.addLine("")
	} else {
		r.addLine("[OK] No circular dependencies detected")
		r.addLine("")
	}
}

func (r *Renderer) renderRootTasks(roots []*types.Task) {
	r.addLine("[ROOT] Root Tasks (no dependencies):")
	for i, root := range roots {
		statusIcon := r.getTaskIcon(root, root.ID)
		r.addLine(fmt.Sprintf("  %d. %s %s (ID: %s)",
			i+1, statusIcon, root.Title, root.ID))
	}
	r.addLine("")
}

func (r *Renderer) renderBlockedTasks(result *ProjectAnalysisResult) error {
	r.addLine("[BLOCKED] Blocked Tasks:")
	r.addLine("")

	blockedCount := 0
	for _, task := range result.AllTasks {
		if r.isTaskBlocked(task) {
			blockedCount++
			if blockedCount <= 10 { // Limit to first 10 for display
				r.addLine(fmt.Sprintf("  %d. %s %s (ID: %s) [Priority: %d]",
					blockedCount, IconBlocked, task.Title, task.ID, task.Priority))
			}
		}
	}

	if blockedCount > 10 {
		r.addLine(fmt.Sprintf("  ... and %d more blocked tasks", blockedCount-10))
	}
	r.addLine("")
	return nil
}

func (r *Renderer) renderTaskStatus(result *TaskAnalysisResult) {
	status := "Ready"
	if result.IsBlocked {
		status = "Blocked"
	} else if result.InCycle {
		status = "In Cycle"
	}

	r.addLine(fmt.Sprintf("Status: %s", status))
	r.addLine("")
}

func (r *Renderer) renderUpstreamDependencies(result *TaskAnalysisResult) {
	if len(result.UpstreamTasks) > 0 {
		r.addLine("[UP] Upstream Dependencies (what this task depends on):")
		for i, dep := range result.UpstreamTasks {
			icon := r.getTaskIcon(dep, dep.ID)
			r.addLine(fmt.Sprintf("  %d. %s %s (ID: %s)",
				i+1, icon, dep.Title, dep.ID))
		}
	} else {
		r.addLine("[UP] No upstream dependencies")
	}
	r.addLine("")
}

func (r *Renderer) renderDownstreamDependents(result *TaskAnalysisResult) {
	if len(result.DownstreamTasks) > 0 {
		r.addLine("[DOWN] Downstream Dependents (tasks that depend on this):")
		for i, dep := range result.DownstreamTasks {
			icon := r.getTaskIcon(dep, dep.ID)
			r.addLine(fmt.Sprintf("  %d. %s %s (ID: %s)",
				i+1, icon, dep.Title, dep.ID))
		}
	} else {
		r.addLine("[DOWN] No downstream dependents")
	}
	r.addLine("")
}

func (r *Renderer) renderBlockingTasks(blockingTasks []*types.Task) {
	r.addLine("[BLOCKING] Blocking Tasks (preventing this task from starting):")
	for i, task := range blockingTasks {
		icon := r.getTaskIcon(task, task.ID)
		r.addLine(fmt.Sprintf("  %d. %s %s (ID: %s)",
			i+1, icon, task.Title, task.ID))
	}
	r.addLine("")
}

func (r *Renderer) renderTaskTree(task *types.Task, relationships []TaskRelationship, depth int) {
	if depth > r.config.MaxDepth {
		return
	}

	indent := strings.Repeat("  ", depth)

	// Find direct dependents
	dependents := r.findDirectDependents(task.ID, relationships)

	if len(dependents) == 0 {
		icon := r.getTaskIcon(task, task.ID)
		r.addLine(fmt.Sprintf("%s  %s %s", indent, icon, task.Title))
		return
	}

	// Sort dependents by priority
	sort.Slice(dependents, func(i, j int) bool {
		return dependents[i].Priority > dependents[j].Priority
	})

	for i, dependent := range dependents {
		isLast := i == len(dependents)-1
		connector := "+-"
		if isLast {
			connector += "-"
		} else {
			connector += "-"
		}

		icon := r.getTaskIcon(dependent, dependent.ID)
		r.addLine(fmt.Sprintf("%s  %s%s %s", indent, icon, connector, dependent.Title))
		r.renderTaskTree(dependent, relationships, depth+1)
	}
}

func (r *Renderer) getTaskIcon(task *types.Task, _ interface{}) TaskIcon {
	switch task.State {
	case "completed":
		return IconCompleted
	case "in-progress":
		return IconInProgress
	case "pending":
		if r.isTaskBlocked(task) {
			return IconBlocked
		}
		return IconReady
	default:
		return IconUnknown
	}
}

func (r *Renderer) isTaskBlocked(task *types.Task) bool {
	// This would need analyzer instance - simplified for now
	return len(task.Dependencies) > 0 && task.State == "pending"
}

func (r *Renderer) findDirectDependents(taskID interface{}, relationships []TaskRelationship) []*types.Task {
	var dependents []*types.Task
	for _, rel := range relationships {
		if rel.FromTask.ID.String() == fmt.Sprintf("%v", taskID) {
			dependents = append(dependents, rel.ToTask)
		}
	}
	return dependents
}

func (r *Renderer) addLine(line string) {
	r.output = append(r.output, line)
}

// Render outputs the current visualization
func (r *Renderer) Render() error {
	fmt.Println(strings.Join(r.output, "\n"))
	return nil
}

// Helper conversion methods
func (r *Renderer) taskToMap(task *types.Task) map[string]interface{} {
	if task == nil {
		return nil
	}
	return map[string]interface{}{
		"id":         task.ID.String(),
		"title":      task.Title,
		"state":      task.State,
		"priority":   task.Priority,
		"complexity": task.Complexity,
	}
}

func (r *Renderer) tasksToMap(tasks []*types.Task) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(tasks))
	for _, task := range tasks {
		if task != nil {
			result = append(result, r.taskToMap(task))
		}
	}
	return result
}