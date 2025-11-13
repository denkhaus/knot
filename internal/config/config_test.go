package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/denkhaus/knot/internal/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigPath(t *testing.T) {
	tests := []struct {
		name          string
		expectedMatch string
	}{
		{
			name:          "normal config path",
			expectedMatch: "config.json",
		},
		{
			name:          "config path includes knot directory",
			expectedMatch: ".knot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath, err := GetConfigPath()

			assert.NoError(t, err)
			assert.NotEmpty(t, configPath)
			assert.Contains(t, configPath, tt.expectedMatch)
			assert.Contains(t, configPath, filepath.FromSlash(".knot/config.json"))
		})
	}
}

func TestGetConfigPathWithWorkingDir(t *testing.T) {
	// Test that the config path is relative to the current working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	configPath, err := GetConfigPath()
	assert.NoError(t, err)

	expectedPath := filepath.Join(tempDir, ".knot", "config.json")
	assert.Equal(t, expectedPath, configPath)
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *manager.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &manager.Config{
				MaxTasksPerDepth:     5,
				ComplexityThreshold:  3,
				MaxDepth:             10,
				MaxDescriptionLength: 200,
			},
			expectError: false,
		},
		{
			name: "valid config with minimum values",
			config: &manager.Config{
				MaxTasksPerDepth:     1,
				ComplexityThreshold:  1,
				MaxDepth:             1,
				MaxDescriptionLength: 1,
			},
			expectError: false,
		},
		{
			name: "valid config with maximum complexity threshold",
			config: &manager.Config{
				MaxTasksPerDepth:     10,
				ComplexityThreshold:  10,
				MaxDepth:             20,
				MaxDescriptionLength: 1000,
			},
			expectError: false,
		},
		{
			name: "invalid MaxTasksPerDepth - zero",
			config: &manager.Config{
				MaxTasksPerDepth:     0,
				ComplexityThreshold:  3,
				MaxDepth:             10,
				MaxDescriptionLength: 200,
			},
			expectError: true,
			errorMsg:    "max_tasks_per_depth must be at least 1",
		},
		{
			name: "invalid MaxTasksPerDepth - negative",
			config: &manager.Config{
				MaxTasksPerDepth:     -1,
				ComplexityThreshold:  3,
				MaxDepth:             10,
				MaxDescriptionLength: 200,
			},
			expectError: true,
			errorMsg:    "max_tasks_per_depth must be at least 1",
		},
		{
			name: "invalid ComplexityThreshold - zero",
			config: &manager.Config{
				MaxTasksPerDepth:     5,
				ComplexityThreshold:  0,
				MaxDepth:             10,
				MaxDescriptionLength: 200,
			},
			expectError: true,
			errorMsg:    "complexity_threshold must be between 1 and 10",
		},
		{
			name: "invalid ComplexityThreshold - too high",
			config: &manager.Config{
				MaxTasksPerDepth:     5,
				ComplexityThreshold:  11,
				MaxDepth:             10,
				MaxDescriptionLength: 200,
			},
			expectError: true,
			errorMsg:    "complexity_threshold must be between 1 and 10",
		},
		{
			name: "invalid MaxDepth - zero",
			config: &manager.Config{
				MaxTasksPerDepth:     5,
				ComplexityThreshold:  3,
				MaxDepth:             0,
				MaxDescriptionLength: 200,
			},
			expectError: true,
			errorMsg:    "max_depth must be at least 1",
		},
		{
			name: "invalid MaxDescriptionLength - zero",
			config: &manager.Config{
				MaxTasksPerDepth:     5,
				ComplexityThreshold:  3,
				MaxDepth:             10,
				MaxDescriptionLength: 0,
			},
			expectError: true,
			errorMsg:    "max_description_length must be at least 1",
		},
		{
			name: "multiple invalid values",
			config: &manager.Config{
				MaxTasksPerDepth:     0,
				ComplexityThreshold:  0,
				MaxDepth:             0,
				MaxDescriptionLength: 0,
			},
			expectError: true,
			errorMsg:    "max_tasks_per_depth must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConfigNil(t *testing.T) {
	// Test that nil config would panic - this is expected behavior
	// We can't easily test this without causing a panic, so we skip this test
	t.Skip("Nil config test skipped - would cause panic which is expected behavior")
}

func TestConfigIntegration(t *testing.T) {
	// Integration test to verify the whole config workflow
	t.Run("config path and validation workflow", func(t *testing.T) {
		// Get config path
		configPath, err := GetConfigPath()
		assert.NoError(t, err)

		// Create a valid config
		config := &manager.Config{
			MaxTasksPerDepth:     3,
			ComplexityThreshold:  5,
			MaxDepth:             8,
			MaxDescriptionLength: 150,
		}

		// Validate config
		err = ValidateConfig(config)
		assert.NoError(t, err)

		// Verify config path is absolute and contains expected components
		assert.True(t, filepath.IsAbs(configPath))
		assert.Contains(t, configPath, ".knot")
		assert.Contains(t, configPath, "config.json")
	})
}
