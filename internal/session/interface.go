package session

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Manager defines the session management interface for MCP services.
// This abstracts session management to enable dependency injection and testing.
type Manager interface {
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

	// Session validation and cleanup
	ValidateSession(sessionID uuid.UUID) bool
	CleanupExpiredSessions(ctx context.Context, timeout time.Duration) error

	// Lifecycle management
	CloseAll(ctx context.Context) error
	GetSessionCount() int
}

// SessionContext represents the session state and context
type SessionContext struct {
	SessionID    uuid.UUID
	ProjectID    *uuid.UUID
	ClientID     string
	CreatedAt    time.Time
	LastActivity time.Time
	Metadata     map[string]interface{} // Flexible metadata storage
}

// SessionInfo provides a summary of session information for logging and monitoring
type SessionInfo struct {
	SessionID     string
	ProjectID     *string
	ClientID      string
	CreatedAt     time.Time
	LastActivity  time.Time
	IsExpired     bool
	MetadataCount int
}
