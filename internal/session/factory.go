package session

import (
	"context"
	"fmt"
	"time"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SessionStorageType defines the type of session storage backend
type SessionStorageType string

const (
	// MemoryStorage uses in-memory sync.Map for session storage (Local Mode)
	MemoryStorage SessionStorageType = "memory"
	// DatabaseStorage uses database (SQLite/PostgreSQL) for persistent session storage (MCP Mode)
	DatabaseStorage SessionStorageType = "database"
)

// databaseSessionManagerImpl implements session.Manager using a database for persistent storage
// Works with PostgreSQL (MCP Mode) via the SessionRepository interface
type databaseSessionManagerImpl struct {
	repo   types.SessionRepository
	logger logger.Logger
}

// CreateSession creates a new session and stores it in the database
func (m *databaseSessionManagerImpl) CreateSession(clientID string) (*SessionContext, error) {
	if clientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}

	ctx := context.Background()

	// Create session in database
	dbSession, err := m.repo.CreateSession(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Create session context
	sessionCtx := &SessionContext{
		SessionID:    dbSession.ID,
		ClientID:     dbSession.ClientID,
		CreatedAt:    dbSession.CreatedAt,
		LastActivity: dbSession.LastActivity,
		Metadata:     dbSession.Metadata,
	}

	m.logger.Debug("Session created in database",
		logger.String("session_id", sessionCtx.SessionID.String()),
		logger.String("client_id", clientID))

	return sessionCtx, nil
}

// CreateSessionWithID creates a new session with a specific session ID and stores it in the database
func (m *databaseSessionManagerImpl) CreateSessionWithID(sessionID uuid.UUID, clientID string) (*SessionContext, error) {
	if clientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}

	ctx := context.Background()

	// Create session in database with specific ID
	dbSession, err := m.repo.CreateSessionWithID(ctx, sessionID, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session with ID: %w", err)
	}

	// Create session context
	sessionCtx := &SessionContext{
		SessionID:    dbSession.ID,
		ClientID:     dbSession.ClientID,
		CreatedAt:    dbSession.CreatedAt,
		LastActivity: dbSession.LastActivity,
		Metadata:     dbSession.Metadata,
	}

	m.logger.Debug("Session created in database with specific ID",
		logger.String("session_id", sessionCtx.SessionID.String()),
		logger.String("client_id", clientID))

	return sessionCtx, nil
}

// GetSession gets a session by ID and updates activity timestamp
func (m *databaseSessionManagerImpl) GetSession(sessionID uuid.UUID) (*SessionContext, error) {
	ctx := context.Background()

	// Get session from database (updates activity automatically)
	dbSession, err := m.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Create session context
	sessionCtx := &SessionContext{
		SessionID:    dbSession.ID,
		ClientID:     dbSession.ClientID,
		CreatedAt:    dbSession.CreatedAt,
		LastActivity: dbSession.LastActivity,
		Metadata:     dbSession.Metadata,
	}

	m.logger.Debug("Session retrieved from database",
		logger.String("session_id", sessionCtx.SessionID.String()))

	return sessionCtx, nil
}

// GetSessionByClientID gets a session by client ID and updates activity timestamp
func (m *databaseSessionManagerImpl) GetSessionByClientID(clientID string) (*SessionContext, error) {
	ctx := context.Background()

	m.logger.Debug("GetSessionByClientID called",
		logger.String("client_id", clientID))

	// List sessions for this client
	dbSessions, err := m.repo.ListSessions(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions for client: %w", err)
	}

	m.logger.Debug("GetSessionByClientID found sessions",
		logger.String("client_id", clientID),
		logger.Int("count", len(dbSessions)))

	if len(dbSessions) == 0 {
		return nil, fmt.Errorf("no sessions found for client: %s", clientID)
	}

	// First, try to find a session with a project set
	for _, dbSession := range dbSessions {
		if dbSession.ProjectID != nil {
			// Create session context
			sessionCtx := &SessionContext{
				SessionID:    dbSession.ID,
				ClientID:     dbSession.ClientID,
				CreatedAt:    dbSession.CreatedAt,
				LastActivity: dbSession.LastActivity,
				Metadata:     dbSession.Metadata,
				ProjectID:    dbSession.ProjectID,
			}

			m.logger.Debug("Session retrieved by client ID (with project)",
				logger.String("client_id", clientID),
				logger.String("session_id", sessionCtx.SessionID.String()))

			return sessionCtx, nil
		}
	}

	// If no session has a project set, use the most recent one
	dbSession := dbSessions[0]

	// Create session context
	sessionCtx := &SessionContext{
		SessionID:    dbSession.ID,
		ClientID:     dbSession.ClientID,
		CreatedAt:    dbSession.CreatedAt,
		LastActivity: dbSession.LastActivity,
		Metadata:     dbSession.Metadata,
		ProjectID:    dbSession.ProjectID,
	}

	m.logger.Debug("Session retrieved by client ID (most recent)",
		logger.String("client_id", clientID),
		logger.String("session_id", sessionCtx.SessionID.String()))

	return sessionCtx, nil
}

