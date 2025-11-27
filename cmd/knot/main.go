// Package main provides the entry point for the KNOT CLI application.
//
// This is the main package that initializes and runs the KNOT project management
// command-line interface. It handles application startup, version information,
// and graceful shutdown procedures.
//
// Key Responsibilities:
//   - Application Entry: Main entry point and initialization
//   - CLI Startup: Launches the CLI application
//   - Version Info: Provides build and version information
//   - Error Handling: Handles application-level errors
//
// Cross-reference: Knot Task 86f3ba2d-3a87-493b-b8fc-96d19f344e89
package main

import (
	"log"
	"os"

	"github.com/denkhaus/knot/v2/internal/app"
)

// Version, commit, and build date are set by ldflags during build
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Set version information from build variables
	app.SetVersionFromBuild(version, commit, date)

	// Create and run the application
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := application.Run(os.Args); err != nil {
		// Error has already been printed by the Run method
		// Just exit with error code without additional logging
		os.Exit(1)
	}
}
