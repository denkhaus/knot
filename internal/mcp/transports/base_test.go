package transports

import (
	"context"
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

func TestNewBaseTransport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockHintIntegration := mocks.NewMockIntegration(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")
	serverConfig := &config.MCPConfig{
		Address: "localhost",
		Port:    8080,
		Transport: config.TransportConfig{
			Mode: config.TransportTypeHTTP,
		},
	}

	t.Run("successful creation", func(t *testing.T) {
		base := NewBaseTransport(
			config.TransportTypeHTTP,
			mcpServer,
			mockManager,
			mockSessionManager,
			mockLogger,
			mockHintIntegration,
			serverConfig,
		)

		require.NotNil(t, base)
		assert.Equal(t, config.TransportTypeHTTP, base.GetType())
		assert.False(t, base.IsRunning())
	})

	t.Run("different transport types", func(t *testing.T) {
		tests := []struct {
			name          string
			transportType config.TransportType
		}{
			{"stdio", config.TransportTypeStdio},
			{"http", config.TransportTypeHTTP},
			{"sse", config.TransportTypeSSE},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				base := NewBaseTransport(
					tt.transportType,
					mcpServer,
					mockManager,
					mockSessionManager,
					mockLogger,
					mockHintIntegration,
					serverConfig,
				)

				assert.Equal(t, tt.transportType, base.GetType())
			})
		}
	})
}

func TestBaseTransport_IsRunning(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockHintIntegration := mocks.NewMockIntegration(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")
	serverConfig := &config.MCPConfig{}

	base := NewBaseTransport(
		config.TransportTypeHTTP,
		mcpServer,
		mockManager,
		mockSessionManager,
		mockLogger,
		mockHintIntegration,
		serverConfig,
	)

	t.Run("initially not running", func(t *testing.T) {
		assert.False(t, base.IsRunning())
	})

	t.Run("set running state", func(t *testing.T) {
		base.setRunning(true)
		assert.True(t, base.IsRunning())

		base.setRunning(false)
		assert.False(t, base.IsRunning())
	})

	t.Run("concurrent access", func(t *testing.T) {
		// Test concurrent reads
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				_ = base.IsRunning()
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestBaseTransport_Logger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockHintIntegration := mocks.NewMockIntegration(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")
	serverConfig := &config.MCPConfig{}

	base := NewBaseTransport(
		config.TransportTypeHTTP,
		mcpServer,
		mockManager,
		mockSessionManager,
		mockLogger,
		mockHintIntegration,
		serverConfig,
	)

	assert.Same(t, mockLogger, base.Logger())
}

func TestBaseTransport_WithContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockHintIntegration := mocks.NewMockIntegration(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")
	serverConfig := &config.MCPConfig{}

	base := NewBaseTransport(
		config.TransportTypeHTTP,
		mcpServer,
		mockManager,
		mockSessionManager,
		mockLogger,
		mockHintIntegration,
		serverConfig,
	)

	t.Run("creates cancellable context", func(t *testing.T) {
		parentCtx := context.Background()
		ctx := base.WithContext(parentCtx)

		require.NotNil(t, ctx)
		assert.NotNil(t, base.getCancelFunc())

		// Cancel and verify
		cancel := base.getCancelFunc()
		cancel()

		select {
		case <-ctx.Done():
			// Context was cancelled as expected
		default:
			t.Error("context should be cancelled")
		}
	})
}

func TestBaseTransport_Stop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockHintIntegration := mocks.NewMockIntegration(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")
	serverConfig := &config.MCPConfig{}

	base := NewBaseTransport(
		config.TransportTypeHTTP,
		mcpServer,
		mockManager,
		mockSessionManager,
		mockLogger,
		mockHintIntegration,
		serverConfig,
	)

	t.Run("successful stop", func(t *testing.T) {
		base.setRunning(true)

		// Set up expectations
		mockLogger.EXPECT().Info("Stopping transport",
			logger.String("type", config.TransportTypeHTTP.String()),
		)
		mockSessionManager.EXPECT().CloseAll(gomock.Any()).Return(nil)
		mockLogger.EXPECT().Info("Transport stopped successfully",
			logger.String("type", config.TransportTypeHTTP.String()),
		)

		ctx := context.Background()
		err := base.Stop(ctx)

		require.NoError(t, err)
		assert.False(t, base.IsRunning())
	})

	t.Run("stop with session manager error", func(t *testing.T) {
		base.setRunning(true)

		// Set up expectations
		mockLogger.EXPECT().Info("Stopping transport",
			logger.String("type", config.TransportTypeHTTP.String()),
		)
		mockSessionManager.EXPECT().CloseAll(gomock.Any()).Return(
			assert.AnError)
		mockLogger.EXPECT().Error("Error closing sessions during transport stop",
			logger.Error(assert.AnError),
		)

		ctx := context.Background()
		err := base.Stop(ctx)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to close sessions")
		assert.False(t, base.IsRunning())
	})

	t.Run("stop with cancel function", func(t *testing.T) {
		base.setRunning(true)

		// Create a cancel function
		parentCtx, cancel := context.WithCancel(context.Background())
		base.setCancelFunc(cancel)

		// Set up expectations
		mockLogger.EXPECT().Info("Stopping transport",
			logger.String("type", config.TransportTypeHTTP.String()),
		)
		mockSessionManager.EXPECT().CloseAll(gomock.Any()).Return(nil)
		mockLogger.EXPECT().Info("Transport stopped successfully",
			logger.String("type", config.TransportTypeHTTP.String()),
		)

		ctx := context.Background()
		err := base.Stop(ctx)

		require.NoError(t, err)

		// Verify parent context was cancelled
		select {
		case <-parentCtx.Done():
			// Context was cancelled as expected
		default:
			t.Error("parent context should be cancelled")
		}
	})
}

func TestBaseTransport_ConcurrentAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockHintIntegration := mocks.NewMockIntegration(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")
	serverConfig := &config.MCPConfig{}

	base := NewBaseTransport(
		config.TransportTypeHTTP,
		mcpServer,
		mockManager,
		mockSessionManager,
		mockLogger,
		mockHintIntegration,
		serverConfig,
	)

	t.Run("concurrent state changes", func(t *testing.T) {
		done := make(chan bool)

		// Start multiple goroutines that modify state
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					base.setRunning(true)
					base.setRunning(false)
				}
				done <- true
			}()
		}

		// Start multiple goroutines that read state
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					_ = base.IsRunning()
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 20; i++ {
			<-done
		}

		// Final state should be consistent
		_ = base.IsRunning() // Should not panic or race
	})
}

func TestBaseTransport_CancelFunc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockHintIntegration := mocks.NewMockIntegration(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")
	serverConfig := &config.MCPConfig{}

	base := NewBaseTransport(
		config.TransportTypeHTTP,
		mcpServer,
		mockManager,
		mockSessionManager,
		mockLogger,
		mockHintIntegration,
		serverConfig,
	)

	t.Run("set and get cancel func", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		base.setCancelFunc(cancel)
		retrieved := base.getCancelFunc()

		// Note: Can't use Same() with functions, just verify they work
		// by calling the retrieved func
		retrieved()

		select {
		case <-ctx.Done():
			// Context was cancelled as expected
		default:
			t.Error("context should be cancelled")
		}
	})

	t.Run("cancel func is callable", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		base.setCancelFunc(cancel)

		// Get and call the cancel func
		cf := base.getCancelFunc()
		if cf != nil {
			cf()
		}

		select {
		case <-ctx.Done():
			// Context was cancelled
		default:
			t.Error("context should be cancelled")
		}
	})
}

