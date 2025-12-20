# Knot MCP Server Setup

This guide explains how to set up and test the knot MCP server with Claude Code.

## Quick Start

1. **Start the MCP server:**
   ```bash
   cd docker
   make mcp
   ```
   This starts PostgreSQL and the knot MCP server.

2. **Connect Claude to MCP server:**
   ```bash
   claude mcp connect http://localhost:9094
   ```

3. **Test the setup:**
   ```
   /test-knot-mcp
   ```

## Details

### What the Setup Does

- **PostgreSQL**: Runs on port 5432 with test database
- **Knot MCP Server**: Runs on port 9094 with HTTP transport
- **Docker Network**: Both services communicate via `knot-network`

### Useful Commands

- `make mcp` - Start database and MCP server
- `make logs-mcp` - View MCP server logs
- `make status` - Check service status
- `make down` - Stop all services
- `make clean` - Stop services and remove volumes

### MCP Tools Available

Once connected, Claude can use these knot MCP tools:
- Project management (create, list, select)
- Task management (create, update, delete)
- Status queries (actionable, ready, blocked)
- Dependency management (add, remove, check)

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Claude   │────▶│ MCP Server  │────▶│ PostgreSQL  │
│   Agent    │ HTTP│  (Docker)   │ SQL │  (Docker)   │
└─────────────┘     └─────────────┘     └─────────────┘
     :9094              :internal          :5432
```

## Testing

Use the `/test-knot-mcp` command to run comprehensive tests of the MCP integration. This will:
1. Verify MCP tool connectivity
2. Test project creation and management
3. Test task operations
4. Test status queries
5. Clean up test data (unless disabled)

## Troubleshooting

### Server not responding
```bash
# Check if server is running
curl http://localhost:9094

# View logs
cd docker && make logs-mcp
```

### Database connection issues
```bash
# Test database connection
cd docker && make test
```

### MCP connection issues
```bash
# Check Claude MCP connections
claude mcp list

# Disconnect and reconnect
claude mcp disconnect
claude mcp connect http://localhost:9094
```

## Brain Integration Note

The MCP server can access brain memories with projectId `556a7f4f-e31a-4656-af1a-d1e44da032fa` for enhanced context and knowledge management.