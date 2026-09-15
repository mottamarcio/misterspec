# Phase 0 Research: Internal Context Command

All unknowns spec.md's own Assumptions section deliberately deferred to
planning are resolved below, including two apparent tensions between
spec.md's own illustrative wording and 015/016's already-shipped
behavior. No `NEEDS CLARIFICATION` markers remain.

## 1. Cache path

**Decision**: A fixed, internal constant path,
`<project root>/.misterspec/cache/context.db`, exactly as
`docs/context-engine-implementation.md` §10.2 illustrates — not a new
`.misterspec/config.yaml` field.

**Rationale**: The path never needs to vary between projects or between
invocations; adding a configuration field for it would be exactly the
kind of speculative knob Constitution Principle IV forbids. The path is
joined against `project.Detect`'s own already-validated root, so no new
path-traversal surface is introduced (Principle VIII).

**Alternatives considered**: A `--cache` flag letting the caller
override the path. Rejected — no current caller has a demonstrated need
to point at a different cache location, and the disposable index is
explicitly non-authoritative (deletable, rebuildable) regardless of
where it lives.

## 2. Transparent index readiness (FR-007, User Story 3)

**Decision**: Every invocation calls `index.Open(cachePath)` (creating
the file and/or recreating an incompatible schema, per 014's own
already-existing `ensureSchema`) followed by one `store.Sync(root,
cfg)` call (014's own already-existing incremental reconciliation)
before ever calling `Collect`. No new index-repair logic is written for
this feature.

**Rationale**: 014 already fully implements every piece User Story 3
asks for — schema-mismatch recreation (`ensureSchema`), incremental
new/changed/deleted reconciliation (`Sync`), and safe rebuildability
from nothing. Phase 8's own job is purely to *call* that existing
machinery before serving a request, not to reimplement any part of it —
the clearest possible application of Constitution Principle VI (DRY)
and Principle IV (YAGNI: no new capability where an existing one
already suffices).

**Alternatives considered**: Only calling `Sync` when a prior request
detects staleness some other way (e.g. a last-synced timestamp cached
in memory across invocations). Rejected — each CLI invocation is a
separate process with no persistent in-memory state to consult, and
`Sync`'s own fingerprint comparison already makes a call against an
up-to-date index cheap (skips every unchanged artifact) — there is
nothing to optimize away.

## 3. Package placement

**Decision**: One new file in each of two already-existing packages —
`internal/context/render.go` (the new `Render` function) and
`internal/cli/internalcmd/context.go` (the new `NewContextCmd`) — no
new subpackage.

**Rationale**: `Render` is pure, dependency-free computation over types
`contextengine` already owns (`Result`, `ResultItem`, `Tier`), exactly
matching 016's own precedent for adding `rank.go`/`budget.go` to the
same package rather than a new one. `NewContextCmd` is one more
`internal/cli/internalcmd` command following the exact shape every
prior command (`references.go`, `backlinks.go`, ...) already
establishes.

**Alternatives considered**: A new `internal/context/render` subpackage.
Rejected for the same reason 016 rejected a `rank` subpackage — no
independent concern exists here (no new dependency, no new I/O, no
alternate implementation ever anticipated).

## 4. `ErrUnsupportedIntent` sentinel

**Decision**: Export a new sentinel, `contextengine.ErrUnsupportedIntent
= errors.New(...)`, in `request.go`, and have `validateIntent` wrap it
(`fmt.Errorf("%w: %v", ErrUnsupportedIntent, i)`) instead of its current
bare `fmt.Errorf`.

**Rationale**: FR-002 requires the CLI to return "a clear, distinct
error identifying the unsupported intent" — `internalcmd.classify`
needs an `errors.Is`-matchable sentinel to map onto its own
`"unsupported_intent"` code, exactly the same pattern every other
`classify` case already follows (`operations.ErrEntityNotFound`,
`operations.ErrInvalidTarget`, ...). This is a minimal, additive change
to 015's own file — the unrecognized-intent condition itself, and its
message text, are unchanged; only its identity becomes checkable.

**Alternatives considered**: Having `internalcmd` re-validate the
intent string itself against a duplicated set of recognized values
before ever calling `Collect`, so it can produce its own error without
touching `contextengine`. Rejected — this would duplicate
`recognizedIntents` in two packages (violating Principle VI's DRY
rule) and risk the two lists silently drifting apart.

## 5. `Tier.String()`

**Decision**: Add one method, `func (t Tier) String() string`,
returning `"mandatory"`, `"structural"`, `"semantic"`, `"text"`, or
`"second_hop"`, to `candidate.go` (015's own file, alongside the `Tier`
type itself).

**Rationale**: Both the JSON `tier` field (FR-006, matching §21's own
illustrative `"tier": "target"`-style output) and the Markdown
`Render` output (FR-010, grouping "by tier/origin") need the exact same
tier-name mapping. Centralizing it once on `Tier` itself, next to where
`Tier`'s own constants are already documented, avoids writing the same
five-case switch twice (Principle VI, DRY) — the single failure mode
015/016 were careful to avoid by keeping `Tier` an unexported-detail
`int` until now.

**Alternatives considered**: A private `tierName(t Tier) string` helper
duplicated in both `internalcmd/context.go` and `context/render.go`.
Rejected — the exact DRY violation this method exists to prevent.

## 6. "Invalid budget" (FR-005) vs. 016's own negative/zero semantics

**Decision**: "Invalid budget" means a `--budget` flag value that fails
to parse as an integer at all (e.g. `--budget=abc`) — a Cobra-level
flag-parsing failure `root.Execute()` already converts into a
`invalid_argument` JSON error automatically (`internal/cli/root.go`),
with no new code needed. A *parseable* budget, including zero or a
negative number, is **not** rejected — it is passed through as
`Request.Budget`'s own real, explicit value, exactly as 016's own
`ApplyBudget` already defines (mandatory content still returned in
full, `BudgetExceeded` flagged, `Overage` reported).

**Rationale**: spec.md's own FR-005 phrasing ("an invalid budget, for
example, negative") textually suggests rejecting negative budgets, but
spec.md's own Assumptions section is explicit that this feature
"introduces no new ranking or budgeting logic of its own" and that
016's Context Result is exposed, not re-decided. 016's own spec (edge
cases, FR-006/FR-007/FR-008) already deliberately made zero/negative
budgets *meaningful, valid* input — the exact mechanism User Story 3 of
016 exists to test. Re-rejecting them here would silently contradict
already-shipped, already-tested behavior one layer up, which this
feature has no mandate to change. The one genuinely invalid case left —
a value that isn't a number at all — has no reasonable numeric meaning
and is correctly rejected; that is what FR-005 protects against in
practice.

**Alternatives considered**: Rejecting budgets below some new,
CLI-specific minimum (e.g. 0). Rejected — this would be new policy
invented at the CLI boundary with no grounding in 016's own
specification, and would make the same logical request behave
differently depending on whether it arrived via direct Go API call
(016) or through the CLI (this feature) — an inconsistency this
project's own layering discipline exists to prevent.

## 7. `--render` stays inside the JSON envelope

**Decision**: `--render` adds one additional string field
(`context.rendered`) to the same JSON success envelope — it never
switches the command's output to raw Markdown on stdout.

**Rationale**: Constitution Principle IX requires internal commands to
"default to structured JSON with no decorative/ANSI output" — a hard,
non-negotiable requirement that applies regardless of `--render`.
`docs/context-engine-implementation.md` §21.1 itself anticipates this
exact resolution: its own closing line reads "JSON should remain the
canonical machine interface," even while its illustrative example
shows bare Markdown. Nesting the rendered pack as one more field
satisfies both documents at once — machine-readable envelope always
(this project's Constitution), human/agent-readable Markdown pack
available on request (the source document's own Phase 8 scope) —
without a conflict needing escalation under Constraint #13.

**Alternatives considered**: A second, separate output mode where
`--render` suppresses the JSON envelope entirely and prints only
Markdown. Rejected — directly violates Principle IX, and would make
scripted callers branch on a flag just to know which parser to use.

## 8. `Task` stays untyped free text

**Decision**: The CLI's `--task` flag is passed through to
`Request.Task` completely unvalidated — exactly the semantics 015
already defined for that field ("the current task's own text, if any").
No artifact-existence check against the target's own Spec is added.

**Rationale**: spec.md's Edge Cases section describes "a task ID
supplied that does not belong to the target's own Spec, or does not
exist at all" as something to report as invalid input — but 015's own
already-shipped `Request.Task` was deliberately specified as free text
used only as a Tier 4 fallback query, never resolved or validated as an
artifact reference (015's own `request.go` doc comment, verbatim:
"Also used as Tier 4's own search query when Query is empty"). Adding a
new existence/ownership check here would be new validation logic this
feature's own Assumptions section explicitly rules out ("this feature
introduces no new indexing or synchronization logic of its own");
by extension, and per Constitution Principle IV, it also introduces no
new *request-validation* logic beyond what orchestrating 015/016
already requires. The common, expected case — a real Task ID string —
still flows through and contributes to retrieval exactly as before;
what changes is only that this feature does not add a policing layer
015 itself never had.

**Alternatives considered**: Resolving `--task` via
`operations.Inspect` before building the request, failing the command
if it doesn't resolve to a Task artifact under the target's own Spec.
Rejected as new scope — this would be new semantic policy layered onto
015's already-settled contract, not orchestration of existing
capability, and no current caller has demonstrated a need for it.

## 9. Index-sync failure classification

**Decision**: An error returned by `index.Open` or `store.Sync` is
passed to the existing `WriteError`/`classify` machinery unchanged;
`classify`'s existing default case (`"unexpected_failure"`, exit code
1) covers it. No new sentinel or classify case is added for this
condition.

**Rationale**: `docs/context-engine-implementation.md` §24 lists "index
cannot be opened" / "index synchronization failed" as illustrative
error classes, but 014 itself exposes no sentinel error type for either
condition today (both surface as plain wrapped I/O/SQL errors) — there
is nothing concrete yet to `errors.Is`-match against. Inventing a new
classify case with no corresponding sentinel to test it against would
be speculative surface (Principle IV); the existing default classify
case already gives a well-formed, distinct-from-success JSON error
(FR-006, User Story 2's own Acceptance Scenario 5), which is what this
feature's own requirements actually demand.

**Alternatives considered**: Adding new sentinels to `internal/context/
index` for "open failed" / "sync failed" specifically so this feature
can classify them distinctly. Rejected as out-of-scope surgery on 014's
already-shipped, already-tested package for a case with no demonstrated
caller need to distinguish it from any other unexpected failure.

## 10. Command shape: one command, five inputs

**Decision**: `misterspec internal context <id> [--intent I]
[--task T] [--query Q] [--budget N] [--render]`, matching
`docs/context-engine-implementation.md` §21's own illustrative
invocation shape and this project's existing `<verb> <id> [flags]`
convention (`inspect <id>`, `references <id>`, `backlinks <id>`).

**Rationale**: One positional argument (the target, required, matching
every existing single-target internal command) plus optional flags for
every other `Request` field keeps the surface identical in shape to
014-established commands, requiring no new CLI convention.

**Alternatives considered**: A JSON request body read from stdin
instead of flags. Rejected — no existing internal command uses stdin
for its request shape; flags keep this command directly consistent
with `inspect`/`references`/`backlinks`.
