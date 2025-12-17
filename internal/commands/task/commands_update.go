package task

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/errors"
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

// updateAction consolidates all task update operations into a single command
// Uses existing update actions as sub-actions to maintain all validation logic
// Reference: Knot Task 15969eda-320f-482b-aac5-ef25e386fbfa
// Reference: Brain Memory 3de5544c-1ac8-40f1-88ae-88a1ae1488a0
func updateAction(injector do.Injector) cli.ActionFunc {
	// Resolve dependencies from DI
	loggerService := do.MustInvoke[logger.Logger](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)

	return func(c *cli.Context) error {
		// Get update values from flags
		title := c.String("title")
		description := c.String("description")
		stateStr := c.String("state")
		priority := c.String("priority")
		complexity := c.Int("complexity")

		// Validate at least one field is provided
		if title == "" && description == "" && stateStr == "" && priority == "" && complexity == 0 {
			return fmt.Errorf("at least one field must be specified for update (title, description, state, priority, or complexity)")
		}

		// Store task ID for reuse
		taskIDStr := c.String("id")

		// Store actor for reuse
		actor := c.String("actor")
		if actor == "" {
			actor = shared.ResolveActor(actor)
			_ = c.Set("actor", actor)
		}

		// Track what was updated for final summary
		var updates []string
		var initialTask *types.Task

		// Get initial task state for comparison
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return errors.InvalidUUIDError("task-id", taskIDStr)
		}
		initialTask, getErr := projectManager.GetTask(context.Background(), taskID)
		if getErr != nil {
			return errors.TaskNotFoundError(taskID)
		}

		// Execute updates using existing actions
		if stateStr != "" {
			// Use state validation and transition logic from updateStateSubAction
			if err := updateStateSubAction(injector)(c); err != nil {
				return err
			}
			updates = append(updates, fmt.Sprintf("State: %s -> %s", initialTask.State, stateStr))
		}

		if title != "" {
			// Use title validation from updateTitleSubAction
			if err := updateTitleSubAction(injector)(c); err != nil {
				return err
			}
			updates = append(updates, fmt.Sprintf("Title: \"%s\" -> \"%s\"", initialTask.Title, title))
		}

		if description != "" {
			// Use description validation from updateDescriptionSubAction
			if err := updateDescriptionSubAction(injector)(c); err != nil {
				return err
			}
			if initialTask.Description == "" {
				updates = append(updates, fmt.Sprintf("Description: (empty) -> \"%s\"", description))
			} else {
				updates = append(updates, fmt.Sprintf("Description: \"%s\" -> \"%s\"", initialTask.Description, description))
			}
		}

		if priority != "" {
			// Use priority validation from updatePrioritySubAction
			if err := updatePrioritySubAction(injector)(c); err != nil {
				return err
			}
			updates = append(updates, fmt.Sprintf("Priority: %s -> %s", initialTask.Priority.ToExternalString(), priority))
		}

		if complexity != 0 {
			// Use complexity validation from updateComplexitySubAction
			if err := updateComplexitySubAction(injector)(c); err != nil {
				return err
			}
			updates = append(updates, fmt.Sprintf("Complexity: %d -> %d", initialTask.Complexity, complexity))
		}

		// Summary
		loggerService.Info("Task updated successfully",
			zap.String("taskID", taskIDStr),
			zap.Strings("updates", updates))

		fmt.Printf("Task updated successfully (ID: %s)\n", taskIDStr)
		fmt.Printf("  Updated by: %s\n", actor)
		for _, update := range updates {
			fmt.Printf("  %s\n", update)
		}

		return nil
	}
}

