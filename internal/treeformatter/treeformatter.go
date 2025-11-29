// Package treeformatter provides utilities for formatting hierarchical task displays
// with consistent styling, emojis, and tree structure characters.
//
// This package follows the single responsibility principle by handling only
// the visual formatting of task trees, without any business logic.
//
// Brain Memory Reference: e4bea247-7f1f-4712-8188-b9b0b4ecb3ea
// Task Reference: 77540a50-676f-402e-8a45-d6272c51b7b2
package treeformatter

import (
	"fmt"
	"strings"

	"github.com/denkhaus/knot/v2/internal/types"
)

// TreeFormatter defines the interface for formatting task trees
type TreeFormatter interface {
	// FormatTaskLine formats a single task line with optional emoji
	FormatTaskLine(task *types.Task) string

	// GetTreePrefix returns the appropriate tree prefix based on position
	GetTreePrefix(isLast bool, depth int, parentPrefix string) string
}

// Config holds configuration options for tree formatting
type Config struct {
	// ShowEmojis enables state emoji indicators
	ShowEmojis bool

	// CompactMode removes extra spacing for tighter display
	CompactMode bool

	// IndentSize controls the number of spaces per indentation level
	IndentSize int
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		ShowEmojis:  false,
		CompactMode: false,
		IndentSize:  2,
	}
}

// DefaultTreeFormatter implements TreeFormatter with standard formatting
type DefaultTreeFormatter struct {
	config *Config
}

// NewDefaultFormatter creates a new DefaultTreeFormatter with default config
func NewDefaultFormatter() *DefaultTreeFormatter {
	return &DefaultTreeFormatter{
		config: DefaultConfig(),
	}
}

// NewFormatter creates a new DefaultTreeFormatter with custom config
func NewFormatter(config *Config) *DefaultTreeFormatter {
	if config == nil {
		config = DefaultConfig()
	}
	return &DefaultTreeFormatter{
		config: config,
	}
}

// StateToEmoji returns the appropriate emoji for each task state
func (f *DefaultTreeFormatter) StateToEmoji(state types.TaskState) string {
	switch state {
	case types.TaskStateCompleted:
		return "✅ "
	case types.TaskStateInProgress:
		return "🔄 "
	case types.TaskStatePending:
		return "⏳ "
	case types.TaskStateBlocked:
		return "🚫 "
	case types.TaskStateCancelled:
		return "❌ "
	case types.TaskStateDeletionPending:
		return "🗑️ "
	default:
		return ""
	}
}

// FormatTaskLine formats a single task line with optional emoji
func (f *DefaultTreeFormatter) FormatTaskLine(task *types.Task) string {
	var emoji string
	if f.config.ShowEmojis {
		emoji = f.StateToEmoji(task.State)
	}

	return fmt.Sprintf("%s%s (ID: %s) - %s", emoji, task.Title, task.ID, task.State)
}

// GetTreePrefix returns the appropriate tree prefix based on position
func (f *DefaultTreeFormatter) GetTreePrefix(isLast bool, depth int, parentPrefix string) string {
	// Root level (depth 0) has no prefix
	if depth == 0 {
		return ""
	}

	// For child tasks, determine the appropriate prefix
	if isLast {
		return parentPrefix + "└── "
	}
	return parentPrefix + "├── "
}

// GetParentPrefix returns the prefix for calculating child prefixes
func (f *DefaultTreeFormatter) GetParentPrefix(prefix string, _isLast bool) string {
	if strings.HasSuffix(prefix, "└── ") {
		return strings.TrimSuffix(prefix, "└── ") + "   "
	}
	if strings.HasSuffix(prefix, "├── ") {
		return strings.TrimSuffix(prefix, "├── ") + "│  "
	}
	return prefix
}
