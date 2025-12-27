# Memory MCP - Proactive Knowledge Building Guideline

## Philosophy: "Never Forget a Decision"

The Memory MCP server is **unentbehrlich** (indispensable) for our workflow. It transforms transient conversations into persistent, searchable knowledge. Every significant decision, insight, and pattern discovery should be captured.

---

## Quick Start: Daily Usage

```bash
# BEFORE starting work - search for context
mcp__memories__search_memory(query="topic you're working on")

# WHEN learning something new - capture it
mcp__memories__add_memories(text="DECISION: What you learned and why...")

# WHEN debugging - store the analysis
mcp__memories__add_memories(text="ANALYSIS: Root cause and solution...")

# WHEN discovering patterns - document them
mcp__memories__add_memories(text="PATTERN: Reusable approach you found...")

# AFTER finishing work - verify coverage
mcp__memories__list_memories
```

**The Golden Rule:** If you spent >5 minutes on it, store it.

---

## Core Principles

### 1. Capture Early, Capture Often
**Rule of thumb:** When in doubt, store it. It's better to have a memory you don't need than to lose knowledge you'll desperately search for later.

### 2. The 5-Minute Rule
If you spent more than 5 minutes researching, debugging, or reasoning about something - **store it as a memory**. This includes:
- Root cause analysis of bugs
- Architecture decisions
- Configuration discoveries
- Command patterns that work
- API integration learnings

### 3. Decision = Memory
Every time you say "we decided to..." or "let's use..." - that's a memory. Capture:
- **What** was decided
- **Why** it was decided (alternatives considered)
- **When** it was decided (context)
- **Trade-offs** acknowledged

### 4. Cross-Reference Everything
Memories should reference beads issues, beads issues should reference memories. This creates a navigable knowledge web.

---

## Proactive Memory Triggers

### Automatic Triggers (ALWAYS create memory):

| Trigger | Example | Memory Topic |
|---------|---------|--------------|
| Architecture decisions | "Use SSE for real-time updates" | architecture,sse,realtime |
| Root cause analysis | "Race condition caused by..." | bug,concurrency,analysis |
| Pattern discovery | "Defer pattern prevents leaks" | patterns,golang,error-handling |
| Technology comparison | "Chose PostgreSQL over MongoDB" | decision,database,comparison |
| Configuration discoveries | "KNOT_LOG_LEVEL=debug enables verbose" | config,debugging,knot |
| Performance insights | "Connection pooling reduces latency 40%" | performance,optimization,database |
| Security considerations | "RS256 allows key rotation" | security,jwt,authentication |
| API integration learnings | "Retry-after header for rate limits" | api,integration,patterns |
| Testing discoveries | "Table tests require struct equality" | testing,golang,discoveries |
| Tooling setup | "Mock generation with go.uber.org/mock" | tooling,mocks,testing |

### Contextual Triggers (use judgment):

| Trigger | When to Store |
|---------|---------------|
| Simple implementation details | Only if non-obvious or clever |
| Workarounds | Always - document why and long-term fix |
| Environment-specific choices | When it differs from local to prod |
| Workflow improvements | When it saves time repeatedly |
| Code refactoring decisions | When pattern is reusable |

---

## Memory Format: Structured for Retrieval

### Decision Template
```
DECISION: [Clear, action-oriented title]

Context:
- Beads Issue: beads-123
- Date: 2025-01-XX
- Phase: [design|implementation|debugging|review]

Decision:
We chose [X] over [Y, Z] because [reasons].

Alternatives Considered:
- [Alternative A]: [pros/cons]
- [Alternative B]: [pros/cons]

Trade-offs:
- Pro: [benefit]
- Con: [cost/limitation]

Implementation Notes:
[Specific details about how it was implemented]

Cross-References:
- Related: beads-456, [brain:abc-def]
- Builds on: [previous-memory-topic]
```

### Bug Analysis Template
```
ANALYSIS: [Bug title]

Context:
- Beads Issue: beads-789
- Symptom: [what went wrong]
- Impact: [severity/scope]

Root Cause:
[Technical explanation of why it happened]

Solution:
[How we fixed it]

Prevention:
[How to prevent this in the future]

Cross-References:
- Fixed in: beads-790
- Related patterns: [brain:xyz-123]
```

### Pattern Discovery Template
```
PATTERN: [Pattern name]

Discovery Context:
- Found while: [beads issue or task]
- Language/Framework: [Go, etc.]

Pattern:
[Clear description of the reusable pattern]

When to Use:
- [Condition 1]
- [Condition 2]

Example:
[Code snippet or concrete example]

Cross-References:
- Used in: beads-123, beads-456
```

