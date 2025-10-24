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

