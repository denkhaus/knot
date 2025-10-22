//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Test namespace for test-related targets
type Test mg.Namespace

// Build builds the knot binary
func Build() error {
	fmt.Println("Building knot binary...")
	return sh.Run("go", "build", "-o", "bin/knot", "./cmd/knot")
}

// Install installs the knot binary
func Install() error {
	fmt.Println("Installing knot binary...")
	return sh.Run("go", "install", "./cmd/knot")
}

// Clean removes build artifacts and test cache
func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	
	// Clean Go cache
	if err := sh.Run("go", "clean"); err != nil {
		return err
	}
	
	// Clean test cache
	if err := sh.Run("go", "clean", "-testcache"); err != nil {
		return err
	}
	
	// Remove coverage files
	os.Remove("coverage.out")
	os.Remove("coverage.html")
	
	// Remove bin directory
	os.RemoveAll("bin")
	
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

// (Test) All runs all tests
func (Test) All() error {
	fmt.Println("Running all tests...")
	return sh.Run("go", "test", "./...")
}

// (Test) Unit runs unit tests only
func (Test) Unit() error {
	fmt.Println("Running unit tests...")
	return sh.Run("go", "test", "-short", "./...")
}

// (Test) Integration runs integration tests only
func (Test) Integration() error {
	fmt.Println("Running integration tests...")
	return sh.Run("go", "test", "-run", "Integration", "./...")
}

// (Test) Coverage runs tests with coverage report
func (Test) Coverage() error {
	fmt.Println("Running tests with coverage...")
	
	if err := sh.Run("go", "test", "-coverprofile=coverage.out", "./..."); err != nil {
		return err
	}
	
	if err := sh.Run("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html"); err != nil {
		return err
	}
	
	fmt.Println("Coverage report generated: coverage.html")
	return nil
}

// (Test) Verbose runs tests with verbose output
func (Test) Verbose() error {
	fmt.Println("Running tests with verbose output...")
	return sh.Run("go", "test", "-v", "./...")
}

// (Test) Race runs tests with race detection
func (Test) Race() error {
	fmt.Println("Running tests with race detection...")
	return sh.Run("go", "test", "-race", "./...")
}

// (Test) Bench runs benchmarks
func (Test) Bench() error {
	fmt.Println("Running benchmarks...")
	return sh.Run("go", "test", "-bench=.", "./...")
}

// Check runs all quality checks
func Check() error {
	mg.Deps(Fmt, Vet, Test.All)
	return nil
}

// Dev runs the development workflow
func Dev() error {
	mg.Deps(Clean, Fmt, Vet, Test.All)
	return Build()
}

// CI runs the CI workflow
func CI() error {
	mg.Deps(Fmt, Vet, Test.Coverage, Test.Race)
	return nil
}