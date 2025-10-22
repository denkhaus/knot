package types

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TaskTemplate represents a template for creating tasks with predefined structure
type TaskTemplate struct {
	ID          uuid.UUID    `json:"id" yaml:"id"`
	Name        string       `json:"name" yaml:"name"`                               // Template name (e.g., "Bug Fix", "Feature Development")
	Description string       `json:"description" yaml:"description"`                 // Template description
	Category    string       `json:"category" yaml:"category"`                       // Category (e.g., "Development", "Testing", "Documentation")
	Tags        []string     `json:"tags" yaml:"tags"`                               // Tags for filtering/searching templates
	Tasks       []TaskSpec   `json:"tasks" yaml:"tasks"`                             // Task specifications in the template
	Variables   []Variable   `json:"variables,omitempty" yaml:"variables,omitempty"` // Template variables for customization
	CreatedAt   time.Time    `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" yaml:"updated_at"`
	CreatedBy   string       `json:"created_by" yaml:"created_by"`
	IsBuiltIn   bool         `json:"is_built_in" yaml:"is_built_in"` // Whether this is a built-in template
}

// TaskSpec defines a task specification within a template
type TaskSpec struct {
	ID           string            `json:"id" yaml:"id"`                                     // Unique ID within template (for referencing)
	Title        string            `json:"title" yaml:"title"`                               // Task title (can contain variables like {{feature_name}})
	Description  string            `json:"description" yaml:"description"`                   // Task description (can contain variables)
	Complexity   int               `json:"complexity" yaml:"complexity"`                     // Default complexity (1-10)
	ParentID     *string           `json:"parent_id,omitempty" yaml:"parent_id,omitempty"`   // Parent task ID within template
	Dependencies []string          `json:"dependencies,omitempty" yaml:"dependencies,omitempty"` // Dependencies within template
	Estimate     *int64            `json:"estimate,omitempty" yaml:"estimate,omitempty"`     // Time estimate in minutes
	Metadata     map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`     // Additional metadata
}

// Variable defines a template variable that can be substituted
type Variable struct {
	Name         string   `json:"name" yaml:"name"`                                   // Variable name (e.g., "feature_name", "bug_id")
	Description  string   `json:"description" yaml:"description"`                     // Variable description
	Type         VarType  `json:"type" yaml:"type"`                                   // Variable type
	Required     bool     `json:"required" yaml:"required"`                           // Whether variable is required
	DefaultValue *string  `json:"default_value,omitempty" yaml:"default_value,omitempty"` // Default value if not provided
	Options      []string `json:"options,omitempty" yaml:"options,omitempty"`         // Valid options for choice type
}

// VarType represents the type of a template variable
type VarType string

const (
	VarTypeString VarType = "string"
	VarTypeInt    VarType = "int"
	VarTypeBool   VarType = "bool"
	VarTypeChoice VarType = "choice" // Select from predefined options
)

// TemplateInstance represents the result of applying a template
type TemplateInstance struct {
	TemplateID   uuid.UUID         `json:"template_id"`
	TemplateName string            `json:"template_name"`
	Variables    map[string]string `json:"variables"`    // Variable values used
	CreatedTasks []uuid.UUID       `json:"created_tasks"` // IDs of tasks created from template
	CreatedAt    time.Time         `json:"created_at"`
	CreatedBy    string            `json:"created_by"`
}

// TemplateFilter represents filtering options for template queries
type TemplateFilter struct {
	Category  *string  `json:"category,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	IsBuiltIn *bool    `json:"is_built_in,omitempty"`
	Search    *string  `json:"search,omitempty"` // Search in name and description
}

// TemplateRepository defines the interface for template persistence
type TemplateRepository interface {
	// Template CRUD operations
	CreateTemplate(ctx context.Context, template *TaskTemplate) error
	GetTemplate(ctx context.Context, id uuid.UUID) (*TaskTemplate, error)
	UpdateTemplate(ctx context.Context, template *TaskTemplate) error
	DeleteTemplate(ctx context.Context, id uuid.UUID) error
	ListTemplates(ctx context.Context, filter TemplateFilter) ([]*TaskTemplate, error)

	// Template instance tracking
	CreateTemplateInstance(ctx context.Context, instance *TemplateInstance) error
	GetTemplateInstances(ctx context.Context, templateID uuid.UUID) ([]*TemplateInstance, error)
	GetTemplateInstancesByProject(ctx context.Context, projectID uuid.UUID) ([]*TemplateInstance, error)
}

// TemplateApplyRequest represents a request to apply a template
type TemplateApplyRequest struct {
	TemplateID  uuid.UUID         `json:"template_id"`
	ProjectID   uuid.UUID         `json:"project_id"`
	ParentID    *uuid.UUID        `json:"parent_id,omitempty"` // Optional parent task for all template tasks
	Variables   map[string]string `json:"variables"`           // Variable substitutions
	DryRun      bool              `json:"dry_run"`             // Preview mode - don't actually create tasks
}

// TemplateApplyResult represents the result of applying a template
type TemplateApplyResult struct {
	Success      bool              `json:"success"`
	CreatedTasks []*Task           `json:"created_tasks,omitempty"` // Tasks that would be/were created
	Errors       []string          `json:"errors,omitempty"`        // Any errors that occurred
	Instance     *TemplateInstance `json:"instance,omitempty"`      // Template instance record
}