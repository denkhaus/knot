package session

import (
	"context"
	"fmt"
	"time"

	"github.com/denkhaus/knot/v2/internal/config"
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

// DatabaseSessionManager implements session.Manager using a database for persistent storage
// Works with PostgreSQL (MCP Mode) via the SessionRepository interface
type DatabaseSessionManager struct {
	repo   types.SessionRepository
	logger logger.Logger
}

// NewDatabaseSessionManager creates a database-backed session manager
func NewDatabaseSessionManager(repo types.SessionRepository, logger logger.Logger) Manager {
	return &DatabaseSessionManager{
		repo:   repo,
		logger: logger,
	}
}

// CreateSession creates a new session and stores it in the database
func (m *DatabaseSessionManager) CreateSession(clientID string) (*SessionContext, error) {
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

// GetSession gets a session by ID and updates activity timestamp
func (m *DatabaseSessionManager) GetSession(sessionID uuid.UUID) (*SessionContext, error) {
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

// DeleteSession removes a session by ID
func (m *DatabaseSessionManager) DeleteSession(sessionID uuid.UUID) error {
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
func (m *DatabaseSessionManager) ListSessions() []*SessionContext {
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
func (m *DatabaseSessionManager) SetProject(sessionID, projectID uuid.UUID) error {
	ctx := context.Background()

	err := m.repo.SetSessionProject(ctx, sessionID, projectID)
	if err != nil {
		return fmt.Errorf("failed to set session project: %w", err)
	}

	m.logger.Debug("Project set for session",
		logger.String("session_id", sessionID.String()),
		logger.String("project_id", projectID.String()))

	return nil
}

// GetProject retrieves the project ID for a session
func (m *DatabaseSessionManager) GetProject(sessionID uuid.UUID) (*uuid.UUID, error) {
	ctx := context.Background()

	projectID, err := m.repo.GetSessionProject(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session project: %w", err)
	}

	return projectID, nil
}

// ClearProject removes the project binding from a session
func (m *DatabaseSessionManager) ClearProject(sessionID uuid.UUID) error {
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
func (m *DatabaseSessionManager) ValidateSession(sessionID uuid.UUID) bool {
	ctx := context.Background()

	return m.repo.ValidateSession(ctx, sessionID)
}

// CleanupExpiredSessions removes sessions that have exceeded the timeout
func (m *DatabaseSessionManager) CleanupExpiredSessions(ctx context.Context, timeout time.Duration) error {
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
func (m *DatabaseSessionManager) CloseAll(ctx context.Context) error {
	// In PostgreSQL, sessions are cleaned up by expiration
	// This is mainly for cleanup during shutdown
	m.logger.Info("All sessions closed")
	return nil
}

// GetSessionCount returns the number of active sessions
func (m *DatabaseSessionManager) GetSessionCount() int {
	ctx := context.Background()

	count, err := m.repo.GetSessionCount(ctx)
	if err != nil {
		m.logger.Error("Failed to get session count", zap.Error(err))
		return 0
	}

	return count
}

// sessionStorageFactory creates session managers based on storage type
// Private implementation of StorageFactory interface
type sessionStorageFactory struct {
	configService config.Service
	logger        logger.Logger
}

// Ensure sessionStorageFactory implements StorageFactory
var _ StorageFactory = (*sessionStorageFactory)(nil)

// NewSessionStorageFactory creates a new session storage factory
func NewSessionStorageFactory(configService config.Service, logger logger.Logger) StorageFactory {
	return &sessionStorageFactory{
		configService: configService,
		logger:        logger,
	}
}

// DetectStorageType determines the appropriate storage type based on configuration
func (f *sessionStorageFactory) DetectStorageType() SessionStorageType {
	// Check if MCP mode is configured with a database
	if f.configService.IsMCPMode() && f.configService.GetMCPConfig().Database.Endpoint != "" {
		return DatabaseStorage
	}
	return MemoryStorage
}

// CreateSessionManager creates a session manager based on the detected storage type
func (f *sessionStorageFactory) CreateSessionManager(ctx context.Context, repo types.Repository) (Manager, error) {
	storageType := f.DetectStorageType()

	f.logger.Info("Creating session manager",
		logger.String("storage_type", string(storageType)))

	switch storageType {
	case DatabaseStorage:
		// For database storage, we need a SessionRepository
		// Try to assert the repository to SessionRepository interface
		if sessionRepo, ok := repo.(types.SessionRepository); ok {
			return f.createDatabaseSessionManager(ctx, sessionRepo)
		}
		return nil, fmt.Errorf("repository does not implement SessionRepository interface for database sessions")
	case MemoryStorage:
		return f.createMemorySessionManager(ctx)
	default:
		return nil, fmt.Errorf("unsupported session storage type: %s", storageType)
	}
}

// createDatabaseSessionManager creates a database-backed session manager for MCP mode
func (f *sessionStorageFactory) createDatabaseSessionManager(ctx context.Context, repo types.SessionRepository) (Manager, error) {
	f.logger.Info("Creating database session manager for MCP mode")

	sessionManager := NewDatabaseSessionManager(repo, f.logger)
	f.logger.Info("Database session manager created successfully")
	return sessionManager, nil
}

// createMemorySessionManager creates an in-memory session manager for Local mode
func (f *sessionStorageFactory) createMemorySessionManager(ctx context.Context) (Manager, error) {
	f.logger.Info("Creating memory session manager for Local mode")

	// Use the existing memory-based session manager
	return newManager(), nil
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
func (f *sessionStorageFactory) GetProviderTypeFromStorageType(storageType SessionStorageType) SessionProviderType {
	switch storageType {
	case DatabaseStorage:
		return DatabaseSessionProvider
	case MemoryStorage:
		return MemorySessionProvider
	default:
		return MemorySessionProvider // Default fallback
	}
}
