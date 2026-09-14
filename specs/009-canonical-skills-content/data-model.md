# Phase 1 Data Model: Canonical Skills Content

This feature's "entities" are content shapes, not Go types — extracted
from `spec.md`'s Key Entities plus `docs/architecture-specification.md`
§38-50. The one Go-level entity is `internal/installer`'s `Resource`,
whose meaning changes per research.md.

## `internal/installer.Resource` (meaning change, no field change)

| Field | Before (005/006) | After (this feature) |
|---|---|---|
| `Name` | The embedded file's bare filename (e.g. `"program.md.tmpl"`) | The resource's path relative to `sourceDir`, `fs.FS`-style forward slashes (e.g. `"create-plan/SKILL.md"`) — a bare filename is the zero-nesting special case, unchanged for every existing flat caller (research.md). |
| `Kind`, `ArtifactType` | Unchanged | Unchanged (`ArtifactType` remains harmless/unused for `"skill"`-kind resources, per 006-agent-adapter's own doc comment). |

## Skill (content entity)

One canonical Skill — a directory `kit/skills/<name>/` containing
exactly one `SKILL.md`.

| Attribute | Value |
|---|---|
| Name | One of the nine §38 names: `create-knowledge-base`, `create-constitution`, `create-program`, `create-feature`, `create-specs`, `create-plan`, `create-tasks`, `implement`, `analyze`. |
| Frontmatter | YAML block: `name` (matches the directory name), `description` (one sentence, non-empty) — the minimum Claude Code's own Skill discovery needs, matching this project's own installed `.claude/skills/speckit-*/SKILL.md` convention. |
| Body | The H2 sections below, in §39's order, every one present and non-empty. |

### Required body sections (§39, verbatim order)

```text
Purpose · Invocation · Responsibility · Inputs · Outputs · Preconditions
Required Context · Optional Context · Conditional Context · Unnecessary Context
Authority · Allowed Reads · Allowed Creates · Allowed Modifications · Forbidden Mutations
Deterministic Operations · Procedure · Decision Rules · Interaction Rules
Validation Rules · Failure Conditions · Stop Conditions · Success Criteria
Postconditions · Idempotency · Resume Behavior
Completion Contract · Recommended Next Step · Related Skills
```

**Validation rules** (FR-002, machine-checked per research.md): every
heading above must appear, in this order, as an H2 (`## `) in every
installed `SKILL.md`.

### Deterministic Operations per Skill (FR-003, SC-004 — the verifiable allowlist)

Every operation named below is one of the ten `misterspec internal <op>`
commands 008-cli-cobra actually registered. No Skill names anything
else (research.md's gap resolution: no `internal project`, no
`internal references`, no Task-ID allocator).

| Skill | Deterministic Operations |
|---|---|
| `create-knowledge-base` | `internal inventory raw`, `internal fingerprint <source>`, `internal create knowledge --slug <slug>`, `internal inspect <knowledge-id>`, `internal validate` |
| `create-constitution` | `internal inventory knowledge`, `internal resolve KNOW-###`, `internal validate` |
| `create-program` | `internal status`, `internal create program`, `internal validate PRG-###` |
| `create-feature` | `internal resolve PRG-###`, `internal children PRG-### --type feature`, `internal create feature --parent PRG-###`, `internal validate FEAT-###` |
| `create-specs` | `internal resolve FEAT-###`, `internal children FEAT-### --type spec`, `internal create spec --parent FEAT-###`, `internal validate SPEC-###` |
| `create-plan` | `internal resolve SPEC-###`, `internal inspect SPEC-###` (its `depends_on`/`supersedes` fields stand in for §46's `references`), `internal create-artifact plan --for SPEC-###`, `internal validate SPEC-###` |
| `create-tasks` | `internal resolve SPEC-###`, `internal inspect SPEC-###`, `internal create-artifact tasks --for SPEC-###`, `internal validate SPEC-###` (Task IDs are authored directly as `## TASK-NNN` headings — no allocator exists; `internal validate` catches duplicates) |
| `implement` | `internal resolve SPEC-###`, `internal inspect SPEC-###`, `internal validate SPEC-###` |
| `analyze` | `internal resolve SPEC-###`, `internal inspect SPEC-###`, `internal create-artifact validation --for SPEC-###`, `internal validate SPEC-###` |

### Completion Contract (§50, per-invocation output shape — content guidance, not a Go type)

Every Skill's **Completion Contract** section must instruct the agent
to end its own invocation reporting all five:

```text
Outcome · Artifacts · Important findings · Attention/unresolved issues
Recommended next step (naming the exact next Skill invocation, or the
earlier-stage Skill to return to — §53's branching rule, FR-005)
```

**Validation rules** (FR-004, machine-checked): every installed
`SKILL.md`'s "Completion Contract" section text must reference all five
concepts above (by their own established names from §50/§51's own
example).

## State / Flow Summary

```text
kit/skills/
  create-knowledge-base/SKILL.md
  create-constitution/SKILL.md
  create-program/SKILL.md
  create-feature/SKILL.md
  create-specs/SKILL.md
  create-plan/SKILL.md
  create-tasks/SKILL.md
  implement/SKILL.md
  analyze/SKILL.md
      ↓ embedded via kit.SkillsFS (fs.Sub, unchanged shape from 008)
      ↓ installer.ListFS/InstallFS, now recursive (research.md)
      ↓ agents/claude.Install (unchanged — already sourceDir ".")
.claude/skills/<name>/SKILL.md   — installed, per-Skill outcome reported
      ↓ Claude Code's own native Skill discovery (unchanged, no misterspec code)
/create-plan, /implement, ...    — invocable by the end user
```
