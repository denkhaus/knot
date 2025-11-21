package task

import (
	"context"
	"fmt"
	"strings"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/shared"

	"github.com/denkhaus/knot/v2/internal/errors"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// DeletionCommands returns task deletion related CLI commands
func DeletionCommands(appCtx *shared.AppContext) []*cli.Command {
	return []*cli.Command{
		{
			Name:   "delete",
			Usage:  "Delete a task with two-step confirmation. Use --all to delete task and all descendants",
			Action: deleteAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Usage:    "Task ID to delete",
					Required: true,
				},
				&cli.BoolFlag{
					Name:  "dry-run",
					Usage: "Show what would be deleted without actually deleting",
					Value: false,
				},
				&cli.BoolFlag{
					Name:  "all",
					Usage: "Delete task and all descendants recursively",
					Value: false,
				},
			},
		},
	}
}

// deleteAction handles task deletion with two-step confirmation
func deleteAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		taskIDStr := c.String("id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return errors.InvalidUUIDError("task-id", taskIDStr)
		}

		dryRun := c.Bool("dry-run")
		deleteAll := c.Bool("all")

		// Get task details
		task, err := appCtx.ProjectManager.GetTask(context.Background(), taskID)
		if err != nil {
			return errors.TaskNotFoundError(taskID)
		}

		// Check if task has children
		children, err := appCtx.ProjectManager.GetChildTasks(context.Background(), taskID)
		if err != nil {
			return errors.WrapWithSuggestion(err, "checking child tasks")
		}

		// Handle task with children based on --all flag
		if len(children) > 0 && !deleteAll {
			return &errors.EnhancedError{
				Operation:   "deleting task",
				Cause:       fmt.Errorf("task has %d child task(s)", len(children)),
				Suggestion:  "Either delete all child tasks first, or use the --all flag to delete the entire hierarchy",
				Example:     fmt.Sprintf("knot task delete --id %s --all", taskID),
				HelpCommand: "knot task children --task-id " + taskID.String(),
			}
		}

		// If --all flag is used, get all descendants for subtree deletion
		var descendants []*types.Task
		if deleteAll {
			descendants, err = getTaskDescendants(appCtx.ProjectManager, taskID)
			if err != nil {
				return errors.WrapWithSuggestion(err, "getting task descendants")
			}
		}

		// Two-step deletion process
		if task.State == types.TaskStateDeletionPending {
			// Second call - actually delete the task or subtree
			if dryRun {
				if deleteAll {
					totalTasks := 1 + len(descendants)
					fmt.Printf("DRY RUN: Task subtree would be permanently deleted (%d tasks, no actual changes made)\n", totalTasks)
				} else {
					fmt.Printf("DRY RUN: Task would be permanently deleted (no actual changes made)\n")
				}
				return nil
			}

			// Show what will be deleted
			if deleteAll {
				fmt.Printf("Final deletion of task subtree:\n")
				fmt.Printf("  %s (ID: %s) [ROOT]\n", task.Title, task.ID)

				if len(descendants) > 0 {
					fmt.Printf("  └── %d descendant task(s):\n", len(descendants))
					for _, desc := range descendants {
						indent := strings.Repeat("  ", desc.Depth-task.Depth+1)
						fmt.Printf("  %s├─ %s (ID: %s)\n", indent, desc.Title, desc.ID)
					}
				}

				totalTasks := 1 + len(descendants)
				fmt.Printf("\nTotal tasks to delete: %d\n", totalTasks)

				// Perform subtree deletion
				err = appCtx.ProjectManager.DeleteTaskSubtree(context.Background(), taskID, appCtx.Actor)
				if err != nil {
					appCtx.Logger.Error("Failed to delete task subtree", zap.Error(err))
					return errors.WrapWithSuggestion(err, "deleting task subtree")
				}

				appCtx.Logger.Info("Task subtree deleted successfully", zap.Int("totalDeleted", totalTasks))
				fmt.Printf("Task subtree permanently deleted: %d task(s) removed\n", totalTasks)
			} else {
				fmt.Printf("Final deletion of task:\n")
				fmt.Printf("  • %s (ID: %s)\n", task.Title, task.ID)
				if task.Description != "" {
					fmt.Printf("    %s\n", task.Description)
				}

				// Perform single task deletion
				err = appCtx.ProjectManager.DeleteTask(context.Background(), taskID, appCtx.Actor)
				if err != nil {
					return &errors.EnhancedError{
						Operation:   "deleting task",
						Cause:       err,
						Suggestion:  "Check if the task still exists or if there are constraint violations",
						HelpCommand: "knot task get --help",
					}
				}

				fmt.Printf("Task permanently deleted: %s\n", task.Title)
			}
			return nil
		} else {
			// First call - mark for deletion
			if dryRun {
				if deleteAll {
					totalTasks := 1 + len(descendants)
					fmt.Printf("DRY RUN: Task subtree would be marked for deletion (%d tasks, no actual changes made)\n", totalTasks)
				} else {
					fmt.Printf("DRY RUN: Task would be marked for deletion (no actual changes made)\n")
				}
				return nil
			}

			// Show what will be marked for deletion
			if deleteAll {
				fmt.Printf("Task subtree to be marked for deletion:\n")
				fmt.Printf("  %s (ID: %s) [ROOT]\n", task.Title, task.ID)
				if task.Description != "" {
					fmt.Printf("    %s\n", task.Description)
				}
				fmt.Printf("    Current State: %s | Complexity: %d\n", task.State, task.Complexity)

				if len(descendants) > 0 {
					fmt.Printf("  └── %d descendant task(s):\n", len(descendants))
					for _, desc := range descendants {
						indent := strings.Repeat("  ", desc.Depth-task.Depth+1)
						fmt.Printf("  %s├─ %s (ID: %s) - State: %s\n", indent, desc.Title, desc.ID, desc.State)
					}
				}

				totalTasks := 1 + len(descendants)
				fmt.Printf("\nTotal tasks to mark for deletion: %d\n", totalTasks)

				// Check for dependencies on any task in the subtree
				err = checkSubtreeDependencies(appCtx, task, descendants)
				if err != nil {
					return err
				}

				fmt.Printf("\nTask subtree marked for deletion. To confirm deletion, run the same command again:\n")
				fmt.Printf("    knot task delete --id %s --all\n", taskID)
			} else {
				fmt.Printf("Task to be marked for deletion:\n")
				fmt.Printf("  • %s (ID: %s)\n", task.Title, task.ID)
				if task.Description != "" {
					fmt.Printf("    %s\n", task.Description)
				}
				fmt.Printf("    Current State: %s | Complexity: %d\n", task.State, task.Complexity)

				// Check for dependencies
				dependencies, err := appCtx.ProjectManager.GetTaskDependencies(context.Background(), taskID)
				if err == nil && len(dependencies) > 0 {
					fmt.Printf("\n  This task depends on %d other task(s):\n", len(dependencies))
					for _, dep := range dependencies {
						fmt.Printf("    • %s (ID: %s)\n", dep.Title, dep.ID)
					}
				}

				dependents, err := appCtx.ProjectManager.GetDependentTasks(context.Background(), taskID)
				if err == nil && len(dependents) > 0 {
					fmt.Printf("\n  %d task(s) depend on this task:\n", len(dependents))
					for _, dep := range dependents {
						fmt.Printf("    • %s (ID: %s)\n", dep.Title, dep.ID)
					}
					fmt.Printf("    These dependencies will be removed.\n")
				}

				fmt.Printf("\nTask marked for deletion. To confirm deletion, run the same command again:\n")
				fmt.Printf("    knot task delete --id %s\n", taskID)
			}

			fmt.Printf("\nTo cancel deletion, change the task state:\n")
			fmt.Printf("    knot task update-state --id %s --state pending\n", taskID)

			if deleteAll {
				fmt.Printf("\nNote: Only the root task is marked as deletion-pending. All descendants will be deleted when confirmed.\n")
			}

			// Mark root task for deletion (triggers subtree deletion if --all was used)
			_, err = appCtx.ProjectManager.UpdateTask(context.Background(), task.ID, task.Title, task.Description, task.Complexity, types.TaskStateDeletionPending, appCtx.Actor)
			if err != nil {
				return &errors.EnhancedError{
					Operation:   "marking task for deletion",
					Cause:       err,
					Suggestion:  "Check if the task state transition is valid",
					HelpCommand: "knot task update-state --help",
				}
			}

			return nil
		}
	}
}


