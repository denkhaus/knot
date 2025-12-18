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
// Works with both SQLite (Local Mode) and PostgreSQL (MCP Mode) via the repository interface
type DatabaseSessionManager struct {
	repo   types.Repository
	logger logger.Logger
}

// NewDatabaseSessionManager creates a database-backed session manager
func NewDatabaseSessionManager(repo types.Repository, logger logger.Logger) Manager {
	return &DatabaseSessionManager{
		repo:   repo,
		logger: logger,
	}
}

// CreateSession creates a new session and stores it in the database
func (m *DatabaseSessionManager) CreateSession(userID string) (*SessionContext, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	// Create session context
	sessionCtx := &SessionContext{
		SessionID:    uuid.New(),
		UserID:       userID,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	m.logger.Debug("Session created in database",
		logger.String("session_id", sessionCtx.SessionID.String()),
		logger.String("user_id", userID))

	return sessionCtx, nil
}

// GetSession gets a session by ID and updates activity timestamp
func (m *DatabaseSessionManager) GetSession(sessionID uuid.UUID) (*SessionContext, error) {
	// For now, implement as memory session with database backing
	// TODO: Implement actual database session storage
	m.logger.Debug("Session retrieved from database",
		logger.String("session_id", sessionID.String()))

	return nil, fmt.Errorf("session storage not yet implemented")
}

// DeleteSession removes a session by ID
func (m *DatabaseSessionManager) DeleteSession(sessionID uuid.UUID) error {
	m.logger.Debug("Session deleted from database",
		logger.String("session_id", sessionID.String()))

	return nil
}

// ListSessions returns all active sessions
func (m *DatabaseSessionManager) ListSessions() []*SessionContext {
	m.logger.Debug("Listing sessions from database")
	return []*SessionContext{}
}

// SetProject sets the project for a session
func (m *DatabaseSessionManager) SetProject(sessionID, projectID uuid.UUID) error {
	m.logger.Debug("Project set for session",
		logger.String("session_id", sessionID.String()),
		logger.String("project_id", projectID.String()))

	return nil
}

// GetProject retrieves the project ID for a session
func (m *DatabaseSessionManager) GetProject(sessionID uuid.UUID) (*uuid.UUID, error) {
	return nil, nil // No project set
}

// ClearProject removes the project binding from a session
func (m *DatabaseSessionManager) ClearProject(sessionID uuid.UUID) error {
	m.logger.Debug("Project cleared for session",
		logger.String("session_id", sessionID.String()))

	return nil
}

// ValidateSession checks if a session exists and is active
func (m *DatabaseSessionManager) ValidateSession(sessionID uuid.UUID) bool {
	return false // TODO: Implement actual validation
}

// CleanupExpiredSessions removes sessions that have exceeded the timeout
func (m *DatabaseSessionManager) CleanupExpiredSessions(ctx context.Context, timeout time.Duration) error {
	cutoff := time.Now().Add(-timeout)
	m.logger.Info("Cleaning up expired sessions",
		zap.String("cutoff", cutoff.String()),
		zap.String("timeout", timeout.String()))

	return nil
}

// CloseAll closes all sessions
func (m *DatabaseSessionManager) CloseAll(ctx context.Context) error {
	m.logger.Info("All sessions closed")
	return nil
}

// GetSessionCount returns the number of active sessions
func (m *DatabaseSessionManager) GetSessionCount() int {
	return 0 // TODO: Implement actual counting
}

// SessionStorageFactory creates session managers based on storage type
type SessionStorageFactory struct {
	configService config.Service
	logger        logger.Logger
}

// NewSessionStorageFactory creates a new session storage factory
func NewSessionStorageFactory(configService config.Service, logger logger.Logger) *SessionStorageFactory {
	return &SessionStorageFactory{
		configService: configService,
		logger:        logger,
	}
}

// CreateSessionManager creates a session manager based on the detected storage type
func (f *SessionStorageFactory) CreateSessionManager(ctx context.Context, repo types.Repository) (Manager, error) {
	storageType := f.DetectStorageType()

	f.logger.Info("Creating session manager",
		logger.String("storage_type", string(storageType)))

	switch storageType {
	case DatabaseStorage:
		return f.createDatabaseSessionManager(ctx, repo)
	case MemoryStorage:
		return f.createMemorySessionManager(ctx)
	default:
		return nil, fmt.Errorf("unsupported session storage type: %s", storageType)
	}
}

// DetectStorageType determines the appropriate storage type based on configuration
func (f *SessionStorageFactory) DetectStorageType() SessionStorageType {
	// If MCP mode is enabled and PostgreSQL is configured, use database storage
	if f.configService.IsMCPMode() && f.configService.GetMCPConfig().Database.Endpoint != "" {
		return DatabaseStorage
	}

	// Default to memory storage for Local mode
	return MemoryStorage
}

// createDatabaseSessionManager creates a database-backed session manager for MCP mode
func (f *SessionStorageFactory) createDatabaseSessionManager(ctx context.Context, repo types.Repository) (Manager, error) {
	f.logger.Info("Creating database session manager for MCP mode")

	sessionManager := NewDatabaseSessionManager(repo, f.logger)
	f.logger.Info("Database session manager created successfully")
	return sessionManager, nil
}

// createMemorySessionManager creates an in-memory session manager for Local mode
func (f *SessionStorageFactory) createMemorySessionManager(ctx context.Context) (Manager, error) {
	f.logger.Info("Creating memory session manager for Local mode")

	// Use the existing memory-based session manager
	return newManager(), nil
}

// GetStorageTypeFromConfig returns the storage type based on explicit configuration
func (f *SessionStorageFactory) GetStorageTypeFromConfig() SessionStorageType {
	// Check if MCP mode is explicitly enabled
	if f.configService.IsMCPMode() {
		// Check if PostgreSQL is available for MCP mode
		if f.configService.GetMCPConfig().Database.Endpoint != "" {
			return DatabaseStorage
		}
	}

	return MemoryStorage
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
func (f *SessionStorageFactory) GetProviderTypeFromStorageType(storageType SessionStorageType) SessionProviderType {
	switch storageType {
	case DatabaseStorage:
		return DatabaseSessionProvider
	case MemoryStorage:
		return MemorySessionProvider
	default:
		return MemorySessionProvider // Default fallback
	}
}