---

## Workflow Integration

### During Work Session

```
1. START: Search memories for context
   ↓
   mcp__memories__search_memory(query="topic keywords")

2. DISCOVERY: When learning something new
   ↓
   mcp__memories__add_memories(text="structured memory content")

3. DECISION: When making a choice
   ↓
   mcp__memories__add_memories(text="DECISION: ...")

4. REFERENCE: Update beads with memory link
   ↓
   bd update beads-123 --notes="Reference: [brain:memory-topic]"

5. END: Review session - capture any missed insights
   ↓
   mcp__memories__search_memory(query="today's topics") to verify coverage
```

### Session Start Ritual

```bash
# 1. Search for relevant context
mcp__memories__search_memory(query="architecture authentication")

# 2. Check beads for work in progress
bd ready
bd list --status in_progress

# 3. For each active issue, check for brain references
bd show beads-123 | grep "brain:"

# 4. Retrieve referenced memories
mcp__memories__search_memory(query="authentication jwt")
```

### Session End Ritual

```bash
# Review today's work - what should be captured?

1. Any decisions made? → Store as DECISION memory
2. Any bugs discovered/fixed? → Store as ANALYSIS memory
3. Any patterns discovered? → Store as PATTERN memory
4. Update beads issues with brain references

# Verify coverage:
mcp__memories__list_memories  # Review recent additions
```

---

## Search Strategy: Finding What You Need

### Effective Search Queries

| Goal | Query Example |
|------|---------------|
| Find architecture decisions | "architecture decision" |
| Find bug patterns | "root cause analysis" |
| Find Go patterns | "golang pattern" |
| Find configuration | "config environment setup" |
| Find API knowledge | "api integration rest" |
| Find testing approaches | "testing pattern mock" |

### Query Refinement

1. **Start broad**, then narrow:
   - First: "authentication"
   - Then: "authentication jwt"
   - Finally: "authentication jwt RS256 key rotation"

2. **Use multiple terms** for semantic search:
   - "concurrency race condition go"
   - "performance optimization database"

3. **Add context terms**:
   - "decision database postgresql"
   - "pattern error handling golang"

---

## Memory Maintenance

### Weekly Review (5 minutes)

```bash
# 1. List recent memories
mcp__memories__list_memories

# 2. Scan for:
#    - Duplicates (merge related memories)
#    - Outdated info (update or note evolution)
#    - Missing cross-references (add beads links)

# 3. Verify memories have corresponding beads references:
#    - Decision memories → should reference implementation beads
#    - Analysis memories → should reference bug fix beads
#    - Pattern memories → should reference usage beads
```

### Monthly Consolidation (15 minutes)

```bash
# 1. Search by topic area
mcp__memories__search_memory(query="architecture")

# 2. Look for consolidation opportunities:
#    - Multiple small memories on same topic → merge into comprehensive guide
#    - Related decisions → create "decision history" memory
#    - Scattered patterns → create "pattern library" memory

# 3. Update consolidated memories
mcp__memories__add_memories(text="COMPREHENSIVE: ...")
mcp__memories__delete_memories(memory_ids=["old-1", "old-2"])
```

---

## Anti-Patterns: What NOT to Do

