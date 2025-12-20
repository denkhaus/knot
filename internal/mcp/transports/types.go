package transports

import (
	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/hints"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/mark3labs/mcp-go/server"
)

// TransportDependencies contains all dependencies needed by transport implementations
// This ensures consistent dependency injection across all transports
type TransportDependencies struct {
	// Core MCP server instance
	MCPServer *server.MCPServer

	// Application services
	ProjectManager  manager.ProjectManager
	SessionManager  session.Manager
	Logger          logger.Logger
	HintIntegration hints.Integration

	// Configuration
	ServerConfig *config.MCPConfig
}