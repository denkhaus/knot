package task

import (
	"context"
	"fmt"
	"strings"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/treeformatter"
	"github.com/denkhaus/knot/v2/internal/utils"

	"github.com/denkhaus/knot/v2/internal/errors"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// printDeletionTree displays tasks using the unified tree format for deletion operations
// Task Reference: 12a4b5a6-908d-411d-959a-462cb3d3d1c2 | Brain Reference: e4bea247-7f1f-4712-8188-b9b0b4ecb3ea
func printDeletionTree(header string, rootTask *types.Task, descendants []*types.Task, showState bool) {
	// Create tree formatter with emoji support
	formatter := treeformatter.NewFormatter(&treeformatter.Config{
		ShowEmojis:   true,
		CompactMode:  false,
		IndentSize:   2,
	})

	fmt.Printf("\n┌── %s\n│\n", header)

	// Display root task with proper formatting (no emoji, use ├── like other tree commands)
	rootLine := fmt.Sprintf("%s (ID: %s) - %s", rootTask.Title, rootTask.ID, rootTask.State)
	fmt.Printf("├── %s [ROOT]\n", rootLine)

	// Display descendants if any - organize them hierarchically with enhanced structure
	if len(descendants) > 0 {
		fmt.Printf("│\n")

		// Build hierarchical structure from flat descendant list
		hierarchy := buildTaskHierarchy(rootTask.ID, descendants)

		// Print hierarchical tree with enhanced structure (no descriptions)
		printDeletionHierarchicalTree(formatter, hierarchy)

		// Add empty line after children (matching desired format)
		fmt.Printf("\n")
	}
}

// TaskNode represents a node in the hierarchical tree structure
type TaskNode struct {
	Task     *types.Task
	Children []*TaskNode
}

// buildTaskHierarchy creates a hierarchical tree structure from a flat list of descendants
func buildTaskHierarchy(rootTaskID uuid.UUID, descendants []*types.Task) []*TaskNode {
	// Create task lookup map
	taskMap := make(map[uuid.UUID]*types.Task)
	for _, task := range descendants {
		taskMap[task.ID] = task
	}

	// Build child-to-parent mapping
	childToParent := make(map[uuid.UUID]uuid.UUID)
	for _, task := range descendants {
		// Find which descendant is the parent of this task, if any
		for _, potentialParent := range descendants {
			if task.ParentID != nil && *task.ParentID == potentialParent.ID {
				childToParent[task.ID] = potentialParent.ID
				break
			}
		}
		// If no descendant parent found, it's a direct child of root
		if _, exists := childToParent[task.ID]; !exists {
			if task.ParentID == nil || *task.ParentID == rootTaskID {
				childToParent[task.ID] = rootTaskID
			}
		}
	}

	// Build node map
	nodeMap := make(map[uuid.UUID]*TaskNode)
	for _, task := range descendants {
		nodeMap[task.ID] = &TaskNode{
			Task:     task,
			Children: []*TaskNode{},
		}
	}

	// Build hierarchy
	var roots []*TaskNode
	for _, task := range descendants {
		node := nodeMap[task.ID]

		parentID, hasParent := childToParent[task.ID]
		if hasParent && parentID != rootTaskID {
			if parentNode, exists := nodeMap[parentID]; exists {
				parentNode.Children = append(parentNode.Children, node)
			}
		} else {
			// This is a root-level child
			roots = append(roots, node)
		}
	}

	// Sort children recursively
	sortTaskNodes(roots)

	return roots
}

// sortTaskNodes recursively sorts task nodes and their children by title
func sortTaskNodes(nodes []*TaskNode) {
	for _, node := range nodes {
		sortTaskNodes(node.Children)
	}
	// Sort the slice itself
	for i := 0; i < len(nodes)-1; i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[i].Task.Title > nodes[j].Task.Title {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}

// printHierarchicalTask prints a task and its children in hierarchical format
func printHierarchicalTask(formatter treeformatter.TreeFormatter, node *TaskNode, prefix string, isLast bool, isRoot bool) {
	// Determine the tree prefix
	var treePrefix string
	if isRoot {
		treePrefix = ""
	} else {
		if isLast {
			treePrefix = prefix + "└── "
		} else {
			treePrefix = prefix + "├── "
		}
	}

	// Print current task
	taskLine := formatter.FormatTaskLine(node.Task)
	fmt.Printf("%s%s\n", treePrefix, taskLine)

	// Add description if available
	if node.Task.Description != "" {
		wrappedLines := utils.WrapText(node.Task.Description, 120)
		for _, line := range wrappedLines {
			if isRoot {
				fmt.Printf("   %s\n", line)
			} else {
				// Use proper continuation prefix for description lines
				if isLast {
					fmt.Printf("      %s\n", line)
				} else {
					fmt.Printf("%s│   %s\n", prefix, line)
				}
			}
		}
	}

	// Print children recursively
	for i, child := range node.Children {
		childIsLast := i == len(node.Children)-1
		var childPrefix string

		if isRoot {
			childPrefix = ""
		} else {
			if isLast {
				childPrefix = prefix + "    "
			} else {
				childPrefix = prefix + "│   "
			}
		}

		printHierarchicalTask(formatter, child, childPrefix, childIsLast, false)
	}
}

// printDeletionHierarchicalTree prints the deletion tree with enhanced structure (similar to regular task tree)
func printDeletionHierarchicalTree(formatter treeformatter.TreeFormatter, nodes []*TaskNode) {
	// Print first level children with proper tree structure
	for i, node := range nodes {
		isLast := i == len(nodes)-1

		var childPrefix string

		if isLast {
			childPrefix = "└─── "
		} else {
			childPrefix = "├─── "
		}

		// Check if this child has its own children
		if len(node.Children) > 0 {
			// Parent task with children - use special formatting
			fmt.Printf("%s┬─ %s\n", strings.TrimSuffix(childPrefix, " "), formatter.FormatTaskLine(node.Task))

			// Add structural line after parent with children
			fmt.Printf("│\n")

			// Print children with proper structure
			printDeletionNodeChildren(formatter, node, "│   ")
		} else {
			// Leaf task
			taskLine := formatter.FormatTaskLine(node.Task)
			fmt.Printf("%s%s\n", childPrefix, taskLine)
		}

		// Add structural line between children (except after last one)
		if !isLast {
			fmt.Printf("│\n")
		}
	}
}

// printDeletionNodeChildren recursively prints children of a node with enhanced structure
func printDeletionNodeChildren(formatter treeformatter.TreeFormatter, node *TaskNode, prefix string) {
	for i, child := range node.Children {
		isLast := i == len(node.Children)-1

		var childPrefix string
		var continuationPrefix string

		if isLast {
			childPrefix = prefix + "└── "
			continuationPrefix = prefix + "    "
		} else {
			childPrefix = prefix + "├── "
			continuationPrefix = prefix + "│   "
		}

		// Print current child (no descriptions to maintain clean tree)
		taskLine := formatter.FormatTaskLine(child.Task)
		fmt.Printf("%s%s\n", childPrefix, taskLine)

		// Recursively print grandchildren
		if len(child.Children) > 0 {
			printDeletionNodeChildren(formatter, child, continuationPrefix)
		}
	}
}

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
				printDeletionTree("Final deletion of task subtree:", task, descendants, false)

				totalTasks := 1 + len(descendants)
				fmt.Printf("Total tasks to delete: %d\n", totalTasks)

				// Perform subtree deletion
				err = appCtx.ProjectManager.DeleteTaskSubtree(context.Background(), taskID, appCtx.Actor)
				if err != nil {
					appCtx.Logger.Error("Failed to delete task subtree", zap.Error(err))
					return errors.WrapWithSuggestion(err, "deleting task subtree")
				}

				appCtx.Logger.Info("Task subtree deleted successfully", zap.Int("totalDeleted", totalTasks))
				fmt.Printf("Task subtree permanently deleted: %d task(s) removed\n", totalTasks)
			} else {
				printDeletionTree("Final deletion of task:", task, nil, false)

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
				printDeletionTree("Task subtree to be marked for deletion:", task, descendants, true)

				totalTasks := 1 + len(descendants)
				fmt.Printf("Total tasks to mark for deletion: %d\n", totalTasks)

				// Check for dependencies on any task in the subtree
				err = checkSubtreeDependencies(appCtx, task, descendants)
				if err != nil {
					return err
				}

				fmt.Printf("\nTask subtree marked for deletion. To confirm deletion, run the same command again:\n")
				fmt.Printf("    knot task delete --id %s --all\n", taskID)
			} else {
				printDeletionTree("Task to be marked for deletion:", task, nil, true)

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
