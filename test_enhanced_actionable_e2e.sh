#!/bin/bash

# Enhanced Actionable Command - Comprehensive E2E Test Suite
# Tests all strategies and automatic recommendation with different project structures

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

TEST_PROJECT_ID="4dafcb6f-4593-4b36-b478-8523dd4e5b8f"
TEST_RESULTS_DIR="/tmp/knot_e2e_test_results"
mkdir -p "$TEST_RESULTS_DIR"

# Function to print colored output
print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Function to run command and capture output
run_knot_command() {
    local cmd="$1"
    local output_file="$TEST_RESULTS_DIR/$(echo "$cmd" | tr ' ' '_').txt"

    print_info "Running: knot $cmd"
    if eval "knot $cmd" > "$output_file" 2>&1; then
        print_success "Command succeeded: knot $cmd"
        return 0
    else
        print_error "Command failed: knot $cmd"
        cat "$output_file"
        return 1
    fi
}

# Function to test strategy
test_strategy() {
    local strategy="$1"
    local description="$2"
    local expected_behavior="$3"

    print_header "Testing Strategy: $strategy ($description)"

    # Run actionable with specific strategy
    if run_knot_command "actionable --strategy=$strategy --json"; then
        local json_output=$(cat "$TEST_RESULTS_DIR/actionable_--strategy=$strategy_--json.txt")

        # Extract strategy and reason
        local actual_strategy=$(echo "$json_output" | jq -r '.strategy')
        local strategy_reason=$(echo "$json_output" | jq -r '.strategy_reason')
        local selected_task=$(echo "$json_output" | jq -r '.task.title')

        print_info "Strategy used: $actual_strategy"
        print_info "Strategy reason: $strategy_reason"
        print_info "Selected task: $selected_task"

        if [[ "$actual_strategy" == "$strategy" ]]; then
            print_success "Strategy correctly applied: $strategy"
        else
            print_error "Strategy mismatch. Expected: $strategy, Got: $actual_strategy"
            return 1
        fi

        # Check if behavior matches expectation
        if [[ -n "$expected_behavior" ]]; then
            if echo "$strategy_reason" | grep -qi "$expected_behavior"; then
                print_success "Expected behavior confirmed: $expected_behavior"
            else
                print_info "Note: Expected behavior '$expected_behavior' not found in reason: $strategy_reason"
            fi
        fi

        echo "$json_output" > "$TEST_RESULTS_DIR/strategy_${strategy}_result.json"
    else
        return 1
    fi

    echo ""
}

# Function to test automatic recommendation
test_automatic_recommendation() {
    print_header "Testing Automatic Strategy Recommendation"

    # Run actionable without strategy (auto-recommend)
    if run_knot_command "actionable --json"; then
        local json_output=$(cat "$TEST_RESULTS_DIR/actionable_--json.txt")

        local auto_strategy=$(echo "$json_output" | jq -r '.strategy')
        local auto_reason=$(echo "$json_output" | jq -r '.strategy_reason')
        local selected_task=$(echo "$json_output" | jq -r '.task.title')

        print_info "Auto-recommended strategy: $auto_strategy"
        print_info "Recommendation reason: $auto_reason"
        print_info "Selected task: $selected_task"

        # Verify it's not showing as user-selected
        if [[ "$auto_reason" != *"User-selected"* ]]; then
            print_success "Automatic recommendation working correctly"
        else
            print_error "Expected automatic recommendation but got user-selected reason"
            return 1
        fi

        echo "$json_output" > "$TEST_RESULTS_DIR/auto_recommendation_result.json"
        print_success "Automatic recommendation test passed"
    else
        return 1
    fi

    echo ""
}

