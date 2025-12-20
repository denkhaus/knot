package shared

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/server"
)

// SessionRegistry manages the integration between our session manager and mcp-go sessions
type SessionRegistry struct {
	sessionManager session.Manager
	mcpServer      *server.MCPServer
	logger         logger.Logger

	// Cache for active MCP sessions
	mcpSessions sync.Map // sessionID -> *MCPClientSession
}

// NewSessionRegistry creates a new session registry
func NewSessionRegistry(sessionManager session.Manager, mcpServer *server.MCPServer, logger logger.Logger) *SessionRegistry {
	return &SessionRegistry{
		sessionManager: sessionManager,
		mcpServer:      mcpServer,
		logger:         logger,
	}
}

// GetOrCreateSession gets an existing session or creates a new one
// This is called by the MCP transport when handling requests
func (r *SessionRegistry) GetOrCreateSession(ctx context.Context, sessionID uuid.UUID, actor string) (*MCPClientSession, error) {
	// First check if we have an internal session
	internalSession, err := r.sessionManager.GetSession(sessionID)
	if err != nil {
		// If session doesn't exist internally, create a new one
		internalSession, err = r.sessionManager.CreateSession(sessionID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to create internal session: %w", err)
		}
	}

	// Check if we already have an MCP session for this ID
	if mcpSession, ok := r.mcpSessions.Load(sessionID.String()); ok {
		if session, ok := mcpSession.(*MCPClientSession); ok {
			// Update last active time
			session.updateLastActive()
			return session, nil
		}
	}

	// Create new MCP session
	mcpSession := NewMCPClientSessionFromInternal(internalSession, actor)

	// Register with mcp-go
	if err := r.mcpServer.RegisterSession(ctx, mcpSession); err != nil {
		// If registration fails, check if session already exists
		if err == server.ErrSessionExists {
			// Session already registered, try to unregister first
			r.logger.Warn("Session already exists in MCP server, unregistering first",
				logger.String("session_id", sessionID.String()))
			r.mcpServer.UnregisterSession(ctx, sessionID.String())

			// Try to register again
			if err := r.mcpServer.RegisterSession(ctx, mcpSession); err != nil {
				return nil, fmt.Errorf("failed to register session with MCP server: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to register session with MCP server: %w", err)
		}
	}

	// Cache the MCP session
	r.mcpSessions.Store(sessionID.String(), mcpSession)

	r.logger.Info("Session registered with MCP server",
		logger.String("session_id", sessionID.String()),
		logger.String("actor", actor))

	return mcpSession, nil
}

// GetSession retrieves a session by ID
func (r *SessionRegistry) GetSession(sessionID uuid.UUID) (*MCPClientSession, error) {
	if mcpSession, ok := r.mcpSessions.Load(sessionID.String()); ok {
		if session, ok := mcpSession.(*MCPClientSession); ok {
			return session, nil
		}
	}
	return nil, fmt.Errorf("session not found: %s", sessionID.String())
}

// SetSessionProject sets the project for a session
func (r *SessionRegistry) SetSessionProject(ctx context.Context, sessionID uuid.UUID, projectID *uuid.UUID, actor string) error {
	// Get MCP session
	mcpSession, err := r.GetSession(sessionID)
	if err != nil {
		return err
	}

	// Update MCP session
	mcpSession.SetProjectID(projectID)

	// Update internal session using session manager
	if projectID != nil {
		if err := r.sessionManager.SetProject(sessionID, *projectID); err != nil {
			return fmt.Errorf("failed to set project in session: %w", err)
		}
	} else {
		if err := r.sessionManager.ClearProject(sessionID); err != nil {
			return fmt.Errorf("failed to clear project in session: %w", err)
		}
	}

	r.logger.Info("Session project updated",
		logger.String("session_id", sessionID.String()),
		logger.String("project_id", projectID.String()))

	return nil
}

// RemoveSession removes a session from the registry
func (r *SessionRegistry) RemoveSession(ctx context.Context, sessionID uuid.UUID) error {
	// Remove from MCP server
	r.mcpServer.UnregisterSession(ctx, sessionID.String())

	// Remove from cache
	r.mcpSessions.Delete(sessionID.String())

	// Remove from internal manager
	if err := r.sessionManager.DeleteSession(sessionID); err != nil {
		r.logger.Warn("Failed to delete internal session",
			logger.String("session_id", sessionID.String()),
			logger.Error(err))
	}

	r.logger.Info("Session removed",
		logger.String("session_id", sessionID.String()))

	return nil
}

// CleanupExpiredSessions removes expired sessions
func (r *SessionRegistry) CleanupExpiredSessions(ctx context.Context, timeout time.Duration) error {
	// Use session manager to cleanup expired sessions
	return r.sessionManager.CleanupExpiredSessions(ctx, timeout)
}

// GetSessionCount returns the number of active sessions
func (r *SessionRegistry) GetSessionCount() int {
	count := 0
	r.mcpSessions.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// SyncExistingSessions syncs existing sessions from the database with MCP server
// This should be called on server startup
func (r *SessionRegistry) SyncExistingSessions(ctx context.Context) error {
	// Get all sessions from session manager
	sessions := r.sessionManager.ListSessions()

	for _, internalSession := range sessions {
		sessionID := internalSession.SessionID

		// Create MCP session
		mcpSession := NewMCPClientSessionFromInternal(internalSession, internalSession.ClientID)

		// Try to register with MCP server
		if err := r.mcpServer.RegisterSession(ctx, mcpSession); err != nil {
			if err == server.ErrSessionExists {
				// Session might already be registered, skip
				r.logger.Debug("Session already exists in MCP server on sync",
					logger.String("session_id", sessionID.String()))
				continue
			}
			r.logger.Warn("Failed to register session during sync",
				logger.String("session_id", sessionID.String()),
				logger.Error(err))
			continue
		}

		// Cache the session
		r.mcpSessions.Store(sessionID.String(), mcpSession)

		r.logger.Info("Session synced",
			logger.String("session_id", sessionID.String()))
	}

	return nil
}