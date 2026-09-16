# Contract: `internal validate` — Constitution Frontmatter Check

Extends the existing `misterspec internal validate` command's output — no new command, no changed invocation. Documents the new, additive Finding shape a caller (a Skill, or a developer) may now see.

## Before this feature

`internal validate`'s JSON `findings` array never contains any entry with `path: "ai/memory/constitution.md"` — the Constitution is entirely outside `ValidateProject`'s scan scope.

## After this feature

When `ai/memory/constitution.md` exists, one additional check runs (alongside the existing Program/Feature/Spec/Knowledge/Learning/Task checks, all unchanged):

```json
{
  "findings": [
    {
      "code": "frontmatter_malformed",
      "severity": "error",
      "path": "ai/memory/constitution.md",
      "message": "..."
    }
  ]
}
```

or

```json
{
  "findings": [
    {
      "code": "required_field_missing",
      "severity": "error",
      "path": "ai/memory/constitution.md",
      "message": "..."
    }
  ]
}
```

When the file doesn't exist yet, or exists with valid `type`/`schema_version` frontmatter, no Constitution-related entry appears in `findings` at all — identical to today's silence, not a new "ok" entry (matching every other check's existing silent-success convention).

## Invariants

- Never mutates `ai/memory/constitution.md` — read-only, like every other `internal validate` check.
- Reuses `CodeFrontmatterMalformed`/`CodeRequiredFieldMissing` — no new Finding code introduced.
- Does not alter any existing Finding this command already produces for Program/Feature/Spec/Knowledge/Learning/Task — `004-structural-validation`'s own contract is otherwise unchanged.

## `create-tasks` Completion Contract Addition (no new command)

Not a new CLI contract — documented content addition to an existing Skill's own output (see `data-model.md`'s "Task Dependency / Parallel-Safe Group"). The completion summary gains explicit lines naming dependency relationships and parallel-safe groups, composed as text from data already in `tasks.md` — never a new machine-readable field, never executed or validated by any command.
