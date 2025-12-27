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
knot status actionable # Shows the next ready task

# 5. Work on the task and update progress
knot task update --id <task-id> --state in-progress
knot task update --id <task-id> --state completed
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
knot task update --id <task-id> --state in-progress
```

### Task State Management

Tasks move through these states: `pending` → `in-progress` → `completed` (or `cancelled`/`blocked`)

```
# Set task as in-progress
knot task update --id <task-id> --state in-progress

# Mark task as completed
knot task update --id <task-id> --state completed

# Check tasks that are ready to work on
knot status ready

# See blocked tasks
knot status blocked
```

### Task Dependencies

Dependencies control task execution order:

```
# Add a dependency (task A depends on task B)
knot dependency add --task-id <task-a-id> --depends-on <task-b-id>

# List dependencies for a task
knot dependency list --task-id <task-id>

# Find the next actionable task
knot status actionable
```

### Task Deletion

Tasks can be deleted with a two-step confirmation process:

```
# Delete a single task (two-step confirmation)
knot task delete --id <task-id>          # Mark for deletion
knot task delete --id <task-id>          # Confirm deletion

# Delete task and all children recursively
knot task delete --id <task-id> --all    # Mark entire hierarchy for deletion
knot task delete --id <task-id> --all    # Confirm deletion

# Dry run to see what would be deleted
knot task delete --id <task-id> --dry-run
knot task delete --id <task-id> --all --dry-run

# Cancel deletion (if marked for deletion)
knot task update --id <task-id> --state pending
```

### Project Structure

Projects can have hierarchical tasks. Tasks with complexity ≥ 8 should be broken down:

```
# Create a subtask
knot task create --parent-id <parent-task-id> --title "<subtask-title>"

# Find tasks needing breakdown
knot status breakdown

# List tasks in the selected project
knot task list

# List tasks with filtering options
knot task list --state pending --complexity 5 --limit 20
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
3. Select project → Use `status ready` command to find next task → Work on task → Update state to `in-progress` → Update state to `completed`

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
knot status breakdown  # Will show the authentication endpoints task (complexity 8)

# Break down complex task into subtasks
knot task create --parent-id <auth-task-id> --title "Design JWT token structure" --complexity 4
knot task create --parent-id <auth-task-id> --title "Implement login endpoint" --complexity 5
knot task create --parent-id <auth-task-id> --title "Implement token validation" --complexity 6

# Set dependencies (login needs user model first)
knot dependency add --task-id <login-task-id> --depends-on <user-model-task-id>

# Find your next task
knot status actionable  # Will show "Setup project structure" first

# Work through tasks systematically
knot task update --id <setup-task-id> --state in-progress
knot task update --id <setup-task-id> --state completed

# Check what's next
knot status actionable  # Will show "Implement user model" as it's now ready
```

### Sync with MCP Server

Knot supports intelligent bidirectional synchronization with an MCP (Model Context Protocol) server that uses PostgreSQL for persistent storage.

**Purpose: Multi-Agent Collaboration**

The sync command is especially useful when multiple agents work on projects and tasks simultaneously—for example, when working in separate Git worktrees.

When working simultaneously in Git worktrees, the MCP server serves as the **single source of truth**. Each agent using local mode would only see their own changes to projects and tasks. Therefore, when collaborating on projects across Git worktrees, all agents should access data from the MCP server.

This ensures every agent working on a project has current information about the state of projects, tasks, and subtasks.

```
# Push local changes to MCP server
knot sync --project-id <project-id> --direction push

# Pull changes from MCP server to local
knot sync --project-id <project-id> --direction pull

# Bidirectional sync (merge changes from both sides)
knot sync --project-id <project-id> --direction bi

# Default: bidirectional sync
knot sync --project-id <project-id>
```

**Sync Modes:**
- `push` - Upload local projects, tasks, and dependencies to MCP server
- `pull` - Download projects, tasks, and dependencies from MCP server
- `bi` - Merge changes from both local and MCP (default)

**What Gets Synced:**
- Projects with all metadata
- Tasks with title, description, complexity, priority, state
- Dependencies between tasks
- Parent-child relationships (subtasks)
- All timestamps and IDs preserved

**Common Sync Workflow:**
```
# 1. Create project locally
knot project create --title "Remote Project" --description "Synced with MCP"

# 2. Work on project locally (create tasks, dependencies, etc.)
knot project select --id <project-id>
knot task create --title "Task 1" --complexity 5

# 3. Push to MCP server
knot sync --project-id <project-id> --direction push

# 4. Continue working on either local or MCP side
# 5. Sync bidirectionally to merge changes
knot sync --project-id <project-id> --direction bi
```

### Important Notes

- **Always select a project first** using `knot project select --id <project-id>` before working with tasks
- Use `knot project get-selected` to check which project is currently active
- All task operations work on the currently selected project
- Sync requires MCP server to be running and accessible

For detailed help on any command, use `knot <command> --help`