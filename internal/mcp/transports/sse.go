package transports

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/hints"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/mark3labs/mcp-go/server"
	"github.com/samber/do/v2"
)

// SSETransport implements the Transport interface for Server-Sent Events
type SSETransport struct {
	*BaseTransport
	sseServer *server.SSEServer
}

// NewSSETransport creates a new SSE transport
func NewSSETransport(deps TransportDependencies) *SSETransport {
	return &SSETransport{
		BaseTransport: NewBaseTransport(config.TransportTypeSSE, deps),
	}
}

// Start starts the SSE transport server using mcp-go's built-in SSE server
func (s *SSETransport) Start(ctx context.Context) error {
	deps := s.Dependencies()
	addr := fmt.Sprintf("%s:%d", deps.ServerConfig.Address, deps.ServerConfig.Port)

	// Create HTTP server with timeout configuration
	httpTimeout := time.Duration(deps.ServerConfig.Transport.SSE.ClientTimeout) * time.Second
	customHTTPServer := &http.Server{
		Addr:         addr,
		ReadTimeout:  httpTimeout,
		WriteTimeout: httpTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// Create SSE server with configuration options
	heartbeatInterval := time.Duration(deps.ServerConfig.Transport.SSE.HeartbeatInterval) * time.Second
	s.sseServer = server.NewSSEServer(
		deps.MCPServer,
		server.WithHTTPServer(customHTTPServer),
		server.WithKeepAlive(true),
		server.WithKeepAliveInterval(heartbeatInterval),
	)

	s.Logger().Info("Starting SSE transport server",
		logger.String("address", addr),
		logger.Int("heartbeat_interval_seconds", deps.ServerConfig.Transport.SSE.HeartbeatInterval),
		logger.Int("client_timeout_seconds", deps.ServerConfig.Transport.SSE.ClientTimeout),
	)

	// Start the SSE server and block
	s.setRunning(true)
	if err := s.sseServer.Start(addr); err != nil {
		s.setRunning(false)
		return fmt.Errorf("SSE server failed to start: %w", err)
	}

	return nil
}

// Stop gracefully stops the SSE transport server
func (s *SSETransport) Stop(ctx context.Context) error {
	s.Logger().Info("Stopping SSE transport server")

	if s.sseServer != nil {
		if err := s.sseServer.Shutdown(ctx); err != nil {
			s.Logger().Error("Error shutting down SSE server", logger.String("error", err.Error()))
			return err
		}
	}

	s.setRunning(false)
	return nil
}

// NewSSETransportProvider creates an SSE transport provider for DI
func NewSSETransportProvider(injector do.Injector) (Transport, error) {
	mcpServer := do.MustInvoke[*server.MCPServer](injector)
	logger := do.MustInvoke[logger.Logger](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)
	sessionManager := do.MustInvoke[session.SessionManager](injector)
	hintIntegration := do.MustInvoke[hints.Integration](injector)
	configService := do.MustInvoke[config.Service](injector)

	// Create transport dependencies
	deps := TransportDependencies{
		MCPServer:       mcpServer,
		ProjectManager:  projectManager,
		SessionManager:  sessionManager,
		Logger:          logger,
		HintIntegration: hintIntegration,
		ServerConfig:    configService.GetMCPConfig(),
	}

	return NewSSETransport(deps), nil
}
