---
name: review
description: Run the code review checklist against recent changes. Use when reviewing code, accepting a PR, or after finishing a feature to verify quality before committing.
---

# Code Review

Run the full code review checklist from [code-review.md](../../../docs/agents/code-review.md) against the files changed in this session (or the files the user specifies).

## Steps

1. **Identify scope**: Use `git diff --name-only` (or the user-specified files) to get the list of changed files.
2. **Read each changed file** and evaluate against every item in the checklist:
   - Structure & Disclosure (outline-first, function length, duplication, silent catches, nesting, embedded data)
   - Testing (happy path tests, parser tests, edge cases, existing tests pass)
   - Naming & Intent (descriptive names, stale comments)
   - Dependencies & Side Effects (isolated side effects, justified deps, clean imports)
   - Refactoring-Specific (cleanup after moves, namespace matches, regression tests)
3. **Check LLM/vibe-generated code rules** if any code was AI-generated.
4. **Check the Reject or Refactor table** — flag any violations.
5. **Output a summary** with pass/fail per checklist item, grouped by file. Use ✅ / ❌ markers.

## Output Format

```
## Review: [file path]
✅ Outline-first: intent clear in first 15 lines
✅ Function length: all under 50 lines
❌ Silent catch: empty catch block at line 42
...

## Summary
Passed: 14/16 checks
Action needed: 2 items (see ❌ above)
```

If all checks pass, say so clearly. If any fail, list the specific file, line, and what to fix.