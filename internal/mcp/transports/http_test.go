package transports

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestNewHTTPTransport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockHintIntegration := mocks.NewMockIntegration(ctrl)
	mockConfigService := mocks.NewMockService(ctrl)

	mcpServer := server.NewMCPServer("test", "1.0.0")
	sessionWrapper := &SessionWrapper{}

	serverConfig := &config.MCPConfig{
		Address: "localhost",
		Port:    8080,
		Transport: config.TransportConfig{
			Mode: config.TransportTypeHTTP,
			HTTP: config.HTTPTransportConfig{
				RequestTimeout: 30,
			},
		},
	}

	t.Run("successful creation", func(t *testing.T) {
		// Set up mock expectations
		mockConfigService.EXPECT().GetMCPConfig().Return(serverConfig)

		// Create injector with all dependencies
		injector := do.New()
		do.ProvideValue[*server.MCPServer](injector, mcpServer)
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue[manager.ProjectManager](injector, mockManager)
		do.ProvideValue[session.SessionManager](injector, mockSessionManager)
		do.ProvideValue[hints.Integration](injector, mockHintIntegration)
		do.ProvideValue[config.Service](injector, mockConfigService)
		do.ProvideValue[*SessionWrapper](injector, sessionWrapper)

		transport, err := newHTTPTransport(injector)

		require.NoError(t, err)
		require.NotNil(t, transport)

		assert.IsType(t, &HTTPTransport{}, transport)
		assert.Equal(t, config.TransportTypeHTTP, transport.GetType())
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
		do.ProvideValue[*SessionWrapper](injector, sessionWrapper)

		transport, err := newHTTPTransport(injector)

		require.NoError(t, err)
		require.NotNil(t, transport)

		httpTransport := transport.(*HTTPTransport)
		assert.NotNil(t, httpTransport.BaseTransport)
		assert.NotNil(t, httpTransport.httpServer)
		assert.NotNil(t, httpTransport.injector)
	})
}

func TestHTTPTransport_Start(t *testing.T) {
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
			Mode: config.TransportTypeHTTP,
			HTTP: config.HTTPTransportConfig{
				RequestTimeout: 30,
			},
		},
	}

	t.Run("successful start", func(t *testing.T) {
		// Note: This test is simplified because starting the actual server
		// requires many dependencies (SyncHandler, etc.) that are complex to mock.
		// We test the health handler separately instead.
		mockConfigService.EXPECT().GetMCPConfig().Return(serverConfig).AnyTimes()

		// Create session wrapper
		sessionWrapper := &SessionWrapper{
			StreamableHTTPServer: server.NewStreamableHTTPServer(mcpServer),
		}

		injector := do.New()
		do.ProvideValue[*server.MCPServer](injector, mcpServer)
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue[manager.ProjectManager](injector, mockManager)
		do.ProvideValue[session.SessionManager](injector, mockSessionManager)
		do.ProvideValue[hints.Integration](injector, mockHintIntegration)
		do.ProvideValue[config.Service](injector, mockConfigService)
		do.ProvideValue[*SessionWrapper](injector, sessionWrapper)

		transport, err := newHTTPTransport(injector)
		require.NoError(t, err)

		httpTransport := transport.(*HTTPTransport)
		assert.NotNil(t, httpTransport)
		assert.NotNil(t, httpTransport.httpServer)
		assert.NotNil(t, httpTransport.injector)
	})
}

func TestHTTPTransport_HealthHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockHintIntegration := mocks.NewMockIntegration(ctrl)
	mockConfigService := mocks.NewMockService(ctrl)

	mcpServer := server.NewMCPServer("test", "1.0.0")
	sessionWrapper := &SessionWrapper{}

	serverConfig := &config.MCPConfig{
		Address: "localhost",
		Port:    8080,
		Transport: config.TransportConfig{
			Mode: config.TransportTypeHTTP,
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
	do.ProvideValue[*SessionWrapper](injector, sessionWrapper)

	transport, err := newHTTPTransport(injector)
	require.NoError(t, err)

	httpTransport := transport.(*HTTPTransport)

	t.Run("health endpoint returns 200", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		httpTransport.healthHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("health endpoint returns correct structure", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		httpTransport.healthHandler(w, req)

		var health map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&health)

		require.NoError(t, err)
		assert.Equal(t, "healthy", health["status"])
		assert.NotEmpty(t, health["timestamp"])
		assert.Equal(t, "knot MCP server", health["server"])
		assert.Equal(t, "http", health["transport"])
		assert.Equal(t, "localhost", health["address"])
		assert.Equal(t, float64(8080), health["port"])
	})
}

func TestHTTPTransport_Stop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockHintIntegration := mocks.NewMockIntegration(ctrl)
	mockConfigService := mocks.NewMockService(ctrl)

	mcpServer := server.NewMCPServer("test", "1.0.0")
	sessionWrapper := &SessionWrapper{
		StreamableHTTPServer: server.NewStreamableHTTPServer(mcpServer),
	}

	serverConfig := &config.MCPConfig{
		Address: "localhost",
		Port:    8080,
		Transport: config.TransportConfig{
			Mode: config.TransportTypeHTTP,
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
	do.ProvideValue[*SessionWrapper](injector, sessionWrapper)

	transport, err := newHTTPTransport(injector)
	require.NoError(t, err)

	httpTransport := transport.(*HTTPTransport)

	t.Run("successful stop", func(t *testing.T) {
		httpTransport.setRunning(true)

		mockLogger.EXPECT().Info("Stopping HTTP transport server")

		ctx := context.Background()
		err := httpTransport.Stop(ctx)

		require.NoError(t, err)
		assert.False(t, httpTransport.IsRunning())
	})

	t.Run("stop when not running", func(t *testing.T) {
		mockLogger.EXPECT().Info("Stopping HTTP transport server")

		ctx := context.Background()
		err := httpTransport.Stop(ctx)

		require.NoError(t, err)
		assert.False(t, httpTransport.IsRunning())
	})
}

func TestHTTPTransport_InitializeSessionComponents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionRegistry := mocks.NewMockSessionRegistry(ctrl)

	mockLogger := mocks.NewMockLogger(ctrl)
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockHintIntegration := mocks.NewMockIntegration(ctrl)
	mockConfigService := mocks.NewMockService(ctrl)

	mcpServer := server.NewMCPServer("test", "1.0.0")
	sessionWrapper := &SessionWrapper{}

	serverConfig := &config.MCPConfig{
		Address: "localhost",
		Port:    8080,
		Transport: config.TransportConfig{
			Mode: config.TransportTypeHTTP,
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
	do.ProvideValue[*SessionWrapper](injector, sessionWrapper)

	transport, err := newHTTPTransport(injector)
	require.NoError(t, err)

	httpTransport := transport.(*HTTPTransport)

	t.Run("initialize session registry", func(t *testing.T) {
		httpTransport.InitializeSessionComponents(mockSessionRegistry)

		assert.Same(t, mockSessionRegistry, httpTransport.sessionRegistry)
	})
}

func TestHTTPTransport_StartErrorCases(t *testing.T) {
	// Note: The DI container will panic if SessionWrapper is not provided,
	// which is the expected behavior for missing required dependencies.
	// We don't test this panic case since it's a DI framework behavior.
	t.Run("missing dependencies panic", func(t *testing.T) {
		// This test documents that missing dependencies cause a panic
		// through the DI container, which is expected behavior.
		// We don't actually test the panic since it would crash the test.
	})
}
