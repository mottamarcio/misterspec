# Contract: `misterspec internal next-task-id`

## §1. Invocation

```text
misterspec internal next-task-id [--dir <path>]
```

- `--dir` — target project directory (default `.`), same convention as
  every other `internal` command.
- Takes no other arguments — the result is project-wide by
  construction (research.md Decision 2); there is no per-Spec variant.

## §2. Success Response

```json
{
  "ok": true,
  "next_task_id": "TASK-048"
}
```

- `next_task_id` — the next available Task number, project-wide,
  formatted exactly like any other `EntityID.String()` (prefix +
  zero-padded number, width from `cfg.IDWidth`).
- When no Task exists anywhere in the project yet, `next_task_id` is
  `"TASK-001"` (spec FR-007).

## §3. Read-Only Guarantee

This command never writes to the filesystem, never modifies any
`tasks.md`, and never persists the returned number anywhere (spec
FR-003). Calling it twice in a row with no intervening write returns
the same value both times.

## §4. Error Responses

Follows the same `{"error": {"code", "message"}}` envelope every other
`internal` command already uses (Constitution Principle IX):

| Code | Condition |
|---|---|
| `project_not_found` | `--dir` does not resolve to a valid MisterSpec project (same condition `project.Detect` already reports for every other `internal` command). |

No other error condition applies — an empty project (no Programs,
Features, Specs, or Tasks at all) is not an error; it returns
`{"ok": true, "next_task_id": "TASK-001"}`.

## §5. Non-Goals

- Does **not** accept a Spec argument — the answer is the same
  regardless of which Spec will use it (research.md Decision 2).
- Does **not** report which Spec(s) currently hold the highest Task
  number, nor any collision information — that remains
  `internal migration-check-tasks`'s job, unchanged (spec FR-005).
- Does **not** reserve, lock, or persist the returned number — a
  second, concurrent caller before any write occurs may receive the
  same value (spec Edge Cases, research.md Decision 4).
