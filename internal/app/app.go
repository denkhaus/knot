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
	"github.com/denkhaus/knot/v2/internal/flags"

	"github.com/urfave/cli/v2"
)

// Version variables that will be set by ldflags during build
var (
	version     = "dev"
	commit      = "unknown" // nolint:unused // set by ldflags during build
	date        = "unknown" // nolint:unused // set by ldflags during build
	usage       = "A CLI tool for hierarchical project and task management with dependencies"
	description = `A CLI tool for hierarchical project and task management with dependencies.
Designed to be the best friend of every LLM agent with structured, parsable outputs and comprehensive error handling.
For new users or LLM agents, run 'knot get-started' for a comprehensive guide to all available commands and usage.`
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
		"Configuration error",
		"Failed to initialize MCP server",
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
	flags := append([]cli.Flag{
		flags.NewActorFlag(),
		flags.NewLogLevelFlag(),
	}, append(flags.NewManagerConfigFlags(), append(flags.NewMCPConfigFlags(), flags.NewSyncConfigFlags()...)...)...)

	// Create CLI app
	cliApp := &cli.App{
		Name:        "knot",
		Usage:       usage,
		Description: description,
		Version:     version,
		Authors: []*cli.Author{
			{
				Name:  "denkhaus",
				Email: "denkhaus@github.com",
			},
		},
		Flags: flags,

		After:  commands.NewAfterCommand(),
		Before: commands.NewBeforeCommand(version),
		Commands: []*cli.Command{
			commands.NewProjectCommand(),
			commands.NewTaskCommand(),
			commands.NewTemplateCommand(),
			commands.NewDependencyCommand(),
			commands.NewSyncCommand(),
			commands.NewConfigCommand(),
			commands.NewHealthCommand(),
			commands.NewStatusCommand(),
			commands.NewValidateCommand(),
			commands.NewGetStartedCommand(),
			commands.NewCompletionCommand(),
			commands.NewMCPCommand(),
		},
	}

	return &App{
		App: cliApp,
	}, nil
}

// Run starts the CLI application
func (a *App) Run(args []string) error {

	if err := a.App.Run(args); err != nil {
		// For user input errors, print them cleanly without JSON logging
		if isUserInputError(err) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "💡 For help getting started with Knot and a list of all commands, run: knot get-started\n")
			return err
		}

		// For internal errors, suggest the get-started command
		fmt.Fprintf(os.Stderr, "💡 For help getting started with Knot and a list of all commands, run: knot get-started\n")
		return err
	}

	return nil
}
