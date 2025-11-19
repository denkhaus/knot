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
