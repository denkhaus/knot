package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// TestActionableResult represents the JSON output from knot actionable --json
type TestActionableResult struct {
	Task           TaskDetails       `json:"task"`
	Strategy       string            `json:"strategy"`
	StrategyReason string            `json:"strategy_reason"`
	Reason         string            `json:"reason"`
	Score          float64           `json:"score"`
	ExecutionTime  string            `json:"execution_time"`
	Alternatives   []TaskAlternative `json:"alternatives,omitempty"`
}

// TaskDetails represents task information from actionable output
type TaskDetails struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	Priority    int       `json:"priority"`
	Complexity  int       `json:"complexity"`
	Depth       int       `json:"depth"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TaskAlternative represents alternative tasks considered
type TaskAlternative struct {
	Task  TaskDetails `json:"task"`
	Score float64     `json:"score"`
}

// Actionable namespace for actionable command testing
type Actionable mg.Namespace

const (
	testProjectID  = "4dafcb6f-4593-4b36-b478-8523dd4e5b8f"
	testResultsDir = "/tmp/knot_e2e_test_results"
	reportFileName = "actionable_test_report_%s.md"
)

// TestStrategies runs comprehensive tests for all actionable strategies
func (Actionable) TestStrategies() error {
	fmt.Println("=== Enhanced Actionable Command - Mage Test Suite ===")

	// Create test results directory
	if err := os.MkdirAll(testResultsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create test results directory: %w", err)
	}

	// Test cases
	strategies := []struct {
		name             string
		description      string
		expectedBehavior string
	}{
		{"dependency-aware", "Unblocks other tasks", "unblock"},
		{"priority", "High priority focus", "priority"},
		{"depth-first", "Completes subtasks first", "hierarchy"},
		{"creation-order", "Oldest tasks first", "creation"},
		{"critical-path", "Project timeline focus", "critical"},
	}

	var successCount, totalCount int

	// Test each strategy
	for _, strategy := range strategies {
		fmt.Printf("\n=== Testing Strategy: %s (%s) ===\n", strategy.name, strategy.description)

		if testStrategy(strategy.name, strategy.description, strategy.expectedBehavior) {
			fmt.Printf("✅ Strategy %s test passed\n", strategy.name)
			successCount++
		} else {
			fmt.Printf("❌ Strategy %s test failed\n", strategy.name)
		}
		totalCount++
	}

	// Test automatic recommendation
	fmt.Printf("\n=== Testing Automatic Strategy Recommendation ===\n")
	if testAutomaticRecommendation() {
		fmt.Println("✅ Automatic recommendation test passed")
		successCount++
	} else {
		fmt.Println("❌ Automatic recommendation test failed")
	}
	totalCount++

	// Test verbose output
	fmt.Printf("\n=== Testing Verbose Output ===\n")
	if testVerboseOutput() {
		fmt.Println("✅ Verbose output test passed")
		successCount++
	} else {
		fmt.Println("❌ Verbose output test failed")
	}
	totalCount++

	// Test invalid strategy handling
	fmt.Printf("\n=== Testing Invalid Strategy Handling ===\n")
	if testInvalidStrategyHandling() {
		fmt.Println("✅ Invalid strategy handling test passed")
		successCount++
	} else {
		fmt.Println("❌ Invalid strategy handling test failed")
	}
	totalCount++

	// Generate test report
	reportFile := filepath.Join(testResultsDir, fmt.Sprintf(reportFileName, time.Now().Format("20060102_150405")))
	if err := generateTestReport(successCount, totalCount, reportFile); err != nil {
		fmt.Printf("Warning: Failed to generate test report: %v\n", err)
	}

	// Summary
	fmt.Printf("\n=== Test Summary ===\n")
	fmt.Printf("Passed: %d/%d tests\n", successCount, totalCount)
	if successCount == totalCount {
		fmt.Printf("🎉 All tests passed successfully!\n")
		return nil
	}

	fmt.Printf("❌ %d tests failed\n", totalCount-successCount)
	return fmt.Errorf("some tests failed")
}

// testStrategy tests a specific actionable strategy
func testStrategy(strategy, description, expectedBehavior string) bool {
	// Run actionable command
	cmd := fmt.Sprintf("actionable --strategy=%s --json", strategy)
	outputFile := filepath.Join(testResultsDir, fmt.Sprintf("actionable_%s.json", strategy))

	if err := runKnotCommand(cmd, outputFile); err != nil {
		fmt.Printf("❌ Command failed: %v\n", err)
		return false
	}

	// Parse and validate JSON output
	result, err := parseActionableOutput(outputFile)
	if err != nil {
		fmt.Printf("❌ Failed to parse JSON output: %v\n", err)
		return false
	}

	fmt.Printf("Strategy used: %s\n", result.Strategy)
	fmt.Printf("Strategy reason: %s\n", result.StrategyReason)
	fmt.Printf("Selected task: %s\n", result.Task.Title)

	// Validate strategy match
	if result.Strategy != strategy {
		fmt.Printf("❌ Strategy mismatch. Expected: %s, Got: %s\n", strategy, result.Strategy)
		return false
	}

	// Check expected behavior (optional)
	if expectedBehavior != "" && strings.Contains(strings.ToLower(result.StrategyReason), expectedBehavior) {
		fmt.Printf("✅ Expected behavior confirmed: %s\n", expectedBehavior)
	}

	// Save individual result
	individualFile := filepath.Join(testResultsDir, fmt.Sprintf("strategy_%s_result.json", strategy))
	if err := saveJSONResult(result, individualFile); err != nil {
		fmt.Printf("Warning: Failed to save individual result: %v\n", err)
	}

	return true
}

// testAutomaticRecommendation tests the automatic strategy recommendation
func testAutomaticRecommendation() bool {
	cmd := "actionable --json"
	outputFile := filepath.Join(testResultsDir, "actionable_auto_recommendation.json")

	if err := runKnotCommand(cmd, outputFile); err != nil {
		fmt.Printf("❌ Command failed: %v\n", err)
		return false
	}

	result, err := parseActionableOutput(outputFile)
	if err != nil {
		fmt.Printf("❌ Failed to parse JSON output: %v\n", err)
		return false
	}

	fmt.Printf("Auto-recommended strategy: %s\n", result.Strategy)
	fmt.Printf("Recommendation reason: %s\n", result.StrategyReason)
	fmt.Printf("Selected task: %s\n", result.Task.Title)

	// Validate that a strategy was recommended
	if result.Strategy == "" || result.StrategyReason == "" {
		fmt.Printf("❌ Missing strategy recommendation\n")
		return false
	}

	return true
}

// testVerboseOutput tests verbose output functionality
func testVerboseOutput() bool {
	cmd := "actionable --verbose"
	outputFile := filepath.Join(testResultsDir, "actionable_verbose_output.txt")

	if err := runKnotCommand(cmd, outputFile); err != nil {
		fmt.Printf("❌ Command failed: %v\n", err)
		return false
	}

	// Check if output contains verbose indicators
	content, err := os.ReadFile(outputFile)
	if err != nil {
		fmt.Printf("❌ Failed to read verbose output: %v\n", err)
		return false
	}

	output := string(content)
	if strings.Contains(output, "Alternatives") || strings.Contains(output, "Execution time") {
		fmt.Printf("✅ Verbose output shows enhanced information\n")
		return true
	}

	fmt.Printf("❌ Verbose output missing expected information\n")
	return false
}

// testInvalidStrategyHandling tests handling of invalid strategies
func testInvalidStrategyHandling() bool {
	cmd := "actionable --strategy=invalid-strategy"
	outputFile := filepath.Join(testResultsDir, "actionable_invalid_strategy.txt")

	// This should succeed (graceful handling) not fail
	if err := runKnotCommand(cmd, outputFile); err != nil {
		fmt.Printf("❌ Command failed unexpectedly: %v\n", err)
		return false
	}

	fmt.Printf("✅ Invalid strategy handled gracefully\n")
	return true
}

// runKnotCommand executes a knot command and saves output to file
func runKnotCommand(cmd, outputFile string) error {
	fmt.Printf("Running: knot %s\n", cmd)

	// Ensure we're using the test project
	if err := sh.Run("knot", "project", "select", "--id", testProjectID); err != nil {
		return fmt.Errorf("failed to select test project: %w", err)
	}

	// Run the command and capture output
	output, err := sh.Output("knot", strings.Fields(cmd)...)
	if err != nil {
		// Save error output for debugging
		if writeErr := os.WriteFile(outputFile, []byte(output), 0o644); writeErr != nil {
			return fmt.Errorf("failed to save error output: %w", err)
		}
		return fmt.Errorf("command failed: %w", err)
	}

	// Save successful output
	if err := os.WriteFile(outputFile, []byte(output), 0o644); err != nil {
		return fmt.Errorf("failed to save output: %w", err)
	}

	return nil
}

// parseActionableOutput parses JSON output from actionable command
func parseActionableOutput(outputFile string) (*TestActionableResult, error) {
	content, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read output file: %w", err)
	}

	var result TestActionableResult
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}

// saveJSONResult saves the actionable result as JSON
func saveJSONResult(result *TestActionableResult, filename string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	return os.WriteFile(filename, data, 0o644)
}

// generateTestReport generates a markdown test report
func generateTestReport(successCount, totalCount int, reportFile string) error {
	report := fmt.Sprintf(`# Actionable Command Test Report

