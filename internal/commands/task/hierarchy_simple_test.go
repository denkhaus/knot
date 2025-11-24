package task

import (
	"testing"

	"github.com/denkhaus/knot/v2/internal/treeformatter"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
)

// TestTaskFormatterIntegration tests that the treeformatter integration works correctly
func TestTaskFormatterIntegration(t *testing.T) {
	// Create tree formatter with emoji support
	formatter := treeformatter.NewFormatter(&treeformatter.Config{
		ShowEmojis:   true,
		CompactMode:  false,
		IndentSize:   2,
	})

	// Test task from the brain memory specification
	task := &types.Task{
		ID:    uuid.MustParse("0a4f4544-2802-493b-9986-e0d7b7a41d2e"),
		Title: "Core Architecture Implementation",
		State: types.TaskStateInProgress,
	}

	// Format the task
	result := formatter.FormatTaskLine(task)

	// Verify the result has the expected format with emoji
	expected := "🔄 Core Architecture Implementation (ID: 0a4f4544-2802-493b-9986-e0d7b7a41d2e) - in-progress"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestTaskFormatterDifferentStates tests all task states with emojis
func TestTaskFormatterDifferentStates(t *testing.T) {
	formatter := treeformatter.NewFormatter(&treeformatter.Config{
		ShowEmojis:   true,
		CompactMode:  false,
		IndentSize:   2,
	})

	tests := []struct {
		state    types.TaskState
		expected string
	}{
		{types.TaskStatePending, "⏳ Test Task (ID: test-id) - pending"},
		{types.TaskStateInProgress, "🔄 Test Task (ID: test-id) - in-progress"},
		{types.TaskStateCompleted, "✅ Test Task (ID: test-id) - completed"},
		{types.TaskStateBlocked, "🚫 Test Task (ID: test-id) - blocked"},
		{types.TaskStateCancelled, "❌ Test Task (ID: test-id) - cancelled"},
		{types.TaskStateDeletionPending, "🗑️ Test Task (ID: test-id) - deletion-pending"},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			task := &types.Task{
				ID:    uuid.MustParse("00000000-0000-0000-0000-000000000000"),
				Title: "Test Task",
				State: tt.state,
			}

			result := formatter.FormatTaskLine(task)

			// Check that it starts with the expected emoji and contains the state
			if len(result) < 2 || result[:2] != tt.expected[:2] {
				t.Errorf("Expected emoji %s for state %s, got %q", tt.expected[:2], tt.state, result)
			}

			if !containsString(result, string(tt.state)) {
				t.Errorf("Expected result to contain state %s, got %q", tt.state, result)
			}
		})
	}
}

// TestTaskFormatterWithoutEmojis tests that emojis can be disabled
func TestTaskFormatterWithoutEmojis(t *testing.T) {
	formatter := treeformatter.NewFormatter(&treeformatter.Config{
		ShowEmojis:   false,
		CompactMode:  false,
		IndentSize:   2,
	})

	task := &types.Task{
		ID:    uuid.MustParse("0a4f4544-2802-493b-9986-e0d7b7a41d2e"),
		Title: "Core Architecture Implementation",
		State: types.TaskStateInProgress,
	}

	result := formatter.FormatTaskLine(task)
	expected := "Core Architecture Implementation (ID: 0a4f4544-2802-493b-9986-e0d7b7a41d2e) - in-progress"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// containsString is a helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}()))
}