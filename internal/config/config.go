package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/denkhaus/knot/v2/internal/manager"
)

// ConfigFile provides file-based configuration management for Knot
// Uses the existing manager.Config struct but adds file persistence
// REFERENCE: pkg/tools/project/interfaces.go lines 64-80 (original Config struct)

// GetConfigPath returns the path to the knot configuration file
func GetConfigPath() (string, error) {
	// Use .knot directory for configuration (same as database)
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	knotDir := filepath.Join(cwd, ".knot")
	configPath := filepath.Join(knotDir, "config.json")
	return configPath, nil
}

// ValidateConfig checks if the configuration values are valid
func ValidateConfig(c *manager.Config) error {
	if c.MaxTasksPerDepth < 1 {
		return fmt.Errorf("max_tasks_per_depth must be at least 1, got %d", c.MaxTasksPerDepth)
	}
	if c.ComplexityThreshold < 1 || c.ComplexityThreshold > 10 {
		return fmt.Errorf("complexity_threshold must be between 1 and 10, got %d", c.ComplexityThreshold)
	}
	if c.MaxDepth < 1 {
		return fmt.Errorf("max_depth must be at least 1, got %d", c.MaxDepth)
	}
	if c.MaxDescriptionLength < 1 {
		return fmt.Errorf("max_description_length must be at least 1, got %d", c.MaxDescriptionLength)
	}
	return nil
}
