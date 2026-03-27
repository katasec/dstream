# Code Style — Language Examples

> Spoke of [code-style.md](code-style.md). Reference when implementing patterns described in the main guide.

---

## Orchestrator Pipelines

The orchestrator sits at disclosure **layer 1** — a sequence of named steps. Each step is a small function (layer 2) that the reader drills into only when needed. The pipeline short-circuits on the first failure.

### Go

Use a `Result[T, E]` type (e.g. `github.com/samber/mo`, or a lightweight custom impl).

```go
func ProcessOrder(ctx context.Context, req OrderRequest) Result[OrderResult, AppError] {
    return validateRequest(req).
        Bind(authorizeAndResolveTenant).
        Bind(fetchAndLockResources).
        Bind(executeBusinessLogic).
        Bind(persistChanges).
        Bind(emitEvents).
        Tap(cleanupOnSuccess)
}
```

**Rules:**
- No `panic` / early `log.Fatal` in business orchestrators.
- Use `Bind`, `Map`, `Tap` / equivalents for composition.
- Each step is a standalone function with a single responsibility.

### C\#

Use `Result<T, E>`, `ErrorOr`, `FluentResults`, or `OneOf`.

```csharp
public async Task<Result<OrderResult, AppError>> ProcessOrderAsync(OrderRequest req)
{
    return await ValidateRequest(req)
        .BindAsync(AuthorizeAndResolveTenant)
        .BindAsync(FetchAndLockResources)
        .BindAsync(ExecuteBusinessLogic)
        .BindAsync(PersistChanges)
        .BindAsync(EmitEvents)
        .Tap(CleanupOnSuccess);
}
```

**Rules:**
- Async flows use `BindAsync` / `MapAsync` / `Tap`.
- Domain/business exceptions → wrapped in Result (no throws from orchestrators).
- Prefer keeping method signatures on a single line; only break for very long parameter lists.

---

## Structural Patterns

### Comment Banners for Step Separation

When a function has logical phases that aren't yet extracted into separate functions, comment banners make the outline scannable:

```go
func handleRequest(ctx context.Context, req Request) error {
    // ─── 1. Validate ────────────────────────────────
    if err := validateInput(req); err != nil {
        return err
    }

    // ─── 2. Authorize ───────────────────────────────
    user, err := authorize(ctx, req.Token)
    if err != nil {
        return err
    }

    // ─── 3. Execute ─────────────────────────────────
    return process(ctx, user, req)
}
```

```csharp
public async Task<Result<Response>> HandleAsync(Request req, CancellationToken ct)
{
    // ─── 1. Validate ────────────────────────────────
    var validated = Validate(req);
    if (validated.IsFailure) return validated.Error;

    // ─── 2. Authorize ───────────────────────────────
    var user = await Authorize(req.Token, ct);
    if (user.IsFailure) return user.Error;

    // ─── 3. Execute ─────────────────────────────────
    return await Process(user.Value, validated.Value, ct);
}
```

Banners are a stepping stone. Once the steps are stable, extract each into a named function and compose via pipeline.

### Block Syntax over Expression-Bodied (C#)

Prefer explicit blocks for multi-line method bodies — the braces make the outline scannable:

```csharp
// ✅ Preferred — clear block structure
public Task<Result<Order>> GetOrderAsync(int id, CancellationToken ct)
{
    return _repository
        .FindAsync(id, ct)
        .BindAsync(ValidateOrder)
        .BindAsync(EnrichWithDetails);
}

// ❌ Avoid — expression body obscures the outline
public Task<Result<Order>> GetOrderAsync(int id, CancellationToken ct)
    => _repository
        .FindAsync(id, ct)
        .BindAsync(ValidateOrder)
        .BindAsync(EnrichWithDetails);
```

### Nesting Depth — Max 2 Levels

Refactor to early returns or extract methods:

```go
// ❌ Too deep — 3+ levels of nesting
func process(items []Item) error {
    for _, item := range items {
        if item.IsActive {
            if item.HasPermission {
                if err := handle(item); err != nil {
                    return err
                }
            }
        }
    }
    return nil
}

// ✅ Flattened with early continues
func process(items []Item) error {
    for _, item := range items {
        if !item.IsActive || !item.HasPermission {
            continue
        }
        if err := handle(item); err != nil {
            return err
        }
    }
    return nil
}
```

---

## Layer Ownership — Validation Example

Each layer validates only what *it* can enforce. No re-guarding what a lower layer already guarantees.

### Go

```go
// Layer 1: Schema / types — enforced by the type system
type CreateUserRequest struct {
    Name  string `validate:"required,min=1"`
    Email string `validate:"required,email"`
    Age   int    `validate:"required,gte=0,lte=150"`
}

// Layer 2: Input validation — cross-field checks the schema can't express
func validateCreateUser(req CreateUserRequest) error {
    // Don't re-check that Name is non-empty — the schema guarantees it.
    // Do check cross-field invariants:
    if req.Age < 13 && req.Email != "" {
        return errors.New("users under 13 cannot have an email on file")
    }
    return nil
}

// Layer 3: Business logic — runtime state
func createUser(ctx context.Context, req CreateUserRequest) Result[User, AppError] {
    // Don't re-validate fields. Do check runtime state:
    existing, _ := repo.FindByEmail(ctx, req.Email)
    if existing != nil {
        return Fail[User](Conflict("email already registered"))
    }
    return Ok(repo.Create(ctx, req))
}
```

### C\#

```csharp
// Layer 1: Schema / types — enforced by model binding + attributes
public record CreateUserRequest(
    [Required, MinLength(1)] string Name,
    [Required, EmailAddress] string Email,
    [Required, Range(0, 150)] int Age
);

// Layer 2: Input validation — cross-field checks
public static Result<CreateUserRequest> Validate(CreateUserRequest req)
{
    // Don't re-check that Name is non-empty — the model binder guarantees it.
    if (req.Age < 13 && !string.IsNullOrEmpty(req.Email))
        return Result.Fail("Users under 13 cannot have an email on file");

    return Result.Ok(req);
}

// Layer 3: Business logic — runtime state
public async Task<Result<User>> CreateUserAsync(CreateUserRequest req, CancellationToken ct)
{
    // Don't re-validate fields. Do check runtime state:
    var existing = await _repo.FindByEmailAsync(req.Email, ct);
    if (existing is not null)
        return Result.Fail<User>("Email already registered");

    return Result.Ok(await _repo.CreateAsync(req, ct));
}
```