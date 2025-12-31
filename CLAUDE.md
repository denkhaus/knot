# knot Project - AI Agent Instructions

## MANDATORY WORKFLOW: Memory + Beads Integration

### 🚨 CRITICAL: Before Starting ANY Work

```bash
# 1. SEARCH memories for context first
mcp__memories__search_memory(query="<topic you're about to work on>")

# 2. THEN check beads for work
bd ready --json
bd list --status in_progress
```

### 🧠 Memory Rules (MANDATORY)

**The 5-Minute Rule**: If you spent >5 minutes on it, store it as a memory.

**ALWAYS create memories for:**
- Architecture decisions → `mcp__memories__add_memories(text="DECISION: ...")`
- Bug root cause analysis → `mcp__memories__add_memories(text="ANALYSIS: ...")`
- Pattern discoveries → `mcp__memories__add_memories(text="PATTERN: ...")`
- Configuration discoveries → `mcp__memories__add_memories(text="CONFIG: ...")`

**ALWAYS cross-reference:**
- Memories → reference beads issues
- Beads issues → reference memories with `[brain:memory-topic]`

**Session End Checklist:**
- [ ] All decisions stored as memories?
- [ ] All bugs analyzed in memories?
- [ ] Beads issues updated with brain references?
- [ ] Ran `mcp__memories__list_memories` to verify?

See @memory.md for complete guidelines and templates.

## Issue Tracking with bd (beads)

This project uses **bd (beads)** for ALL issue tracking. Do NOT use markdown TODOs.

Read @AGENTS.md for complete beads workflow details.

## Coding Standards

- Maintain reusable utility functions at the utils package
- When testing knot create a dedicated test project. DO NOT test on existing knot projects regarding the development of knot itself
