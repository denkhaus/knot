package task

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/denkhaus/knot/internal/shared"

	"github.com/denkhaus/knot/internal/errors"
	"github.com/denkhaus/knot/internal/types"
	"github.com/denkhaus/knot/internal/validation"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// validateProjectID validates and returns the project ID from the CLI context

// Commands returns all task-related CLI commands
func Commands(appCtx *shared.AppContext) []*cli.Command {
	// Basic task commands
	basicCommands := []*cli.Command{
		{
			Name:   "create",
			Usage:  "Create a new task",
			Action: createAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "title",
					Aliases:  []string{"t"},
					Usage:    "Task title",
					Required: true,
				},
				&cli.StringFlag{
					Name:    "description",
					Aliases: []string{"d"},
					Usage:   "Task description",
				},
				&cli.StringFlag{
					Name:  "parent-id",
					Usage: "Parent task ID (for subtasks)",
				},
				&cli.IntFlag{
					Name:    "complexity",
					Aliases: []string{"c"},
					Usage:   "Task complexity (1-10)",
					Value:   5,
					EnvVars: []string{"KNOT_DEFAULT_COMPLEXITY"},
				},
				&cli.StringFlag{
					Name:    "priority",
					Aliases: []string{"p"},
					Usage:   "Task priority (low, medium, high)",
					Value:   "medium",
				},
			},
		},
		{
			Name:   "list",
			Usage:  "List tasks with advanced filtering options",
			Action: listAction(appCtx),
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "json",
					Aliases: []string{"j"},
					Usage:   "Output in JSON format",
				},
				&cli.StringFlag{
					Name:    "state",
					Aliases: []string{"s"},
					Usage:   "Filter by task state (pending, in-progress, completed, blocked, cancelled)",
				},
				&cli.StringFlag{
					Name:    "priority",
					Aliases: []string{"p"},
					Usage:   "Filter by task priority (low, medium, high)",
				},
				&cli.IntFlag{
					Name:  "complexity-min",
					Usage: "Filter by minimum complexity (1-10)",
				},
				&cli.IntFlag{
					Name:  "complexity-max",
					Usage: "Filter by maximum complexity (1-10)",
				},
				&cli.IntFlag{
					Name:  "complexity",
					Usage: "Filter by exact complexity (1-10)",
				},
				&cli.StringFlag{
					Name:    "search",
					Aliases: []string{"q"},
					Usage:   "Search in task titles and descriptions",
				},
				&cli.IntFlag{
					Name:  "depth-max",
					Usage: "Filter by maximum depth in hierarchy",
				},
				&cli.IntFlag{
					Name:    "limit",
					Aliases: []string{"l"},
					Usage:   "Maximum number of tasks to show",
				},
				&cli.StringFlag{
					Name:  "sort",
					Usage: "Sort by field (title, complexity, state, priority, created, depth)",
					Value: "created",
				},
				&cli.BoolFlag{
					Name:  "reverse",
					Usage: "Reverse sort order",
				},
			},
		},
		{
			Name:   "update-state",
			Usage:  "Update task state",
			Action: updateStateAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Usage:    "Task ID",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "state",
					Aliases:  []string{"s"},
					Usage:    "New state (pending, in-progress, completed, blocked, cancelled)",
					Required: true,
				},
			},
		},
		{
			Name:   "update-title",
			Usage:  "Update task title",
			Action: updateTitleAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Usage:    "Task ID",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "title",
					Aliases:  []string{"t"},
					Usage:    "New task title",
					Required: true,
				},
			},
		},
		{
			Name:   "update-description",
			Usage:  "Update task description",
			Action: updateDescriptionAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Usage:    "Task ID",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "description",
					Aliases:  []string{"d"},
					Usage:    "New task description",
					Required: true,
				},
			},
		},
		{
			Name:   "update-priority",
			Usage:  "Update task priority",
			Action: updatePriorityAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Usage:    "Task ID",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "priority",
					Aliases:  []string{"p"},
					Usage:    "New task priority (low, medium, high)",
					Required: true,
				},
			},
		},
	}

	// Hierarchy navigation commands
	hierarchyCommands := HierarchyCommands(appCtx)

	// Task deletion commands
	deletionCommands := DeletionCommands(appCtx)

	// Bulk operation commands
	bulkCommands := BulkCommands(appCtx)

	// Combine all commands
	allCommands := make([]*cli.Command, 0, len(basicCommands)+len(hierarchyCommands)+len(deletionCommands)+len(bulkCommands))
	allCommands = append(allCommands, basicCommands...)
	allCommands = append(allCommands, hierarchyCommands...)
	allCommands = append(allCommands, deletionCommands...)
	allCommands = append(allCommands, bulkCommands...)

	return allCommands
}

func createAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		projectID, err := shared.ResolveProjectID(c, appCtx)
		if err != nil {
			return err
		}

		title := c.String("title")
		description := c.String("description")
		complexity := c.Int("complexity")
		priority := c.String("priority")
		actor := c.String("actor")

		// Create input validator
		validator := validation.NewInputValidator()

		// Validate inputs
		if err := validator.ValidateTaskTitle(title); err != nil {
			return errors.NewValidationError("invalid task title", err)
		}

		if err := validator.ValidateTaskDescription(description); err != nil {
			return errors.NewValidationError("invalid task description", err)
		}

		if err := validator.ValidateComplexity(complexity); err != nil {
			return errors.NewValidationError("invalid complexity", err)
		}

		// Validate priority
		if err := validator.ValidateTaskPriority(priority); err != nil {
			return errors.NewValidationError("invalid priority", err)
		}

		// Default to $USER if actor is not provided
		if actor == "" {
			actor = os.Getenv("USER")
			if actor == "" {
				actor = "unknown"
			}
		}

		// Validate complexity
		if err := errors.ValidateComplexity(complexity); err != nil {
			return err
		}

		var parentID *uuid.UUID
		if parentIDStr := c.String("parent-id"); parentIDStr != "" {
			parsed, err := uuid.Parse(parentIDStr)
			if err != nil {
				return errors.InvalidUUIDError("parent-id", parentIDStr)
			}
			parentID = &parsed
		}

		appCtx.Logger.Info("Creating task",
			zap.String("title", title),
			zap.String("projectID", projectID.String()),
			zap.Int("complexity", complexity),
			zap.String("priority", priority),
			zap.String("actor", actor))

		task, err := appCtx.ProjectManager.CreateTask(context.Background(), projectID, parentID, title, description, complexity, types.TaskPriority(priority), actor)
		if err != nil {
			appCtx.Logger.Error("Failed to create task", zap.Error(err))
			return errors.WrapWithSuggestion(err, "creating task")
		}

		appCtx.Logger.Info("Task created successfully", zap.String("taskID", task.ID.String()), zap.String("actor", actor))

		fmt.Printf("Created task: %s (ID: %s)\n", task.Title, task.ID)
		fmt.Printf("  Created by: %s\n", actor)
		if task.Description != "" {
			fmt.Printf("  Description: %s\n", task.Description)
		}
		fmt.Printf("  Complexity: %d\n", task.Complexity)
		fmt.Printf("  Priority: %s\n", task.Priority)
		fmt.Printf("  State: %s\n", task.State)
		if parentID != nil {
			fmt.Printf("  Parent: %s\n", *parentID)
		}

		// Show workflow reminder for task state management
		fmt.Printf("\nReminder: Set this task to 'in-progress' before starting work:\n")
		fmt.Printf("  knot task update-state --id %s --state in-progress\n", task.ID)

		// Show breakdown suggestion for high complexity tasks
		if complexity >= 8 {
			fmt.Printf("\nNote: This task has high complexity (%d >= 8 threshold).\n", complexity)
			fmt.Printf("Consider breaking it down into smaller subtasks:\n")
			fmt.Printf("  knot task create --parent-id %s --title \"Subtask 1\"\n", task.ID)
			fmt.Printf("  knot breakdown  # to see all tasks needing breakdown\n")
		}

		return nil
	}
}

func listAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		projectID, err := shared.ResolveProjectID(c, appCtx)
		if err != nil {
			return err
		}

		appCtx.Logger.Info("Listing tasks", zap.String("projectID", projectID.String()))

		tasks, err := appCtx.ProjectManager.ListTasksForProject(context.Background(), projectID)
		if err != nil {
			appCtx.Logger.Error("Failed to list tasks", zap.Error(err))
			return errors.WrapWithSuggestion(err, "listing tasks")
		}

		// Apply filters
		filteredTasks := applyTaskFilters(tasks, c)

		// Apply sorting
		sortedTasks := applyTaskSorting(filteredTasks, c)

		// Apply limit
		finalTasks := applyTaskLimit(sortedTasks, c)

		appCtx.Logger.Info("Tasks filtered and sorted",
			zap.Int("originalCount", len(tasks)),
			zap.Int("filteredCount", len(finalTasks)))

		if len(finalTasks) == 0 {
			fmt.Printf("No tasks found matching the specified criteria.\n")
			return nil
		}

		// Check if JSON output is requested
		if c.Bool("json") {
			return outputTasksAsJSON(finalTasks)
		}

		// Show project context indicator
		shared.ShowProjectContextWithSeparator(c, appCtx)

		// Show filter summary if filters were applied
		if hasFiltersApplied(c) {
			fmt.Printf("Found %d task(s) matching criteria (out of %d total):\n\n", len(finalTasks), len(tasks))
		} else {
			fmt.Printf("Found %d task(s):\n\n", len(finalTasks))
		}

		for _, task := range finalTasks {
			indent := ""
			for i := 0; i < task.Depth; i++ {
				indent += "  "
			}
			fmt.Printf("%s* %s (ID: %s)\n", indent, task.Title, task.ID)
			if task.Description != "" {
				fmt.Printf("%s  %s\n", indent, task.Description)
			}
			fmt.Printf("%s  State: %s | Priority: %s | Complexity: %d | Depth: %d\n", indent, task.State, task.Priority, task.Complexity, task.Depth)
			fmt.Println()
		}
		return nil
	}
}

func updateStateAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		taskIDStr := c.String("id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return errors.InvalidUUIDError("task-id", taskIDStr)
		}

		stateStr := c.String("state")
		actor := c.String("actor")

		// Default to $USER if actor is not provided
		if actor == "" {
			actor = os.Getenv("USER")
			if actor == "" {
				actor = "unknown"
			}
		}

		// Basic state validation
		if err := errors.ValidateTaskState(stateStr); err != nil {
			return err
		}

		newState := types.TaskState(stateStr)

		appCtx.Logger.Info("Updating task state",
			zap.String("taskID", taskID.String()),
			zap.String("newState", stateStr),
			zap.String("actor", actor))

		// Get current task to preserve other fields
		task, err := appCtx.ProjectManager.GetTask(context.Background(), taskID)
		if err != nil {
			appCtx.Logger.Error("Failed to get task", zap.Error(err))
			return errors.TaskNotFoundError(taskID)
		}

		// Validate state transition
		validator := validation.NewStateValidator()
		if err := validator.ValidateTransition(task.State, newState, task); err != nil {
			// EnhancedError already contains user-friendly formatting
			// No need to log this as it's a user input validation error
			return err
		}

		// Update task state
		updatedTask, err := appCtx.ProjectManager.UpdateTaskState(context.Background(), taskID, newState, actor)
		if err != nil {
			appCtx.Logger.Error("Failed to update task state", zap.Error(err))
			return errors.WrapWithSuggestion(err, "updating task state")
		}

		appCtx.Logger.Info("Task state updated successfully", zap.String("actor", actor))
		fmt.Printf("Updated task state: %s -> %s\n", task.State, updatedTask.State)
		fmt.Printf("  Updated by: %s\n", actor)
		return nil
	}
}

