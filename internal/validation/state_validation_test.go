package validation

import (
	"testing"
	"time"

	"github.com/denkhaus/knot/internal/errors"
	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidationErrorHandling tests that validation errors provide clean, helpful messages
func TestValidationErrorHandling(t *testing.T) {
	validator := NewStateValidator()
	
	// Create a test task
	task := &types.Task{
		ID:          uuid.New(),
		ProjectID:   uuid.New(),
		Title:       "Test Task",
		Description: "Test task for validation",
		State:       types.TaskStatePending,
		Complexity:  5,
		Depth:       0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name           string
		fromState      types.TaskState
		toState        types.TaskState
		task           *types.Task
		expectError    bool
		errorType      string
		errorContains  []string
	}{
		{
			name:          "Invalid state transition - pending to completed",
			fromState:     types.TaskStatePending,
			toState:       types.TaskStateCompleted,
			task:          task,
			expectError:   true,
			errorType:     "*errors.EnhancedError",
			errorContains: []string{"transition", "pending", "completed", "in-progress"},
		},
		{
			name:          "Valid state transition - pending to in-progress",
			fromState:     types.TaskStatePending,
			toState:       types.TaskStateInProgress,
			task:          task,
			expectError:   false,
		},
		{
			name:          "Valid state transition - completed to pending (reopen)",
			fromState:     types.TaskStateCompleted,
			toState:       types.TaskStatePending,
			task:          task,
			expectError:   false,
		},
		{
			name:          "Valid state transition - in-progress to completed",
			fromState:     types.TaskStateInProgress,
			toState:       types.TaskStateCompleted,
			task:          task,
			expectError:   false,
		},
		{
			name:          "Valid state transition - any to cancelled",
			fromState:     types.TaskStateInProgress,
			toState:       types.TaskStateCancelled,
			task:          task,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Update task state for the test
			testTask := *tt.task
			testTask.State = tt.fromState
			
			err := validator.ValidateTransition(tt.fromState, tt.toState, &testTask)
			
			if tt.expectError {
				require.Error(t, err, "Expected error for transition %s -> %s", tt.fromState, tt.toState)
				
				// Check if it's an EnhancedError (clean error message)
				if tt.errorType == "*errors.EnhancedError" {
					enhancedErr, ok := err.(*errors.EnhancedError)
					require.True(t, ok, "Expected EnhancedError, got %T", err)
					
					// Verify error contains expected strings
					for _, contains := range tt.errorContains {
						assert.Contains(t, enhancedErr.Error(), contains, 
							"Error message should contain '%s'", contains)
					}
					
					// Verify EnhancedError has helpful fields
					assert.NotEmpty(t, enhancedErr.Operation, "EnhancedError should have Operation")
					assert.NotEmpty(t, enhancedErr.Suggestion, "EnhancedError should have Suggestion")
					assert.NotEmpty(t, enhancedErr.Example, "EnhancedError should have Example")
				}
			} else {
				assert.NoError(t, err, "Expected no error for transition %s -> %s", tt.fromState, tt.toState)
			}
		})
	}
}

// TestValidationErrorMessages tests that error messages are user-friendly
func TestValidationErrorMessages(t *testing.T) {
	validator := NewStateValidator()
	
	task := &types.Task{
		ID:          uuid.New(),
		ProjectID:   uuid.New(),
		Title:       "Test Task",
		State:       types.TaskStatePending,
		Complexity:  5,
	}

	// Test invalid transition
	err := validator.ValidateTransition(types.TaskStatePending, types.TaskStateCompleted, task)
	require.Error(t, err)
	
	enhancedErr, ok := err.(*errors.EnhancedError)
	require.True(t, ok, "Should return EnhancedError")
	
	// Test that error message is clean and helpful
	errorMsg := enhancedErr.Error()
	
	// Should contain key information
	assert.Contains(t, errorMsg, "transition", "Error should mention transition")
	assert.Contains(t, errorMsg, "pending", "Error should mention current state")
	assert.Contains(t, errorMsg, "completed", "Error should mention target state")
	
	// Should have helpful suggestion
	assert.NotEmpty(t, enhancedErr.Suggestion, "Should provide suggestion")
	assert.Contains(t, enhancedErr.Suggestion, "in-progress", "Should suggest intermediate state")
	
	// Should have example command
	assert.NotEmpty(t, enhancedErr.Example, "Should provide example")
	assert.Contains(t, enhancedErr.Example, "knot task update-state", "Should show correct command")
	assert.Contains(t, enhancedErr.Example, task.ID.String(), "Should include task ID")
}

// TestStateValidatorBasics tests basic validator functionality
func TestStateValidatorBasics(t *testing.T) {
	validator := NewStateValidator()
	
	// Test valid states
	validStates := []string{"pending", "in-progress", "completed", "blocked", "cancelled"}
	for _, state := range validStates {
		assert.True(t, validator.IsValidState(state), "State %s should be valid", state)
	}
	
	// Test invalid states
	invalidStates := []string{"invalid", "unknown", "", "PENDING", "InProgress"}
	for _, state := range invalidStates {
		assert.False(t, validator.IsValidState(state), "State %s should be invalid", state)
	}
	
	// Test transition matrix
	matrix := validator.GetStateTransitionMatrix()
	assert.NotEmpty(t, matrix, "Transition matrix should not be empty")
	
	// Test specific transitions
	pendingTransitions := matrix["pending"]
	assert.Contains(t, pendingTransitions, "in-progress", "Pending should allow transition to in-progress")
	assert.Contains(t, pendingTransitions, "blocked", "Pending should allow transition to blocked")
	assert.Contains(t, pendingTransitions, "cancelled", "Pending should allow transition to cancelled")
	assert.NotContains(t, pendingTransitions, "completed", "Pending should not allow direct transition to completed")
}

// TestLenientValidation tests the lenient validation mode
func TestLenientValidation(t *testing.T) {
	validator := NewStateValidator()
	
	task := &types.Task{
		ID:          uuid.New(),
		ProjectID:   uuid.New(),
		Title:       "Test Task",
		State:       types.TaskStatePending,
		Complexity:  5,
	}

	// Test lenient validation for pending -> completed (should warn, not error)
	err, warnings := validator.ValidateTransitionLenient(types.TaskStatePending, types.TaskStateCompleted, task)
	
	// Should not error in lenient mode (transition is not allowed, so it should error)
	assert.Error(t, err, "Lenient validation should error for invalid transition pending -> completed")
	
	// Test a valid transition that triggers warnings (high complexity completion)
	highComplexityTask := &types.Task{
		ID:          uuid.New(),
		ProjectID:   uuid.New(),
		Title:       "High Complexity Task",
		State:       types.TaskStateInProgress,
		Complexity:  9, // High complexity
	}
	
	err, warnings = validator.ValidateTransitionLenient(types.TaskStateInProgress, types.TaskStateCompleted, highComplexityTask)
	
	// Should not error but provide warnings
	assert.NoError(t, err, "Lenient validation should not error for valid transition")
	assert.NotEmpty(t, warnings, "Lenient validation should provide warnings for high complexity")
	assert.Contains(t, warnings[0], "high_complexity", "Warning should mention high complexity")
}