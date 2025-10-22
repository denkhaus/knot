package task

import (
	"embed"
	"fmt"

	"github.com/denkhaus/knot/internal/shared"
	"github.com/urfave/cli/v2"
)

//go:embed get_started.md
var getStartedFS embed.FS

// GetStartedAction provides a summary of available commands for LLM agents
func GetStartedAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		content, err := getStartedFS.ReadFile("get_started.md")
		if err != nil {
			return fmt.Errorf("failed to read get-started content: %w", err)
		}
		
		fmt.Print(string(content))
		return nil
	}
}