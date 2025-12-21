package transports

import (
	"context"
	"fmt"
	"net/http"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/mcp/shared"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/server"
	"github.com/samber/do/v2"
)

// CustomSessionIDManager implements the mcp-go SessionIdManager interface
type CustomSessionIDManager struct {
	sessionRegistry shared.SessionRegistry
	logger          logger.Logger
}

// CustomSessionIDManagerResolver implements the mcp-go SessionIdManagerResolver interface
type CustomSessionIDManagerResolver struct {
	sessionRegistry shared.SessionRegistry
	logger          logger.Logger
}

// NewCustomSessionIDManagerResolver creates a new resolver that provides session managers
func NewCustomSessionIDManagerResolver(sessionRegistry shared.SessionRegistry, logger logger.Logger) *CustomSessionIDManagerResolver {
	return &CustomSessionIDManagerResolver{
		sessionRegistry: sessionRegistry,
		logger:          logger,
	}
}

// ResolveSessionIdManager resolves a SessionIdManager based on the HTTP request
func (r *CustomSessionIDManagerResolver) ResolveSessionIdManager(req *http.Request) server.SessionIdManager {
	fmt.Printf("🔍 [CustomSessionIDManagerResolver.ResolveSessionIdManager] CALLED! Request: %s %s\n", req.Method, req.URL.Path)

	// Extract session from header if present
	sessionID := req.Header.Get("X-MCP-Session-ID")
	if sessionID != "" {
		fmt.Printf("✅ [CustomSessionIDManagerResolver] Found session ID in header: %s\n", sessionID)
	} else {
		fmt.Printf("ℹ️ [CustomSessionIDManagerResolver] No session ID in header, will create new session\n")
	}

	// Check all headers for debugging
	fmt.Printf("📋 [CustomSessionIDManagerResolver] All headers: %v\n", req.Header)

	// Return our custom session manager that will handle session creation/validation
	return &CustomSessionIDManager{
		sessionRegistry: r.sessionRegistry,
		logger:          r.logger,
	}
}

// Generate generates a new session ID
func (m *CustomSessionIDManager) Generate() string {
	sessionID := uuid.New()
	fmt.Printf("🔥 [CustomSessionIDManager.Generate] CALLED! Creating session: %s\n", sessionID.String())

	// Create session in our registry
	ctx := context.Background()
	actor := shared.ActorMCPUser
	mcpSession, err := m.sessionRegistry.GetOrCreateSession(ctx, sessionID, actor)
	if err != nil {
		fmt.Printf("❌ [CustomSessionIDManager.Generate] FAILED to create session: %v\n", err)
		// Log error but still return the ID
		// The session will be created lazily when needed
	} else {
		fmt.Printf("✅ [CustomSessionIDManager.Generate] Successfully created session: %s (Actor: %s)\n", mcpSession.SessionID(), actor)
	}

	fmt.Printf("🚀 [CustomSessionIDManager.Generate] Returning session ID: %s\n", sessionID.String())
	return sessionID.String()
}

// Validate validates a session ID
func (m *CustomSessionIDManager) Validate(sessionID string) (bool, error) {
	fmt.Printf("🔍 [CustomSessionIDManager.Validate] CALLED! Validating session: %s\n", sessionID)

	// Extract UUID from session ID (remove "mcp-session-" prefix if present)
	cleanSessionID := sessionID
	if len(sessionID) > 12 && sessionID[:12] == "mcp-session-" {
		cleanSessionID = sessionID[12:] // Remove prefix
		fmt.Printf("✂️ [CustomSessionIDManager.Validate] Removed mcp-session- prefix: %s\n", cleanSessionID)
	}

	// Try to parse the UUID
	uuidSessionID, err := uuid.Parse(cleanSessionID)
	if err != nil {
		fmt.Printf("❌ [CustomSessionIDManager.Validate] Invalid UUID format: %v\n", err)
		return false, fmt.Errorf("invalid session ID format: %w", err)
	}

	fmt.Printf("✅ [CustomSessionIDManager.Validate] Parsed UUID successfully: %s\n", uuidSessionID.String())

	ctx := context.Background()
	actor := shared.ActorMCPUser
	mcpSession, err := m.sessionRegistry.GetOrCreateSession(ctx, uuidSessionID, actor)
	if err != nil {
		fmt.Printf("❌ [CustomSessionIDManager.Validate] FAILED to get/create session: %v\n", err)
		return false, fmt.Errorf("failed to get or create session: %w", err)
	}

	fmt.Printf("✅ [CustomSessionIDManager.Validate] Successfully validated session: %s\n", mcpSession.SessionID())
	// Return false for isTerminated (session is active), nil for error
	return false, nil // Session is valid and not terminated
}

