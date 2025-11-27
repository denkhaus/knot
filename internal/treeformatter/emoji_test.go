package treeformatter

import (
	"testing"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
)

// TestStateToEmoji tests the emoji mapping for all task states
func TestStateToEmoji(t *testing.T) {
	formatter := NewDefaultFormatter()

	tests := []struct {
		name     string
		state    types.TaskState
		expected string
	}{
		{
			name:     "completed state",
			state:    types.TaskStateCompleted,
			expected: "✅ ",
		},
		{
			name:     "in-progress state",
			state:    types.TaskStateInProgress,
			expected: "🔄 ",
		},
		{
			name:     "pending state",
			state:    types.TaskStatePending,
			expected: "⏳ ",
		},
		{
			name:     "blocked state",
			state:    types.TaskStateBlocked,
			expected: "🚫 ",
		},
		{
			name:     "cancelled state",
			state:    types.TaskStateCancelled,
			expected: "❌ ",
		},
		{
			name:     "deletion-pending state",
			state:    types.TaskStateDeletionPending,
			expected: "🗑️ ",
		},
		{
			name:     "unknown state",
			state:    types.TaskState("unknown"),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.StateToEmoji(tt.state)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestStateToEmojiDisabled tests that emojis can be disabled
func TestStateToEmojiDisabled(t *testing.T) {
	formatter := NewDefaultFormatter()
	formatter.config.ShowEmojis = false

	task := &types.Task{
		ID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		Title: "Test Task",
		State: types.TaskStateCompleted,
	}

	result := formatter.FormatTaskLine(task)
	expected := "Test Task (ID: 550e8400-e29b-41d4-a716-446655440000) - completed"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestStateToEmojiWithFallback tests fallback behavior for unsupported states
func TestStateToEmojiWithFallback(t *testing.T) {
	formatter := NewDefaultFormatter()
	formatter.config.ShowEmojis = true

	// Test with a custom state that's not in the enum
	task := &types.Task{
		ID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		Title: "Custom State Task",
		State: types.TaskState("custom-state"),
	}

	result := formatter.FormatTaskLine(task)
	expected := "Custom State Task (ID: 550e8400-e29b-41d4-a716-446655440000) - custom-state"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
