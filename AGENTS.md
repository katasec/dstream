# AGENTS.md — dstream

> DStream CLI orchestrator. Part of the [DStream ecosystem](https://github.com/katasec/dstream-mission-control).

## Role

`dstream` is the **central orchestrator** — it reads HCL task definitions, resolves provider binaries (local path or OCI reference), and wires input → output providers via stdin/stdout pipes. It is the hub; providers are the spokes.

## Design Docs

- Architecture & protocol: `docs/design/design.md`
- Plugin system: `docs/plugins/`
- Ecosystem inventory: [dstream-mission-control/docs/repository-inventory.md](https://github.com/katasec/dstream-mission-control/blob/main/docs/repository-inventory.md)

## Code Style (Go)

**Governing principle: Progressive Disclosure.** Code reveals intent in layers — what at the top, how one level deeper.

- **Outline-first**: The top 15–20 lines of any file or function must disclose intent and flow. If a reader must scroll to understand what a function does, refactor.
- **Small, composable functions**: Each function is one named step (~20–40 lines max). Callers read step names; drilling in reveals implementation.
- **Top-down method ordering**: Entry points first, helpers below. File reads like an outline.
- **Error handling**: Handle errors explicitly. No `_ = fn()`. No silent swallowing. Return errors up the chain or log them at the boundary.
- **No deeply nested branching**: Max 2 levels. Use early returns.
- **Side effects isolated**: DB, network, file, process exec — in clearly named functions, not mixed with logic.
- **Zero warnings**: Build must pass with zero warnings. Treat warnings as errors.
- **No speculative abstractions**: Build for what the task requires. Three similar lines of code is better than a premature abstraction.

## Behaviour Rules

- **Propose before editing**: Describe what you're about to change and why. Wait for a clear go-ahead before modifying files.
- **Test every change**: No untested code ships. Run `go test ./...` before considering work done.
- **Build before push**: `go build ./...` must succeed with zero errors and zero warnings.
- **Focus**: Only change what the task requires. Do not refactor surrounding code, add comments, or improve unrelated areas.
- **Document mistakes**: If something goes wrong, record it inline: `**Incident (YYYY-MM-DD):** [what happened → why → what was fixed]`.

## Provider Protocol

Providers communicate over stdin/stdout using JSON envelopes. The orchestrator sends a startup command envelope and streams data between input and output providers. Do not break this contract — changes to envelope format require a protocol version bump.

## Task Context

Tasks arrive via GitHub Actions `workflow_dispatch` from `katasec/dstream-mission-control`. The issue body is your primary context. Read `## Task`, `## Context`, and `## Acceptance criteria` carefully before writing any code.
