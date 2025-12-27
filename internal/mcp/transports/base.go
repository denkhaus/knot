package transports

import (
	"context"
	"fmt"
	"sync"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/hints"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/mark3labs/mcp-go/server"
)

// BaseTransport provides common functionality for all transport implementations
type BaseTransport struct {
	transportType config.TransportType
	mcpServer     *server.MCPServer
	projectManager manager.ProjectManager
	sessionManager session.SessionManager
	logger         logger.Logger
	hintIntegration hints.Integration
	serverConfig   *config.MCPConfig
	running        bool
	mu             sync.RWMutex
	cancelFunc     context.CancelFunc
}

// NewBaseTransport creates a new base transport with direct dependencies
func NewBaseTransport(
	transportType config.TransportType,
	mcpServer *server.MCPServer,
	projectManager manager.ProjectManager,
	sessionManager session.SessionManager,
	logger logger.Logger,
	hintIntegration hints.Integration,
	serverConfig *config.MCPConfig,
) *BaseTransport {
	return &BaseTransport{
		transportType:  transportType,
		mcpServer:      mcpServer,
		projectManager: projectManager,
		sessionManager: sessionManager,
		logger:         logger,
		hintIntegration: hintIntegration,
		serverConfig:   serverConfig,
		running:        false,
	}
}

// GetType returns the transport type
func (b *BaseTransport) GetType() config.TransportType {
	return b.transportType
}

// IsRunning returns true if the transport is currently running
func (b *BaseTransport) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running
}

// setRunning sets the running state
func (b *BaseTransport) setRunning(running bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.running = running
}

// setCancelFunc sets the cancellation function
func (b *BaseTransport) setCancelFunc(cancel context.CancelFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cancelFunc = cancel
}

// getCancelFunc returns the cancellation function
func (b *BaseTransport) getCancelFunc() context.CancelFunc {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.cancelFunc
}

// Logger returns the logger instance
func (b *BaseTransport) Logger() logger.Logger {
	return b.logger
}

// Stop provides a common stop implementation that can be used by concrete transports
func (b *BaseTransport) Stop(ctx context.Context) error {
	b.logger.Info("Stopping transport",
		logger.String("type", b.transportType.String()),
	)

	b.setRunning(false)

	// Cancel the context if available
	if cancel := b.getCancelFunc(); cancel != nil {
		cancel()
	}

	// Close all sessions
	if err := b.sessionManager.CloseAll(ctx); err != nil {
		b.logger.Error("Error closing sessions during transport stop",
			logger.Error(err),
		)
		return fmt.Errorf("failed to close sessions: %w", err)
	}

	b.logger.Info("Transport stopped successfully",
		logger.String("type", b.transportType.String()),
	)

	return nil
}

// WithContext creates a context with cancellation for the transport
func (b *BaseTransport) WithContext(parent context.Context) context.Context {
	ctx, cancel := context.WithCancel(parent)
	b.setCancelFunc(cancel)
	return ctx
}