// DeleteSession removes a session by ID
func (m *databaseSessionManagerImpl) DeleteSession(sessionID uuid.UUID) error {
	ctx := context.Background()

	err := m.repo.DeleteSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	m.logger.Debug("Session deleted from database",
		logger.String("session_id", sessionID.String()))

	return nil
}

// ListSessions returns all active sessions
func (m *databaseSessionManagerImpl) ListSessions() []*SessionContext {
	ctx := context.Background()

	// Get all sessions (empty clientID means all clients)
	dbSessions, err := m.repo.ListSessions(ctx, "")
	if err != nil {
		m.logger.Error("Failed to list sessions", zap.Error(err))
		return []*SessionContext{}
	}

	// Convert to session contexts
	sessions := make([]*SessionContext, len(dbSessions))
	for i, dbSession := range dbSessions {
		sessions[i] = &SessionContext{
			SessionID:    dbSession.ID,
			ClientID:     dbSession.ClientID,
			CreatedAt:    dbSession.CreatedAt,
			LastActivity: dbSession.LastActivity,
			Metadata:     dbSession.Metadata,
		}
	}

	m.logger.Debug("Listing sessions from database",
		logger.Int("count", len(sessions)))

	return sessions
}

// SetProject sets the project for a session
func (m *databaseSessionManagerImpl) SetProject(sessionID, projectID uuid.UUID) error {
	ctx := context.Background()

	fmt.Printf("🔧 [SetProject] Setting project %s for session %s\n", projectID.String(), sessionID.String())

	err := m.repo.SetSessionProject(ctx, sessionID, projectID)
	if err != nil {
		fmt.Printf("❌ [SetProject] Failed to set project: %v\n", err)
		return fmt.Errorf("failed to set session project: %w", err)
	}

	fmt.Printf("✅ [SetProject] Successfully set project %s for session %s\n", projectID.String(), sessionID.String())

	m.logger.Debug("Project set for session",
		logger.String("session_id", sessionID.String()),
		logger.String("project_id", projectID.String()))

	return nil
}

// GetProject retrieves the project ID for a session
func (m *databaseSessionManagerImpl) GetProject(sessionID uuid.UUID) (*uuid.UUID, error) {
	ctx := context.Background()

	projectID, err := m.repo.GetSessionProject(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session project: %w", err)
	}

	return projectID, nil
}

// ClearProject removes the project binding from a session
func (m *databaseSessionManagerImpl) ClearProject(sessionID uuid.UUID) error {
	ctx := context.Background()

	err := m.repo.ClearSessionProject(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to clear session project: %w", err)
	}

	m.logger.Debug("Project cleared for session",
		logger.String("session_id", sessionID.String()))

	return nil
}

// ValidateSession checks if a session exists and is active
func (m *databaseSessionManagerImpl) ValidateSession(sessionID uuid.UUID) bool {
	ctx := context.Background()

	return m.repo.ValidateSession(ctx, sessionID)
}

// CleanupExpiredSessions removes sessions that have exceeded the timeout
func (m *databaseSessionManagerImpl) CleanupExpiredSessions(ctx context.Context, timeout time.Duration) error {
	cutoff := time.Now().Add(-timeout)

	err := m.repo.CleanupExpiredSessions(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	m.logger.Info("Cleaning up expired sessions",
		zap.Time("cutoff", cutoff),
		zap.String("timeout", timeout.String()))

	return nil
}

// CloseAll closes all sessions
func (m *databaseSessionManagerImpl) CloseAll(ctx context.Context) error {
	// In PostgreSQL, sessions are cleaned up by expiration
	// This is mainly for cleanup during shutdown
	m.logger.Info("All sessions closed")
	return nil
}

// GetSessionCount returns the number of active sessions
func (m *databaseSessionManagerImpl) GetSessionCount() int {
	ctx := context.Background()

	count, err := m.repo.GetSessionCount(ctx)
	if err != nil {
		m.logger.Error("Failed to get session count", zap.Error(err))
		return 0
	}

	return count
}

// IsDatabaseStorage checks if the storage type is Database
func IsDatabaseStorage(storageType SessionStorageType) bool {
	return storageType == DatabaseStorage
}

// IsMemoryStorage checks if the storage type is Memory
func IsMemoryStorage(storageType SessionStorageType) bool {
	return storageType == MemoryStorage
}

// GetProviderTypeFromStorageType converts storage type to provider type
func GetProviderTypeFromStorageType(storageType SessionStorageType) SessionProviderType {
	switch storageType {
	case DatabaseStorage:
		return DatabaseSessionProvider
	case MemoryStorage:
		return MemorySessionProvider
	default:
		return MemorySessionProvider // Default fallback
	}
}