func updateStateSubAction(injector do.Injector) cli.ActionFunc {
	// Resolve dependencies from DI
	loggerService := do.MustInvoke[logger.Logger](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)

	return func(c *cli.Context) error {
		taskIDStr := c.String("id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return errors.InvalidUUIDError("task-id", taskIDStr)
		}

		stateStr := c.String("state")
		actor := c.String("actor")

		// Default to $USER if actor is not provided
		actor = shared.ResolveActor(actor)

		// Basic state validation
		if err := errors.ValidateTaskState(stateStr); err != nil {
			return err
		}

		newState := types.TaskState(stateStr)

		loggerService.Info("Updating task state",
			zap.String("taskID", taskID.String()),
			zap.String("newState", stateStr),
			zap.String("actor", actor))

		// Resolve project context to ensure task belongs to current project
		projectID, err := resolveProjectIDWithDI(c, injector)
		if err != nil {
			loggerService.Error("Failed to resolve project context", zap.Error(err))
			return err
		}

		loggerService.Info("Task update debug",
			zap.String("taskID", taskID.String()),
			zap.String("currentProjectID", projectID.String()),
			zap.String("newState", stateStr),
			zap.String("actor", actor))

		// Get current task to preserve other fields
		task, err := projectManager.GetTask(context.Background(), taskID)
		if err != nil {
			loggerService.Error("Failed to get task", zap.Error(err))
			return errors.TaskNotFoundError(taskID)
		}

		// Validate task belongs to current project
		if task.ProjectID != projectID {
			return fmt.Errorf("task %s belongs to project %s, but current project is %s",
				taskID, task.ProjectID, projectID)
		}

		// Validate state transition
		validator := validation.NewStateValidator()
		if err := validator.ValidateTransition(task.State, newState, task); err != nil {
			// EnhancedError already contains user-friendly formatting
			// No need to log this as it's a user input validation error
			return err
		}

		// Update task state
		updatedTask, err := projectManager.UpdateTaskState(context.Background(), taskID, newState, actor)
		if err != nil {
			loggerService.Error("Failed to update task state", zap.Error(err))
			return errors.WrapWithSuggestion(err, "updating task state")
		}

		loggerService.Info("Task state updated successfully", zap.String("actor", actor))
		fmt.Printf("Updated task state: %s -> %s\n", task.State, updatedTask.State)
		fmt.Printf("  Updated by: %s\n", actor)
		return nil
	}
}

func updateTitleSubAction(injector do.Injector) cli.ActionFunc {
	// Resolve dependencies from DI
	loggerService := do.MustInvoke[logger.Logger](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)

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
		actor = shared.ResolveActor(actor)

		loggerService.Info("Updating task title",
			zap.String("taskID", taskID.String()),
			zap.String("newTitle", newTitle),
			zap.String("actor", actor))

		// Get current task to check if it exists and get old title
		task, err := projectManager.GetTask(context.Background(), taskID)
		if err != nil {
			loggerService.Error("Failed to get task", zap.Error(err))
			return errors.TaskNotFoundError(taskID)
		}

		oldTitle := task.Title

		// Update task title
		updatedTask, err := projectManager.UpdateTaskTitle(context.Background(), taskID, newTitle, actor)
		if err != nil {
			loggerService.Error("Failed to update task title", zap.Error(err))
			return errors.WrapWithSuggestion(err, "updating task title")
		}

		loggerService.Info("Task title updated successfully", zap.String("actor", actor))
		fmt.Printf("Updated task title: \"%s\" -> \"%s\"\n", oldTitle, updatedTask.Title)
		fmt.Printf("  Updated by: %s\n", actor)
		return nil
	}
}

func updateDescriptionSubAction(injector do.Injector) cli.ActionFunc {
	// Resolve dependencies from DI
	loggerService := do.MustInvoke[logger.Logger](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)

	return func(c *cli.Context) error {
		taskIDStr := c.String("id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return errors.InvalidUUIDError("task-id", taskIDStr)
		}

		newDescription := c.String("description")
		actor := c.String("actor")

		// Default to $USER if actor is not provided
		actor = shared.ResolveActor(actor)

		loggerService.Info("Updating task description",
			zap.String("taskID", taskID.String()),
			zap.String("newDescription", newDescription),
			zap.String("actor", actor))

		// Get current task to check if it exists and get old description
		task, err := projectManager.GetTask(context.Background(), taskID)
		if err != nil {
			loggerService.Error("Failed to get task", zap.Error(err))
			return errors.TaskNotFoundError(taskID)
		}

		oldDescription := task.Description

		// Update task description
		updatedTask, err := projectManager.UpdateTaskDescription(context.Background(), taskID, newDescription, actor)
		if err != nil {
			loggerService.Error("Failed to update task description", zap.Error(err))
			return errors.WrapWithSuggestion(err, "updating task description")
		}

		loggerService.Info("Task description updated successfully", zap.String("actor", actor))
		if oldDescription == "" {
			fmt.Printf("Updated task description: (empty) -> \"%s\"\n", updatedTask.Description)
		} else {
			fmt.Printf("Updated task description: \"%s\" -> \"%s\"\n", oldDescription, updatedTask.Description)
		}
		fmt.Printf("  Updated by: %s\n", actor)
		return nil
	}
}

