package template

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/denkhaus/knot/internal/shared"
	"github.com/denkhaus/knot/internal/templates"
	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
)

// loadAllTemplates loads both user and built-in templates, with user templates taking precedence
func loadAllTemplates() ([]*types.TaskTemplate, error) {
	var allTemplates []*types.TaskTemplate
	templateMap := make(map[string]*types.TaskTemplate)

	// Load built-in templates first
	builtInTemplates, err := templates.LoadBuiltInTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to load built-in templates: %w", err)
	}

	// Add built-in templates to map
	for _, template := range builtInTemplates {
		templateMap[strings.ToLower(template.Name)] = template
	}

	// Load user templates (these override built-in templates with same name)
	userTemplates, err := templates.LoadUserTemplates()
	if err != nil {
		// Don't fail if user templates can't be loaded, just log and continue
		// In a real implementation, we'd use proper logging here
	} else {
		// User templates take precedence
		for _, template := range userTemplates {
			templateMap[strings.ToLower(template.Name)] = template
		}
	}

	// Convert map back to slice
	for _, template := range templateMap {
		allTemplates = append(allTemplates, template)
	}

	return allTemplates, nil
}

// loadBuiltInTemplates loads all built-in templates from embedded filesystem
func loadBuiltInTemplates() ([]*types.TaskTemplate, error) {
	return templates.LoadBuiltInTemplates()
}

// loadTemplateFromFile loads a template from a YAML file (delegates to templates package)
func loadTemplateFromFile(filePath string) (*types.TaskTemplate, error) {
	return templates.LoadTemplateFromFile(filePath)
}

// findTemplateByName finds a template by name from all available templates (user + built-in)
func findTemplateByName(name string) (*types.TaskTemplate, error) {
	allTemplates, err := loadAllTemplates()
	if err != nil {
		return nil, err
	}

	for _, template := range allTemplates {
		if strings.EqualFold(template.Name, name) {
			return template, nil
		}
	}

	return nil, fmt.Errorf("template '%s' not found", name)
}

