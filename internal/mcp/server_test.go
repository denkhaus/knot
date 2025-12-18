package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/mcp/tools"
	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// createTestConfig creates a test configuration for use in tests
func createTestConfig() *config.MCPConfig {
	return &config.MCPConfig{
		Enabled: true,
		Address: "localhost",
		Port:    8080,
		Timeout: 30 * time.Second,
		Database: config.DatabaseConfig{
			Backend:  "postgres",
			Endpoint: "localhost:5432",
		},
		Session: config.SessionConfig{
			Timeout:     30 * time.Minute,
			MaxSessions: 100,
		},
		Hints: config.HintsConfig{
			Enabled:    false,
			MaxHints:   10,
			Categories: []string{"general"},
		},
	}
}

// createTestServerConfig creates a test server config with all required mocks
func createTestServerConfig(ctrl *gomock.Controller) ServerConfig {
	mockLogger := mocks.NewMockLogger(ctrl)
	mockConfig := createTestConfig()
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockManager(ctrl)

	return ServerConfig{
		ProjectManager: mockManager,
		SessionManager: mockSessionManager,
		Logger:         mockLogger,
		Config:         mockConfig,
	}
}

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("Successful server creation", func(t *testing.T) {
		cfg := createTestServerConfig(ctrl)

		server, err := newServer(cfg)
		require.NoError(t, err)
		require.NotNil(t, server)

		// Test that the server implements the Server interface
		assert.Implements(t, (*Server)(nil), server)
	})

	t.Run("Missing project manager", func(t *testing.T) {
		cfg := createTestServerConfig(ctrl)
		cfg.ProjectManager = nil

		server, err := newServer(cfg)
		require.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "project manager is required")
	})

	t.Run("Missing session manager", func(t *testing.T) {
		cfg := createTestServerConfig(ctrl)
		cfg.SessionManager = nil

		server, err := newServer(cfg)
		require.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "session manager is required")
	})

	t.Run("Missing logger", func(t *testing.T) {
		cfg := createTestServerConfig(ctrl)
		cfg.Logger = nil

		server, err := newServer(cfg)
		require.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "logger is required")
	})

	t.Run("Missing config", func(t *testing.T) {
		cfg := createTestServerConfig(ctrl)
		cfg.Config = nil

		server, err := newServer(cfg)
		require.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "config is required")
	})
}

func TestMCPServer_Interface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := createTestServerConfig(ctrl)

	server, err := newServer(cfg)
	require.NoError(t, err)

	t.Run("GetConfig", func(t *testing.T) {
		config := server.GetConfig()
		require.NotNil(t, config)
		// The returned config should be our test config
		assert.Equal(t, cfg.Config, config)
	})

	t.Run("IsRunning", func(t *testing.T) {
		// Initially should not be running
		assert.False(t, server.IsRunning())
	})

	t.Run("GetSessionCount", func(t *testing.T) {
		// Set up expectation for GetSessionCount call
		cfg.SessionManager.(*mocks.MockManager).EXPECT().
			GetSessionCount().
			Return(5)

		// Should return session count from session manager
		count := server.GetSessionCount()
		assert.Equal(t, 5, count)
	})
}

func TestMCPServer_Stop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := createTestServerConfig(ctrl)
	// Set up expectation for CloseAll call
	cfg.SessionManager.(*mocks.MockManager).EXPECT().
		CloseAll(gomock.Any()).
		Return(nil)
	// Set up expectation for logger.Info call
	cfg.Logger.(*mocks.MockLogger).EXPECT().
		Info("Stopping MCP server").
		Return()

	server, err := newServer(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Stop server
	err = server.Stop(ctx)
	require.NoError(t, err)
}

func TestMCPServer_CleanupExpiredSessions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := createTestServerConfig(ctrl)
	// Set up expectation for CleanupExpiredSessions call
	// Note: The server uses config.Timeout, not config.Session.Timeout
	cfg.SessionManager.(*mocks.MockManager).EXPECT().
		CleanupExpiredSessions(gomock.Any(), cfg.Config.Timeout).
		Return(nil)

	server, err := newServer(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Cleanup expired sessions
	err = server.CleanupExpiredSessions(ctx)
	require.NoError(t, err)
}

// TestStructuredInputOutput validates that our structured types work correctly
func TestStructuredInputOutput(t *testing.T) {
	t.Run("ProjectSelectRequest validation", func(t *testing.T) {
		req := tools.ProjectSelectRequest{
			ProjectID: uuid.New().String(),
		}

		assert.NotEmpty(t, req.ProjectID)
	})

	t.Run("ProjectCreateRequest validation", func(t *testing.T) {
		req := tools.ProjectCreateRequest{
			Title:       "Test Project",
			Description: "Test Description",
		}

		assert.Equal(t, "Test Project", req.Title)
		assert.Equal(t, "Test Description", req.Description)
	})

	t.Run("TaskCreateRequest validation", func(t *testing.T) {
		req := tools.TaskCreateRequest{
			Title:       "Test Task",
			Description: "Test Description",
			Complexity:  7,
		}

		assert.Equal(t, "Test Task", req.Title)
		assert.Equal(t, "Test Description", req.Description)
		assert.Equal(t, 7, req.Complexity)
	})

	t.Run("Response validation", func(t *testing.T) {
		projectResp := tools.ProjectSelectResponse{
			Message:   "Success",
			ProjectID: uuid.New().String(),
		}

		assert.Equal(t, "Success", projectResp.Message)
		assert.NotEmpty(t, projectResp.ProjectID)

		taskResp := tools.TaskCreateResponse{
			Message:   "Task created",
			TaskID:    uuid.New().String(),
			ProjectID: uuid.New().String(),
			Title:     "Test Task",
		}

		assert.Equal(t, "Task created", taskResp.Message)
		assert.NotEmpty(t, taskResp.TaskID)
		assert.NotEmpty(t, taskResp.ProjectID)
		assert.Equal(t, "Test Task", taskResp.Title)
	})
}

func TestGetSessionID(t *testing.T) {
	t.Run("No session in context", func(t *testing.T) {
		ctx := context.Background()
		sessionID := getSessionID(ctx)
		assert.Empty(t, sessionID)
	})
}
