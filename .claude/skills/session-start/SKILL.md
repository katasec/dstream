---
name: session-start
description: Orient at the start of a coding session. Use at the beginning of every session to understand project status, active work, and what to do next.
---

# Session Start

Follow the session continuity protocol from [AGENTS.md](../../../AGENTS.md).

## Steps

1. **Read** `docs/plan.md` — specifically the `## Current Status` block.
2. **Identify active work** from the `## Active Work` table. Follow the link to the relevant spoke file in `docs/plan/`.
3. **Read the spoke file** for full context on the current task.
4. **Summarise** to the user:
   - What was completed last session
   - What's currently in progress
   - What's next
   - Any blockers or decisions needed

## Output Format

```
## Session Orient

**Last session**: [what was completed]
**In progress**: [current active work]
**Next up**: [what to tackle this session]
**Blockers**: [any decisions or dependencies needed, or "None"]

Ready to continue with: [specific task name]
```

If `docs/plan.md` doesn't exist yet or has no active work, say so and ask the user what they'd like to work on.