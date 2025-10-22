#!/bin/bash

# Coverage Analysis Script for Knot Project
# Generates comprehensive test coverage reports and analysis

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COVERAGE_DIR="$PROJECT_ROOT/coverage"
COVERAGE_FILE="$COVERAGE_DIR/coverage.out"
HTML_REPORT="$COVERAGE_DIR/coverage.html"
SUMMARY_FILE="$COVERAGE_DIR/coverage_summary.txt"

echo "🔍 Knot Project - Test Coverage Analysis"
echo "========================================"

# Create coverage directory
mkdir -p "$COVERAGE_DIR"

# Clean previous coverage data
rm -f "$COVERAGE_FILE" "$HTML_REPORT" "$SUMMARY_FILE"

echo "📊 Running tests with coverage..."

# Run tests with coverage, excluding vendor and generated code
cd "$PROJECT_ROOT"
go test -coverprofile="$COVERAGE_FILE" -covermode=atomic ./... 2>/dev/null || {
    echo "⚠️  Some tests failed, but continuing with coverage analysis..."
    go test -coverprofile="$COVERAGE_FILE" -covermode=atomic ./... 2>&1 | grep -E "(PASS|FAIL|coverage)" > "$COVERAGE_DIR/test_results.log" || true
}

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ Failed to generate coverage file"
    exit 1
fi

echo "📈 Generating coverage reports..."

# Generate HTML report
go tool cover -html="$COVERAGE_FILE" -o "$HTML_REPORT"

# Generate detailed function coverage
go tool cover -func="$COVERAGE_FILE" > "$COVERAGE_DIR/function_coverage.txt"

# Extract overall coverage percentage
TOTAL_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep "total:" | awk '{print $3}')

echo "📋 Coverage Summary"
echo "==================="
echo "Overall Coverage: $TOTAL_COVERAGE"
echo ""

# Generate detailed analysis
cat > "$SUMMARY_FILE" << EOF
# Knot Project - Test Coverage Analysis Report
Generated: $(date)

## Overall Coverage: $TOTAL_COVERAGE

## Coverage by Package:
EOF

# Add package-by-package breakdown
echo "📦 Package Coverage Breakdown:"
go test -cover ./... 2>/dev/null | grep -E "coverage:" | while read line; do
    package=$(echo "$line" | awk '{print $1}')
    coverage=$(echo "$line" | awk '{print $4}')
    echo "  $package: $coverage"
    echo "$package: $coverage" >> "$SUMMARY_FILE"
done

echo "" >> "$SUMMARY_FILE"
echo "## Detailed Function Coverage:" >> "$SUMMARY_FILE"
cat "$COVERAGE_DIR/function_coverage.txt" >> "$SUMMARY_FILE"

# Identify uncovered areas
echo "" >> "$SUMMARY_FILE"
echo "## Areas Needing Attention (0% coverage):" >> "$SUMMARY_FILE"
grep "0.0%" "$COVERAGE_DIR/function_coverage.txt" | head -20 >> "$SUMMARY_FILE" || echo "No functions with 0% coverage found" >> "$SUMMARY_FILE"

# Identify well-covered areas
echo "" >> "$SUMMARY_FILE"
echo "## Well-Covered Areas (>80% coverage):" >> "$SUMMARY_FILE"
grep -E "(8[0-9]|9[0-9]|100)\..*%" "$COVERAGE_DIR/function_coverage.txt" | head -20 >> "$SUMMARY_FILE" || echo "No functions with >80% coverage found" >> "$SUMMARY_FILE"

echo ""
echo "📊 Coverage Analysis Complete!"
echo "   - Overall Coverage: $TOTAL_COVERAGE"
echo "   - HTML Report: $HTML_REPORT"
echo "   - Summary: $SUMMARY_FILE"
echo "   - Function Details: $COVERAGE_DIR/function_coverage.txt"

# Check if we're meeting the 80% target
COVERAGE_NUM=$(echo "$TOTAL_COVERAGE" | sed 's/%//')
TARGET=80

if (( $(echo "$COVERAGE_NUM >= $TARGET" | bc -l) )); then
    echo "🎯 SUCCESS: Coverage target of $TARGET% achieved!"
else
    NEEDED=$(echo "$TARGET - $COVERAGE_NUM" | bc -l)
    echo "📈 PROGRESS: Need $NEEDED% more coverage to reach $TARGET% target"
fi

echo ""
echo "🔧 Recommendations:"
echo "   1. Focus on packages with 0% coverage"
echo "   2. Add tests for uncovered functions"
echo "   3. Review failing tests and fix state validation issues"
echo "   4. Consider integration tests for CLI commands"

# Generate coverage badge info
BADGE_COLOR="red"
if (( $(echo "$COVERAGE_NUM >= 80" | bc -l) )); then
    BADGE_COLOR="brightgreen"
elif (( $(echo "$COVERAGE_NUM >= 60" | bc -l) )); then
    BADGE_COLOR="yellow"
elif (( $(echo "$COVERAGE_NUM >= 40" | bc -l) )); then
    BADGE_COLOR="orange"
fi

echo "Coverage-$TOTAL_COVERAGE-$BADGE_COLOR" > "$COVERAGE_DIR/badge.txt"

echo ""
echo "📁 All reports saved to: $COVERAGE_DIR/"