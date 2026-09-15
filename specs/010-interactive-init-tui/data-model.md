# Phase 1 Data Model: Interactive Init TUI

Entities extracted from `spec.md`'s Key Entities plus
`docs/architecture-specification.md` §37's own `Screen` model. This
feature adds one new package, `internal/tui`, and one additive field to
`internal/bootstrap.InspectResult`.

## `internal/bootstrap.InspectResult` (extended, additive)

| Field | Type | Notes |
|---|---|---|
| `Empty` | bool | NEW. `true` when `!Initialized` and `targetDir` contains no entries (or does not exist yet); `false` otherwise (including whenever `Initialized` is `true` — an initialized project is never reported empty). research.md. |

**Validation rules**: Existing fields (`Initialized`, `ProjectRoot`,
`InstalledAgent`, `AgentInstalled`) and `Inspect`'s exported signature
are unchanged — every 007/008/009 caller keeps compiling and passing
unmodified (named regression gate).

## `internal/tui.Screen`

`docs/architecture-specification.md` §37's own state model, verbatim:

```go
type Screen int

const (
    ScreenInspect Screen = iota
    ScreenNonEmptyWarning
    ScreenAgentSelection
    ScreenPreview
    ScreenInstalling
    ScreenSuccess
    ScreenError
)
```

## `internal/tui.Model`

The Bubble Tea model — UI state only, per §37's own rule ("The Bubble
Tea model should contain UI state. It should not contain filesystem
business rules.").

| Field | Type | Notes |
|---|---|---|
| `screen` | `Screen` | Current screen. |
| `targetDir` | string | The directory being bootstrapped (`--dir`'s value, default `.`). |
| `registry` | `*agents.Registry` | Passed in at construction — `builtin.Default()` from the real CLI, a fixture in tests. |
| `skills` | `fs.FS` | Passed in at construction — `kit.SkillsFS` from the real CLI, a fixture in tests. |
| `inspect` | `bootstrap.InspectResult` | Populated once `ScreenInspect`'s command completes. |
| `inspectErr` | error | Set only for a genuine `Inspect` failure (e.g. `project.ErrInvalidConfiguration`) — distinct from a normal "not initialized" or "non-empty" result. |
| `agents` | `[]agents.Adapter` | `registry.List()`'s result, cached once at `ScreenAgentSelection`. |
| `cursor` | int | Selected index within the current screen's choice list (agent list, or a warning/preview's Yes/No choice). |
| `selectedAgent` | `agents.Adapter` | Set once the user confirms a choice on `ScreenAgentSelection`. |
| `preview` | `previewContent` | Assembled once, entering `ScreenPreview` (research.md's read-only composition). |
| `outcome` | `bootstrap.BootstrapOutcome` | Populated on a successful `ScreenInstalling` → `ScreenSuccess` transition. |
| `installErr` | error | Populated on a failed `ScreenInstalling` → `ScreenError` transition. |
| `quitting` | bool | Set when the user cancels at any screen (FR-009) — the next `Update` returns `tea.Quit`. |

### `previewContent` (unexported, internal to `internal/tui`)

| Field | Type | Notes |
|---|---|---|
| `configPath` | string | The fixed `.misterspec/config.yaml` path (FR-004). |
| `templates` | `[]installer.Resource` | `installer.List()` — the kit resources that will install. |
| `skillResources` | `[]installer.Resource` | `installer.ListFS(skills, ".", "skill")` — the selected agent's Skills that will install. |
| `agentTargetPath` | string | `selectedAgent.TargetPath()` — where Skills land. |

**Validation rules**: `previewContent` is built once and never
recomputed after the user confirms — what `ScreenPreview` shows is
exactly what `bootstrap.Bootstrap` (called identically, with the same
`targetDir`/`selectedAgent`/`registry`/`skills`) then performs (FR-005),
since both read from the same already-proven, deterministic sources.

## State / Flow Summary (§36's transaction model, concretely)

```text
ScreenInspect                                                    [US1, US2]
  bootstrap.Inspect(targetDir)
    err != nil (genuine failure)      → ScreenError
    Initialized || !Empty             → ScreenNonEmptyWarning     [US2]
    else                              → ScreenAgentSelection

ScreenNonEmptyWarning                                             [US2]
  user selects "Cancel"               → quitting = true (FR-009)
  user selects "Continue anyway"      → ScreenAgentSelection

ScreenAgentSelection                                              [US1]
  registry.List() → agents (cached once)
  user picks one, confirms            → build preview → ScreenPreview
  user cancels                        → quitting = true

ScreenPreview                                                     [US1]
  user confirms                       → ScreenInstalling
  user declines                       → quitting = true (FR-005: nothing written)

ScreenInstalling (transient — runs bootstrap.Bootstrap as a tea.Cmd)  [US1]
  success  → outcome populated  → ScreenSuccess
  failure  → installErr populated → ScreenError                  [US3]

ScreenSuccess / ScreenError
  any key                             → quitting = true
```