// Terminate terminates a session
func (m *CustomSessionIDManager) Terminate(sessionID string) (bool, error) {
	fmt.Printf("🗑️ [CustomSessionIDManager.Terminate] CALLED! Terminating session: %s\n", sessionID)

	// Extract UUID from session ID (remove "mcp-session-" prefix if present)
	cleanSessionID := sessionID
	if len(sessionID) > 12 && sessionID[:12] == "mcp-session-" {
		cleanSessionID = sessionID[12:] // Remove prefix
		fmt.Printf("✂️ [CustomSessionIDManager.Terminate] Removed mcp-session- prefix: %s\n", cleanSessionID)
	}

	// Try to parse the UUID
	uuidSessionID, err := uuid.Parse(cleanSessionID)
	if err != nil {
		fmt.Printf("❌ [CustomSessionIDManager.Terminate] Invalid session ID format: %v\n", err)
		return false, fmt.Errorf("invalid session ID format: %w", err)
	}

	// Remove session from our registry
	ctx := context.Background()
	err = m.sessionRegistry.RemoveSession(ctx, uuidSessionID)
	if err != nil {
		fmt.Printf("❌ [CustomSessionIDManager.Terminate] Failed to remove session: %v\n", err)
		return false, fmt.Errorf("failed to remove session: %w", err)
	}

	fmt.Printf("✅ [CustomSessionIDManager.Terminate] Successfully terminated session: %s\n", sessionID)
	// Return false for isNotAllowed (termination is allowed), nil for error
	return false, nil // Successfully terminated
}

// SessionWrapper wraps the mcp-go StreamableHTTPServer to automatically register sessions
type SessionWrapper struct {
	*server.StreamableHTTPServer
	sessionRegistry shared.SessionRegistry
	logger          logger.Logger
}

// NewSessionWrapper creates a new session wrapper
func NewSessionWrapper(injector do.Injector) (*SessionWrapper, error) {
	fmt.Printf("🏗️ [NewSessionWrapper] Creating new session wrapper...\n")

	loggerService := do.MustInvoke[logger.Logger](injector)
	sessionRegistry := do.MustInvoke[shared.SessionRegistry](injector)
	mcpServer := do.MustInvoke[*server.MCPServer](injector)

	fmt.Printf("🔧 [NewSessionWrapper] Creating CustomSessionIDManagerResolver with registry\n")

	// Create the resolver that will provide session managers dynamically
	resolver := NewCustomSessionIDManagerResolver(sessionRegistry, loggerService)

	wrapper := &SessionWrapper{
		StreamableHTTPServer: server.NewStreamableHTTPServer(mcpServer,
			// Configure session management - use stateless with our custom resolver
			server.WithStateLess(true),
			server.WithSessionIdManagerResolver(resolver),
		),
		sessionRegistry: sessionRegistry,
		logger:          loggerService,
	}

	fmt.Printf("✅ [NewSessionWrapper] Session wrapper created successfully with resolver\n")
	return wrapper, nil
}

// ServeHTTP wraps the original ServeHTTP to handle session registration
func (sw *SessionWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract session from the request if it exists
	sessionID := r.Header.Get("X-MCP-Session-ID")
	if sessionID != "" {
		// Try to parse the session ID
		if uuid, err := uuid.Parse(sessionID); err == nil {
			// Ensure session exists in our registry
			ctx := context.Background()
			actor := shared.ActorMCPUser
			_, err = sw.sessionRegistry.GetOrCreateSession(ctx, uuid, actor)
			if err != nil {
				sw.logger.Warn("Failed to ensure session exists",
					logger.String("session_id", sessionID),
					logger.Error(err))
			}
		}
	}

	// Call the original handler - sessions will be validated/created via the CustomSessionIDManager
	sw.StreamableHTTPServer.ServeHTTP(w, r)
}
