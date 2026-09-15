# Phase 1 Data Model: Project Bootstrap

Entities extracted from `spec.md` § Key Entities. This feature adds one
new package, `internal/bootstrap`, composing `internal/project`,
`internal/installer`, and `internal/agents` (research.md) — none of
those three packages themselves change.

## `internal/bootstrap`

### InspectResult

The outcome of inspecting a target directory — US1's "Inspection
Result" key entity.

| Field | Type | Notes |
|---|---|---|
| `Initialized` | bool | `false` for an uninitialized directory (FR-001); `true` if `targetDir` (or an ancestor — `project.Detect`'s own semantics, research.md) is already a misterspec project. |
| `ProjectRoot` | string | Set only when `Initialized`; the detected project's root (may differ from `targetDir` if it's a nested subdirectory — see research.md). |
| `InstalledAgent` | string | Set only when `Initialized` and an agent has actually been installed (`agents.CurrentInstall` returned `true`); empty otherwise (Edge Case: "already a project but no agent installed" — FR-002, Acceptance Scenario 3). |
| `AgentInstalled` | bool | Mirrors `agents.CurrentInstall`'s own second return value directly — `true` iff `InstalledAgent` is meaningful, never inferred from `InstalledAgent != ""` alone. |

**Validation rules**: `Inspect` never writes anything (FR-001, FR-002).
`project.ErrInvalidConfiguration` (a project marker exists but is
malformed — Edge Case) is propagated as a distinct returned `error`, not
folded into `InspectResult{Initialized: false}` — an invalid project is
not the same thing as no project (mirrors `project.Detect`'s own
`ErrNotInitialized` vs. `ErrInvalidConfiguration` distinction).

### BootstrapOutcome

The result of one bootstrap attempt — US2's "Bootstrap Outcome" key
entity. Never a single pass/fail flag (FR-008, SC-004).

| Field | Type | Notes |
|---|---|---|
| `ProjectRoot` | string | The bootstrapped project's root (`targetDir`, created if it didn't exist — Edge Case). |
| `ConfigWritten` | bool | Whether `.misterspec/config.yaml` was written successfully. |
| `TemplateOutcomes` | `[]installer.Outcome` | Per-kit-resource results, reused directly from `internal/installer.Install` — not a parallel type (FR-005). |
| `AgentInstall` | `agents.InstallResult` | The chosen agent's own `Install` result, reused directly — includes its own per-Skill `Outcomes` (FR-006). |

**Validation rules**: `Bootstrap` rejects (before writing anything) an
already-initialized target (`ErrAlreadyInitialized`, FR-004, SC-001) or
an unregistered agent ID (`ErrUnknownAgent`, FR-007, SC-002) — both
checked via `Inspect` and `registry.Get` respectively, before any file
is touched. A failure partway through (e.g. config write succeeds but
agent install fails) is reported via `BootstrapOutcome`'s own per-part
fields plus a non-nil `error` — never a partially-written resource
(FR-009, relying on `installer`'s and `agents.RecordInstall`'s existing
atomic-write guarantees for every individual file).

### VerifyResult

The outcome of verifying a completed bootstrap — US3's "Verification
Result" key entity.

| Field | Type | Notes |
|---|---|---|
| `Detected` | bool | Whether `targetDir` now detects as a valid project (FR-010, reusing `project.Detect` via `Inspect`). |
| `InstalledAgent` | string | The agent actually recorded as installed, per `Inspect`. |
| `AgentMatches` | bool | Whether `InstalledAgent` equals the `wantAgentID` the caller expected (FR-011, SC-005). |

**Validation rules**: `Verify` performs no writes; it is `Inspect` plus
one comparison (research.md's DRY decision) — never a second,
independent detection or record-reading implementation.

## State / Flow Summary

```text
Inspect(targetDir)                                                 [US1]
  project.Detect(targetDir)
    ErrNotInitialized        → InspectResult{Initialized: false}, nil
    ErrInvalidConfiguration  → InspectResult{}, err (propagated distinctly)
    *Project{Root, Config}   → agents.CurrentInstall(Root)
        (record, true,  nil) → InspectResult{true, Root, record.Agent.ID, true}, nil
        (_,      false, nil) → InspectResult{true, Root, "", false}, nil

Bootstrap(targetDir, agentID, registry, skills)                    [US2]
  Inspect(targetDir)
    Initialized == true      → BootstrapOutcome{}, ErrAlreadyInitialized
  registry.Get(agentID)
    ok == false              → BootstrapOutcome{}, ErrUnknownAgent
  os.MkdirAll(targetDir)                          # Edge Case: dir may not exist yet
  writeDefaultConfig(targetDir)  → .misterspec/config.yaml (atomic write)
  installer.Install(targetDir, overwrite=false)   → []installer.Outcome
  adapter.Install(ctx, agents.InstallRequest{targetDir, skills, false})
    → agents.InstallResult (includes agents.RecordInstall's write)
  → BootstrapOutcome{targetDir, true, outcomes, installResult}, nil

Verify(targetDir, wantAgentID)                                     [US3]
  Inspect(targetDir)
    → VerifyResult{
        Detected:       result.Initialized,
        InstalledAgent: result.InstalledAgent,
        AgentMatches:   result.InstalledAgent == wantAgentID,
      }, nil
```
