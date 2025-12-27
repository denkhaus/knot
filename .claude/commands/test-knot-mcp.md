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

### Phase 3.5: Task Deletion Testing

#### 3.5.1 Test Single Task Deletion
Verify task deletion works correctly:
- Create a standalone task (no dependencies)
- Delete the task using `task_delete`
- Verify task no longer appears in any status queries
- Verify no orphaned dependencies remain

**Expected Results:**
```
Create Task: task_id = "test-single-task"
Delete Task: task_delete(task_id = "test-single-task")
Verify:
  - task_get("test-single-task") returns error/not found
  - status_ready() does not list deleted task
  - status_actionable() does not list deleted task
  - dependency_list() shows no broken references
```

#### 3.5.2 Test Task Deletion with Dependencies
Verify cascading deletion behavior:
- Create a parent task
- Create subtasks that depend on parent
- Delete parent task
- Verify behavior: subtasks are either (a) also deleted OR (b) become unblocked with updated dependencies

**Expected Results:**
```
Setup:
  parent_task = create_task("Parent Task")
  subtask1 = create_task("Subtask 1")
  subtask2 = create_task("Subtask 2")
  dependency_add(subtask1, depends_on: parent_task)
  dependency_add(subtask2, depends_on: parent_task)

Test: task_delete(parent_task)

Verification Steps:
  1. Check if parent_task is deleted
  2. Check subtask1 and subtask2 status:
     - If cascade delete: both subtasks should also be deleted
     - If dependency cleanup: subtasks exist with no parent dependency
  3. Verify no broken dependency references
  4. Verify status_blocked() is accurate
```

#### 3.5.3 Test Task Deletion with Subtasks
Verify deletion of task hierarchy:
- Create parent task with subtasks
- Delete parent task
- Verify all subtasks are handled correctly

**Expected Results:**
```
Setup:
  epic_task = create_task("Epic Task")
  subtask1 = create_task("Subtask 1", parent_id: epic_task.id)
  subtask2 = create_task("Subtask 2", parent_id: epic_task.id)

Test: task_delete(epic_task.id)

Verification:
  - epic_task no longer exists
  - All subtasks (subtask1, subtask2) are also deleted
  - No orphaned tasks in hierarchy
  - status_ready() and status_actionable() show clean state
```

### Phase 4: Project Deletion Testing

#### 4.1 Test Project Deletion with Complete Task Hierarchy
Verify project deletion removes all associated data:
- Create project with tasks, subtasks, and dependencies
- Delete entire project
- Verify no orphaned data remains

**Test Steps:**
```bash
# 1. Create test project
project = project_create(title="Delete Test Project")
project_id = project.id

# 2. Select project
project_select(project_id)

# 3. Create comprehensive task hierarchy
main_task = task_create(title="Main Task", complexity: 5)
subtask1 = task_create(title="Subtask 1", complexity: 3)
subtask2 = task_create(title="Subtask 2", complexity: 7)
nested_subtask = task_create(title="Nested Subtask", parent_id: subtask1.id)
blocked_task = task_create(title="Blocked Task")

# 4. Add dependencies
dependency_add(blocked_task.id, depends_on: main_task.id)
dependency_add(subtask1.id, depends_on: main_task.id)

# 5. Verify project state before deletion
stats_before = status_project()
expected_counts = {
  total_tasks: 5,
  open_tasks: 5,
  dependencies: 2
}

# 6. Record all task IDs for verification
task_ids = [main_task.id, subtask1.id, subtask2.id, nested_subtask.id, blocked_task.id]

# 7. Delete project
project_delete(project_id)

# 8. Verification phase - ALL must pass:
verify_project_deleted()
verify_all_tasks_deleted(task_ids)
verify_no_orphaned_dependencies()
verify_project_list_clean()
```

**Expected Results - Detailed Verification:**

```
## Verification 1: Project Deletion
project_list() should NOT contain the deleted project_id
project_get(project_id) should return error/not found

## Verification 2: All Tasks Deleted
For each task_id in task_ids:
  - task_get(task_id) should return error/not found
  - status_ready() should NOT list any of these tasks
  - status_actionable() should NOT list any of these tasks
  - status_blocked() should NOT list any of these tasks

## Verification 3: No Orphaned Dependencies
dependency_list() queries should return:
  - Empty results for deleted tasks
  - No "depends_on" references to deleted task_ids
  - No broken foreign key references

## Verification 4: Project Statistics Clean
status_project() should show:
  - Total projects decreased by 1
  - Total tasks decreased by 5
  - No negative counts
  - Consistent statistics across all queries

## Verification 5: Clean State
- List all projects - deleted project not in list
- Try to select deleted project - should fail
- Create new project with same name - should succeed (no conflicts)
```

#### 4.2 Test Multiple Project Deletion
Verify deleting multiple projects maintains integrity:
- Create project A with tasks
- Create project B with tasks
- Create project C with tasks
- Delete projects A and B
- Verify only project C remains with all its data intact

