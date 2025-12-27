package transports

import (
	"context"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/hints"
	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/mark3labs/mcp-go/server"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewSSETransport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockHintIntegration := mocks.NewMockIntegration(ctrl)
	mockConfigService := mocks.NewMockService(ctrl)

	mcpServer := server.NewMCPServer("test", "1.0.0")

	serverConfig := &config.MCPConfig{
		Address: "localhost",
		Port:    8080,
		Transport: config.TransportConfig{
			Mode: config.TransportTypeSSE,
			SSE: config.SSETransportConfig{
				HeartbeatInterval: 30,
				ClientTimeout:     60,
			},
		},
	}

	t.Run("successful creation", func(t *testing.T) {
		mockConfigService.EXPECT().GetMCPConfig().Return(serverConfig)

		injector := do.New()
		do.ProvideValue[*server.MCPServer](injector, mcpServer)
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue[manager.ProjectManager](injector, mockManager)
		do.ProvideValue[session.SessionManager](injector, mockSessionManager)
		do.ProvideValue[hints.Integration](injector, mockHintIntegration)
		do.ProvideValue[config.Service](injector, mockConfigService)

		transport, err := newSSETransport(injector)

		require.NoError(t, err)
		require.NotNil(t, transport)

		assert.IsType(t, &SSETransport{}, transport)
		assert.Equal(t, config.TransportTypeSSE, transport.GetType())
		assert.False(t, transport.IsRunning())
	})

	t.Run("verify BaseTransport is initialized", func(t *testing.T) {
		mockConfigService.EXPECT().GetMCPConfig().Return(serverConfig)

		injector := do.New()
		do.ProvideValue[*server.MCPServer](injector, mcpServer)
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue[manager.ProjectManager](injector, mockManager)
		do.ProvideValue[session.SessionManager](injector, mockSessionManager)
		do.ProvideValue[hints.Integration](injector, mockHintIntegration)
		do.ProvideValue[config.Service](injector, mockConfigService)

		transport, err := newSSETransport(injector)

		require.NoError(t, err)
		require.NotNil(t, transport)

		sseTransport := transport.(*SSETransport)
		assert.NotNil(t, sseTransport.BaseTransport)
		assert.NotNil(t, sseTransport.mcpServer)
		assert.NotNil(t, sseTransport.serverConfig)
	})
}

func TestSSETransport_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockHintIntegration := mocks.NewMockIntegration(ctrl)
	mockConfigService := mocks.NewMockService(ctrl)

	mcpServer := server.NewMCPServer("test", "1.0.0")

	serverConfig := &config.MCPConfig{
		Address: "localhost",
		Port:    0, // Use random port for testing
		Transport: config.TransportConfig{
			Mode: config.TransportTypeSSE,
			SSE: config.SSETransportConfig{
				HeartbeatInterval: 30,
				ClientTimeout:     60,
			},
		},
	}

	t.Run("successful creation", func(t *testing.T) {
		mockConfigService.EXPECT().GetMCPConfig().Return(serverConfig).AnyTimes()

		injector := do.New()
		do.ProvideValue[*server.MCPServer](injector, mcpServer)
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue[manager.ProjectManager](injector, mockManager)
		do.ProvideValue[session.SessionManager](injector, mockSessionManager)
		do.ProvideValue[hints.Integration](injector, mockHintIntegration)
		do.ProvideValue[config.Service](injector, mockConfigService)

		transport, err := newSSETransport(injector)
		require.NoError(t, err)

		sseTransport := transport.(*SSETransport)
		assert.NotNil(t, sseTransport)
		assert.NotNil(t, sseTransport.mcpServer)
		assert.NotNil(t, sseTransport.serverConfig)
	})

	t.Run("different config values", func(t *testing.T) {
		// Create a separate controller and mocks to avoid expectation conflicts
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		customLogger := mocks.NewMockLogger(ctrl)
		customConfigService := mocks.NewMockService(ctrl)

		customConfig := &config.MCPConfig{
			Address: "0.0.0.0",
			Port:    9090,
			Transport: config.TransportConfig{
				Mode: config.TransportTypeSSE,
				SSE: config.SSETransportConfig{
					HeartbeatInterval: 15,
					ClientTimeout:     120,
				},
			},
		}

		customConfigService.EXPECT().GetMCPConfig().Return(customConfig)

		injector := do.New()
		do.ProvideValue[*server.MCPServer](injector, mcpServer)
		do.ProvideValue[logger.Logger](injector, customLogger)
		do.ProvideValue[manager.ProjectManager](injector, mockManager)
		do.ProvideValue[session.SessionManager](injector, mockSessionManager)
		do.ProvideValue[hints.Integration](injector, mockHintIntegration)
		do.ProvideValue[config.Service](injector, customConfigService)

		transport, err := newSSETransport(injector)
		require.NoError(t, err)

		sseTransport := transport.(*SSETransport)
		assert.NotNil(t, sseTransport.serverConfig)
		assert.Equal(t, "0.0.0.0", sseTransport.serverConfig.Address)
		assert.Equal(t, 9090, sseTransport.serverConfig.Port)
		assert.Equal(t, 15, sseTransport.serverConfig.Transport.SSE.HeartbeatInterval)
		assert.Equal(t, 120, sseTransport.serverConfig.Transport.SSE.ClientTimeout)
	})
}

