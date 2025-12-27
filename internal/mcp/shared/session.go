package shared

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/server"
)

// Actor constants for consistent, type-safe actor handling
const (
	ActorAnonymous = "anonymous"
	ActorMCPUser   = "mcp-user"
)

// GetSessionUUIDFromContext extracts and parses session ID as UUID from MCP context
// Handles the MCP library's session ID format by removing the "mcp-session-" prefix
func GetSessionUUIDFromContext(ctx context.Context) (uuid.UUID, error) {
	session := server.ClientSessionFromContext(ctx)
	if session == nil {
		return uuid.Nil, fmt.Errorf("no session found in context")
	}

	sessionID := session.SessionID()
	if sessionID == "" {
		return uuid.Nil, fmt.Errorf("empty session ID")
	}

	// Extract UUID from session ID (remove "mcp-session-" prefix if present)
	if len(sessionID) > 12 && sessionID[:12] == "mcp-session-" {
		sessionID = sessionID[12:] // Remove prefix
	}

	parsedUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid session ID format: %w", err)
	}

	return parsedUUID, nil
}

// GetSessionActor extracts actor information from session context
// Returns the actor stored in the MCP client session from context
// Returns empty string if no session or no actor is set (no fallback)
func GetSessionActor(ctx context.Context) string {
	session := server.ClientSessionFromContext(ctx)
	if session == nil {
		return ""
	}

	// Get actor from MCP client session if it implements GetActor
	if mcpSession, ok := session.(interface{ GetActor() string }); ok {
		return mcpSession.GetActor()
	}

	return ""
}
