// Package dependency provides CLI commands for dependency management in KNOT.
//
// This package implements command-line interface commands for managing
// task dependencies, relationship visualization, and dependency analysis.
//
// Key Commands:
//   - AddDependency: Creates dependencies between tasks
//   - RemoveDependency: Removes task dependencies
//   - ListDependencies: Shows dependency relationships
//   - VisualizeDependencies: Displays dependency graphs
//
// Cross-reference: Knot Task 86f3ba2d-3a87-493b-b8fc-96d19f344e89
package dependency

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/errors"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// Commands returns all dependency-related CLI commands
func Commands() []*cli.Command {
	// Basic commands
	basicCommands := []*cli.Command{
		{
			Name:   "add",
			Usage:  "Add task dependency",
			Action: addAction(),
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
			Action: removeAction(),
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
			Action: listAction(),
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
	enhancedCommands := EnhancedCommands()

	// Combine all commands
	allCommands := make([]*cli.Command, 0, len(basicCommands)+len(enhancedCommands))
	allCommands = append(allCommands, basicCommands...)
	allCommands = append(allCommands, enhancedCommands...)

	return allCommands
}

func addAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		container := shared.GetContainerFromContext(c)
		projectManager := container.GetProjectManager()
		loggerService := container.GetLogger()
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

		loggerService.Info("Adding task dependency",
			zap.String("taskID", taskID.String()),
			zap.String("dependsOnID", dependsOnID.String()),
			zap.String("actor", actor))

		_, err = projectManager.AddTaskDependency(context.Background(), taskID, dependsOnID, actor)
		if err != nil {
			loggerService.Error("Failed to add dependency", zap.Error(err))
			return errors.WrapWithSuggestion(err, "adding task dependency")
		}

		loggerService.Info("Dependency added successfully", zap.String("actor", actor))
		fmt.Printf("Added dependency: %s now depends on %s\n", taskID, dependsOnID)
		fmt.Printf("  Added by: %s\n", actor)
		return nil
	}
}

func removeAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		container := shared.GetContainerFromContext(c)
		projectManager := container.GetProjectManager()
		loggerService := container.GetLogger()

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

		loggerService.Info("Removing task dependency",
			zap.String("taskID", taskID.String()),
			zap.String("dependsOnID", dependsOnID.String()),
			zap.String("actor", actor))

		_, err = projectManager.RemoveTaskDependency(context.Background(), taskID, dependsOnID, actor)
		if err != nil {
			loggerService.Error("Failed to remove dependency", zap.Error(err))
			return fmt.Errorf("failed to remove dependency: %w", err)
		}

		loggerService.Info("Dependency removed successfully", zap.String("actor", actor))
		fmt.Printf("Removed dependency: %s no longer depends on %s\n", taskID, dependsOnID)
		fmt.Printf("  Removed by: %s\n", actor)
		return nil
	}
}

func listAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		container := shared.GetContainerFromContext(c)
		projectManager := container.GetProjectManager()
		loggerService := container.GetLogger()

		taskIDStr := c.String("task-id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return fmt.Errorf("invalid task ID: %w", err)
		}

		loggerService.Info("Listing task dependencies", zap.String("taskID", taskID.String()))

		dependencies, err := projectManager.GetTaskDependencies(context.Background(), taskID)
		if err != nil {
			loggerService.Error("Failed to get dependencies", zap.Error(err))
			return fmt.Errorf("failed to get dependencies: %w", err)
		}

		dependents, err := projectManager.GetDependentTasks(context.Background(), taskID)
		if err != nil {
			loggerService.Error("Failed to get dependents", zap.Error(err))
			return fmt.Errorf("failed to get dependents: %w", err)
		}

		loggerService.Info("Dependencies retrieved",
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
