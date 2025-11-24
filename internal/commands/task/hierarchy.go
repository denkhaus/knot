package task

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/treeformatter"

	"github.com/denkhaus/knot/v2/internal/types"
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
				shared.NewQuietFlag(),
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

// wrapText wraps text at the specified width and returns a slice of lines
func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	for len(text) > width {
		// Find the last space before width to avoid breaking words
		breakPoint := width
		for i := width; i > 0; i-- {
			if text[i] == ' ' {
				breakPoint = i
				break
			}
		}

		if breakPoint == 0 {
			// No space found, break at width
			breakPoint = width
		}

		lines = append(lines, text[:breakPoint])
		text = strings.TrimSpace(text[breakPoint:])
	}

	if len(text) > 0 {
		lines = append(lines, text)
	}

	return lines
}

// printChildrenUsingTreeFormat displays child tasks using the unified tree format
// Task Reference: bf1b0a75-3f12-49dd-9391-26d2e632fc2b | Brain Reference: e4bea247-7f1f-4712-8188-b9b0b4ecb3ea
func printChildrenUsingTreeFormat(children []*types.Task, parentTask *types.Task, recursive bool) {
	// Create tree formatter with emoji support for visual hierarchy
	formatter := treeformatter.NewFormatter(&treeformatter.Config{
		ShowEmojis:   true,
		CompactMode:  false,
		IndentSize:   2,
	})

	// Display children using proper tree structure
	for i, child := range children {
		isLast := i == len(children)-1
		var prefix string

		if recursive {
			// Calculate relative depth for proper indentation
			relativeDepth := child.Depth - parentTask.Depth - 1
			if relativeDepth > 0 {
				// For nested children, build the prefix with proper ancestors
				// We need to show the tree structure up to this depth
				prefix = ""
				for d := 0; d < relativeDepth; d++ {
					if d == relativeDepth-1 {
						// This is the level where our child is
						if isLast {
							prefix += "└── "
						} else {
							prefix += "├── "
						}
					} else {
						// For levels above, assume they have more children (use │)
						prefix += "│  "
					}
				}
			} else {
				// Direct children of parent
				if isLast {
					prefix = "└── "
				} else {
					prefix = "├── "
				}
			}
		} else {
			// Non-recursive mode: all are direct children at same level
			if isLast {
				prefix = "└── "
			} else {
				prefix = "├── "
			}
		}

		// Print task line using unified format
		taskLine := formatter.FormatTaskLine(child)
		fmt.Printf("%s%s\n", prefix, taskLine)

		// Add description if available with proper continuation and text wrapping
		if child.Description != "" {
			// Build description prefix that maintains tree continuity
			var descPrefix string
			if isLast {
				// Last task: └── becomes spaces for proper alignment
				descPrefix = strings.Replace(prefix, "└── ", "   ", -1)
			} else {
				// Non-last tasks: ├── becomes │ for continuity
				descPrefix = strings.Replace(prefix, "├── ", "│  ", -1)
			}

			// Wrap description at 120 characters
			wrappedLines := wrapText(child.Description, 120)
			for _, line := range wrappedLines {
				fmt.Printf("%s%s\n", descPrefix, line)
			}
		}

		// Add separator line with box drawing continuity (except for last task)
		if !isLast {
			// Build separator prefix that maintains vertical line
			separatorPrefix := strings.Replace(prefix, "├── ", "│", -1)
			separatorPrefix = strings.Replace(separatorPrefix, "└── ", "│", -1)
			fmt.Printf("%s\n", separatorPrefix)
		}
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

		if len(children) == 0 {
			fmt.Printf("No child tasks found for '%s' (ID: %s).\n", parentTask.Title, taskID)
			return nil
		}

		fmt.Printf("\n┌── Children of '%s' (ID: %s):\n│\n", parentTask.Title, taskID)

		// Sort by depth first, then by title
		sort.Slice(children, func(i, j int) bool {
			if children[i].Depth != children[j].Depth {
				return children[i].Depth < children[j].Depth
			}
			return children[i].Title < children[j].Title
		})

		// Display children using unified tree format with emojis and box drawing
		printChildrenUsingTreeFormat(children, parentTask, recursive)

		// Add empty line before total count
		fmt.Println()

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
}