// outputTasksAsJSON outputs tasks in JSON format
func outputTasksAsJSON(tasks []*types.Task) error {
	jsonData, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks to JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// outputSingleTaskAsJSON outputs a single task in JSON format
func outputSingleTaskAsJSON(task *types.Task) error {
	jsonData, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal task to JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

func updateTitleAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		taskIDStr := c.String("id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return errors.InvalidUUIDError("task-id", taskIDStr)
		}

		newTitle := c.String("title")
		actor := c.String("actor")
		if newTitle == "" {
			return fmt.Errorf("title cannot be empty")
		}

		// Default to $USER if actor is not provided
		if actor == "" {
			actor = os.Getenv("USER")
			if actor == "" {
				actor = "unknown"
			}
		}

		appCtx.Logger.Info("Updating task title",
			zap.String("taskID", taskID.String()),
			zap.String("newTitle", newTitle),
			zap.String("actor", actor))

		// Get current task to check if it exists and get old title
		task, err := appCtx.ProjectManager.GetTask(context.Background(), taskID)
		if err != nil {
			appCtx.Logger.Error("Failed to get task", zap.Error(err))
			return errors.TaskNotFoundError(taskID)
		}

		oldTitle := task.Title

		// Update task title
		updatedTask, err := appCtx.ProjectManager.UpdateTaskTitle(context.Background(), taskID, newTitle, actor)
		if err != nil {
			appCtx.Logger.Error("Failed to update task title", zap.Error(err))
			return errors.WrapWithSuggestion(err, "updating task title")
		}

		appCtx.Logger.Info("Task title updated successfully", zap.String("actor", actor))
		fmt.Printf("Updated task title: \"%s\" -> \"%s\"\n", oldTitle, updatedTask.Title)
		fmt.Printf("  Updated by: %s\n", actor)
		return nil
	}
}

func updateDescriptionAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		taskIDStr := c.String("id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return errors.InvalidUUIDError("task-id", taskIDStr)
		}

		newDescription := c.String("description")
		actor := c.String("actor")

		// Default to $USER if actor is not provided
		if actor == "" {
			actor = os.Getenv("USER")
			if actor == "" {
				actor = "unknown"
			}
		}

		appCtx.Logger.Info("Updating task description",
			zap.String("taskID", taskID.String()),
			zap.String("newDescription", newDescription),
			zap.String("actor", actor))

		// Get current task to check if it exists and get old description
		task, err := appCtx.ProjectManager.GetTask(context.Background(), taskID)
		if err != nil {
			appCtx.Logger.Error("Failed to get task", zap.Error(err))
			return errors.TaskNotFoundError(taskID)
		}

		oldDescription := task.Description

		// Update task description
		updatedTask, err := appCtx.ProjectManager.UpdateTaskDescription(context.Background(), taskID, newDescription, actor)
		if err != nil {
			appCtx.Logger.Error("Failed to update task description", zap.Error(err))
			return errors.WrapWithSuggestion(err, "updating task description")
		}

		appCtx.Logger.Info("Task description updated successfully", zap.String("actor", actor))
		if oldDescription == "" {
			fmt.Printf("Updated task description: (empty) -> \"%s\"\n", updatedTask.Description)
		} else {
			fmt.Printf("Updated task description: \"%s\" -> \"%s\"\n", oldDescription, updatedTask.Description)
		}
		fmt.Printf("  Updated by: %s\n", actor)
		return nil
	}
}

func updatePriorityAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		taskIDStr := c.String("id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return errors.InvalidUUIDError("task-id", taskIDStr)
		}

		priority := c.String("priority")
		actor := c.String("actor")

		// Default to $USER if actor is not provided
		if actor == "" {
			actor = os.Getenv("USER")
			if actor == "" {
				actor = "unknown"
			}
		}

		// Validate priority
		validator := validation.NewInputValidator()
		if err := validator.ValidateTaskPriority(priority); err != nil {
			return errors.NewValidationError("invalid priority", err)
		}

		appCtx.Logger.Info("Updating task priority",
			zap.String("taskID", taskID.String()),
			zap.String("newPriority", priority),
			zap.String("actor", actor))

		// Get current task to check if it exists and get old priority
		task, err := appCtx.ProjectManager.GetTask(context.Background(), taskID)
		if err != nil {
			appCtx.Logger.Error("Failed to get task", zap.Error(err))
			return errors.TaskNotFoundError(taskID)
		}

		oldPriority := task.Priority

		// Update task priority using the service method
		updatedTask, err := appCtx.ProjectManager.UpdateTaskPriority(context.Background(), taskID, types.TaskPriority(priority), actor)
		if err != nil {
			appCtx.Logger.Error("Failed to update task priority", zap.Error(err))
			return errors.WrapWithSuggestion(err, "updating task priority")
		}

		appCtx.Logger.Info("Task priority updated successfully", zap.String("actor", actor))
		fmt.Printf("Updated task priority: \"%s\" -> \"%s\"\n", oldPriority, updatedTask.Priority)
		fmt.Printf("  Updated by: %s\n", actor)
		return nil
	}
}

