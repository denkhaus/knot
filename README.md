# Knot

[![CI](https://github.com/denkhaus/knot/workflows/CI/badge.svg)](https://github.com/denkhaus/knot/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/Coverage-35.2%25-yellow.svg)](./coverage/coverage.html)
[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/release/denkhaus/knot.svg)](https://github.com/denkhaus/knot/releases/latest)
[![GitHub issues](https://img.shields.io/github/issues/denkhaus/knot.svg)](https://github.com/denkhaus/knot/issues)
[![GitHub stars](https://img.shields.io/github/stars/denkhaus/knot.svg)](https://github.com/denkhaus/knot/stargazers)

A standalone CLI tool for hierarchical project and task management with dependencies. Specifically designed to be the best friend of every LLM agent - with structured, parsable outputs, comprehensive error handling, and an enhanced `get-started` command that provides immediate workflow guidance for AI agents. Perfect for organizing complex workflows and project hierarchies.

## 🤖 LLM Agent First

Knot is specifically designed for AI agents with:

- **Enhanced `get-started` Command**: Comprehensive workflow guidance with emoji indicators and practical examples
- **Structured Outputs**: Machine-readable JSON outputs for all list commands
- **Intelligent Task Discovery**: `actionable`, `ready`, and `blocked` commands for smart workflow management
- **Quick Start Workflow**: 5-step process that gets agents productive immediately
- **Typical LLM Workflow Examples**: Complete API development project walkthrough

## Features

### Core Management

- **Project Management**: Create, list, update, and delete projects with full lifecycle support
- **Project Context System**: Select a project once, then work seamlessly without repeating project IDs
- **Hierarchical Task Management**: Create tasks with parent-child relationships and unlimited depth
- **Task Dependencies**: Manage complex task dependencies and blocking relationships with cycle detection
- **Smart Complexity Management**: Auto-reduce parent task complexity when subtasks are added
- **Local SQLite Storage**: Automatic .knot directory with persistent SQLite database and connection pooling
- **Secure File Permissions**: Database files with 600 permissions, directories with 700 - owner access only
- **Clean CLI Output**: No JSON logs in normal operation, debug mode available
- **LLM-Friendly**: Structured, parsable outputs perfect for AI agents
- **Enhanced LLM Onboarding**: Improved `get-started` command with workflow examples and emoji indicators
- **Help Integration**: The `get-started` command is explicitly mentioned in the main help output (`knot --help`) for easy discovery
- **Audit Trail**: Track all changes with actor information using `--actor` flag

### Advanced Task Operations

- **Bulk Operations**: Update multiple tasks simultaneously, create from JSON, bulk delete with safety checks
- **Task Duplication**: Duplicate tasks between projects with state reset
- **Two-Step Deletion**: Safe task deletion with confirmation and subtree deletion support
- **State Management**: Comprehensive task state transitions with validation
- **Priority Management**: Task prioritization with high/medium/low levels

### Workflow Analysis & Discovery

- **Ready Tasks**: Find tasks with no blockers that are ready to work on
- **Blocked Tasks**: Identify tasks blocked by dependencies with detailed blocking information
- **Actionable Tasks**: Smart recommendation of the next task to work on
- **Breakdown Analysis**: Find high-complexity tasks that need to be broken down into subtasks
- **Dependency Analysis**: Comprehensive dependency chains, cycle detection, and validation
- **Hierarchy Navigation**: Navigate task trees with parent/child/root/descendant commands

### Template System

- **Task Templates**: Pre-built templates for common workflows (bug-fix, feature-development, code-review)
- **Template Variables**: Dynamic template instantiation with variable substitution
- **Custom Templates**: Create and manage custom task templates
- **Template Seeding**: Automatic seeding of built-in templates
- **Conditional Tasks**: Template tasks with conditional inclusion based on variables

### Configuration & Validation

- **Configurable Settings**: Complexity thresholds, hierarchy limits, description length limits
- **State Validation**: Task state transition validation and checks
- **Input Validation**: Comprehensive validation of task titles, descriptions, complexity, and priorities
- **Health Checks**: Database connectivity, integrity checks, and performance monitoring
- **Enhanced Error Handling**: User-friendly error messages with suggestions and examples

### Advanced Filtering & Search

- **Multi-Criteria Filtering**: Filter tasks by state, priority, complexity, depth, and search terms
- **Flexible Sorting**: Sort by title, complexity, state, priority, creation date, or depth
- **JSON Output**: Machine-readable JSON output for all list commands
- **Pagination**: Limit results and paginate through large task lists

## Installation

```bash
go install github.com/denkhaus/knot/cmd/knot@latest
```

Or build locally:

```bash
git clone https://github.com/denkhaus/knot.git
cd knot
go build -o knot cmd/knot/main.go
```

## Quick Start

```bash
# Get comprehensive guidance for LLM agents
knot get-started

# Create a project
knot project create --title "My Project" --description "Project description"

# List available projects to see project IDs
knot project list

# Select the project to work with (do this once)
knot project select --id <project-uuid>

# Now work seamlessly without repeating project ID
knot task create --title "Implement feature" --complexity 5
knot task list
knot ready
knot blocked
knot actionable

# Check which project is currently selected
knot project get-selected

# Switch to a different project when needed
knot project select --id <other-project-uuid>
```

## Core Commands

### Project Management

```bash
# Create project
knot project create --title "Web App" --description "Main web application"

# List all projects
knot project list

# Update project
knot project update --id <project-uuid> --name "Updated Name"

# Delete project
knot project delete --id <project-uuid>
```

### Task Management

```bash
# Select project first (do this once)
knot project select --id <project-uuid>

# Create root task
knot task create --title "Feature X" --complexity 7

# Create subtask
knot task create --title "Subtask" --parent-id <parent-task-uuid>

# Update task state
knot task update-state --id <task-uuid> --state in-progress

# Update task details
knot task update-title --id <task-uuid> --title "New Title"
knot task update-description --id <task-uuid> --description "New desc"
knot task update-priority --id <task-uuid> --priority high

# Get detailed task information
knot task get --id <task-uuid>

# Get task information as JSON
knot task get --id <task-uuid> --json

# List with filtering
knot task list --state pending --complexity-min 5 --search "feature"

# Delete task (two-step process)
knot task delete --id <task-uuid>  # Mark for deletion
knot task delete --id <task-uuid>  # Confirm deletion

# Delete task with all children
knot task delete-subtree --id <task-uuid>
```

### Hierarchy Navigation

```bash
# Get task children
knot task children --task-id <task-uuid>

# Get all descendants recursively
knot task children --task-id <task-uuid> --recursive

# Get parent task
knot task parent --task-id <task-uuid>

# Get root tasks
knot task roots

# Show task tree
knot task tree --max-depth 3
```

### Dependency Management

```bash
# Add dependency
knot dependency add --task-id <task-uuid> --depends-on <other-task-uuid>

# Remove dependency
knot dependency remove --task-id <task-uuid> --depends-on <other-task-uuid>

# List dependencies
knot dependency list --task-id <task-uuid>

# Enhanced dependency visualization with character-based indicators
knot dependency show --project                    # Show project overview
knot dependency show --task-id <id>              # Show specific task dependencies
knot dependency show --tree                     # Show dependency tree structure
knot dependency show --graph                    # Show dependency graph with all connections

# Show dependency chain
knot dependency chain --task-id <task-uuid> --upstream --downstream

# Find dependent tasks
knot dependency dependents --task-id <task-uuid> --recursive

# Detect circular dependencies with enhanced cycle detection
knot dependency cycles

# Validate all dependencies
knot dependency validate
```

### Workflow Analysis

```bash
# Select project first (if not already selected)
knot project select --id <project-uuid>

# Find ready tasks (no blockers)
knot ready --limit 5

# Find blocked tasks
knot blocked --limit 10

# Get next actionable task with intelligent strategy selection
knot actionable                                   # Use auto-recommended strategy
knot actionable --strategy dependency-aware     # Prioritize tasks that unblock others
knot actionable --strategy depth-first          # Complete subtasks before moving to other branches
knot actionable --strategy priority             # Focus on high-priority tasks first
knot actionable --strategy creation-order       # Original knot behavior (oldest first)
knot actionable --strategy critical-path        # Focus on tasks affecting project timeline
knot actionable --verbose                      # Show detailed selection reasoning and alternatives
knot actionable --json                         # Output result as JSON

# Find tasks needing breakdown
knot breakdown --threshold 8
```

### Bulk Operations

```bash
# Select project first (if not already selected)
knot project select --id <project-uuid>

# Bulk update tasks
knot task bulk-update --task-ids "<task-uuid-1>,<task-uuid-2>,<task-uuid-3>" --state completed

# Bulk create from JSON
knot task bulk-create --file tasks.json

# Bulk delete with confirmation
knot task bulk-delete --task-ids "<task-uuid-1>,<task-uuid-2>,<task-uuid-3>" --dry-run
knot task bulk-delete --task-ids "<task-uuid-1>,<task-uuid-2>,<task-uuid-3>" --force

# Duplicate task to another project
knot task duplicate --task-id <task-uuid> --target-project-id <target-project-uuid>

# List tasks by state
knot task list-by-state --state pending --json
```

### Template Management

Templates are reusable task blueprints that help standardize common workflows. Each template is defined in YAML format with the following structure:

```yaml
name: "Template Name"                        # Human-readable name
description: "Template description"         # Detailed description
category: "Category"                        # Category for organization (e.g. "Development")
tags: ["tag1", "tag2"]                     # Tags for searching and filtering
variables:                                 # List of variables for customization
  - name: "variable_name"                  # Variable name
    type: "string"                         # Type: string, int, bool, or choice
    required: true                         # Whether variable is required
    default_value: "default"              # Optional default value
    description: "What this variable is for"
    options: ["option1", "option2"]       # For choice type variables
tasks:                                     # List of tasks to be created
  - id: "unique_task_id"                   # Unique ID within template (for dependencies)
    title: "Task title with {{variable}}"  # Title with variable substitution
    description: "Task description"        # Description with variable substitution
    complexity: 5                          # Task complexity (1-10)
    estimate: 120                          # Time estimate in minutes (optional)
    parent_id: "parent_task_id"           # Parent task ID within template (optional)
    dependencies: ["other_task_id"]       # Dependencies within template (optional)
    metadata:                             # Additional metadata (optional)
      conditional: "{{variable_name}}"    # Only include if variable matches condition
```

**Key Features:**

- __Variable Substitution__: Use `{{variable_name}}` in titles and descriptions
- **Task Dependencies**: Define dependency relationships between tasks
- **Conditional Tasks**: Include tasks based on variable values using metadata
- **Nested Tasks**: Define parent-child relationships within the template
- **Time Estimates**: Plan work with time estimates for each task

**Built-in Templates:**

- `bug-fix`: Complete bug fix workflow with investigation, implementation, and testing
- `feature-development`: Full feature development lifecycle from design to deployment
- `code-review`: Systematic code review process

**Template Commands:**

```bash
# List all available templates
knot template list

# Show detailed information about a template
knot template show --name "feature-development"

# Apply a template to your project (select project first)
knot project select --id <project-uuid>
knot template apply --name "bug-fix" --var bug_id="BUG-123" --var bug_description="Issue description"

# Validate a template file before using it
knot template validate --file my-template.yaml

# Create a custom template
knot template create --file my-template.yaml

# Update template
knot template update --id <template-id> --file updated-template.yaml

# Delete template
knot template delete --id <template-id>

# Seed built-in templates
knot template seed

# Show detailed information about a template including source
knot template info --name "feature-development"

# Edit a user template in your default editor
knot template edit --name "my-template"
```

**Creating Custom Templates:**

1. Create a YAML file with the template definition
2. Save it in `.knot/templates/` directory to make it available as a user template
3. Use the template with the `apply` command

**Variable Types:**

- `string`: Free-form text input
- `int`: Integer numbers
- `bool`: Boolean values (true/false)
- `choice`: Pick from predefined options

```bash
# Show current configuration
knot config show

# Set complexity threshold
knot config set --key complexity-threshold --value 8

# Set maximum hierarchy depth
knot config set --key max-depth --value 5

# Reset to defaults
knot config reset
```

### Health & Validation

```bash
# Check database health
knot health check

# Select project first (if not already selected)
knot project select --id <project-uuid>

# Validate task states
knot validate states

# Validate task hierarchy
knot validate hierarchy

# Check database integrity
knot health integrity

# Performance check
knot health performance
```

## Advanced Usage

### JSON Output

Most commands support `--json` flag for machine-readable output:

```bash
# Select project first (if not already selected)
knot project select --id <project-uuid>

# Get JSON output for tasks and analysis
knot task list --json
knot task get --id <task-uuid> --json
knot ready --json
knot project list --json
```

### Actor Tracking

Track who makes changes using the `--actor` flag:

```bash
# Select project first
knot project select --id <project-uuid>

# Track changes with actor information
knot --actor "john.doe" task create --title "New task"
knot --actor "jane.smith" task update-state --id <task-uuid> --state completed
```

### Environment Variables

```bash
export KNOT_ACTOR="your-name"
export KNOT_DEFAULT_COMPLEXITY=5
export KNOT_COMPLEXITY_THRESHOLD=8
export KNOT_LOG_LEVEL=debug
```

### Complex Filtering

```bash
# Select project first (if not already selected)
knot project select --id <project-uuid>

# Find high-priority pending tasks with complexity 5-8
knot task list \
  --state pending \
  --priority high \
  --complexity-min 5 \
  --complexity-max 8 \
  --sort complexity \
  --reverse

# Search for tasks containing "api" in title or description
knot task list --search "api" --limit 10
```

### Template Variables Example

```yaml
# feature-template.yaml
name: "Feature Development"
description: "Complete workflow for developing new features from design to deployment"
category: "Development"
tags: ["feature", "development", "design", "testing"]
variables:
  - name: "feature_name"
    description: "Name of the feature to be developed"
    type: "string"
    required: true
  - name: "feature_description"
    description: "Detailed description of the feature"
    type: "string"
    required: true
  - name: "complexity_level"
    description: "Overall feature complexity"
    type: "choice"
    required: true
    options: ["Simple", "Medium", "Complex"]
  - name: "include_api"
    description: "Does this feature require API changes?"
    type: "bool"
    required: false
    default_value: "false"
  - name: "include_ui"
    description: "Does this feature require UI changes?"
    type: "bool"
    required: false
    default_value: "true"

tasks:
  - id: "requirements"
    title: "Define Requirements for {{feature_name}}"
    description: "Gather and document detailed requirements for: {{feature_description}}"
    complexity: 4
    estimate: 240  # 4 hours

  - id: "design"
    title: "Design {{feature_name}}"
    description: "Create technical design and architecture for: {{feature_description}}"
    complexity: 5
    dependencies: ["requirements"]
    estimate: 360  # 6 hours

  - id: "api_design"
    title: "API Design for {{feature_name}}"
    description: "Design API endpoints and data models for: {{feature_description}}"
    complexity: 4
    dependencies: ["design"]
    estimate: 180  # 3 hours
    metadata:
      conditional: "{{include_api}}"

  - id: "ui_mockups"
    title: "UI Mockups for {{feature_name}}"
    description: "Create UI mockups and user flow for: {{feature_description}}"
    complexity: 3
    dependencies: ["design"]
    estimate: 240  # 4 hours
    metadata:
      conditional: "{{include_ui}}"

  - id: "backend_implementation"
    title: "Backend Implementation for {{feature_name}}"
    description: "Implement backend logic and data layer for: {{feature_description}}"
    complexity: 6
    dependencies: ["api_design"]
    estimate: 480  # 8 hours

  - id: "frontend_implementation"
    title: "Frontend Implementation for {{feature_name}}"
    description: "Implement user interface for: {{feature_description}}"
    complexity: 5
    dependencies: ["ui_mockups", "backend_implementation"]
    estimate: 360  # 6 hours
    metadata:
      conditional: "{{include_ui}}"

  - id: "unit_tests"
    title: "Unit Tests for {{feature_name}}"
    description: "Write comprehensive unit tests for: {{feature_description}}"
    complexity: 4
    dependencies: ["backend_implementation"]
    estimate: 240  # 4 hours

  - id: "integration_tests"
    title: "Integration Tests for {{feature_name}}"
    description: "Write integration tests for: {{feature_description}}"
    complexity: 5
    dependencies: ["frontend_implementation", "unit_tests"]
    estimate: 300  # 5 hours

  - id: "documentation"
    title: "Documentation for {{feature_name}}"
    description: "Write user and technical documentation for: {{feature_description}}"
    complexity: 3
    dependencies: ["integration_tests"]
    estimate: 180  # 3 hours

  - id: "code_review"
    title: "Code Review for {{feature_name}}"
    description: "Comprehensive code review for: {{feature_description}}"
    complexity: 3
    dependencies: ["documentation"]
    estimate: 120  # 2 hours

  - id: "deployment"
    title: "Deploy {{feature_name}}"
    description: "Deploy feature to production: {{feature_description}}"
    complexity: 3
    dependencies: ["code_review"]
    estimate: 90   # 1.5 hours
```

## Task States

- **pending**: Task is ready to be started
- **in-progress**: Task is currently being worked on
- **completed**: Task has been finished
- **blocked**: Task cannot proceed due to dependencies
- **cancelled**: Task has been cancelled
- **deletion-pending**: Task marked for deletion (two-step deletion)

## Priority Levels

- **low**: Low priority tasks
- **medium**: Medium priority tasks (default)
- **high**: High priority tasks

## Configuration Options

- **complexity-threshold**: Tasks with complexity >= this value need breakdown (default: 8)
- **max-depth**: Maximum hierarchy depth allowed (default: 10)
- **max-tasks-per-depth**: Maximum tasks per hierarchy level (default: 100)
- **max-description-length**: Maximum task description length (default: 1000)
- **auto-reduce-complexity**: Automatically reduce parent complexity when subtasks added (default: true)

## Recent Enhancements

### v2.2+ Major Architectural Improvements

#### **🔧 Enhanced Actionable Command with Intelligent Strategies**
- **5 Selection Strategies**: `dependency-aware`, `depth-first`, `priority`, `creation-order`, `critical-path`
- **Auto-Recommendation**: Intelligent strategy analysis based on project characteristics
- **Performance Caching**: Thread-safe caching layer for dependency graphs and task scores
- **Enhanced Error Context**: Rich error messages with task IDs and recovery suggestions
- **Modular Architecture**: Refactored into focused components following single responsibility principle
- **Detailed Reasoning**: Verbose mode showing selection alternatives and scoring details

#### **📊 Improved Dependency Visualization**
- **Character-Based Display**: Clean text indicators instead of emoji (`[READY]`, `[WORK]`, `[DONE]`, `[BLOCK]`)
- **Enhanced Show Command**: Multiple visualization modes (`--project`, `--tree`, `--graph`, `--blocks`)
- **Task Status First**: Status indicators displayed before task names for immediate recognition
- **Circular Dependency Detection**: Advanced cycle detection with detailed analysis
- **Dependency Tree Views**: Hierarchical tree structures with clear parent-child relationships

#### **⚡ Architecture & Performance Refactoring**
- **Modular Component Design**: Split large analyzer.go (473 lines) into focused files <500 lines
- **Task Map Utilities**: Centralized task lookup eliminating code duplication
- **Performance Caching**: Intelligent caching for dependency graphs and computations
- **Enhanced Error Handling**: Context-rich errors with actionable recovery suggestions

### v2.1+ Improvements

- **🧪 Enhanced Test Coverage**: Increased from 8.6% to 35.2% overall coverage
- **📚 Comprehensive Documentation**: Added Godoc documentation for all public interfaces
- **🔒 Security Improvements**: Implemented secure file permissions (700/600)
- **⚡ Performance Optimizations**: Refactored long methods and improved actionable task logic
- **🎯 LLM-Focused Features**: Enhanced get-started command with practical workflow examples
- **🐛 Critical Bug Fixes**: Resolved database initialization issues and improved hierarchy display

### Test Coverage by Module

- **Repository**: 29.8% → 33.8% (SQLite operations, migrations)
- **Commands**: ~30% → 57.5% (CLI operations, task analysis)
- **Manager**: Improved coverage for business logic and task workflows
- **Types**: Full documentation coverage for core interfaces

## Database

Knot uses SQLite for local storage with automatic database creation in the `.knot` directory. The database includes:

- Connection pooling for performance
- Automatic migrations with secure permission handling
- Data integrity constraints
- Transaction support for bulk operations
- Owner-only file permissions (600) for security

## Error Handling

Knot provides enhanced error messages with:

- Clear problem descriptions
- Actionable suggestions
- Example commands
- Help command references

## Examples

### Complete Feature Development Workflow

```bash
# Create project and select it
PROJECT_ID=$(knot project create --name "Web App" --json | jq -r '.id')
knot project select --id $PROJECT_ID

# Apply feature template
knot template apply --name "feature-development"

# Find ready work
knot ready

# Start working on first task
TASK_ID=$(knot ready --json | jq -r '.[0].id')
knot task update-state --id $TASK_ID --state in-progress

# Complete task and find next
knot task update-state --id $TASK_ID --state completed
knot actionable
```

### Bug Fix Workflow

```bash
# Select project first
knot project select --id <project-uuid>

# Create bug fix from template
knot template apply --name "bug-fix" \
  --var bug_id="BUG-123" \
  --var bug_description="Login form validation error" \
  --var priority="High"

# Track progress
knot blocked
knot ready
```

### Dependency Management

```bash
# Select project first
knot project select --id <project-uuid>

# Create tasks with dependencies
DESIGN_ID=$(knot task create --title "Design API" --json | jq -r '.id')
IMPL_ID=$(knot task create --title "Implement API" --json | jq -r '.id')
TEST_ID=$(knot task create --title "Test API" --json | jq -r '.id')

# Set up dependency chain
knot dependency add --task-id $IMPL_ID --depends-on $DESIGN_ID
knot dependency add --task-id $TEST_ID --depends-on $IMPL_ID

# Validate dependency chain
knot dependency chain --task-id $TEST_ID --upstream
knot dependency cycles
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.