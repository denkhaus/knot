package selection

import (
	"fmt"

	"github.com/google/uuid"
)

// EnhancedError provides detailed error context with task IDs and recovery suggestions
type EnhancedError struct {
	Type           string            `json:"type"`
	Message        string            `json:"message"`
	TaskID         *uuid.UUID        `json:"task_id,omitempty"`
	TaskTitle      string            `json:"task_title,omitempty"`
	ProjectID      *uuid.UUID        `json:"project_id,omitempty"`
	ValidationErrs []ValidationError `json:"validation_errors,omitempty"`
	Suggestions    []string          `json:"suggestions,omitempty"`
	RecoveryAction string            `json:"recovery_action,omitempty"`
}

// Error implements the error interface
func (e EnhancedError) Error() string {
	baseMsg := e.Message
	if e.TaskID != nil {
		baseMsg = fmt.Sprintf("Task %s (%s): %s", *e.TaskID, e.TaskTitle, baseMsg)
	}

	if len(e.Suggestions) > 0 {
		baseMsg += fmt.Sprintf(" Suggestions: %v", e.Suggestions)
	}

	return baseMsg
}

// ErrorBuilder helps construct enhanced errors with context
type ErrorBuilder struct {
	err *EnhancedError
}

// NewErrorBuilder creates a new error builder
func NewErrorBuilder(errorType, message string) *ErrorBuilder {
	return &ErrorBuilder{
		err: &EnhancedError{
			Type:        errorType,
			Message:     message,
			Suggestions: make([]string, 0),
		},
	}
}

// WithTask adds task context to the error
func (eb *ErrorBuilder) WithTask(taskID uuid.UUID, taskTitle string) *ErrorBuilder {
	eb.err.TaskID = &taskID
	eb.err.TaskTitle = taskTitle
	return eb
}

// WithProject adds project context to the error
func (eb *ErrorBuilder) WithProject(projectID uuid.UUID) *ErrorBuilder {
	eb.err.ProjectID = &projectID
	return eb
}

// WithValidationErrors adds validation error details
func (eb *ErrorBuilder) WithValidationErrors(validationErrs []ValidationError) *ErrorBuilder {
	eb.err.ValidationErrs = validationErrs
	return eb
}

// WithSuggestions adds recovery suggestions
func (eb *ErrorBuilder) WithSuggestions(suggestions ...string) *ErrorBuilder {
	eb.err.Suggestions = append(eb.err.Suggestions, suggestions...)
	return eb
}

// WithRecoveryAction adds a specific recovery action
func (eb *ErrorBuilder) WithRecoveryAction(action string) *ErrorBuilder {
	eb.err.RecoveryAction = action
	return eb
}

// Build creates the final enhanced error
func (eb *ErrorBuilder) Build() *EnhancedError {
	return eb.err
}

// Common enhanced error constructors

// NewTaskNotFoundError creates an enhanced error for missing tasks
func NewTaskNotFoundError(taskID uuid.UUID, taskTitle string) *EnhancedError {
	return NewErrorBuilder(ErrorTypeValidation, "task not found in dependency graph").
		WithTask(taskID, taskTitle).
		WithSuggestions("Verify the task exists in the project", "Check if the task ID is correct").
		WithRecoveryAction("Create the missing task or verify the task ID").
		Build()
}

// NewCircularDependencyError creates an enhanced error for circular dependencies
func NewCircularDependencyError(cyclicTasks []uuid.UUID, taskTitles map[uuid.UUID]string) *EnhancedError {
	suggestions := []string{
		"Review and remove circular references",
		"Break the cycle by removing one dependency",
		"Consider creating intermediate tasks to resolve the cycle",
	}

	message := fmt.Sprintf("circular dependencies detected involving %d task(s)", len(cyclicTasks))

	err := NewErrorBuilder(ErrorTypeCircularDep, message).
		WithSuggestions(suggestions...).
		WithRecoveryAction("Resolve circular dependencies before continuing")

	// Add task context for the first few cyclic tasks
	if len(cyclicTasks) > 0 && len(cyclicTasks) <= 3 {
		err.err.TaskID = &cyclicTasks[0]
		if title, exists := taskTitles[cyclicTasks[0]]; exists {
			err.err.TaskTitle = title
		}
	}

	return err.Build()
}

// NewDependencyNotCompletedError creates an enhanced error for incomplete dependencies
func NewDependencyNotCompletedError(taskID uuid.UUID, taskTitle, depTitle string, depID uuid.UUID) *EnhancedError {
	return NewErrorBuilder(ErrorTypeValidation, "dependency not completed").
		WithTask(taskID, taskTitle).
		WithSuggestions(
			fmt.Sprintf("Complete dependency '%s' first", depTitle),
			"Remove the dependency if it's no longer needed",
			"Check if the dependency is marked as completed",
		).
		WithRecoveryAction("Complete or resolve dependency constraints").
		Build()
}

// NewNoActionableTasksError creates an enhanced error when no tasks are actionable
func NewNoActionableTasksError(projectID uuid.UUID, totalTasks int) *EnhancedError {
	suggestions := []string{
		"Complete some dependencies to unblock tasks",
		"Create new tasks without dependencies",
		"Review task states and update as needed",
		"Check if tasks are incorrectly marked as completed",
	}

	message := fmt.Sprintf("no actionable tasks found out of %d total tasks", totalTasks)

	return NewErrorBuilder(ErrorTypeNoActionable, message).
		WithProject(projectID).
		WithSuggestions(suggestions...).
		WithRecoveryAction("Complete existing tasks or create new actionable tasks").
		Build()
}

// NewConfigurationError creates an enhanced error for configuration issues
func NewConfigurationError(configField, value string, validOptions []string) *EnhancedError {
	message := fmt.Sprintf("invalid configuration value for %s: %s", configField, value)
	suggestions := []string{
		fmt.Sprintf("Use one of the valid options: %v", validOptions),
		"Check the configuration documentation",
		"Verify the configuration file format",
	}

	return NewErrorBuilder(ErrorTypeInvalidConfig, message).
		WithSuggestions(suggestions...).
		WithRecoveryAction("Update configuration with valid values").
		Build()
}