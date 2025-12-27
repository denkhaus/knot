package session

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SessionManager defines the session management interface for MCP services.
// This abstracts session management to enable dependency injection and testing.
//
// NOTE: This interface is ONLY used in MCP mode (PostgreSQL), not in local CLI mode.
// - MCP Mode: Sessions are created via MCP server, Actor is set once via project_select
// - CLI Mode: No sessions, Actor is passed directly to each command call
type SessionManager interface {
	// Session lifecycle management
	CreateSession(clientID string) (*SessionContext, error)
	CreateSessionWithID(sessionID uuid.UUID, clientID string) (*SessionContext, error)
	GetSession(sessionID uuid.UUID) (*SessionContext, error)
	GetSessionByClientID(clientID string) (*SessionContext, error)
	DeleteSession(sessionID uuid.UUID) error
	ListSessions() []*SessionContext

	// Project context management
	SetProject(sessionID, projectID uuid.UUID) error
	GetProject(sessionID uuid.UUID) (*uuid.UUID, error)
	ClearProject(sessionID uuid.UUID) error

	// Actor management (MCP mode only)
	// SetActor stores the actor for this session - called once during project_select
	// The actor is then used for all subsequent MCP tool calls in this session
	SetActor(sessionID uuid.UUID, actor string) error

	// Session validation and cleanup
	ValidateSession(sessionID uuid.UUID) bool
	CleanupExpiredSessions(ctx context.Context, timeout time.Duration) error

	// Lifecycle management
	CloseAll(ctx context.Context) error
	GetSessionCount() int
}

// SessionContext represents the session state and context
// Used only in MCP mode - not used in local CLI mode
type SessionContext struct {
	SessionID    uuid.UUID
	ProjectID    *uuid.UUID
	ClientID     string
	Actor        string                // Actor for this session (set via project_select, used for all MCP tools)
	CreatedAt    time.Time
	LastActivity time.Time
	//TODO: has Metadata a actual usage? If not remove that field
	Metadata     map[string]interface{} // Flexible metadata storage
}
