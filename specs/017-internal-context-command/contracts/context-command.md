# Phase 8 Contract: Internal Context Command

**Reconciled against the actual implementation (tasks.md T007-T015)** —
two drifts corrected below (`--budget`'s flag type; `Tier.String()`'s
actual file). Everything else matches what was planned, verbatim.

## CLI surface

```text
misterspec internal context <id> \
    [--intent planning|tasks|implementation|validation|analysis] \
    [--task <text>] \
    [--query <text>] \
    [--budget <int>] \
    [--render] \
    [--dir <path>]
```

- `<id>` — required, exactly one positional argument. Any entity type
  `contextengine.Collect` accepts (Program, Feature, Spec, Knowledge,
  Learning — the same five types `operations.References` already
  covers).
- `--intent` — optional. Omitted or `""` means no intent preference. Any
  value other than the five recognized ones is rejected
  (`unsupported_intent`).
- `--task` — optional free text, passed through unvalidated
  (research.md #8).
- `--query` — optional free text.
- `--budget` — optional, declared as a **string** flag (corrected during
  implementation from the originally planned `int` — see research.md
  #6's implementation-time amendment in tasks.md T007), manually parsed
  via `strconv.Atoi` inside `RunE`. Omitted means "use 016's own
  `DefaultBudget`." A non-numeric value returns `invalid_argument`
  directly from `RunE` (not a Cobra-level flag-parse failure — that
  fallback in `internal/cli/root.go` only fires for a full
  `root.Execute()` invocation, not a leaf command's own `Execute()`, so
  this command validates itself). A parseable zero or negative value is
  accepted and passed through (research.md #6).
- `--render` — optional boolean flag. When set, the response's
  `context.rendered` field carries a Markdown context pack; omitted or
  false, `context.rendered` is `null`.
- `--dir` — optional, defaults to `"."`, matching every other internal
  command.

## Go API surface (extended)

```go
package contextengine

// request.go
var ErrUnsupportedIntent = errors.New("contextengine: unsupported intent")

// result.go (corrected during implementation — Tier is declared in
// result.go, not candidate.go as originally planned)
func (t Tier) String() string

// render.go
func Render(req Request, result Result) string
```

No other exported symbol in `internal/context` changes shape.

## JSON envelope

Success:

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
      {"path": "...", "heading": "...", "tier": "mandatory", "reasons": ["target"], "score": 100, "tokens": 640}
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

Failure (unchanged envelope shape every internal command already uses):

```json
{"ok": false, "error": {"code": "entity_not_found", "message": "..."}}
{"ok": false, "error": {"code": "entity_ambiguous", "message": "..."}}
{"ok": false, "error": {"code": "unsupported_intent", "message": "..."}}
{"ok": false, "error": {"code": "invalid_argument", "message": "..."}}
{"ok": false, "error": {"code": "project_not_initialized", "message": "..."}}
```

## Error Code table (this feature's additions only)

| Condition | Code | Exit Code | Source |
|---|---|---|---|
| `<id>` does not resolve to exactly one existing artifact | `entity_not_found` | 3 | existing (`operations.ErrEntityNotFound`, unchanged) |
| `<id>` resolves ambiguously | `entity_ambiguous` | 3 | existing (`operations.ErrEntityAmbiguous`, unchanged) |
| `<id>` is not one of the five referenceable entity types | `invalid_target` | 2 | existing (`operations.ErrInvalidTarget`, unchanged) |
| `--intent` is not `""` or one of the five recognized values | `unsupported_intent` | 2 | **new** (`contextengine.ErrUnsupportedIntent`) |
| `--budget` is not a valid integer | `invalid_argument` | 2 | `RunE`'s own `strconv.Atoi` check (corrected from the originally planned Cobra-level fallback — see above) |
| No project detected at `--dir` | `project_not_initialized` | 6 | existing (`project.ErrNotInitialized`, unchanged) |
| Index cannot be opened or synchronized | `unexpected_failure` | 1 | existing default case (research.md #9) |

## Behavioral guarantees carried over from 015/016 (not re-tested here, only orchestrated)

- Tier always dominates Score in `items[]`'s own order (016 FR-002/FR-003).
- Mandatory content (Tier 0/1) always fully present in `items[]`
  regardless of `--budget` (016 FR-006).
- `budget_exceeded`/`overage` report accurately when mandatory content
  alone exceeds the resolved budget (016 FR-007).
- `diagnostics` values are internally consistent
  (`tokens_available - tokens_excluded == tokens_selected`) (016 FR-010).
- Zero project artifacts, reference-graph data, or search-index *source*
  data are ever modified — the disposable cache database is the only
  thing this command ever writes to (this feature's own FR-008).
