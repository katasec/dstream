# Code Health Audit

> Spoke of [code-style.md](code-style.md). Use this prompt to run a full project review.

Copy-paste the prompt below into a new session to get a prioritised task list. The agent will read the code-style guide and audit the codebase against it.

---

> **Prompt:**
>
> Read `code-style.md` and `code-review.md` — specifically the Governing Principle, Corollaries, Reject or Refactor table, and the full Code Review Checklist. Then audit the codebase:
>
> 1. **Scan all source files** in `src/` and `tests/` — check every function length, nesting depth, silent error handlers, duplicated patterns, stale comments, and unused imports.
> 2. **Check test coverage** — identify public methods with zero direct unit tests. Flag custom parsers/serializers without tests. Note any empty catch / error-swallowing blocks.
> 3. **Check outline readability** — for each file, is the intent clear in the first 15–20 lines? Flag any file where you must scroll past implementation detail to understand what it does.
> 4. **Check refactoring hygiene** — empty directories, namespace/package mismatches, orphan references, stale TODO comments.
> 5. **Output a prioritised task list** — group into: 🔴 Fix now (bugs, silent failures), 🟡 Fix soon (style violations, missing tests), 🟢 Nice to have (naming, minor cleanup). Include file path and line number for each finding.
>
> Do not fix anything — just produce the audit report.