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
// For now, returns a default actor since we don't have user authentication
// TODO(knot-lom): the actor must be defined in the process of session creation to not have a dummy value
func GetSessionActor(ctx context.Context) string {
	return ActorMCPUser
}
