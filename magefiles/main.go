//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Default target to run when none is specified
var Default = Build

const (
	binaryName   = "knot"
	packagePath  = "./cmd/knot"
	coverageFile = "coverage.out"
	coverageHTML = "coverage.html"
)

// Namespaces for organizing targets
type Test mg.Namespace
type Docker mg.Namespace
type Release mg.Namespace

// Build builds the binary for current platform
func Build() error {
	mg.Deps(Deps)
	fmt.Println("Building binary...")
	
	ldflags := fmt.Sprintf("-s -w -X main.version=%s -X main.commit=%s -X main.date=%s",
		getVersion(), getCommit(), time.Now().Format(time.RFC3339))
	
	return sh.Run("go", "build", "-ldflags", ldflags, "-o", binaryName, packagePath)
}

// Deps downloads and installs dependencies
func Deps() error {
	fmt.Println("Installing dependencies...")
	return sh.Run("go", "mod", "download")
}

// Install installs the binary to $GOPATH/bin
func Install() error {
	mg.Deps(Deps)
	fmt.Println("Installing binary...")
	return sh.Run("go", "install", packagePath)
}

// Run builds and runs the application with help
func Run() error {
	mg.Deps(Build)
	fmt.Println("Running application...")
	return sh.Run("./"+binaryName, "--help")
}

