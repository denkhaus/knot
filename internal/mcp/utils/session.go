package utils

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/mcp/shared"
	"github.com/denkhaus/knot/v2/internal/session"
)

// TODO: move this to the shared package
// GetSelectedProject returns the project ID from the current session.
// Used by tools that should only work with the currently selected project.
// Returns a user-friendly error message suggesting to use project_select first.
func GetSelectedProject(
	ctx context.Context,
	sessionManager session.SessionManager,
) (string, error) {
	// Get session UUID from context
	sessionUUID, err := shared.GetSessionUUIDFromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("no session found")
	}

	// Get session using session manager with client ID lookup
	sess, err := sessionManager.GetSessionByClientID(sessionUUID.String())
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session has a project selected
	if sess.ProjectID == nil {
		return "", fmt.Errorf("no project selected. Please use the 'project_select' tool to select a project first")
	}

	return sess.ProjectID.String(), nil
}
