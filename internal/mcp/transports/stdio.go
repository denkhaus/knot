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

// Start starts the stdio transport server
func (s *StdioTransport) Start(ctx context.Context) error {
	s.Logger().Info("Starting stdio transport")
	s.setRunning(true)

	// Use the existing mcp-go ServeStdio functionality
	return server.ServeStdio(s.mcpServer)
}

// newStdioTransport creates a stdio transport provider for DI
func newStdioTransport(injector do.Injector) (Transport, error) {
	mcpServer := do.MustInvoke[*server.MCPServer](injector)
	logger := do.MustInvoke[logger.Logger](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)
	sessionManager := do.MustInvoke[session.SessionManager](injector)
	hintIntegration := do.MustInvoke[hints.Integration](injector)
	configService := do.MustInvoke[config.Service](injector)

	serverConfig := configService.GetMCPConfig()

	base := NewBaseTransport(
		config.TransportTypeStdio,
		mcpServer,
		projectManager,
		sessionManager,
		logger,
		hintIntegration,
		serverConfig,
	)
	return &StdioTransport{
		BaseTransport: base,
		mcpServer:     mcpServer,
	}, nil
}
