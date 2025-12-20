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
	"github.com/samber/do/v2"
)

type SessionRegistry interface {
	GetOrCreateSession(ctx context.Context, sessionID uuid.UUID, actor string) (*MCPClientSession, error)
	GetSession(sessionID uuid.UUID) (*MCPClientSession, error)
	RemoveSession(ctx context.Context, sessionID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context, timeout time.Duration) error
	SyncExistingSessions(ctx context.Context) error
	GetSessionCount() int
}

// SessionRegistry manages the integration between our session manager and mcp-go sessions
type sessionRegistryImpl struct {
	sessionManager session.Manager
	mcpServer      *server.MCPServer
	logger         logger.Logger

	// Cache for active MCP sessions
	mcpSessions sync.Map // sessionID -> *MCPClientSession
}

// NewSessionRegistry creates a new session registry
func NewSessionRegistry(injector do.Injector) (SessionRegistry, error) {
	sessionManager := do.MustInvoke[session.Manager](injector)
	loggerService := do.MustInvoke[logger.Logger](injector)
	mcpServer := do.MustInvoke[*server.MCPServer](injector)

	return &sessionRegistryImpl{
		sessionManager: sessionManager,
		mcpServer:      mcpServer,
		logger:         loggerService,
	}, nil
}

// GetOrCreateSession gets an existing session or creates a new one
// This is called by the MCP transport when handling requests
func (r *sessionRegistryImpl) GetOrCreateSession(ctx context.Context, sessionID uuid.UUID, actor string) (*MCPClientSession, error) {
	fmt.Printf("🔥🔥🔥 GetOrCreateSession called with sessionID: %s, actor: %s\n", sessionID.String(), actor)
	r.logger.Info("🔥 GetOrCreateSession called with sessionID",
		logger.String("session_id", sessionID.String()),
		logger.String("actor", actor))

	// First check if we have an internal session by ID
	internalSession, err := r.sessionManager.GetSession(sessionID)
	if err != nil {
		r.logger.Info("Session not found by ID, trying client ID lookup",
			logger.String("session_id", sessionID.String()))

		// If session doesn't exist by ID, try to find by client ID (for MCP sessions)
		internalSession, err = r.sessionManager.GetSessionByClientID(sessionID.String())
		if err != nil {
			r.logger.Info("Session not found by client ID either, creating new session",
				logger.String("client_id", sessionID.String()))

			// If no session exists by client ID either, create a new one with the provided sessionID
			internalSession, err = r.sessionManager.CreateSessionWithID(sessionID, sessionID.String())
			if err != nil {
				return nil, fmt.Errorf("failed to create internal session: %w", err)
			}
		} else {
			r.logger.Info("Found existing session by client ID",
				logger.String("client_id", sessionID.String()),
				logger.String("found_session_id", internalSession.SessionID.String()))
		}
	} else {
		r.logger.Info("Found existing session by ID",
			logger.String("session_id", sessionID.String()))
	}

	// Check if we already have an MCP session for this ID
	if mcpSession, ok := r.mcpSessions.Load(sessionID.String()); ok {
		if session, ok := mcpSession.(*MCPClientSession); ok {
			// Update last active time
			session.updateLastActive()
			return session, nil
		}
	}

	// Create new MCP session with the correct session ID
	mcpSession := NewMCPClientSession(sessionID, actor)

	// Copy project data from internal session if it exists
	if internalSession.ProjectID != nil {
		mcpSession.SetProjectID(internalSession.ProjectID)
	}

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
func (r *sessionRegistryImpl) GetSession(sessionID uuid.UUID) (*MCPClientSession, error) {
	// First check the MCP sessions cache
	if mcpSession, ok := r.mcpSessions.Load(sessionID.String()); ok {
		if session, ok := mcpSession.(*MCPClientSession); ok {
			return session, nil
		}
	}

	// If not found in cache, try to get from session manager and create MCP session
	internalSession, err := r.sessionManager.GetSession(sessionID)
	if err != nil {
		// If session doesn't exist by ID, try to find by client ID (for MCP sessions)
		internalSession, err = r.sessionManager.GetSessionByClientID(sessionID.String())
		if err != nil {
			// Session doesn't exist in session manager - create it automatically
			// Create a new internal session with the provided sessionID
			newInternalSession, err := r.sessionManager.CreateSessionWithID(sessionID, sessionID.String())
			if err != nil {
				return nil, fmt.Errorf("failed to create session: %w", err)
			}
			internalSession = newInternalSession
		}
	}

	// Create MCP session with the correct session ID (the one being looked up)
	mcpSession := NewMCPClientSession(sessionID, ActorMCPUser)

	// Copy project data from internal session if it exists
	if internalSession.ProjectID != nil {
		mcpSession.SetProjectID(internalSession.ProjectID)
	}

	// Register with MCP server
	ctx := context.Background()
	if err := r.mcpServer.RegisterSession(ctx, mcpSession); err != nil {
		if err == server.ErrSessionExists {
			// Session already exists in MCP server, unregister first
			r.logger.Warn("Session already exists in MCP server during GetSession, unregistering first",
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

	r.logger.Info("Session recovered and registered",
		logger.String("session_id", sessionID.String()))

	return mcpSession, nil
}

// SetSessionProject sets the project for a session
func (r *sessionRegistryImpl) SetSessionProject(ctx context.Context, sessionID uuid.UUID, projectID *uuid.UUID, actor string) error {
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
func (r *sessionRegistryImpl) RemoveSession(ctx context.Context, sessionID uuid.UUID) error {
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
func (r *sessionRegistryImpl) CleanupExpiredSessions(ctx context.Context, timeout time.Duration) error {
	// Use session manager to cleanup expired sessions
	return r.sessionManager.CleanupExpiredSessions(ctx, timeout)
}

// GetSessionCount returns the number of active sessions
func (r *sessionRegistryImpl) GetSessionCount() int {
	count := 0
	r.mcpSessions.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// SyncExistingSessions syncs existing sessions from the database with MCP server
// This should be called on server startup
func (r *sessionRegistryImpl) SyncExistingSessions(ctx context.Context) error {
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
