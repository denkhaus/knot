package task

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/treeformatter"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// TreeNode represents a task node in JSON tree format
type TreeNode struct {
	*types.Task
	Children []*TreeNode `json:"children,omitempty"`
}

// TreeAction shows task hierarchy as a tree
func TreeAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		projectID, err := shared.ResolveProjectID(c, appCtx)
		if err != nil {
			return err
		}

		maxDepth := c.Int("max-depth")
		rootTaskIDStr := c.String("root-task-id")

		appCtx.Logger.Info("Showing task tree",
			zap.String("projectID", projectID.String()),
			zap.Int("maxDepth", maxDepth),
			zap.String("rootTaskID", rootTaskIDStr))

		var startingTasks []*types.Task

		if rootTaskIDStr != "" {
			// Start from specific task
			rootTaskID, err := uuid.Parse(rootTaskIDStr)
			if err != nil {
				return fmt.Errorf("invalid root task ID: %w", err)
			}

			// Use batch loading for consistency with other optimizations
			tasks, err := appCtx.ProjectManager.GetTasksWithDependencies(context.Background(), []uuid.UUID{rootTaskID})
			if err != nil {
				return fmt.Errorf("failed to get root task: %w", err)
			}
			if len(tasks) == 0 {
				return fmt.Errorf("root task not found")
			}
			startingTasks = tasks
			fmt.Printf("Task tree starting from '%s':\n\n", tasks[0].Title)
		} else {
			// Start from project roots
			roots, err := appCtx.ProjectManager.GetRootTasks(context.Background(), projectID)
			if err != nil {
				return fmt.Errorf("failed to get root tasks: %w", err)
			}
			startingTasks = roots
		}

		if len(startingTasks) == 0 {
			fmt.Println("No tasks found.")
			return nil
		}

		// Sort starting tasks
		sort.Slice(startingTasks, func(i, j int) bool {
			return startingTasks[i].Title < startingTasks[j].Title
		})

		// Show headers for non-JSON mode (skip if quiet)
		if !c.Bool("json") && !c.Bool("quiet") {
			if rootTaskIDStr != "" {
				fmt.Printf("Task tree starting from '%s':\n\n", startingTasks[0].Title)
			} else {
				fmt.Printf("Task tree for project %s:\n\n", projectID)
			}
		}

		// Output JSON if requested
		if c.Bool("json") {
			var treeNodes []*TreeNode
			for _, task := range startingTasks {
				treeNode, err := buildTreeJSON(appCtx.ProjectManager, task, 0, maxDepth)
				if err != nil {
					return fmt.Errorf("failed to build JSON tree: %w", err)
				}
				treeNodes = append(treeNodes, treeNode)
			}

			jsonData, err := json.MarshalIndent(treeNodes, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal tree to JSON: %w", err)
			}
			fmt.Println(string(jsonData))
			return nil
		}

		for _, task := range startingTasks {
			if err := printTaskTree(appCtx.ProjectManager, task, 0, maxDepth, ""); err != nil {
				return fmt.Errorf("failed to print task tree: %w", err)
			}
		}

		return nil
	}
}

// buildTreeJSON recursively builds a JSON tree structure
func buildTreeJSON(projectManager manager.ProjectManager, task *types.Task, currentDepth, maxDepth int) (*TreeNode, error) {
	// Check depth limit
	if maxDepth > 0 && currentDepth >= maxDepth {
		return &TreeNode{Task: task, Children: []*TreeNode{}}, nil
	}

	node := &TreeNode{
		Task:     task,
		Children: []*TreeNode{},
	}

	// Get children
	children, err := projectManager.GetChildTasks(context.Background(), task.ID)
	if err != nil {
		return nil, err
	}

	// Sort children
	sort.Slice(children, func(i, j int) bool {
		return children[i].Title < children[j].Title
	})

	// Build child nodes
	for _, child := range children {
		childNode, err := buildTreeJSON(projectManager, child, currentDepth+1, maxDepth)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, childNode)
	}

	return node, nil
}