// confirmDeletion prompts user for confirmation
// Currently unused but kept for potential future use
// func confirmDeletion(itemType, itemName string) bool {
// 	fmt.Printf("\nAre you sure you want to delete this %s?\n", itemType)
// 	fmt.Printf("   %s\n", itemName)
// 	fmt.Printf("\nThis action cannot be undone. Type 'yes' to confirm: ")
//
// 	var response string
// 	_, _ = fmt.Scanln(&response)
//
// 	return strings.ToLower(strings.TrimSpace(response)) == "yes"
// }

// getTaskDescendants recursively gets all descendants of a task (renamed to avoid conflict)
func getTaskDescendants(projectManager manager.ProjectManager, taskID uuid.UUID) ([]*types.Task, error) {
	var result []*types.Task
	visited := make(map[uuid.UUID]bool)

	var collectDescendants func(uuid.UUID) error
	collectDescendants = func(id uuid.UUID) error {
		if visited[id] {
			return nil
		}
		visited[id] = true

		children, err := projectManager.GetChildTasks(context.Background(), id)
		if err != nil {
			return err
		}

		for _, child := range children {
			result = append(result, child)
			if err := collectDescendants(child.ID); err != nil {
				return err
			}
		}

		return nil
	}

	if err := collectDescendants(taskID); err != nil {
		return nil, err
	}

	return result, nil
}

