package validation

import (
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/errors"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeletionPendingStateValidation tests the deletion-pending state validation rules
func TestDeletionPendingStateValidation(t *testing.T) {
	validator := NewStateValidator()

	// Create a test task in deletion-pending state
	task := &types.Task{
		ID:          uuid.New(),
		ProjectID:   uuid.New(),
		Title:       "Task to be deleted",
		Description: "Task marked for deletion",
		State:       types.TaskStateDeletionPending,
		Complexity:  5,
		Depth:       0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name          string
		fromState     types.TaskState
		toState       types.TaskState
		expectError   bool
		errorContains []string
	}{
		{
			name:          "deletion-pending to pending - should fail",
			fromState:     types.TaskStateDeletionPending,
			toState:       types.TaskStatePending,
			expectError:   true,
			errorContains: []string{"invalid state transition", "deletion-pending"},
		},
		{
			name:          "deletion-pending to in-progress - should fail",
			fromState:     types.TaskStateDeletionPending,
			toState:       types.TaskStateInProgress,
			expectError:   true,
			errorContains: []string{"invalid state transition", "deletion-pending"},
		},
		{
			name:          "deletion-pending to completed - should fail",
			fromState:     types.TaskStateDeletionPending,
			toState:       types.TaskStateCompleted,
			expectError:   true,
			errorContains: []string{"invalid state transition", "deletion-pending"},
		},
		{
			name:          "deletion-pending to blocked - should fail",
			fromState:     types.TaskStateDeletionPending,
			toState:       types.TaskStateBlocked,
			expectError:   true,
			errorContains: []string{"invalid state transition", "deletion-pending"},
		},
		{
			name:          "deletion-pending to cancelled - should fail",
			fromState:     types.TaskStateDeletionPending,
			toState:       types.TaskStateCancelled,
			expectError:   true,
			errorContains: []string{"invalid state transition", "deletion-pending"},
		},
		{
			name:        "deletion-pending to deletion-pending - should succeed (no-op)",
			fromState:   types.TaskStateDeletionPending,
			toState:     types.TaskStateDeletionPending,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Update task state for the test
			testTask := *task
			testTask.State = tt.fromState

			err := validator.ValidateTransition(tt.fromState, tt.toState, &testTask)

			if tt.expectError {
				require.Error(t, err, "Expected error for transition %s -> %s", tt.fromState, tt.toState)

				// Check if it's an EnhancedError with helpful message
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
			} else {
				assert.NoError(t, err, "Expected no error for transition %s -> %s", tt.fromState, tt.toState)
			}
		})
	}
}

