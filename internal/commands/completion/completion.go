package completion

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// ShellType represents supported shell types
type ShellType string

const (
	ShellBash ShellType = "bash"
	ShellZsh  ShellType = "zsh"
)

// CompletionAction implements the shell completion functionality
func CompletionAction() cli.ActionFunc {
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
