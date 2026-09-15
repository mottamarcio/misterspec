# Phase 1 Data Model: Internal Context Command

Two small Go additions to `internal/context` (015/016's own package),
one new function in the same package, one new file in
`internal/cli/internalcmd`, and one new CLI-boundary JSON shape. No new
persisted schema — the disposable index's own schema (014) is reused
unchanged.

## `contextengine.ErrUnsupportedIntent` (new sentinel, `request.go`)

```go
// ErrUnsupportedIntent is returned (wrapped) by validateIntent when a
// Request's Intent is neither "" nor one of the recognized values
// (FR-002, research.md #4) — an internalcmd.classify-matchable
// sentinel, not just an error string.
var ErrUnsupportedIntent = errors.New("contextengine: unsupported intent")
```

`validateIntent` is modified to return
`fmt.Errorf("%w: %v", ErrUnsupportedIntent, i)` instead of its current
bare `fmt.Errorf(...)` — same condition, same message shape, now
`errors.Is`-matchable.

## `contextengine.Tier.String()` (new method, `candidate.go`)

```go
// String returns t's own stable, lowercase label — used both by
// Render's own Markdown section headings and by NewContextCmd's own
// JSON "tier" field (research.md #5). Never changes across releases
// without a deliberate, reviewed decision, since callers may match on
// it.
func (t Tier) String() string {
    switch t {
    case TierMandatory:
        return "mandatory"
    case TierStructural:
        return "structural"
    case TierSemantic:
        return "semantic"
    case TierText:
        return "text"
    case TierSecondHop:
        return "second_hop"
    default:
        return "unknown"
    }
}
```

## `contextengine.Render` (new function, `render.go`)

```go
// Render converts result into a well-formed, readable Markdown context
// pack (FR-010): one "## <Tier label>" heading per Tier present among
// result.Items (in ascending Tier order, matching result's own already-
// established ordering — Render performs no re-sorting of its own),
// followed by one subsection per item naming its Path/Heading, its
// Content, and its own Reasons/Tokens. Contains every item present in
// result — Render never adds, drops, reorders, or re-scores anything
// (FR-011); it is a pure, read-only presentation of an already-computed
// Result. req is used only for a short header line naming the Target
// and Intent — it plays no role in item selection.
func Render(req Request, result Result) string
```

Grouping key: `minTier(item.Reasons)` — the same "effective Tier" 015's
own `mergeAndSort`/016's own `Rank` already use, so `Render`'s own
grouping matches the same Tier boundary FR-002/FR-003 already
established, computed once via the already-exported `Tier` value each
`ResultItem`'s own `Reasons` collectively carry (no new tier-detection
logic; reuses the existing unexported `minTier` helper already present
in `candidate.go`).

## Context Command Request (CLI-boundary struct, `internalcmd/context.go`)

Not a new domain type — purely the flag-parsing shape `NewContextCmd`'s
`RunE` assembles into a `contextengine.Request` before calling
`Collect`/`Rank`/`ApplyBudget`:

