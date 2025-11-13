package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

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

	// Create coverage directory if it doesn't exist
	if err := os.MkdirAll("coverage", 0755); err != nil {
		return err
	}

	if err := sh.Run("go", "test", "-coverprofile="+coverageFile, "-covermode=set", "./..."); err != nil {
		fmt.Printf("Error running tests: %v\n", err)
		return err
	}

	if err := sh.Run("go", "tool", "cover", "-html="+coverageFile, "-o", coverageHTML); err != nil {
		return err
	}

	fmt.Println("Coverage report generated: " + coverageHTML)

	// Generate coverage badge
	t := Test{}
	if err := t.Badge(); err != nil {
		fmt.Printf("Warning: Failed to generate coverage badge: %v\n", err)
	}

	return nil
}

// Badge generates coverage badge and updates README
func (Test) Badge() error {
	fmt.Println("Generating coverage badge...")

	// Get coverage percentage
	output, err := sh.Output("go", "tool", "cover", "-func="+coverageFile)
	if err != nil {
		return err
	}

	// Parse coverage percentage from output
	lines := strings.Split(output, "\n")
	var coveragePercent string
	for _, line := range lines {
		if strings.Contains(line, "total:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				coveragePercent = strings.TrimSuffix(parts[2], "%")
				break
			}
		}
	}

	if coveragePercent == "" {
		return fmt.Errorf("could not parse coverage percentage")
	}

	// Determine badge color
	coverage, err := strconv.ParseFloat(coveragePercent, 64)
	if err != nil {
		return err
	}

	var color string
	switch {
	case coverage >= 80:
		color = "green"
	case coverage >= 50:
		color = "yellow"
	default:
		color = "red"
	}

	// Create coverage directory if it doesn't exist
	if err := os.MkdirAll("coverage", 0755); err != nil {
		return err
	}

	// Write badge info to file
	badgeText := fmt.Sprintf("Coverage-%s%%-%s", coveragePercent, color)
	if err := os.WriteFile("coverage/badge.txt", []byte(badgeText), 0644); err != nil {
		return err
	}

	// Update README with new coverage percentage
	t := Test{}
	if err := t.UpdateReadmeBadge(coveragePercent); err != nil {
		fmt.Printf("Warning: Failed to update README badge: %v\n", err)
	}

	fmt.Printf("Coverage badge generated: %s\n", badgeText)
	return nil
}

// UpdateReadmeBadge updates the coverage badge in README.md
func (Test) UpdateReadmeBadge(coveragePercent string) error {
	readmePath := "README.md"

	// Read README content
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}

	// Replace coverage badge line
	lines := strings.Split(string(content), "\n")

	// Determine badge color
	coverage, err := strconv.ParseFloat(coveragePercent, 64)
	if err != nil {
		return err
	}

	var color string
	switch {
	case coverage >= 80:
		color = "green"
	case coverage >= 50:
		color = "yellow"
	default:
		color = "red"
	}

	for i, line := range lines {
		if strings.Contains(line, "[![Coverage](https://img.shields.io/badge/Coverage-") {
			lines[i] = fmt.Sprintf("[![Coverage](https://img.shields.io/badge/Coverage-%s%%25-%s.svg)](./coverage/coverage.html)", coveragePercent, color)
			break
		}
	}

	// Write updated content back
	updatedContent := strings.Join(lines, "\n")
	return os.WriteFile(readmePath, []byte(updatedContent), 0644)
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
