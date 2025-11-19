// Package selection provides dependency-aware task selection capabilities for knot.
//
// This package implements intelligent task selection strategies that consider task dependencies,
// hierarchy, priorities, and project characteristics to recommend the optimal next task to work on.
//
// Key Components:
//   - TaskSelector: Main interface for task selection
//   - DependencyAnalyzer: Builds and analyzes task dependency graphs
//   - ScoringStrategy: Different strategies for scoring and ranking tasks
//   - TaskFilter: Filters tasks based on actionability criteria
//   - ConfigProvider: Manages configuration for selection behavior
//
// Usage Example:
//
//	// Create a selector with dependency-aware strategy
//	selector, err := NewTaskSelector(StrategyDependencyAware, DefaultConfig())
//	if err != nil {
//		return err
//	}
//
//	// Select the next actionable task
//	task, err := selector.SelectNextActionableTask(allTasks)
//	if err != nil {
//		return err
//	}
//
//	fmt.Printf("Next task to work on: %s\n", task.Title)
//	fmt.Printf("Reason: %s\n", selector.GetSelectionReason())
package selection

import (
	"fmt"

	"github.com/denkhaus/knot/v2/internal/types"
)

// Package-level convenience functions for easy integration

// SelectActionableTask is a convenience function for quick task selection with default configuration
func SelectActionableTask(tasks []*types.Task, strategy Strategy) (*types.Task, error) {
	selector, err := NewTaskSelector(strategy, DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create selector: %w", err)
	}

	return selector.SelectNextActionableTask(tasks)
}

// SelectActionableTaskWithConfig is a convenience function for task selection with custom configuration
func SelectActionableTaskWithConfig(tasks []*types.Task, config *Config) (*types.Task, error) {
	selector, err := NewTaskSelector(config.Strategy, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create selector: %w", err)
	}

	return selector.SelectNextActionableTask(tasks)
}

// AnalyzeProjectAndRecommendStrategy analyzes project characteristics and recommends the best strategy
func AnalyzeProjectAndRecommendStrategy(tasks []*types.Task) (Strategy, string, error) {
	analyzer := &ProjectAnalyzer{}
	characteristics := analyzer.AnalyzeProject(tasks)

	strategy, reason := analyzer.RecommendStrategy(characteristics)
	return strategy, reason, nil
}

// ValidateTaskDependencies checks for circular dependencies and missing references
func ValidateTaskDependencies(tasks []*types.Task) ([]ValidationError, error) {
	analyzer := NewDependencyAnalyzer(DefaultConfig())

	graph, err := analyzer.BuildDependencyGraph(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build dependency graph: %w", err)
	}

	var validationErrors []ValidationError

	// Check for circular dependencies
	if graph.HasCycles {
		for _, taskID := range graph.CyclicTasks {
			validationErrors = append(validationErrors, ValidationError{
				TaskID:  taskID,
				Message: "Task is part of a circular dependency",
				Type:    "circular_dependency",
			})
		}
	}

	// Check for missing dependencies
	taskMap := make(map[string]*types.Task)
	for _, task := range tasks {
		taskMap[task.ID.String()] = task
	}

	for _, task := range tasks {
		for _, depID := range task.Dependencies {
			if _, exists := taskMap[depID.String()]; !exists {
				validationErrors = append(validationErrors, ValidationError{
					TaskID:  task.ID,
					Message: fmt.Sprintf("Dependency %s not found", depID),
					Type:    "missing_dependency",
				})
			}
		}
	}

	return validationErrors, nil
}

// GetActionableTasks returns all currently actionable tasks
func GetActionableTasks(tasks []*types.Task, config *Config) ([]*types.Task, error) {
	if config == nil {
		config = DefaultConfig()
	}

	analyzer := NewDependencyAnalyzer(config)
	filter := NewTaskFilter(analyzer, config)

	return filter.FilterActionableTasks(tasks)
}

// ScoreTasks scores all actionable tasks using the specified strategy
func ScoreTasks(tasks []*types.Task, strategy Strategy, config *Config) ([]*TaskScore, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Override config strategy
	config.Strategy = strategy

	selector, err := NewTaskSelector(strategy, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create selector: %w", err)
	}

	// Build dependency graph
	graph, err := selector.analyzer.BuildDependencyGraph(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// Get actionable tasks
	actionableTasks, err := selector.filter.FilterActionableTasks(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to filter actionable tasks: %w", err)
	}

	// Score tasks
	return selector.scoreActionableTasks(actionableTasks, graph)
}

// CreateDependencyGraph creates a dependency graph for visualization or analysis
func CreateDependencyGraph(tasks []*types.Task) (*DependencyGraph, error) {
	analyzer := NewDependencyAnalyzer(DefaultConfig())
	return analyzer.BuildDependencyGraph(tasks)
}

// FormatTaskSelection formats a task selection result for display
func FormatTaskSelection(task *types.Task, reason string, verbose bool) string {
	if task == nil {
		return "No task selected"
	}

	utils := &TaskSelectionUtils{}
	result := &SelectionResult{
		SelectedTask: task,
		Reason:       reason,
		Score: &TaskScore{
			Task: task,
		},
	}

	return utils.FormatSelectionResult(result, verbose)
}
