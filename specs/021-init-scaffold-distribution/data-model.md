# Phase 1 Data Model: Init Scaffolding and Binary Distribution

## `bootstrap.BootstrapOutcome` (modified)

```go
type BootstrapOutcome struct {
    ProjectRoot            string
    ConfigWritten          bool
    // DirectoriesScaffolded replaces TemplateOutcomes — the relative
    // paths (project-root-relative, forward-slash) of every directory
    // scaffoldDirectories created or verified present, in a fixed,
    // deterministic order (research.md #2, #4).
    DirectoriesScaffolded []string
    AgentInstall           agents.InstallResult
}
```

Fixed order (matches `project.Configuration`'s own field declaration
order, so output is stable and reviewable):

```text
ai              (ArtifactsDir)
ai/raw          (RawDir)
ai/knowledge    (KnowledgeDir)
ai/memory       (parent of ConstitutionPath)
ai/memory/learnings (LearningsDir)
ai/programs     (ProgramsRoot)
```

## `scaffoldDirectories` (new, `internal/bootstrap/scaffold.go`)

```go
// scaffoldDirectories creates every directory project.Configuration's
// own Default* constants name, relative to targetDir, if it does not
// already exist (os.MkdirAll — idempotent), and writes a zero-byte
// .gitkeep into any of them that is genuinely empty afterward (never
// into one that already has real content — FR-006). Returns the
// relative paths of every directory scaffolded, in the fixed order
// above.
func scaffoldDirectories(targetDir string) ([]string, error)
```

## `tui.previewContent` (modified)

```go
type previewContent struct {
    configPath      string
    // directories replaces templates — the same relative paths
    // scaffoldDirectories will create, computed from
    // project.Default* constants directly (nothing exists on disk yet
    // at Preview time) — research.md #5.
    directories     []string
    skillResources  []resourceSummary
    agentTargetPath string
}
```

## `internal context` JSON success payload — `init`'s own shape (modified)

```json
{
  "ok": true,
  "bootstrap": {
    "project_root": ".",
    "config_written": true,
    "directories": ["ai", "ai/raw", "ai/knowledge", "ai/memory", "ai/memory/learnings", "ai/programs"],
    "agent": {
      "adapter_id": "claude-code",
      "integration_path": ".claude/skills",
      "outcomes": [ ... ]
    }
  }
}
```

The `"templates"` key (`[{"name":...,"status":...}, ...]`) is removed
entirely — research.md #4.

## Release asset naming (`.github/workflows/release.yml`)

| GOOS | GOARCH | Asset name |
|---|---|---|
| linux | amd64 | `misterspec_linux_amd64` |
| linux | arm64 | `misterspec_linux_arm64` |
| darwin | amd64 | `misterspec_darwin_amd64` |
| darwin | arm64 | `misterspec_darwin_arm64` |
| windows | amd64 | `misterspec_windows_amd64.exe` |

Each built via `GOOS=<goos> GOARCH=<goarch> CGO_ENABLED=0 go build -o <asset-name> ./cmd/misterspec`, attached verbatim to the GitHub Release matching the pushed tag.
