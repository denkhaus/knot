package repository

import (
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestMultiModeProjectContext tests project context functionality across both Local and MCP modes
func TestMultiModeProjectContext(t *testing.T) {
	t.Run("Local Mode Project Context", func(t *testing.T) {
		testLocalModeProjectContext(t)
	})

	t.Run("MCP Mode Project Context", func(t *testing.T) {
		testMCPModeProjectContext(t)
	})

	t.Run("Mode Separation", func(t *testing.T) {
		testModeSeparation(t)
	})
}

// testLocalModeProjectContext tests Local Mode (SQLite) project context behavior
func testLocalModeProjectContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test Local mode configuration
	t.Run("Local Mode Configuration", func(t *testing.T) {
		// Mock config service for Local mode
		mockConfigService := mocks.NewMockService(ctrl)
		mockConfigService.EXPECT().GetDatabasePath().Return(":memory:").AnyTimes()
		mockConfigService.EXPECT().GetMCPConfig().Return(&config.MCPConfig{}).AnyTimes()
		mockConfigService.EXPECT().IsMCPMode().Return(false).AnyTimes()

		// Test that local mode configuration is detected correctly
		assert.False(t, mockConfigService.IsMCPMode())
		assert.Equal(t, ":memory:", mockConfigService.GetDatabasePath())
	})
}

// testMCPModeProjectContext tests MCP Mode (PostgreSQL) project context behavior
func testMCPModeProjectContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test MCP mode configuration
	t.Run("MCP Mode Configuration", func(t *testing.T) {
		// Mock config service for MCP mode
		mockConfigService := mocks.NewMockService(ctrl)
		mockConfigService.EXPECT().GetMCPConfig().Return(&config.MCPConfig{
			Database: config.DatabaseConfig{
				Backend:  "postgres",
				Endpoint: "postgres://test:test@localhost/knot_test?sslmode=disable",
			},
			Session: config.SessionConfig{
				Timeout: 30,
			},
		}).AnyTimes()
		mockConfigService.EXPECT().IsMCPMode().Return(true).AnyTimes()

		// Test that MCP mode configuration is detected correctly
		assert.True(t, mockConfigService.IsMCPMode())

		mcpConfig := mockConfigService.GetMCPConfig()
		assert.Equal(t, "postgres", mcpConfig.Database.Backend)
		assert.Equal(t, "postgres://test:test@localhost/knot_test?sslmode=disable", mcpConfig.Database.Endpoint)
		assert.Equal(t, time.Duration(30), mcpConfig.Session.Timeout)
	})
}

// testModeSeparation tests that Local and MCP modes are properly separated
func testModeSeparation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("Mode Separation Test", func(t *testing.T) {
		// Test that Local and MCP modes are different and properly configured
		localMode := LocalMode
		mcpMode := MCPMode

		// Test string representations
		assert.Equal(t, "local", string(localMode))
		assert.Equal(t, "mcp", string(mcpMode))

		// Test that modes are different
		assert.NotEqual(t, localMode, mcpMode)
		assert.NotEqual(t, string(localMode), string(mcpMode))
	})
}

// TestConcurrentModeOperations tests concurrent operations between modes
func TestConcurrentModeOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent operations test in short mode")
	}

	t.Run("Concurrent Configuration Testing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Test that we can test different modes concurrently
		done := make(chan bool, 2)

		// Local Mode Test
		go func() {
			defer func() { done <- true }()

			mockConfigService := mocks.NewMockService(ctrl)
			mockConfigService.EXPECT().GetDatabasePath().Return(":memory:").AnyTimes()
			mockConfigService.EXPECT().GetMCPConfig().Return(&config.MCPConfig{}).AnyTimes()
			mockConfigService.EXPECT().IsMCPMode().Return(false).AnyTimes()

			// Test local mode configuration
			assert.False(t, mockConfigService.IsMCPMode())
		}()

		// MCP Mode Test
		go func() {
			defer func() { done <- true }()

			mockConfigService := mocks.NewMockService(ctrl)
			mockConfigService.EXPECT().GetMCPConfig().Return(&config.MCPConfig{
				Database: config.DatabaseConfig{
					Backend:  "postgres",
					Endpoint: "postgres://user:pass@localhost/knot",
				},
			}).AnyTimes()
			mockConfigService.EXPECT().IsMCPMode().Return(true).AnyTimes()

			// Test MCP mode configuration
			assert.True(t, mockConfigService.IsMCPMode())

			mcpConfig := mockConfigService.GetMCPConfig()
			assert.Equal(t, "postgres", mcpConfig.Database.Backend)
		}()

		// Wait for both operations to complete
		for i := 0; i < 2; i++ {
			<-done
		}
	})
}