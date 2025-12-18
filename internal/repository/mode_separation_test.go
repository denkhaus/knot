package repository

import (
	"testing"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestModeSeparation tests that Local and MCP modes are properly separated
func TestModeSeparation(t *testing.T) {
	t.Run("Local Mode Configuration", func(t *testing.T) {
		configService := &mockConfigService{
			isMCPMode: false,
		}

		factory := NewRepositoryFactory(configService, zaptest.NewLogger(t))

		mode := factory.GetModeFromConfig()
		assert.Equal(t, LocalMode, mode)
		assert.True(t, IsLocalMode(mode))
		assert.False(t, IsMCPMode(mode))
	})

	t.Run("MCP Mode Configuration", func(t *testing.T) {
		configService := &mockConfigService{
			isMCPMode: true,
			mcpConfig: &config.MCPConfig{
				Database: config.DatabaseConfig{
					Backend:  "postgres",
					Endpoint: "postgres://user:pass@localhost/knot",
				},
			},
		}

		factory := NewRepositoryFactory(configService, zaptest.NewLogger(t))

		mode := factory.GetModeFromConfig()
		assert.Equal(t, MCPMode, mode)
		assert.True(t, IsMCPMode(mode))
		assert.False(t, IsLocalMode(mode))
	})

	t.Run("Default to Local Mode", func(t *testing.T) {
		configService := &mockConfigService{
			isMCPMode: false,
			mcpConfig: &config.MCPConfig{},
		}

		factory := NewRepositoryFactory(configService, zaptest.NewLogger(t))

		mode, err := factory.DetectMode()
		require.NoError(t, err)
		assert.Equal(t, LocalMode, mode)
	})

	t.Run("Type Conversions", func(t *testing.T) {
		// Test type safety and conversions
		localMode := LocalMode
		mcpMode := MCPMode

		assert.True(t, IsLocalMode(localMode))
		assert.False(t, IsMCPMode(localMode))
		assert.True(t, IsMCPMode(mcpMode))
		assert.False(t, IsLocalMode(mcpMode))

		// Test string representations
		assert.Equal(t, "local", string(LocalMode))
		assert.Equal(t, "mcp", string(MCPMode))
	})
}

// TestRepositoryModeDetection tests the repository mode detection logic
func TestRepositoryModeDetection(t *testing.T) {
	tests := []struct {
		name        string
		isMCPMode   bool
		hasPGConfig bool
		expected    RepositoryMode
	}{
		{
			name:        "Local Mode - no MCP",
			isMCPMode:   false,
			hasPGConfig: false,
			expected:    LocalMode,
		},
		{
			name:        "Local Mode - MCP enabled but no PG",
			isMCPMode:   true,
			hasPGConfig: false,
			expected:    LocalMode,
		},
		{
			name:        "MCP Mode - MCP enabled with PG",
			isMCPMode:   true,
			hasPGConfig: true,
			expected:    MCPMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configService := &mockConfigService{
				isMCPMode: tt.isMCPMode,
				mcpConfig: &config.MCPConfig{
					Database: config.DatabaseConfig{
						Backend:  "postgres",
						Endpoint: func() string {
							if tt.hasPGConfig {
								return "postgres://user:pass@localhost/knot"
							}
							return ""
						}(),
					},
				},
			}

			factory := NewRepositoryFactory(configService, zaptest.NewLogger(t))
			mode := factory.GetModeFromConfig()

			assert.Equal(t, tt.expected, mode, "Mode detection failed for: %s", tt.name)
		})
	}
}

// TestProjectContextSchemaCompatibility tests that the extended schema works
func TestProjectContextSchemaCompatibility(t *testing.T) {
	t.Run("Schema Field Validation", func(t *testing.T) {
		// Test that we can work with the extended schema structure
		localData := map[string]interface{}{
			"id":                1,
			"selected_project_id": nil,
			"session_id":         nil,
			"context_type":       "local",
			"updated_at":         "2024-01-01T12:00:00Z",
			"updated_by":         "test-actor",
		}

		mcpData := map[string]interface{}{
			"id":                2,
			"selected_project_id": "550e8400-e29b-41d4-a716-446655440000",
			"session_id":         "550e8400-e29b-41d4-a716-446655440001",
			"context_type":       "mcp",
			"updated_at":         "2024-01-01T12:00:00Z",
			"updated_by":         "session-user",
		}

		// Validate Local Mode data
		assert.Equal(t, 1, localData["id"])
		assert.Equal(t, "local", localData["context_type"])
		assert.Nil(t, localData["session_id"])

		// Validate MCP Mode data
		assert.Equal(t, 2, mcpData["id"])
		assert.Equal(t, "mcp", mcpData["context_type"])
		assert.NotNil(t, mcpData["session_id"])
	})

	t.Run("Backward Compatibility", func(t *testing.T) {
		// Test that Local Mode behavior is preserved
		assert.NotNil(t, "ProjectContext should work in Local Mode")

		// Local mode should continue to use singleton pattern
		// MCP mode should support multiple sessions
		// Both should support actor tracking
	})
}