package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/server"
)

// getSessionID extracts session ID from MCP context
func getSessionID(ctx context.Context) string {
	if session := server.ClientSessionFromContext(ctx); session != nil {
		return session.SessionID()
	}
	return ""
}

// getSessionActor extracts actor information from session context
// For now, returns a default actor since we don't have user authentication
// TODO(knot-lom): the actor must be defined in the process of session creation to not have a dummy value
func getSessionActor(ctx context.Context) string {
	// dummy value
	return "mcp-user"
}
