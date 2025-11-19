package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// GetUserTemplatesDir returns the local project templates directory path
func GetUserTemplatesDir() (string, error) {
	// Use local .knot directory in current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	return filepath.Join(cwd, ".knot", "templates"), nil
}

// EnsureUserTemplatesDir creates the user templates directory if it doesn't exist
func EnsureUserTemplatesDir() error {
	dir, err := GetUserTemplatesDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

// LoadUserTemplates loads all user templates from ~/.knot/templates/
func LoadUserTemplates() ([]*types.TaskTemplate, error) {
	var templates []*types.TaskTemplate

	templatesDir, err := GetUserTemplatesDir()
	if err != nil {
		return templates, err
	}

	// Check if directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		return templates, nil // No user templates directory, return empty list
	}

	// Find all YAML files in the directory
	files, err := filepath.Glob(filepath.Join(templatesDir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to find user template files: %w", err)
	}

	for _, file := range files {
		template, err := LoadTemplateFromFile(file)
		if err != nil {
			// Log error but continue with other templates
			continue
		}
		template.IsBuiltIn = false
		templates = append(templates, template)
	}

	return templates, nil
}

// LoadTemplateFromFile loads a template from a YAML file
func LoadTemplateFromFile(filePath string) (*types.TaskTemplate, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var template types.TaskTemplate
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse template YAML: %w", err)
	}

	// Generate ID if not present
	if template.ID == uuid.Nil {
		template.ID = uuid.New()
	}

	return &template, nil
}

// SaveUserTemplate saves a template to the user templates directory
func SaveUserTemplate(template *types.TaskTemplate) error {
	if err := EnsureUserTemplatesDir(); err != nil {
		return err
	}

	templatesDir, err := GetUserTemplatesDir()
	if err != nil {
		return err
	}

	// Generate filename from template name
	filename := strings.ToLower(strings.ReplaceAll(template.Name, " ", "-")) + ".yaml"
	filePath := filepath.Join(templatesDir, filename)

	data, err := yaml.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template to YAML: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}

// DeleteUserTemplate deletes a user template by name
func DeleteUserTemplate(templateName string) error {
	templatesDir, err := GetUserTemplatesDir()
	if err != nil {
		return err
	}

	// Generate filename from template name
	filename := strings.ToLower(strings.ReplaceAll(templateName, " ", "-")) + ".yaml"
	filePath := filepath.Join(templatesDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("user template '%s' not found", templateName)
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete template file: %w", err)
	}

	return nil
}

// GetUserTemplateFilePath returns the file path for a user template
func GetUserTemplateFilePath(templateName string) (string, error) {
	templatesDir, err := GetUserTemplatesDir()
	if err != nil {
		return "", err
	}

	filename := strings.ToLower(strings.ReplaceAll(templateName, " ", "-")) + ".yaml"
	return filepath.Join(templatesDir, filename), nil
}

// UserTemplateExists checks if a user template exists
func UserTemplateExists(templateName string) bool {
	filePath, err := GetUserTemplateFilePath(templateName)
	if err != nil {
		return false
	}

	_, err = os.Stat(filePath)
	return !os.IsNotExist(err)
}