### ❌ Don't Store:
- Trivial findings (< 1 minute to discover)
- Transient information (temporary workarounds without plan to fix)
- Obvious implementation details (standard API usage)
- Information already in code comments (unless it's a WHY comment)
- Duplicate information (search before storing)

### ❌ Don't Vague:
- "Fixed a bug" → "ANALYSIS: Race condition in token validation under concurrent load"
- "Used Go" → "DECISION: Use Go for backend services due to built-in concurrency"
- "Changed config" → "DECISION: Increase connection pool to 50 for 40% latency reduction"

### ❌ Don't Isolate:
- Never store without considering: "Will I be able to find this again?"
- Always include: context, cross-references, searchable keywords

---

## Quality Checklist

Before moving on from a task, ask:

- [ ] Did we make any decisions? → Store as DECISION memory
- [ ] Did we debug anything? → Store as ANALYSIS memory
- [ ] Did we discover any patterns? → Store as PATTERN memory
- [ ] Are beads issues updated with brain references?
- [ ] Is the memory searchable? (good keywords in content)
- [ ] Will I understand this memory 6 months from now?

---

## Examples: Good vs Bad

### ❌ Bad Memory
```
Changed the database connection pool
```

### ✅ Good Memory
```
DECISION: Increase PostgreSQL connection pool to 50

Context:
- Beads Issue: beads-456
- Date: 2025-01-15
- Symptom: High latency during peak load

Decision:
Increase connection pool from 10 to 50 based on load testing.

Analysis:
- 10 connections caused connection wait times under load
- 50 connections reduced p95 latency by 40%
- 100 connections showed diminishing returns

Trade-offs:
- Pro: Better latency, handles peak load
- Con: Increased memory usage (~50MB), requires connection monitoring

Implementation:
internal/config/postgres.go: Set MaxOpenConns=50, MaxIdleConns=25

Cross-References:
- Implemented in: beads-456
- Load test results: beads-455
- Related: [brain:performance-testing-patterns]
```

---

## Tool Reference

### How the Memory System Actually Works

**Understanding the automatic behavior:**

1. **Granular Splitting**: When you add structured content, the system automatically splits it into focused, individual memories
   - One `add_memories` call → Multiple discrete memories stored
   - Each memory represents a single fact/decision/concept
   - Better for semantic search and retrieval

2. **Semantic Search**: Returns ranked results with relevance scores (0.0 to 1.0)
   - Higher score = better match to your query
   - Uses semantic understanding, not just keyword matching
   - Finds related concepts even with different wording

3. **Memory Metadata** (stored automatically):
   - `id`: UUID for deletion/management
   - `hash`: Content fingerprint for deduplication
   - `created_at`: When the memory was added
   - `updated_at`: Last modification (null if never updated)
   - `metadata.source_app`: Which app created it
   - `metadata.mcp_client`: Which MCP client (e.g., "claude")

### Memory MCP Tools

```bash
# Add memories (system auto-splits into granular pieces)
mcp__memories__add_memories \
  text="DECISION: Your structured content here..."

# Returns array of created memories with IDs:
# {"results": [
#   {"id": "uuid-1", "memory": "First fact extracted", ...},
#   {"id": "uuid-2", "memory": "Second fact extracted", ...}
# ]}

# Search memories (semantic search with relevance scores)
mcp__memories__search_memory \
  query="keywords describing what you're looking for"

# Returns ranked results:
# {"results": [
#   {"id": "uuid", "memory": "Content", "score": 0.85, ...},
#   ...
# ]}

# List all memories (with full metadata)
mcp__memories__list_memories

# Delete specific memories (by UUID)
mcp__memories__delete_memories \
  memory_ids=["uuid-1", "uuid-2"]

# Delete all memories (use with extreme caution!)
mcp__memories__delete_all_memories
```

### Optimal Memory Formatting

Since the system auto-splits content:

**DO:**
```
DECISION: Use PostgreSQL for primary database

Context:
- Beads Issue: beads-456
- Date: 2025-01-26

Decision:
PostgreSQL chosen over MongoDB for relational data integrity.

Reasons:
- ACID transactions required for financial data
- Complex joins needed for reporting
- Strong SQL skills in team

Trade-offs:
- Pro: Data integrity, complex queries, mature ecosystem
- Con: More complex schema migrations vs document DB

Cross-References:
- Related: beads-123 (schema design)
- Pattern: [brain:relational-database-patterns]
```

**DON'T (one big block):**
```
DECISION: PostgreSQL. We chose Postgres. It has ACID. We need transactions. Also MongoDB was considered. We said no to Mongo. The team knows SQL. We need joins. ACID is good for financial data.
```

The system will split the first example into focused, searchable memories. The second creates confused, overlapping fragments.

---

## Making It Indispensable

### How to Build the Habit

1. **Week 1**: Set a timer every 30 minutes - ask "Did I learn anything worth storing?"
2. **Week 2**: After completing any task, ask "What should I remember from this?"
3. **Week 3**: Before starting work, search memories - demonstrate value by finding past knowledge
4. **Week 4**: It's automatic - you can't imagine working without it

### Measuring Success

You're using the memory system effectively when:
- ✅ You search memories before making decisions
- ✅ You find past solutions to current problems
- ✅ New team members can search memories to get up to speed
- ✅ You can explain "why" the code is the way it is (from memories)
- ✅ You rarely have to rediscover the same thing twice

---

## Summary: The Memory Mindset

> "The best time to plant a tree was 20 years ago. The second best time is now."
> — Chinese Proverb

**The best time to start storing memories was at the beginning of the project. The second best time is now.**

Every conversation, every decision, every discovery - capture it. Build a knowledge base that makes you and your team exponentially more effective over time.

The memory system isn't a documentation tool - it's your **second brain**. Use it proactively, use it constantly, and watch your project's collective intelligence grow.

---

*Generated for the knot project to ensure no important decision is ever forgotten.*
