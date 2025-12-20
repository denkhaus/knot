package mcp

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/hints"
	"github.com/denkhaus/knot/v2/internal/mcp/tools"
	"github.com/denkhaus/knot/v2/internal/mcp/transports"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/mark3labs/mcp-go/server"
)


// mcpServerImpl is the private implementation of the Server interface
type mcpServerImpl struct {
	*server.MCPServer
	projectManager  manager.ProjectManager
	sessions        session.Manager
	logger          logger.Logger
	config          *config.MCPConfig
	hintIntegration hints.Integration
	running         bool
	transport       transports.Transport
}

// ServerConfig holds configuration for creating an MCP server
type ServerConfig struct {
	ProjectManager  manager.ProjectManager
	SessionManager  session.Manager
	Logger          logger.Logger
	Config          *config.MCPConfig
	HintIntegration hints.Integration
}

// NewServer creates a new MCP server with multi-project support using dependency injection
func newServer(cfg ServerConfig) (Server, error) {
	if cfg.ProjectManager == nil {
		return nil, fmt.Errorf("project manager is required")
	}
	if cfg.SessionManager == nil {
		return nil, fmt.Errorf("session manager is required")
	}
	if cfg.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if cfg.Config == nil {
		return nil, fmt.Errorf("config is required")
	}

	// Create mcp-go server
	mcpServer := server.NewMCPServer("knot", "1.0.0")

	// Create transport dependencies
	transportDeps := transports.TransportDependencies{
		MCPServer:      mcpServer,
		ProjectManager: cfg.ProjectManager,
		SessionManager: cfg.SessionManager,
		Logger:         cfg.Logger,
		HintIntegration: cfg.HintIntegration,
		ServerConfig:   cfg.Config,
	}

	// Create transport based on configuration
	transport, err := transports.CreateTransport(transportDeps, cfg.Config.Transport.Mode)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	s := &mcpServerImpl{
		MCPServer:      mcpServer,
		projectManager: cfg.ProjectManager,
		sessions:       cfg.SessionManager,
		logger:         cfg.Logger,
		config:         cfg.Config,
		hintIntegration: cfg.HintIntegration,
		running:        false,
		transport:      transport,
	}

	// Register core tools
	if err := s.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return s, nil
}

// registerTools registers all MCP tools with the server
func (s *mcpServerImpl) registerTools() error {
	// Register navigation tools (project_select, project_list, project_get)
	tools.RegisterNavigationTools(s.MCPServer, s.projectManager, s.sessions)

	// Register project management tools (project_create, project_update, project_delete)
	tools.RegisterProjectManagementTools(s.MCPServer, s.projectManager)

	// Register task management tools (task_create, task_get, task_update, task_update_state, task_delete)
	tools.RegisterTaskManagementTools(s.MCPServer, s.projectManager, s.sessions)

	// Register status and query tools (status_ready, status_actionable, status_project)
	tools.RegisterStatusTools(s.MCPServer, s.projectManager, s.sessions)

	return nil
}

// Start starts the MCP server with the configured transport
func (s *mcpServerImpl) Start() error {
	s.logger.Info("Starting MCP server",
		logger.String("transport", s.config.Transport.Mode.String()),
		logger.String("address", s.config.Address),
		logger.Int("port", s.config.Port),
		logger.String("database", s.config.Database.Endpoint),
	)

	s.running = true
	s.logger.Info("MCP server started",
		logger.String("transport", s.transport.GetType().String()),
	)

	// Start the configured transport
	return s.transport.Start(context.Background())
}

// Stop gracefully stops the MCP server
func (s *mcpServerImpl) Stop(ctx context.Context) error {
	s.logger.Info("Stopping MCP server")
	s.running = false

	// Stop the transport
	if err := s.transport.Stop(ctx); err != nil {
		s.logger.Error("Error stopping transport", logger.Error(err))
		return err
	}

	return nil
}

// IsRunning returns true if the server is currently running
func (s *mcpServerImpl) IsRunning() bool {
	return s.running
}

// GetSessionCount returns the number of active sessions
func (s *mcpServerImpl) GetSessionCount() int {
	return s.sessions.GetSessionCount()
}

// CleanupExpiredSessions removes expired sessions using the configured timeout
func (s *mcpServerImpl) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessions.CleanupExpiredSessions(ctx, s.config.Timeout)
}

// GetConfig returns the server configuration
func (s *mcpServerImpl) GetConfig() interface{} {
	return s.config
}

// getSessionID extracts session ID from context
func getSessionID(ctx context.Context) string {
	if session := server.ClientSessionFromContext(ctx); session != nil {
		return session.SessionID()
	}
	return ""
}
