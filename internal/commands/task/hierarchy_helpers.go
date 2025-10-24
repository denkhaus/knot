package task

import (
	"context"
	"fmt"
	"sort"

	"github.com/denkhaus/knot/internal/manager"
	"github.com/denkhaus/knot/internal/shared"
	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

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

			task, err := appCtx.ProjectManager.GetTask(context.Background(), rootTaskID)
			if err != nil {
				return fmt.Errorf("failed to get root task: %w", err)
			}
			startingTasks = []*types.Task{task}
			fmt.Printf("Task tree starting from '%s':\n\n", task.Title)
		} else {
			// Start from project roots
			roots, err := appCtx.ProjectManager.GetRootTasks(context.Background(), projectID)
			if err != nil {
				return fmt.Errorf("failed to get root tasks: %w", err)
			}
			startingTasks = roots
			fmt.Printf("Task tree for project %s:\n\n", projectID)
		}

		if len(startingTasks) == 0 {
			fmt.Println("No tasks found.")
			return nil
		}

		// Sort starting tasks
		sort.Slice(startingTasks, func(i, j int) bool {
			return startingTasks[i].Title < startingTasks[j].Title
		})

		for _, task := range startingTasks {
			if err := printTaskTree(appCtx.ProjectManager, task, 0, maxDepth, ""); err != nil {
				return fmt.Errorf("failed to print task tree: %w", err)
			}
		}

		return nil
	}
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

// printTaskTree recursively prints a task and its children as a tree
func printTaskTree(projectManager manager.ProjectManager, task *types.Task, currentDepth, maxDepth int, prefix string) error {
	// Check depth limit
	if maxDepth > 0 && currentDepth >= maxDepth {
		return nil
	}

	// Print current task
	fmt.Printf("%s+- %s (ID: %s) - %s\n", prefix, task.Title, task.ID, task.State)

	// Get children
	children, err := projectManager.GetChildTasks(context.Background(), task.ID)
	if err != nil {
		return err
	}

	// Sort children
	sort.Slice(children, func(i, j int) bool {
		return children[i].Title < children[j].Title
	})

	// Print children
	for i, child := range children {
		childPrefix := prefix
		if i == len(children)-1 {
			childPrefix += "   "
		} else {
			childPrefix += "|  "
		}

		if err := printTaskTree(projectManager, child, currentDepth+1, maxDepth, childPrefix); err != nil {
			return err
		}
	}

	return nil
}
