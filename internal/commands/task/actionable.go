package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/denkhaus/knot/v2/internal/selection"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/utils"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// ActionableAction finds the next actionable task using dependency-aware selection
func ActionableAction() cli.ActionFunc {

	return func(c *cli.Context) error {
		container := shared.GetContainerFromCLIContext(c)
		projectManager := container.GetProjectManager()
		loggerService := container.GetLogger()

		projectID, err := projectManager.ResolveProjectID(c.Context)
		if err != nil {
			return err
		}

		loggerService.Info("Finding next actionable task with enhanced selection",
			zap.String("projectID", projectID.String()))

		// Get all tasks in the project
		allTasks, err := projectManager.ListTasksForProject(context.Background(), projectID)
		if err != nil {
			loggerService.Error("Failed to get project tasks", zap.Error(err))
			return fmt.Errorf("failed to get project tasks: %w", err)
		}

		// Parse strategy from CLI flag or auto-recommend if not provided
		var strategy selection.Strategy
		var strategyReason string

		if c.IsSet("strategy") {
			// User explicitly provided a strategy
			strategyStr := c.String("strategy")
			strategy = selection.ParseStrategy(strategyStr)
			strategyReason = fmt.Sprintf("User-selected %s strategy", strategy.String())
		} else {
			// Auto-recommend strategy based on project analysis
			recommendedStrategy, reason, err := selection.AnalyzeProjectAndRecommendStrategy(allTasks)
			if err != nil {
				loggerService.Warn("Failed to analyze project for strategy recommendation, using dependency-aware", zap.Error(err))
				strategy = selection.StrategyDependencyAware
				strategyReason = "Using default dependency-aware strategy (analysis failed)"
			} else {
				strategy = recommendedStrategy
				strategyReason = reason
			}
		}

		// Get configuration
		config := selection.DefaultConfig()
		config.Strategy = strategy

		// Apply configuration overrides from CLI flags
		if c.Bool("allow-parent-with-subtasks") {
			config.Behavior.AllowParentWithSubtasks = true
		}
		if c.Bool("prefer-pending") {
			config.Behavior.PreferInProgress = false
		}

		// Create selector
		selector, err := selection.NewTaskSelector(strategy, config)
		if err != nil {
			loggerService.Error("Failed to create task selector", zap.Error(err))
			return fmt.Errorf("failed to create task selector: %w", err)
		}

		// Select next actionable task
		selectedTask, err := selector.SelectNextActionableTask(allTasks)
		if err != nil {
			// Handle specific error types
			if selErr, ok := err.(*selection.SelectionError); ok {
				switch selErr.Type {
				case selection.ErrorTypeNoTasks:
					fmt.Println("No tasks found in project")
					return nil
				case selection.ErrorTypeNoActionable:
					fmt.Println("No actionable tasks available")
					return nil
				case selection.ErrorTypeDeadlock:
					fmt.Printf("No actionable tasks found: %s\n", selErr.Message)
					return nil
				case selection.ErrorTypeCircularDep:
					fmt.Printf("Circular dependencies detected: %s\n", selErr.Message)
					fmt.Println("Please resolve the circular dependencies before continuing")
					return nil
				default:
					return fmt.Errorf("task selection failed: %w", err)
				}
			}
			return fmt.Errorf("failed to select actionable task: %w", err)
		}

		// Get selection result for additional context
		result := selector.GetLastResult()

		// Output JSON if requested
		if c.Bool("json") {
			output := map[string]any{
				"task":            selectedTask,
				"strategy":        strategy.String(),
				"strategy_reason": strategyReason,
				"reason":          result.Reason,
				"score":           result.Score.Score,
				"execution_time":  result.ExecutionTime.String(),
			}

			if c.Bool("verbose") && len(result.Alternatives) > 0 {
				output["alternatives"] = result.Alternatives[:min(5, len(result.Alternatives))]
			}

			jsonData, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal result to JSON: %w", err)
			}
			fmt.Println(string(jsonData))
			return nil
		}

		// Show project context indicator
		shared.ShowProjectContextWithSeparator(c)

		// Output formatted text
		fmt.Printf("Next actionable task (strategy: %s):\n\n", strategy.String())
		fmt.Printf("* %s (ID: %s)\n", selectedTask.Title, selectedTask.ID)

		if selectedTask.Description != "" {
			fmt.Printf("%s\n", strings.Repeat("═", 120))
			wrappedDesc := utils.WrapText(selectedTask.Description, 120)
			for _, line := range wrappedDesc {
				fmt.Printf("  %s\n", line)
			}
			fmt.Printf("%s\n", strings.Repeat("═", 120))
		}

		fmt.Printf("  State: %s | Complexity: %d | Priority: %d\n",
			selectedTask.State, selectedTask.Complexity, selectedTask.Priority)

		if selectedTask.Depth > 0 {
			fmt.Printf("  Depth: %d", selectedTask.Depth)
			if selectedTask.ParentID != nil {
				fmt.Printf(" | Parent: %s", *selectedTask.ParentID)
			}
			fmt.Println()
		}

		// Show strategy reasoning and selection reasoning
		fmt.Printf("\nStrategy: %s\n", strategyReason)
		fmt.Printf("Selection reason: %s\n", result.Reason)

		if result.Score.UnblockedTaskCount > 0 {
			fmt.Printf("Will unblock: %d task(s)\n", result.Score.UnblockedTaskCount)
		}

		if result.Score.DependentCount > 0 {
			fmt.Printf("Dependent tasks: %d\n", result.Score.DependentCount)
		}

		// Show alternatives if verbose mode and available
		if c.Bool("verbose") && len(result.Alternatives) > 0 {
			fmt.Printf("\nAlternatives considered:\n")
			for i, alt := range result.Alternatives[:min(3, len(result.Alternatives))] {
				fmt.Printf("  %d. %s (score: %.2f)\n", i+1, alt.Task.Title, alt.Score)
			}
		}

		fmt.Printf("\nExecution time: %v\n", result.ExecutionTime)

		return nil
	}
}
