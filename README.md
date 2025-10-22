# Knot

A standalone CLI tool for hierarchical project and task management with dependencies. Perfect for LLM agents to organize complex workflows.

## Features

### Core Management

- **Project Management**: Create, list, update, and delete projects with full lifecycle support
- **Hierarchical Task Management**: Create tasks with parent-child relationships and unlimited depth
- **Task Dependencies**: Manage complex task dependencies and blocking relationships with cycle detection
- **Smart Complexity Management**: Auto-reduce parent task complexity when subtasks are added
- **Local SQLite Storage**: Automatic .knot directory with persistent SQLite database and connection pooling
- **Clean CLI Output**: No JSON logs in normal operation, debug mode available
- **LLM-Friendly**: Structured, parsable outputs perfect for AI agents
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
# Create a project
knot project create --name "My Project" --description "Project description"

# Create a task
knot --project-id <project-uuid> task create --title "Implement feature" --complexity 5

# List tasks
knot --project-id <project-uuid> task list

# Find ready work
knot --project-id <project-uuid> ready

# Check what's blocked
knot --project-id <project-uuid> blocked

# Find next actionable task
knot --project-id <project-uuid> actionable
```

## Core Commands

### Project Management

```bash
# Create project
knot project create --name "Web App" --description "Main web application"

# List all projects
knot project list

# Update project
knot project update --id <project-uuid> --name "Updated Name"

# Delete project
knot project delete --id <project-uuid>
```

### Task Management

```bash
# Create root task
knot --project-id <project-uuid> task create --title "Feature X" --complexity 7

# Create subtask
knot --project-id <project-uuid> task create --title "Subtask" --parent-id <parent-task-uuid>

# Update task state
knot --project-id <project-uuid> task update-state --id <task-uuid> --state in-progress

# Update task details
knot --project-id <project-uuid> task update-title --id <task-uuid> --title "New Title"
knot --project-id <project-uuid> task update-description --id <task-uuid> --description "New desc"
knot --project-id <project-uuid> task update-priority --id <task-uuid> --priority high

# List with filtering
knot --project-id <project-uuid> task list --state pending --complexity-min 5 --search "feature"

# Delete task (two-step process)
knot --project-id <project-uuid> task delete --id <task-uuid>  # Mark for deletion
knot --project-id <project-uuid> task delete --id <task-uuid>  # Confirm deletion

# Delete task with all children
knot --project-id <project-uuid> task delete-subtree --id <task-uuid>
```

### Hierarchy Navigation

```bash
# Get task children
knot --project-id <project-uuid> task children --task-id <task-uuid>

# Get all descendants recursively
knot --project-id <project-uuid> task children --task-id <task-uuid> --recursive

# Get parent task
knot --project-id <project-uuid> task parent --task-id <task-uuid>

# Get root tasks
knot --project-id <project-uuid> task roots

# Show task tree
knot --project-id <project-uuid> task tree --max-depth 3
```

### Dependency Management

```bash
# Add dependency
knot --project-id <project-uuid> dependency add --task-id <task-uuid> --depends-on <other-task-uuid>

# Remove dependency
knot --project-id <project-uuid> dependency remove --task-id <task-uuid> --depends-on <other-task-uuid>

# List dependencies
knot --project-id <project-uuid> dependency list --task-id <task-uuid>

# Show dependency chain
knot --project-id <project-uuid> dependency chain --task-id <task-uuid> --upstream --downstream

# Find dependent tasks
knot --project-id <project-uuid> dependency dependents --task-id <task-uuid> --recursive

# Detect circular dependencies
knot --project-id <project-uuid> dependency cycles

# Validate all dependencies
knot --project-id <project-uuid> dependency validate
```

### Workflow Analysis

```bash
# Find ready tasks (no blockers)
knot --project-id <project-uuid> ready --limit 5

# Find blocked tasks
knot --project-id <project-uuid> blocked --limit 10

# Get next actionable task
knot --project-id <project-uuid> actionable

# Find tasks needing breakdown
knot --project-id <project-uuid> breakdown --threshold 8
```

### Bulk Operations

```bash
# Bulk update tasks
knot --project-id <project-uuid> task bulk-update --task-ids "<task-uuid-1>,<task-uuid-2>,<task-uuid-3>" --state completed

# Bulk create from JSON
knot --project-id <project-uuid> task bulk-create --file tasks.json

# Bulk delete with confirmation
knot --project-id <project-uuid> task bulk-delete --task-ids "<task-uuid-1>,<task-uuid-2>,<task-uuid-3>" --dry-run
knot --project-id <project-uuid> task bulk-delete --task-ids "<task-uuid-1>,<task-uuid-2>,<task-uuid-3>" --force

