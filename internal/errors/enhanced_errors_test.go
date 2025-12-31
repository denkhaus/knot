package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEnhancedErrorError(t *testing.T) {
	// Test with all fields populated
	err := &EnhancedError{
		Operation:   "test operation",
		Cause:       fmt.Errorf("original error"),
		Suggestion:  "try again",
		Example:     "example command",
		HelpCommand: "help command",
	}

	expectedParts := []string{
		"original error",
		"Suggestion: try again",
		"Example: example command",
		"For more help: help command",
	}

	result := err.Error()
	for _, part := range expectedParts {
		assert.Contains(t, result, part)
	}
}

func TestEnhancedErrorUnwrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	enhancedErr := &EnhancedError{
		Cause: originalErr,
	}

	unwrapped := enhancedErr.Unwrap()
	assert.Equal(t, originalErr, unwrapped)
}

func TestEnhancedErrorWithoutCause(t *testing.T) {
	err := &EnhancedError{
		Operation:  "test operation",
		Suggestion: "try again",
	}

	result := err.Error()
	assert.Contains(t, result, "Error in test operation")
	assert.Contains(t, result, "Suggestion: try again")
}

func TestInvalidUUIDError(t *testing.T) {
	fieldName := "task-id"
	value := "invalid-uuid"

	err := InvalidUUIDError(fieldName, value)

	assert.Equal(t, "parsing task-id", err.Operation)
	assert.Contains(t, err.Cause.Error(), "invalid task-id format")
	assert.Contains(t, err.Suggestion, "Ensure task-id is a valid UUID")
	assert.Contains(t, err.Example, "550e8400-e29b-41d4-a716-446655440000")
	assert.Contains(t, err.HelpCommand, "knot project list")
}

func TestTaskNotFoundError(t *testing.T) {
	taskID := uuid.New()

	err := TaskNotFoundError(taskID)

	assert.Equal(t, "finding task", err.Operation)
	assert.Contains(t, err.Cause.Error(), taskID.String())
	assert.Contains(t, err.Suggestion, "Verify the task ID exists")
	assert.Contains(t, err.Example, "knot task list")
	assert.Contains(t, err.HelpCommand, "knot project list")
}

func TestProjectNotFoundError(t *testing.T) {
	projectID := uuid.New()

	err := ProjectNotFoundError(projectID)

	assert.Equal(t, "finding project", err.Operation)
	assert.Contains(t, err.Cause.Error(), projectID.String())
	assert.Contains(t, err.Suggestion, "Check if the project exists")
	assert.Contains(t, err.Example, "knot project create")
	assert.Contains(t, err.HelpCommand, "knot project list")
}

func TestInvalidTaskStateError(t *testing.T) {
	state := "invalid-state"

	err := InvalidTaskStateError(state)

	assert.Equal(t, "validating task state", err.Operation)
	assert.Contains(t, err.Cause.Error(), "invalid task state")
	assert.Contains(t, err.Suggestion, "Use one of the valid states")
	assert.Contains(t, err.Suggestion, "pending, in-progress, completed, blocked, cancelled")
	assert.Contains(t, err.Example, "knot task update")
	assert.Contains(t, err.HelpCommand, "knot task update --help")
}

func TestCircularDependencyError(t *testing.T) {
	taskID := uuid.New()
	dependsOnID := uuid.New()

	err := CircularDependencyError(taskID, dependsOnID)

	assert.Equal(t, "adding task dependency", err.Operation)
	assert.Contains(t, err.Cause.Error(), taskID.String())
	assert.Contains(t, err.Cause.Error(), dependsOnID.String())
	assert.Contains(t, err.Suggestion, "Remove existing dependencies")
	assert.Contains(t, err.Example, "knot dependency cycles")
	assert.Contains(t, err.HelpCommand, "knot dependency validate")
}

func TestDatabaseConnectionError(t *testing.T) {
	operation := "database operation"
	cause := fmt.Errorf("connection failed")

	err := DatabaseConnectionError(operation, cause)

	assert.Equal(t, operation, err.Operation)
	assert.Equal(t, cause, err.Cause)
	assert.Contains(t, err.Suggestion, "Check if the .knot directory exists")
	assert.Contains(t, err.Example, "ls -la .knot/")
	assert.Contains(t, err.HelpCommand, "knot project list")
}