func TestNewTransport(t *testing.T) {
	// Allow any number of log calls for all tests

	t.Run("creates stdio transport", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockConfigService := mocks.NewMockService(ctrl)
		mockManager := mocks.NewMockProjectManager(ctrl)
		mockSessionManager := mocks.NewMockSessionManager(ctrl)
		mockHintIntegration := mocks.NewMockIntegration(ctrl)
		mcpServer := server.NewMCPServer("test", "1.0.0")

		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

		serverConfig := &config.MCPConfig{
			Address: "localhost",
			Port:    8080,
			Transport: config.TransportConfig{
				Mode: config.TransportTypeStdio,
			},
		}

		mockConfigService.EXPECT().GetMCPConfig().Return(serverConfig).AnyTimes()

		injector := do.New()
		do.ProvideValue[config.Service](injector, mockConfigService)
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue(injector, mcpServer)
		do.ProvideValue[manager.ProjectManager](injector, mockManager)
		do.ProvideValue[session.SessionManager](injector, mockSessionManager)
		do.ProvideValue[hints.Integration](injector, mockHintIntegration)

		transport, err := NewTransport(injector)

		require.NoError(t, err)
		require.NotNil(t, transport)
		assert.Equal(t, config.TransportTypeStdio, transport.GetType())
	})

	t.Run("creates sse transport", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockConfigService := mocks.NewMockService(ctrl)
		mockManager := mocks.NewMockProjectManager(ctrl)
		mockSessionManager := mocks.NewMockSessionManager(ctrl)
		mockHintIntegration := mocks.NewMockIntegration(ctrl)
		mcpServer := server.NewMCPServer("test", "1.0.0")

		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

		serverConfig := &config.MCPConfig{
			Address: "localhost",
			Port:    8080,
			Transport: config.TransportConfig{
				Mode: config.TransportTypeSSE,
			},
		}

		mockConfigService.EXPECT().GetMCPConfig().Return(serverConfig).AnyTimes()

		injector := do.New()
		do.ProvideValue[config.Service](injector, mockConfigService)
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue(injector, mcpServer)
		do.ProvideValue[manager.ProjectManager](injector, mockManager)
		do.ProvideValue[session.SessionManager](injector, mockSessionManager)
		do.ProvideValue[hints.Integration](injector, mockHintIntegration)

		transport, err := NewTransport(injector)

		require.NoError(t, err)
		require.NotNil(t, transport)
		assert.Equal(t, config.TransportTypeSSE, transport.GetType())
	})

	t.Run("returns error for invalid transport type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigService := mocks.NewMockService(ctrl)
		mockLogger := mocks.NewMockLogger(ctrl)
		mockManager := mocks.NewMockProjectManager(ctrl)
		mockSessionManager := mocks.NewMockSessionManager(ctrl)
		mockHintIntegration := mocks.NewMockIntegration(ctrl)
		mcpServer := server.NewMCPServer("test", "1.0.0")

		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

		serverConfig := &config.MCPConfig{
			Address: "localhost",
			Port:    8080,
			Transport: config.TransportConfig{
				Mode: config.TransportType("invalid"),
			},
		}

		mockConfigService.EXPECT().GetMCPConfig().Return(serverConfig)

		injector := do.New()
		do.ProvideValue[config.Service](injector, mockConfigService)
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue(injector, mcpServer)
		do.ProvideValue[manager.ProjectManager](injector, mockManager)
		do.ProvideValue[session.SessionManager](injector, mockSessionManager)
		do.ProvideValue[hints.Integration](injector, mockHintIntegration)

		transport, err := NewTransport(injector)

		assert.Error(t, err)
		assert.Nil(t, transport)
		assert.Contains(t, err.Error(), "invalid transport type")
	})
}
