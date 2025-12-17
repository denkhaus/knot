// Package task provides CLI commands for task management operations in KNOT.
//
// This package implements command-line interface commands for creating, updating,
// managing, and analyzing tasks within the KNOT project management system.
//
// Key Commands:
//   - CreateTask: Creates new tasks with hierarchical relationships
//   - ListTasks: Displays tasks with filtering and sorting options
//   - UpdateTask: Modifies task state, priority, and properties
//   - DeleteTask: Removes tasks and manages dependencies
//   - TaskAnalysis: Provides task breakdown and complexity analysis
//
// Cross-reference: Knot Task 86f3ba2d-3a87-493b-b8fc-96d19f344e89
package task

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/denkhaus/knot/v2/internal/errors"
	"github.com/denkhaus/knot/v2/internal/flags"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/denkhaus/knot/v2/internal/utils"
	"github.com/denkhaus/knot/v2/internal/validation"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// resolveProjectIDWithDI resolves the project ID using dependency injection
func resolveProjectIDWithDI(c *cli.Context, injector do.Injector) (uuid.UUID, error) {
	// Get project manager from DI
	projectManager := do.MustInvoke[manager.ProjectManager](injector)

	// Get project from database stored context
	if contextProjectID, err := projectManager.GetSelectedProject(c.Context); err == nil && contextProjectID != nil {
		return *contextProjectID, nil
	}

	// No project available
	return uuid.Nil, errors.NoProjectContextError()
}

// showProjectContextWithDI displays the current project context using dependency injection
func showProjectContextWithDI(c *cli.Context, injector do.Injector) bool {
	// Skip context display for JSON output or quiet mode
	if c.Bool("json") || c.Bool("quiet") {
		return false
	}

	// Get project manager from DI
	projectManager := do.MustInvoke[manager.ProjectManager](injector)

	// Get selected project
	selectedProjectID, err := projectManager.GetSelectedProject(c.Context)
	if err != nil || selectedProjectID == nil {
		return false
	}

	// Get project details
	project, err := projectManager.GetProject(context.Background(), *selectedProjectID)
	if err != nil {
		return false
	}

	// Display context indicator
	fmt.Printf("[Project: %s]\n", project.Title)
	return true
}

// Commands returns all task-related CLI commands
func Commands(injector do.Injector) []*cli.Command {
	// Basic task commands
	basicCommands := []*cli.Command{
		{
			Name:   "create",
			Usage:  "Create a new task",
			Action: createAction(injector),
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
			Name:   "get",
			Usage:  "Get a task by ID",
			Action: getAction(injector),
			Flags: []cli.Flag{
				flags.NewJSONFlag(),
				flags.NewQuietFlag(),
				flags.NewTaskIDFlag(),
			},
		},
		{
			Name:   "list",
			Usage:  "List tasks with filtering options",
			Action: listAction(injector),
			Flags: []cli.Flag{
				flags.NewJSONFlag(),
				flags.NewQuietFlag(),
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
					Name:  "complexity",
					Usage: "Minimum complexity filter - shows tasks with this complexity or higher (1-10)",
				},
				&cli.StringFlag{
					Name:    "search",
					Aliases: []string{"q"},
					Usage:   "Search in task titles and descriptions",
				},
				&cli.IntFlag{
					Name:    "limit",
					Aliases: []string{"l"},
					Usage:   "Maximum number of tasks to show (default: 20)",
					Value:   20,
				},
			},
		},
		{
			Name:  "update",
			Usage: "Update task fields",
			Description: `Update one or more task fields in a single command.
This follows the single responsibility principle by having one command handle all updates.

Examples:
  knot task update --id <task-id> --state in-progress
  knot task update --id <task-id> --title "New title" --priority high
  knot task update --id <task-id> --description "New description" --complexity 6
  knot task update --id <task-id> --state completed --title "Done" --complexity 3`,
			Action: updateAction(injector),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Usage:    "Task ID",
					Required: true,
				},
				&cli.StringFlag{
					Name:    "title",
					Aliases: []string{"t"},
					Usage:   "New task title",
				},
				&cli.StringFlag{
					Name:    "description",
					Aliases: []string{"d"},
					Usage:   "New task description",
				},
				&cli.StringFlag{
					Name:    "state",
					Aliases: []string{"s"},
					Usage:   "New state (pending, in-progress, completed, blocked, cancelled)",
				},
				&cli.StringFlag{
					Name:    "priority",
					Aliases: []string{"p"},
					Usage:   "New task priority (low, medium, high)",
				},
				&cli.IntFlag{
					Name:    "complexity",
					Aliases: []string{"c"},
					Usage:   "New task complexity (1-10)",
				},
			},
		},
	}

	// Hierarchy navigation commands
	hierarchyCommands := HierarchyCommands(injector)

	// Task deletion commands
	deletionCommands := DeletionCommands(injector)

	// Bulk operation commands
	bulkCommands := BulkCommands(injector)

	// Combine all commands
	allCommands := make([]*cli.Command, 0, len(basicCommands)+len(hierarchyCommands)+len(deletionCommands)+len(bulkCommands))
	allCommands = append(allCommands, basicCommands...)
	allCommands = append(allCommands, hierarchyCommands...)
	allCommands = append(allCommands, deletionCommands...)
	allCommands = append(allCommands, bulkCommands...)

	return allCommands
}