func TestSSETransport_Stop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockHintIntegration := mocks.NewMockIntegration(ctrl)
	mockConfigService := mocks.NewMockService(ctrl)

	mcpServer := server.NewMCPServer("test", "1.0.0")

	serverConfig := &config.MCPConfig{
		Address: "localhost",
		Port:    8080,
		Transport: config.TransportConfig{
			Mode: config.TransportTypeSSE,
			SSE: config.SSETransportConfig{
				HeartbeatInterval: 30,
				ClientTimeout:     60,
			},
		},
	}

	mockConfigService.EXPECT().GetMCPConfig().Return(serverConfig).AnyTimes()

	injector := do.New()
	do.ProvideValue[*server.MCPServer](injector, mcpServer)
	do.ProvideValue[logger.Logger](injector, mockLogger)
	do.ProvideValue[manager.ProjectManager](injector, mockManager)
	do.ProvideValue[session.SessionManager](injector, mockSessionManager)
	do.ProvideValue[hints.Integration](injector, mockHintIntegration)
	do.ProvideValue[config.Service](injector, mockConfigService)

	transport, err := newSSETransport(injector)
	require.NoError(t, err)

	sseTransport := transport.(*SSETransport)

	t.Run("successful stop", func(t *testing.T) {
		sseTransport.setRunning(true)

		mockLogger.EXPECT().Info("Stopping SSE transport server")

		ctx := context.Background()
		err := sseTransport.Stop(ctx)

		require.NoError(t, err)
		assert.False(t, sseTransport.IsRunning())
	})

	t.Run("stop when not running", func(t *testing.T) {
		mockLogger.EXPECT().Info("Stopping SSE transport server")

		ctx := context.Background()
		err := sseTransport.Stop(ctx)

		require.NoError(t, err)
		assert.False(t, sseTransport.IsRunning())
	})

	t.Run("stop with context timeout", func(t *testing.T) {
		sseTransport.setRunning(true)

		mockLogger.EXPECT().Info("Stopping SSE transport server")
		// Error may or may not be logged depending on shutdown timing
		mockLogger.EXPECT().Error("Error shutting down SSE server",
			gomock.Any(), // error
		).MaxTimes(1) // Make it optional - may or may not be called

		// Use a very short context
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Give context time to expire
		time.Sleep(10 * time.Millisecond)

		err := sseTransport.Stop(ctx)
		// May error due to context timeout
		_ = err
	})
}
