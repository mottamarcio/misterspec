# Phase 1 Data Model: Multi-Agent Skill Integration

No new domain types. This feature adds five new implementations of an
already-existing interface (`agents.Adapter`, 006-agent-adapter) and
edits four already-existing Markdown Skill files' own content — nothing
new to model at the data level beyond what 006 already defined.

## `agents.Adapter` (unchanged interface, five new implementations)

```go
// Already defined in internal/agents/adapter.go — unmodified.
type Adapter interface {
    ID() string
    Name() string
    TargetPath() string
    Install(ctx context.Context, req InstallRequest) (InstallResult, error)
}
```

Five new concrete types, one per new package, each identical in shape
to `claude.Adapter` (research.md #2):

| Package | `ID()` | `Name()` | `TargetPath()` |
|---|---|---|---|
| `internal/agents/agy` | `"agy"` | `"Antigravity"` | `".agents/skills"` |
| `internal/agents/codex` | `"codex"` | `"Codex CLI"` | `".agents/skills"` |
| `internal/agents/copilot` | `"copilot"` | `"GitHub Copilot"` | `".github/skills"` |
| `internal/agents/cursoragent` | `"cursor-agent"` | `"Cursor"` | `".cursor/skills"` |
| `internal/agents/devin` | `"devin"` | `"Devin for Terminal"` | `".devin/skills"` |

Each `Install` method body is identical to `claude.Adapter.Install`
verbatim (research.md #2):

```go
func (a Adapter) Install(ctx context.Context, req agents.InstallRequest) (agents.InstallResult, error) {
    target := filepath.Join(req.ProjectRoot, targetPath)
    outcomes, err := installer.InstallFS(req.Skills, ".", "skill", target, req.Overwrite)
    if err != nil {
        return agents.InstallResult{}, err
    }
    result := agents.InstallResult{AdapterID: a.ID(), IntegrationPath: targetPath, Outcomes: outcomes}
    if err := agents.RecordInstall(req.ProjectRoot, result); err != nil {
        return agents.InstallResult{}, err
    }
    return result, nil
}
```

Note that both `agy` and `codex` resolve to the same `TargetPath()`
(`.agents/skills`) — this is a real, verified fact (research.md #1),
not a bug: both agents currently scan the identical project-relative
directory for Skills. Each remains its own distinct adapter (its own
`ID()`, its own `Name()`, independently selectable and independently
recorded in `install.json`) — installing for one does not imply the
other is installed, and `InstallRequest.Overwrite`'s own existing
no-clobber semantics apply exactly as they would for any two adapters
that happen to share a target, matching FR-004's own requirement that
one agent's installation never corrupt another's.

## `builtin.Default()` (extended, not new)

```go
// internal/agents/builtin/builtin.go
func Default() *agents.Registry {
    return agents.NewRegistry(
        claude.New(),
        agy.New(),
        codex.New(),
        copilot.New(),
        cursoragent.New(),
        devin.New(),
    )
}
```

## Canonical Skill file (existing artifact type, four instances edited)

No schema change. Each of the four files
(`kit/skills/{implement,create-plan,create-tasks,analyze}/SKILL.md`)
keeps its own existing frontmatter (`name`, `description`) and its own
existing fixed section structure (009-canonical-skills-content). The
edit touches exactly two sections per file:

1. **`Deterministic Operations` → Required operations**: one new
   bullet added, naming the specific intent (research.md #5):

   ```text
   - `internal context <target> --intent <intent>` — request a
     budgeted Context Pack before broader exploration.
   ```

2. **`Procedure`**: one new numbered step inserted immediately after
   the existing `resolve`/`inspect` step(s), before the Skill's own
   first exploration/decision step:

   ```text
   N. Run `internal context <target> --intent <intent>` and begin from
      its returned items. If the request fails, proceed using this
      Skill's own Required Context below instead (research.md #7).
   ```

Additionally, one sentence is added (in `Required Context` or
`Optional Context`, whichever already discusses exploration scope for
that Skill) making explicit that the agent remains free to read
further project files or invoke further deterministic operations
beyond what the Context Pack returns (FR-007).

`<target>` and `<intent>` per Skill (research.md #5):

| Skill | `<target>` | `<intent>` |
|---|---|---|
| `implement` | the Spec ID being implemented | `implementation` |
| `create-plan` | the Spec ID being planned | `planning` |
| `create-tasks` | the Spec ID being decomposed | `tasks` |
| `analyze` | the Spec ID being verified | `validation` |
