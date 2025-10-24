## Knot CLI - Getting Started for LLM Agents

Knot is a hierarchical project and task management tool with dependencies. It's designed to help structure complex projects into manageable tasks.

### Essential Project Commands

```
# Create a new project
knot project create --name "<project-name>" --description "<project-description>"

# List all projects
knot project list

# Get project details
knot project get --id <project-id>
```

### Project Selection (Required First)

```
# Select a project to work with (required before task operations)
knot project select --id <project-id>

# Check which project is currently selected
knot project get-selected

# Clear project selection
knot project clear-selection
```

### Essential Task Commands

```
# Create a new task (requires project selection first)
knot task create --title "<task-title>" --description "<task-description>" --complexity 5

# List tasks in the selected project
knot task list

# Update a task state
knot task update-state --id <task-id> --state in-progress
```

### Task State Management

Tasks move through these states: `pending` → `in-progress` → `completed` (or `cancelled`/`blocked`)

```
# Set task as in-progress
knot task update-state --id <task-id> --state in-progress

# Mark task as completed
knot task update-state --id <task-id> --state completed

# Check tasks that are ready to work on
knot ready

# See blocked tasks
knot blocked
```

### Task Dependencies

Dependencies control task execution order:

```
# Add a dependency (task A depends on task B)
knot dependency add --task-id <task-a-id> --depends-on <task-b-id>

# List dependencies for a task
knot dependency list --task-id <task-id>

# Find the next actionable task
knot actionable
```

### Project Structure

Projects can have hierarchical tasks. Tasks with complexity ≥ 8 should be broken down:

```
# Create a subtask
knot task create --parent-id <parent-task-id> --title "<subtask-title>"

# Find tasks needing breakdown
knot breakdown

# List tasks with hierarchical view
knot task list --depth-max 3
```

### Templates for Common Patterns

Use templates to create standardized sets of tasks:

```
# List available templates
knot template list

# Apply a template (requires project selection first)
knot template apply --name <template-name>

# Apply with variables
knot template apply --name <template-name> --var name=value
```

### Key Concepts
- **Project**: Container for related tasks
- **Task**: Individual work unit with title, description, complexity (1-10), and state
- **Dependencies**: Ensure tasks are completed in correct order
- **Complexity**: Numerical estimate of effort (1-10); tasks ≥8 should be broken down
- **State**: Tracks task progress (pending, in-progress, completed, blocked, cancelled)

### Common Workflows

1. Create project → Select project → Create tasks → Set dependencies → Work through tasks
2. For complex tasks (complexity ≥8) → Break down into subtasks → Work on subtasks
3. Select project → Use `ready` command to find next task → Work on task → Update state to `in-progress` → Update state to `completed`

### Important Notes

- **Always select a project first** using `knot project select --id <project-id>` before working with tasks
- Use `knot project get-selected` to check which project is currently active
- All task operations work on the currently selected project

For detailed help on any command, use `knot <command> --help`