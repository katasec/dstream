# Code Review Checklist

> Spoke of [code-style.md](code-style.md). Load when reviewing or accepting changes.

Before accepting any change — whether authored by a human, pair-programmed with an LLM, or generated — verify:

## Structure & Disclosure
- [ ] Top-level intent clear in 15–20 lines (the outline-first rule)
- [ ] No functions over ~40–50 lines without clear decomposition
- [ ] No duplicated patterns — check for existing helpers before writing new ones
- [ ] No silent error swallowing — no empty `catch { }`, no `_ = fn()` (Go), no unchecked return values. Handle errors explicitly or log them
- [ ] No deeply nested branching (>2 levels) — refactor to early returns or extract methods
- [ ] No runtime data embedded in the binary — configuration belongs in config files or external stores

## Testing
- [ ] New public methods have at least one test exercising the happy path
- [ ] Custom parsers / serializers have unit tests (high-risk for subtle breakage)
- [ ] Edge cases considered: null inputs, empty collections, unresolved template patterns
- [ ] Existing tests still pass — no silent regressions

## Naming & Intent
- [ ] Function names describe *what* they do, not *how* (e.g., `ResolveScaling` not `CascadeOverrides`)
- [ ] No stale comments (especially `// TODO`, path references, orphan doc links)

## Dependencies & Side Effects
- [ ] Side effects isolated to clearly named functions — not mixed with pure logic
- [ ] No new dependencies unless justified (check if existing packages cover the need)
- [ ] Imports cleaned up — no unused imports left behind

## Refactoring-Specific
- [ ] After moving code: original location cleaned up (no empty directories, no orphan references)
- [ ] Package / namespace matches project intent (flag mismatches, don't ignore them)
- [ ] Regression tests ran after every move — not batched

---

## LLM / Vibe-Generated Code

Allowed only if:

- Wrapped behind one well-named step (preserves progressive disclosure)
- Project compiles / runs after every change
- Includes cancellation / cleanup when async or background
- Does not expand blast radius (side effects contained)

---

## Reject or Refactor

Flag any of the following for immediate rework — each violates progressive disclosure:

| Violation | Why it breaks disclosure |
|-----------|------------------------|
| Monolithic function (> ~40–50 lines) | No layers to disclose into — reader must parse everything at once |
| Orchestrator contains business logic | Layer 1 leaks layer-2 detail |
| Validation + IO + domain rules mixed | Multiple concerns at the same disclosure level |
| Deeply nested branching / error checks | Reader must hold too many conditions in working memory |
| Scattered / unclassified errors | Error semantics invisible at the top level |
| Uncontained side effects | Reader can't trust function names — hidden behaviour |