// getAllDescendants recursively gets all descendants of a task
func getAllDescendants(projectManager manager.ProjectManager, taskID uuid.UUID) ([]*types.Task, error) {
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

// printTaskTree recursively prints a task and its children as a tree using unified formatting
// Task Reference: 4eaf63a8-5a5c-4572-ad63-0c039f2943ac | Brain Reference: e4bea247-7f1f-4712-8188-b9b0b4ecb3ea
func printTaskTree(projectManager manager.ProjectManager, task *types.Task, currentDepth, maxDepth int, prefix string) error {
	// Check depth limit
	if maxDepth > 0 && currentDepth >= maxDepth {
		return nil
	}

	// Create tree formatter with emoji support for visual hierarchy
	formatter := treeformatter.NewFormatter(&treeformatter.Config{
		ShowEmojis:  true,
		CompactMode: false,
		IndentSize:  2,
	})

	// Print current task with proper tree prefix
	taskLine := formatter.FormatTaskLine(task)
	if currentDepth == 0 {
		// Root task - no prefix
		fmt.Printf("┌─ %s\n", taskLine)
	} else {
		fmt.Printf("%s%s\n", prefix, taskLine)
	}

	// Get children (maintaining existing logic)
	children, err := projectManager.GetChildTasks(context.Background(), task.ID)
	if err != nil {
		return err
	}

	// Sort children (maintaining existing logic)
	sort.Slice(children, func(i, j int) bool {
		return children[i].Title < children[j].Title
	})

	// If root task has children, add structural line
	if currentDepth == 0 && len(children) > 0 {
		fmt.Printf("│\n")
	}

	// Print children recursively with proper tree structure
	for i, child := range children {
		isLast := i == len(children)-1

		var childPrefix string
		var structuralPrefix string

		if currentDepth == 0 {
			// First level children
			if isLast {
				childPrefix = "└─── "
				structuralPrefix = "    "
			} else {
				childPrefix = "├─── "
				structuralPrefix = "│   "
			}
		} else {
			// Deeper levels
			if isLast {
				childPrefix = prefix + "└── "
				structuralPrefix = prefix + "    "
			} else {
				childPrefix = prefix + "├── "
				structuralPrefix = prefix + "│   "
			}
		}

		// Check if this child has its own children
		grandchildren, err := projectManager.GetChildTasks(context.Background(), child.ID)
		if err == nil && len(grandchildren) > 0 {
			// Parent task with children - use special formatting
			fmt.Printf("%s┬─ %s\n", strings.TrimSuffix(childPrefix, " "), formatter.FormatTaskLine(child))
			// Add structural line after parent with children
			fmt.Printf("%s│\n", structuralPrefix)

			// Print grandchildren with enhanced structure
			if err := printTaskTreeWithStructure(projectManager, child, currentDepth+1, maxDepth, structuralPrefix); err != nil {
				return err
			}
		} else {
			// Leaf task
			if err := printTaskTreeWithPrefix(projectManager, child, currentDepth+1, maxDepth, childPrefix, structuralPrefix); err != nil {
				return err
			}
		}

		// Add structural line between siblings (except after last one)
		if !isLast && currentDepth == 0 {
			fmt.Printf("│\n")
		}
	}

	return nil
}

// printTaskTreeWithPrefix is a helper that prints tasks with explicit tree prefixes
func printTaskTreeWithPrefix(projectManager manager.ProjectManager, task *types.Task, currentDepth, maxDepth int, treePrefix, continuationPrefix string) error {
	// Check depth limit
	if maxDepth > 0 && currentDepth >= maxDepth {
		return nil
	}

	// Create tree formatter
	formatter := treeformatter.NewFormatter(&treeformatter.Config{
		ShowEmojis:  true,
		CompactMode: false,
		IndentSize:  2,
	})

	// Print current task with its tree prefix
	taskLine := formatter.FormatTaskLine(task)
	fmt.Printf("%s%s\n", treePrefix, taskLine)

	// Get children
	children, err := projectManager.GetChildTasks(context.Background(), task.ID)
	if err != nil {
		return err
	}

	// Sort children
	sort.Slice(children, func(i, j int) bool {
		return children[i].Title < children[j].Title
	})

	// Print children recursively
	for i, child := range children {
		isLast := i == len(children)-1

		var childTreePrefix string
		var childContinuationPrefix string

		if isLast {
			childTreePrefix = continuationPrefix + "└── "
			childContinuationPrefix = continuationPrefix + "    "
		} else {
			childTreePrefix = continuationPrefix + "├── "
			childContinuationPrefix = continuationPrefix + "│   "
		}

		if err := printTaskTreeWithPrefix(projectManager, child, currentDepth+1, maxDepth, childTreePrefix, childContinuationPrefix); err != nil {
			return err
		}
	}

	return nil
}

// printTaskTreeWithStructure prints a task tree with enhanced structural lines
func printTaskTreeWithStructure(projectManager manager.ProjectManager, parentTask *types.Task, currentDepth, maxDepth int, prefix string) error {
	// Create tree formatter
	formatter := treeformatter.NewFormatter(&treeformatter.Config{
		ShowEmojis:  true,
		CompactMode: false,
		IndentSize:  2,
	})

	// Get children of the parent task
	children, err := projectManager.GetChildTasks(context.Background(), parentTask.ID)
	if err != nil {
		return err
	}

	// Sort children
	sort.Slice(children, func(i, j int) bool {
		return children[i].Title < children[j].Title
	})

	// Print children with enhanced structure
	for i, child := range children {
		isLast := i == len(children)-1

		var childPrefix string
		var structuralPrefix string

		if isLast {
			childPrefix = prefix + "└── "
			structuralPrefix = prefix + "    "
		} else {
			childPrefix = prefix + "├── "
			structuralPrefix = prefix + "│   "
		}

		// Check if this child has its own children
		grandchildren, err := projectManager.GetChildTasks(context.Background(), child.ID)
		if err == nil && len(grandchildren) > 0 {
			// Parent task with children - use special formatting
			fmt.Printf("%s┬─ %s\n", strings.TrimSuffix(childPrefix, " "), formatter.FormatTaskLine(child))
			// Add structural line after parent with children
			fmt.Printf("%s│\n", structuralPrefix)

			// Recursively print grandchildren
			if err := printTaskTreeWithStructure(projectManager, child, currentDepth+1, maxDepth, structuralPrefix); err != nil {
				return err
			}
		} else {
			// Leaf task
			taskLine := formatter.FormatTaskLine(child)
			fmt.Printf("%s%s\n", childPrefix, taskLine)
		}

		// Add structural line between siblings (except after last one at this level)
		if !isLast {
			fmt.Printf("%s│\n", prefix)
		}
	}

	return nil
}
