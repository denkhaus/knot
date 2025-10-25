package main

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Default target to run when none is specified
var Default = Build.Build

// Demo runs a demo workflow
func Demo() error {
	mg.Deps(Build{}.Build)
	fmt.Println("Running demo workflow...")

	commands := [][]string{
		{"./knot", "project", "create", "--name", "Demo Project", "--description", "A demo project"},
		{"./knot", "project", "list"},
		{"./knot", "health", "check"},
	}

	for _, cmd := range commands {
		fmt.Printf("$ %s\n", strings.Join(cmd, " "))
		if err := sh.Run(cmd[0], cmd[1:]...); err != nil {
			return err
		}
		fmt.Println()
	}

	return nil
}

// Help shows available targets with descriptions
func Help() error {
	fmt.Println("Knot Project Lifecycle Management")
	fmt.Println()

	fmt.Println("Main Targets:")
	mainTargets := map[string]string{
		"build":   "Build binary for current platform",
		"install": "Install binary to $GOPATH/bin",
		"run":     "Build and run application",
		"demo":    "Run demo workflow",
		"clean":   "Remove build artifacts",
		"check":   "Run all quality checks",
		"setup":   "Setup development environment",
		"ci":      "Run CI pipeline",
		"version": "Show version information",
	}

	for target, desc := range mainTargets {
		fmt.Printf("  %-12s %s\n", target, desc)
	}

	fmt.Println("\nBuild Targets:")
	fmt.Println("  build:build         Build binary for current platform")
	fmt.Println("  build:deps          Download and install dependencies")
	fmt.Println("  build:install       Install binary to $GOPATH/bin")
	fmt.Println("  build:run           Build and run application")
	fmt.Println("  build:clean         Remove build artifacts")
	fmt.Println("  build:version       Show version information")

	fmt.Println("\nTest Targets:")
	fmt.Println("  test:all            Run all tests")
	fmt.Println("  test:unit           Run unit tests only")
	fmt.Println("  test:integration    Run integration tests")
	fmt.Println("  test:coverage       Run tests with coverage and update badges")
	fmt.Println("  test:badge          Generate coverage badge and update README")
	fmt.Println("  test:race           Run tests with race detection")
	fmt.Println("  test:bench          Run benchmarks")
	fmt.Println("  test:view           Open coverage report")

	fmt.Println("\nUtility Targets:")
	fmt.Println("  util:fmt            Format code")
	fmt.Println("  util:vet            Run go vet")
	fmt.Println("  util:lint           Run golangci-lint")
	fmt.Println("  util:lintfix        Run linter with auto-fix")
	fmt.Println("  util:security       Run security checks")
	fmt.Println("  util:generate       Run go generate")
	fmt.Println("  util:tidy           Tidy go modules")
	fmt.Println("  util:update         Update dependencies")
	fmt.Println("  util:check          Run all quality checks")
	fmt.Println("  util:ci             Run CI pipeline")

	fmt.Println("\nDocker Targets:")
	fmt.Println("  docker:build        Build Docker image")
	fmt.Println("  docker:run          Run Docker container")

	fmt.Println("\nRelease Targets:")
	fmt.Println("  release:build       Build release binaries for all platforms")
	fmt.Println("  release:single      Build release binary for current platform")
	fmt.Println("  release:release     Create a full release using goreleaser")
	fmt.Println("  release:snapshot    Create a snapshot release")
	fmt.Println("  release:dryrun      Perform a dry run of the release process")
	fmt.Println("  release:manifest    Generate release manifest without publishing")

	fmt.Println("\nUsage: mage <target>")
	fmt.Println("Default target: build")
	return nil
}
