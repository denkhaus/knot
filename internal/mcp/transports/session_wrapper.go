package transports

import (
	"context"
	"net/http"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/mcp/shared"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/server"
)

// CustomSessionIDManager implements the mcp-go SessionIdManager interface
type CustomSessionIDManager struct {
	sessionRegistry *shared.SessionRegistry
}

// Generate generates a new session ID
func (m *CustomSessionIDManager) Generate() string {
	sessionID := uuid.New()

	// Create session in our registry
	ctx := context.Background()
	actor := shared.ActorMCPUser
	_, err := m.sessionRegistry.GetOrCreateSession(ctx, sessionID, actor)
	if err != nil {
		// Log error but still return the ID
		// The session will be created lazily when needed
	}

	return sessionID.String()
}

// Validate validates a session ID
func (m *CustomSessionIDManager) Validate(sessionID string) (bool, error) {
	// Try to parse the UUID
	if _, err := uuid.Parse(sessionID); err != nil {
		return false, nil
	}

	// Check if session exists in our registry or can be created
	uuidSessionID, err := uuid.Parse(sessionID)
	if err != nil {
		return false, nil
	}

	ctx := context.Background()
	actor := shared.ActorMCPUser
	_, err = m.sessionRegistry.GetOrCreateSession(ctx, uuidSessionID, actor)

	// If session doesn't exist and we can't create it, that's an error
	if err != nil {
		return false, nil
	}

	return true, nil
}

// Terminate terminates a session
func (m *CustomSessionIDManager) Terminate(sessionID string) (bool, error) {
	// Try to parse the UUID
	uuidSessionID, err := uuid.Parse(sessionID)
	if err != nil {
		return false, nil // Invalid session ID
	}

	// Remove session from our registry
	ctx := context.Background()
	err = m.sessionRegistry.RemoveSession(ctx, uuidSessionID)
	if err != nil {
		return false, nil // Error removing session
	}

	return true, nil // Successfully terminated
}

// SessionWrapper wraps the mcp-go StreamableHTTPServer to automatically register sessions
type SessionWrapper struct {
	*server.StreamableHTTPServer
	sessionRegistry *shared.SessionRegistry
	logger          logger.Logger
}

// NewSessionWrapper creates a new session wrapper
func NewSessionWrapper(mcpServer *server.MCPServer, sessionRegistry *shared.SessionRegistry, logger logger.Logger) *SessionWrapper {
	return &SessionWrapper{
		StreamableHTTPServer: server.NewStreamableHTTPServer(
			mcpServer,
			// Configure session management - use stateless with our custom ID manager
			server.WithSessionIdManager(&CustomSessionIDManager{sessionRegistry: sessionRegistry}),
		),
		sessionRegistry: sessionRegistry,
		logger:          logger,
	}
}

// ServeHTTP wraps the original ServeHTTP to handle session registration
func (sw *SessionWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Call the original handler - sessions will be validated/created via the CustomSessionIDManager
	sw.StreamableHTTPServer.ServeHTTP(w, r)
}
