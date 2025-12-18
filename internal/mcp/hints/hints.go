package hints

import (
	"context"
)

// HintIntegration provides the main interface for hint integration with MCP tools
type HintIntegration struct {
	generator *HintGenerator
}

// NewHintIntegration creates a new hint integration service
func NewHintIntegration(enabled bool, maxHints int, categories []string) *HintIntegration {
	return &HintIntegration{
		generator: NewHintGenerator(enabled, maxHints, categories),
	}
}

// GetSessionContext extracts session context information from a context
func (hi *HintIntegration) GetSessionContext(ctx context.Context, projectManager interface{}, sessionManager interface{}) (*SessionState, error) {
	// This would need to be implemented based on the actual session management interface
	// For now, return a basic session state
	return &SessionState{
		Actor:       "mcp-user",
		TaskCount:   0,
		RecentTools: []string{},
	}, nil
}

// GenerateToolHints generates hints for a specific tool execution
func (hi *HintIntegration) GenerateToolHints(ctx context.Context, toolName string, result interface{}, projectManager interface{}, sessionManager interface{}) []Hint {
	sessionState, err := hi.GetSessionContext(ctx, projectManager, sessionManager)
	if err != nil {
		// Fallback to basic session state
		sessionState = &SessionState{
			Actor:       "mcp-user",
			TaskCount:   0,
			RecentTools: []string{},
		}
	}

	hints := hi.generator.GenerateHints(ctx, toolName, result, sessionState)
	return hi.generator.FilterHints(hints)
}

// GenerateContextualHints generates hints based on current session state
func (hi *HintIntegration) GenerateContextualHints(ctx context.Context, projectManager interface{}, sessionManager interface{}) []Hint {
	sessionState, err := hi.GetSessionContext(ctx, projectManager, sessionManager)
	if err != nil {
		return hi.generator.GenerateGeneralHints()
	}

	var hints []Hint

	// Check if no project is selected
	if sessionState.ProjectID == nil || *sessionState.ProjectID == "" {
		hints = append(hints, hi.generator.GenerateNoProjectSelectedHints()...)
	} else if sessionState.TaskCount == 0 {
		// Project exists but no tasks
		hints = append(hints, Hint{
			Type:        string(HintTypeNextAction),
			Title:       "Project Ready - No Tasks Yet",
			Description: "Create your first task to get started with this project",
			NextTools:   []string{"task_create"},
			Context:     map[string]interface{}{"project_empty": true},
		})
	}

	// Add general hints if enabled
	if hi.generator.enabled {
		generalHints := hi.generator.GenerateGeneralHints()
		hints = append(hints, generalHints...)
	}

	return hi.generator.FilterHints(hints)
}

// WrapToolResponse wraps a tool response with generated hints
func (hi *HintIntegration) WrapToolResponse(response interface{}, hints []Hint) interface{} {
	if len(hints) == 0 {
		return response
	}

	// Create a response wrapper that includes hints
	wrapper := map[string]interface{}{
		"response": response,
		"hints":    hints,
	}

	return wrapper
}

// FormatHintsForOutput formats hints for MCP protocol output
func (hi *HintIntegration) FormatHintsForOutput(hints []Hint) []interface{} {
	var formatted []interface{}
	for _, hint := range hints {
		hintMap := map[string]interface{}{
			"type":        hint.Type,
			"title":       hint.Title,
			"description": hint.Description,
		}
		if len(hint.NextTools) > 0 {
			hintMap["next_tools"] = hint.NextTools
		}
		if len(hint.Context) > 0 {
			hintMap["context"] = hint.Context
		}
		formatted = append(formatted, hintMap)
	}
	return formatted
}

// GetHintStats provides statistics about hint generation
func (hi *HintIntegration) GetHintStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":    hi.generator.enabled,
		"max_hints":  hi.generator.maxHints,
		"categories": hi.generator.categories,
	}
}
