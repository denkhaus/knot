---
name: test-knot-mcp
description: Test knot MCP server functionality through direct tool calls
parameters:
  type: object
  properties:
    cleanup:
      type: boolean
      default: true
      description: Whether to cleanup test data after testing
    project_name:
      type: string
      default: "mcp-test-project"
      description: Name of the test project to create
---
# Knot MCP Server Test Workflow

This command tests the knot MCP server functionality through direct tool calls in a development environment.

## Important Context Understanding

### Workspace Setup
- **Current workspace**: `/home/denkhaus/dev/gomodules/knot` - This is the **knot CLI and MCP server source code**
- **Test project**: Separate project created through MCP tools for testing functionality
- **Key insight**: We're testing the MCP server that we're currently developing in this workspace

### Code Changes Impact
⚠️ **CRITICAL**: When source code changes are made in the knot workspace:
1. The Docker container **must be restarted** for changes to take effect
2. The MCP server inside container needs to be rebuilt with new code
3. All test processes start over with fresh container

### Container Restart Workflow
```bash
# After code changes:
docker compose down
docker compose up -d --build  # Rebuild with latest code
# Then restart entire test workflow from beginning
```

## Prerequisites

### 1. Environment Setup
Start the MCP server with current source code:
```bash
docker compose up -d --build
```

### 2. MCP Connection
Connect Claude to the running MCP server:
```bash
claude mcp add --transport http knot-mcp http://localhost:9094/mcp
```

### 3. Verify Connection
Check that MCP tools are accessible and responsive.

## Complete Test Workflow

### Phase 1: Environment Verification

#### 1.1 Test MCP Connection
Verify basic MCP server connectivity:
- List available MCP tools (should see knot-mcp tools)
- Test basic tool responsiveness with simple calls
- Check server health with proper timeout commands:
  ```bash
  # Health check with timeout (prevents hanging on SSE)
  timeout 5 curl -s http://localhost:9094/health || echo "Health endpoint timeout"

  # Quick MCP endpoint test
  timeout 3 curl -s -o /dev/null -w "%{http_code}" http://localhost:9094/mcp
  ```

#### 1.2 Clean Test State
**IMPORTANT**: Always start with clean test state:
```bash
# List all projects to check for existing test projects
# If test project exists, delete it first to ensure clean state
# This prevents interference from previous test runs
```

### Phase 2: Project Management Testing

#### 2.1 Create Test Project
Create a fresh test project:
- Use specified `project_name` parameter (default: "mcp-test-project")
- Include clear description indicating it's a test project
- Verify project creation succeeded

#### 2.2 Verify Project Operations
Test project-level functionality:
- List all projects to verify creation
- Select the test project as active
- Retrieve project details to confirm data integrity

### Phase 3: Task Management Testing

#### 3.1 Create Task Hierarchy
Create comprehensive test tasks:
- **Main task**: Medium complexity parent task
- **Subtask 1**: Low complexity, depends on main task
- **Subtask 2**: High complexity, depends on main task
- **Independent task**: No dependencies
- **Blocked task**: Depends on subtask to test dependency blocking

#### 3.2 Test Task Lifecycle
Verify complete task management:
- Create tasks with various complexity levels
- Update task details (title, description, priority)
- Change task states (pending → in_progress → completed)
- Test task deletion if supported

#### 3.3 Test Dependency Management
Verify task dependency functionality:
- Create parent-child relationships
- Test dependency blocking
- Verify dependency queries work correctly

### Phase 4: Status and Query Testing

#### 4.1 Test Status Queries
Verify all status endpoints:
- Get actionable tasks (pending + in_progress)
- List ready tasks (pending state only)
- Check for blocked tasks and their blockers
- Get project statistics and health

#### 4.2 Test Data Consistency
Verify data integrity:
- Cross-reference task counts between different endpoints
- Verify dependency relationships are bidirectional
- Check that status updates propagate correctly

### Phase 5: Edge Cases and Error Handling

