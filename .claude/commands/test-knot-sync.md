---
name: test-knot-sync
description: Comprehensive test for knot sync functionality between local SQLite and MCP PostgreSQL environments
parameters:
  type: object
  properties:
    cleanup:
      type: boolean
      default: true
      description: Whether to cleanup test data after testing
    project_name:
      type: string
      default: "sync-test"
      description: Base name for the test project
    test_mode:
      type: string
      enum: ["all", "push", "pull", "bidirectional", "quick"]
      default: "quick"
      description: Which sync mode(s) to test (quick = smoke test only)
    skip_restart:
      type: boolean
      default: false
      description: Skip docker-compose restart (use if server already running)
---


## 🚨 CRITICAL: ALWAYS USE ./knot_local

**NEVER use the global `knot` command during testing!**

- ❌ `knot project list` - WRONG! Uses outdated global binary
- ✅ `./knot_local project list` - CORRECT! Uses current development build

**Why this matters:**
- `knot` (global) = installed version, possibly outdated
- `./knot_local` = your latest code changes, built from source
- Testing with `knot` gives FALSE NEGATIVES - you're not testing your changes!

**Before testing, always run:**
```bash
go build -o knot_local ./cmd/knot
```

**If you accidentally use `knot`, stop and restart with `./knot_local`!**


# Knot Sync Functionality Test Workflow

This command tests the sync functionality between local SQLite (CLI) and MCP PostgreSQL (server) environments.

## Quick Reference

### Critical Distinctions
- **Local CLI**: Uses `./knot_local` (local build) with SQLite at `.knot/knot.db` at the workspace root.
- **MCP Server**: PostgreSQL in Docker, accessed via MCP tools
- **⚠️ ALWAYS use `./knot_local`** - global `knot` may be outdated

### MCP vs CLI Commands

| Operation | CLI Command | MCP Tool |
|-----------|-------------|----------|
| List projects | `./knot_local project list` | `mcp__knot-mcp__project_list` |
| Create project | `./knot_local project create --t "Title"` | `mcp__knot-mcp__project_create` |
| Select project | `./knot_local project select --id <ID>` | `mcp__knot-mcp__project_select` |
| List tasks | `./knot_local task list` | `mcp__knot-mcp__status_ready` |
| Create task | `./knot_local task create -t "Title"` | `mcp__knot-mcp__task_create` |
| Add dependency | `./knot_local dependency add` | `mcp__knot-mcp__dependency_add` |
| List dependencies | `./knot_local dependency list -id <ID>` | `mcp__knot-mcp__dependency_list` |
| Sync | `./knot_local sync --project-id <ID> --direction <dir>` | N/A (server receives) |

## Prerequisites

### 1. Build and Start MCP Server

```bash
# Build local knot binary
go build -o knot_local ./cmd/knot

# Start Docker services
docker compose up -d --build

# Verify services
docker compose ps
# Both services should be "healthy"
```

### 2. Learn Knot CLI Functions

**IMPORTANT: Before testing, review all available knot commands and features**

```bash
# Display comprehensive knot CLI documentation
./knot_local get-started
```

This shows all available commands, workflows, and best practices. Review this to ensure proper usage of the knot tool during testing.

### 3. Verify Setup

```bash
# Check MCP tools available. NOTE: This are mcp tools available to the agent. **NOT commands available to bash**
mcp__knot-mcp__project_list

# Check CLI works
./knot_local project list
```

## Test Modes

### Quick Test (Smoke Test)
Fast verification that sync works at all.

```bash
# Create local project
./knot_local project create -t "Quick Test" -d "Smoke test"
PROJECT_ID=$(./knot_local project list | grep -A1 "Quick Test" | grep "ID:" | cut -d: -f2 | xargs)

./knot_local project select -id $PROJECT_ID
./knot_local task create -t "Test Task" -c 5
TASK_ID=$(./knot_local task list | grep "Test Task" | grep "ID:" | cut -d: -f2 | xargs)

# Add second task and dependency
./knot_local task create -t "Dependent Task" -c 3
DEP_ID=$(./knot_local task list | grep "Dependent Task" | grep "ID:" | cut -d: -f2 | xargs)
./knot_local dependency add --task-id $DEP_ID --depends-on $TASK_ID

# Push sync
./knot_local sync -p $PROJECT_ID -d push

# Verify on MCP NOTE: This are mcp tools available to the agent. **NOT commands available to bash**
mcp__knot-mcp__project_select --project_id=$PROJECT_ID
mcp__knot-mcp__dependency_list --task_id=$DEP_ID

echo "✅ Quick test PASSED" if dependency exists; else echo "❌ Quick test FAILED"
```

### Push Mode Test
Verify local → MCP sync with dependencies.

