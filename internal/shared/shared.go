package shared

import (
	"github.com/denkhaus/knot/internal/errors"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
)

// ResolveProjectID resolves the project ID from stored context
func ResolveProjectID(c *cli.Context, appCtx *AppContext) (uuid.UUID, error) {
	// Get project from database stored context
	if contextProjectID, err := appCtx.ProjectManager.GetSelectedProject(c.Context); err == nil && contextProjectID != nil {
		return *contextProjectID, nil
	}
	
	// No project available
	return uuid.Nil, errors.NoProjectContextError()
}

// ValidateProjectID validates and returns the project ID from the CLI context
// DEPRECATED: Use ResolveProjectID instead for context-aware resolution
func ValidateProjectID(c *cli.Context) (uuid.UUID, error) {
	projectIDStr := c.String("project-id")
	if projectIDStr == "" {
		return uuid.Nil, errors.MissingRequiredFlagError("project-id", c.Command.FullName())
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return uuid.Nil, errors.InvalidUUIDError("project-id", projectIDStr)
	}

	return projectID, nil
}