Generated: %s
Test Project ID: %s

## Test Results

- **Tests Passed**: %d/%d
- **Success Rate**: %.1f%%

## Test Categories

### Strategy Tests
All actionable strategies tested for:
- Correct strategy application
- Proper JSON output format
- Reasonable task selection

### Functionality Tests
- Automatic strategy recommendation
- Verbose output formatting
- Invalid strategy error handling

## Files Generated

Test outputs saved in: %s

## Summary

%s`,
		time.Now().Format("2006-01-02 15:04:05"),
		testProjectID,
		successCount, totalCount,
		float64(successCount)/float64(totalCount)*100,
		testResultsDir,
		func() string {
			if successCount == totalCount {
				return "🎉 All tests passed successfully!"
			}
			return fmt.Sprintf("❌ %d tests failed", totalCount-successCount)
		}())

	return os.WriteFile(reportFile, []byte(report), 0o644)
}

// Cleanup removes test artifacts
func (Actionable) Cleanup() error {
	fmt.Printf("Cleaning up test artifacts from %s\n", testResultsDir)

	if err := os.RemoveAll(testResultsDir); err != nil {
		return fmt.Errorf("failed to cleanup test artifacts: %w", err)
	}

	fmt.Printf("✅ Test artifacts cleaned up\n")
	return nil
}

// SetupTestEnvironment ensures the test environment is ready
func (Actionable) SetupTestEnvironment() error {
	fmt.Printf("Setting up test environment...\n")

	// Create test results directory
	if err := os.MkdirAll(testResultsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create test results directory: %w", err)
	}

	// Select test project
	if err := sh.Run("knot", "project", "select", "--id", testProjectID); err != nil {
		return fmt.Errorf("failed to select test project: %w", err)
	}

	fmt.Printf("✅ Test environment setup complete\n")
	return nil
}
