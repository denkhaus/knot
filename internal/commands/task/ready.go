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

// ReadyAction shows tasks that are ready to work on (no blockers)
func ReadyAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		projectID, err := shared.ResolveProjectID(c, appCtx)
		if err != nil {
			return err
		}

		appCtx.Logger.Info("Finding ready tasks", zap.String("projectID", projectID.String()))

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

		// Find ready tasks (pending/in-progress with no blockers)
		var readyTasks []*types.Task
		for _, task := range allTasks {
			if task.State == types.TaskStatePending || task.State == types.TaskStateInProgress {
				if utils.IsTaskReady(task, taskMap) {
					readyTasks = append(readyTasks, task)
				}
			}
		}

		appCtx.Logger.Info("Ready tasks found", zap.Int("count", len(readyTasks)))

		if len(readyTasks) == 0 {
			if c.Bool("json") {
				fmt.Println("[]")
				return nil
			}
			fmt.Println("No ready tasks found. All tasks are either completed, blocked, or cancelled.")
			return nil
		}

		// Apply limit if specified
		limit := c.Int("limit")
		if limit > 0 && len(readyTasks) > limit {
			readyTasks = readyTasks[:limit]
		}

		// Check if JSON output is requested
		if c.Bool("json") {
			return utils.OutputTasksAsJSON(readyTasks)
		}

		// Show project context indicator
		shared.ShowProjectContextWithSeparator(c, appCtx)

		if limit > 0 && len(readyTasks) == limit {
			fmt.Printf("Ready work (showing %d of %d tasks with no blockers):\n\n", limit, len(readyTasks))
		} else {
			fmt.Printf("Ready work (%d tasks with no blockers):\n\n", len(readyTasks))
		}

		for i, task := range readyTasks {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, task.Title, task.ID)
			if task.Description != "" {
				fmt.Printf("   %s\n", task.Description)
			}
			fmt.Printf("   State: %s | Complexity: %d\n", task.State, task.Complexity)
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
