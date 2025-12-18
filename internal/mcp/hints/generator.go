package hints

import (
	"context"
	"strings"
)

// HintType represents different types of hints
type HintType string

const (
	HintTypeNextAction   HintType = "next_action"
	HintTypeWarning      HintType = "warning"
	HintTypeSuggestion   HintType = "suggestion"
	HintTypeBestPractice HintType = "best_practice"
)

// Hint represents a suggestion for next actions for the agent
type Hint struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	NextTools   []string               `json:"next_tools,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// HintGenerator provides context-aware hint generation for MCP operations
type HintGenerator struct {
	enabled    bool
	maxHints   int
	categories []string
}

// NewHintGenerator creates a new hint generator
func NewHintGenerator(enabled bool, maxHints int, categories []string) *HintGenerator {
	return &HintGenerator{
		enabled:    enabled,
		maxHints:   maxHints,
		categories: categories,
	}
}

// GenerateHints generates hints based on the current operation and session state
func (hg *HintGenerator) GenerateHints(ctx context.Context, toolName string, result interface{}, sessionState *SessionState) []Hint {
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

	return hints
}

// SessionState represents the current session context state
type SessionState struct {
	ProjectID   *string
	Actor       string
	TaskCount   int
	RecentTools []string
}

// appendProjectCreatedHints adds hints after project creation
func (hg *HintGenerator) appendProjectCreatedHints(hints []Hint, result interface{}, session *SessionState) []Hint {
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

// appendProjectSelectedHints adds hints after project selection
func (hg *HintGenerator) appendProjectSelectedHints(hints []Hint, result interface{}, session *SessionState) []Hint {
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

// appendTaskCreatedHints adds hints after task creation
func (hg *HintGenerator) appendTaskCreatedHints(hints []Hint, result interface{}, session *SessionState) []Hint {
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

// appendTaskUpdatedHints adds hints after task updates
func (hg *HintGenerator) appendTaskUpdatedHints(hints []Hint, result interface{}, session *SessionState) []Hint {
	hints = append(hints, Hint{
		Type:        string(HintTypeNextAction),
		Title:       "Task Updated",
		Description: "Continue working on this task or move to other pending work",
		NextTools:   []string{"status_ready", "task_list", "project_list"},
		Context:     map[string]interface{}{"task_just_updated": true},
	})

	return hints
}

// appendTaskDeletedHints adds hints after task deletion
func (hg *HintGenerator) appendTaskDeletedHints(hints []Hint, result interface{}, session *SessionState) []Hint {
	hints = append(hints, Hint{
		Type:        string(HintTypeNextAction),
		Title:       "Task Deleted",
		Description: "Review remaining tasks or create new work",
		NextTools:   []string{"status_ready", "task_list"},
		Context:     map[string]interface{}{"task_just_deleted": true},
	})

	return hints
}

// appendProjectListHints adds hints after project listing
func (hg *HintGenerator) appendProjectListHints(hints []Hint, result interface{}, session *SessionState) []Hint {
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

// GenerateNoProjectSelectedHints generates hints when no project is selected
func (hg *HintGenerator) GenerateNoProjectSelectedHints() []Hint {
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

// GenerateGeneralHints generates general hints for navigation and discovery
func (hg *HintGenerator) GenerateGeneralHints() []Hint {
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
func (hg *HintGenerator) FilterHints(hints []Hint) []Hint {
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

// hintCategoryEnabled checks if a hint type/category is enabled
func (hg *HintGenerator) hintCategoryEnabled(hintType string) bool {
	if len(hg.categories) == 0 {
		return true
	}

	for _, category := range hg.categories {
		category = strings.ToLower(category)
		hintType = strings.ToLower(hintType)

		// Map hint types to categories
		switch category {
		case "general":
			if strings.Contains(hintType, "best_practice") || strings.Contains(hintType, "suggestion") {
				return true
			}
		case "actions":
			if strings.Contains(hintType, "next_action") {
				return true
			}
		case "warnings":
			if strings.Contains(hintType, "warning") {
				return true
			}
		case "all":
			return true
		}

		// Direct category match
		if strings.Contains(hintType, category) {
			return true
		}
	}

	return false
}
