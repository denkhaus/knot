package errors

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// EnhancedError wraps an error with helpful suggestions and examples
type EnhancedError struct {
	Operation   string
	Cause       error
	Suggestion  string
	Example     string
	HelpCommand string
}

func (e *EnhancedError) Error() string {
	var parts []string
	
	// Main error message
	if e.Cause != nil {
		parts = append(parts, e.Cause.Error())
	} else {
		parts = append(parts, fmt.Sprintf("Error in %s", e.Operation))
	}
	
	// Add suggestion if available
	if e.Suggestion != "" {
		parts = append(parts, fmt.Sprintf("Suggestion: %s", e.Suggestion))
	}
	
	// Add example if available
	if e.Example != "" {
		parts = append(parts, fmt.Sprintf("Example: %s", e.Example))
	}
	
	// Add help command if available
	if e.HelpCommand != "" {
		parts = append(parts, fmt.Sprintf("For more help: %s", e.HelpCommand))
	}
	
	return strings.Join(parts, "\n")
}

func (e *EnhancedError) Unwrap() error {
	return e.Cause
}

// Common error enhancement functions

// InvalidUUIDError creates an enhanced error for invalid UUID parsing
func InvalidUUIDError(fieldName, value string) *EnhancedError {
	return &EnhancedError{
		Operation:   fmt.Sprintf("parsing %s", fieldName),
		Cause:       fmt.Errorf("invalid %s format: '%s'", fieldName, value),
		Suggestion:  fmt.Sprintf("Ensure %s is a valid UUID (36 characters with hyphens)", fieldName),
		Example:     "550e8400-e29b-41d4-a716-446655440000",
		HelpCommand: "knot project list  # to see available project IDs",
	}
}

// TaskNotFoundError creates an enhanced error for missing tasks
func TaskNotFoundError(taskID uuid.UUID) *EnhancedError {
	return &EnhancedError{
		Operation:   "finding task",
		Cause:       fmt.Errorf("task not found: %s", taskID),
		Suggestion:  "Verify the task ID exists in the current project",
		Example:     "knot task list  # to see available tasks",
		HelpCommand: "knot project list  # to see available projects",
	}
}

// ProjectNotFoundError creates an enhanced error for missing projects
func ProjectNotFoundError(projectID uuid.UUID) *EnhancedError {
	return &EnhancedError{
		Operation:   "finding project",
		Cause:       fmt.Errorf("project not found: %s", projectID),
		Suggestion:  "Check if the project exists or create a new one",
		Example:     "knot project create --title \"My Project\" --description \"Description\"",
		HelpCommand: "knot project list  # to see available projects",
	}
}

// InvalidTaskStateError creates an enhanced error for invalid task states
func InvalidTaskStateError(state string) *EnhancedError {
	validStates := []string{"pending", "in-progress", "completed", "blocked", "cancelled"}
	return &EnhancedError{
		Operation:   "validating task state",
		Cause:       fmt.Errorf("invalid task state: '%s'", state),
		Suggestion:  fmt.Sprintf("Use one of the valid states: %s", strings.Join(validStates, ", ")),
		Example:     "knot task update-state --id <task-id> --state completed",
		HelpCommand: "knot task update-state --help",
	}
}

// CircularDependencyError creates an enhanced error for circular dependencies
func CircularDependencyError(taskID, dependsOnID uuid.UUID) *EnhancedError {
	return &EnhancedError{
		Operation:   "adding task dependency",
		Cause:       fmt.Errorf("circular dependency detected between %s and %s", taskID, dependsOnID),
		Suggestion:  "Remove existing dependencies that create a cycle, or restructure your task hierarchy",
		Example:     "knot dependency cycles  # to detect all cycles",
		HelpCommand: "knot dependency validate",
	}
}

// DatabaseConnectionError creates an enhanced error for database issues
func DatabaseConnectionError(operation string, cause error) *EnhancedError {
	return &EnhancedError{
		Operation:  operation,
		Cause:      cause,
		Suggestion: "Check if the .knot directory exists and is writable, or try running from a different directory",
		Example:    "ls -la .knot/  # check database directory permissions",
		HelpCommand: "knot project list  # test database connectivity",
	}
}