```bash
# Setup
./knot_local project create -t "Push Test" -d "Testing push sync"
P_ID=$(./knot_local project list | grep "Push Test" | grep -oP 'ID: \K[^ ]+')
./knot_local project select -id $P_ID

# Create task hierarchy
./knot_local task create -t "Parent" -c 5
PARENT=$(./knot_local task list | grep "Parent" | grep -oP 'ID: \K[^ ]+')
./knot_local task create --parent-id $PARENT -t "Child 1" -c 3
CHILD1=$(./knot_local task list | grep "Child 1" | grep -oP 'ID: \K[^ ]+')
./knot_local task create --parent-id $PARENT -t "Child 2" -c 4
CHILD2=$(./knot_local task list | grep "Child 2" | grep -oP 'ID: \K[^ ]+')

# Add dependency
./knot_local dependency add --task-id $CHILD2 --depends-on $CHILD1

# Sync and verify
./knot_local sync --project-id $P_ID --direction push

# Verify on MCP NOTE: This are mcp tools available to the agent. **NOT commands available to bash**
mcp__knot-mcp__project_select --project_id=$P_ID
mcp__knot-mcp__dependency_list --task_id=$CHILD2
# Should show Child 2 depends on Child 1
```

### Pull Mode Test
Verify MCP → local sync with dependencies.

```bash
# Create on MCP NOTE: This are mcp tools available to the agent. **NOT commands available to bash**
MCP_PROJECT=$(mcp__knot-mcp__project_create --title="Pull Test" --description="Testing pull" | jq -r '.project_id')
mcp__knot-mcp__project_select --project_id=$MCP_PROJECT

PARENT=$(mcp__knot-mcp__task_create --title="MCP Parent" --complexity=5 | jq -r '.task_id')
SUB1=$(mcp__knot-mcp__task_create --parent_id=$PARENT --title="Sub 1" --complexity=3 | jq -r '.task_id')
SUB2=$(mcp__knot-mcp__task_create --parent_id=$PARENT --title="Sub 2" --complexity=4 | jq -r '.task_id')

mcp__knot-mcp__dependency_add --task_id=$SUB2 --depends_on_task_id=$SUB1

# Pull and verify locally
./knot_local sync -p $MCP_PROJECT -d pull
./knot_local project select -id $MCP_PROJECT
./knot_local dependency list --task-id $SUB2
# Should show Sub 2 depends on Sub 1
```

### Bidirectional Test
Verify merge of divergent changes.

```bash
# Create local and sync
./knot_local project create -t "Bi Test" -d "Bidirectional sync"
BI_ID=$(./knot_local project list | grep "Bi Test" | grep -oP 'ID: \K[^ ]+')
./knot_local project select -id $BI_ID
./knot_local task create -t "Local Task"
./knot_local sync -p $BI_ID -d bi

# Modify on both sides
LOCAL_TASK=$(./knot_local task list | grep "Local Task" | grep -oP 'ID: \K[^ ]+')
./knot_local task update --id=$LOCAL_TASK -t "Local Updated"
MCP_TASK=$(mcp__knot-mcp__status_ready --limit=1 | jq -r '.tasks[0].id')
mcp__knot-mcp__task_update --task_id=$MCP_TASK --title="MCP Updated"

# Sync bidirectionally
./knot_local sync -p $BI_ID -d bi

# Verify both changes present
./knot_local task list | grep "Local Updated"
mcp__knot-mcp__status_ready | grep "MCP Updated"
```

## Helper Functions

### Extract IDs from Output
```bash
# Get project ID by title
get_project_id() {
  ./knot_local project list | grep -A2 "$1" | grep "ID:" | cut -d: -f2 | xargs
}

# Get task ID by title
get_task_id() {
  ./knot_local task list | grep -A2 "$1" | grep "ID:" | cut -d: -f2 | xargs
}
```

### Verify Dependency Sync
```bash
verify_dependency() {
  local task_id=$1
  local dep_title=$2

  # Check locally
  local has_dep=$(./knot_local dependency list --task-id $task_id | grep -c "$dep_title")

  # Check on MCP
  local mcp_has_dep=$(mcp__knot-mcp__dependency_list --task_id=$task_id | grep -c "$dep_title")

  if [ $has_dep -gt 0 ] && [ $mcp_has_dep -gt 0 ]; then
    echo "✅ Dependency verified: $dep_title"
    return 0
  else
    echo "❌ Dependency missing: $dep_title"
    return 1
  fi
}
```

### Clean Test Projects
```bash
# Clean local test projects
./knot_local project list | grep "Test" | while read line; do
  id=$(echo "$line" | grep -oP 'ID: \K[^ ]+')
  [ -n "$id" ] && ./knot_local project delete --id $id <<EOF
yes
EOF
done

# Clean MCP test projects
docker compose exec -T postgres psql -U knot_user -d knot_test -c "DELETE FROM projects WHERE title LIKE '%Test%';" 2>/dev/null
```

## Troubleshooting

