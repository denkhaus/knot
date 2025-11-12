package task

import (
	"context"
	"fmt"
	"sort"

	"github.com/denkhaus/knot/internal/shared"

	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// HierarchyCommands returns hierarchy navigation CLI commands
func HierarchyCommands(appCtx *shared.AppContext) []*cli.Command {
	return []*cli.Command{
		{
			Name:   "children",
			Usage:  "Get direct children of a task",
			Action: ChildrenAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "task-id",
					Usage:    "Parent task ID",
					Required: true,
				},
				&cli.BoolFlag{
					Name:  "recursive",
					Usage: "Show all descendants (children of children)",
					Value: false,
				},
			},
		},
		{
			Name:   "parent",
			Usage:  "Get parent task of a task",
			Action: ParentAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "task-id",
					Usage:    "Child task ID",
					Required: true,
				},
			},
		},
		{
			Name:   "roots",
			Usage:  "Get root tasks of a project",
			Action: RootsAction(appCtx),
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:  "limit",
					Usage: "Maximum number of root tasks to show",
					Value: 0, // 0 means no limit
				},
			},
		},
		{
			Name:   "tree",
			Usage:  "Show task hierarchy as a tree",
			Action: TreeAction(appCtx),
			Flags: []cli.Flag{
				shared.NewJSONFlag(),
				&cli.IntFlag{
					Name:  "max-depth",
					Usage: "Maximum depth to show (0 = no limit)",
					Value: 0,
				},
				&cli.StringFlag{
					Name:  "root-task-id",
					Usage: "Show tree starting from specific task",
				},
			},
		},
	}
}

// ChildrenAction gets direct children of a task
func ChildrenAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		taskIDStr := c.String("task-id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return fmt.Errorf("invalid task ID: %w", err)
		}

		recursive := c.Bool("recursive")

		appCtx.Logger.Info("Getting child tasks",
			zap.String("taskID", taskID.String()),
			zap.Bool("recursive", recursive))

		// Get the parent task for context
		parentTask, err := appCtx.ProjectManager.GetTask(context.Background(), taskID)
		if err != nil {
			appCtx.Logger.Error("Failed to get parent task", zap.Error(err))
			return fmt.Errorf("failed to get parent task: %w", err)
		}

		var children []*types.Task
		if recursive {
			children, err = getAllDescendants(appCtx.ProjectManager, taskID)
		} else {
			children, err = appCtx.ProjectManager.GetChildTasks(context.Background(), taskID)
		}

		if err != nil {
			appCtx.Logger.Error("Failed to get child tasks", zap.Error(err))
			return fmt.Errorf("failed to get child tasks: %w", err)
		}

		fmt.Printf("Children of '%s' (ID: %s):\n\n", parentTask.Title, taskID)

		if len(children) == 0 {
			fmt.Println("No child tasks found.")
			return nil
		}

		// Sort by depth first, then by title
		sort.Slice(children, func(i, j int) bool {
			if children[i].Depth != children[j].Depth {
				return children[i].Depth < children[j].Depth
			}
			return children[i].Title < children[j].Title
		})

		for i, child := range children {
			indent := ""
			if recursive {
				// Show indentation based on relative depth
				relativeDepth := child.Depth - parentTask.Depth - 1
				for d := 0; d < relativeDepth; d++ {
					indent += "  "
				}
			}

			fmt.Printf("%s%d. %s (ID: %s)\n", indent, i+1, child.Title, child.ID)
			if child.Description != "" {
				fmt.Printf("%s   %s\n", indent, child.Description)
			}
			fmt.Printf("%s   State: %s | Complexity: %d | Depth: %d\n",
				indent, child.State, child.Complexity, child.Depth)
			fmt.Println()
		}

		if recursive {
			fmt.Printf("Total: %d descendants\n", len(children))
		} else {
			fmt.Printf("Total: %d direct children\n", len(children))
		}

		return nil
	}
}

// ParentAction gets parent task of a task
func ParentAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		taskIDStr := c.String("task-id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return fmt.Errorf("invalid task ID: %w", err)
		}

		appCtx.Logger.Info("Getting parent task", zap.String("taskID", taskID.String()))

		// Get the child task first
		childTask, err := appCtx.ProjectManager.GetTask(context.Background(), taskID)
		if err != nil {
			appCtx.Logger.Error("Failed to get task", zap.Error(err))
			return fmt.Errorf("failed to get task: %w", err)
		}

		fmt.Printf("Parent of '%s' (ID: %s):\n\n", childTask.Title, taskID)

		if childTask.ParentID == nil {
			fmt.Println("This is a root task (no parent).")
			return nil
		}

		parentTask, err := appCtx.ProjectManager.GetParentTask(context.Background(), taskID)
		if err != nil {
			appCtx.Logger.Error("Failed to get parent task", zap.Error(err))
			return fmt.Errorf("failed to get parent task: %w", err)
		}

		if parentTask == nil {
			fmt.Println("Parent task not found (orphaned task).")
			return nil
		}

		fmt.Printf("* %s (ID: %s)\n", parentTask.Title, parentTask.ID)
		if parentTask.Description != "" {
			fmt.Printf("  %s\n", parentTask.Description)
		}
		fmt.Printf("  State: %s | Complexity: %d | Depth: %d\n",
			parentTask.State, parentTask.Complexity, parentTask.Depth)

		return nil
	}
}

// RootsAction gets root tasks of a project
func RootsAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		projectID, err := shared.ResolveProjectID(c, appCtx)
		if err != nil {
			return err
		}

		limit := c.Int("limit")

		appCtx.Logger.Info("Getting root tasks",
			zap.String("projectID", projectID.String()),
			zap.Int("limit", limit))

		rootTasks, err := appCtx.ProjectManager.GetRootTasks(context.Background(), projectID)
		if err != nil {
			appCtx.Logger.Error("Failed to get root tasks", zap.Error(err))
			return fmt.Errorf("failed to get root tasks: %w", err)
		}

		fmt.Printf("Root tasks for project %s:\n\n", projectID)

		if len(rootTasks) == 0 {
			fmt.Println("No root tasks found.")
			return nil
		}

		// Sort by title for consistent output
		sort.Slice(rootTasks, func(i, j int) bool {
			return rootTasks[i].Title < rootTasks[j].Title
		})

		// Apply limit if specified
		if limit > 0 && len(rootTasks) > limit {
			fmt.Printf("Root tasks (showing %d of %d):\n\n", limit, len(rootTasks))
			rootTasks = rootTasks[:limit]
		} else {
			fmt.Printf("Root tasks (%d total):\n\n", len(rootTasks))
		}

		for i, task := range rootTasks {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, task.Title, task.ID)
			if task.Description != "" {
				fmt.Printf("   %s\n", task.Description)
			}
			fmt.Printf("   State: %s | Complexity: %d\n", task.State, task.Complexity)
			fmt.Println()
		}

		return nil
	}
} // getAllDescendants and other missing functions