func TestMissingRequiredFlagError(t *testing.T) {
	tests := []struct {
		name                 string
		flagName             string
		context              string
		expectProjectContext bool
	}{
		{
			name:                 "project-id flag",
			flagName:             "project-id",
			context:              "",
			expectProjectContext: true,
		},
		{
			name:                 "task-id flag",
			flagName:             "task-id",
			context:              "task update",
			expectProjectContext: false,
		},
		{
			name:                 "other flag",
			flagName:             "other-flag",
			context:              "command",
			expectProjectContext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MissingRequiredFlagError(tt.flagName, tt.context)

			if tt.expectProjectContext {
				// For project-id, it uses project context logic
				assert.Equal(t, "project context resolution", err.Operation)
				assert.Contains(t, err.Cause.Error(), "no project is currently selected")
				assert.Contains(t, err.Suggestion, "Select a project first")
			} else {
				assert.Equal(t, "parsing command flags", err.Operation)
				assert.Contains(t, err.Cause.Error(), tt.flagName)
				assert.Contains(t, err.Suggestion, "Add the --"+tt.flagName+" flag")
			}
		})
	}
}

func TestComplexityOutOfRangeError(t *testing.T) {
	complexity := 15

	err := ComplexityOutOfRangeError(complexity)

	assert.Equal(t, "validating task complexity", err.Operation)
	assert.Contains(t, err.Cause.Error(), "complexity 15 is out of range")
	assert.Contains(t, err.Suggestion, "Use a complexity value between 1 and 10")
	assert.Contains(t, err.Example, "knot task create")
	assert.Contains(t, err.HelpCommand, "knot task create --help")
}

func TestTooManyTasksError(t *testing.T) {
	currentCount := 50
	maxAllowed := 20
	depth := 3

	err := TooManyTasksError(currentCount, maxAllowed, depth)

	assert.Equal(t, "creating task", err.Operation)
	assert.Contains(t, err.Cause.Error(), "maximum tasks per depth exceeded: 50/20 at depth 3")
	assert.Contains(t, err.Suggestion, "Break down existing complex tasks")
	assert.Contains(t, err.Example, "export KNOT_MAX_TASKS_PER_DEPTH=200")
	assert.Contains(t, err.HelpCommand, "knot breakdown")
}

func TestNewValidationError(t *testing.T) {
	message := "validation failed"
	cause := fmt.Errorf("field error")

	err := NewValidationError(message, cause)

	assert.Equal(t, "input validation", err.Operation)
	assert.Equal(t, cause, err.Cause)
	assert.Contains(t, err.Suggestion, "Check your input")
	assert.Contains(t, err.Example, "Ensure titles are under 200 characters")
	assert.Contains(t, err.HelpCommand, "knot --help")
}

func TestEmptyResultError(t *testing.T) {
	tests := []struct {
		name               string
		operation          string
		context            string
		expectedSuggestion string
	}{
		{
			name:               "list tasks",
			operation:          "list tasks",
			context:            "tasks",
			expectedSuggestion: "Create some tasks first",
		},
		{
			name:               "list projects",
			operation:          "list projects",
			context:            "projects",
			expectedSuggestion: "Create a project first",
		},
		{
			name:               "find actionable tasks",
			operation:          "find actionable tasks",
			context:            "actionable tasks",
			expectedSuggestion: "Check if all tasks are completed",
		},
		{
			name:               "other operation",
			operation:          "other operation",
			context:            "other",
			expectedSuggestion: "Check your query parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EmptyResultError(tt.operation, tt.context)

			assert.Equal(t, tt.operation, err.Operation)
			assert.Contains(t, err.Cause.Error(), "no results found")
			assert.Contains(t, err.Suggestion, tt.expectedSuggestion)
			assert.Contains(t, err.HelpCommand, "knot --help")
		})
	}
}

func TestNoProjectContextError(t *testing.T) {
	err := NoProjectContextError()

	assert.Equal(t, "project context resolution", err.Operation)
	assert.Contains(t, err.Cause.Error(), "no project is currently selected")
	assert.Contains(t, err.Suggestion, "Select a project first")
	assert.Contains(t, err.Example, "knot project select")
	assert.Contains(t, err.HelpCommand, "knot project list")
}

func TestEnhancedErrorWithEmptyFields(t *testing.T) {
	// Test with empty suggestion, example, and help command
	err := &EnhancedError{
		Operation: "test",
		Cause:     fmt.Errorf("test error"),
		// All other fields are empty
	}

	result := err.Error()
	assert.Contains(t, result, "test error")
	// Should not contain suggestion/example/help sections since they're empty
	lines := strings.Split(result, "\n")
	assert.Len(t, lines, 1) // Only the main error message
}

