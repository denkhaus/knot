package hints

import (
	"context"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/shared"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/mark3labs/mcp-go/server"
	"github.com/samber/do/v2"
)

// Private implementation of the Generator interface
type hintGenerator struct {
	enabled    bool
	maxHints   int
	categories []string
	logger     logger.Logger
}

// NewHintGeneratorImpl creates a new hint generator with DI (private implementation)
func NewHintGeneratorImpl(cfg *config.MCPConfig, logger logger.Logger) Generator {
	return &hintGenerator{
		enabled:    cfg.Hints.Enabled,
		maxHints:   cfg.Hints.MaxHints,
		categories: cfg.Hints.Categories,
		logger:     logger,
	}
}

// GenerateHints generates hints based on tool execution and session state
func (hg *hintGenerator) GenerateHints(ctx context.Context, toolName string, result interface{}, sessionState *SessionState) []Hint {
	if !hg.enabled {
		return nil
	}

	var hints []Hint

	// Generate hints based on the specific tool executed
	switch toolName {
	case "project_create":
		hints = hg.appendProjectCreatedHints(hints, result, sessionState)
	case "project_select":
		hints = hg.appendProjectSelectedHints(hints, result, sessionState)
	case "task_create":
		hints = hg.appendTaskCreatedHints(hints, result, sessionState)
	case "task_update", "task_update_state":
		hints = hg.appendTaskUpdatedHints(hints, result, sessionState)
	case "task_delete":
		hints = hg.appendTaskDeletedHints(hints, result, sessionState)
	case "project_list":
		hints = hg.appendProjectListHints(hints, result, sessionState)
	}

	// Apply limit
	if len(hints) > hg.maxHints {
		hints = hints[:hg.maxHints]
	}

	hg.logger.Debug("Generated hints",
		logger.String("tool", toolName),
		logger.Int("count", len(hints)))

	return hints
}

// GenerateNoProjectSelectedHints generates hints when no project is selected
func (hg *hintGenerator) GenerateNoProjectSelectedHints() []Hint {
	if !hg.enabled {
		return nil
	}

	return []Hint{
		{
			Type:        string(HintTypeNextAction),
			Title:       "No Project Selected",
			Description: "Select a project to work on tasks, or create a new project",
			NextTools:   []string{"project_select", "project_create", "project_list"},
			Context:     map[string]interface{}{"no_project_selected": true},
		},
		{
			Type:        string(HintTypeSuggestion),
			Title:       "Getting Started",
			Description: "Use project_list to see available projects, or project_create to start fresh",
			NextTools:   []string{"project_list"},
			Context:     map[string]interface{}{"getting_started": true},
		},
	}
}

// GenerateGeneralHints generates general navigation and discovery hints
func (hg *hintGenerator) GenerateGeneralHints() []Hint {
	if !hg.enabled {
		return nil
	}

	return []Hint{
		{
			Type:        string(HintTypeBestPractice),
			Title:       "Project Organization",
			Description: "Keep tasks organized with proper dependencies and priorities",
			NextTools:   []string{"project_list", "status_ready"},
			Context:     map[string]interface{}{"best_practice": true},
		},
		{
			Type:        string(HintTypeSuggestion),
			Title:       "Task Management",
			Description: "Use status commands to find actionable work and monitor progress",
			NextTools:   []string{"status_ready", "status_actionable"},
			Context:     map[string]interface{}{"task_management": true},
		},
	}
}

// FilterHints filters hints by enabled categories
func (hg *hintGenerator) FilterHints(hints []Hint) []Hint {
	if len(hg.categories) == 0 {
		return hints
	}

	var filtered []Hint
	for _, hint := range hints {
		if hg.hintCategoryEnabled(hint.Type) {
			filtered = append(filtered, hint)
		}
	}
	return filtered
}

// Private implementation of the Integration interface
type hintIntegration struct {
	generator Generator
	logger    logger.Logger
}

// NewHintIntegrationImpl creates a new hint integration service with DI (private implementation)
func NewHintIntegrationImpl(generator Generator, logger logger.Logger) Integration {
	return &hintIntegration{
		generator: generator,
		logger:    logger,
	}
}

