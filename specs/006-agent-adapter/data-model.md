# Phase 1 Data Model: Agent Adapter Layer

Entities extracted from `spec.md` § Key Entities. This feature adds a new
package (`internal/agents`, plus `internal/agents/claude` and
`internal/agents/builtin`), and extends `internal/installer` additively
(research.md).

## `internal/installer` (extended)

| Symbol | Type | Notes |
|---|---|---|
| `InstallFS` | `func(source fs.FS, sourceDir, targetDir string, overwrite bool) ([]Outcome, error)` | Generalized form of `Install`; `Install` becomes a thin wrapper over it (research.md). |
| `ListFS` | `func(source fs.FS, sourceDir string) []Resource` | Generalized form of `List`. |

**Validation rules**: `Install`/`List`'s existing behavior and guarantees
(FR-003, FR-004, FR-005, FR-006, FR-007 from 005-embedded-kit) are
unchanged — verified by 005's existing test suite passing unmodified.

## `internal/agents`

### Adapter

The interface every coding-agent integration implements
(`docs/architecture-specification.md` §34).

| Method | Notes |
|---|---|
| `ID() string` | Stable identifier, e.g. `"claude-code"`. |
| `Name() string` | Human-readable, e.g. `"Claude Code"`. |
| `TargetPath() string` | The agent's integration location, relative to a project root, e.g. `".claude/skills"`. |
| `Install(ctx context.Context, req InstallRequest) (InstallResult, error)` | Materializes Skills into `TargetPath()` and writes the installation record — one call, one outcome (research.md). |

### InstallRequest

| Field | Type | Notes |
|---|---|---|
| `ProjectRoot` | string | The target project's root directory. |
| `Skills` | `fs.FS` | Source of Skill resources — a fixture in this feature's own tests; `kit.SkillsFS` once Phase 6 exists (research.md). |
| `Overwrite` | bool | Same meaning as `installer.Install`'s `overwrite`. |

### InstallResult

The in-memory outcome of one `Adapter.Install` call (FR-006).

| Field | Type | Notes |
|---|---|---|
| `AdapterID` | string | Which adapter ran. |
| `IntegrationPath` | string | Where it installed to, relative to `ProjectRoot`. |
| `Outcomes` | `[]installer.Outcome` | Per-resource results, reused directly from `internal/installer` — not a second, parallel per-resource result type. |

### InstallRecord

The on-disk record, `docs/architecture-specification.md` §22's schema
exactly.

| Field | Type | JSON key | Notes |
|---|---|---|---|
| `SchemaVersion` | int | `schema_version` | Always `1`. |
| `MisterspecVersion` | string | `misterspec_version` | Hardcoded `"0.1.0"` (research.md). |
| `AgentID` | string | `agent.id` | |
| `IntegrationPath` | string | `agent.integration_path` | |

**Validation rules** (FR-006, FR-008): `InstallRecord` is derived from an
`InstallResult` after a successful `Install`, written atomically to
`<ProjectRoot>/.misterspec/install.json` (reusing `internal/installer`'s
atomic-write helper). It is explicitly non-authoritative project state
(`docs/architecture-specification.md` §22, Constitution Principle III) —
a report of what happened, never re-derived as "the" source of truth for
what's actually on disk.

### Registry

| Method | Notes |
|---|---|
| `NewRegistry(adapters ...Adapter) *Registry` | Constructor; not global mutable state (research.md). |
| `(*Registry) List() []Adapter` | Every registered adapter, sorted by `ID()` for deterministic output. |
| `(*Registry) Get(id string) (Adapter, bool)` | `bool` false — not an error — for an unregistered ID (FR-003). |

## `internal/agents/claude`

| Symbol | Notes |
|---|---|
| `New() agents.Adapter` | Constructs the Claude Code adapter. `ID() == "claude-code"`, `TargetPath() == ".claude/skills"` — matching this project's own established convention. |

## `internal/agents/builtin`

| Symbol | Notes |
|---|---|
| `Default() *agents.Registry` | `agents.NewRegistry(claude.New())` — every concrete adapter this build knows about (research.md's cycle-avoidance decision). |

## State / Flow Summary

```text
Default() → *Registry{claude-code}                              [US1]
Registry.List() → []Adapter (sorted by ID, no filesystem access)
Registry.Get("claude-code") → Adapter, true
Registry.Get("unknown-id")  → nil, false

Adapter.Install(ctx, InstallRequest{ProjectRoot, Skills, Overwrite}) [US2]
  installer.InstallFS(req.Skills, ".", ProjectRoot+"/"+TargetPath(), Overwrite)
    → []installer.Outcome (Installed/Skipped/Failed per resource,
       identical guarantees to 005-embedded-kit's Install)
  recordInstall(ProjectRoot, InstallRecord{1, "0.1.0", ID(), TargetPath()})
    → atomic write to <ProjectRoot>/.misterspec/install.json
  → InstallResult{AdapterID, IntegrationPath, Outcomes}

CurrentInstall(projectRoot)                                       [US3]
  read <projectRoot>/.misterspec/install.json
    not present → InstallRecord{}, false, nil   (FR-007's "not installed")
    present, valid → InstallRecord, true, nil
    present, malformed → InstallRecord{}, false, error
```
