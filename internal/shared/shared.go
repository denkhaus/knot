package shared

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/errors"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
)

// ResolveProjectID resolves the project ID from stored context
func ResolveProjectID(c *cli.Context) (uuid.UUID, error) {
	container := GetContainerFromCLIContext(c)
	projectManager := container.GetProjectManager()

	// Get project from database stored context
	if contextProjectID, err := projectManager.GetSelectedProject(c.Context); err == nil && contextProjectID != nil {
		return *contextProjectID, nil
	}

	// No project available
	return uuid.Nil, errors.NoProjectContextError()
}

// ShowProjectContext displays the current project context if one is selected
// Returns true if context was shown, false if no project is selected
func ShowProjectContext(c *cli.Context) bool {
	container := GetContainerFromCLIContext(c)
	projectManager := container.GetProjectManager()

	// Skip context display for JSON output or quiet mode
	if c.Bool("json") || c.Bool("quiet") {
		return false
	}

	// Get selected project
	selectedProjectID, err := projectManager.GetSelectedProject(c.Context)
	if err != nil || selectedProjectID == nil {
		return false
	}

	// Get project details
	project, err := projectManager.GetProject(context.Background(), *selectedProjectID)
	if err != nil {
		return false
	}

	// Display context indicator
	fmt.Printf("[Project: %s]\n", project.Title)
	return true
}

// ShowProjectContextWithSeparator displays project context with a separator line
func ShowProjectContextWithSeparator(c *cli.Context) bool {
	if ShowProjectContext(c) {
		fmt.Println()
		return true
	}
	return false
}
