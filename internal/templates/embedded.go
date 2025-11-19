package templates

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

//go:embed bug-fix.yaml code-review.yaml feature-development.yaml
var embeddedTemplates embed.FS

// LoadBuiltInTemplates loads all built-in templates from embedded filesystem
func LoadBuiltInTemplates() ([]*types.TaskTemplate, error) {
	var templates []*types.TaskTemplate

	entries, err := embeddedTemplates.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded templates: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		template, err := loadTemplateFromEmbedded(entry.Name())
		if err != nil {
			// Log error but continue with other templates
			continue
		}

		template.IsBuiltIn = true
		templates = append(templates, template)
	}

	return templates, nil
}

// loadTemplateFromEmbedded loads a template from the embedded filesystem
func loadTemplateFromEmbedded(filename string) (*types.TaskTemplate, error) {
	data, err := embeddedTemplates.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded template %s: %w", filename, err)
	}

	var template types.TaskTemplate
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse template YAML %s: %w", filename, err)
	}

	// Generate deterministic ID for built-in templates based on filename
	if template.ID == uuid.Nil {
		template.ID = generateDeterministicUUID(filename)
	}

	return &template, nil
}

// GetEmbeddedTemplateFS returns the embedded filesystem for external access
func GetEmbeddedTemplateFS() fs.FS {
	return embeddedTemplates
}

// ListEmbeddedTemplateFiles returns a list of embedded template filenames
func ListEmbeddedTemplateFiles() ([]string, error) {
	var files []string

	entries, err := embeddedTemplates.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded templates: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// GetEmbeddedTemplateContent returns the raw content of an embedded template
func GetEmbeddedTemplateContent(filename string) ([]byte, error) {
	return embeddedTemplates.ReadFile(filename)
}

// generateDeterministicUUID creates a deterministic UUID based on filename
// This ensures built-in templates always have the same UUID
func generateDeterministicUUID(filename string) uuid.UUID {
	// Create deterministic UUID based on filename
	hash := sha256.Sum256([]byte("knot-builtin-template:" + filename))

	// Use first 16 bytes of hash to create UUID
	var uuidBytes [16]byte
	copy(uuidBytes[:], hash[:16])

	// Set version (4) and variant bits
	uuidBytes[6] = (uuidBytes[6] & 0x0f) | 0x40 // Version 4
	uuidBytes[8] = (uuidBytes[8] & 0x3f) | 0x80 // Variant 10

	return uuid.UUID(uuidBytes)
}
