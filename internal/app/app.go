// Package app provides the main application entry point and CLI setup for KNOT.
//
// This package contains the main application initialization, command registration,
// and CLI configuration for the KNOT project management system. It orchestrates
// all command modules and provides the unified interface for the knot CLI tool.
//
// Key Components:
//   - App: Main application structure and lifecycle management
//   - Command Registration: Registers all CLI commands and subcommands
//   - CLI Setup: Configures command-line interface and flags
//
// Cross-reference: Knot Task 86f3ba2d-3a87-493b-b8fc-96d19f344e89
package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/denkhaus/knot/v2/internal/commands"
	"github.com/denkhaus/knot/v2/internal/errors"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/repository/inmemory"
	"github.com/denkhaus/knot/v2/internal/repository/sqlite"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/templates"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// Version variables that will be set by ldflags during build
var (
	version = "dev"
	commit  = "unknown" // nolint:unused // set by ldflags during build
	date    = "unknown" // nolint:unused // set by ldflags during build
)

// SetVersionFromBuild allows setting version information from build time variables
func SetVersionFromBuild(v, c, d string) {
	version = v
	commit = c
	date = d
}

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
	appCtx := createAppContext()

	// Create CLI app
	cliApp := &cli.App{
		Name:  "knot",
		Usage: "A CLI tool for hierarchical project and task management with dependencies",
		Description: `A CLI tool for hierarchical project and task management with dependencies.
Designed to be the best friend of every LLM agent with structured, parsable outputs and comprehensive error handling.
For new users or LLM agents, run 'knot get-started' for a comprehensive guide to all available commands and usage.`,
		Version: version,
		Authors: []*cli.Author{
			{
				Name:  "denkhaus",
				Email: "denkhaus@example.com",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "actor",
				Usage:   "Actor name for audit trail (default: $USER)",
				EnvVars: []string{"KNOT_ACTOR", "USER"},
			},
			shared.NewLogLevelFlag(),
		},
		Before: func(c *cli.Context) error {
			// Configure logger based on log-level flag
			logLevel := c.String("log-level")
			logger.SetLogLevel(logLevel)

			// Update appCtx logger reference after reconfiguration
			appCtx.Logger = logger.GetLogger()

			appCtx.SetActor(c.String("actor"))
			appCtx.Logger.Info("Knot CLI started", zap.String("version", version))
			return nil
		},
		Commands: []*cli.Command{
			commands.NewProjectCommand(appCtx),
			commands.NewTaskCommand(appCtx),
			commands.NewTemplateCommand(appCtx),
			commands.NewDependencyCommand(appCtx),
			commands.NewConfigCommand(appCtx),
			commands.NewHealthCommand(appCtx),
			commands.NewStatusCommand(appCtx),
			commands.NewValidateCommand(appCtx),
			commands.NewGetStartedCommand(appCtx),
			commands.NewCompletionCommand(appCtx),
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
			fmt.Fprintf(os.Stderr, "💡 For help getting started with Knot and a list of all commands, run: knot get-started\n")
			return err
		}

		// For internal errors, use the logger but also suggest the get-started command
		a.context.Logger.Error("Application error", zap.Error(err))
		fmt.Fprintf(os.Stderr, "💡 For help getting started with Knot and a list of all commands, run: knot get-started\n")
		return err
	}

	return nil
}

func createAppContext() *shared.AppContext {
	// Initialize logger
	appLogger := logger.GetLogger()

	// Initialize repository (SQLite with fallback to in-memory)
	var repo types.Repository
	var err error

	repo, err = sqlite.NewRepository("",
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
	return shared.NewAppContext(projectManager, appLogger)
}
