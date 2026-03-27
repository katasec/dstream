---
name: session-end
description: Wrap up a coding session by updating project status and running the docs freshness check. Use at the end of every session before signing off.
---

# Session End

Follow the end-of-session protocol from [AGENTS.md](../../../AGENTS.md).

## Steps

1. **Update `docs/plan.md`**:
   - Update `## Current Status` to reflect what was completed, what's in progress, and what's next.
   - Update the `## Active Work` table if items changed status.

2. **Update spoke file(s)** in `docs/plan/`:
   - Check off completed items.
   - Add notes on decisions made or issues encountered.

3. **Docs freshness check** — review what changed this session and update the corresponding docs:

   | If you changed... | Update... |
   |---|---|
   | Core architecture, data models, system design | `docs/design/design.md` |
   | Future improvements, deferred decisions | `docs/design/future-refinements.md` |
   | Conventions, code style, build/test workflow | `AGENTS.md` + relevant `docs/agents/` spoke |
   | Progress, phases, task status | `docs/plan.md` + relevant `docs/plan/` spoke |

4. **Summarise** to the user what was updated.

## Output Format

```
## Session Wrap-Up

**Completed this session**:
- [list of completed items]

**Updated docs**:
- docs/plan.md — status updated
- docs/plan/[spoke].md — [items checked off]
- [any other docs updated via freshness check]

**Next session should start with**:
- [specific task or area]
```