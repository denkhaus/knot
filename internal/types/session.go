package types

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Session represents a client session in MCP mode
type Session struct {
	ID           uuid.UUID                 `json:"id"`
	ClientID     string                    `json:"client_id"` // MCP client identifier
	CreatedAt    time.Time                 `json:"created_at"`
	LastActivity time.Time                 `json:"last_activity"`
	ExpiresAt    *time.Time                `json:"expires_at,omitempty"`
	Metadata     map[string]interface{}    `json:"metadata,omitempty"`
	Actor        string                    `json:"actor,omitempty"` // Who is operating the session
	Status       SessionStatus             `json:"status"`
	ProjectID    *uuid.UUID                `json:"project_id,omitempty"`
}

// SessionStatus represents the status of a session
type SessionStatus string

const (
	SessionStatusActive   SessionStatus = "active"
	SessionStatusInactive SessionStatus = "inactive"
	SessionStatusExpired  SessionStatus = "expired"
)

// SessionRepository defines the interface for session storage operations
// This is only used in MCP mode, not in local CLI mode
type SessionRepository interface {
	// CreateSession creates a new session
	CreateSession(ctx context.Context, clientID string) (*Session, error)

	// CreateSessionWithID creates a new session with a specific session ID
	CreateSessionWithID(ctx context.Context, sessionID uuid.UUID, clientID string) (*Session, error)

	// GetSession retrieves a session by ID and updates last activity
	GetSession(ctx context.Context, sessionID uuid.UUID) (*Session, error)

	// DeleteSession removes a session by ID
	DeleteSession(ctx context.Context, sessionID uuid.UUID) error

	// ListSessions returns all sessions for a client, or all sessions if clientID is empty
	ListSessions(ctx context.Context, clientID string) ([]*Session, error)

	// UpdateSessionActivity updates the last activity timestamp
	UpdateSessionActivity(ctx context.Context, sessionID uuid.UUID) error

	// SetSessionProject associates a project with a session
	SetSessionProject(ctx context.Context, sessionID, projectID uuid.UUID) error

	// GetSessionProject retrieves the project ID associated with a session
	GetSessionProject(ctx context.Context, sessionID uuid.UUID) (*uuid.UUID, error)

	// ClearSessionProject removes the project association from a session
	ClearSessionProject(ctx context.Context, sessionID uuid.UUID) error

	// CleanupExpiredSessions removes sessions that have expired
	CleanupExpiredSessions(ctx context.Context, before time.Time) error

	// GetSessionCount returns the number of active sessions
	GetSessionCount(ctx context.Context) (int, error)

	// ValidateSession checks if a session exists and is active
	ValidateSession(ctx context.Context, sessionID uuid.UUID) bool
}