package task

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/denkhaus/knot/v2/internal/utils"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// DuplicateAction creates a copy of a task
func DuplicateAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		container := shared.GetContainerFromContext(c)
		projectManager := container.GetProjectManager()
		loggerService := container.GetLogger()

		taskIDStr := c.String("task-id")
		if taskIDStr == "" {
			return fmt.Errorf("task-id is required")
		}

		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return fmt.Errorf("invalid task ID: %w", err)
		}

		projectIDStr := c.String("target-project-id")
		if projectIDStr == "" {
			return fmt.Errorf("target-project-id is required")
		}

		targetProjectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			return fmt.Errorf("invalid target project ID: %w", err)
		}

		loggerService.Info("Duplicating task",
			zap.String("taskID", taskID.String()),
			zap.String("targetProjectID", targetProjectID.String()))

		duplicatedTask, err := projectManager.DuplicateTask(context.Background(), taskID, targetProjectID)
		if err != nil {
			loggerService.Error("Failed to duplicate task", zap.Error(err))
			return fmt.Errorf("failed to duplicate task: %w", err)
		}

		fmt.Printf("Task duplicated successfully:\n")
		fmt.Printf("  Original: %s\n", taskID)
		fmt.Printf("  New: %s (ID: %s)\n", duplicatedTask.Title, duplicatedTask.ID)
		fmt.Printf("  Target Project: %s\n", targetProjectID)
		fmt.Printf("  State: %s (reset to pending)\n", duplicatedTask.State)
		fmt.Printf("  Complexity: %d\n", duplicatedTask.Complexity)

		return nil
	}
}

// ListByStateAction lists tasks filtered by state
func ListByStateAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		container := shared.GetContainerFromContext(c)
		projectManager := container.GetProjectManager()
		loggerService := container.GetLogger()

		projectID, err := projectManager.ResolveProjectID(c.Context)
		if err != nil {
			return err
		}

		stateStr := c.String("state")
		if stateStr == "" {
			return fmt.Errorf("state is required")
		}

		state := types.TaskState(stateStr)

		loggerService.Info("Listing tasks by state",
			zap.String("projectID", projectID.String()),
			zap.String("state", stateStr))

		tasks, err := projectManager.ListTasksByState(context.Background(), projectID, state)
		if err != nil {
			loggerService.Error("Failed to list tasks by state", zap.Error(err))
			return fmt.Errorf("failed to list tasks by state: %w", err)
		}

		if len(tasks) == 0 {
			fmt.Printf("No tasks found with state '%s' in project %s\n", state, projectID)
			return nil
		}

		// Check if JSON output is requested
		if c.Bool("json") {
			return utils.OutputTasksAsJSON(tasks)
		}

		fmt.Printf("Tasks with state '%s' (%d found):\n\n", state, len(tasks))
		for i, task := range tasks {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, task.Title, task.ID)
			if task.Description != "" {
				fmt.Printf("   %s\n", task.Description)
			}
			fmt.Printf("   Complexity: %d | Depth: %d\n", task.Complexity, task.Depth)
			if task.ParentID != nil {
				fmt.Printf("   Parent: %s\n", *task.ParentID)
			}
			fmt.Println()
		}

		return nil
	}
}

// BulkCreateAction creates multiple tasks from JSON input
func BulkCreateAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		container := shared.GetContainerFromContext(c)
		projectManager := container.GetProjectManager()
		loggerService := container.GetLogger()

		projectID, err := projectManager.ResolveProjectID(c.Context)
		if err != nil {
			return err
		}

		inputFile := c.String("file")
		if inputFile == "" {
			return fmt.Errorf("file is required")
		}

		// Read input file
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		// Parse JSON input
		var taskInputs []struct {
			Title       string  `json:"title"`
			Description string  `json:"description"`
			Complexity  int     `json:"complexity"`
			ParentID    *string `json:"parent_id,omitempty"`
		}

		if err := json.Unmarshal(data, &taskInputs); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}

		if len(taskInputs) == 0 {
			return fmt.Errorf("no tasks found in input file")
		}

		actor := shared.GetActorFromContext(c)

		loggerService.Info("Bulk creating tasks",
			zap.Int("taskCount", len(taskInputs)),
			zap.String("projectID", projectID.String()),
			zap.String("actor", actor))

		var createdTasks []types.Task
		for i, input := range taskInputs {
			// Validate input
			if input.Title == "" {
				return fmt.Errorf("task %d: title is required", i+1)
			}
			if input.Complexity <= 0 {
				input.Complexity = 5 // Default complexity
			}

			var parentID *uuid.UUID
			if input.ParentID != nil && *input.ParentID != "" {
				parsed, err := uuid.Parse(*input.ParentID)
				if err != nil {
					return fmt.Errorf("task %d: invalid parent ID '%s': %w", i+1, *input.ParentID, err)
				}
				parentID = &parsed
			}

			// Create task
			task, err := projectManager.CreateTask(
				context.Background(),
				projectID,
				parentID,
				input.Title,
				input.Description,
				input.Complexity,
				types.TaskPriorityMedium,
				actor,
			)
			if err != nil {
				loggerService.Error("Failed to create task", zap.Error(err), zap.Int("taskIndex", i))
				return fmt.Errorf("failed to create task %d ('%s'): %w", i+1, input.Title, err)
			}

			createdTasks = append(createdTasks, *task)
		}

		fmt.Printf("Successfully created %d tasks:\n", len(createdTasks))
		for i, task := range createdTasks {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, task.Title, task.ID)
			if task.Description != "" {
				fmt.Printf("   %s\n", task.Description)
			}
			fmt.Printf("   Complexity: %d | State: %s\n", task.Complexity, task.State)
		}
		fmt.Printf("  Created by: %s\n", actor)

		return nil
	}
}