### Container Issues
```bash
# Restart services
docker compose down && docker compose up -d --build

# Check logs
docker compose logs knot-mcp --tail=50

# Verify database connection
docker compose exec postgres pg_isready -U knot_user -d knot_test
```

### CLI Shows "No such command"
```bash
# Rebuild local binary
go build -o knot_local ./cmd/knot

# Verify sync command exists
./knot_local sync --help
```

### Dependencies Not Syncing
```bash
# Check if GetTasksWithDependencies is being called
export KNOT_LOG_LEVEL=debug
./knot_local sync -p $PROJECT_ID -d pull 2>&1 | grep -i "dependency"

# Verify endpoint returns dependencies
curl -s "http://localhost:9094/api/sync/tasks?project_id=$PROJECT_ID" | jq .
# Tasks should have "dependencies" array
```

### Quick Verification Commands
```bash
# Count tasks locally vs MCP
local_count=$(./knot_local task list | grep -c "ID:")
mcp_count=$(mcp__knot-mcp__status_ready | jq '.total')
echo "Local: $local_count, MCP: $mcp_count"

# Compare specific task
./knot_local task get --id $TASK_ID
mcp__knot-mcp__task_get --task_id=$TASK_ID
```

## Expected Results

### Success Criteria
- ✅ Local cli commands and remote mcp tool calls work without error
- ✅ Projects sync in both directions
- ✅ Tasks with all fields sync correctly
- ✅ **Dependencies sync in ALL modes** (push, pull, bidirectional)
- ✅ Parent-child relationships preserved
- ✅ State transitions sync correctly
- ✅ Timestamps, Actor (UpdatedBy) on tasks and projects preserved
- ✅ No duplicate or orphaned records

### Critical Validation Points
1. **Dependencies**: Must be present in both local and MCP after sync
2. **Parent IDs**: Subtasks must maintain parent relationships
3. **States**: Task state changes propagate correctly
4. **Timestamps**: CreatedAt matches, UpdatedAt reasonable

## Cleanup

```bash
# Automatic cleanup (default)
# Clean test projects from both environments
# See helper function above

# Manual cleanup
docker compose down -v  # Remove volumes (fresh database)
docker compose up -d --build  # Restart fresh
```

## 🚨 Bug Reporting with Beads

### When to Create Bug Issues

**IMMEDIATE ACTION REQUIRED**: When any test fails or shows unexpected behavior, create a Beads issue BEFORE continuing with further testing.

### Bug Creation Workflow

```bash
# 1. Stop testing when bug is discovered
# 2. Create bug issue
bd create \
  --title "[SYNC BUG] <short-description>" \
  --type bug \
  --priority <0-4> \
  --description "
## Problem Description
<What went wrong>

## Reproduction Steps
1. <Step 1>
2. <Step 2>
3. <Step 3>

## Expected Behavior
<What should have happened>

## Actual Behavior
<What actually happened>

## Environment
- Test Phase: <Push/Pull/Bidirectional/Dependencies/State>
- Sync Mode: <push/pull/bi>
- Container Rebuilt: <yes/no>

## Test Data
Project ID: <ID>
Task IDs: <IDs>
Command: <exact command used>

## Error Messages
<Any errors or logs>
"

# 3. Verify issue was created
bd show <issue-id>

# 4. Only continue testing if bug is not blocking
```

### Priority Guidelines

| Priority | When to Use | Examples |
|----------|-------------|----------|
| **P0** | Data loss, sync completely broken | Projects deleted, dependencies lost |
| **P1** | Core sync functionality fails | Push/pull/bi doesn't work |
| **P2** | Partial functionality broken | Specific field not syncing |
| **P3** | Minor issues | Cosmetic problems, rare edge cases |
| **P4** | Enhancements | Performance, logging improvements |

### Bug Issue Template Structure

```bash
## Acceptance Criteria
- [ ] Bug is reproducible
- [ ] Fix implemented
- [ ] All sync modes pass
- [ ] No regressions

## Related Information
- Discovered during test-knot-sync
- Relevant files: <file:line> if known
```

## Best Practices

### During Development
1. **Always use `./knot_local`** - rebuild after code changes
2. **Test small increments** - verify each sync mode separately
3. **Check dependencies immediately** - they're the first thing to break
4. **Use debug logging** - `KNOT_LOG_LEVEL=debug` for issues

### When Code Changes
```bash
# 1. Rebuild binary
go build -o knot_local ./cmd/knot

# 2. Restart container
docker compose down && docker compose up -d --build

# 3. Run quick test
# (see Quick Test section above)

# 4. If quick test passes, run full tests
```

### Common Pitfalls
- ❌ Using global `knot` instead of `./knot_local`
- ❌ Forgetting to rebuild container after code changes
- ❌ Not verifying dependencies in both local and MCP
- ❌ Skipping project selection before task operations
- ✅ Always verify on BOTH sides after sync
