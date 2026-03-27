# Code Style & Design Principles

> Spoke of [AGENTS.md](../../AGENTS.md). Load when writing code.

---

## The Governing Principle: Progressive Disclosure

**Progressive disclosure** is the single organising idea behind every convention in this guide.

Code reveals intent in layers:

| Layer | Shows | Reader action |
|-------|-------|---------------|
| **Top** | *What* happens — intent, flow, named steps | Scan to orient |
| **One level deeper** | *How* each step works — implementation | Drill into the one step you care about |

You should **never need to go deeper than one step** to understand any single concern. If you do, the code has failed to disclose progressively — refactor.

> **Mantra:** Code must read like a trustworthy outline — every decision is visible at the top or exactly one small, composable step away.

Everything below is a **corollary** of this principle.

---

## Corollaries

### 1. Outline-First Programming

The top layer of any file, class, or function must disclose *what*, not *how*.

**15–20 line rule**: If the intent and high-level flow aren't clear in the first 15–20 lines, the code violates progressive disclosure — refactor.

### 2. Small, Composable Functions (SRP / Unix Philosophy)

Small single-responsibility functions *create the layers* for progressive disclosure. Each function is one named step. The caller reads step names (layer 1); drilling into any step reveals implementation (layer 2). Without small functions there are no layers — just a wall of detail.

### 3. Railway-Oriented Orchestrators

Pipeline composition (Result chaining via Bind / Map / Tap) keeps orchestrators at layer 1 — a readable sequence of named steps that short-circuits on failure. The orchestrator *delegates*; it never mixes validation, business logic, and IO. Orchestrators return Result types — never throw/panic for domain flow.

### 4. Each Layer Owns Exactly Its Concerns

Every validation lives in exactly one layer — the lowest that can enforce it. Higher layers assume lower layers did their job. Duplicate guards break disclosure by scattering the same concern across layers. Before adding a check: *"Does a lower layer already guarantee this?"* If yes, don't duplicate it.

### 5. Configuration Is Data, Not Code

Runtime data (templates, config values, feature flags) lives in configuration, not in the binary. This keeps the code layer about *logic and flow* — its proper disclosure level. If you're writing a custom parser to extract data baked into the binary, the data is in the wrong place.

### 6. Top-Down Method Ordering

Order methods by invocation order: public entry points first, then the methods they call, then the methods *those* call. The file reads like an outline — high-level flow at the top, supporting detail below. This is progressive disclosure applied to file layout.

### 7. Zero-Warnings Build Policy

The build must complete with **zero warnings**. Treat warnings as errors — unused variables, nullable issues, unhandled cases. Warnings are deferred bugs. Suppressing them (`#pragma`, `//nolint`, etc.) requires explicit approval.

---

## Structural Rules

- Side effects (DB, network, file, external) isolated in clearly named functions.
- Validation and auth at the edge (first pipeline steps).
- Multiple orchestration levels allowed if the outline stays clear at each level.
- Deep nesting, mixed concerns, scroll novels → **banned**.
- Prefer single-line method signatures; only break for very long parameter lists.
- Prefer explicit block syntax (`{ return ...; }`) over expression-bodied (`=>`) for multi-line bodies.

---

## Reference Spokes

| Topic | When to load | Detail |
|-------|-------------|--------|
| Language Examples | Implementing patterns (Go, C#) | [code-style-examples.md](code-style-examples.md) |
| Code Review | Reviewing or accepting changes | [code-review.md](code-review.md) |
| Code Audit | Running a full project health check | [code-audit.md](code-audit.md) |