// TestTransitionsToDeletionPending tests transitions TO deletion-pending state
func TestTransitionsToDeletionPending(t *testing.T) {
	validator := NewStateValidator()

	task := &types.Task{
		ID:          uuid.New(),
		ProjectID:   uuid.New(),
		Title:       "Test Task",
		Description: "Task for testing transitions to deletion-pending",
		Complexity:  3,
		Depth:       0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	validTransitions := []struct {
		name      string
		fromState types.TaskState
	}{
		{"pending to deletion-pending", types.TaskStatePending},
		{"in-progress to deletion-pending", types.TaskStateInProgress},
		{"completed to deletion-pending", types.TaskStateCompleted},
		{"blocked to deletion-pending", types.TaskStateBlocked},
		{"cancelled to deletion-pending", types.TaskStateCancelled},
	}

	for _, tt := range validTransitions {
		t.Run(tt.name, func(t *testing.T) {
			testTask := *task
			testTask.State = tt.fromState

			err := validator.ValidateTransition(tt.fromState, types.TaskStateDeletionPending, &testTask)
			assert.NoError(t, err, "Transition %s -> deletion-pending should be valid", tt.fromState)
		})
	}
}

// TestDeletionPendingValidationRule tests the specific validation rule
func TestDeletionPendingValidationRule(t *testing.T) {
	validator := NewStateValidator()

	task := &types.Task{
		ID:         uuid.New(),
		ProjectID:  uuid.New(),
		Title:      "Deletion Pending Task",
		State:      types.TaskStateDeletionPending,
		Complexity: 4,
	}

	t.Run("deletion_pending_protection rule", func(t *testing.T) {
		// Test that transitions from deletion-pending are blocked
		err := validator.ValidateTransition(types.TaskStateDeletionPending, types.TaskStatePending, task)
		require.Error(t, err)

		enhancedErr, ok := err.(*errors.EnhancedError)
		require.True(t, ok, "Should return EnhancedError")

		// Check that it's blocked at the transition matrix level
		assert.Contains(t, enhancedErr.Error(), "invalid state transition", "Should mention invalid transition")
		assert.Contains(t, enhancedErr.Error(), "deletion-pending", "Should mention deletion-pending state")
		assert.Contains(t, enhancedErr.Example, "knot task delete", "Should show delete command example")
		assert.Contains(t, enhancedErr.Example, task.ID.String(), "Should include task ID in example")
	})

	t.Run("self-transition allowed", func(t *testing.T) {
		// Test that self-transition (no-op) is allowed
		err := validator.ValidateTransition(types.TaskStateDeletionPending, types.TaskStateDeletionPending, task)
		assert.NoError(t, err, "Self-transition should be allowed")
	})
}

// TestDeletionPendingStateMatrix tests the complete state transition matrix
func TestDeletionPendingStateMatrix(t *testing.T) {
	validator := NewStateValidator()

	matrix := validator.GetStateTransitionMatrix()

	t.Run("deletion-pending has no outgoing transitions", func(t *testing.T) {
		deletionPendingTransitions := matrix["deletion-pending"]

		// Should only allow self-transition (which is filtered out in getValidTransitionsFrom)
		assert.Len(t, deletionPendingTransitions, 0,
			"deletion-pending should have no valid outgoing transitions (except self)")
	})

	t.Run("all states can transition to deletion-pending", func(t *testing.T) {
		allStates := []string{"pending", "in-progress", "completed", "blocked", "cancelled"}

		for _, state := range allStates {
			transitions := matrix[state]
			assert.Contains(t, transitions, "deletion-pending",
				"State %s should allow transition to deletion-pending", state)
		}
	})
}

// TestDeletionPendingEdgeCases tests edge cases for deletion-pending state
func TestDeletionPendingEdgeCases(t *testing.T) {
	validator := NewStateValidator()

	t.Run("task with dependencies in deletion-pending", func(t *testing.T) {
		// Task with dependencies should still be protected from transitions
		taskWithDeps := &types.Task{
			ID:           uuid.New(),
			ProjectID:    uuid.New(),
			Title:        "Task with dependencies",
			State:        types.TaskStateDeletionPending,
			Dependencies: []uuid.UUID{uuid.New(), uuid.New()},
			Complexity:   5,
		}

		// Should still prevent transitions even with dependencies
		err := validator.ValidateTransition(types.TaskStateDeletionPending, types.TaskStatePending, taskWithDeps)
		require.Error(t, err, "Should prevent transition even for tasks with dependencies")

		enhancedErr, ok := err.(*errors.EnhancedError)
		require.True(t, ok, "Should return EnhancedError")
		assert.Contains(t, enhancedErr.Error(), "invalid state transition")
	})

	t.Run("high complexity task in deletion-pending", func(t *testing.T) {
		// High complexity task should still be protected
		highComplexityTask := &types.Task{
			ID:         uuid.New(),
			ProjectID:  uuid.New(),
			Title:      "High complexity task",
			State:      types.TaskStateDeletionPending,
			Complexity: 10, // Maximum complexity
		}

		err := validator.ValidateTransition(types.TaskStateDeletionPending, types.TaskStateCompleted, highComplexityTask)
		require.Error(t, err, "Should prevent transition for high complexity tasks")
	})

	t.Run("completed task moved to deletion-pending", func(t *testing.T) {
		// Test that completed tasks can be moved to deletion-pending
		completedTask := &types.Task{
			ID:          uuid.New(),
			ProjectID:   uuid.New(),
			Title:       "Completed task",
			State:       types.TaskStateCompleted,
			Complexity:  3,
			CompletedAt: &time.Time{},
		}

		err := validator.ValidateTransition(types.TaskStateCompleted, types.TaskStateDeletionPending, completedTask)
		assert.NoError(t, err, "Completed tasks should be allowed to transition to deletion-pending")
	})
}

// TestDeletionPendingLenientValidation tests lenient validation for deletion-pending
func TestDeletionPendingLenientValidation(t *testing.T) {
	validator := NewStateValidator()

	task := &types.Task{
		ID:         uuid.New(),
		ProjectID:  uuid.New(),
		Title:      "Test Task",
		State:      types.TaskStateDeletionPending,
		Complexity: 5,
	}

	t.Run("lenient validation still prevents invalid transitions", func(t *testing.T) {
		// Even in lenient mode, deletion-pending protection should be enforced
		err, warnings := validator.ValidateTransitionLenient(types.TaskStateDeletionPending, types.TaskStatePending, task)

		// Should still error (not just warn) because this is a security/data integrity rule
		assert.Error(t, err, "Lenient validation should still prevent deletion-pending transitions")
		assert.Empty(t, warnings, "Should error, not warn, for deletion-pending violations")

		enhancedErr, ok := err.(*errors.EnhancedError)
		require.True(t, ok, "Should return EnhancedError even in lenient mode")
		assert.Contains(t, enhancedErr.Error(), "invalid state transition")
	})

	t.Run("lenient validation allows valid transitions to deletion-pending", func(t *testing.T) {
		// Valid transitions should work in lenient mode
		pendingTask := &types.Task{
			ID:         uuid.New(),
			ProjectID:  uuid.New(),
			Title:      "Pending Task",
			State:      types.TaskStatePending,
			Complexity: 5,
		}

		err, warnings := validator.ValidateTransitionLenient(types.TaskStatePending, types.TaskStateDeletionPending, pendingTask)
		assert.NoError(t, err, "Valid transition to deletion-pending should work in lenient mode")
		assert.Empty(t, warnings, "Should not generate warnings for valid transitions")
	})
}
