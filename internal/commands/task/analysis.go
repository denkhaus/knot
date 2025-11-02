package task

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/internal/shared"
	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// analysis.go contains task analysis and discovery commands
// - actionable: find next actionable task
// - breakdown: find tasks needing breakdown
// - ready: show ready tasks (moved from commands.go)
// - blocked: show blocked tasks (moved from commands.go)

// isTaskReady checks if a task has all its dependencies completed
func isTaskReady(task *types.Task, taskMap map[uuid.UUID]*types.Task) bool {
	// If task has no dependencies, it's ready
	if len(task.Dependencies) == 0 {
		return true
	}

	// Check if all dependencies are completed
	for _, depID := range task.Dependencies {
		depTask, exists := taskMap[depID]
		if !exists || depTask.State != types.TaskStateCompleted {
			return false
		}
	}

	return true
}

// hasActiveSubtasks checks if a task has any active (pending/in-progress) subtasks
func hasActiveSubtasks(parent *types.Task, allTasks []*types.Task) bool {
	for _, task := range allTasks {
		if task.ParentID != nil && *task.ParentID == parent.ID {
			// Found a subtask, check if it's active
			if task.State == types.TaskStatePending || task.State == types.TaskStateInProgress {
				return true
			}
		}
	}
	return false
}

// ActionableAction finds the next actionable task in a project
func ActionableAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		projectID, err := shared.ResolveProjectID(c, appCtx)
		if err != nil {
			return err
		}

		appCtx.Logger.Info("Finding next actionable task", zap.String("projectID", projectID.String()))

		// Get all tasks in the project
		allTasks, err := appCtx.ProjectManager.ListTasksForProject(context.Background(), projectID)
		if err != nil {
			appCtx.Logger.Error("Failed to get project tasks", zap.Error(err))
			return fmt.Errorf("failed to get project tasks: %w", err)
		}

		// Create a map of task IDs to tasks for quick lookup
		taskMap := make(map[uuid.UUID]*types.Task)
		for _, task := range allTasks {
			taskMap[task.ID] = task
		}

		// Separate tasks by state
		var pendingTasks, inProgressTasks []*types.Task
		for _, task := range allTasks {
			switch task.State {
			case types.TaskStatePending:
				pendingTasks = append(pendingTasks, task)
			case types.TaskStateInProgress:
				inProgressTasks = append(inProgressTasks, task)
			}
		}

		// Show project context indicator
		shared.ShowProjectContextWithSeparator(c, appCtx)

		// Prioritize in-progress tasks first
		if len(inProgressTasks) > 0 {
			// For in-progress tasks, find one that has all its dependencies met and no active subtasks
			for _, task := range inProgressTasks {
				if isTaskReady(task, taskMap) && !hasActiveSubtasks(task, allTasks) {
					fmt.Printf("Next actionable task (in-progress):\n\n")
					fmt.Printf("* %s (ID: %s)\n", task.Title, task.ID)
					if task.Description != "" {
						fmt.Printf("  %s\n", task.Description)
					}
					fmt.Printf("  State: %s | Complexity: %d\n", task.State, task.Complexity)
					if task.Depth > 0 {
						fmt.Printf("  Depth: %d", task.Depth)
						if task.ParentID != nil {
							fmt.Printf(" | Parent: %s", *task.ParentID)
						}
						fmt.Println()
					}
					return nil
				}
			}
			// If no in-progress task has its dependencies met, this indicates an inconsistency
			fmt.Println("Warning: In-progress tasks exist but none have all dependencies met - possible data inconsistency")
		}

		// For pending tasks, find one that has all its dependencies met and no active subtasks
		for _, task := range pendingTasks {
			if isTaskReady(task, taskMap) && !hasActiveSubtasks(task, allTasks) {
				fmt.Printf("Next actionable task:\n\n")
				fmt.Printf("* %s (ID: %s)\n", task.Title, task.ID)
				if task.Description != "" {
					fmt.Printf("  %s\n", task.Description)
				}
				fmt.Printf("  State: %s | Complexity: %d\n", task.State, task.Complexity)
				if task.Depth > 0 {
					fmt.Printf("  Depth: %d", task.Depth)
					if task.ParentID != nil {
						fmt.Printf(" | Parent: %s", *task.ParentID)
					}
					fmt.Println()
				}
				return nil
			}
		}

		// If we reach here, check for specific scenarios
		if len(pendingTasks) > 0 {
			fmt.Println("No actionable tasks found: All pending tasks have unmet dependencies (possible deadlock scenario)")
		} else {
			fmt.Println("No actionable tasks found: No pending or in-progress tasks available")
		}

		return nil
	}
}

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
				if isTaskReady(task, taskMap) {
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
			return outputTasksAsJSON(readyTasks)
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
				if !isTaskReady(task, taskMap) && len(task.Dependencies) > 0 {
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