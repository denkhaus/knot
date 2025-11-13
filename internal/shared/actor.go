package shared

import (
	"os"
	"github.com/urfave/cli/v2"
)

// GetActorFromContext resolves the actor from CLI context with proper fallback logic
// This function eliminates duplicate actor resolution code across CLI commands
//
// Priority order:
// 1. CLI --actor flag (if provided)
// 2. urfave/cli default (USER env var automatically handled)
// 3. "unknown" as final fallback
//
// Parameters:
// - c: CLI context from urfave/cli/v2
//
// Returns:
// - string: resolved actor name
//
// Usage example:
//   actor := shared.GetActorFromContext(c)
//
// Related to task: a42c4861-f7f7-4d03-9a29-53b965a7ee1e
func GetActorFromContext(c *cli.Context) string {
	actor := c.String("actor")
	if actor == "" {
		actor = "unknown"
	}
	return actor
}

// ResolveActor resolves actor with fallback logic for non-CLI contexts
// This function can be used in helper functions that receive actor as parameter
func ResolveActor(actor string) string {
	if actor == "" {
		actor = os.Getenv("USER")
		if actor == "" {
			actor = "unknown"
		}
	}
	return actor
}