package commands

import (
	"github.com/denkhaus/knot/v2/internal/commands/completion"
	"github.com/denkhaus/knot/v2/internal/commands/config"
	"github.com/denkhaus/knot/v2/internal/commands/dependency"
	"github.com/denkhaus/knot/v2/internal/commands/health"
	cmdmcp "github.com/denkhaus/knot/v2/internal/commands/mcp"
	"github.com/denkhaus/knot/v2/internal/commands/project"
	"github.com/denkhaus/knot/v2/internal/commands/task"
	"github.com/denkhaus/knot/v2/internal/commands/template"
	"github.com/denkhaus/knot/v2/internal/commands/validation"
	"github.com/denkhaus/knot/v2/internal/di"
	"github.com/denkhaus/knot/v2/internal/flags"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/templates"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// NewBeforeCommand creates a Before action function that initializes DI and configures services
func NewBeforeCommand(container *di.Container, version string) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		// Initialize DI container with all services
		injector := container.RegisterAllServices(c.Context, c)

		// For now, keep using the global logger approach until full migration
		logLevel := c.String("log-level")
		logger.SetLogLevel(logLevel)

		// Use DI template service for seeding instead of static function
		templateService := do.MustInvoke[templates.Service](injector)
		loggerService := do.MustInvoke[logger.Logger](injector)

		if err := templateService.CheckAndSeedIfNeeded(); err != nil {
			loggerService.Warn("Failed to seed templates during initialization", zap.Error(err))
		} else {
			loggerService.Debug("Template seeding check completed successfully")
		}

		// Store the injector in the container for later use by commands
		container.SetInjector(injector)

		// Add commands now that DI is initialized
		c.App.Commands = []*cli.Command{
			NewProjectCommand(injector),
			NewTaskCommand(injector),
			NewTemplateCommand(injector),
			NewDependencyCommand(injector),
			NewConfigCommand(injector),
			NewHealthCommand(injector),
			NewStatusCommand(injector),
			NewValidateCommand(injector),
			NewGetStartedCommand(injector),
			NewCompletionCommand(injector),
			NewMCPCommand(injector),
		}

		loggerService.Info("Knot CLI started with DI", zap.String("version", version))

		return nil
	}
}

func NewDependencyCommand(injector do.Injector) *cli.Command {
	return &cli.Command{
		Name:        "dependency",
		Aliases:     []string{"dep"},
		Usage:       "Task dependency management",
		Subcommands: dependency.Commands(injector),
	}
}

func NewTemplateCommand(injector do.Injector) *cli.Command {
	return &cli.Command{
		Name:        "template",
		Aliases:     []string{"tmpl"},
		Usage:       "Task template management commands",
		Subcommands: template.Commands(injector),
	}
}

func NewHealthCommand(injector do.Injector) *cli.Command {
	return &cli.Command{
		Name:        "health",
		Usage:       "Database health and connectivity checks",
		Subcommands: health.Commands(injector),
	}
}

func NewConfigCommand(injector do.Injector) *cli.Command {
	return &cli.Command{
		Name:        "config",
		Aliases:     []string{"cfg"},
		Usage:       "Configuration management",
		Subcommands: config.Commands(injector),
	}
}

// Command creates the shell completion command (renamed from CompletionCommand to avoid stuttering)
func NewCompletionCommand(injector do.Injector) *cli.Command {
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
		Action: completion.CompletionAction(injector),
	}
}

func NewTaskCommand(injector do.Injector) *cli.Command {
	return &cli.Command{
		Name:        "task",
		Aliases:     []string{"t"},
		Usage:       "Task management commands",
		Subcommands: task.Commands(injector),
	}
}

func NewGetStartedCommand(injector do.Injector) *cli.Command {
	return &cli.Command{
		Name:   "get-started",
		Usage:  "Get started guide for LLM agents with available commands and usage",
		Action: task.GetStartedAction(injector),
	}
}

func NewValidateCommand(injector do.Injector) *cli.Command {
	return &cli.Command{
		Name:        "validate",
		Usage:       "Task state validation and transition checks",
		Subcommands: validation.Commands(injector),
	}
}

