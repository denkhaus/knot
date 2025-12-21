package hints

import (
	"context"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/session"
)

// Generator defines the interface for generating context-aware hints
type Generator interface {
	// GenerateHints generates hints based on tool execution and session state
	GenerateHints(ctx context.Context, toolName string, result interface{}, sessionState *SessionState) []Hint

	// GenerateNoProjectSelectedHints generates hints when no project is selected
	GenerateNoProjectSelectedHints() []Hint

	// GenerateGeneralHints generates general navigation and discovery hints
	GenerateGeneralHints() []Hint

	// FilterHints filters hints by enabled categories
	FilterHints(hints []Hint) []Hint
}

// Integration defines the interface for integrating hints with MCP tools
type Integration interface {
	// GetSessionContext extracts session context information
	GetSessionContext(ctx context.Context, projectManager manager.ProjectManager, sessionManager session.SessionManager) (*SessionState, error)

	// GenerateToolHints generates hints for specific tool execution
	GenerateToolHints(ctx context.Context, toolName string, result interface{}, projectManager manager.ProjectManager, sessionManager session.SessionManager) []Hint

	// GenerateContextualHints generates hints based on current session state
	GenerateContextualHints(ctx context.Context, projectManager manager.ProjectManager, sessionManager session.SessionManager) []Hint

	// WrapToolResponse wraps a tool response with generated hints
	WrapToolResponse(response interface{}, hints []Hint) interface{}

	// FormatHintsForOutput formats hints for MCP protocol output
	FormatHintsForOutput(hints []Hint) []interface{}

	// GetHintStats provides statistics about hint generation
	GetHintStats() map[string]interface{}
}
