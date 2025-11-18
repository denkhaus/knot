package task

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/internal/shared"
	"github.com/denkhaus/knot/internal/types"
	"github.com/denkhaus/knot/internal/utils"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// BlockedAction shows tasks that are blocked by dependencies
func BlockedAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		projectID, err := shared.ResolveProjectID(c, appCtx)
		if err != nil {
			return err
		}

		appCtx.Logger.Info("Finding blocked tasks", zap.String("projectID", projectID.String()))

		// Get all tasks in the project
		allTasks, err := appCtx.ProjectManager.ListTasksForProject(context.Background(), projectID)
		if err != nil {
			appCtx.Logger.Error("Failed to get project tasks", zap.Error(err))
			return fmt.Errorf("failed to get project tasks: %w", err)
		}

		// Create a map for quick task lookup
		taskMap := make(map[uuid.UUID]*types.Task)
		for _, task := range allTasks {
			taskMap[task.ID] = task
		}

		// Find blocked tasks (pending/in-progress with unmet dependencies)
		var blockedTasks []*types.Task
		for _, task := range allTasks {
			if task.State == types.TaskStatePending || task.State == types.TaskStateInProgress {
				if !utils.IsTaskReady(task, taskMap) && len(task.Dependencies) > 0 {
					blockedTasks = append(blockedTasks, task)
				}
			}
		}

		appCtx.Logger.Info("Blocked tasks found", zap.Int("count", len(blockedTasks)))

		// Show project context indicator
		shared.ShowProjectContextWithSeparator(c, appCtx)

		if len(blockedTasks) == 0 {
			fmt.Println("No blocked tasks found. All tasks are either ready, completed, or have no dependencies.")
			return nil
		}

		// Apply limit if specified
		limit := c.Int("limit")
		if limit > 0 && len(blockedTasks) > limit {
			fmt.Printf("Blocked tasks (showing %d of %d):\n\n", limit, len(blockedTasks))
			blockedTasks = blockedTasks[:limit]
		} else {
			fmt.Printf("Blocked tasks (%d):\n\n", len(blockedTasks))
		}

		for i, task := range blockedTasks {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, task.Title, task.ID)
			if task.Description != "" {
				fmt.Printf("   %s\n", task.Description)
			}
			fmt.Printf("   State: %s | Complexity: %d\n", task.State, task.Complexity)

			// Show blocking dependencies
			fmt.Printf("   Blocked by %d dependencies:\n", len(task.Dependencies))
			for _, depID := range task.Dependencies {
				if depTask, exists := taskMap[depID]; exists {
					fmt.Printf("     -> %s (ID: %s) - %s\n", depTask.Title, depTask.ID, depTask.State)
				} else {
					fmt.Printf("     -> Unknown task (ID: %s)\n", depID)
				}
			}
			fmt.Println()
		}

		return nil
	}
}
