package transports

import (
	"context"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/mark3labs/mcp-go/server"
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