# Duplicate task to another project
knot --project-id <project-uuid> task duplicate --task-id <task-uuid> --target-project-id <target-project-uuid>

# List tasks by state
knot --project-id <project-uuid> task list-by-state --state pending --json
```

### Template Management

```bash
# List available templates
knot template list

# Create project from template
knot --project-id <project-uuid> template apply --template-name "feature-development"

# Create custom template
knot template create --name "My Template" --file template.yaml

# Update template
knot template update --id <template-id> --file updated-template.yaml

# Delete template
knot template delete --id <template-id>

# Seed built-in templates
knot template seed
```

### Configuration

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

# Validate task states
knot --project-id <project-uuid> validate states

# Validate task hierarchy
knot --project-id <project-uuid> validate hierarchy

# Check database integrity
knot health integrity

# Performance check
knot health performance
```

## Advanced Usage

### JSON Output

Most commands support `--json` flag for machine-readable output:

```bash
knot --project-id <project-uuid> task list --json
knot --project-id <project-uuid> ready --json
knot project list --json
```

### Actor Tracking

Track who makes changes using the `--actor` flag:

```bash
knot --actor "john.doe" --project-id <project-uuid> task create --title "New task"
knot --actor "jane.smith" --project-id <project-uuid> task update-state --id <task-uuid> --state completed
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
# Find high-priority pending tasks with complexity 5-8
knot --project-id <project-uuid> task list \
  --state pending \
  --priority high \
  --complexity-min 5 \
  --complexity-max 8 \
  --sort complexity \
  --reverse

# Search for tasks containing "api" in title or description
knot --project-id <project-uuid> task list --search "api" --limit 10
```

### Template Variables Example

```yaml
# feature-template.yaml
name: "Feature Development"
variables:
  - name: "feature_name"
    type: "string"
    required: true
  - name: "complexity_level"
    type: "choice"
    options: ["Simple", "Medium", "Complex"]

tasks:
  - title: "Design {{feature_name}}"
    description: "Design the {{feature_name}} feature"
    complexity: 5
  - title: "Implement {{feature_name}}"
    description: "Implement {{feature_name}} with {{complexity_level}} complexity"
    complexity: 7
    dependencies: ["design"]
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

## Database

Knot uses SQLite for local storage with automatic database creation in the `.knot` directory. The database includes:

- Connection pooling for performance
- Automatic migrations
- Data integrity constraints
- Transaction support for bulk operations

## Error Handling

Knot provides enhanced error messages with:

- Clear problem descriptions
- Actionable suggestions
- Example commands
- Help command references

## Examples

### Complete Feature Development Workflow

```bash
# Example with environment variables for convenience
PROJECT_ID=$(knot project create --name "Web App" --json | jq -r '.id')

# Apply feature template
knot --project-id $PROJECT_ID template apply --template-name "feature-development"

# Find ready work
knot --project-id $PROJECT_ID ready

# Start working on first task
TASK_ID=$(knot --project-id $PROJECT_ID ready --json | jq -r '.[0].id')
knot --project-id $PROJECT_ID task update-state --id $TASK_ID --state in-progress

# Complete task and find next
knot --project-id $PROJECT_ID task update-state --id $TASK_ID --state completed
knot --project-id $PROJECT_ID actionable
```

### Bug Fix Workflow

```bash
# Create bug fix from template
knot --project-id <project-uuid> template apply --template-name "bug-fix" \
  --var bug_id="BUG-123" \
  --var bug_description="Login form validation error" \
  --var priority="High"

# Track progress
knot --project-id <project-uuid> blocked
knot --project-id <project-uuid> ready
```

### Dependency Management

```bash
# Create tasks with dependencies
DESIGN_ID=$(knot --project-id <project-uuid> task create --title "Design API" --json | jq -r '.id')
IMPL_ID=$(knot --project-id <project-uuid> task create --title "Implement API" --json | jq -r '.id')
TEST_ID=$(knot --project-id <project-uuid> task create --title "Test API" --json | jq -r '.id')

# Set up dependency chain
knot --project-id <project-uuid> dependency add --task-id $IMPL_ID --depends-on $DESIGN_ID
knot --project-id <project-uuid> dependency add --task-id $TEST_ID --depends-on $IMPL_ID

# Validate dependency chain
knot --project-id <project-uuid> dependency chain --task-id $TEST_ID --upstream
knot --project-id <project-uuid> dependency cycles
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.