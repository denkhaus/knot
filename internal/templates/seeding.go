package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/denkhaus/knot/internal/types"
	"gopkg.in/yaml.v3"
)

// SeedingMetadata tracks which templates have been seeded
type SeedingMetadata struct {
	SeededTemplates map[string]SeededTemplateInfo `yaml:"seeded_templates"`
	LastSeedTime    string                        `yaml:"last_seed_time"`
}

// SeededTemplateInfo contains information about a seeded template
type SeededTemplateInfo struct {
	Name     string `yaml:"name"`
	SeededAt string `yaml:"seeded_at"`
}

// AutoSeedTemplates automatically seeds built-in templates to user directory on first .knot creation
func AutoSeedTemplates() error {
	// Check if this is first-time setup by looking for seeding metadata
	templatesDir, err := GetUserTemplatesDir()
	if err != nil {
		return fmt.Errorf("failed to get templates directory: %w", err)
	}
	metadataPath := filepath.Join(templatesDir, "metadata.yaml")
	if _, err := os.Stat(metadataPath); err == nil {
		// Metadata exists, seeding already done
		return nil
	}

	// Ensure templates directory exists
	if err := EnsureUserTemplatesDir(); err != nil {
		return fmt.Errorf("failed to ensure templates directory: %w", err)
	}

	// Load built-in templates
	builtInTemplates, err := LoadBuiltInTemplates()
	if err != nil {
		return fmt.Errorf("failed to load built-in templates: %w", err)
	}

	// Seed each built-in template
	metadata := &SeedingMetadata{
		SeededTemplates: make(map[string]SeededTemplateInfo),
		LastSeedTime:    getCurrentTimestamp(),
	}

	for _, template := range builtInTemplates {
		if err := seedTemplate(template); err != nil {
			// Log error but continue with other templates
			fmt.Printf("Warning: Failed to seed template '%s': %v\n", template.Name, err)
			continue
		}

		// Use template UUID as key, store template info as value
		templateKey := template.ID.String()
		metadata.SeededTemplates[templateKey] = SeededTemplateInfo{
			Name:     template.Name,
			SeededAt: getCurrentTimestamp(),
		}
	}

	// Save seeding metadata
	if err := saveMetadata(metadataPath, metadata); err != nil {
		return fmt.Errorf("failed to save seeding metadata: %w", err)
	}

	seededCount := len(metadata.SeededTemplates)
	fmt.Printf("Auto-seeded %d built-in templates to .knot/templates/\n", seededCount)
	return nil
}

// seedTemplate copies a built-in template to the user templates directory
func seedTemplate(template *types.TaskTemplate) error {
	templatesDir, err := GetUserTemplatesDir()
	if err != nil {
		return err
	}

	// Generate filename from template name
	filename := strings.ToLower(strings.ReplaceAll(template.Name, " ", "-")) + ".yaml"
	filePath := filepath.Join(templatesDir, filename)

	// Check if file already exists (user might have created it manually)
	if _, err := os.Stat(filePath); err == nil {
		// File exists, don't overwrite
		return nil
	}

	// Marshal template to YAML
	data, err := yaml.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template to YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}

// getKnotDir returns the .knot directory path
func getKnotDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	return filepath.Join(cwd, ".knot"), nil
}

// saveMetadata saves seeding metadata to file
func saveMetadata(path string, metadata *SeedingMetadata) error {
	data, err := yaml.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// getCurrentTimestamp returns current timestamp as string
func getCurrentTimestamp() string {
	// Use a simple timestamp format
	return fmt.Sprintf("%d", 1640995200) // Unix timestamp placeholder
}

// CheckAndSeedIfNeeded checks if seeding is needed and performs it
func CheckAndSeedIfNeeded() error {
	knotDir, err := getKnotDir()
	if err != nil {
		return err
	}

	// Check if .knot directory exists
	if _, err := os.Stat(knotDir); os.IsNotExist(err) {
		// .knot directory doesn't exist, no need to seed
		return nil
	}

	// Check if templates directory exists
	templatesDir, err := GetUserTemplatesDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		// Templates directory doesn't exist, create it and seed
		return AutoSeedTemplates()
	}

	// Load current seeding metadata
	metadata, err := LoadSeedingMetadata()
	if err != nil {
		return fmt.Errorf("failed to load seeding metadata: %w", err)
	}

	// If no templates have been seeded yet, perform seeding
	if len(metadata.SeededTemplates) == 0 {
		return AutoSeedTemplates()
	}

	// Check if there are new built-in templates that haven't been seeded
	builtInTemplates, err := LoadBuiltInTemplates()
	if err != nil {
		return fmt.Errorf("failed to load built-in templates: %w", err)
	}

	newTemplatesFound := false
	for _, template := range builtInTemplates {
		if _, exists := metadata.SeededTemplates[template.ID.String()]; !exists {
			// New template found, seed it
			if err := seedTemplate(template); err != nil {
				fmt.Printf("Warning: Failed to seed new template '%s': %v\n", template.Name, err)
				continue
			}

			// Update metadata
			if err := UpdateSeededTemplate(template); err != nil {
				fmt.Printf("Warning: Failed to update metadata for template '%s': %v\n", template.Name, err)
			} else {
				fmt.Printf("Auto-seeded new template: %s\n", template.Name)
				newTemplatesFound = true
			}
		}
	}

	if newTemplatesFound {
		fmt.Println("New built-in templates have been seeded to .knot/templates/")
	}

	return nil
}
