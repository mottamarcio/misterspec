# Contract: Requirement coverage and dependency-cycle validation

This documents the new `validation.Finding` codes and the `internal validate` behavior this feature adds — Constitution Principle IX requires every outcome be a structured, stable-code Finding, never new prose-only output.

## 1. New Finding codes (`internal/validation/findings.go`)

| Code | Meaning | `Path` | Notes |
|---|---|---|---|
| `duplicate_requirement_id` | Two `### R<N>` headings with the same number in one Spec's body. | the Spec's `spec.md` | Mirrors `duplicate_id`'s shape but scoped to Requirement headings inside one file (spec FR-002). |
| `uncovered_requirement` | A declared Requirement has zero same-Spec `Serves:` reference anywhere in that Spec's own `tasks.md`. | the Spec's `spec.md` | `Message` names the Requirement number (spec FR-007). |
| `unknown_requirement_reference` | A `Serves:` reference's Requirement number does not exist in the named Spec, or the reference is syntactically malformed. | the Spec's `tasks.md#TASK-NNN` | `Message` names the raw reference text and the Task (spec FR-004). |
| `cross_spec_requirement_reference` | A `Serves:` reference names a Spec other than the Task's own owning Spec. | the Spec's `tasks.md#TASK-NNN` | Distinct from `unknown_requirement_reference` — the reference may be syntactically well-formed and even point at a real requirement, just not in this Task's own Spec (spec FR-005). |
| `task_without_requirement` | A Task has no `Serves:` line at all. | the Spec's `tasks.md#TASK-NNN` | (spec FR-006). |
| `dependency_cycle` | A cycle exists in the project-wide Spec `depends_on` graph. | the first Spec in the cycle's own `spec.md` | `Message` includes the full ordered path (spec FR-009). Reported once per distinct cycle, not once per participating Spec. |
| `phase_gate_blocked` | A Spec's `status` is not `draft` and it has an `uncovered_requirement` and/or `task_without_requirement` condition. | the Spec's `spec.md` | Emitted *in addition to* the underlying `uncovered_requirement`/`task_without_requirement` Finding(s) — it is the phase-aware escalation, not a replacement (spec FR-011, FR-012). |

All six follow the existing `Finding{Code, Severity: SeverityError, Path, Message}` shape (`internal/validation/findings.go`) — no new fields, no new envelope.

## 2. `validation.ValidateProject` — new checks wired in

Extends the existing per-entity-type loop (`internal/validation/validator.go:26-55`) and the existing Task-duplicate loop pattern (031-canonical-task-identity) with, for every discovered Spec:

1. Parse that Spec's own `### R<N>` headings (`requirements.go`, research.md Decision 2) → `duplicate_requirement_id` Finding(s) if any.
2. Parse that Spec's own `tasks.md` (if it exists — absence is not an error, mirrors `ids.ScanTasks`'s existing "empty is fine" behavior) for every Task's `Serves:` line(s) → `unknown_requirement_reference` / `cross_spec_requirement_reference` / `task_without_requirement` Finding(s).
3. Compute `RequirementCoverageReport.UncoveredRequirements` → `uncovered_requirement` Finding(s) for any Requirement number never referenced.
4. Apply the phase gate (research.md Decision 6) → `phase_gate_blocked` Finding when `status != "draft"` and step 2/3 found anything blocking.

Once per project (not once per Spec):

5. Build the `DependencyGraph` from every discovered Spec's `Metadata.DependsOn` and run cycle detection (research.md Decision 5) → one `dependency_cycle` Finding per distinct cycle found.

## 3. `validation.ValidateEntity(root, cfg, "SPEC-###")` — single-Spec parity (spec FR-013)

Runs steps 1–4 above scoped to the named Spec only (identical logic and Finding codes to `ValidateProject`'s per-Spec pass). For step 5 (cycles), the named Spec's participation is checked against the *same* project-wide graph — a Spec not on any cycle produces no `dependency_cycle` Finding; a Spec that *is* on a cycle produces the same Finding, with the same full path, that `ValidateProject` would produce for it. This mirrors `checkDependencyList`'s existing behavior of resolving against a full-project `ids.Scan(root, cfg, ids.Spec)` even when validating one Spec in isolation — validating one Spec has always meant "check this Spec's own relationships to the rest of the project," not "pretend the rest of the project doesn't exist."

## 4. Example: `internal validate` output shape

```json
{
  "ok": true,
  "valid": false,
  "findings": [
    {"code": "uncovered_requirement", "severity": "error", "path": "ai/.../SPEC-014/spec.md", "message": "SPEC-014: requirement R2 has no Task serving it"},
    {"code": "unknown_requirement_reference", "severity": "error", "path": "ai/.../SPEC-014/tasks.md#TASK-003", "message": "TASK-003 references SPEC-014:R9, which does not exist"},
    {"code": "cross_spec_requirement_reference", "severity": "error", "path": "ai/.../SPEC-014/tasks.md#TASK-004", "message": "TASK-004 (owned by SPEC-014) references SPEC-011:R1 — coverage must reference the task's own Spec"},
    {"code": "task_without_requirement", "severity": "error", "path": "ai/.../SPEC-014/tasks.md#TASK-005", "message": "TASK-005 has no Serves: reference"},
    {"code": "phase_gate_blocked", "severity": "error", "path": "ai/.../SPEC-014/spec.md", "message": "SPEC-014 has status \"ready\" but incomplete requirement coverage — not ready for implementation"},
    {"code": "dependency_cycle", "severity": "error", "path": "ai/.../SPEC-001/spec.md", "message": "dependency cycle: SPEC-001 -> SPEC-002 -> SPEC-003 -> SPEC-001"}
  ]
}
```

No change to `internal validate`'s own CLI surface (`internal/cli/internalcmd/validate.go`) — it already forwards whatever `ValidateProject`/`ValidateEntity` return; only the set of possible Finding codes grows.
