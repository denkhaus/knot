---
description: Comprehensive task management workflow using KNOT CLI tool. This skill provides systematic guidance for project and task management, ensuring every piece of work is documented, tracked, and completed with proper dependencies. Use when managing any complex project that requires structured task tracking, dependency management, and systematic workflow organization.
name: knot-task-management
---

# KNOT Task Management Skill

This skill provides comprehensive guidance for systematic project and task management using the KNOT CLI tool. KNOT serves as the single source of truth for all project management activities.

## Installation

Before using KNOT for task management, you need to install the tool:

```bash
go install github.com/denkhaus/knot/cmd/knot@latest
```

This installs the KNOT CLI tool globally on your system. After installation, you can verify it's working with:

```bash
knot --help
```

## Database and Version Control

### KNOT Database Location
- KNOT stores all project and task data in a `.knot` directory in your workspace
- This directory contains the complete project management memory and history
- **IMPORTANT**: Always commit the `.knot` directory to Git to preserve project management knowledge

### Why Version Control .knot Directory
- **Project Memory**: Maintains complete history of all tasks, decisions, and progress
- **Next Steps**: Documents what needs to be done next when returning to a project
- **Roadmap**: Provides historical context and project evolution
- **Team Collaboration**: Enables team members to understand project history and current state
- **Continuity**: Prevents loss of project management knowledge across time and team changes

### Git Configuration
Add the `.knot` directory to your Git repository:

```bash
# Add .knot directory to Git
git add .knot/

# Commit the project management data
git commit -m "Add KNOT project management database"

# Push to preserve project memory
git push
```

### Circular Dependency Prevention
KNOT automatically prevents circular dependencies:
- The tool detects and blocks attempts to create circular dependency chains
- This ensures tasks can always be completed in a logical order
- No manual dependency cycle checking is required

## Core Principles

### 1. Document Everything Before Starting

- **NEVER** begin work without creating a task first
- **ALWAYS** document work as tasks in KNOT before implementation
- **EVERY** decision, bug fix, and feature must have a corresponding task

### 2. Maintain Real-Time State Updates

- Update task to `in-progress` immediately when starting work
- Update task to `completed` immediately when finishing work
- **NEVER** let task states become stale

### 3. Break Down Complex Work

- Tasks with complexity ≥8 **MUST** be broken down into subtasks
- Use `knot breakdown` to identify tasks needing breakdown
- Create hierarchical task structures for complex projects

### 4. Dependency Management

- **ALWAYS** set dependencies between related tasks
- Use `knot dependency add` to link tasks that depend on each other
- Dependencies ensure proper task execution order

## Essential Workflow Commands

### Project Setup

```bash
# Create new project
knot project create --title "Project Name" --description "Detailed project description"

# Select active project (REQUIRED before task operations)
knot project select --id <project-id>

# Verify selected project
knot project get-selected
```

### Task Creation and Management

```bash
# Create new task
knot task create --title "Task Title" --description "Detailed task description" --complexity 5

# Create subtask under parent task
knot task create --parent-id <parent-task-id> --title "Subtask Title" --complexity 3

# Update task state
knot task update-state --id <task-id> --state in-progress
knot task update-state --id <task-id> --state completed

# List tasks with hierarchy
knot task list --depth-max 3
```

### Dependency and Workflow Management

```bash
# Add dependency (Task A depends on Task B)
knot dependency add --task-id <task-a-id> --depends-on <task-b-id>

# Find next actionable tasks
knot actionable

# Check tasks needing breakdown
knot breakdown

# See blocked tasks
knot blocked

# Check ready tasks
knot ready
```

## Task Complexity Guidelines

### Complexity Levels

- **1-3**: Simple tasks (quick fixes, small documentation updates)
- **4-6**: Standard tasks (feature implementation, testing)
- **7-8**: Complex tasks (require planning, multiple components)
- **9-10**: Very complex tasks (MUST be broken down immediately)

### Breakdown Rules

- Tasks ≥8 complexity **MUST** be broken down before starting
- Use subtasks to create manageable work units
- Maintain logical dependencies between subtasks

## Systematic Work Workflow

### Before Any Work

