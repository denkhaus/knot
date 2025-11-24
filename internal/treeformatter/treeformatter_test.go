package treeformatter

import (
	"testing"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
)

// TestTreeFormatterInterface validates that TreeFormatter implements the expected interface
func TestTreeFormatterInterface(t *testing.T) {
	var _ TreeFormatter = &DefaultTreeFormatter{}
}

// TestFormatTaskLine tests the basic task line formatting
func TestFormatTaskLine(t *testing.T) {
	formatter := NewDefaultFormatter()

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

// TestFormatTaskLineWithEmoji tests task line formatting with state emojis
func TestFormatTaskLineWithEmoji(t *testing.T) {
	formatter := NewDefaultFormatter()
	formatter.config.ShowEmojis = true

	tests := []struct {
		name     string
		task     *types.Task
		expected string
	}{
		{
			name: "completed task",
			task: &types.Task{
				ID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
				Title: "Completed Task",
				State: types.TaskStateCompleted,
			},
			expected: "✅ Completed Task (ID: 550e8400-e29b-41d4-a716-446655440001) - completed",
		},
		{
			name: "in-progress task",
			task: &types.Task{
				ID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
				Title: "In Progress Task",
				State: types.TaskStateInProgress,
			},
			expected: "🔄 In Progress Task (ID: 550e8400-e29b-41d4-a716-446655440002) - in-progress",
		},
		{
			name: "pending task",
			task: &types.Task{
				ID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
				Title: "Pending Task",
				State: types.TaskStatePending,
			},
			expected: "⏳ Pending Task (ID: 550e8400-e29b-41d4-a716-446655440003) - pending",
		},
		{
			name: "blocked task",
			task: &types.Task{
				ID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"),
				Title: "Blocked Task",
				State: types.TaskStateBlocked,
			},
			expected: "🚫 Blocked Task (ID: 550e8400-e29b-41d4-a716-446655440004) - blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatTaskLine(tt.task)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestGetTreePrefix tests the generation of tree prefixes
func TestGetTreePrefix(t *testing.T) {
	formatter := NewDefaultFormatter()

	tests := []struct {
		name           string
		isLast         bool
		depth          int
		parentPrefix   string
		expectedPrefix string
	}{
		{
			name:           "root task",
			isLast:         true,
			depth:          0,
			parentPrefix:   "",
			expectedPrefix: "",
		},
		{
			name:           "last child",
			isLast:         true,
			depth:          1,
			parentPrefix:   "",
			expectedPrefix: "└── ",
		},
		{
			name:           "middle child",
			isLast:         false,
			depth:          1,
			parentPrefix:   "",
			expectedPrefix: "├── ",
		},
		{
			name:           "nested last child",
			isLast:         true,
			depth:          2,
			parentPrefix:   "│  ",
			expectedPrefix: "│  └── ",
		},
		{
			name:           "nested middle child",
			isLast:         false,
			depth:          2,
			parentPrefix:   "│  ",
			expectedPrefix: "│  ├── ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.GetTreePrefix(tt.isLast, tt.depth, tt.parentPrefix)
			if result != tt.expectedPrefix {
				t.Errorf("Expected %q, got %q", tt.expectedPrefix, result)
			}
		})
	}
}