// Helper functions for task filtering, sorting, and limiting

// applyTaskFilters applies all specified filters to the task list
func applyTaskFilters(tasks []*types.Task, c *cli.Context) []*types.Task {
	var filtered []*types.Task

	for _, task := range tasks {
		// State filter
		if state := c.String("state"); state != "" {
			if string(task.State) != state {
				continue
			}
		}

		// Priority filter
		if priority := c.String("priority"); priority != "" {
			if string(task.Priority) != priority {
				continue
			}
		}

		// Complexity filters
		if complexity := c.Int("complexity"); complexity > 0 {
			if task.Complexity != complexity {
				continue
			}
		}
		if complexityMin := c.Int("complexity-min"); complexityMin > 0 {
			if task.Complexity < complexityMin {
				continue
			}
		}
		if complexityMax := c.Int("complexity-max"); complexityMax > 0 {
			if task.Complexity > complexityMax {
				continue
			}
		}

		// Depth filter
		if depthMax := c.Int("depth-max"); depthMax >= 0 {
			if task.Depth > depthMax {
				continue
			}
		}

		// Search filter (case-insensitive search in title and description)
		if search := c.String("search"); search != "" {
			searchLower := strings.ToLower(search)
			titleMatch := strings.Contains(strings.ToLower(task.Title), searchLower)
			descMatch := strings.Contains(strings.ToLower(task.Description), searchLower)
			if !titleMatch && !descMatch {
				continue
			}
		}

		// If we get here, the task passed all filters
		filtered = append(filtered, task)
	}

	return filtered
}

// applyTaskSorting sorts the task list based on the specified criteria
func applyTaskSorting(tasks []*types.Task, c *cli.Context) []*types.Task {
	sortField := c.String("sort")
	reverse := c.Bool("reverse")

	// Make a copy to avoid modifying the original slice
	sorted := make([]*types.Task, len(tasks))
	copy(sorted, tasks)

	sort.Slice(sorted, func(i, j int) bool {
		var less bool

		switch sortField {
		case "title":
			less = strings.ToLower(sorted[i].Title) < strings.ToLower(sorted[j].Title)
		case "complexity":
			less = sorted[i].Complexity < sorted[j].Complexity
		case "state":
			less = string(sorted[i].State) < string(sorted[j].State)
		case "priority":
			// Sort by priority: high -> medium -> low
			priorityOrder := map[string]int{"high": 0, "medium": 1, "low": 2}
			less = priorityOrder[string(sorted[i].Priority)] < priorityOrder[string(sorted[j].Priority)]
		case "depth":
			less = sorted[i].Depth < sorted[j].Depth
		case "created":
			fallthrough
		default:
			// Default sort by creation time (using ID as proxy since tasks are created sequentially)
			less = sorted[i].ID.String() < sorted[j].ID.String()
		}

		if reverse {
			return !less
		}
		return less
	})

	return sorted
}

// applyTaskLimit applies the limit to the task list
func applyTaskLimit(tasks []*types.Task, c *cli.Context) []*types.Task {
	limit := c.Int("limit")
	if limit <= 0 || limit >= len(tasks) {
		return tasks
	}
	return tasks[:limit]
}

// hasFiltersApplied checks if any filters were applied
func hasFiltersApplied(c *cli.Context) bool {
	return c.String("state") != "" ||
		c.String("priority") != "" ||
		c.Int("complexity") > 0 ||
		c.Int("complexity-min") > 0 ||
		c.Int("complexity-max") > 0 ||
		c.Int("depth-max") >= 0 ||
		c.String("search") != "" ||
		c.Int("limit") > 0 ||
		c.String("sort") != "created" ||
		c.Bool("reverse")
}
