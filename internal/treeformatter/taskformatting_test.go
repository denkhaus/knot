package treeformatter

import (
	"testing"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
)

// TestFormatTaskLineCompact tests compact mode formatting
func TestFormatTaskLineCompact(t *testing.T) {
	config := DefaultConfig()
	config.CompactMode = true
	formatter := NewFormatter(config)

	task := &types.Task{
		ID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		Title: "Test Task",
		State: types.TaskStatePending,
	}

	result := formatter.FormatTaskLine(task)
	expected := "Test Task (ID: 550e8400-e29b-41d4-a716-446655440000) - pending"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestFormatTaskLineWithAllStates tests formatting with all possible states
func TestFormatTaskLineWithAllStates(t *testing.T) {
	formatter := NewDefaultFormatter()
	formatter.config.ShowEmojis = true

	taskID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	baseTitle := "Test Task"

	tests := []struct {
		name     string
		state    types.TaskState
		expected string
	}{
		{
			name:     "pending",
			state:    types.TaskStatePending,
			expected: "⏳ Test Task (ID: 550e8400-e29b-41d4-a716-446655440000) - pending",
		},
		{
			name:     "in-progress",
			state:    types.TaskStateInProgress,
			expected: "🔄 Test Task (ID: 550e8400-e29b-41d4-a716-446655440000) - in-progress",
		},
		{
			name:     "completed",
			state:    types.TaskStateCompleted,
			expected: "✅ Test Task (ID: 550e8400-e29b-41d4-a716-446655440000) - completed",
		},
		{
			name:     "blocked",
			state:    types.TaskStateBlocked,
			expected: "🚫 Test Task (ID: 550e8400-e29b-41d4-a716-446655440000) - blocked",
		},
		{
			name:     "cancelled",
			state:    types.TaskStateCancelled,
			expected: "❌ Test Task (ID: 550e8400-e29b-41d4-a716-446655440000) - cancelled",
		},
		{
			name:     "deletion-pending",
			state:    types.TaskStateDeletionPending,
			expected: "🗑️ Test Task (ID: 550e8400-e29b-41d4-a716-446655440000) - deletion-pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &types.Task{
				ID:    taskID,
				Title: baseTitle,
				State: tt.state,
			}

			result := formatter.FormatTaskLine(task)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestFormatTaskLineWithLongTitle tests formatting with long titles
func TestFormatTaskLineWithLongTitle(t *testing.T) {
	formatter := NewDefaultFormatter()

	task := &types.Task{
		ID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		Title: "This is a very long task title that should be formatted correctly without any issues",
		State: types.TaskStateInProgress,
	}

	result := formatter.FormatTaskLine(task)
	expected := "This is a very long task title that should be formatted correctly without any issues (ID: 550e8400-e29b-41d4-a716-446655440000) - in-progress"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestFormatTaskLineWithEmojiDisabled tests that emojis can be disabled
func TestFormatTaskLineWithEmojiDisabled(t *testing.T) {
	formatter := NewDefaultFormatter()
	formatter.config.ShowEmojis = false

	task := &types.Task{
		ID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		Title: "Completed Task",
		State: types.TaskStateCompleted,
	}

	result := formatter.FormatTaskLine(task)
	expected := "Completed Task (ID: 550e8400-e29b-41d4-a716-446655440000) - completed"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestFormatTaskLineWithSpecialCharacters tests formatting with special characters
func TestFormatTaskLineWithSpecialCharacters(t *testing.T) {
	formatter := NewDefaultFormatter()

	task := &types.Task{
		ID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		Title: "Task with special chars: 'quotes', \"double quotes\", & symbols",
		State: types.TaskStatePending,
	}

	result := formatter.FormatTaskLine(task)
	expected := "Task with special chars: 'quotes', \"double quotes\", & symbols (ID: 550e8400-e29b-41d4-a716-446655440000) - pending"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestDifferentConfigurations tests various configuration combinations
func TestDifferentConfigurations(t *testing.T) {
	taskID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	task := &types.Task{
		ID:    taskID,
		Title: "Config Test Task",
		State: types.TaskStateCompleted,
	}

	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name:     "default config",
			config:   DefaultConfig(),
			expected: "Config Test Task (ID: 550e8400-e29b-41d4-a716-446655440000) - completed",
		},
		{
			name: "emojis enabled",
			config: &Config{
				ShowEmojis:   true,
				CompactMode:  false,
				IndentSize:   2,
			},
			expected: "✅ Config Test Task (ID: 550e8400-e29b-41d4-a716-446655440000) - completed",
		},
		{
			name: "compact with emojis",
			config: &Config{
				ShowEmojis:   true,
				CompactMode:  true,
				IndentSize:   2,
			},
			expected: "✅ Config Test Task (ID: 550e8400-e29b-41d4-a716-446655440000) - completed",
		},
		{
			name: "custom indent size",
			config: &Config{
				ShowEmojis:   false,
				CompactMode:  false,
				IndentSize:   4,
			},
			expected: "Config Test Task (ID: 550e8400-e29b-41d4-a716-446655440000) - completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(tt.config)
			result := formatter.FormatTaskLine(task)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}