// GetSessionContext extracts session context information
func (hi *hintIntegration) GetSessionContext(ctx context.Context, projectManager manager.ProjectManager, sessionManager session.SessionManager) (*SessionState, error) {
	// Get session UUID from context
	sessionID, err := shared.GetSessionUUIDFromContext(ctx)
	if err != nil {
		hi.logger.Warn("Invalid session ID format for hint generation",
			logger.String("error", err.Error()))
		return &SessionState{
			Actor:       shared.ActorAnonymous,
			TaskCount:   0,
			RecentTools: []string{},
		}, nil
	}

	// Get session from session manager
	sess, err := sessionManager.GetSession(sessionID)
	if err != nil {
		hi.logger.Warn("Failed to get session for hint generation",
			logger.String("session_id", sessionID.String()),
			logger.Error(err))
		// Try to get actor from MCP context as fallback
		actor := shared.GetSessionActor(ctx)
		if actor == "" {
			actor = shared.ActorMCPUser
		}
		return &SessionState{
			Actor:       actor,
			TaskCount:   0,
			RecentTools: []string{},
		}, nil
	}

	// Get task count if project is selected
	taskCount := 0
	if sess.ProjectID != nil {
		tasks, err := projectManager.ListTasksForProject(ctx, *sess.ProjectID)
		if err != nil {
			hi.logger.Warn("Failed to get tasks for hint generation",
				logger.String("project_id", sess.ProjectID.String()),
				logger.Error(err))
		} else {
			taskCount = len(tasks)
		}
	}

	// Convert project ID to string pointer
	var projectIDStr *string
	if sess.ProjectID != nil {
		pStr := sess.ProjectID.String()
		projectIDStr = &pStr
	}

	// Get actor from session - this is the real actor, not a placeholder
	actor := sess.Actor
	if actor == "" {
		// Try to get from MCP context as additional fallback
		actor = shared.GetSessionActor(ctx)
		if actor == "" {
			actor = shared.ActorMCPUser
		}
	}

	return &SessionState{
		ProjectID:   projectIDStr,
		Actor:       actor,
		TaskCount:   taskCount,
		RecentTools: []string{}, // Not tracked in current session implementation
	}, nil
}

// GenerateToolHints generates hints for specific tool execution
func (hi *hintIntegration) GenerateToolHints(ctx context.Context, toolName string, result interface{}, projectManager manager.ProjectManager, sessionManager session.SessionManager) []Hint {
	sessionState, err := hi.GetSessionContext(ctx, projectManager, sessionManager)
	if err != nil {
		hi.logger.Warn("Failed to get session context for hints",
			logger.String("tool", toolName),
			logger.Error(err))
		// Fallback to basic session state - try to get actor from MCP context
		actor := shared.GetSessionActor(ctx)
		if actor == "" {
			actor = shared.ActorMCPUser
		}
		sessionState = &SessionState{
			Actor:       actor,
			TaskCount:   0,
			RecentTools: []string{},
		}
	}

	hints := hi.generator.GenerateHints(ctx, toolName, result, sessionState)
	return hi.generator.FilterHints(hints)
}

// GenerateContextualHints generates hints based on current session state
func (hi *hintIntegration) GenerateContextualHints(ctx context.Context, projectManager manager.ProjectManager, sessionManager session.SessionManager) []Hint {
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
	if len(hints) > 0 {
		generalHints := hi.generator.GenerateGeneralHints()
		hints = append(hints, generalHints...)
	}

	return hi.generator.FilterHints(hints)
}

