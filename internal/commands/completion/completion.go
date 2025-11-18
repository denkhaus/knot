package completion

import (
	"fmt"

	"github.com/denkhaus/knot/internal/shared"
	"github.com/urfave/cli/v2"
)

// ShellType represents supported shell types
type ShellType string

const (
	ShellBash ShellType = "bash"
	ShellZsh  ShellType = "zsh"
)

// CompletionCommand creates the shell completion command
func CompletionCommand(appCtx *shared.AppContext) *cli.Command {
	return &cli.Command{
		Name:  "completion",
		Usage: "Generate shell completion scripts",
		Description: `Generate shell completion scripts for bash and zsh.
These scripts enable tab completion for knot commands, flags, and dynamic values.

Examples:
  knot completion bash > /etc/bash_completion.d/knot
  knot completion zsh > /usr/local/share/zsh-completions/_knot
  source <(knot completion bash)

Installation:
  # Bash (system-wide)
  knot completion bash | sudo tee /etc/bash_completion.d/knot

  # Bash (user-specific)
  knot completion bash > ~/.local/share/bash-completion/completions/knot

  # Zsh (system-wide)
  knot completion zsh | sudo tee /usr/local/share/zsh-completions/_knot

  # Zsh (user-specific)
  mkdir -p ~/.zsh/completions
  knot completion zsh > ~/.zsh/completions/_knot`,
		Action: CompletionAction(appCtx),
	}
}

// CompletionAction implements the shell completion functionality
func CompletionAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		args := c.Args()
		if args.Len() != 1 {
			return fmt.Errorf("usage: knot completion <bash|zsh>")
		}

		shell := ShellType(args.Get(0))
		return GenerateCompletion(shell)
	}
}

// GenerateCompletion generates completion script for the specified shell
func GenerateCompletion(shell ShellType) error {
	switch shell {
	case ShellBash:
		completion := NewBashCompletion()
		return completion.Generate()
	case ShellZsh:
		completion := NewZshCompletion()
		return completion.Generate()
	default:
		return fmt.Errorf("unsupported shell: %s. Supported shells: bash, zsh", shell)
	}
}