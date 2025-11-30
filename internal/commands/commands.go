package commands

import (
	"github.com/denkhaus/knot/v2/internal/commands/completion"
	"github.com/denkhaus/knot/v2/internal/commands/config"
	"github.com/denkhaus/knot/v2/internal/commands/dependency"
	"github.com/denkhaus/knot/v2/internal/commands/health"
	"github.com/denkhaus/knot/v2/internal/commands/project"
	"github.com/denkhaus/knot/v2/internal/commands/task"
	"github.com/denkhaus/knot/v2/internal/commands/template"
	"github.com/denkhaus/knot/v2/internal/commands/validation"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/urfave/cli/v2"
)

func NewDependencyCommand(appCtx *shared.AppContext) *cli.Command {
	return &cli.Command{
		Name:        "dependency",
		Aliases:     []string{"dep"},
		Usage:       "Task dependency management",
		Subcommands: dependency.Commands(appCtx),
	}
}

func NewTemplateCommand(appCtx *shared.AppContext) *cli.Command {
	return &cli.Command{
		Name:        "template",
		Aliases:     []string{"tmpl"},
		Usage:       "Task template management commands",
		Subcommands: template.Commands(appCtx),
	}
}

func NewHealthCommand(appCtx *shared.AppContext) *cli.Command {
	return &cli.Command{
		Name:        "health",
		Usage:       "Database health and connectivity checks",
		Subcommands: health.Commands(appCtx),
	}
}

func NewConfigCommand(appCtx *shared.AppContext) *cli.Command {
	return &cli.Command{
		Name:        "config",
		Aliases:     []string{"cfg"},
		Usage:       "Configuration management",
		Subcommands: config.Commands(appCtx),
	}
}

// Command creates the shell completion command (renamed from CompletionCommand to avoid stuttering)
func NewCompletionCommand(appCtx *shared.AppContext) *cli.Command {
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
		Action: completion.CompletionAction(appCtx),
	}
}

func NewTaskCommand(appCtx *shared.AppContext) *cli.Command {
	return &cli.Command{
		Name:        "task",
		Aliases:     []string{"t"},
		Usage:       "Task management commands",
		Subcommands: task.Commands(appCtx),
	}
}

func NewBreakdownCommand(appCtx *shared.AppContext) *cli.Command {
	return &cli.Command{
		Name:   "breakdown",
		Usage:  "Find tasks that need breakdown based on complexity",
		Action: task.BreakdownAction(appCtx),
		Flags: []cli.Flag{
			shared.NewTaskLimitFlag(),
			shared.NewJSONFlag(),
			shared.NewQuietFlag(),
			&cli.IntFlag{
				Name:    "threshold",
				Aliases: []string{"t"},
				Usage:   "Complexity threshold for breakdown (default: 8)",
				Value:   8,
				EnvVars: []string{"KNOT_COMPLEXITY_THRESHOLD"},
			},
		},
	}
}

func NewBlockedCommand(appCtx *shared.AppContext) *cli.Command {
	return &cli.Command{
		Name:   "blocked",
		Usage:  "Show tasks blocked by dependencies",
		Action: task.BlockedAction(appCtx),
		Flags: []cli.Flag{
			shared.NewTaskLimitFlag(),
			shared.NewJSONFlag(),
		},
	}
}

func NewGetStartedCommand(appCtx *shared.AppContext) *cli.Command {
	return &cli.Command{
		Name:   "get-started",
		Usage:  "Get started guide for LLM agents with available commands and usage",
		Action: task.GetStartedAction(appCtx),
	}
}

func NewReadyCommand(appCtx *shared.AppContext) *cli.Command {
	return &cli.Command{
		Name:   "ready",
		Usage:  "Show tasks with no blockers (ready to work on)",
		Action: task.ReadyAction(appCtx),
		Flags: []cli.Flag{
			shared.NewTaskLimitFlag(),
			shared.NewJSONFlag(),
		},
	}
}

func NewValidateCommand(appCtx *shared.AppContext) *cli.Command {
	return &cli.Command{
		Name:        "validate",
		Usage:       "Task state validation and transition checks",
		Subcommands: validation.Commands(appCtx),
	}
}

func NewProjectCommand(appCtx *shared.AppContext) *cli.Command {
	return &cli.Command{
		Name:        "project",
		Aliases:     []string{"p"},
		Usage:       "Project management commands",
		Subcommands: project.Commands(appCtx),
		Flags: []cli.Flag{
			shared.NewJSONFlag(),
		},
	}
}

// NewActionableCommand creates the enhanced actionable command with new flags
func NewActionableCommand(appCtx *shared.AppContext) *cli.Command {
	return &cli.Command{
		Name:    "actionable",
		Aliases: []string{"next"},
		Usage:   "Find the next actionable task using intelligent selection",
		Description: `Find the next actionable task using dependency-aware selection strategies.

Available strategies:
  - dependency-aware: Prioritizes tasks that unblock others (default)
  - depth-first: Complete subtasks before moving to other branches
  - priority: Focus on high-priority tasks first
  - creation-order: Original knot behavior (oldest first)
  - critical-path: Focus on tasks affecting project timeline

Examples:
  knot task actionable                           # Use default dependency-aware strategy
  knot task actionable --strategy=depth-first   # Prioritize completing branches
  knot task actionable --strategy=priority      # Focus on high-priority tasks
  knot task actionable --verbose --json         # Detailed JSON output`,
		Action: task.ActionableAction(appCtx),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "strategy",
				Aliases: []string{"s"},
				Usage:   "Selection strategy: dependency-aware, depth-first, priority, creation-order, critical-path (auto-recommended if not specified)",
			},
			&cli.BoolFlag{
				Name:  "allow-parent-with-subtasks",
				Usage: "Allow selection of parent tasks even when subtasks exist",
			},
			&cli.BoolFlag{
				Name:  "prefer-pending",
				Usage: "Prefer pending tasks over in-progress tasks",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show detailed selection reasoning and alternatives",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output result as JSON",
			},
		},
	}
}
