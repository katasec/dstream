# Architecture Baseline Workstream

> Build a clear current-state baseline, identify mismatch risks, and turn findings into an implementation roadmap.

## Scope

- Document what DStream does today (not just intended behavior).
- Identify code/design/docs mismatches that can produce bugs or onboarding friction.
- Prioritize short-term stabilization and medium-term ecosystem enablement.

## Discover -> Catalogue -> Classify -> Propose

### Discover (Done)

- Confirmed core model is three-process stdin/stdout orchestration for `type = "providers"` tasks.
- Confirmed provider binaries can be resolved by local path or pulled via OCI reference.
- Confirmed ORAS pull + local cache path behavior for provider binaries.
- Confirmed command-envelope pattern (`command` + `config`) used for provider startup.

### Catalogue (In Progress)

- [x] Architecture narrative in [docs/design/design.md](../design/design.md)
- [ ] Protocol reference doc (wire examples, required fields, lifecycle semantics)
- [ ] Runtime sequence diagrams (run vs init/plan/status/destroy)
- [ ] Legacy path inventory (gRPC/go-plugin usage, docs, and removal risk)
- [ ] Historical snapshot/backfill requirements and options (full load + CDC handoff)

### Classify (Planned)

Candidate gap categories:

1. Product/documentation alignment gaps.
2. Runtime behavior ambiguity gaps.
3. Plugin authoring friction gaps.
4. Testing and compatibility assurance gaps.
5. Historical data bootstrap and incremental handoff gaps.

### Propose (Planned)

Phased roadmap proposal to follow after catalogue/classification.

## Initial Risk Register

1. Legacy and current execution modes coexist, which can confuse users and maintainers.
2. Protocol contract is implemented but not yet published as a versioned spec.
3. Lifecycle command routing semantics may not be intuitive for provider authors.
4. Ecosystem scalability depends on stronger authoring and compatibility tooling.
5. No first-class historical snapshot capability yet, which limits production onboarding for large existing datasets.

## Near-Term Deliverables

1. Protocol v1 spec doc with canonical examples.
2. Provider author quickstart and packaging checklist.
3. Compatibility test harness for provider startup and stream semantics.
4. Deprecation decision record for legacy gRPC path.
5. Snapshot/backfill design proposal covering:
	- dataset snapshot strategy,
	- CDC cutoff/handoff marker,
	- idempotency and dedupe expectations,
	- restart and replay behavior,
	- operational controls for partial reruns.

## Notes

- This workstream is documentation-first to reduce implementation churn.
- Follow-up implementation tasks should be split into small, test-backed changes.
