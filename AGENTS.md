# AGENTS.md - Project Context

> Hub document. Behavior prefs + session protocol here; all other conventions in spoke files under `docs/agents/`.

## Agent Behavior Preferences
- **Ask before editing**: Always propose changes and wait for approval before modifying any files.
- **Explain first**: Describe what you plan to change and why before making edits.
- **Provide ballpark estimates**: When discussing task scope or planning work, give rough time estimates (e.g., "~1 hour", "half a session", "2–3 sessions"). These are needed for work planning — don't hedge excessively, just give a practical ballpark.
- **Test every change**: Every new feature or bug fix must include a test that verifies the behavior. Design the test, add it to the test suite, and confirm all tests pass before considering the work complete. No untested code ships.
- **Review every change against the checklist**: After writing or modifying code, check the result against the Code Review Checklist in [code-review.md](docs/agents/code-review.md). No code ships without passing that checklist — structure, testing, naming, dependencies, and refactoring cleanness. This catches duplicated patterns, silent catch blocks, and other issues that are easy to introduce under time pressure.
- **Discover before coding**: Before implementing a significant feature or fixing a complex bug, follow a structured assessment: Discover → Catalogue → Classify → Propose → Execute. No code is written until the proposal step is approved. This prevents fixing the wrong thing, over-engineering, and surprise changes.
- **Small, composable functions**: Follow the Unix philosophy — write small single-responsibility functions that do one thing well, then compose them. No large monolithic methods. For complex flows, compose via the railway/pipeline pattern — small steps piped together, short-circuiting on failure. This keeps the codebase understandable, maintainable, evolvable, and debuggable for both humans and LLMs. See [guidelines.md](docs/agents/guidelines.md) for examples, [code-style.md](docs/agents/code-style.md) for the pipeline pattern.
- **Document mistakes inline**: When a bug, incident, or wrong approach is discovered, record it inline in the relevant spoke — what happened, why, and what was changed. These notes prevent future agents from repeating the same mistake. Format: `**Incident (YYYY-MM-DD):** [what happened → why it failed → what was fixed]`.

## Session Continuity

All progress is tracked in [docs/plan.md](docs/plan.md) (hub) with detailed spoke files in `docs/plan/`. Follow this workflow every session:

1. **Start of session**: Read `docs/plan.md` — the `## Current Status` block tells you where things stand, what's complete, and what's next. The `## Active Work` table links to the relevant spoke file for detail.
2. **Identify active work**: Follow the link from the Active Work table to the relevant spoke file (e.g., `plan/phase-2-feature-x.md`). Read that file for full context on the current task.
3. **During work**: As you complete tasks (checkboxes, status changes), update the **spoke file** inline — check off items, add notes. Update `## Current Status` in `plan.md` if the high-level summary changes.
4. **End of session**: Before signing off:
   - Update `## Current Status` in `docs/plan.md` to reflect what was completed, what's in progress, and what's next.
   - Update the relevant spoke file(s) with detailed progress.
   - **Docs freshness check** — if your changes affected any of the following, update the corresponding doc:
     | Change type | Update |
     |-------------|--------|
     | Core architecture, data models, system design | [docs/design/design.md](docs/design/design.md) |
     | Future improvements, deferred decisions | [docs/design/future-refinements.md](docs/design/future-refinements.md) |
     | Conventions, code style, build/test workflow, extension patterns | AGENTS.md + relevant `docs/agents/` spoke |
     | Progress, phases, task status | [docs/plan.md](docs/plan.md) + relevant `docs/plan/` spoke |
   - Most sessions only touch `plan.md` + one spoke file. Design docs and AGENTS.md change rarely — but when they should, catch it here.
5. **Single source of truth**: Each doc owns its scope. Do NOT duplicate architecture details in plan.md or progress tracking in the design doc. plan.md is the hub (status + index); spoke files have the detail.

---

## Convention Spokes

| Topic | When to load | Detail |
|-------|-------------|--------|
| Code Style & Principles | Writing code | [code-style.md](docs/agents/code-style.md) |
| Code Review | Reviewing or accepting changes | [code-review.md](docs/agents/code-review.md) |
| Code Audit | Running a full project health check | [code-audit.md](docs/agents/code-audit.md) |
| Language Examples | Implementing patterns (Go, C#) | [code-style-examples.md](docs/agents/code-style-examples.md) |
| Testing & CI | Running tests, writing tests, CI workflows | [testing.md](docs/agents/testing.md) |
| Architecture & Design | Understanding the system, onboarding | [architecture.md](docs/agents/architecture.md) |
| Agent Guidelines & Refs | General orientation, workspace conventions | [guidelines.md](docs/agents/guidelines.md) |

## Skills & Hooks

Skills are invocable workflows; hooks are deterministic safety guardrails. Source of truth is `.agents/` (symlinked into each tool's native location).

### Skills

| Skill | Invoke | What it does |
|-------|--------|-------------|
| `/review` | After writing code | Runs the code review checklist against changed files |
| `/session-start` | Start of every session | Reads plan.md, orients on active work |
| `/session-end` | End of every session | Updates plan.md, runs docs freshness check |
| `/audit` | Periodic health check | Full codebase audit against code style guide |

### Hooks

| Hook | When | What it prevents |
|------|------|-----------------|
| `block-dangerous-commands.sh` | Before shell commands | `rm -rf`, force push, pipe-to-shell, fork bombs |
| `protect-secrets.sh` | Before file access | Reading/editing `.env`, `*.pem`, `*.key`, credentials |

See [.agents/hooks/README.md](.agents/hooks/README.md) for wiring instructions per tool.

---

## Plan & Progress

> See [docs/plan.md](docs/plan.md) for implementation status, active work, and phase details.
