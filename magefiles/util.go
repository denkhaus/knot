package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	binaryName   = "knot"
	packagePath  = "./cmd/knot"
	coverageFile = "coverage.out"
	coverageHTML = "coverage.html"
)

// Util namespace for utility targets
type Util mg.Namespace

// Fmt formats the code
func (Util) Fmt() error {
	fmt.Println("Formatting code...")
	return sh.Run("go", "fmt", "./...")
}

// Vet runs go vet
func (Util) Vet() error {
	fmt.Println("Running go vet...")
	return sh.Run("go", "vet", "./...")
}

// Generate runs go generate
func (Util) Generate() error {
	fmt.Println("Running go generate...")
	return sh.Run("go", "generate", "./...")
}

// Tidy cleans up go.mod and go.sum
func (Util) Tidy() error {
	fmt.Println("Tidying go modules...")
	return sh.Run("go", "mod", "tidy")
}

// Lint runs golangci-lint
func (Util) Lint() error {
	fmt.Println("Running linter...")
	return sh.Run("golangci-lint", "run", "--timeout=5m")
}

// LintFix runs golangci-lint with auto-fix
func (Util) LintFix() error {
	fmt.Println("Running linter with auto-fix...")
	return sh.Run("golangci-lint", "run", "--fix", "--timeout=5m")
}

// Security runs security checks
func (Util) Security() error {
	fmt.Println("Running security checks...")

	// Check if gosec is installed
	if _, err := exec.LookPath("gosec"); err != nil {
		fmt.Println("Installing gosec...")
		if err := sh.Run("go", "install", "github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"); err != nil {
			return err
		}
	}

	return sh.Run("gosec", "./...")
}

// Update updates all dependencies
func (Util) Update() error {
	fmt.Println("Updating dependencies...")

	if err := sh.Run("go", "get", "-u", "./..."); err != nil {
		return err
	}

	return sh.Run("go", "mod", "tidy")
}

// Setup sets up development environment
func (Util) Setup() error {
	fmt.Println("Setting up development environment...")

	tools := []string{
		"github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
		"github.com/goreleaser/goreleaser@latest",
		"github.com/securecodewarrior/gosec/v2/cmd/gosec@latest",
	}

	for _, tool := range tools {
		fmt.Printf("Installing %s...\n", tool)
		if err := sh.Run("go", "install", tool); err != nil {
			return err
		}
	}

	fmt.Println("Development environment ready!")
	return nil
}

// Check runs all quality checks
func (Util) Check() error {
	fmt.Println("Running all quality checks...")
	mg.Deps(Util.Fmt, Util.Vet, Util.Lint, Test.Race)
	fmt.Println("All checks passed!")
	return nil
}

// CI runs all CI checks
func (Util) CI() error {
	fmt.Println("Running CI pipeline...")
	mg.Deps(Util.Tidy, Util.Generate, Util.Check, Test.Coverage)
	fmt.Println("CI pipeline completed successfully!")
	return nil
}

// Docker namespace for Docker-related targets
type Docker mg.Namespace

// Build builds Docker image
func (Docker) Build() error {
	mg.Deps(Test.All)
	fmt.Println("Building Docker image...")
	return sh.Run("docker", "build", "-t", "knot:latest", ".")
}

// Run runs Docker container
func (Docker) Run() error {
	mg.Deps(Docker.Build)
	fmt.Println("Running Docker container...")
	return sh.Run("docker", "run", "--rm", "-it", "knot:latest", "--help")
}

// Helper functions
func getVersion() string {
	if version := os.Getenv("VERSION"); version != "" {
		return version
	}

	// Try to get version from git tag
	if output, err := sh.Output("git", "describe", "--tags", "--always", "--dirty"); err == nil {
		return strings.TrimSpace(output)
	}

	return "dev"
}

func getCommit() string {
	if output, err := sh.Output("git", "rev-parse", "--short", "HEAD"); err == nil {
		return strings.TrimSpace(output)
	}
	return "unknown"
}
