package mcp

import (
	configsvc "github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/hints"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/mark3labs/mcp-go/server"
	"github.com/samber/do/v2"
)

// NewMCPServer creates a provider function for the raw MCP Server from mcp-go
// This provides the underlying server implementation for dependency injection
func NewMCPServer(injector do.Injector) (*server.MCPServer, error) {
	// Create the basic mcp-go server
	mcpServer := server.NewMCPServer("knot", "1.0.0")
	return mcpServer, nil
}

// NewServer creates a provider function for the MCP Server
// This follows the dependency injection pattern used throughout the application
func NewServer(injector do.Injector) (Server, error) {
	// Resolve dependencies from DI
	configService := do.MustInvoke[configsvc.Service](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)
	sessionManager := do.MustInvoke[session.SessionManager](injector)
	loggerService := do.MustInvoke[logger.Logger](injector)
	hintIntegration := do.MustInvoke[hints.Integration](injector)
	mcpServer := do.MustInvoke[*server.MCPServer](injector)

	// Get MCP configuration
	mcpConfig := configService.GetMCPConfig()

	// Create server configuration
	serverConfig := ServerConfig{
		ProjectManager:  projectManager,
		SessionManager:  sessionManager,
		Logger:          loggerService,
		Config:          mcpConfig,
		HintIntegration: hintIntegration,
		MCPServer:       mcpServer,
		Injector:        injector,
	}

	// Create and return the server
	return newServer(serverConfig)
}