| Flag | Go type | Maps to | Default when omitted |
|---|---|---|---|
| `<id>` (positional, required) | `string` | `Request.Target` | n/a — `cobra.ExactArgs(1)` |
| `--intent` | `string` | `Request.Intent` (cast) | `""` (no intent preference) |
| `--task` | `string` | `Request.Task` | `""` |
| `--query` | `string` | `Request.Query` | `""` |
| `--budget` | `int`, tracked via `cmd.Flags().Changed("budget")` | `Request.Budget *int` | `nil` (016's own `DefaultBudget` resolved downstream) |
| `--render` | `bool` | (not part of `Request` — controls whether `Render` is called after `ApplyBudget`) | `false` |
| `--dir` | `string` | `project.Detect`'s own root, matching every existing internal command | `"."` |

Validation, in order, each returning through the existing
`WriteError`/`classify` machinery on failure:

1. `project.Detect(dir)` — existing project-resolution errors
   (`project.ErrNotInitialized`) unchanged.
2. `index.Open(cachePath)` + `store.Sync(root, cfg)` — establishes
   Index Readiness (FR-007); any error classifies as
   `unexpected_failure` (research.md #9).
3. `contextengine.Collect(root, cfg, store, req)` — surfaces
   `operations.ErrEntityNotFound` / `ErrEntityAmbiguous` / `ErrInvalidTarget`
   (FR-004) and the new `contextengine.ErrUnsupportedIntent` (FR-002),
   all already classified or newly classified per research.md #4.
4. A `--budget` value that fails Cobra's own int parsing is caught
   before `RunE` ever runs, automatically becoming an
   `invalid_argument` error via `internal/cli/root.go`'s existing
   Cobra-level fallback (FR-005, research.md #6) — no explicit check
   needed in `RunE` itself.

## Context Command Result (CLI-boundary JSON shape)

Success (`ok: true`):

```json
{
  "ok": true,
  "context": {
    "target": "SPEC-014",
    "intent": "implementation",
    "budget": 6000,
    "estimated_tokens": 4382,
    "budget_exceeded": false,
    "overage": 0,
    "items": [
      {
        "path": "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/spec.md",
        "heading": "Requirements",
        "tier": "mandatory",
        "reasons": ["target"],
        "score": 100,
        "tokens": 640
      }
    ],
    "diagnostics": {
      "candidates_considered": 126,
      "items_selected": 14,
      "tokens_available": 48230,
      "tokens_selected": 4382,
      "tokens_excluded": 43848,
      "reduction_percent": 90.9
    },
    "rendered": null
  }
}
```

Field derivation — every field is copied straight from 016's own
`Result`/`ResultItem`/`Diagnostics`, no new computation:

- `target`, `intent`, `budget` — echoed from the resolved `Request` (the
  budget actually used: `*req.Budget` or `contextengine.DefaultBudget`).
- `estimated_tokens` = `Diagnostics.TokensSelected`.
- `budget_exceeded` = `Result.BudgetExceeded`; `overage` = `Result.Overage`.
- `items[]` = `Result.Items`, one entry per `ResultItem`: `path` =
  `Path`, `heading` = `Heading`, `tier` = `Tier.String()` of
  `minTier(Reasons)`, `reasons` = each `Reason.Relation` string
  (deduplicated already by 015's own merge), `score` = `Score`,
  `tokens` = `Tokens`.
- `diagnostics` = `Result.Diagnostics`, field-for-field, snake_cased.
- `rendered` — `null` when `--render` was not passed; the
  `contextengine.Render` output (a Markdown string) otherwise (FR-010,
  research.md #7).

Failure (`ok: false`) uses the existing, unchanged envelope every
internal command already produces:

```json
{"ok": false, "error": {"code": "entity_not_found", "message": "..."}}
```

New `classify` case (`errors.go`):

```go
case errors.Is(err, contextengine.ErrUnsupportedIntent):
    return "unsupported_intent", 2
```

(Exit code `2`, matching the existing `invalid_target` /
`unsupported_type` / `invalid_argument` family of CLI-boundary/
input-shape problems.)

## Index Readiness (conceptual precondition, not a new type)

Not a persisted or returned entity — the state `NewContextCmd`'s
`RunE` establishes, in order, before calling `Collect`:

1. `index.Open(cachePath)` — creates the cache directory/file if
   missing; recreates the schema from scratch if `PRAGMA user_version`
   doesn't match (014's own existing `ensureSchema`, unmodified).
2. `store.Sync(root, cfg)` — reconciles new/changed/deleted artifacts
   against the now-current schema (014's own existing `Sync`,
   unmodified); a database that was just recreated behaves identically
   to an empty one, so every artifact indexes as new.

No new struct models this — it is pure orchestration, matching
research.md #2's decision to add zero new index-repair logic.
