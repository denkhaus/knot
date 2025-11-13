package errors

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
)

// WrapWithSuggestion wraps any error with helpful suggestions based on context
func WrapWithSuggestion(err error, operation string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// UUID parsing errors
	if strings.Contains(errStr, "invalid UUID") {
		if strings.Contains(operation, "project") {
			return InvalidUUIDError("project-id", extractValueFromError(errStr))
		}
		if strings.Contains(operation, "task") {
			return InvalidUUIDError("task-id", extractValueFromError(errStr))
		}
	}

	// Not found errors
	if strings.Contains(errStr, "not found") {
		if strings.Contains(operation, "project") {
			if id := extractUUIDFromError(errStr); id != uuid.Nil {
				return ProjectNotFoundError(id)
			}
		}
		if strings.Contains(operation, "task") {
			if id := extractUUIDFromError(errStr); id != uuid.Nil {
				return TaskNotFoundError(id)
			}
		}
	}

	// Database connection errors
	if strings.Contains(errStr, "database") || strings.Contains(errStr, "connection") {
		return DatabaseConnectionError(operation, err)
	}

	// State validation errors
	if strings.Contains(errStr, "invalid.*state") || strings.Contains(operation, "state") {
		if state := extractValueFromError(errStr); state != "" {
			return InvalidTaskStateError(state)
		}
	}

	// Circular dependency errors
	if strings.Contains(errStr, "circular") || strings.Contains(errStr, "cycle") {
		return CircularDependencyError(uuid.Nil, uuid.Nil) // Generic circular dependency error
	}

	// Return original error if no enhancement is available
	return err
}

// HandleCLIError provides enhanced error handling for CLI commands
func HandleCLIError(c *cli.Context, err error, operation string) error {
	if err == nil {
		return nil
	}

	// Check for missing required flags
	if strings.Contains(err.Error(), "Required flag") {
		flagName := extractFlagNameFromError(err.Error())
		commandContext := buildCommandContext(c)
		return MissingRequiredFlagError(flagName, commandContext)
	}

	// Wrap with context-aware suggestions
	return WrapWithSuggestion(err, operation)
}

// extractValueFromError extracts a value from error messages like "invalid format: 'value'"
func extractValueFromError(errStr string) string {
	// Look for patterns like 'value' or "value"
	if start := strings.Index(errStr, "'"); start != -1 {
		if end := strings.Index(errStr[start+1:], "'"); end != -1 {
			return errStr[start+1 : start+1+end]
		}
	}
	if start := strings.Index(errStr, "\""); start != -1 {
		if end := strings.Index(errStr[start+1:], "\""); end != -1 {
			return errStr[start+1 : start+1+end]
		}
	}
	return ""
}

// extractUUIDFromError attempts to extract a UUID from error messages
func extractUUIDFromError(errStr string) uuid.UUID {
	words := strings.Fields(errStr)
	for _, word := range words {
		if id, err := uuid.Parse(word); err == nil {
			return id
		}
	}
	return uuid.Nil
}

// extractFlagNameFromError extracts flag name from CLI error messages
func extractFlagNameFromError(errStr string) string {
	// Look for patterns like "Required flag \"flag-name\""
	if start := strings.Index(errStr, "\""); start != -1 {
		if end := strings.Index(errStr[start+1:], "\""); end != -1 {
			return errStr[start+1 : start+1+end]
		}
	}
	return "unknown"
}

// buildCommandContext builds a command context string for help messages
func buildCommandContext(c *cli.Context) string {
	var parts []string

	// Add command hierarchy
	if c.Command != nil {
		parts = append(parts, c.Command.Name)
	}

	// Add parent command if available
	if c.Command != nil && c.Command.Category != "" {
		parts = append([]string{c.Command.Category}, parts...)
	}

	// Fallback to app name
	if len(parts) == 0 && c.App != nil {
		parts = append(parts, c.App.Name)
	}

	return strings.Join(parts, " ")
}

// ValidateComplexity validates task complexity and returns enhanced error if invalid
func ValidateComplexity(complexity int) error {
	if complexity < 1 || complexity > 10 {
		return ComplexityOutOfRangeError(complexity)
	}
	return nil
}

// ValidateTaskState validates task state and returns enhanced error if invalid
func ValidateTaskState(state string) error {
	validStates := map[string]bool{
		"pending":     true,
		"in-progress": true,
		"completed":   true,
		"blocked":     true,
		"cancelled":   true,
	}

	if !validStates[state] {
		return InvalidTaskStateError(state)
	}
	return nil
}

// FormatSuggestionList formats a list of suggestions for display
func FormatSuggestionList(title string, suggestions []string) string {
	if len(suggestions) == 0 {
		return ""
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("💡 %s:", title))

	for i, suggestion := range suggestions {
		parts = append(parts, fmt.Sprintf("   %d. %s", i+1, suggestion))
	}

	return strings.Join(parts, "\n")
}