func TestEnhancedErrorErrorWithMultilineError(t *testing.T) {
	// Test error with cause that has multiline message
	err := &EnhancedError{
		Operation:  "test operation",
		Cause:      errors.New("line1\nline2\nline3"),
		Suggestion: "try again",
	}

	result := err.Error()
	assert.Contains(t, result, "line1")
	assert.Contains(t, result, "line2")
	assert.Contains(t, result, "line3")
	assert.Contains(t, result, "Suggestion: try again")
}

func TestEnhancedErrorWithNilCause(t *testing.T) {
	err := &EnhancedError{
		Operation:  "test operation",
		Cause:      nil,
		Suggestion: "try again",
	}

	result := err.Error()
	assert.Contains(t, result, "Error in test operation")
	assert.Contains(t, result, "Suggestion: try again")
}

func TestEnhancedErrorFormatting(t *testing.T) {
	// Test proper formatting with multiple sections
	err := &EnhancedError{
		Operation:   "test op",
		Cause:       fmt.Errorf("cause error"),
		Suggestion:  "suggestion text",
		Example:     "example text",
		HelpCommand: "help command",
	}

	result := err.Error()
	parts := strings.Split(result, "\n")

	assert.Len(t, parts, 4) // 4 parts: cause, suggestion, example, help
	assert.Contains(t, parts[0], "cause error")
	assert.Contains(t, parts[1], "Suggestion: suggestion text")
	assert.Contains(t, parts[2], "Example: example text")
	assert.Contains(t, parts[3], "For more help: help command")
}

func TestEnhancedErrorUnwrapNil(t *testing.T) {
	err := &EnhancedError{
		Cause: nil,
	}

	unwrapped := err.Unwrap()
	assert.Nil(t, unwrapped)
}

func TestWrapWithSuggestion(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		operation  string
		wantType   string
		wantContains string
	}{
		{
			name:       "nil error returns nil",
			err:        nil,
			operation:  "test",
			wantType:   "nil",
		},
		{
			name:       "UUID parsing error for project",
			err:        fmt.Errorf("invalid UUID length: 'not-a-uuid'"),
			operation:  "project",
			wantType:   "EnhancedError",
			wantContains: "invalid project-id format",
		},
		{
			name:       "UUID parsing error for task",
			err:        fmt.Errorf("invalid UUID length: 'not-a-uuid'"),
			operation:  "task",
			wantType:   "EnhancedError",
			wantContains: "invalid task-id format",
		},
		{
			name:       "not found error for project",
			err:        fmt.Errorf("project not found"),
			operation:  "project",
			wantType:   "error",
		},
		{
			name:       "database connection error",
			err:        fmt.Errorf("database connection failed"),
			operation:  "connect",
			wantType:   "EnhancedError",
			wantContains: "database connection",
		},
		{
			name:       "state validation error",
			err:        fmt.Errorf("invalid state: 'bad-state'"),
			operation:  "state",
			wantType:   "EnhancedError",
			wantContains: "invalid task state",
		},
		{
			name:       "circular dependency error",
			err:        fmt.Errorf("circular dependency detected"),
			operation:  "dependency",
			wantType:   "EnhancedError",
			wantContains: "circular",
		},
		{
			name:       "unknown error returns as-is",
			err:        fmt.Errorf("some unknown error"),
			operation:  "test",
			wantType:   "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapWithSuggestion(tt.err, tt.operation)

			if tt.wantType == "nil" {
				assert.Nil(t, result)
			} else if tt.wantType == "EnhancedError" {
				assert.IsType(t, &EnhancedError{}, result)
				if tt.wantContains != "" {
					assert.Contains(t, result.Error(), tt.wantContains)
				}
			} else {
				assert.Equal(t, tt.err, result)
			}
		})
	}
}

func TestHandleCLIError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		operation  string
		wantContains string
	}{
		{
			name:       "nil error returns nil",
			err:        nil,
			operation:  "test",
		},
		{
			name:       "missing required flag error",
			err:        fmt.Errorf("Required flag \"task-id\" not set"),
			operation:  "task update",
			wantContains: "required flag",
		},
		{
			name:       "other error gets wrapped",
			err:        fmt.Errorf("some error"),
			operation:  "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For nil context, WrapWithSuggestion is called which doesn't crash
			result := HandleCLIError(nil, tt.err, tt.operation)

			if tt.err == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				if tt.wantContains != "" {
					assert.Contains(t, result.Error(), tt.wantContains)
				}
			}
		})
	}
}