1. **Check current project**: `knot project get-selected`
2. **Create task** if not exists: `knot task create`
3. **Set dependencies** if needed: `knot dependency add`
4. **Check for breakdown**: `knot breakdown`
5. **Break down complex tasks** if required

### Starting Work

1. **Find next task**: `knot actionable`
2. **Update state**: `knot task update-state --id <task-id> --state in-progress`
3. **Work on task systematically**

### During Work

1. **Watch for new work items** that emerge
2. **Create new tasks immediately** for discovered work
3. **Set dependencies** between new and existing tasks
4. **Continue with current task**

### Completing Work

1. **Verify task completion** fully
2. **Update state**: `knot task update-state --id <task-id> --state completed`
3. **Check next actionable**: `knot actionable`
4. **Continue systematic workflow**

## Dynamic Task Creation Patterns

### When to Create New Tasks

**IMMEDIATELY** create new tasks when:

- New requirements emerge during implementation
- Bugs are discovered that need fixing
- Architecture decisions are needed
- Testing requirements are identified
- Documentation updates become necessary
- Performance optimizations are discovered
- Security issues are found
- Refactoring becomes necessary

### Task Creation Examples

```bash
# While implementing Feature X, discover need for Y
knot task create --title "Implement requirement Y" --description "Discovered during Feature X implementation" --complexity 5

# Create dependency relationship
knot dependency add --task-id <feature-y-id> --depends-on <feature-x-id>
```

## Multi-Level Task Hierarchies

### Project Structure Example

```ini
Project: Web Application Development
├── Task: Database Setup (Complexity: 6)
├── Task: Authentication System (Complexity: 9)
│   ├── Subtask: User Model Design (Complexity: 4)
│   ├── Subtask: JWT Implementation (Complexity: 6)
│   │   ├── Sub-Subtask: Token Generation Logic (Complexity: 3)
│   │   ├── Sub-Subtask: Token Validation Middleware (Complexity: 4)
│   │   └── Sub-Subtask: Token Refresh Mechanism (Complexity: 5)
│   └── Subtask: Authentication Endpoints (Complexity: 8)
│       ├── Sub-Subtask: Registration Endpoint (Complexity: 4)
│       ├── Sub-Subtask: Login Endpoint (Complexity: 5)
│       ├── Sub-Subtask: Logout Endpoint (Complexity: 3)
│       └── Sub-Subtask: Password Reset Flow (Complexity: 7)
│           ├── Sub-Subtask: Reset Request Handler (Complexity: 3)
│           ├── Sub-Subtask: Secure Token Generation (Complexity: 4)
│           ├── Sub-Subtask: Email Service Integration (Complexity: 5)
│           └── Sub-Subtask: Password Update Logic (Complexity: 4)
│
├── Task: API Development (Complexity: 8)
│   ├── Subtask: User CRUD Operations (Complexity: 5)
│   │   ├── Sub-Subtask: Create User Endpoint (Complexity: 3)
│   │   ├── Sub-Subtask: Read User Endpoint (Complexity: 2)
│   │   ├── Sub-Subtask: Update User Endpoint (Complexity: 4)
│   │   └── Sub-Subtask: Delete User Endpoint (Complexity: 3)
│   └── Subtask: Data Validation (Complexity: 4)
│       ├── Sub-Subtask: Input Sanitization (Complexity: 3)
│       └── Sub-Subtask: Validation Rules Engine (Complexity: 4)
└── Task: Frontend Development (Complexity: 7)
    ├── Subtask: React Components (Complexity: 6)
    │   ├── Sub-Subtask: User Authentication Components (Complexity: 4)
    │   ├── Sub-Subtask: Dashboard Components (Complexity: 5)
    │   └── Sub-Subtask: Form Components (Complexity: 3)
    └── Subtask: State Management (Complexity: 5)
        ├── Sub-Subtask: Redux Store Setup (Complexity: 3)
        └── Sub-Subtask: API Integration Layer (Complexity: 4)
```

### Multi-Level Hierarchy Benefits

**Multi Level Depth**: KNOT supports multiple levels of task hierarchy, allowing you to break down work across several levels as needed.

**Progressive Refinement**: Start with high-level tasks and progressively break them down into more detailed subtasks as you understand the requirements better.

**Flexible Granularity**: Different parts of your project can have different levels of detail based on their complexity and current understanding.

