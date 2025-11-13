package templates

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// LoadSeedingMetadata loads seeding metadata from file
func LoadSeedingMetadata() (*SeedingMetadata, error) {
	templatesDir, err := GetUserTemplatesDir()
	if err != nil {
		return nil, err
	}

	metadataPath := filepath.Join(templatesDir, "metadata.yaml")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		// Return empty metadata if file doesn't exist
		return &SeedingMetadata{
			SeededTemplates: make(map[string]SeededTemplateInfo),
			LastSeedTime:    "",
		}, nil
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata SeedingMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata YAML: %w", err)
	}

	// Initialize map if nil (for backward compatibility)
	if metadata.SeededTemplates == nil {
		metadata.SeededTemplates = make(map[string]SeededTemplateInfo)
	}

	return &metadata, nil
}

// IsTemplateSeeded checks if a template with given UUID has been seeded
func IsTemplateSeeded(templateID uuid.UUID) (bool, error) {
	metadata, err := LoadSeedingMetadata()
	if err != nil {
		return false, err
	}

	_, exists := metadata.SeededTemplates[templateID.String()]
	return exists, nil
}

// GetSeededTemplateInfo returns information about a seeded template
func GetSeededTemplateInfo(templateID uuid.UUID) (*SeededTemplateInfo, error) {
	metadata, err := LoadSeedingMetadata()
	if err != nil {
		return nil, err
	}

	info, exists := metadata.SeededTemplates[templateID.String()]
	if !exists {
		return nil, fmt.Errorf("template %s not found in seeded templates", templateID.String())
	}

	return &info, nil
}

// UpdateSeededTemplate updates or adds a seeded template entry
func UpdateSeededTemplate(template *types.TaskTemplate) error {
	metadata, err := LoadSeedingMetadata()
	if err != nil {
		return err
	}

	templateKey := template.ID.String()
	metadata.SeededTemplates[templateKey] = SeededTemplateInfo{
		Name:     template.Name,
		SeededAt: getCurrentTimestamp(),
	}

	return saveSeedingMetadata(metadata)
}

// RemoveSeededTemplate removes a template from seeded templates tracking
func RemoveSeededTemplate(templateID uuid.UUID) error {
	metadata, err := LoadSeedingMetadata()
	if err != nil {
		return err
	}

	delete(metadata.SeededTemplates, templateID.String())
	return saveSeedingMetadata(metadata)
}

// saveSeedingMetadata saves metadata to file
func saveSeedingMetadata(metadata *SeedingMetadata) error {
	templatesDir, err := GetUserTemplatesDir()
	if err != nil {
		return err
	}

	metadataPath := filepath.Join(templatesDir, "metadata.yaml")
	return saveMetadata(metadataPath, metadata)
}

// ListSeededTemplates returns all seeded template information
func ListSeededTemplates() (map[string]SeededTemplateInfo, error) {
	metadata, err := LoadSeedingMetadata()
	if err != nil {
		return nil, err
	}

	return metadata.SeededTemplates, nil
}

// GetSeedingStats returns statistics about seeded templates
func GetSeedingStats() (map[string]interface{}, error) {
	metadata, err := LoadSeedingMetadata()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_seeded":   len(metadata.SeededTemplates),
		"last_seed_time": metadata.LastSeedTime,
		"templates":      make([]map[string]string, 0),
	}

	for uuid, info := range metadata.SeededTemplates {
		templateInfo := map[string]string{
			"uuid":      uuid,
			"name":      info.Name,
			"seeded_at": info.SeededAt,
		}
		stats["templates"] = append(stats["templates"].([]map[string]string), templateInfo)
	}

	return stats, nil
}
