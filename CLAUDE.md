- maintain reusable utility functions at the utils package

# CRITICAL WARNING: NEVER TEST DELETION ON FOUNDATIONAL TASKS

## The Incident
When testing the deletion command update, I accidentally deleted the "Create Shared Tree Formatting Utilities" task (ID: 77540a50-676f-402e-8a45-d6272c51b7b2) which was the root task containing all our tree formatting work.

## What Went Wrong
- Tested deletion command on actual production tasks
- Used real tasks instead of creating test tasks
- Didn't consider the impact of subtree deletion on dependencies
- Failed to use dry-run mode for initial testing

## Prevention Rules
1. **NEVER test deletion on tasks you're actively working on**
2. **ALWAYS create dedicated test tasks for deletion testing**
3. **USE dry-run mode first for any deletion testing**
4. **CHECK task dependencies before testing deletion**
5. **VERIFY task hierarchy before deleting anything**
6. **CONSIDER using a separate test project for destructive operations**

## Safe Testing Practices
```bash
# Always test with dry-run first
knot task delete --id [TASK_ID] --all --dry-run

# Create test tasks specifically for testing
knot task create --title "TEST DELETION" --description "Test task for deletion command"

# Check dependencies before deletion
knot task dependencies --id [TASK_ID]
```

## Recovery Steps
If accidental deletion occurs:
1. Cancel deletion immediately if still in deletion-pending state
2. Check if tasks can be restored from backups
3. Recreate lost tasks and their relationships
4. Verify all dependent work is still functional

## Consequences
- Loss of foundational work
- Broken dependencies
- Reverted progress
- Time wasted recreating work
- Potential project delays

**LESSON: Always consider the impact before testing destructive operations on production data!**
