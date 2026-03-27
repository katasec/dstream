# Testing & CI

> Spoke of [AGENTS.md](../../AGENTS.md). Load when running tests, writing tests, or debugging CI.

## Testing Workflow

After making code changes, follow this flow to ensure quality and catch issues early:

### Local Testing Steps
1. **Build**: Verify compilation with zero warnings. Treat warnings as errors — fix them before committing.
2. **Unit tests**: Run the fast test suite for logic and mock-level feedback.
3. **Integration tests** (if applicable): Run against a local environment for end-to-end validation.
4. **Cleanup**: Remove build artifacts and temp files.

<!-- Customize: Replace with your project's actual commands -->
<!--
### Key Commands
| Command | What it does |
|---------|-------------|
| `make build` | Compile the project |
| `make test` | Run unit tests |
| `make test-integration` | Run integration tests |
| `make clean` | Clean build artifacts |
-->

### CI/CD
<!-- Customize: Describe your CI pipeline triggers and steps -->
- CI runs automatically on pushes/PRs to `main`.
- Local testing prevents CI surprises — always test before pushing.

---

## Testing Principles

### Test Every Change
Every new feature or bug fix must include a test that verifies the behavior. No untested code ships.

### Test Fixtures Are Truth
If your project uses golden files, fixture files, or snapshot tests:
- **Fixtures are read-only** — they represent expected output from a known-good source. Never modify them to match your code's output.
- **Direction of adaptation** — when a variance is found, the fix goes into your code, not the fixture.
- **Never generate fixtures from your own output** — this creates a circular reference where regressions silently absorb into the fixtures.

### Edge Cases
Always consider: null/nil inputs, empty collections, boundary values, unresolved template patterns, concurrent access.

### Test Completeness
When a new capability is added to the system, all existing test generators must be updated to exercise it. A feature that isn't covered by integration tests is untested — even if unit tests cover it in isolation.

---

## Test Categories

<!-- Customize: Fill in with your project's test categories -->

| Category | Scope | Speed | When to run |
|----------|-------|-------|-------------|
| Unit | Single function / class | Fast | Every change |
| Integration | Multiple components, real dependencies | Medium | Before push |
| E2E | Full system | Slow | CI / pre-release |