// Demo runs a demo workflow
func Demo() error {
	mg.Deps(Build)
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

// Clean removes build artifacts
func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	
	artifacts := []string{
		binaryName,
		binaryName + ".exe",
		"bin/",
		"dist/",
		coverageFile,
		coverageHTML,
	}
	
	for _, artifact := range artifacts {
		if err := os.RemoveAll(artifact); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	
	// Clean Go cache
	sh.Run("go", "clean", "-cache")
	sh.Run("go", "clean", "-testcache")
	
	fmt.Println("Clean completed!")
	return nil
}

// Fmt formats the code
func Fmt() error {
	fmt.Println("Formatting code...")
	return sh.Run("go", "fmt", "./...")
}

// Vet runs go vet
func Vet() error {
	fmt.Println("Running go vet...")
	return sh.Run("go", "vet", "./...")
}

// Generate runs go generate
func Generate() error {
	fmt.Println("Running go generate...")
	return sh.Run("go", "generate", "./...")
}

// Tidy cleans up go.mod and go.sum
func Tidy() error {
	fmt.Println("Tidying go modules...")
	return sh.Run("go", "mod", "tidy")
}

// Lint runs golangci-lint
func Lint() error {
	fmt.Println("Running linter...")
	return sh.Run("golangci-lint", "run", "--timeout=5m")
}

// LintFix runs golangci-lint with auto-fix
func LintFix() error {
	fmt.Println("Running linter with auto-fix...")
	return sh.Run("golangci-lint", "run", "--fix", "--timeout=5m")
}

// Security runs security checks
func Security() error {
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

// Check runs all quality checks
func Check() error {
	fmt.Println("Running all quality checks...")
	mg.Deps(Fmt, Vet, Lint, Test.Race)
	fmt.Println("All checks passed!")
	return nil
}

// Setup sets up development environment
func Setup() error {
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

// CI runs all CI checks
func CI() error {
	fmt.Println("Running CI pipeline...")
	mg.Deps(Tidy, Generate, Check, Test.Coverage)
	fmt.Println("CI pipeline completed successfully!")
	return nil
}

// Update updates all dependencies
func Update() error {
	fmt.Println("Updating dependencies...")
	
	if err := sh.Run("go", "get", "-u", "./..."); err != nil {
		return err
	}
	
	return sh.Run("go", "mod", "tidy")
}

// Version shows version information
func Version() error {
	fmt.Printf("Version: %s\n", getVersion())
	fmt.Printf("Commit: %s\n", getCommit())
	fmt.Printf("Date: %s\n", time.Now().Format(time.RFC3339))
	return nil
}

// Help shows available targets with descriptions
func Help() error {
	fmt.Println("Knot Project Lifecycle Management")
	fmt.Println()
	
	fmt.Println("Main Targets:")
	mainTargets := map[string]string{
		"build":    "Build binary for current platform",
		"install":  "Install binary to $GOPATH/bin",
		"run":      "Build and run application",
		"demo":     "Run demo workflow",
		"clean":    "Remove build artifacts",
		"check":    "Run all quality checks",
		"setup":    "Setup development environment",
		"ci":       "Run CI pipeline",
		"version":  "Show version information",
	}
	
	for target, desc := range mainTargets {
		fmt.Printf("  %-12s %s\n", target, desc)
	}
	
	fmt.Println("\nTest Targets:")
	fmt.Println("  test:all        Run all tests")
	fmt.Println("  test:unit       Run unit tests only")
	fmt.Println("  test:integration Run integration tests")
	fmt.Println("  test:coverage   Run tests with coverage")
	fmt.Println("  test:race       Run tests with race detection")
	fmt.Println("  test:bench      Run benchmarks")
	fmt.Println("  test:view       Open coverage report")
	
	fmt.Println("\nDocker Targets:")
	fmt.Println("  docker:build    Build Docker image")
	fmt.Println("  docker:run      Run Docker container")
	
	fmt.Println("\nRelease Targets:")
	fmt.Println("  release:build   Build for all platforms")
	fmt.Println("  release:dry     Run dry release")
	fmt.Println("  release:create  Create release with GoReleaser")
	
	fmt.Println("\nUtility Targets:")
	utilTargets := map[string]string{
		"fmt":       "Format code",
		"vet":       "Run go vet",
		"lint":      "Run golangci-lint",
		"lintfix":   "Run linter with auto-fix",
		"security":  "Run security checks",
		"generate":  "Run go generate",
		"tidy":      "Tidy go modules",
		"update":    "Update dependencies",
	}
	
	for target, desc := range utilTargets {
		fmt.Printf("  %-12s %s\n", target, desc)
	}
	
	fmt.Println("\nUsage: mage <target>")
	fmt.Println("Default target: build")
	return nil
}

// Test namespace targets
func (Test) All() error {
	fmt.Println("Running all tests...")
	return sh.Run("go", "test", "-v", "./...")
}

func (Test) Unit() error {
	fmt.Println("Running unit tests...")
	return sh.Run("go", "test", "-short", "-v", "./...")
}

func (Test) Integration() error {
	fmt.Println("Running integration tests...")
	return sh.Run("go", "test", "-run", "Integration", "-v", "./...")
}

func (Test) Coverage() error {
	fmt.Println("Running tests with coverage...")
	if err := sh.Run("go", "test", "-coverprofile="+coverageFile, "-covermode=atomic", "./..."); err != nil {
		return err
	}
	
	if err := sh.Run("go", "tool", "cover", "-html="+coverageFile, "-o", coverageHTML); err != nil {
		return err
	}
	
	fmt.Println("Coverage report generated: " + coverageHTML)
	return nil
}

func (Test) Race() error {
	fmt.Println("Running tests with race detection...")
	return sh.Run("go", "test", "-race", "-v", "./...")
}

func (Test) Bench() error {
	fmt.Println("Running benchmarks...")
	return sh.Run("go", "test", "-bench=.", "-benchmem", "./...")
}

func (Test) View() error {
	mg.Deps(Test.Coverage)
	fmt.Println("Opening coverage report...")
	
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	case "windows":
		cmd = "start"
	default:
		return fmt.Errorf("unsupported platform")
	}
	
	return sh.Run(cmd, coverageHTML)
}

// Docker namespace targets
func (Docker) Build() error {
	mg.Deps(Test.All)
	fmt.Println("Building Docker image...")
	return sh.Run("docker", "build", "-t", "knot:latest", ".")
}

func (Docker) Run() error {
	mg.Deps(Docker.Build)
	fmt.Println("Running Docker container...")
	return sh.Run("docker", "run", "--rm", "-it", "knot:latest", "--help")
}

// Release namespace targets
func (Release) Build() error {
	mg.Deps(Test.All)
	fmt.Println("Building for all platforms...")

	platforms := []struct {
		goos, goarch string
	}{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
	}

	// Create dist directory
	if err := os.MkdirAll("dist", 0755); err != nil {
		return err
	}

	for _, platform := range platforms {
		env := map[string]string{
			"GOOS":        platform.goos,
			"GOARCH":      platform.goarch,
			"CGO_ENABLED": "0",
		}

		binary := fmt.Sprintf("dist/%s-%s-%s", binaryName, platform.goos, platform.goarch)
		if platform.goos == "windows" {
			binary += ".exe"
		}

		ldflags := fmt.Sprintf("-s -w -X main.version=%s -X main.commit=%s -X main.date=%s",
			getVersion(), getCommit(), time.Now().Format(time.RFC3339))

		fmt.Printf("Building %s...\n", binary)
		if err := sh.RunWith(env, "go", "build", "-ldflags", ldflags, "-o", binary, packagePath); err != nil {
			return err
		}
	}

	fmt.Println("All binaries built successfully!")
	return nil
}

func (Release) Dry() error {
	mg.Deps(Check)
	fmt.Println("Running dry release...")
	return sh.Run("goreleaser", "release", "--snapshot", "--clean")
}

func (Release) Create() error {
	mg.Deps(Check)
	fmt.Println("Creating release...")
	return sh.Run("goreleaser", "release", "--clean")
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