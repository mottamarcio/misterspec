# Contract: `internal prepare` and the two new validation Finding codes

This documents the machine-readable contract this feature adds (Constitution Principle IX).

## 1. `internal prepare SPEC-### [--task TASK-###]`

**Args**: `SPEC-###` (required, positional) — the owning Spec. `--task TASK-###` (optional) — bare local form is sufficient since the Spec is already given; a composite `SPEC-###:TASK-###` naming a *different* Spec than the positional argument is rejected as `invalid_argument`.

**Success shape** (explicit Task, ready):

```json
{
  "ok": true,
  "preparation": {
    "task": "SPEC-014:TASK-003",
    "heading": "Add refresh-token rotation",
    "ready": true,
    "blockers": [],
    "requirements": [
      {"ref": "SPEC-014:R1", "content": "### R1 — Refresh tokens must rotate on every use.\n...", "fingerprint": "sha256:..."}
    ],
    "scope": "internal/auth/refresh.go",
    "verify": "go test ./internal/auth/... -run TestRefreshRotation",
    "plan_sections": [
      {"heading": "Implementation Sequence", "content": "...", "fingerprint": "sha256:...", "matched_requirements": ["SPEC-014:R1"]}
    ]
  }
}
```

**Success shape** (explicit Task, blocked) — same `ok: true` envelope, `ready: false`, `blockers` named, every other field still populated (data-model.md's validation rule):

```json
{"ok": true, "preparation": {"task": "SPEC-014:TASK-004", "heading": "...", "ready": false, "blockers": ["SPEC-014:TASK-003"], "requirements": [...], "scope": "...", "verify": "...", "plan_sections": [...]}}
```

A blocked preparation's process exit code is `10` (a new, dedicated code — distinct from `validate`'s own `valid: false` exit `4`, since "this Task is blocked" is a different condition than "this project has a structural problem") — `ok` stays `true` throughout; only the exit code signals "not actionable yet" to a calling script, consistent with how `validate` already distinguishes `ok` from its own semantic outcome.

**Success shape** (no `--task` given, auto-selected):

Identical shape, with the selected Task's own identity in `task`. When no Task in the Spec is Ready:

```json
{"ok": true, "preparation": null, "message": "no Task in SPEC-014 is ready — every Task is either complete or blocked"}
```

Exit code `0` in this case — an empty Spec-wide result is a valid, informative outcome (spec FR-012), not a blocked-Task condition.

## 2. New error cases

| Condition | Code | Exit |
|---|---|---|
| `--task` names a Task that doesn't exist in the Spec | `entity_not_found` (existing sentinel) | 3 |
| `--task`'s composite form names a different Spec than the positional argument | `invalid_argument` (existing sentinel) | 2 |
| The Spec has no `tasks.md` yet | `entity_not_found` — informative, naming the missing `tasks.md` (spec Edge Cases) | 3 |

No new error sentinel is introduced — every failure case reuses `operations`/`internalcmd`'s existing vocabulary.

## 3. New `validation.Finding` codes

| Code | Meaning | `Path` |
|---|---|---|
| `task_dependency_cycle` | A cycle exists in one Spec's Task-to-Task `Depends on:` graph (including self-dependency). | that Spec's `tasks.md#TASK-NNN` for the first Task in the cycle |
| `invalid_task_dependency` | A `Depends on:` entry names a nonexistent Task number, or a Task in a different Spec. | the declaring Task's own `tasks.md#TASK-NNN` |

Both follow the existing `Finding{Code, Severity: SeverityError, Path, Message}` shape (`internal/validation/findings.go`) — wired into `ValidateProject`/`ValidateEntity` the same way 032's `dependency_cycle`/coverage codes already are.

## 4. Example: catching a Task-dependency cycle before `prepare` ever runs

```json
{"ok": true, "valid": false, "findings": [
  {"code": "task_dependency_cycle", "severity": "error", "path": "ai/.../SPEC-014/tasks.md#TASK-002", "message": "task dependency cycle: SPEC-014:TASK-002 -> SPEC-014:TASK-003 -> SPEC-014:TASK-002"}
]}
```

`internal prepare SPEC-014 --task TASK-002` against the same project surfaces the same cycle as its own `ready: false` / blocked outcome (spec FR-008) — `internal validate` and `internal prepare` never disagree about the same underlying fact, mirroring 031/032's own single-shared-parser precedent.
