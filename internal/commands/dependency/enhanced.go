package dependency

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

// EnhancedCommands returns enhanced dependency-related CLI commands
func EnhancedCommands(appCtx *shared.AppContext) []*cli.Command {
	return []*cli.Command{
		{
			Name:   "dependents",
			Usage:  "List tasks that depend on this task",
			Action: dependentsAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "task-id",
					Usage:    "Task ID",
					Required: true,
				},
				&cli.BoolFlag{
					Name:  "recursive",
					Usage: "Show all transitive dependents (tasks that depend on dependents)",
					Value: false,
				},
			},
		},
		{
			Name:   "chain",
			Usage:  "Show dependency chain for a task",
			Action: chainAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "task-id",
					Usage:    "Task ID",
					Required: true,
				},
				&cli.BoolFlag{
					Name:  "upstream",
					Usage: "Show upstream dependencies (what this task depends on)",
					Value: true,
				},
				&cli.BoolFlag{
					Name:  "downstream",
					Usage: "Show downstream dependencies (what depends on this task)",
					Value: false,
				},
			},
		},
		{
			Name:   "cycles",
			Usage:  "Detect circular dependencies in project",
			Action: cyclesAction(appCtx),
		},
		{
			Name:   "validate",
			Usage:  "Validate all dependencies in project",
			Action: validateAction(appCtx),
		},
	}
}

func dependentsAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		taskIDStr := c.String("task-id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return fmt.Errorf("invalid task ID: %w", err)
		}

		recursive := c.Bool("recursive")

		appCtx.Logger.Info("Getting dependent tasks",
			zap.String("taskID", taskID.String()),
			zap.Bool("recursive", recursive))

		var dependents []*types.Task
		if recursive {
			dependents, err = getAllTransitiveDependents(appCtx.ProjectManager, taskID)
		} else {
			dependents, err = appCtx.ProjectManager.GetDependentTasks(context.Background(), taskID)
		}

		if err != nil {
			appCtx.Logger.Error("Failed to get dependents", zap.Error(err))
			return fmt.Errorf("failed to get dependents: %w", err)
		}

		// Get the original task for context
		task, err := appCtx.ProjectManager.GetTask(context.Background(), taskID)
		if err != nil {
			appCtx.Logger.Error("Failed to get task", zap.Error(err))
			return fmt.Errorf("failed to get task: %w", err)
		}

		fmt.Printf("Tasks that depend on '%s' (ID: %s):\n\n", task.Title, taskID)

		if len(dependents) == 0 {
			fmt.Println("No tasks depend on this task.")
			return nil
		}

		// Sort by state and title for consistent output
		sort.Slice(dependents, func(i, j int) bool {
			if dependents[i].State != dependents[j].State {
				return dependents[i].State < dependents[j].State
			}
			return dependents[i].Title < dependents[j].Title
		})

		for i, dep := range dependents {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, dep.Title, dep.ID)
			if dep.Description != "" {
				fmt.Printf("   %s\n", dep.Description)
			}
			fmt.Printf("   State: %s | Complexity: %d\n", dep.State, dep.Complexity)
			if dep.Depth > 0 {
				fmt.Printf("   Depth: %d", dep.Depth)
				if dep.ParentID != nil {
					fmt.Printf(" | Parent: %s", *dep.ParentID)
				}
				fmt.Println()
			}
			fmt.Println()
		}

		if recursive {
			fmt.Printf("Total: %d transitive dependents\n", len(dependents))
		} else {
			fmt.Printf("Total: %d direct dependents\n", len(dependents))
		}

		return nil
	}
}

// getAllTransitiveDependents recursively gets all tasks that depend on the given task
func getAllTransitiveDependents(projectManager manager.ProjectManager, taskID uuid.UUID) ([]*types.Task, error) {
	visited := make(map[uuid.UUID]bool)
	var result []*types.Task

	var collectDependents func(uuid.UUID) error
	collectDependents = func(id uuid.UUID) error {
		if visited[id] {
			return nil
		}
		visited[id] = true

		dependents, err := projectManager.GetDependentTasks(context.Background(), id)
		if err != nil {
			return err
		}

		for _, dep := range dependents {
			if !visited[dep.ID] {
				result = append(result, dep)
				if err := collectDependents(dep.ID); err != nil {
					return err
				}
			}
		}

		return nil
	}

	if err := collectDependents(taskID); err != nil {
		return nil, err
	}

	return result, nil
}