**Clear Responsibility**: Each level can represent different scopes - epics, features, user stories, technical tasks, or implementation details.

### Hierarchical Workflow

1. **Start with high-level planning tasks**
2. **Break down complex tasks into subtasks**
3. **Set dependencies between levels**
4. **Work through actionable tasks systematically**
5. **Use `knot actionable` to find next ready task**

## Dependency Management Best Practices

### Dependency Types

- **Sequential**: Task B requires Task A completion
- **Prerequisite**: Task B needs Task A output/resources
- **Blocking**: Task A blocks Task B progress

### Dependency Guidelines

- **Minimize circular dependencies**
- **Create clear dependency chains**
- **Document dependency reasons in task descriptions**
- **Review dependencies regularly**

## Templates for Common Workflows

### Feature Development Template

```bash
knot task create --title "Research: Feature X" --complexity 3
knot task create --title "Design: Feature X" --complexity 4
knot task create --title "Implementation: Feature X" --complexity 7
knot task create --title "Testing: Feature X" --complexity 5
knot task create --title "Documentation: Feature X" --complexity 3

# Set dependencies
knot dependency add --task-id <design-id> --depends-on <research-id>
knot dependency add --task-id <implementation-id> --depends-on <design-id>
knot dependency add --task-id <testing-id> --depends-on <implementation-id>
knot dependency add --task-id <documentation-id> --depends-on <testing-id>
```

### Bug Fix Template

```bash
knot task create --title "Investigation: Bug Description" --complexity 4
knot task create --title "Root Cause Analysis" --complexity 5
knot task create --title "Fix Implementation" --complexity 6
knot task create --title "Testing & Verification" --complexity 4

# Set dependencies
knot dependency add --task-id <root-cause-id> --depends-on <investigation-id>
knot dependency add --task-id <fix-id> --depends-on <root-cause-id>
knot dependency add --task-id <testing-id> --depends-on <fix-id>
```

## Continuous Workflow Management

### Daily Workflow Routine

1. **Morning check**: `knot actionable` to see ready tasks
2. **Work systematically** through actionable tasks
3. **Create new tasks** as work emerges
4. **Update states** immediately when starting/finishing
5. **Evening review**: Check project progress and tomorrow's tasks

### Project Health Monitoring

```bash
# Check project progress
knot project get --id <project-id>

# Review blocked tasks
knot blocked

# Identify tasks needing breakdown
knot breakdown

# Review overall task structure
knot task list --depth-max 3
```

## Critical Rules Summary

### MUST DO

- [ ] **Always create task before starting work**
- [ ] **Update task state immediately when starting work**
- [ ] **Update task state immediately when completing work**
- [ ] **Break down tasks with complexity ≥8**
- [ ] **Set dependencies between related tasks**
- [ ] **Use KNOT as single source of truth**
- [ ] **Check `knot actionable` to find next task**

### NEVER DO

- [ ] **Start work without creating a task**
- [ ] **Let task states become stale**
- [ ] **Ignore tasks needing breakdown**
- [ ] **Work without setting dependencies**
- [ ] **Keep project information outside KNOT**

## Integration with Development Workflow

### Before Writing Code

1. **Select project**: `knot project get-selected`
2. **Create implementation task**: `knot task create`
3. **Check dependencies**: `knot actionable`
4. **Set task in-progress**: `knot task update-state --state in-progress`

### During Development

1. **Watch for emerging work**
2. **Create tasks for discovered requirements**
3. **Maintain dependency relationships**
4. **Keep task states current**

### After Completion

1. **Verify work completion**
2. **Update task state**: `knot task update-state --state completed`
3. **Check next task**: `knot actionable`
4. **Continue systematic workflow**

## Troubleshooting Common Issues

### No Actionable Tasks

- Run `knot blocked` to see what's blocking progress
- Run `knot breakdown` to check if tasks need breaking down
- Review project structure with `knot task list`

### Complex Task Management

- Use `knot breakdown` regularly
- Create subtasks for complex work
- Set clear dependencies between subtasks

### Dependency Conflicts

- Review dependency chains
- Remove unnecessary dependencies
- Ensure clear task relationships

This skill ensures systematic, traceable, and complete project management using KNOT as the central authority for all work tracking and execution.