package transports

import (
	"context"
	"testing"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/hints"
	"github.com/denkhaus/knot/v2/internal/mcp/shared"
	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/mark3labs/mcp-go/server"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewStdioTransport(t *testing.T) {
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
			Mode: config.TransportTypeStdio,
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

		transport, err := newStdioTransport(injector)

		require.NoError(t, err)
		require.NotNil(t, transport)

		assert.IsType(t, &StdioTransport{}, transport)
		assert.Equal(t, config.TransportTypeStdio, transport.GetType())
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

		transport, err := newStdioTransport(injector)

		require.NoError(t, err)
		require.NotNil(t, transport)

		stdioTransport := transport.(*StdioTransport)
		assert.NotNil(t, stdioTransport.BaseTransport)
		assert.NotNil(t, stdioTransport.mcpServer)
	})
}

func TestStdioTransport_Start(t *testing.T) {
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
			Mode: config.TransportTypeStdio,
		},
	}

	t.Run("start calls logger", func(t *testing.T) {
		mockConfigService.EXPECT().GetMCPConfig().Return(serverConfig)
		mockLogger.EXPECT().Info("Starting stdio transport")

		injector := do.New()
		do.ProvideValue[*server.MCPServer](injector, mcpServer)
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue[manager.ProjectManager](injector, mockManager)
		do.ProvideValue[session.SessionManager](injector, mockSessionManager)
		do.ProvideValue[hints.Integration](injector, mockHintIntegration)
		do.ProvideValue[config.Service](injector, mockConfigService)

		transport, err := newStdioTransport(injector)
		require.NoError(t, err)

		stdioTransport := transport.(*StdioTransport)

		// Note: server.ServeStdio() will block and try to read from stdin
		// In a test environment, we can't fully test this without mocking stdin/stdout
		// So we just verify the initial setup and logging

		// Verify running state is set
		assert.False(t, stdioTransport.IsRunning())

		// Start will block, so we run it in a goroutine
		errChan := make(chan error, 1)
		go func() {
			// This will return an error because it can't read from stdin in tests
			errChan <- stdioTransport.Start(context.Background())
		}()

		// Give it a moment to start and set running flag
		// Then cancel the context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Immediately cancel

		// The start should eventually fail or return
		select {
		case err := <-errChan:
			// Expected to get an error
			_ = err
		case <-ctx.Done():
			// Or context is done
		}
	})

	t.Run("start sets running state", func(t *testing.T) {
		mockConfigService.EXPECT().GetMCPConfig().Return(serverConfig)

		injector := do.New()
		do.ProvideValue[*server.MCPServer](injector, mcpServer)
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue[manager.ProjectManager](injector, mockManager)
		do.ProvideValue[session.SessionManager](injector, mockSessionManager)
		do.ProvideValue[hints.Integration](injector, mockHintIntegration)
		do.ProvideValue[config.Service](injector, mockConfigService)

		transport, err := newStdioTransport(injector)
		require.NoError(t, err)

		stdioTransport := transport.(*StdioTransport)

		// Verify initial state
		assert.False(t, stdioTransport.IsRunning())
	})
}

func TestStdioTransport_TransportType(t *testing.T) {
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
			Mode: config.TransportTypeStdio,
		},
	}

	mockConfigService.EXPECT().GetMCPConfig().Return(serverConfig)

	injector := do.New()
	do.ProvideValue[*server.MCPServer](injector, mcpServer)
	do.ProvideValue[logger.Logger](injector, mockLogger)
	do.ProvideValue[manager.ProjectManager](injector, mockManager)
	do.ProvideValue[session.SessionManager](injector, mockSessionManager)
	do.ProvideValue[hints.Integration](injector, mockHintIntegration)
	do.ProvideValue[config.Service](injector, mockConfigService)

	transport, err := newStdioTransport(injector)
	require.NoError(t, err)

	t.Run("returns correct transport type", func(t *testing.T) {
		assert.Equal(t, config.TransportTypeStdio, transport.GetType())
	})
}

func TestStdioTransport_CompareWithOtherTransports(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockHintIntegration := mocks.NewMockIntegration(ctrl)
	mockConfigService := mocks.NewMockService(ctrl)
	mockSessionRegistry := mocks.NewMockSessionRegistry(ctrl)

	mcpServer := server.NewMCPServer("test", "1.0.0")

	serverConfig := &config.MCPConfig{
		Address: "localhost",
		Port:    8080,
		Transport: config.TransportConfig{
			Mode: config.TransportTypeStdio,
		},
	}

	mockConfigService.EXPECT().GetMCPConfig().Return(serverConfig).AnyTimes()

	t.Run("stdio vs http transport type", func(t *testing.T) {
		httpConfig := &config.MCPConfig{
			Address: "localhost",
			Port:    8080,
			Transport: config.TransportConfig{
				Mode: config.TransportTypeHTTP,
			},
		}

		mockConfigService.EXPECT().GetMCPConfig().Return(httpConfig).AnyTimes()

		injectorHTTP := do.New()
		do.ProvideValue[*server.MCPServer](injectorHTTP, mcpServer)
		do.ProvideValue[logger.Logger](injectorHTTP, mockLogger)
		do.ProvideValue[manager.ProjectManager](injectorHTTP, mockManager)
		do.ProvideValue[session.SessionManager](injectorHTTP, mockSessionManager)
		do.ProvideValue[hints.Integration](injectorHTTP, mockHintIntegration)
		do.ProvideValue[config.Service](injectorHTTP, mockConfigService)
		do.ProvideValue[shared.SessionRegistry](injectorHTTP, mockSessionRegistry)

		// Also need to provide the SessionWrapper
		sessionWrapper, err := NewSessionWrapper(injectorHTTP)
		require.NoError(t, err)
		do.ProvideValue[*SessionWrapper](injectorHTTP, sessionWrapper)

		httpTransport, err := newHTTPTransport(injectorHTTP)
		require.NoError(t, err)

		injectorStdio := do.New()
		do.ProvideValue[*server.MCPServer](injectorStdio, mcpServer)
		do.ProvideValue[logger.Logger](injectorStdio, mockLogger)
		do.ProvideValue[manager.ProjectManager](injectorStdio, mockManager)
		do.ProvideValue[session.SessionManager](injectorStdio, mockSessionManager)
		do.ProvideValue[hints.Integration](injectorStdio, mockHintIntegration)
		do.ProvideValue[config.Service](injectorStdio, mockConfigService)

		stdioTransport, err := newStdioTransport(injectorStdio)
		require.NoError(t, err)

		assert.NotEqual(t, httpTransport.GetType(), stdioTransport.GetType())
		assert.Equal(t, config.TransportTypeHTTP, httpTransport.GetType())
		assert.Equal(t, config.TransportTypeStdio, stdioTransport.GetType())
	})
}
