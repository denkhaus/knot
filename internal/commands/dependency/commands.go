package dependency

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/internal/errors"
	"github.com/denkhaus/knot/internal/shared"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// Commands returns all dependency-related CLI commands
func Commands(appCtx *shared.AppContext) []*cli.Command {
	// Basic commands
	basicCommands := []*cli.Command{
		{
			Name:   "add",
			Usage:  "Add task dependency",
			Action: addAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "task-id",
					Usage:    "Task ID",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "depends-on",
					Usage:    "Task ID that this task depends on",
					Required: true,
				},
			},
		},
		{
			Name:   "remove",
			Usage:  "Remove task dependency",
			Action: removeAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "task-id",
					Usage:    "Task ID",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "depends-on",
					Usage:    "Task ID to remove dependency from",
					Required: true,
				},
			},
		},
		{
			Name:   "list",
			Usage:  "List task dependencies",
			Action: listAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "task-id",
					Usage:    "Task ID",
					Required: true,
				},
			},
		},
	}

	// Enhanced commands
	enhancedCommands := EnhancedCommands(appCtx)

	// Combine all commands
	allCommands := make([]*cli.Command, 0, len(basicCommands)+len(enhancedCommands))
	allCommands = append(allCommands, basicCommands...)
	allCommands = append(allCommands, enhancedCommands...)

	return allCommands
}

func addAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		taskIDStr := c.String("task-id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return errors.InvalidUUIDError("task-id", taskIDStr)
		}

		dependsOnStr := c.String("depends-on")
		dependsOnID, err := uuid.Parse(dependsOnStr)
		if err != nil {
			return errors.InvalidUUIDError("depends-on", dependsOnStr)
		}

		actor := c.String("actor")

		appCtx.Logger.Info("Adding task dependency",
			zap.String("taskID", taskID.String()),
			zap.String("dependsOnID", dependsOnID.String()),
			zap.String("actor", actor))

		_, err = appCtx.ProjectManager.AddTaskDependency(context.Background(), taskID, dependsOnID, actor)
		if err != nil {
			appCtx.Logger.Error("Failed to add dependency", zap.Error(err))
			return errors.WrapWithSuggestion(err, "adding task dependency")
		}

		appCtx.Logger.Info("Dependency added successfully", zap.String("actor", actor))
		fmt.Printf("Added dependency: %s now depends on %s\n", taskID, dependsOnID)
		fmt.Printf("  Added by: %s\n", actor)
		return nil
	}
}

func removeAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		taskIDStr := c.String("task-id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return fmt.Errorf("invalid task ID: %w", err)
		}

		dependsOnStr := c.String("depends-on")
		dependsOnID, err := uuid.Parse(dependsOnStr)
		if err != nil {
			return fmt.Errorf("invalid depends-on ID: %w", err)
		}

		actor := c.String("actor")

		appCtx.Logger.Info("Removing task dependency",
			zap.String("taskID", taskID.String()),
			zap.String("dependsOnID", dependsOnID.String()),
			zap.String("actor", actor))

		_, err = appCtx.ProjectManager.RemoveTaskDependency(context.Background(), taskID, dependsOnID, actor)
		if err != nil {
			appCtx.Logger.Error("Failed to remove dependency", zap.Error(err))
			return fmt.Errorf("failed to remove dependency: %w", err)
		}

		appCtx.Logger.Info("Dependency removed successfully", zap.String("actor", actor))
		fmt.Printf("Removed dependency: %s no longer depends on %s\n", taskID, dependsOnID)
		fmt.Printf("  Removed by: %s\n", actor)
		return nil
	}
}

func listAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		taskIDStr := c.String("task-id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return fmt.Errorf("invalid task ID: %w", err)
		}

		appCtx.Logger.Info("Listing task dependencies", zap.String("taskID", taskID.String()))

		dependencies, err := appCtx.ProjectManager.GetTaskDependencies(context.Background(), taskID)
		if err != nil {
			appCtx.Logger.Error("Failed to get dependencies", zap.Error(err))
			return fmt.Errorf("failed to get dependencies: %w", err)
		}

		dependents, err := appCtx.ProjectManager.GetDependentTasks(context.Background(), taskID)
		if err != nil {
			appCtx.Logger.Error("Failed to get dependents", zap.Error(err))
			return fmt.Errorf("failed to get dependents: %w", err)
		}

		appCtx.Logger.Info("Dependencies retrieved",
			zap.Int("dependencies", len(dependencies)),
			zap.Int("dependents", len(dependents)))

		fmt.Printf("Dependencies for task %s:\n\n", taskID)

		if len(dependencies) > 0 {
			fmt.Println("This task depends on:")
			for _, dep := range dependencies {
				fmt.Printf("  • %s (ID: %s) - %s\n", dep.Title, dep.ID, dep.State)
			}
		} else {
			fmt.Println("This task has no dependencies.")
		}

		fmt.Println()

		if len(dependents) > 0 {
			fmt.Println("Tasks that depend on this task:")
			for _, dep := range dependents {
				fmt.Printf("  • %s (ID: %s) - %s\n", dep.Title, dep.ID, dep.State)
			}
		} else {
			fmt.Println("No tasks depend on this task.")
		}

		return nil
	}
}