func createAction(injector do.Injector) cli.ActionFunc {
	// Resolve dependencies from DI
	loggerService := do.MustInvoke[logger.Logger](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)

	return func(c *cli.Context) error {
		projectID, err := resolveProjectIDWithDI(c, injector)
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
		actor = shared.ResolveActor(actor)

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

		loggerService.Info("Creating task",
			zap.String("title", title),
			zap.String("projectID", projectID.String()),
			zap.Int("complexity", complexity),
			zap.String("priority", priority),
			zap.String("actor", actor))

		task, err := projectManager.CreateTask(context.Background(), projectID, parentID, title, description, complexity, utils.ParsePriority(priority), actor)
		if err != nil {
			loggerService.Error("Failed to create task", zap.Error(err))
			return errors.WrapWithSuggestion(err, "creating task")
		}

		loggerService.Info("Task created successfully", zap.String("taskID", task.ID.String()), zap.String("actor", actor))

		fmt.Printf("Created task: %s (ID: %s)\n", task.Title, task.ID)
		fmt.Printf("  Created by: %s\n", actor)
		if task.Description != "" {
			fmt.Printf("  Description: %s\n", task.Description)
		}
		fmt.Printf("  Complexity: %d\n", task.Complexity)

		fmt.Printf("  Priority: %s\n", task.Priority.ToExternalString())
		fmt.Printf("  State: %s\n", task.State)
		if parentID != nil {
			fmt.Printf("  Parent: %s\n", *parentID)
		}

		// Show workflow reminder for task state management
		fmt.Printf("\nReminder: Set this task to 'in-progress' before starting work:\n")
		fmt.Printf("  knot task update --id %s --state in-progress\n", task.ID)

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

func listAction(injector do.Injector) cli.ActionFunc {
	// Resolve dependencies from DI
	loggerService := do.MustInvoke[logger.Logger](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)

	return func(c *cli.Context) error {
		projectID, err := resolveProjectIDWithDI(c, injector)
		if err != nil {
			return err
		}

		loggerService.Info("Listing tasks", zap.String("projectID", projectID.String()))

		tasks, err := projectManager.ListTasksForProject(context.Background(), projectID)
		if err != nil {
			loggerService.Error("Failed to list tasks", zap.Error(err))
			return errors.WrapWithSuggestion(err, "listing tasks")
		}

		// Apply filters
		filteredTasks := applyTaskFilters(tasks, c)

		// Apply sorting
		sortedTasks := applyTaskSorting(filteredTasks, c)

		// Apply limit
		finalTasks := applyTaskLimit(sortedTasks, c)

		loggerService.Info("Tasks filtered and sorted",
			zap.Int("originalCount", len(tasks)),
			zap.Int("filteredCount", len(finalTasks)))

		if len(finalTasks) == 0 {
			fmt.Printf("No tasks found matching the specified criteria.\n")
			return nil
		}

		// Check if JSON output is requested
		if c.Bool("json") {
			return utils.OutputTasksAsJSON(finalTasks)
		}

		// Show project context indicator using DI
		showProjectContextWithDI(c, injector)

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

			// Show parent information for better hierarchy understanding
			parentInfo := ""
			if task.ParentID != nil {
				parentInfo = fmt.Sprintf(" (Parent: %s)", *task.ParentID)
			}

			fmt.Printf("%s* %s (ID: %s)%s\n", indent, task.Title, task.ID, parentInfo)
			if task.Description != "" {
				fmt.Printf("%s%s\n", indent, strings.Repeat("═", 120))
				wrappedDesc := utils.WrapText(task.Description, 120)
				for _, line := range wrappedDesc {
					fmt.Printf("%s  %s\n", indent, line)
				}
				fmt.Printf("%s%s\n", indent, strings.Repeat("═", 120))
			}

			fmt.Printf("%s  State: %s | Priority: %s | Complexity: %d | Depth: %d\n", indent, task.State, task.Priority.ToExternalString(), task.Complexity, task.Depth)
			fmt.Println()
		}
		return nil
	}
}

func getAction(injector do.Injector) cli.ActionFunc {
	// Resolve dependencies from DI
	loggerService := do.MustInvoke[logger.Logger](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)

	return func(c *cli.Context) error {
		taskIDStr := c.String("id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return errors.InvalidUUIDError("task-id", taskIDStr)
		}

		loggerService.Info("Getting task", zap.String("taskID", taskID.String()))

		// Get the task
		task, err := projectManager.GetTask(context.Background(), taskID)
		if err != nil {
			loggerService.Error("Failed to get task", zap.Error(err))
			return errors.TaskNotFoundError(taskID)
		}

		// Check if JSON output is requested
		if c.Bool("json") {
			return utils.OutputTaskAsJSON(task)
		}

		// Show project context indicator using DI
		showProjectContextWithDI(c, injector)

		// Display task details
		fmt.Printf("Task Details:\n")
		fmt.Printf("  ID: %s\n", task.ID)
		fmt.Printf("  Title: %s\n", task.Title)
		if task.Description != "" {
			fmt.Printf("  Description:\n")
			fmt.Printf("  %s\n", strings.Repeat("═", 120))
			wrappedDesc := utils.WrapText(task.Description, 116) // Leave space for indent
			for _, line := range wrappedDesc {
				fmt.Printf("    %s\n", line)
			}
			fmt.Printf("  %s\n", strings.Repeat("═", 120))
		}
		fmt.Printf("  State: %s\n", task.State)
		fmt.Printf("  Priority: %s\n", task.Priority.ToExternalString())
		fmt.Printf("  Complexity: %d\n", task.Complexity)
		fmt.Printf("  Depth: %d\n", task.Depth)
		fmt.Printf("  Created: %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Updated: %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Created By: %s\n", task.CreatedBy)

		if task.ParentID != nil {
			fmt.Printf("  Parent ID: %s\n", *task.ParentID)
		}

		if task.CompletedAt != nil {
			fmt.Printf("  Completed At: %s\n", task.CompletedAt.Format("2006-01-02 15:04:05"))
		}

		return nil
	}
}

// Helper functions for task filtering, sorting, and limiting

// applyTaskFilters applies all specified filters to the task list
// Knot Task 75010287-8330-4001-8f31-a824ce6c5d09: Simplified task list flags
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
			if task.Priority.ToExternalString() != priority {
				continue
			}
		}

		// Complexity filter - now acts as minimum complexity
		if complexity := c.Int("complexity"); complexity > 0 {
			if task.Complexity < complexity {
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

// applyTaskSorting sorts the task list using hierarchy as default
// Knot Task 75010287-8330-4001-8f31-a824ce6c5d09: Simplified to hierarchy sort only
func applyTaskSorting(tasks []*types.Task, c *cli.Context) []*types.Task {
	// Make a copy to avoid modifying the original slice
	sorted := make([]*types.Task, len(tasks))
	copy(sorted, tasks)

	sort.Slice(sorted, func(i, j int) bool {
		// Hierarchical sort: first by depth, then by creation time within each level
		if sorted[i].Depth != sorted[j].Depth {
			return sorted[i].Depth < sorted[j].Depth
		}
		// Within the same depth, sort by creation time using ID as proxy
		return sorted[i].ID.String() < sorted[j].ID.String()
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
// Knot Task 75010287-8330-4001-8f31-a824ce6c5d09: Simplified filter checks
func hasFiltersApplied(c *cli.Context) bool {
	return c.String("state") != "" ||
		c.String("priority") != "" ||
		c.Int("complexity") > 0 ||
		c.String("search") != "" ||
		c.Int("limit") > 0 && c.Int("limit") < 20 // Only count as filter if less than default
}