// BulkDeleteAction deletes multiple tasks with safety checks
func BulkDeleteAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		container := shared.GetContainerFromContext(c)
		projectManager := container.GetProjectManager()
		loggerService := container.GetLogger()

		taskIDsStr := c.String("task-ids")
		if taskIDsStr == "" {
			return fmt.Errorf("task-ids are required")
		}

		// Parse comma-separated task IDs
		taskIDStrings := strings.Split(taskIDsStr, ",")
		var taskIDs []uuid.UUID
		for _, idStr := range taskIDStrings {
			idStr = strings.TrimSpace(idStr)
			taskID, err := uuid.Parse(idStr)
			if err != nil {
				return fmt.Errorf("invalid task ID '%s': %w", idStr, err)
			}
			taskIDs = append(taskIDs, taskID)
		}

		if len(taskIDs) == 0 {
			return fmt.Errorf("no valid task IDs provided")
		}

		dryRun := c.Bool("dry-run")
		force := c.Bool("force")

		loggerService.Info("Bulk deleting tasks",
			zap.Int("taskCount", len(taskIDs)),
			zap.Bool("dryRun", dryRun),
			zap.Bool("force", force))

		// Get task details for confirmation using optimized batch loading
		tasksToDelete, err := projectManager.GetTasksWithDependencies(context.Background(), taskIDs)
		if err != nil {
			loggerService.Error("Failed to get tasks", zap.Error(err), zap.Strings("taskIDs", utils.ConvertUUIDsToStrings(taskIDs)))
			return fmt.Errorf("failed to get tasks: %w", err)
		}

		// Show what will be deleted
		fmt.Printf("Tasks to be deleted (%d):\n", len(tasksToDelete))
		for i, task := range tasksToDelete {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, task.Title, task.ID)
			if task.Description != "" {
				fmt.Printf("   %s\n", task.Description)
			}
			fmt.Printf("   State: %s | Complexity: %d\n", task.State, task.Complexity)
		}

		if dryRun {
			fmt.Println("\nDry run mode - no tasks were actually deleted.")
			return nil
		}

		// Confirmation prompt (unless force flag is used)
		if !force {
			fmt.Printf("\nAre you sure you want to delete these %d tasks? (y/N): ", len(tasksToDelete))
			var response string
			_, _ = fmt.Scanln(&response)
			if response != "y" && response != "Y" && response != "yes" && response != "YES" {
				fmt.Println("Deletion cancelled.")
				return nil
			}
		}

		// Get actor for deletion
		actor := shared.GetActorFromContext(c)

		// Delete tasks
		var deletedCount int
		for _, taskID := range taskIDs {
			err := projectManager.DeleteTask(context.Background(), taskID, actor)
			if err != nil {
				loggerService.Error("Failed to delete task", zap.Error(err), zap.String("taskID", taskID.String()))
				fmt.Printf("Failed to delete task %s: %v\n", taskID, err)
				continue
			}
			deletedCount++
		}

		fmt.Printf("Successfully deleted %d out of %d tasks\n", deletedCount, len(taskIDs))
		if deletedCount < len(taskIDs) {
			fmt.Printf("Warning: %d tasks could not be deleted (see errors above)\n", len(taskIDs)-deletedCount)
		}

		return nil
	}
}

// BulkCommands returns bulk operation CLI commands
func BulkCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:   "bulk-create",
			Usage:  "Create multiple tasks from JSON file",
			Action: BulkCreateAction(),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "file",
					Aliases:  []string{"f"},
					Usage:    "JSON file containing task definitions",
					Required: true,
				},
			},
		},
		{
			Name:   "bulk-delete",
			Usage:  "Delete multiple tasks with safety checks",
			Action: BulkDeleteAction(),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "task-ids",
					Usage:    "Comma-separated list of task IDs",
					Required: true,
				},
				&cli.BoolFlag{
					Name:  "dry-run",
					Usage: "Show what would be deleted without actually deleting",
				},
				&cli.BoolFlag{
					Name:  "force",
					Usage: "Skip confirmation prompt",
				},
			},
		},
		{
			Name:   "duplicate",
			Usage:  "Duplicate a task to another project",
			Action: DuplicateAction(),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "task-id",
					Usage:    "Task ID to duplicate",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "target-project-id",
					Usage:    "Target project ID",
					Required: true,
				},
			},
		},
		{
			Name:   "list-by-state",
			Usage:  "List tasks filtered by state",
			Action: ListByStateAction(),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "state",
					Aliases:  []string{"s"},
					Usage:    "Task state (pending, in-progress, completed, blocked, cancelled)",
					Required: true,
				},
				&cli.BoolFlag{
					Name:    "json",
					Aliases: []string{"j"},
					Usage:   "Output in JSON format",
				},
			},
		},
	}
}
