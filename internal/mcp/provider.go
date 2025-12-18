package mcp

import (
	configsvc "github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/session"
	"github.com/samber/do/v2"
)

// NewServer creates a provider function for the MCP Server
// This follows the dependency injection pattern used throughout the application
func NewServer(injector do.Injector) (Server, error) {
	// Resolve dependencies from DI
	configService := do.MustInvoke[configsvc.Service](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)
	sessionManager := do.MustInvoke[session.Manager](injector)
	loggerService := do.MustInvoke[logger.Logger](injector)

	// Get MCP configuration
	mcpConfig := configService.GetMCPConfig()

	// Create server configuration
	serverConfig := ServerConfig{
		ProjectManager: projectManager,
		SessionManager: sessionManager,
		Logger:         loggerService,
		Config:         mcpConfig,
	}

	// Create and return the server
	return newServer(serverConfig)
}
