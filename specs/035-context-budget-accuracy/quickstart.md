# Quickstart: Validating budget/estimator accuracy

This guide proves the feature works end-to-end once implemented, against a disposable temp project. It mirrors what the Go tests for this feature assert.

## Prerequisites

- A built `misterspec` binary reflecting this feature's changes.
- A scratch project with a Spec that has several Chunks of clearly different sizes (some small, some individually larger than a small test budget), including at least one section containing a fenced code block large enough to matter, so the coherent-unit fallback has something real to split.

## Scenario 1 — The estimator is named in every response (User Story 3 / FR-001)

```sh
misterspec internal context --target SPEC-014 --mode manifest
```

**Pass condition**: `diagnostics.estimator == "default"` — present in every output mode (`--mode manifest`, `--mode package`, plain Markdown's own diagnostics-bearing JSON companion, per however each mode already surfaces diagnostics).

## Scenario 2 — Soft and hard limits are independent (User Story 2, Acceptance Scenario 1)

```sh
misterspec internal context --target SPEC-014 --budget 500 --hard-limit 50000
```

**Pass condition**: even though mandatory content alone comfortably exceeds the 500-token soft budget, `budget_exceeded: false` — the soft budget only shapes how much *optional* content is included, never flags mandatory content as "exceeded" by itself (research.md Decision 3).

```sh
misterspec internal context --target SPEC-014 --budget 50000 --hard-limit 100
```

**Pass condition**: if mandatory content exceeds 100 tokens, `budget_exceeded: true`, `overage = mandatory_tokens - 100` — driven by the hard limit, not the generous soft budget.

## Scenario 3 — Omitting `--hard-limit` never means unbounded (FR-006)

```sh
misterspec internal context --target SPEC-014
```

**Pass condition**: `diagnostics.hard_limit == 12000` (the resolved default) is present and finite even though `--hard-limit` was never passed.

## Scenario 4 — Backfill keeps trying smaller same-tier candidates (User Story 1, Acceptance Scenario 1–2)

```sh
misterspec internal context --target SPEC-014 --budget 300
```

**Pass condition**: with a tight soft budget, a small optional candidate that fits appears in `items` even when a larger, higher-scored candidate of the *same* tier was excluded first — `diagnostics.exclusions` lists the larger one with `reason: "did_not_fit_remaining_budget"`; the smaller one is not also listed as excluded (it was included). No candidate from a lower-priority tier appears in `items` while any higher-priority-tier candidate remains undecided in `diagnostics.exclusions`'s own absence from `items` (FR-009 — tier order is never violated).

## Scenario 5 — An oversized section is split, never truncated mid-code-block (User Story 1, Acceptance Scenario 3 / FR-011–FR-012)

```sh
misterspec internal context --target SPEC-014 --budget 150
```

**Pass condition**: for the one Chunk containing the large fenced code block, either (a) the block appears in `items` complete and un-truncated (its own coherent unit fit), or (b) the whole Chunk is excluded (`reason: "no_coherent_unit_fit"`) — never a response containing a fenced block whose closing fence is missing.

## Scenario 6 — `schema_version` signals the redefinition (FR contract)

```sh
misterspec internal context --target SPEC-014 --mode manifest
```

**Pass condition**: `schema_version == 2` (up from `1` in 033) in every response.
