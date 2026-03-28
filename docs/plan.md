# Project Plan

> Hub document for session continuity. Spokes in `plan/`. Track completed work in [completed.md](plan/completed.md).

---

## Current Status

<!-- Update this block at the start and end of every session -->
<!-- Format: What was completed, what's in progress, what's next -->

**Last updated**: 2026-03-29

- Completed: Captured architecture intent and current-state runtime model (HCL + three-process orchestration + OCI/ORAS distribution) in design documentation.
- In progress: Build a validated gap list (doc/code/protocol) and convert into prioritized execution phases.
- Next: Define protocol/spec hardening tasks, ecosystem onboarding tasks, deprecation strategy for legacy gRPC pathway, and snapshot/backfill design for historical data transfer.

---

## Active Work

| Item | Status | Link |
|------|--------|------|
| Current-state architecture baseline | 🚧 In progress | [architecture-baseline.md](plan/architecture-baseline.md) |

<!-- Add rows as work begins. Link to spoke files in plan/ for detail. -->
<!-- Example: -->
<!-- | Implement user auth | 🚧 In progress | [user-auth.md](plan/user-auth.md) | -->

---

## Work Queue

### Next Up

| Item | Priority | Link |
|------|----------|------|
| Protocol specification v1 (command/data envelope contract) | P0 | [architecture-baseline.md](plan/architecture-baseline.md) |
| Runtime behavior parity audit (README vs code) | P0 | [architecture-baseline.md](plan/architecture-baseline.md) |
| Historical snapshot/backfill design (CDC bootstrap) | P0 | [architecture-baseline.md](plan/architecture-baseline.md) |
| Provider authoring guide + compatibility test harness | P1 | [architecture-baseline.md](plan/architecture-baseline.md) |
| Legacy gRPC path deprecation plan | P1 | [architecture-baseline.md](plan/architecture-baseline.md) |

### Before Production

| Item | Link |
|------|------|
| Versioned protocol and compatibility policy | [architecture-baseline.md](plan/architecture-baseline.md) |
| Provider artifact integrity strategy (signing/provenance) | [architecture-baseline.md](plan/architecture-baseline.md) |
| End-to-end conformance tests across sample providers | [architecture-baseline.md](plan/architecture-baseline.md) |
| Historical + incremental handoff strategy for large datasets | [architecture-baseline.md](plan/architecture-baseline.md) |

---

## Completed

> See [completed.md](plan/completed.md) for full history.