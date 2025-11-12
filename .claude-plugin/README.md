# KNOT Task Management Claude Plugin

This is the official Claude plugin for the KNOT Task Management system, providing systematic guidance for project and task management workflows.

## Overview

The KNOT plugin enables Claude to provide comprehensive task management guidance using the KNOT CLI tool. It ensures systematic project organization, dependency management, and workflow optimization for development teams.
KNOT provides sophisticated task/todo management persisted between sessions and recallable longterm memory as your project grows.

## Installation

1. Install the KNOT CLI tool:

```bash
go install github.com/denkhaus/knot/cmd/knot@latest
```

2. The plugin will be automatically available when this repository is used as a Claude skill.

## Features

- **Project Management**: Create and manage projects with hierarchical task structures
- **Task Dependencies**: Build complex dependency chains with cycle prevention
- **Workflow Optimization**: Identify actionable tasks and bottlenecks
- **Template System**: Use pre-built templates for common workflows
- **Bulk Operations**: Efficiently manage multiple tasks simultaneously
- **LLM-Friendly**: Structured outputs perfect for AI agents

## Usage

Once installed, you can use the skill by asking Claude to help with task management:

- "Set up a new project with KNOT for my web application"
- "Create a hierarchical task structure for implementing authentication"
- "Help me identify the next actionable tasks"
- "Set up dependencies between my frontend and backend tasks"

## Documentation

See [knot-task-management/SKILL.md](./skills/knot-task-management/SKILL.md) for comprehensive usage documentation.

## Development

This plugin follows the standard Claude marketplace structure:

```ini
.claude-plugin/
├── marketplace.json      # Plugin metadata and marketplace configuration
├── README.md            # This file
└── skills/
    └── knot-task-management/
        └── SKILL.md     # Skill definition and documentation
```

## Contributing

To contribute to this plugin:

1. Fork the main KNOT repository
2. Make changes to the skill documentation
3. Update marketplace.json if needed
4. Submit a pull request

## License

MIT License - see the main repository LICENSE file for details.