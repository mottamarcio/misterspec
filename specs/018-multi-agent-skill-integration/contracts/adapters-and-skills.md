# Phase 8-parallel Contract: Multi-Agent Skill Integration

**To be reconciled against the actual implementation during
`/speckit-implement`** (matching every prior feature's own contract
convention: written from the plan, updated only if implementation
reveals genuine drift).

## Adapter registry surface (extended, not new)

```go
package agents // unchanged interface (006-agent-adapter)

type Adapter interface {
    ID() string
    Name() string
    TargetPath() string
    Install(ctx context.Context, req InstallRequest) (InstallResult, error)
}
```

Six adapters are registered by `builtin.Default()` after this feature:

| `ID()` | `Name()` | `TargetPath()` |
|---|---|---|
| `claude-code` | Claude Code | `.claude/skills` (unchanged) |
| `agy` | Antigravity | `.agents/skills` |
| `codex` | Codex CLI | `.agents/skills` |
| `copilot` | GitHub Copilot | `.github/skills` |
| `cursor-agent` | Cursor | `.cursor/skills` |
| `devin` | Devin for Terminal | `.devin/skills` |

`agents.Registry.List()`/`Get(id)` (006, unmodified) already expose
these generically — no new discovery API.

## Installed Skill file shape (unchanged, per agent)

Every adapter installs the exact same source content
(`InstallRequest.Skills`, ultimately `kit.SkillsFS`) byte-for-byte,
under `<ProjectRoot>/<TargetPath()>/<skill-name>/SKILL.md` — the open
Agent Skills convention every one of the six agents scans (research.md
#1):

```markdown
---
name: <skill-name>
description: <one-line description>
---

<Skill body — unchanged by this feature>
```

## `internal context` call added to four Skill files (CLI usage, not a new command)

No new command. Each of the four Skills below gains one call to the
already-existing `misterspec internal context` command (017):

```text
misterspec internal context <target-id> --intent <intent>
```

| Skill file | `<intent>` |
|---|---|
| `kit/skills/implement/SKILL.md` | `implementation` |
| `kit/skills/create-plan/SKILL.md` | `planning` |
| `kit/skills/create-tasks/SKILL.md` | `tasks` |
| `kit/skills/analyze/SKILL.md` | `validation` |

Response handling is instruction-level, not code: on success, the
Skill begins from the returned `context.items`; on any error response
(`ok: false`), the Skill proceeds using its own pre-existing Required
Context/exploration approach instead (research.md #7) — it is never
treated as a Failure Condition.

## Error/behavioral guarantees carried over from 006/017 (not re-tested here, only orchestrated)

- `InstallFS`'s own no-silent-overwrite and containment guarantees
  apply unchanged to every new adapter (006 FR-004/FR-006).
- `internal context`'s own JSON envelope, error codes, and read-only
  guarantee apply unchanged to every Skill's new call (017 FR-006,
  FR-008).
- Installing for one agent never modifies another agent's own already-
  installed files (this feature's own FR-004), verified even for the
  two adapters (`agy`, `codex`) that happen to share a `TargetPath()`.
