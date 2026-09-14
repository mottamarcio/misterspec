# Phase 1 Contracts: Project Bootstrap Package API

Like the prior six features, no HTTP/CLI surface yet. This feature's
contract is one new package, `internal/bootstrap`, composing
`internal/project`, `internal/installer`, and `internal/agents`
unmodified (research.md).

**Reconciled against the actual implementation (T013)** — unlike every
prior feature so far, no drift was found: every signature below (types,
field names, function signatures, both sentinel errors) matches the
shipped code exactly, including `Bootstrap`'s internal ordering (Inspect
check, then `registry.Get`, then `os.MkdirAll`, then config, then
templates, then the agent's own `Install`) as drafted. The one detail
this draft left implicit and worth naming explicitly now that it's
implemented: `writeDefaultConfig` and the `configFilePath` constant it
uses are unexported, package-local to `internal/bootstrap` (mirroring
`project`'s own unexported `configFilePath`) — never part of this
package's public surface, consistent with research.md's decision that
config-writing stays out of `internal/project` itself without needing to
be part of `bootstrap`'s public API either.

## `internal/bootstrap` — User Story 1: Inspect

```go
package bootstrap

// InspectResult is the outcome of inspecting a target directory,
// without writing anything (FR-001, FR-002).
type InspectResult struct {
    Initialized    bool
    ProjectRoot    string
    InstalledAgent string
    AgentInstalled bool
}

// Inspect determines whether targetDir is already a misterspec project
// and, if so, which agent (if any) is currently installed there. It
// performs no filesystem writes.
//
// Inspect returns (result, nil) for both "not yet a project" and
// "already a project" — those are not error conditions. It returns
// (InspectResult{}, err) with errors.Is(err, project.ErrInvalidConfiguration)
// only when a project marker exists but its configuration is malformed
// (Edge Case) — kept distinct from "not initialized" the same way
// project.Detect itself keeps ErrNotInitialized and
// ErrInvalidConfiguration distinct.
func Inspect(targetDir string) (InspectResult, error)
```

**Guarantees**: `Inspect` never creates, modifies, or deletes anything
(FR-001, FR-002). It reuses `project.Detect` and `agents.CurrentInstall`
directly — no parallel detection or record-reading logic (research.md).

## `internal/bootstrap` — User Story 2: Bootstrap

```go
package bootstrap

// ErrAlreadyInitialized is returned by Bootstrap when targetDir (or an
// ancestor) is already a misterspec project (FR-004, SC-001). Nothing is
// written when this is returned.
var ErrAlreadyInitialized = errors.New("bootstrap: already initialized")

// ErrUnknownAgent is returned by Bootstrap when agentID is not
// registered in registry (FR-007, SC-002). Nothing is written when this
// is returned.
var ErrUnknownAgent = errors.New("bootstrap: unknown agent")

// BootstrapOutcome is the result of one Bootstrap call — the specific
// outcome of its configuration, kit-resource, and agent-install parts,
// never a single pass/fail flag (FR-008, SC-004).
type BootstrapOutcome struct {
    ProjectRoot      string
    ConfigWritten    bool
    TemplateOutcomes []installer.Outcome
    AgentInstall     agents.InstallResult
}

// Bootstrap creates a new misterspec project at targetDir (creating the
// directory itself if it does not yet exist — Edge Case) for the given,
// already-chosen agentID: it writes a default project configuration,
// installs the framework's kit resources, and installs Skills for that
// agent via its own registered Adapter.
//
// Bootstrap checks Inspect and registry.Get(agentID) before writing
// anything (FR-004, FR-007, SC-001, SC-002). A failure partway through
// is never a silently partial resource — every individual file this
// composes (config.yaml, each kit template, each Skill, install.json)
// is written atomically by the package that owns it
// (project/installer/agents); BootstrapOutcome's own fields plus the
// returned error report exactly how far the attempt got (FR-009).
func Bootstrap(targetDir, agentID string, registry *agents.Registry, skills fs.FS) (BootstrapOutcome, error)
```

**Guarantees**:
- `Bootstrap` against an already-initialized target is always rejected,
  never silently overwriting anything already there (FR-004, SC-001,
  100% of attempts).
- `Bootstrap` naming an unregistered agent ID is always rejected before
  any file is written (FR-007, SC-002, 100% of attempts).
- The written configuration matches `project.Configuration`'s actual,
  already-implemented schema exactly (`Default*` constants reused
  directly) — not a second, independently-typed config writer
  (research.md).
- `BootstrapOutcome` always reports the specific outcome of every part
  attempted — never collapsed into one aggregate pass/fail (FR-008,
  SC-004).

## `internal/bootstrap` — User Story 3: Verify

```go
package bootstrap

// VerifyResult is the outcome of verifying a completed bootstrap: does
// the project detect successfully, and does the installed agent match
// what was requested.
type VerifyResult struct {
    Detected       bool
    InstalledAgent string
    AgentMatches   bool
}

// Verify confirms a completed bootstrap is genuinely usable: targetDir
// detects as a valid project, and the agent actually installed matches
// wantAgentID (FR-011, SC-005).
//
// Verify is Inspect plus one comparison — not a second, independent
// detection or record-reading implementation (research.md).
func Verify(targetDir, wantAgentID string) (VerifyResult, error)
```

**Guarantees**: A successful bootstrap's target always verifies with
`Detected == true` and `AgentMatches == true` for the agent it was
actually bootstrapped with, 100% of the time (SC-003, SC-005).

## Cross-cutting: error → future JSON mapping

Extends the running convention from 001-006. This feature's two new
sentinel errors follow the same pattern every prior feature's have:

| Condition | Future JSON shape |
|---|---|
| `Inspect` → `InspectResult{Initialized: false}, nil` | `{"ok": true, "initialized": false}` |
| `Inspect` → `err` wrapping `project.ErrInvalidConfiguration` | `{"ok": false, "error": "invalid_configuration", ...}` |
| `Bootstrap` → `err` wrapping `ErrAlreadyInitialized` | `{"ok": false, "error": "already_initialized"}` |
| `Bootstrap` → `err` wrapping `ErrUnknownAgent` | `{"ok": false, "error": "unknown_agent"}` |
| `BootstrapOutcome` | Its four fields map directly to a future response's `config`/`templates`/`agent` sections — each part's own outcome, not one aggregate flag. |
| `Verify` → `VerifyResult` | `{"ok": true, "detected": ..., "installed_agent": ..., "agent_matches": ...}` |
