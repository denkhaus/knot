package app

import (
	"fmt"
	"os"
	"strings"

	configCommands "github.com/denkhaus/knot/internal/commands/config"
	"github.com/denkhaus/knot/internal/commands/dependency"
	"github.com/denkhaus/knot/internal/commands/health"
	"github.com/denkhaus/knot/internal/commands/project"
	"github.com/denkhaus/knot/internal/commands/task"
	"github.com/denkhaus/knot/internal/commands/template"
	validationCommands "github.com/denkhaus/knot/internal/commands/validation"
	"github.com/denkhaus/knot/internal/errors"
	"github.com/denkhaus/knot/internal/logger"
	"github.com/denkhaus/knot/internal/manager"
	"github.com/denkhaus/knot/internal/repository/inmemory"
	"github.com/denkhaus/knot/internal/repository/sqlite"
	"github.com/denkhaus/knot/internal/shared"
	"github.com/denkhaus/knot/internal/templates"
	"github.com/denkhaus/knot/internal/types"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// App represents the CLI application
type App struct {
	*cli.App
	context *shared.AppContext
}

// isUserInputError checks if an error is due to user input (like missing required flags)
// rather than an internal application error
func isUserInputError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's an EnhancedError - these are user-facing validation errors
	if _, ok := err.(*errors.EnhancedError); ok {
		return true
	}

	errMsg := err.Error()

	// Common user input errors from urfave/cli
	userErrorPatterns := []string{
		"Required flag",
		"flag provided but not defined",
		"invalid value",
		"command not found",
		"incorrect usage",
		"flag needs an argument",
		"project-id is required",
		"required flag --project-id not provided",
		"No help topic for",
	}

	for _, pattern := range userErrorPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// New creates a new CLI application with all dependencies initialized
func New() (*App, error) {
	// Initialize logger
	appLogger := logger.GetLogger()

	// Initialize repository (SQLite with fallback to in-memory)
	var repo types.Repository
	var err error

	repo, err = sqlite.NewRepository(
		sqlite.WithLogger(appLogger),
		sqlite.WithAutoMigrate(true),
	)
	if err != nil {
		appLogger.Warn("Failed to initialize SQLite repository, falling back to in-memory", zap.Error(err))
		repo = inmemory.NewMemoryRepository()
	} else {
		appLogger.Info("SQLite repository initialized successfully")
		
		// Initialize templates automatically after successful database setup
		if err := templates.CheckAndSeedIfNeeded(); err != nil {
			appLogger.Warn("Failed to seed templates during initialization", zap.Error(err))
		} else {
			appLogger.Debug("Template seeding check completed successfully")
		}
	}

	// Initialize project manager
	config := manager.DefaultConfig()
	projectManager := manager.NewManagerWithRepository(repo, config)

	// Create application context
	appCtx := shared.NewAppContext(projectManager, appLogger)

	// Create CLI app
	cliApp := &cli.App{
		Name:    "knot",
		Usage:   "A CLI tool for hierarchical project and task management with dependencies",
		Version: "1.0.0",
		Authors: []*cli.Author{
			{
				Name:  "denkhaus",
				Email: "denkhaus@example.com",
			},
		},
		Flags: []cli.Flag{
			NewProjectIDFlag(),
			&cli.StringFlag{
				Name:    "actor",
				Usage:   "Actor name for audit trail (default: $USER)",
				EnvVars: []string{"KNOT_ACTOR", "USER"},
			},
			NewLogLevelFlag(),
		},
		Before: func(c *cli.Context) error {
			// Configure logger based on log-level flag
			logLevel := c.String("log-level")
			logger.SetLogLevel(logLevel)
			
			// Update appCtx logger reference after reconfiguration
			appCtx.Logger = logger.GetLogger()
			
			appCtx.SetActor(c.String("actor"))
			appCtx.Logger.Info("Knot CLI started", zap.String("version", "1.0.0"))
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:        "project",
				Aliases:     []string{"p"},
				Usage:       "Project management commands",
				Subcommands: project.Commands(appCtx),
			},
			{
				Name:        "task",
				Aliases:     []string{"t"},
				Usage:       "Task management commands",
				Subcommands: task.Commands(appCtx),
			},
			{
				Name:        "template",
				Aliases:     []string{"tmpl"},
				Usage:       "Task template management commands",
				Subcommands: template.Commands(appCtx),
			},
			{
				Name:        "dependency",
				Aliases:     []string{"dep"},
				Usage:       "Task dependency management",
				Subcommands: dependency.Commands(appCtx),
			},
			{
				Name:        "config",
				Aliases:     []string{"cfg"},
				Usage:       "Configuration management",
				Subcommands: configCommands.Commands(appCtx),
			},
			{
				Name:        "health",
				Usage:       "Database health and connectivity checks",
				Subcommands: health.Commands(appCtx),
			},
			{
				Name:        "validate",
				Usage:       "Task state validation and transition checks",
				Subcommands: validationCommands.Commands(appCtx),
			},
			{
				Name:   "ready",
				Usage:  "Show tasks with no blockers (ready to work on)",
				Action: task.ReadyAction(appCtx),
				Flags: []cli.Flag{
					NewTaskLimitFlag(),
					NewJSONFlag(),
				},
			},
			{
				Name:   "blocked",
				Usage:  "Show tasks blocked by dependencies",
				Action: task.BlockedAction(appCtx),
				Flags: []cli.Flag{
					NewTaskLimitFlag(),
					NewJSONFlag(),
				},
			},
			{
				Name:   "actionable",
				Usage:  "Find the next actionable task in a project",
				Action: task.ActionableAction(appCtx),
				Flags: []cli.Flag{
					NewJSONFlag(),
				},
			},
			{
				Name:   "breakdown",
				Usage:  "Find tasks that need breakdown based on complexity",
				Action: task.BreakdownAction(appCtx),
				Flags: []cli.Flag{
					NewTaskLimitFlag(),
					NewJSONFlag(),
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

	return &App{
		App:     cliApp,
		context: appCtx,
	}, nil
}

// Run starts the CLI application
func (a *App) Run(args []string) error {
	defer logger.Sync()

	if err := a.App.Run(args); err != nil {
		// For user input errors, print them cleanly without JSON logging
		if isUserInputError(err) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return err
		}

		// For internal errors, use the logger
		a.context.Logger.Error("Application error", zap.Error(err))
		return err
	}

	return nil
}