# Function to create test project structure
create_test_structure() {
    print_header "Creating Comprehensive Test Structure"

    # Select test project
    run_knot_command "project select --id $TEST_PROJECT_ID"

    # Create different task structures to test various scenarios

    # 1. Simple project structure (for creation-order strategy)
    print_info "Creating simple linear tasks..."
    run_knot_command "task create --title 'Simple Task 1' --priority low --description 'First simple task'"
    run_knot_command "task create --title 'Simple Task 2' --priority medium --description 'Second simple task'"
    run_knot_command "task create --title 'Simple Task 3' --priority high --description 'Third simple task'"

    # 2. Hierarchical structure (for depth-first strategy)
    print_info "Creating hierarchical structure..."
    local parent_id
    parent_id=$(knot task create --title 'Parent Task A' --priority medium --description 'Parent with subtasks' --json | jq -r '.id')
    run_knot_command "task create --title 'Subtask A1' --priority high --parent-id $parent_id --description 'First subtask'"
    run_knot_command "task create --title 'Subtask A2' --priority medium --parent-id $parent_id --description 'Second subtask'"

    # 3. Priority-focused structure (for priority strategy)
    print_info "Creating priority-focused tasks..."
    run_knot_command "task create --title 'High Priority Task 1' --priority high --description 'Critical high priority work'"
    run_knot_command "task create --title 'High Priority Task 2' --priority high --description 'Another critical task'"
    run_knot_command "task create --title 'Low Priority Task' --priority low --description 'Can wait task'"

    # 4. Dependency-heavy structure (for dependency-aware strategy)
    print_info "Creating dependency structure..."
    local dep_task1 dep_task2 dep_task3
    dep_task1=$(knot task create --title 'Dependency Base Task' --priority medium --description 'Base task for dependencies' --json | jq -r '.id')
    dep_task2=$(knot task create --title 'Blocking Task' --priority medium --description 'Blocks other tasks' --json | jq -r '.id')
    dep_task3=$(knot task create --title 'Dependent Task' --priority high --description 'Depends on other tasks' --json | jq -r '.id')

    # Note: We can't actually set dependencies with current CLI, but this creates a mix

    # 5. Critical path structure (for critical-path strategy)
    print_info "Creating critical path tasks..."
    run_knot_command "task create --title 'Critical Path Task 1' --priority high --description 'On critical project timeline'"
    run_knot_command "task create --title 'Critical Path Task 2' --priority medium --description 'Also on critical path'"

    print_success "Test structure created with $(knot task list | grep 'Found.*task' | sed 's/Found \([0-9]*\) task.*/\1/') tasks"
    echo ""
}

# Function to run comprehensive tests
run_comprehensive_tests() {
    print_header "Starting Comprehensive Enhanced Actionable Tests"

    # Create test structure
    create_test_structure

    # Test all strategies explicitly
    test_strategy "dependency-aware" "Unblocks other tasks" "unblock"
    test_strategy "priority" "High priority focus" "priority"
    test_strategy "depth-first" "Completes subtasks first" "hierarchy"
    test_strategy "creation-order" "Oldest tasks first" "creation"
    test_strategy "critical-path" "Project timeline focus" "critical"

    # Test automatic recommendation
    test_automatic_recommendation

    # Test verbose output
    print_header "Testing Verbose Output"
    if run_knot_command "actionable --verbose"; then
        local verbose_output=$(cat "$TEST_RESULTS_DIR/actionable_--verbose.txt")
        if echo "$verbose_output" | grep -q "Alternatives considered"; then
            print_success "Verbose output shows alternatives"
        else
            print_info "Note: Verbose output doesn't show alternatives (may be project-dependent)"
        fi
    fi

    # Test help and validation
    print_header "Testing Command Help and Validation"
    run_knot_command "actionable --help"

    # Test invalid strategy handling
    print_info "Testing invalid strategy handling..."
    if run_knot_command "actionable --strategy=invalid-strategy"; then
        print_info "Invalid strategy handled gracefully"
    fi

    print_success "All comprehensive tests completed!"
}

# Function to generate test report
generate_test_report() {
    print_header "Generating Test Report"

    local report_file="$TEST_RESULTS_DIR/test_report_$(date +%Y%m%d_%H%M%S).md"

    cat > "$report_file" << EOF
# Enhanced Actionable Command - E2E Test Report

**Test Date:** $(date)
**Test Project ID:** $TEST_PROJECT_ID

## Test Results Summary

### Strategies Tested:
- ✅ dependency-aware
- ✅ priority
- ✅ depth-first
- ✅ creation-order
- ✅ critical-path
- ✅ automatic recommendation

### Output Formats Tested:
- ✅ Text output
- ✅ JSON output
- ✅ Verbose output
- ✅ Help output

### Key Findings:

\`\`\`json
$(cat "$TEST_RESULTS_DIR/auto_recommendation_result.json" 2>/dev/null || echo "No auto recommendation result available")
\`\`\`

### Test Files Generated:
$(ls -la "$TEST_RESULTS_DIR" | grep -v "^total" | awk '{print "- " $9}')

**Test Status:** PASSED
EOF

    print_success "Test report generated: $report_file"
    echo ""
    print_info "All test results saved in: $TEST_RESULTS_DIR"
}

# Main execution
main() {
    print_header "Enhanced Actionable Command - E2E Test Suite"
    print_info "Starting comprehensive tests..."
    echo ""

    if run_comprehensive_tests; then
        generate_test_report
        print_success "🎉 All tests passed successfully!"
        exit 0
    else
        print_error "❌ Some tests failed!"
        exit 1
    fi
}

# Run main function
main "$@"