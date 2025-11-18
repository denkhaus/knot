package task

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/internal/shared"
	"github.com/denkhaus/knot/internal/types"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// BreakdownAction finds tasks that need breakdown based on complexity threshold
func BreakdownAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		projectID, err := shared.ResolveProjectID(c, appCtx)
		if err != nil {
			return err
		}

		// TODO: Get complexity threshold from config (default: 8)
		complexityThreshold := c.Int("threshold")
		if complexityThreshold == 0 {
			complexityThreshold = 8 // Default from original pkg/tools/project
		}

		appCtx.Logger.Info("Finding tasks needing breakdown",
			zap.String("projectID", projectID.String()),
			zap.Int("threshold", complexityThreshold))

		// Get all tasks in the project
		allTasks, err := appCtx.ProjectManager.ListTasksForProject(context.Background(), projectID)
		if err != nil {
			appCtx.Logger.Error("Failed to get project tasks", zap.Error(err))
			return fmt.Errorf("failed to get project tasks: %w", err)
		}

		// Find tasks with complexity >= threshold that have no children
		var needsBreakdown []*types.Task
		for _, task := range allTasks {
			if task.Complexity >= complexityThreshold {
				// Check if task has children by looking for tasks with this task as parent
				hasChildren := false
				for _, otherTask := range allTasks {
					if otherTask.ParentID != nil && *otherTask.ParentID == task.ID {
						hasChildren = true
						break
					}
				}
				if !hasChildren {
					needsBreakdown = append(needsBreakdown, task)
				}
			}
		}

		appCtx.Logger.Info("Tasks needing breakdown found", zap.Int("count", len(needsBreakdown)))

		if len(needsBreakdown) == 0 {
			fmt.Printf("No tasks need breakdown (complexity >= %d with no subtasks)\n", complexityThreshold)
			return nil
		}

		// Apply limit if specified
		limit := c.Int("limit")
		if limit > 0 && len(needsBreakdown) > limit {
			fmt.Printf("Tasks needing breakdown (showing %d of %d with complexity >= %d):\n\n",
				limit, len(needsBreakdown), complexityThreshold)
			needsBreakdown = needsBreakdown[:limit]
		} else {
			fmt.Printf("Tasks needing breakdown (%d tasks with complexity >= %d):\n\n",
				len(needsBreakdown), complexityThreshold)
		}

		for i, task := range needsBreakdown {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, task.Title, task.ID)
			if task.Description != "" {
				fmt.Printf("   %s\n", task.Description)
			}
			fmt.Printf("   State: %s | Complexity: %d (>= %d threshold)\n",
				task.State, task.Complexity, complexityThreshold)
			if task.Depth > 0 {
				fmt.Printf("   Depth: %d", task.Depth)
				if task.ParentID != nil {
					fmt.Printf(" | Parent: %s", *task.ParentID)
				}
				fmt.Println()
			}
			fmt.Println()
		}

		return nil
	}
}
