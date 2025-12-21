package transports

import (
	"context"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/hints"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/mark3labs/mcp-go/server"
	"github.com/samber/do/v2"
)

// StdioTransport implements the Transport interface for stdio communication
type StdioTransport struct {
	*BaseTransport
	mcpServer *server.MCPServer
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(deps TransportDependencies) *StdioTransport {
	return &StdioTransport{
		BaseTransport: NewBaseTransport(config.TransportTypeStdio, deps),
		mcpServer:     deps.MCPServer,
	}
}

// Start starts the stdio transport server
func (s *StdioTransport) Start(ctx context.Context) error {
	s.Logger().Info("Starting stdio transport")
	s.setRunning(true)

	// Use the existing mcp-go ServeStdio functionality
	return server.ServeStdio(s.mcpServer)
}

// NewStdioTransportProvider creates a stdio transport provider for DI
func NewStdioTransportProvider(injector do.Injector) (Transport, error) {
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

	return NewStdioTransport(deps), nil
}
