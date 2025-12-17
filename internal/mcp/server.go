package mcp

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/session"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Structured request/response types for MCP tools

type ProjectSelectRequest struct {
	ProjectID string `json:"project_id" jsonschema_description:"The ID of the project to select" jsonschema:"required"`
}

type ProjectSelectResponse struct {
	Message   string `json:"message" jsonschema_description:"Confirmation message"`
	ProjectID string `json:"project_id" jsonschema_description:"Selected project ID"`
}

type ProjectCreateRequest struct {
	Title       string `json:"title" jsonschema_description:"Project title" jsonschema:"required"`
	Description string `json:"description,omitempty" jsonschema_description:"Project description"`
}

type ProjectCreateResponse struct {
	Message   string `json:"message" jsonschema_description:"Confirmation message"`
	ProjectID string `json:"project_id" jsonschema_description:"Created project ID"`
	Title     string `json:"title" jsonschema_description:"Project title"`
}

type TaskCreateRequest struct {
	Title       string  `json:"title" jsonschema_description:"Task title" jsonschema:"required"`
	Description string  `json:"description,omitempty" jsonschema_description:"Task description"`
	Complexity  int     `json:"complexity,omitempty" jsonschema_description:"Task complexity (1-10)" jsonschema:"minimum=1,maximum=10,default=5"`
}

type TaskCreateResponse struct {
	Message   string `json:"message" jsonschema_description:"Confirmation message"`
	TaskID    string `json:"task_id" jsonschema_description:"Created task ID"`
	ProjectID string `json:"project_id" jsonschema_description:"Project ID"`
	Title     string `json:"title" jsonschema_description:"Task title"`
}

// mcpServerImpl is the private implementation of the Server interface
type mcpServerImpl struct {
	*server.MCPServer
	projectManager manager.ProjectManager
	sessions       session.Manager
	logger         logger.Logger
	config         *config.MCPConfig
	running        bool
}

// ServerConfig holds configuration for creating an MCP server
type ServerConfig struct {
	ProjectManager manager.ProjectManager
	SessionManager session.Manager
	Logger         logger.Logger
	Config         *config.MCPConfig
}

// NewServer creates a new MCP server with multi-project support using dependency injection
func NewServer(cfg ServerConfig) (Server, error) {
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

	s := &mcpServerImpl{
		MCPServer:     mcpServer,
		projectManager: cfg.ProjectManager,
		sessions:       cfg.SessionManager,
		logger:         cfg.Logger,
		config:         cfg.Config,
		running:        false,
	}

	// Register core tools
	if err := s.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return s, nil
}

// registerTools registers all MCP tools with the server
func (s *mcpServerImpl) registerTools() error {
	// Project selection tool
	projectSelectTool := mcp.NewTool("project_select",
		mcp.WithDescription("Select the active project for this session"),
		mcp.WithInputSchema[ProjectSelectRequest](),
		mcp.WithOutputSchema[ProjectSelectResponse](),
	)
	s.AddTool(projectSelectTool, mcp.NewStructuredToolHandler(s.handleProjectSelect))

	// Project creation tool
	projectCreateTool := mcp.NewTool("project_create",
		mcp.WithDescription("Create a new project"),
		mcp.WithInputSchema[ProjectCreateRequest](),
		mcp.WithOutputSchema[ProjectCreateResponse](),
	)
	s.AddTool(projectCreateTool, mcp.NewStructuredToolHandler(s.handleProjectCreate))

	// Task creation tool
	taskCreateTool := mcp.NewTool("task_create",
		mcp.WithDescription("Create a new task in the selected project"),
		mcp.WithInputSchema[TaskCreateRequest](),
		mcp.WithOutputSchema[TaskCreateResponse](),
	)
	s.AddTool(taskCreateTool, mcp.NewStructuredToolHandler(s.handleTaskCreate))

	return nil
}

