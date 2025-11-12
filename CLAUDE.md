# Brain Tools Usage Guide

## Project Setup
Always load the project UUID from `.project` file when starting work. The current project ID is: `556a7f4f-e31a-4656-af1a-d1e44da032fa`

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