#### 5.1 Test Edge Cases
Verify robust handling:
- Invalid task IDs
- Circular dependencies (if applicable)
- Invalid state transitions
- Empty project operations

#### 5.2 Test Error Messages
Verify clear error reporting:
- Meaningful error messages for invalid operations
- Proper HTTP status codes
- Consistent error format

## Expected Results

### Success Criteria
- ✅ MCP tools respond correctly and quickly
- ✅ Project operations create and manage projects successfully
- ✅ Task lifecycle management works end-to-end
- ✅ Dependencies block/unblock tasks as expected
- ✅ Status queries return accurate, consistent data
- ✅ Error handling provides clear feedback
- ✅ Database operations maintain integrity

### Data Validation
- All created entities have valid IDs
- Timestamps are properly set
- Relationships are correctly established
- No orphaned records or broken references

## Iterative Testing Workflow

### When Testing Code Changes
1. **Make code changes** in knot workspace
2. **Rebuild and restart** container:
   ```bash
   docker compose down
   docker compose up -d --build
   ```
3. **Run complete test workflow** from Phase 1
4. **Compare results** with previous test run
5. **Document any differences** or regressions

### 🚨 CRITICAL: Bug Issue Creation

**If problems are discovered during testing, immediately create Beads issues with type "bug".**

Add comprehensive description to the beads issues with references in code files and define proper quality management through comprehensive acceptance constraints

This is especially important to:
- Document found problems immediately
- Record reproduction steps
- Make corrections traceable

### Test Data Isolation
Each test run should:
1. Start with clean state (delete existing test project)
2. Create fresh test data
3. Verify all operations work correctly
4. Clean up test data (unless `cleanup: false`)

## Cleanup

### Automatic Cleanup (default)
- Delete test project
- Verify all test data removed
- Confirm clean state for next test run

### Manual Inspection Mode (`cleanup: false`)
- Preserve test project for detailed inspection
- Document state for debugging
- Provide manual cleanup commands

## Troubleshooting

### Common Issues

#### MCP Server Problems
```bash
# Check server status (with timeout to avoid hanging on SSE)
timeout 5 curl -s http://localhost:9094/health || echo "Health check timeout - server may be running but not responding"

# Alternative: Check if MCP endpoint responds
timeout 5 curl -X POST http://localhost:9094/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0"}}}' \
  || echo "MCP endpoint not responding"

# Check container logs
docker compose logs knot-mcp

# Restart with rebuild
docker compose down
docker compose up -d --build
```

#### Database Issues
```bash
# Check database connectivity
docker compose exec knot-mcp sqlite3 /app/data/knot.db ".tables"

# Reset database (if needed)
docker compose down
docker volume rm knot_knot-data
docker compose up -d
```

#### Test State Issues
- Always delete existing test project before starting
- Verify clean state between test runs
- Check for zombie tasks or broken dependencies

### Debug Commands
```bash
# List all MCP tools
# Should show knot-mcp tools in the list

# Check project state
# Should show clean or expected test state

# Verify container status
docker compose ps

# Test server connectivity (with proper timeouts)
timeout 3 curl -s http://localhost:9094/health || echo "Health endpoint timeout"
timeout 3 curl -s -o /dev/null -w "%{http_code}" http://localhost:9094/mcp || echo "MCP endpoint not reachable"

# Real-time logs
docker compose logs -f knot-mcp
```

## Test Report Template

After each test run, document:
```
Test Run: [timestamp]
Code Changes: [yes/no + description]
Container Rebuilt: [yes/no]
Results:
- MCP Connection: ✅/❌
- Project Creation: ✅/❌
- Task Management: ✅/❌
- Dependencies: ✅/❌
- Status Queries: ✅/❌
- Data Integrity: ✅/❌
Issues Found: [description]
Notes: [additional observations]
```

This structured approach ensures consistent, thorough testing of the knot MCP server functionality while maintaining clear separation between the development workspace and test environment.