// WrapToolResponse wraps a tool response with generated hints
func (hi *hintIntegration) WrapToolResponse(response interface{}, hints []Hint) interface{} {
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
func (hi *hintIntegration) FormatHintsForOutput(hints []Hint) []interface{} {
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
func (hi *hintIntegration) GetHintStats() map[string]interface{} {
	// This is a simple implementation - in a full version, we might track
	// generation statistics over time
	return map[string]interface{}{
		"implementation": "private_integration",
		"generator_type": "DI-enabled",
	}
}

// getSessionIDFromContext extracts session ID from MCP context
func getSessionIDFromContext(ctx context.Context) string {
	if session := server.ClientSessionFromContext(ctx); session != nil {
		return session.SessionID()
	}
	return ""
}

// Helper functions (moved from generator.go)

func (hg *hintGenerator) appendProjectCreatedHints(hints []Hint, result interface{}, session *SessionState) []Hint {
	hints = append(hints, Hint{
		Type:        string(HintTypeNextAction),
		Title:       "Project Created Successfully",
		Description: "Consider creating your first task or exploring project structure",
		NextTools:   []string{"task_create", "project_get"},
		Context:     map[string]interface{}{"project_just_created": true},
	})

	hints = append(hints, Hint{
		Type:        string(HintTypeSuggestion),
		Title:       "Project Setup",
		Description: "Start by defining your project goals and initial tasks",
		NextTools:   []string{"task_create"},
		Context:     map[string]interface{}{"setup_phase": true},
	})

	return hints
}

func (hg *hintGenerator) appendProjectSelectedHints(hints []Hint, result interface{}, session *SessionState) []Hint {
	if session != nil && session.TaskCount == 0 {
		hints = append(hints, Hint{
			Type:        string(HintTypeNextAction),
			Title:       "Project Selected - Ready for Tasks",
			Description: "This project has no tasks yet. Create your first task to get started",
			NextTools:   []string{"task_create"},
			Context:     map[string]interface{}{"project_empty": true},
		})
	} else {
		hints = append(hints, Hint{
			Type:        string(HintTypeNextAction),
			Title:       "Project Selected",
			Description: "You can now work with tasks in this project or select a different one",
			NextTools:   []string{"task_create", "task_list", "status_ready"},
			Context:     map[string]interface{}{"project_has_tasks": true},
		})
	}

	return hints
}

func (hg *hintGenerator) appendTaskCreatedHints(hints []Hint, result interface{}, session *SessionState) []Hint {
	hints = append(hints, Hint{
		Type:        string(HintTypeNextAction),
		Title:       "Task Created Successfully",
		Description: "Consider adding dependencies, subtasks, or moving to the next task",
		NextTools:   []string{"dependency_add", "task_create", "status_ready"},
		Context:     map[string]interface{}{"task_just_created": true},
	})

	if session != nil && session.TaskCount <= 3 {
		hints = append(hints, Hint{
			Type:        string(HintTypeSuggestion),
			Title:       "Early Project Setup",
			Description: "Focus on defining the main structure and dependencies",
			NextTools:   []string{"task_create"},
			Context:     map[string]interface{}{"early_project": true},
		})
	}

	return hints
}

func (hg *hintGenerator) appendTaskUpdatedHints(hints []Hint, result interface{}, session *SessionState) []Hint {
	hints = append(hints, Hint{
		Type:        string(HintTypeNextAction),
		Title:       "Task Updated",
		Description: "Continue working on this task or move to other pending work",
		NextTools:   []string{"status_ready", "task_list", "project_list"},
		Context:     map[string]interface{}{"task_just_updated": true},
	})

	return hints
}

func (hg *hintGenerator) appendTaskDeletedHints(hints []Hint, result interface{}, session *SessionState) []Hint {
	hints = append(hints, Hint{
		Type:        string(HintTypeNextAction),
		Title:       "Task Deleted",
		Description: "Review remaining tasks or create new work",
		NextTools:   []string{"status_ready", "task_list"},
		Context:     map[string]interface{}{"task_just_deleted": true},
	})

	return hints
}

func (hg *hintGenerator) appendProjectListHints(hints []Hint, result interface{}, session *SessionState) []Hint {
	// Get project count from result if possible
	var projectCount int
	if projects, ok := result.([]interface{}); ok {
		projectCount = len(projects)
	}

	if projectCount == 0 {
		hints = append(hints, Hint{
			Type:        string(HintTypeNextAction),
			Title:       "No Projects Found",
			Description: "Create your first project to start using Knot",
			NextTools:   []string{"project_create"},
			Context:     map[string]interface{}{"no_projects": true},
		})
	} else if projectCount == 1 {
		hints = append(hints, Hint{
			Type:        string(HintTypeSuggestion),
			Title:       "Single Project Available",
			Description: "Select this project to start working on tasks",
			NextTools:   []string{"project_select"},
			Context:     map[string]interface{}{"single_project": true},
		})
	} else {
		hints = append(hints, Hint{
			Type:        string(HintTypeSuggestion),
			Title:       "Multiple Projects Available",
			Description: "Select a project or create a new one",
			NextTools:   []string{"project_select", "project_create"},
			Context:     map[string]interface{}{"multiple_projects": true},
		})
	}

	return hints
}

func (hg *hintGenerator) hintCategoryEnabled(hintType string) bool {
	if len(hg.categories) == 0 {
		return true
	}

	// Simple implementation - assume all categories are enabled if categories list is not empty
	// In a full implementation, this would check against specific categories
	return true
}

// DI Provider functions

// NewHintGeneratorProvider provides hint generator service via DI
func NewHintGeneratorProvider(injector do.Injector) (Generator, error) {
	configService := do.MustInvoke[config.Service](injector)
	logger := do.MustInvoke[logger.Logger](injector)
	cfg := configService.GetMCPConfig()
	return NewHintGeneratorImpl(cfg, logger), nil
}

// NewHintIntegrationProvider provides hint integration service via DI
func NewHintIntegrationProvider(injector do.Injector) (Integration, error) {
	generator := do.MustInvoke[Generator](injector)
	logger := do.MustInvoke[logger.Logger](injector)
	return NewHintIntegrationImpl(generator, logger), nil
}
