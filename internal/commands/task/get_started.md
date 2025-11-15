## Knot CLI - Getting Started for LLM Agents

Knot is a hierarchical project and task management tool with dependencies. It's designed to help structure complex projects into manageable tasks.

### 🚀 Quick Start Workflow

```
# 1. Create your first project
knot project create --title "My Project" --description "Project description"

# 2. Select the project (required before any task operations)
knot project select --id <project-id>

# 3. Create your first task
knot task create --title "Initial Setup" --description "Project setup tasks" --complexity 3

# 4. Find out what to work on next
knot actionable        # Shows the next ready task

# 5. Work on the task and update progress
knot task update-state --id <task-id> --state in-progress
knot task update-state --id <task-id> --state completed
```

### 📋 Essential Project Commands

```
# Create a new project
knot project create --title "<project-name>" --description "<project-description>"

# List all projects
knot project list

# Get project details and progress
knot project get --id <project-id>

# Switch between projects
knot project select --id <project-id>
knot project get-selected
```

### Essential Task Commands

```
# Create a new task (requires project selection first)
knot task create --title "<task-title>" --description "<task-description>" --complexity 5

# List tasks in the selected project
knot task list

# Get a specific task by ID
knot task get --id <task-id>

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

### 🔄 Typical LLM Agent Workflow

```
# Start a new coding project
knot project create --title "API Development" --description "REST API with user authentication"
knot project select --id <project-id>

# Create high-level tasks
knot task create --title "Setup project structure" --complexity 3
knot task create --title "Implement user model" --complexity 5
knot task create --title "Create authentication endpoints" --complexity 8

# Check what needs breakdown
knot breakdown  # Will show the authentication endpoints task (complexity 8)

# Break down complex task into subtasks
knot task create --parent-id <auth-task-id> --title "Design JWT token structure" --complexity 4
knot task create --parent-id <auth-task-id> --title "Implement login endpoint" --complexity 5
knot task create --parent-id <auth-task-id> --title "Implement token validation" --complexity 6

# Set dependencies (login needs user model first)
knot dependency add --task-id <login-task-id> --depends-on <user-model-task-id>

# Find your next task
knot actionable  # Will show "Setup project structure" first

# Work through tasks systematically
knot task update-state --id <setup-task-id> --state in-progress
knot task update-state --id <setup-task-id> --state completed

# Check what's next
knot actionable  # Will show "Implement user model" as it's now ready
```

### Important Notes

- **Always select a project first** using `knot project select --id <project-id>` before working with tasks
- Use `knot project get-selected` to check which project is currently active
- All task operations work on the currently selected project

For detailed help on any command, use `knot <command> --help`