// MissingRequiredFlagError creates an enhanced error for missing CLI flags
func MissingRequiredFlagError(flagName, commandContext string) *EnhancedError {
	var example string
	var helpCommand string
	
	if commandContext == "" {
		commandContext = "command"
	}
	
	switch flagName {
	case "project-id":
		example = "knot " + commandContext + " (select a project first with: knot project select --id <project-id>)"
		helpCommand = "knot project list  # to see available projects"
	case "task-id":
		example = "knot " + commandContext + " --task-id <task-id>"
		helpCommand = "knot task list  # to see available tasks"
	default:
		example = "knot " + commandContext + " --" + flagName + " <value>"
		helpCommand = "knot --help"
	}
	
	// Safe help command generation
	if commandContext != "" && commandContext != "command" {
		fields := strings.Fields(commandContext)
		if len(fields) > 0 {
			helpCommand = "knot " + fields[0] + " --help"
		}
	}
	
	var flagType string
	if flagName == "project-id" {
		// project-id is no longer a global flag, suggest using project selection instead
		return &EnhancedError{
			Operation:   "project context resolution",
			Cause:       fmt.Errorf("no project is currently selected"),
			Suggestion:  "Select a project first using the project selection command",
			Example:     "knot project select --id <project-id>",
			HelpCommand: "knot project list  # to see available projects",
		}
	} else {
		flagType = "required flag"
	}
	
	return &EnhancedError{
		Operation:   "parsing command flags",
		Cause:       fmt.Errorf("%s --%s not provided", flagType, flagName),
		Suggestion:  fmt.Sprintf("Add the --%s flag with a valid value", flagName),
		Example:     example,
		HelpCommand: helpCommand,
	}
}

// ComplexityOutOfRangeError creates an enhanced error for invalid complexity values
func ComplexityOutOfRangeError(complexity int) *EnhancedError {
	return &EnhancedError{
		Operation:   "validating task complexity",
		Cause:       fmt.Errorf("complexity %d is out of range", complexity),
		Suggestion:  "Use a complexity value between 1 and 10 (1=very simple, 10=very complex)",
		Example:     "knot task create --title \"Task\" --complexity 5",
		HelpCommand: "knot task create --help",
	}
}

// TooManyTasksError creates an enhanced error for task limits
func TooManyTasksError(currentCount, maxAllowed int, depth int) *EnhancedError {
	return &EnhancedError{
		Operation:   "creating task",
		Cause:       fmt.Errorf("maximum tasks per depth exceeded: %d/%d at depth %d", currentCount, maxAllowed, depth),
		Suggestion:  "Break down existing complex tasks into subtasks, or increase the limit via environment variable",
		Example:     "export KNOT_MAX_TASKS_PER_DEPTH=200  # increase limit",
		HelpCommand: "knot breakdown  # find tasks to break down",
	}
}

// NewValidationError creates an enhanced error for validation failures
func NewValidationError(message string, cause error) *EnhancedError {
	return &EnhancedError{
		Operation:   "input validation",
		Cause:       cause,
		Suggestion:  "Check your input and try again with valid values",
		Example:     "Ensure titles are under 200 characters and descriptions under 2000 characters",
		HelpCommand: "knot --help",
	}
}

// EmptyResultError creates an enhanced error for empty query results
func EmptyResultError(operation, context string) *EnhancedError {
	var suggestion, example string
	
	switch operation {
	case "list tasks":
		suggestion = "Create some tasks first, or check if you're using the correct project ID"
		example = "knot task create --title \"First Task\""
	case "list projects":
		suggestion = "Create a project first to get started"
		example = "knot project create --title \"My Project\" --description \"Project description\""
	case "find actionable tasks":
		suggestion = "Check if all tasks are completed, blocked, or have unmet dependencies"
		example = "knot ready  # see tasks ready to work on"
	default:
		suggestion = "Check your query parameters or create the required resources first"
		example = "knot --help  # see available commands"
	}
	
	return &EnhancedError{
		Operation:   operation,
		Cause:       fmt.Errorf("no results found for %s", context),
		Suggestion:  suggestion,
		Example:     example,
		HelpCommand: "knot --help",
	}
}
// NoProjectContextError creates an enhanced error when no project context is available
func NoProjectContextError() *EnhancedError {
	return &EnhancedError{
		Operation:   "project context resolution",
		Cause:       fmt.Errorf("no project is currently selected"),
		Suggestion:  "Select a project first to work with tasks and other project-specific commands",
		Example:     "knot project select --id <project-id>",
		HelpCommand: "knot project list  # to see available projects",
	}
}
