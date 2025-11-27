package main

import (
	"fmt"
	"os"

	"strconv"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)



// Coverage namespace for coverage-related targets
type Coverage mg.Namespace

// All runs tests with coverage for selected packages
func (Coverage) All() error {
	fmt.Println("Running tests with coverage for selected packages...")

	// Create coverage directory if it doesn't exist
	if err := os.MkdirAll("coverage", 0755); err != nil {
		return err
	}

	// Define packages to include in coverage. Exclude generated code and unnecessary packages.
	coverpkg := "github.com/denkhaus/knot/v2/internal/app," +
		"github.com/denkhaus/knot/v2/internal/commands/completion," +
		"github.com/denkhaus/knot/v2/internal/commands/config," +
		"github.com/denkhaus/knot/v2/internal/commands/dependency," +
		"github.com/denkhaus/knot/v2/internal/commands/health," +
		"github.com/denkhaus/knot/v2/internal/commands/project," +
		"github.com/denkhaus/knot/v2/internal/commands/task," +
		"github.com/denkhaus/knot/v2/internal/commands/template," +
		"github.com/denkhaus/knot/v2/internal/commands/validation," +
		"github.com/denkhaus/knot/v2/internal/config," +
		"github.com/denkhaus/knot/v2/internal/errors," +
		"github.com/denkhaus/knot/v2/internal/logger," +
		"github.com/denkhaus/knot/v2/internal/manager," +
		"github.com/denkhaus/knot/v2/internal/repository/inmemory," +
		"github.com/denkhaus/knot/v2/internal/repository/sqlite," +
		"github.com/denkhaus/knot/v2/internal/selection," +
		"github.com/denkhaus/knot/v2/internal/shared," +
		"github.com/denkhaus/knot/v2/internal/treeformatter," +
		"github.com/denkhaus/knot/v2/internal/types," +
		"github.com/denkhaus/knot/v2/internal/validation"

	if err := sh.Run("go", "test", "-coverprofile="+coverageFile, "-covermode=set", "-coverpkg="+coverpkg, "./..."); err != nil {
		fmt.Printf("Error running tests: %v\n", err)
		return err
	}

	if err := sh.Run("go", "tool", "cover", "-html="+coverageFile, "-o", coverageHTML); err != nil {
		return err
	}

	fmt.Println("Coverage report generated: " + coverageHTML)

	// Generate coverage badge
	c := Coverage{}
	if err := c.Badge(); err != nil {
		fmt.Printf("Warning: Failed to generate coverage badge: %v\n", err)
	}

	return nil
}

// Badge generates coverage badge and updates README
func (Coverage) Badge() error {
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
	c := Coverage{}
	if err := c.UpdateReadmeBadge(coveragePercent); err != nil {
		fmt.Printf("Warning: Failed to update README badge: %v\n", err)
	}

	fmt.Printf("Coverage badge generated: %s\n", badgeText)
	return nil
}

// UpdateReadmeBadge updates the coverage badge in README.md
func (Coverage) UpdateReadmeBadge(coveragePercent string) error {
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