func NewProjectCommand(injector do.Injector) *cli.Command {
	return &cli.Command{
		Name:        "project",
		Aliases:     []string{"p"},
		Usage:       "Project management commands",
		Subcommands: project.Commands(injector),
		Flags: []cli.Flag{
			flags.NewJSONFlag(),
		},
	}
}

// NewStatusCommand creates a status command with subcommands for different status views
func NewStatusCommand(injector do.Injector) *cli.Command {
	return &cli.Command{
		Name:    "status",
		Usage:   "Show tasks by status (actionable, ready, blocked, breakdown)",
		Aliases: []string{"st"},
		Description: `Status commands provide different views into your task workflow:

  actionable: Find the next actionable task using intelligent selection
  ready:     Show tasks with no blockers (ready to work on)
  blocked:   Show tasks blocked by dependencies
  breakdown: Find tasks that need breakdown based on complexity

Examples:
  knot status actionable                           # Use default dependency-aware strategy
  knot status actionable --strategy=priority      # Focus on high-priority tasks
  knot status ready --json                        # Show ready tasks in JSON format
  knot status blocked --limit=10                  # Show first 10 blocked tasks
  knot status breakdown --threshold=5             # Find tasks needing breakdown (complexity >= 5)`,
		Subcommands: []*cli.Command{
			{
				Name:    "actionable",
				Aliases: []string{"next", "act"},
				Usage:   "Find the next actionable task using intelligent selection",
				Description: `Find the next actionable task using dependency-aware selection strategies.

Available strategies:
  - dependency-aware: Prioritizes tasks that unblock others (default)
  - depth-first: Complete subtasks before moving to other branches
  - priority: Focus on high-priority tasks first
  - creation-order: Original knot behavior (oldest first)
  - critical-path: Focus on tasks affecting project timeline

Examples:
  knot status actionable                           # Use default dependency-aware strategy
  knot status actionable --strategy=depth-first   # Prioritize completing branches
  knot status actionable --strategy=priority      # Focus on high-priority tasks
  knot status actionable --verbose --json         # Detailed JSON output`,
				Action: task.ActionableAction(injector),
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
			},
			{
				Name:    "ready",
				Aliases: []string{"rd"},
				Usage:   "Show tasks with no blockers (ready to work on)",
				Action:  task.ReadyAction(injector),
				Flags: []cli.Flag{
					flags.NewTaskLimitFlag(),
					flags.NewJSONFlag(),
				},
			},
			{
				Name:    "blocked",
				Aliases: []string{"blk"},
				Usage:   "Show tasks blocked by dependencies",
				Action:  task.BlockedAction(injector),
				Flags: []cli.Flag{
					flags.NewTaskLimitFlag(),
					flags.NewJSONFlag(),
				},
			},
			{
				Name:    "breakdown",
				Aliases: []string{"bd"},
				Usage:   "Find tasks that need breakdown based on complexity",
				Action:  task.BreakdownAction(injector),
				Flags: []cli.Flag{
					flags.NewTaskLimitFlag(),
					flags.NewJSONFlag(),
					flags.NewQuietFlag(),
					&cli.IntFlag{
						Name:    "threshold",
						Aliases: []string{"t"},
						Usage:   "Complexity threshold for breakdown (default: 8)",
						Value:   8,
						EnvVars: []string{"KNOT_COMPLEXITY_THRESHOLD"},
					},
				},
			},
		},
	}
}

// NewMCPCommand creates a Model Context Protocol (MCP) server command
func NewMCPCommand(injector do.Injector) *cli.Command {
	return &cli.Command{
		Name:  "mcp",
		Usage: "Model Context Protocol (MCP) server commands",
		Description: `Model Context Protocol (MCP) server provides a central hub for multi-project
task management that can be accessed by MCP-compatible clients.

The MCP server supports:
  - Multiple concurrent projects with session-based selection
  - Curated set of core Knot operations
  - Basic hint system for next-action guidance
  - Thread-safe operations for multiple clients

Examples:
  knot mcp server                                    # Start MCP server with defaults
  knot mcp server --address 0.0.0.0 --port 9090     # Start on custom address/port
  knot mcp server --log-level debug                 # Enable debug logging`,
		Subcommands: cmdmcp.Commands(injector),
	}
}
