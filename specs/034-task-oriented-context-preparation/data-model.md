# Phase 1 Data Model: Recuperação orientada à tarefa e preparação de execução

No persisted storage is added (Constitution Principle III) — every type below is computed fresh, in memory, from `spec.md`/`tasks.md`/`plan.md` content on each `prepare` or `validate` call.

## TaskDependency (`internal/validation`)

One Task's own `Depends on:` declaration, parsed the same way `TaskCoverage` already parses `Serves:` (research.md Decision 1).

| Field | Type | Description |
|---|---|---|
| `Task` | `ids.EntityID` (`Type == ids.Task`) | The declaring Task. |
| `DependsOn` | `[]ids.EntityID` (`Type == ids.Task`) | Every Task number this Task depends on, within the same Spec. Empty when the line is absent or explicitly `none`. |
| `MalformedDependsOn` | `[]string` | Any entry that failed to parse as a bare `TASK-###` — captured, not silently dropped (mirrors `TaskCoverage.MalformedReferences`). |

**Validation rules** (spec FR-006, FR-009):
- A `DependsOn` entry naming a Task number that does not exist in the same Spec's `tasks.md` is invalid — reported as `invalid_task_dependency`.
- A `DependsOn` entry using the composite `SPEC-###:TASK-###` form naming a *different* Spec than the declaring Task's own is invalid — Task dependencies are local to their own Spec, mirroring 031's own per-Spec Task identity scoping.
- A cycle (including self-dependency, the degenerate one-node case) among `DependsOn` edges within one Spec is detected and reported as `task_dependency_cycle`, reusing `internal/validation/dependency_graph.go`'s existing `DependencyGraph`/`detectCycles` at Task-number granularity (research.md Decision 2).

## TaskReadiness (`internal/prepare`)

The computed answer to "can this Task start now" (spec.md "Task Readiness").

| Field | Type | Description |
|---|---|---|
| `Task` | `ids.EntityID` | The Task being evaluated. |
| `Ready` | `bool` | `true` iff every entry in its `TaskDependency.DependsOn` has Status `"complete"` (via the existing per-Task checkbox status, `operations.inspectTask`'s own convention). An empty `DependsOn` is trivially `Ready`. |
| `Blockers` | `[]ids.EntityID` | The subset of `DependsOn` whose Status is not `"complete"` — empty when `Ready` is `true`. |

Recomputed from the current filesystem state on every call (spec FR-013) — never cached (research.md Decision 6).

## TaskFields (`internal/prepare`)

The new `Verify:`/`Scope:` lines, parsed the same way as `Serves:`/`Depends on:` (research.md Decision 1).

| Field | Type | Description |
|---|---|---|
| `Task` | `ids.EntityID` | The Task. |
| `Verify` | `string` | The `Verify:` line's own text, verbatim. Empty when absent. |
| `Scope` | `string` | The `Scope:` line's own text, verbatim. Empty when absent. |

## RequirementText (`internal/prepare`)

One served Requirement's own text, for a Task's assembled context (spec.md "Task Preparation").

| Field | Type | Description |
|---|---|---|
| `Ref` | `validation.RequirementRef` | The Requirement (`SPEC-###:R#`), from the Task's existing `Serves:` references. |
| `Content` | `string` | That Requirement's own section body, verbatim, from `spec.md`'s `### R<N>` section (via `artifacts.ParseDocument`). |
| `Fingerprint` | `string` | `contextengine.Fingerprint(Content)` — `"sha256:<hex>"`, reused for consistency with 033's per-item shape. |

## PlanSectionAssociation (`internal/prepare`)

One Plan section whose own content mentions a Requirement the Task serves (spec.md "Plan Section Association", research.md Decision 4).

| Field | Type | Description |
|---|---|---|
| `Heading` | `string` | The Plan section's own heading (e.g. `"Implementation Sequence"`). |
| `Content` | `string` | That section's own body, verbatim. |
| `Fingerprint` | `string` | `contextengine.Fingerprint(Content)`. |
| `MatchedRequirements` | `[]validation.RequirementRef` | Which of the Task's own served Requirements this section's text references — always non-empty (that's what makes the association exist). |

## TaskPreparation (`internal/prepare` — the `prepare` response)

The single response bundle (spec.md "Task Preparation", FR-001).

| Field | Type | Description |
|---|---|---|
| `Task` | `ids.EntityID` | The prepared Task's identity. |
| `Heading` | `string` | The Task's own `## TASK-NNN — <title>` heading text. |
| `Readiness` | `TaskReadiness` | Ready/blocked + named blockers. |
| `Requirements` | `[]RequirementText` | Every Requirement this Task serves, with its own text. |
| `Scope` | `string` | From `TaskFields.Scope`. |
| `Verify` | `string` | From `TaskFields.Verify`. |
| `PlanSections` | `[]PlanSectionAssociation` | Zero or more — empty is valid (spec Edge Cases: a Task with no Plan-section match still prepares successfully). |

**Validation rule** (spec FR-007): when `Readiness.Ready` is `false`, `Requirements`/`Scope`/`Verify`/`PlanSections` are still computed and returned (they're harmless facts about the Task), but the response as a whole is presented as blocked, not as ready-to-execute — the distinction lives entirely in `Readiness`, not in withholding the other fields.

## AutoSelection (`internal/prepare`)

The result of choosing a Task when none is named (spec.md FR-011/FR-012, research.md Decision 6).

| Outcome | Meaning |
|---|---|
| A `TaskPreparation` for the lowest-numbered Ready Task | The normal case. |
| An explicit "no Task is ready" result | Every Task in the Spec is either complete or blocked — spec FR-012 requires this to be stated clearly, never an empty/arbitrary response. |

## State / Lifecycle

No new lifecycle states. `TaskReadiness.Ready` is a derived, read-only value — a Task's own completion state (the existing checkbox) is the only thing that changes it, and that change happens through normal authoring (checking the box), never through `prepare` itself (FR-002).
