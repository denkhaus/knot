package shared

import (
	"sync/atomic"
	"time"

	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPClientSession implements the mcp-go ClientSession interface
// It wraps our internal session data and provides the required methods
type MCPClientSession struct {
	// Core session data
	id               string
	notificationChan chan mcp.JSONRPCNotification
	isInitialized    int32 // atomic bool

	// Application-specific data
	actor       string
	projectID   *uuid.UUID
	createdAt   time.Time
	lastActive  time.Time

	// Optional extended interfaces
	logLevel         *mcp.LoggingLevel
	tools            map[string]server.ServerTool
	resources        map[string]server.ServerResource
	resourceTemplates map[string]server.ServerResourceTemplate
	clientInfo       *mcp.Implementation
	clientCapabilities *mcp.ClientCapabilities
}

// NewMCPClientSession creates a new MCP client session
func NewMCPClientSession(sessionID uuid.UUID, actor string) *MCPClientSession {
	return &MCPClientSession{
		id:               sessionID.String(),
		notificationChan: make(chan mcp.JSONRPCNotification, 100),
		actor:            actor,
		createdAt:        time.Now(),
		lastActive:       time.Now(),
		tools:            make(map[string]server.ServerTool),
		resources:        make(map[string]server.ServerResource),
		resourceTemplates: make(map[string]server.ServerResourceTemplate),
	}
}

// NewMCPClientSessionFromInternal creates an MCP client session from our internal session data
func NewMCPClientSessionFromInternal(internalSession *session.SessionContext, actor string) *MCPClientSession {
	return &MCPClientSession{
		id:               internalSession.SessionID.String(),
		notificationChan: make(chan mcp.JSONRPCNotification, 100),
		isInitialized:    1, // Internal sessions are considered initialized
		actor:            actor,
		projectID:        internalSession.ProjectID,
		createdAt:        internalSession.CreatedAt,
		lastActive:       internalSession.LastActivity,
		tools:            make(map[string]server.ServerTool),
		resources:        make(map[string]server.ServerResource),
		resourceTemplates: make(map[string]server.ServerResourceTemplate),
	}
}

// ClientSession interface implementation

// SessionID returns the unique session identifier
func (s *MCPClientSession) SessionID() string {
	return s.id
}

// NotificationChannel returns the channel for sending notifications to the client
func (s *MCPClientSession) NotificationChannel() chan<- mcp.JSONRPCNotification {
	return s.notificationChan
}

// Initialize marks the session as ready to accept notifications
func (s *MCPClientSession) Initialize() {
	atomic.StoreInt32(&s.isInitialized, 1)
}

// Initialized returns whether the session is ready to accept notifications
func (s *MCPClientSession) Initialized() bool {
	return atomic.LoadInt32(&s.isInitialized) == 1
}

// GetActor returns the session actor
func (s *MCPClientSession) GetActor() string {
	return s.actor
}

// GetProjectID returns the selected project ID for this session
func (s *MCPClientSession) GetProjectID() *uuid.UUID {
	return s.projectID
}

// SetProjectID sets the selected project for this session
func (s *MCPClientSession) SetProjectID(projectID *uuid.UUID) {
	s.projectID = projectID
	s.updateLastActive()
}

// GetCreatedAt returns when the session was created
func (s *MCPClientSession) GetCreatedAt() time.Time {
	return s.createdAt
}

// GetLastActive returns when the session was last active
func (s *MCPClientSession) GetLastActive() time.Time {
	return s.lastActive
}

// updateLastActive updates the last active timestamp
func (s *MCPClientSession) updateLastActive() {
	s.lastActive = time.Now()
}

// Extended interface implementations (optional)

// SessionWithLogging implementation
func (s *MCPClientSession) SetLogLevel(level mcp.LoggingLevel) {
	s.logLevel = &level
}

func (s *MCPClientSession) GetLogLevel() mcp.LoggingLevel {
	if s.logLevel == nil {
		return mcp.LoggingLevelInfo
	}
	return *s.logLevel
}

// SessionWithTools implementation
func (s *MCPClientSession) GetSessionTools() map[string]server.ServerTool {
	// Return a copy to prevent external modification
	result := make(map[string]server.ServerTool, len(s.tools))
	for k, v := range s.tools {
		result[k] = v
	}
	return result
}

func (s *MCPClientSession) SetSessionTools(tools map[string]server.ServerTool) {
	s.tools = make(map[string]server.ServerTool)
	for k, v := range tools {
		s.tools[k] = v
	}
}

// SessionWithResources implementation
func (s *MCPClientSession) GetSessionResources() map[string]server.ServerResource {
	result := make(map[string]server.ServerResource, len(s.resources))
	for k, v := range s.resources {
		result[k] = v
	}
	return result
}

func (s *MCPClientSession) SetSessionResources(resources map[string]server.ServerResource) {
	s.resources = make(map[string]server.ServerResource)
	for k, v := range resources {
		s.resources[k] = v
	}
}

// SessionWithResourceTemplates implementation
func (s *MCPClientSession) GetSessionResourceTemplates() map[string]server.ServerResourceTemplate {
	result := make(map[string]server.ServerResourceTemplate, len(s.resourceTemplates))
	for k, v := range s.resourceTemplates {
		result[k] = v
	}
	return result
}

func (s *MCPClientSession) SetSessionResourceTemplates(templates map[string]server.ServerResourceTemplate) {
	s.resourceTemplates = make(map[string]server.ServerResourceTemplate)
	for k, v := range templates {
		s.resourceTemplates[k] = v
	}
}

// SessionWithClientInfo implementation
func (s *MCPClientSession) GetClientInfo() mcp.Implementation {
	if s.clientInfo == nil {
		return mcp.Implementation{}
	}
	return *s.clientInfo
}

func (s *MCPClientSession) SetClientInfo(clientInfo mcp.Implementation) {
	s.clientInfo = &clientInfo
}

func (s *MCPClientSession) GetClientCapabilities() mcp.ClientCapabilities {
	if s.clientCapabilities == nil {
		return mcp.ClientCapabilities{}
	}
	return *s.clientCapabilities
}

func (s *MCPClientSession) SetClientCapabilities(capabilities mcp.ClientCapabilities) {
	s.clientCapabilities = &capabilities
}

// Close closes the session and cleans up resources
func (s *MCPClientSession) Close() {
	close(s.notificationChan)
}

// Ensure MCPClientSession implements ClientSession
var _ server.ClientSession = (*MCPClientSession)(nil)