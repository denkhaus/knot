# Brain Tools Usage Guide

## Project Setup
Always load the project UUID from `.project` file when starting work.

## Workflow Overview

### Dual-Layer System
- **Brain Tools**: Memory layer for storing important notes, user decisions, and project knowledge
- **Knot Tool**: Task management for concrete development tasks

### Proactive Memory Storage Strategy
**Critical**: Immediately store important information to brain tools when the user provides:
- Technical preferences (languages, tools, frameworks)
- Coding style or patterns
- Project requirements or constraints
- User opinions or feedback
- Problem-solving approaches
- Learning style or experience level
- Important decisions made during development
- Best practices for this project

## Brain Tools Purpose
The brain tools serve as a comprehensive knowledge database that:
- Stores user-provided information during programming process
- Maintains important decisions and principles across sessions
- Preserves project-specific best practices
- Provides context for future programming sessions
- Enables continuity of knowledge over time

## Session Workflow
1. **Load Project**: Read `.project` file to get current project UUID
2. **Retrieve Memories**: Search existing memories for relevant context
3. **Proactive Storage**: Store new information immediately when provided by user
4. **Build Knowledge Base**: Continuously expand the memory database
5. **Reference Past Decisions**: Use stored memories to maintain consistency

## Memory Categories to Store
- Technical architecture decisions
- Code style preferences
- Tool and framework choices
- Project constraints and requirements
- User feedback and preferences
- Problem-solving patterns
- Best practices discovered
- Important debugging insights
- Performance optimization decisions
- Security considerations

## Memory Maintenance Protocol

### Keeping Brain Up-to-Date
**Critical**: Brain memories must be actively maintained to ensure accuracy:

1. **Memory Lifecycle Management**:
   - Regularly review and refine existing memories
   - Update memories when new information becomes available
   - Consolidate related memories into comprehensive entries
   - Delete outdated or obsolete memories
   - Ensure all memories remain current and accurate

2. **Before Creating New Memories**:
   - Search for related existing memories first
   - Update existing memories instead of creating duplicates
   - Merge overlapping memories into coherent entries
   - Only create new memories for truly new information

3. **Memory Quality Standards**:
   - Keep memories concise but comprehensive
   - Update memories when project requirements change
   - Refine memories based on new user feedback
   - Delete memories that no longer apply to current context

4. **Regular Maintenance**:
   - Audit memories for current relevance
   - Remove outdated information
   - Update memories when user preferences evolve
   - Use memory relationships to maintain consistency
   - Maintain clean, current knowledge base

This creates a growing knowledge base that ensures consistency and preserves the user's intent across all development sessions while preventing outdated memory accumulation.

## Cross-Reference System: Brain Tools ↔ KNOT Tasks

### Bidirectional Reference Capability
**Critical Enhancement**: Brain memories and KNOT tasks/projects can reference each other for complete traceability:

#### Memory → Task References
- Store KNOT task IDs in memories when referencing specific work items
- Example: Memory ID `b307c53a-6128-4401-8ffd-236ca77e96d9` references test analysis work
- Enables quick lookup of detailed task context when retrieving memories
- Provides complete audit trail from knowledge to implementation

#### Task → Memory References
- Include memory IDs in KNOT task descriptions or use external documentation
- Example: Task ID `32edbaf1-75b0-4418-a911-d94006df7bd6` (JSON consistency) references relevant memories
- Creates comprehensive task context with supporting documentation
- Enables detailed task understanding through linked knowledge base

### Reference Implementation Patterns

**Memory Storage with Task References**:
```
Store memory with task ID in content/metadata:
"Related to KNOT task: [task-id]"
"Supporting documentation for: [task-title]"
```

**Task Management with Memory References**:
```
Update task description with memory references:
"See brain memory: [memory-id] for detailed analysis"
"Related context stored in memory: [memory-title]"
```

### Benefits of Cross-Reference System
1. **Complete Traceability**: From decision (memory) to implementation (task)
2. **Rich Task Context**: Tasks reference detailed knowledge and analysis
3. **Knowledge Integration**: Memories understand their implementation impact
4. **Audit Trail**: Full history of decisions and their execution
5. **Context Preservation**: Detailed knowledge available during task execution
6. **Workflow Continuity**: Seamless transition between knowledge and action

This creates a powerful integrated system where knowledge management (Brain Tools) and task execution (KNOT) work together with full bidirectional reference capabilities.

## Development Workflow Guidelines

### Critical Development Requirements

#### 1. Code Changes Installation Protocol
**MANDATORY**: After any code changes in KNOT, you MUST run `mage build:install` to rebuild and update the globally installed KNOT binary. This ensures that the testing environment uses the updated code with all changes.

**Process**:
```bash
# After making code changes
mage build:install

# This will:
# 1. Compile the latest code
# 2. Install it globally to $GOPATH/bin
# 3. Replace the previous KNOT binary
# 4. Make new functionality available for testing
```

**Why**: The CLI commands you test (`knot task ...`, `knot project ...`, etc.) use the globally installed binary. Without rebuilding, you would be testing the old version without your changes.

#### 2. Tool Usage Policy
**KNOT as Todo Tool Replacement**:
- KNOT Tool is the official drop-in replacement for Todo Tool in Claude Code environments
- Todo Tool should NOT be used for task management anymore
- All task operations should use KNOT commands
- KNOT provides superior hierarchical task management, state tracking, and auto-parent completion

#### 3. Cross-Reference Documentation System
**Task ID and Memory ID References in Code**:
- Always include relevant Task IDs and Memory IDs in code comments
- This creates direct links between implementation and detailed documentation
- Enables quick navigation from code to task context and technical specifications
- Maintains traceability between implementation decisions and requirements

**Comment Examples**:
```go
// Auto-parent completion logic (Task ID: 06afc996-9a4e-4e75-a03d-8289d13042e3)
// See brain memory: 4bd7bc0a-4382-4fb3-8601-4facbbf1abc6 for technical specifications
func (s *service) evaluateAndUpdateParentTask(ctx context.Context, parentID uuid.UUID, actor string) error {
    // Implementation details...
}

// TODO: Use shared actor resolution function from task a42c4861-f7f7-4d03-9a29-53b965a7ee1e
actor := c.String("actor")
if actor == "" {
    actor = os.Getenv("USER")
    if actor == "" {
        actor = "unknown"
    }
}
```

#### 4. Code Artifact Management
**No Build Artifacts in Codebase**:
- All image operations, compilations, and build artifacts must use temporary directories
- Codebase must remain clean from compiled artifacts, images, and temporary files
- Use system temp directories or build-specific output directories
- This keeps the repository clean and focused on source code

**Best Practices**:
```bash
# Use temporary directories for build operations
TEMP_DIR=$(mktemp -d)
# Perform operations in $TEMP_DIR
# Clean up afterwards with rm -rf $TEMP_DIR
```

### Implementation Quality Standards

#### Code Documentation Requirements
- Every significant function should reference related tasks/memories
- Technical decisions should link to brain memories with full specifications
- TODO comments should reference specific task IDs for follow-up
- This creates maintainable code with full context preservation

#### Testing Protocol
1. Make code changes
2. Run `mage build:install` to rebuild KNOT
3. Test functionality with updated binary
4. Update related tasks with test results
5. Store learnings in brain memories for future reference

This workflow ensures that all development work is properly tracked, documented, and maintains high quality standards with full context preservation.