**Expected Results:**
```
Setup:
  projects = [project_create("Project A"), project_create("Project B"), project_create("Project C")]
  for each p in projects[:2]:
    select p
    create_tasks(count: 3)

Test: Delete projects A and B

Verification:
  - project_list() shows only Project C
  - Project C's tasks are all intact
  - Projects A and B tasks are completely removed
  - No cross-project data leakage
```

#### 4.3 Test Project Deletion Error Handling
Verify proper error handling for edge cases:
- Try to delete non-existent project
- Try to delete currently selected project
- Try to delete project while tasks are in various states

**Expected Results:**
```
Error Cases:
1. Delete non-existent project:
   - Should return clear error message
   - Should not affect existing projects

2. Delete currently selected project:
   - Should succeed
   - Next operation should handle no-active-project state gracefully

3. Delete project with in-progress tasks:
   - Should succeed (all tasks deleted)
   - No orphaned in-progress state remains
```

### Phase 5: Status and Query Testing

#### 5.1 Test Status Queries
Verify all status endpoints:
- Get actionable tasks (pending + in_progress)
- List ready tasks (pending state only)
- Check for blocked tasks and their blockers
- Get project statistics and health

#### 5.2 Test Data Consistency
Verify data integrity:
- Cross-reference task counts between different endpoints
- Verify dependency relationships are bidirectional
- Check that status updates propagate correctly

#### 5.3 Test Dependency-Aware Actionable Selection with Deep Nesting

This phase verifies the `status_actionable` MCP tool uses the same selection logic as `knot status actionable` CLI command, especially for deeply nested project/task structures with complex dependencies.

**Test Setup: Complex Nested Hierarchy**

Create a comprehensive test structure to verify dependency-aware selection:

```
Project: "Deep Nest Test"

Root Task A (complexity: 5, priority: high)
├── Subtask A.1 (complexity: 3, priority: medium)
│   └── Subtask A.1.a (complexity: 2, priority: low)
├── Subtask A.2 (complexity: 4, priority: high)
│   ├── Subtask A.2.a (complexity: 3, priority: medium)
│   │   └── Subtask A.2.a.i (complexity: 1, priority: low)
│   └── Subtask A.2.b (complexity: 2, priority: medium)
└── Subtask A.3 (complexity: 6, priority: low)

Root Task B (complexity: 8, priority: medium) - [depends on: A.2]
├── Subtask B.1 (complexity: 3, priority: high) - [depends on: A.2.a]
└── Subtask B.2 (complexity: 4, priority: medium)

Root Task C (complexity: 4, priority: high) - [independent]
```

**Test Steps:**

1. **Create the hierarchy using MCP tools:**
   - Create root tasks (A, B, C) with specified properties
   - Create subtasks with proper parent_id relationships
   - Add dependencies between tasks using `dependency_add`

2. **Verify initial actionable selection:**
   - Call `status_actionable` MCP tool
   - Expected: Root Task C (independent, high priority) OR A.1 (deepest pending task with no deps)
   - Verify the reasoning message includes correct factors (priority, depth, unblocking)

3. **Test progression through hierarchy:**
   - Mark selected task as in_progress using `task_update_state`
   - Call `status_actionable` again
   - Verify it selects the next logical task based on:
     - Completed dependencies
     - Depth-first preference (deeper subtasks preferred)
     - Priority weighting
     - Unblocking potential (tasks that unblock dependents)

4. **Test dependency blocking:**
   - Call `status_actionable` before completing A.2
   - Verify B and B.1 are NOT selected (blocked by dependency on A.2/A.2.a)
   - Complete A.2.a then A.2
   - Call `status_actionable` again
   - Verify B or B.1 now becomes actionable (unblocked)

5. **Compare with CLI actionable command:**
   - Run `./knot status actionable` in the project
   - Verify same task is selected as by MCP tool
   - Compare reasoning messages - should mention same factors

**Expected Behaviors for Dependency-Aware Selection:**

1. **Depth-First Selection:** Deeper subtasks preferred to complete branches
   - A.1.a should be selected before A.2 (same depth, but priority may influence)
   - A.2.a.i should be selected before B (when both are ready)

2. **Dependency Unblocking:** Tasks that unblock others are prioritized
   - A.2 has high unblocking potential (unblocks B, B.1, A.2.a, A.2.b)
   - After A.2 is complete, B and B.1 become available

3. **Priority Weighting:** High priority tasks get preference
   - Root Task A (high priority) should be selected over similar complexity medium priority tasks

4. **In-Progress Preference:** Already in-progress tasks get 20% score bonus
   - Once A.1 is in-progress, it should be preferred over other pending tasks

**Verification Points:**

```json
// Example status_actionable response structure:
{
  "project_id": "...",
  "tasks": [
    {
      "id": "task-id",
      "title": "Task Title",
      "state": "pending",
      "priority": "high",
      "complexity": 5
    }
  ],
  "total": 1,
  "message": "selected using dependency-aware strategy | will unblock 3 task(s) | high priority | subtask (completing branch) | on critical path (length 4)"
}
```

