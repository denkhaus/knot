package shared

import (
	"github.com/denkhaus/knot/internal/errors"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
)

// ValidateProjectID validates and returns the project ID from the CLI context
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
