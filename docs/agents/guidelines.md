# Agent Guidelines & Workspace Conventions

> Spoke of [AGENTS.md](../../AGENTS.md). Load for general orientation and workspace conventions.

## Workspace Layout

<!-- Customize: Fill in your project's directory structure -->
<!--
| Directory | Purpose |
|-----------|---------|
| `src/` | Application source code |
| `tests/` | Test suites |
| `docs/` | Documentation (agents, design, plan) |
| `scripts/` | Build and utility scripts |
-->

## Working Conventions

- **Canonical paths**: Use absolute paths when referencing files in conversation. Use relative paths in code and docs.
- **Method ordering**: Order methods by invocation (top-down) — public entry points first, helpers below. This is progressive disclosure applied to file layout.
- **Focus on declarative, testable implementations**: Prefer pure functions and composition over imperative mutation.
- **Build before push**: Always run a local build + test before pushing. Zero warnings, zero failures.

## Local Dependencies

<!-- Customize: List local repos, SDKs, or tools the agent may need to reference -->
<!--
| Dependency | Location | Notes |
|------------|----------|-------|
| Framework source | `/path/to/framework` | Read source — don't decompile packages |
-->

## External References

<!-- Customize: List relevant documentation links -->
<!--
- Framework docs: https://...
- API reference: https://...
-->