func TestValidateComplexity(t *testing.T) {
	tests := []struct {
		name        string
		complexity  int
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid complexity",
			complexity: 5,
			wantErr:    false,
		},
		{
			name:        "complexity too low",
			complexity:  0,
			wantErr:     true,
			errContains: "out of range",
		},
		{
			name:        "complexity too high",
			complexity:  11,
			wantErr:     true,
			errContains: "out of range",
		},
		{
			name:       "minimum valid complexity",
			complexity: 1,
			wantErr:    false,
		},
		{
			name:       "maximum valid complexity",
			complexity: 10,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateComplexity(tt.complexity)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTaskState(t *testing.T) {
	tests := []struct {
		name        string
		state       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid state: pending",
			state:   "pending",
			wantErr: false,
		},
		{
			name:    "valid state: in-progress",
			state:   "in-progress",
			wantErr: false,
		},
		{
			name:    "valid state: completed",
			state:   "completed",
			wantErr: false,
		},
		{
			name:    "valid state: blocked",
			state:   "blocked",
			wantErr: false,
		},
		{
			name:    "valid state: cancelled",
			state:   "cancelled",
			wantErr: false,
		},
		{
			name:        "invalid state",
			state:       "invalid-state",
			wantErr:     true,
			errContains: "invalid task state",
		},
		{
			name:        "empty state",
			state:       "",
			wantErr:     true,
			errContains: "invalid task state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTaskState(tt.state)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFormatSuggestionList(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		suggestions []string
		wantContains string
	}{
		{
			name:        "empty suggestions returns empty string",
			title:       "Test",
			suggestions: []string{},
			wantContains: "",
		},
		{
			name:        "single suggestion",
			title:       "Test Title",
			suggestions: []string{"suggestion 1"},
			wantContains: "💡 Test Title:",
		},
		{
			name:        "multiple suggestions",
			title:       "Test Title",
			suggestions: []string{"suggestion 1", "suggestion 2", "suggestion 3"},
			wantContains: "   3. suggestion 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSuggestionList(tt.title, tt.suggestions)

			if tt.wantContains == "" {
				assert.Empty(t, result)
			} else {
				assert.Contains(t, result, tt.wantContains)
			}
		})
	}
}

func TestExtractValueFromError(t *testing.T) {
	tests := []struct {
		name  string
		errStr string
		want  string
	}{
		{
			name:  "extract value from single quotes",
			errStr: "invalid format: 'test-value'",
			want:  "test-value",
		},
		{
			name:  "extract value from double quotes",
			errStr: `invalid format: "test-value"`,
			want:  "test-value",
		},
		{
			name:  "no quotes returns empty",
			errStr: "invalid format test-value",
			want:  "",
		},
		{
			name:  "mixed quotes extracts first",
			errStr: `"first" and 'second'`,
			want:  "second",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractValueFromError(tt.errStr)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestExtractUUIDFromError(t *testing.T) {
	validUUID := uuid.New()
	tests := []struct {
		name  string
		errStr string
		want  uuid.UUID
	}{
		{
			name:  "extract valid UUID",
			errStr: fmt.Sprintf("task %s not found", validUUID),
			want:  validUUID,
		},
		{
			name:  "no UUID returns nil",
			errStr: "task not found",
			want:  uuid.Nil,
		},
		{
			name:  "multiple UUIDs returns first",
			errStr: fmt.Sprintf("task %s depends on %s", validUUID, uuid.New()),
			want:  validUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractUUIDFromError(tt.errStr)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestExtractFlagNameFromError(t *testing.T) {
	tests := []struct {
		name  string
		errStr string
		want  string
	}{
		{
			name:  "extract flag name from error",
			errStr: `Required flag "task-id" not set`,
			want:  "task-id",
		},
		{
			name:  "no quotes returns unknown",
			errStr: "Required flag task-id not set",
			want:  "unknown",
		},
		{
			name:  "empty string returns unknown",
			errStr: "",
			want:  "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFlagNameFromError(tt.errStr)
			assert.Equal(t, tt.want, result)
		})
	}
}

