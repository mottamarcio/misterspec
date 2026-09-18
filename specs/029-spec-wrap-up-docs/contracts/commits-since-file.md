# Contract: `internal commits-since-file` Command

New deterministic CLI command under the hidden `misterspec internal …` namespace, consumed by `mister-wrap-up`.

## Invocation

```text
misterspec internal commits-since-file --path ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/plan.md
```

- `--path` is required — a repo-relative path to the file whose "added" commit anchors the range (always the named Spec's own `plan.md` in practice, but the operation itself is file-agnostic).

## Output (JSON)

```json
{
  "ok": true,
  "available": true,
  "commits": [
    {"hash": "a1b2c3d", "subject": "Implement TASK-001", "author_date": "2026-09-10"},
    {"hash": "e4f5a6b", "subject": "Add plan.md for SPEC-014", "author_date": "2026-09-09"}
  ]
}
```

Newest first, matching `git log`'s own default order. When history can't be determined (not a Git repository, or the file was never committed):

```json
{
  "ok": true,
  "available": false,
  "commits": []
}
```

`available: false` is never an error — `ok` stays `true`. `ok: false` is reserved for a genuine operational failure (e.g. `git` itself not installed).

## Invariants

- Read-only — never mutates the repository or any file.
- `available: false`, never a thrown error, when the target project isn't a Git repository or the file was never committed (spec.md FR-007) — the calling Skill's own Procedure branches on this exactly the way `022`'s `GitSkippedReason` already lets `create.go` branch on "not a repo" vs. "disabled" without treating either as a failure.
- Commit range is always `[commit that added --path, inclusive] .. [HEAD, inclusive]` on whatever branch is currently checked out — never derived from commit message content.

## `mister-wrap-up` Output Contract (the Skill's own file, not a new CLI contract)

Not a machine-readable API — a Markdown file at `cortex/SPEC-###-<slug>.md` (see `data-model.md`). No frontmatter schema is enforced and `internal validate` never checks it, unlike every other canonical artifact — it is downstream documentation material, regenerated wholesale on every `/mister-wrap-up` run (spec.md FR-009).
