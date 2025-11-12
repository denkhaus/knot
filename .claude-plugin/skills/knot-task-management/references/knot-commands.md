# KNOT CLI Commands Reference

## Project Management Commands

### Create Project
```bash
knot project create --title "Project Name" --description "Project description"
```
- Creates a new project with title and description
- Returns project ID for selection

### Select Project
```bash
knot project select --id <project-id>
```
- Sets the active project for all task operations
- Required before any task management commands

### Get Selected Project
```bash
knot project get-selected
```
- Shows currently selected project details
- Use to verify active project context

### List Projects
```bash
knot project list
```
- Shows all available projects
- Use for project overview and selection

### Get Project Details
```bash
knot project get --id <project-id>
```
- Shows detailed project information and progress
- Use for project status review

## Task Management Commands

### Create Task
```bash
knot task create --title "Task Title" --description "Task description" --complexity 5
```
- Creates a new task in selected project
- Complexity: 1-10 (≥8 requires breakdown)

### Create Subtask
```bash
knot task create --parent-id <parent-task-id> --title "Subtask Title" --complexity 3
```
- Creates a subtask under specified parent task
- Builds hierarchical task structure

### Update Task State
```bash
knot task update-state --id <task-id> --state in-progress|completed|pending|blocked
```
- Updates task progress state
- Critical for workflow management

### List Tasks
```bash
knot task list --depth-max 3
```
- Shows tasks in selected project
- Use depth-max for hierarchical view

### Get Task Details
```bash
knot task get --id <task-id>
```
- Shows detailed task information
- Use for task status review

## Dependency Management Commands

### Add Dependency
```bash
knot dependency add --task-id <task-a-id> --depends-on <task-b-id>
```
- Makes Task A depend on Task B completion
- Essential for workflow sequencing

### List Dependencies
```bash
knot dependency list --task-id <task-id>
```
- Shows all dependencies for specified task
- Use for dependency review

## Workflow Management Commands

### Find Actionable Tasks
```bash
knot actionable
```
- Shows tasks ready to start (no unmet dependencies)
- Primary command for finding next work

### Find Tasks Needing Breakdown
```bash
knot breakdown
```
- Shows tasks with complexity ≥8 that need subtasks
- Use for task complexity management

### Find Blocked Tasks
```bash
knot blocked
```
- Shows tasks blocked by dependencies
- Use for identifying workflow issues

### Find Ready Tasks
```bash
knot ready
```
- Shows tasks ready to work on
- Alternative to `knot actionable`

## Template Commands

### List Templates
```bash
knot template list
```
- Shows available task templates
- Use for standardized task creation

### Apply Template
```bash
knot template apply --name <template-name> --var name=value
```
- Applies template to create standardized task sets
- Use with variables for customization

## Command Options and Flags

### Common Flags
- `--title`: Task/project title (required)
- `--description`: Detailed description (required)
- `--complexity`: 1-10 complexity rating (required for tasks)
- `--id`: Unique identifier for tasks/projects
- `--parent-id`: Parent task ID for subtasks
- `--state`: Task state (pending|in-progress|completed|blocked)
- `--depth-max`: Maximum hierarchy depth for listing

### Output Formats
Most commands support output formatting options:
- Default: Human-readable format
- `--json`: JSON output for scripting
- `--table`: Table format for overview

## Error Handling

### Common Errors
- **No project selected**: Always run `knot project select` first
- **Task not found**: Verify task ID with `knot task list`
- **Dependency cycle**: Check for circular dependencies
- **Invalid complexity**: Use 1-10 range

### Troubleshooting Commands
```bash
# Check current context
knot project get-selected

# Verify task exists
knot task get --id <task-id>

# Check dependency status
knot dependency list --task-id <task-id>
```