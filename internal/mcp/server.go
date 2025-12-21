package mcp

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/hints"
	"github.com/denkhaus/knot/v2/internal/mcp/shared"
	"github.com/denkhaus/knot/v2/internal/mcp/tools"
	"github.com/denkhaus/knot/v2/internal/mcp/transports"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/mark3labs/mcp-go/server"
	"github.com/samber/do/v2"
)

// mcpServerImpl is the private implementation of the Server interface
type mcpServerImpl struct {
	*server.MCPServer
	projectManager  manager.ProjectManager
	sessionManager  session.SessionManager
	sessionRegistry shared.SessionRegistry
	logger          logger.Logger
	config          *config.MCPConfig
	hintIntegration hints.Integration
	running         bool
	transport       transports.Transport
}

// ServerConfig holds configuration for creating an MCP server
type ServerConfig struct {
	ProjectManager  manager.ProjectManager
	SessionManager  session.SessionManager
	Logger          logger.Logger
	Config          *config.MCPConfig
	HintIntegration hints.Integration
	MCPServer       *server.MCPServer
	Injector        do.Injector // Added for DI access
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

	// Use the injected mcp-go server
	mcpServer := cfg.MCPServer
	if mcpServer == nil {
		return nil, fmt.Errorf("MCPServer is required in ServerConfig")
	}

	// Session components will be created via DI container when needed

	// Transport will be injected via DI
	transport := do.MustInvoke[transports.Transport](cfg.Injector)

	// SessionRegistry will be injected via DI if available
	var sessionRegistry shared.SessionRegistry
	if sessionReg, err := do.Invoke[shared.SessionRegistry](cfg.Injector); err == nil {
		sessionRegistry = sessionReg
	}

	s := &mcpServerImpl{
		MCPServer:       mcpServer,
		projectManager:  cfg.ProjectManager,
		sessionManager:  cfg.SessionManager,
		sessionRegistry: sessionRegistry,
		logger:          cfg.Logger,
		config:          cfg.Config,
		hintIntegration: cfg.HintIntegration,
		running:         false,
		transport:       transport,
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
	tools.RegisterNavigationTools(s.MCPServer, s.projectManager, s.sessionManager)

	// Register project management tools (project_create, project_update, project_delete)
	tools.RegisterProjectManagementTools(s.MCPServer, s.projectManager)

	// Register task management tools (task_create, task_get, task_update, task_update_state, task_delete)
	tools.RegisterTaskManagementTools(s.MCPServer, s.projectManager, s.sessionManager, s.sessionRegistry)

	// Register status and query tools (status_ready, status_actionable, status_project, status_blocked, status_tree)
	tools.RegisterStatusTools(s.MCPServer, s.projectManager, s.sessionManager)

	// Register dependency management tools (dependency_list, dependency_add, dependency_remove, dependency_check)
	tools.RegisterDependencyTools(s.MCPServer, s.projectManager, s.sessionManager)

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

	// Sync existing sessions with MCP server
	if err := s.sessionRegistry.SyncExistingSessions(context.Background()); err != nil {
		s.logger.Warn("Failed to sync existing sessions",
			logger.Error(err))
	}

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

	// Close all sessions
	if err := s.sessionManager.CloseAll(ctx); err != nil {
		s.logger.Warn("Error closing sessions", logger.Error(err))
	}

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
	if s.sessionRegistry != nil {
		return s.sessionRegistry.GetSessionCount()
	}
	// Fall back to session manager if session registry is not available
	return s.sessionManager.GetSessionCount()
}

// CleanupExpiredSessions removes expired sessions using the configured timeout
func (s *mcpServerImpl) CleanupExpiredSessions(ctx context.Context) error {
	if s.sessionRegistry != nil {
		return s.sessionRegistry.CleanupExpiredSessions(ctx, s.config.Timeout)
	}
	// Fall back to session manager if session registry is not available
	return s.sessionManager.CleanupExpiredSessions(ctx, s.config.Timeout)
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