func updatePrioritySubAction(injector do.Injector) cli.ActionFunc {
	// Resolve dependencies from DI
	loggerService := do.MustInvoke[logger.Logger](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)

	return func(c *cli.Context) error {
		taskIDStr := c.String("id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return errors.InvalidUUIDError("task-id", taskIDStr)
		}

		priority := c.String("priority")
		actor := c.String("actor")

		// Default to $USER if actor is not provided
		actor = shared.ResolveActor(actor)

		// Validate priority
		validator := validation.NewInputValidator()
		if err := validator.ValidateTaskPriority(priority); err != nil {
			return errors.NewValidationError("invalid priority", err)
		}

		loggerService.Info("Updating task priority",
			zap.String("taskID", taskID.String()),
			zap.String("newPriority", priority),
			zap.String("actor", actor))

		// Get current task to check if it exists and get old priority
		task, err := projectManager.GetTask(context.Background(), taskID)
		if err != nil {
			loggerService.Error("Failed to get task", zap.Error(err))
			return errors.TaskNotFoundError(taskID)
		}

		oldPriority := task.Priority

		// Update task priority using the service method
		updatedTask, err := projectManager.UpdateTaskPriority(context.Background(), taskID, utils.ParsePriority(priority), actor)
		if err != nil {
			loggerService.Error("Failed to update task priority", zap.Error(err))
			return errors.WrapWithSuggestion(err, "updating task priority")
		}

		loggerService.Info("Task priority updated successfully", zap.String("actor", actor))
		fmt.Printf("Updated task priority: \"%s\" -> \"%s\"\n", oldPriority.ToExternalString(), updatedTask.Priority.ToExternalString())
		fmt.Printf("  Updated by: %s\n", actor)
		return nil
	}
}

func updateComplexitySubAction(injector do.Injector) cli.ActionFunc {
	// Resolve dependencies from DI
	loggerService := do.MustInvoke[logger.Logger](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)

	return func(c *cli.Context) error {
		taskIDStr := c.String("id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return errors.InvalidUUIDError("task-id", taskIDStr)
		}

		complexity := c.Int("complexity")
		actor := c.String("actor")

		// Default to $USER if actor is not provided
		actor = shared.ResolveActor(actor)

		// Validate complexity
		if err := errors.ValidateComplexity(complexity); err != nil {
			return err
		}

		loggerService.Info("Updating task complexity",
			zap.String("taskID", taskID.String()),
			zap.Int("newComplexity", complexity),
			zap.String("actor", actor))

		// Get current task to check if it exists and get old complexity
		task, err := projectManager.GetTask(context.Background(), taskID)
		if err != nil {
			loggerService.Error("Failed to get task", zap.Error(err))
			return errors.TaskNotFoundError(taskID)
		}

		oldComplexity := task.Complexity

		// Update task complexity using the service method
		updatedTask, err := projectManager.UpdateTaskComplexity(context.Background(), taskID, complexity, actor)
		if err != nil {
			loggerService.Error("Failed to update task complexity", zap.Error(err))
			return errors.WrapWithSuggestion(err, "updating task complexity")
		}

		loggerService.Info("Task complexity updated successfully", zap.String("actor", actor))
		fmt.Printf("Updated task complexity: %d -> %d\n", oldComplexity, updatedTask.Complexity)
		fmt.Printf("  Updated by: %s\n", actor)
		return nil
	}
}