// checkSubtreeDependencies checks for external dependencies on tasks in the subtree
func checkSubtreeDependencies(appCtx *shared.AppContext, rootTask *types.Task, descendants []*types.Task) error {
	allTasks := append([]*types.Task{rootTask}, descendants...)

	// Check dependencies for root task
	dependencies, err := appCtx.ProjectManager.GetTaskDependencies(context.Background(), rootTask.ID)
	if err == nil && len(dependencies) > 0 {
		fmt.Printf("\n  Root task depends on %d other task(s):\n", len(dependencies))
		for _, dep := range dependencies {
			fmt.Printf("    • %s (ID: %s)\n", dep.Title, dep.ID)
		}
	}

	// Check for external dependents (tasks outside the subtree that depend on tasks in the subtree)
	var externalDependents []*types.Task
	subtreeTaskIDs := make(map[uuid.UUID]bool)
	for _, task := range allTasks {
		subtreeTaskIDs[task.ID] = true
	}

	for _, task := range allTasks {
		dependents, err := appCtx.ProjectManager.GetDependentTasks(context.Background(), task.ID)
		if err != nil {
			continue
		}

		for _, dependent := range dependents {
			// If the dependent is not in our subtree, it's an external dependency
			if !subtreeTaskIDs[dependent.ID] {
				externalDependents = append(externalDependents, dependent)
			}
		}
	}

	if len(externalDependents) > 0 {
		fmt.Printf("\n  %d external task(s) depend on tasks in this subtree:\n", len(externalDependents))
		for _, dep := range externalDependents {
			fmt.Printf("    • %s (ID: %s)\n", dep.Title, dep.ID)
		}
		fmt.Printf("    These dependencies will be removed when the subtree is deleted.\n")
	}

	return nil
}