**Key Assertions:**
- Only 1 task returned (the next actionable task)
- Message contains reasoning factors from selection logic
- Task state is either "pending" or "in_progress" (never "blocked")
- Task has all dependencies completed
- Deep nesting doesn't break dependency resolution
- Circular dependencies are detected and reported properly

**Error Cases to Test:**

1. **Empty project:** No tasks → returns empty list with "No tasks found" message
2. **All blocked:** All tasks have unmet dependencies → returns "No actionable tasks available"
3. **Circular dependencies:** A→B→A → returns error message about cycle detection
4. **Deadlock:** No tasks can progress → returns detailed deadlock explanation

**Success Criteria:**
- ✅ `status_actionable` returns same task as `knot status actionable`
- ✅ Reasoning message includes dependency-aware factors
- ✅ Deep nesting (4+ levels) doesn't break selection
- ✅ Complex dependency chains are correctly resolved
- ✅ Priority, depth, and unblocking potential influence selection
- ✅ In-progress tasks get preference when configured
- ✅ Error messages are clear and actionable

### Phase 6: Edge Cases and Error Handling

#### 6.1 Test Edge Cases
Verify robust handling:
- Invalid task IDs
- Circular dependencies (if applicable)
- Invalid state transitions
- Empty project operations

#### 6.2 Test Error Messages
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
- ✅ **Dependency-aware actionable selection works correctly with deep nesting**
- ✅ **status_actionable MCP tool returns same results as knot status actionable CLI**
- ✅ **Selection reasoning includes priority, depth, and unblocking factors**
- ✅ **Complex dependency chains are correctly resolved**
- ✅ Error handling provides clear feedback
- ✅ Database operations maintain integrity
- ✅ **Task deletion works correctly for standalone tasks**
- ✅ **Task deletion properly handles dependencies and subtasks**
- ✅ **Project deletion removes all associated tasks and dependencies**
- ✅ **No orphaned data after deletion operations**

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
4. **Test deletion operations on complex hierarchies**
5. Verify complete cleanup (no orphaned data)
6. Clean up test data (unless `cleanup: false`)

## Cleanup

### Automatic Cleanup (default)
- Delete test project (this tests project_delete)
- Verify all test data removed (including tasks, subtasks, dependencies)
- Confirm clean state for next test run
- Run orphaned data check to ensure no remnants remain

### Manual Inspection Mode (`cleanup: false`)
- Preserve test project for detailed inspection
- Document state for debugging
- Provide manual cleanup commands
- Include verification that deletion can be done manually

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
docker compose exec postgres psql -U knot_user -d knot_test -c "\dt"

# Check database tables and records
docker compose exec postgres psql -U knot_user -d knot_test \
  -c "SELECT tablename FROM pg_tables WHERE schemaname = 'public';"

# Check projects table
docker compose exec postgres psql -U knot_user -d knot_test \
  -c "SELECT id, title FROM projects;"

# Check tasks table
docker compose exec postgres psql -U knot_user -d knot_test \
  -c "SELECT id, title, project_id FROM tasks LIMIT 10;"

# Reset database (if needed)
docker compose down -v
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
- Dependency-Aware Actionable Selection: ✅/❌
- Deep Nesting Test (4+ levels): ✅/❌
- MCP vs CLI Selection Consistency: ✅/❌
- Data Integrity: ✅/❌
- Task Deletion: ✅/❌
- Project Deletion: ✅/❌
- Orphaned Data Check: ✅/❌
Issues Found: [description]
Notes: [additional observations]
```

### Actionable Selection Testing Checklist

#### Dependency-Aware Selection Tests
- [ ] Deep nesting (4+ levels) works correctly
- [ ] Depth-first preference selects deeper subtasks
- [ ] Priority weighting influences selection
- [ ] Unblocking potential is calculated correctly
- [ ] In-progress tasks get preference when configured
- [ ] Circular dependencies are detected and reported
- [ ] Deadlock situations are explained clearly

#### MCP vs CLI Consistency Tests
- [ ] `status_actionable` returns same task as `knot status actionable`
- [ ] Reasoning messages include same factors
- [ ] Empty project handling matches
- [ ] All blocked handling matches
- [ ] Error messages are consistent

#### Complex Scenarios
- [ ] Cross-branch dependencies work correctly
- [ ] Multiple dependency chains resolve properly
- [ ] Parent-child relationships don't break selection
- [ ] Mixed priority/depth scenarios handled correctly

### Deletion Testing Checklist

#### Task Deletion Tests
- [ ] Single task deletion works
- [ ] Task with dependencies is handled correctly
- [ ] Parent task with subtasks deletes entire hierarchy
- [ ] No orphaned dependencies after deletion
- [ ] Status queries return clean results after deletion

#### Project Deletion Tests
- [ ] Project deletion removes all tasks
- [ ] All subtasks and nested tasks are deleted
- [ ] All dependencies are cleaned up
- [ ] No orphaned records in database
- [ ] Project list no longer contains deleted project
- [ ] Creating new project with same name succeeds

This structured approach ensures consistent, thorough testing of the knot MCP server functionality while maintaining clear separation between the development workspace and test environment.
