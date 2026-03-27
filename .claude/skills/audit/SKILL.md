---
name: audit
description: Run a full code health audit against the code style guide. Use for periodic project health checks or when onboarding to a new codebase.
---

# Code Health Audit

Run a full project review against the principles in [code-style.md](../../../docs/agents/code-style.md) and the checklists in [code-review.md](../../../docs/agents/code-review.md).

## Steps

1. **Read** `docs/agents/code-style.md` (governing principle + corollaries) and `docs/agents/code-review.md` (full checklist + reject/refactor table).

2. **Scan all source files** in `src/` and `tests/`:
   - Function length (flag >40–50 lines)
   - Nesting depth (flag >2 levels)
   - Silent error handlers (empty catch blocks, swallowed errors)
   - Duplicated patterns (check for existing helpers)
   - Stale comments (`// TODO`, orphan references)
   - Unused imports

3. **Check test coverage**:
   - Public methods with zero direct unit tests
   - Custom parsers/serializers without tests
   - Empty catch / error-swallowing blocks

4. **Check outline readability**:
   - For each file: is intent clear in the first 15–20 lines?
   - Flag files where you must scroll past implementation to understand purpose

5. **Check refactoring hygiene**:
   - Empty directories
   - Namespace/package mismatches
   - Orphan references
   - Stale TODO comments

6. **Output a prioritised task list** grouped by severity:

## Output Format

```
## Code Health Audit — [date]

### 🔴 Fix Now (bugs, silent failures)
- [file:line] Description of issue

### 🟡 Fix Soon (style violations, missing tests)
- [file:line] Description of issue

### 🟢 Nice to Have (naming, minor cleanup)
- [file:line] Description of issue

### Summary
- Files scanned: N
- Issues found: N (🔴 X, 🟡 Y, 🟢 Z)
```

**Do not fix anything** — just produce the audit report.