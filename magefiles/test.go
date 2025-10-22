package main

import (
	"fmt"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Test namespace for test-related targets
type Test mg.Namespace

// All runs all tests
func (Test) All() error {
	fmt.Println("Running all tests...")
	return sh.Run("go", "test", "-v", "./...")
}

// Unit runs unit tests only
func (Test) Unit() error {
	fmt.Println("Running unit tests...")
	return sh.Run("go", "test", "-short", "-v", "./...")
}

// Integration runs integration tests
func (Test) Integration() error {
	fmt.Println("Running integration tests...")
	return sh.Run("go", "test", "-run", "Integration", "-v", "./...")
}

// Coverage runs tests with coverage
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

// Race runs tests with race detection
func (Test) Race() error {
	fmt.Println("Running tests with race detection...")
	return sh.Run("go", "test", "-race", "-v", "./...")
}

// Bench runs benchmarks
func (Test) Bench() error {
	fmt.Println("Running benchmarks...")
	return sh.Run("go", "test", "-bench=.", "-benchmem", "./...")
}

// View opens coverage report
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
