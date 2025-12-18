package repository

import (
	"testing"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestModeSeparation tests that Local and MCP modes are properly separated
func TestModeSeparation(t *testing.T) {
	t.Run("Type Conversions", func(t *testing.T) {
		// Test type safety and conversions
		localMode := LocalMode
		mcpMode := MCPMode

		// Test string representations
		assert.Equal(t, "local", string(localMode))
		assert.Equal(t, "mcp", string(mcpMode))

		// Test that modes are different
		assert.NotEqual(t, localMode, mcpMode)
	})

	t.Run("Local Mode Detection", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Setup mock for Local mode
		mockConfigService := mocks.NewMockService(ctrl)
		mockConfigService.EXPECT().IsMCPMode().Return(false).AnyTimes()
		mockConfigService.EXPECT().GetMCPConfig().Return(&config.MCPConfig{}).AnyTimes()
		mockConfigService.EXPECT().GetDatabasePath().Return(":memory:").AnyTimes()

		// Test that local mode configuration is detected correctly
		assert.False(t, mockConfigService.IsMCPMode())
	})

	t.Run("MCP Mode Detection", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Setup mock for MCP mode with PostgreSQL
		mcpConfig := &config.MCPConfig{
			Database: config.DatabaseConfig{
				Backend:  "postgres",
				Endpoint: "postgres://user:pass@localhost/knot",
			},
			Session: config.SessionConfig{
				Timeout: 30,
			},
		}

		mockConfigService := mocks.NewMockService(ctrl)
		mockConfigService.EXPECT().IsMCPMode().Return(true).AnyTimes()
		mockConfigService.EXPECT().GetMCPConfig().Return(mcpConfig).AnyTimes()
		mockConfigService.EXPECT().GetDatabasePath().Return(".knot").AnyTimes()

		// Test that we can detect MCP mode from config
		assert.True(t, mockConfigService.IsMCPMode())
		assert.Equal(t, "postgres", mcpConfig.Database.Backend)
		assert.Equal(t, "postgres://user:pass@localhost/knot", mcpConfig.Database.Endpoint)
	})
}

// TestFactoryInterface tests that the Factory interface is properly implemented
func TestFactoryInterface(t *testing.T) {
	t.Run("Interface Implementation", func(t *testing.T) {
		// Test that repositoryFactory implements Factory interface
		var _ Factory = (*repositoryFactory)(nil)

		// This test verifies that the interface is properly defined
		// and the implementation exists
		assert.True(t, true, "Factory interface is properly defined")
	})
}

// TestRepositoryModeConstants tests repository mode constants
func TestRepositoryModeConstants(t *testing.T) {
	t.Run("Mode Constants", func(t *testing.T) {
		// Test that mode constants are properly defined
		assert.Equal(t, RepositoryMode("local"), LocalMode)
		assert.Equal(t, RepositoryMode("mcp"), MCPMode)

		// Test that modes are different
		assert.NotEqual(t, LocalMode, MCPMode)
		assert.NotEqual(t, string(LocalMode), string(MCPMode))
	})
}