// applyTemplateFilters applies filters to template list
func applyTemplateFilters(templates []*types.TaskTemplate, c *cli.Context) []*types.TaskTemplate {
	var filtered []*types.TaskTemplate

	for _, template := range templates {
		// Category filter
		if category := c.String("category"); category != "" {
			if !strings.EqualFold(template.Category, category) {
				continue
			}
		}

		// Tags filter
		if tags := c.StringSlice("tags"); len(tags) > 0 {
			hasTag := false
			for _, filterTag := range tags {
				for _, templateTag := range template.Tags {
					if strings.EqualFold(templateTag, filterTag) {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		// Search filter
		if search := c.String("search"); search != "" {
			searchLower := strings.ToLower(search)
			nameMatch := strings.Contains(strings.ToLower(template.Name), searchLower)
			descMatch := strings.Contains(strings.ToLower(template.Description), searchLower)
			if !nameMatch && !descMatch {
				continue
			}
		}

		// Built-in filter
		if c.Bool("built-in") && !template.IsBuiltIn {
			continue
		}

		filtered = append(filtered, template)
	}

	return filtered
}

// validateTemplate validates a template structure
func validateTemplate(template *types.TaskTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if len(template.Tasks) == 0 {
		return fmt.Errorf("template must have at least one task")
	}

	// Validate task IDs are unique
	taskIDs := make(map[string]bool)
	for _, task := range template.Tasks {
		if task.ID == "" {
			return fmt.Errorf("task ID is required")
		}
		if taskIDs[task.ID] {
			return fmt.Errorf("duplicate task ID: %s", task.ID)
		}
		taskIDs[task.ID] = true

		if task.Title == "" {
			return fmt.Errorf("task title is required for task %s", task.ID)
		}

		if task.Complexity < 1 || task.Complexity > 10 {
			return fmt.Errorf("task complexity must be between 1 and 10 for task %s", task.ID)
		}
	}

	// Validate dependencies reference existing tasks
	for _, task := range template.Tasks {
		for _, depID := range task.Dependencies {
			if !taskIDs[depID] {
				return fmt.Errorf("task %s references non-existent dependency: %s", task.ID, depID)
			}
		}
		if task.ParentID != nil && !taskIDs[*task.ParentID] {
			return fmt.Errorf("task %s references non-existent parent: %s", task.ID, *task.ParentID)
		}
	}

	// Validate variables
	varNames := make(map[string]bool)
	for _, variable := range template.Variables {
		if variable.Name == "" {
			return fmt.Errorf("variable name is required")
		}
		if varNames[variable.Name] {
			return fmt.Errorf("duplicate variable name: %s", variable.Name)
		}
		varNames[variable.Name] = true

		if variable.Type == types.VarTypeChoice && len(variable.Options) == 0 {
			return fmt.Errorf("choice variable %s must have options", variable.Name)
		}
	}

	return nil
}

// validateTemplateVariables validates that required variables are provided
func validateTemplateVariables(template *types.TaskTemplate, variables map[string]string) error {
	for _, variable := range template.Variables {
		value, provided := variables[variable.Name]
		
		if variable.Required && !provided {
			return fmt.Errorf("required variable '%s' not provided", variable.Name)
		}

		if provided {
			// Validate choice variables
			if variable.Type == types.VarTypeChoice {
				validOption := false
				for _, option := range variable.Options {
					if value == option {
						validOption = true
						break
					}
				}
				if !validOption {
					return fmt.Errorf("variable '%s' must be one of: %s", variable.Name, strings.Join(variable.Options, ", "))
				}
			}
		}
	}

	return nil
}

// applyTemplate applies a template to create tasks
func applyTemplate(appCtx *shared.AppContext, template *types.TaskTemplate, projectID uuid.UUID, parentID *uuid.UUID, variables map[string]string, dryRun bool) (*types.TemplateApplyResult, error) {
	result := &types.TemplateApplyResult{
		Success: true,
	}

	// Apply default values for missing variables
	finalVariables := make(map[string]string)
	for k, v := range variables {
		finalVariables[k] = v
	}
	for _, variable := range template.Variables {
		if _, exists := finalVariables[variable.Name]; !exists && variable.DefaultValue != nil {
			finalVariables[variable.Name] = *variable.DefaultValue
		}
	}

	// Create task ID mapping for dependencies
	taskIDMap := make(map[string]uuid.UUID)
	
	// First pass: create all tasks without dependencies
	for _, taskSpec := range template.Tasks {
		// Skip conditional tasks that don't match
		if shouldSkipTask(taskSpec, finalVariables) {
			continue
		}

		task := &types.Task{
			ID:          uuid.New(),
			ProjectID:   projectID,
			Title:       substituteVariables(taskSpec.Title, finalVariables),
			Description: substituteVariables(taskSpec.Description, finalVariables),
			State:       types.TaskStatePending,
			Complexity:  taskSpec.Complexity,
			Estimate:    taskSpec.Estimate,
		}

		// Set parent if this task has a parent in the template
		if taskSpec.ParentID != nil {
			if parentTaskID, exists := taskIDMap[*taskSpec.ParentID]; exists {
				task.ParentID = &parentTaskID
			}
		} else if parentID != nil {
			// Use the provided parent ID for root tasks
			task.ParentID = parentID
		}

		taskIDMap[taskSpec.ID] = task.ID

		if !dryRun {
			actor := appCtx.GetActor()
			createdTask, err := appCtx.ProjectManager.CreateTask(
				context.Background(),
				projectID,
				task.ParentID,
				task.Title,
				task.Description,
				task.Complexity,
				types.TaskPriorityMedium,
				actor,
			)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to create task '%s': %v", task.Title, err))
				result.Success = false
				continue
			}
			result.CreatedTasks = append(result.CreatedTasks, createdTask)
		} else {
			result.CreatedTasks = append(result.CreatedTasks, task)
		}
	}

	// TODO: Second pass for dependencies (requires dependency management implementation)

	if !dryRun && result.Success {
		// Create template instance record
		instance := &types.TemplateInstance{
			TemplateID:   template.ID,
			TemplateName: template.Name,
			Variables:    finalVariables,
			CreatedAt:    appCtx.ProjectManager.GetCurrentTime(),
			CreatedBy:    appCtx.GetActor(),
		}
		for _, task := range result.CreatedTasks {
			instance.CreatedTasks = append(instance.CreatedTasks, task.ID)
		}
		result.Instance = instance
	}

	return result, nil
}

// shouldSkipTask determines if a task should be skipped based on conditional metadata
func shouldSkipTask(taskSpec types.TaskSpec, variables map[string]string) bool {
	if conditional, exists := taskSpec.Metadata["conditional"]; exists {
		// Simple boolean variable check
		if strings.HasPrefix(conditional, "{{") && strings.HasSuffix(conditional, "}}") {
			varName := strings.Trim(conditional, "{}")
			if value, exists := variables[varName]; exists {
				return value != "true" && value != "yes" && value != "1"
			}
			return true // Skip if variable not found
		}
	}
	return false
}

// substituteVariables replaces template variables in text
func substituteVariables(text string, variables map[string]string) string {
	result := text
	for name, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", name)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// Output functions
func outputTemplatesAsJSON(templates []*types.TaskTemplate) error {
	data, err := json.MarshalIndent(templates, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal templates to JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputTemplateAsJSON(template *types.TaskTemplate) error {
	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal template to JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputApplyResultAsJSON(result *types.TemplateApplyResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result to JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}