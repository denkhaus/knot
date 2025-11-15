package shared

import (
	"context"
	"fmt"

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

// ShowProjectContext displays the current project context if one is selected
// Returns true if context was shown, false if no project is selected
func ShowProjectContext(c *cli.Context, appCtx *AppContext) bool {
	// Skip context display for JSON output or quiet mode
	if c.Bool("json") || c.Bool("quiet") {
		return false
	}

	// Get selected project
	selectedProjectID, err := appCtx.ProjectManager.GetSelectedProject(c.Context)
	if err != nil || selectedProjectID == nil {
		return false
	}

	// Get project details
	project, err := appCtx.ProjectManager.GetProject(context.Background(), *selectedProjectID)
	if err != nil {
		return false
	}

	// Display context indicator
	fmt.Printf("[Project: %s]\n", project.Title)
	return true
}

// ShowProjectContextWithSeparator displays project context with a separator line
func ShowProjectContextWithSeparator(c *cli.Context, appCtx *AppContext) bool {
	if ShowProjectContext(c, appCtx) {
		fmt.Println()
		return true
	}
	return false
}