// Start starts the MCP server with stdio transport
func (s *mcpServerImpl) Start() error {
	s.logger.Info("Starting MCP server",
		logger.String("address", s.config.Address),
		logger.Int("port", s.config.Port),
		logger.String("database", s.config.Database.Endpoint),
	)

	s.running = true
	s.logger.Info("MCP server started on stdio")
	return server.ServeStdio(s.MCPServer)
}

// Stop gracefully stops the MCP server
func (s *mcpServerImpl) Stop(ctx context.Context) error {
	s.logger.Info("Stopping MCP server")
	s.running = false

	// Close all sessions
	if err := s.sessions.CloseAll(ctx); err != nil {
		s.logger.Error("Error closing sessions", logger.Error(err))
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

// ExecutionContext provides context for tool execution
type ExecutionContext struct {
	Context       context.Context
	Project       *session.SessionContext
	SessionID     string
	ProjectManager manager.ProjectManager
	Logger        logger.Logger
}

// handleProjectSelect handles project selection for a session
func (s *mcpServerImpl) handleProjectSelect(ctx context.Context, request mcp.CallToolRequest, args ProjectSelectRequest) (ProjectSelectResponse, error) {
	projectID, err := uuid.Parse(args.ProjectID)
	if err != nil {
		return ProjectSelectResponse{}, fmt.Errorf("invalid project_id format: %w", err)
	}

	sessionID := getSessionID(ctx)
	sessionIDUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return ProjectSelectResponse{}, fmt.Errorf("invalid session_id format: %w", err)
	}

	if err := s.sessions.SetProject(sessionIDUUID, projectID); err != nil {
		return ProjectSelectResponse{}, fmt.Errorf("failed to set project: %w", err)
	}

	return ProjectSelectResponse{
		Message:   fmt.Sprintf("Selected project %s for session", projectID),
		ProjectID: projectID.String(),
	}, nil
}

// handleProjectCreate handles project creation
func (s *mcpServerImpl) handleProjectCreate(ctx context.Context, request mcp.CallToolRequest, args ProjectCreateRequest) (ProjectCreateResponse, error) {
	// Create project using project manager
	project, err := s.projectManager.CreateProject(ctx, args.Title, args.Description, "")
	if err != nil {
		return ProjectCreateResponse{}, fmt.Errorf("failed to create project: %w", err)
	}

	return ProjectCreateResponse{
		Message:   fmt.Sprintf("Created project %s (ID: %s)", project.Title, project.ID),
		ProjectID: project.ID.String(),
		Title:     project.Title,
	}, nil
}

// handleTaskCreate handles task creation
func (s *mcpServerImpl) handleTaskCreate(ctx context.Context, request mcp.CallToolRequest, args TaskCreateRequest) (TaskCreateResponse, error) {
	complexity := args.Complexity
	if complexity == 0 {
		complexity = 5 // Default complexity
	}

	// Get session context
	sessionID := getSessionID(ctx)
	sessionIDUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return TaskCreateResponse{}, fmt.Errorf("invalid session_id format: %w", err)
	}

	session, err := s.sessions.GetSession(sessionIDUUID)
	if err != nil {
		return TaskCreateResponse{}, fmt.Errorf("failed to get session: %w", err)
	}

	if session.ProjectID == nil {
		return TaskCreateResponse{}, fmt.Errorf("no project selected for this session")
	}

	// Create task using project manager
	task, err := s.projectManager.CreateTask(ctx, *session.ProjectID, nil, args.Title, args.Description, complexity, types.TaskPriorityMedium, "")
	if err != nil {
		return TaskCreateResponse{}, fmt.Errorf("failed to create task: %w", err)
	}

	return TaskCreateResponse{
		Message:   fmt.Sprintf("Created task %s (ID: %s) in project %s", task.Title, task.ID, session.ProjectID),
		TaskID:    task.ID.String(),
		ProjectID: session.ProjectID.String(),
		Title:     task.Title,
	}, nil
}