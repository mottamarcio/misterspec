# Implementation Plan: Evidências de Execução e Validade por Fingerprint

**Branch**: `041-task-evidence-fingerprint` | **Date**: 2026-09-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/041-task-evidence-fingerprint/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Today `internal/prepare/scan.go`'s `taskStatus` derives a Task's
`"complete"` state purely from its `- [x]` checkbox — the exact gap
PROP-10 targets — and every downstream consumer (`readiness.go`'s
dependency-satisfaction check, `selection.go`'s "skip completed Tasks"
logic) trusts that string outright. The technical approach: (1) a new
deterministic operation, `internal capture-evidence`, computes a Task's
own Section-content fingerprint, the current Git revision and working-
tree dirty state, optionally runs one explicit, fully-specified
verification command, writes its full output to a new sibling
`<specDir>/evidence/<TASK-ID>-<timestamp>.log`, and returns every
`Evidence-*:` field's value as structured JSON — it never rewrites
`tasks.md` itself, since Constitution Principle VII already assigns
"task evidence" edits to `/mister-implement`, not the binary
(research.md #4); (2) `/mister-implement` writes those returned values
as new flat `Evidence-*:` lines in the Task's own body, the same
convention `Serves:`/`Depends on:`/`Verify:`/`Scope:` already use
(research.md #2); (3) `scan.go`'s `taskStatus` is redefined so
`"complete"` requires both the checkbox and a `Verified` evidence state
— `readiness.go`/`selection.go` need zero code changes, since they
already gate on that one string (research.md #1); (4) a new
`taskEvidenceFindings` check, wired into `internal/validation` exactly
where `specCoverageFindings` already is, surfaces an unverified, stale,
or failed Task as a distinct `validate` Finding (research.md #8).

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: `github.com/spf13/cobra` v1.10.2 (CLI). `os/exec` for the one explicit, caller-specified verification command this feature runs (no new third-party process/sandboxing library — research.md #5). No new third-party dependency.
**Storage**: Filesystem only (Constitution Principle III). Evidence is additive `Evidence-*:` lines inside a Task's own existing `tasks.md` Section body (research.md #2) plus one new log file per capture under a new `<specDir>/evidence/` directory (research.md #3) — both git-tracked, human-readable, fully reconstructable; no database, no new binary-format store.
**Testing**: `go test` — unit tests for `TaskContentFingerprint` (the new Task-scoped fingerprint helper), `vcs.HeadCommit`/`IsWorkingTreeDirty`, the new `internal/evidence` parsing/state-derivation package, `scan.go`'s redefined `taskStatus`, and the three new `internal/validation` Codes; filesystem-integration tests for `internal capture-evidence`'s declared-vs-automated paths (including a command that fails, one that times out, and one that would try to escape the project root); an end-to-end test proving `internal prepare`'s readiness computation withholds "ready" for a Task whose only upstream dependency is checked-but-unverified.
**Target Platform**: cross-platform CLI binary (Linux/macOS/Windows); the one subprocess this feature ever runs is the caller's own explicitly-named command, same portability expectations any other locally-run test/build command already has.
**Project Type**: single Go project (CLI + internal libraries)
**Performance Goals**: no new performance target; the added fingerprint/Git-status computation is on the same order as the file-based `operations.Fingerprint`/`vcs.CommitsSinceFileAdded` calls this codebase already performs on demand, not in any hot path.
**Constraints**: no implicit execution of Markdown-embedded text as a command (Principle I, spec FR-003); no new sandboxing/permission layer invented for the subprocess this feature runs — the calling agent/session's own existing tool-approval boundary remains the actual trust boundary (Principle IV, research.md #5); `--dir` MUST be rejected if it would resolve outside the project root, reusing `artifacts.RelativeWithinRoot` (Principle VIII, same guarantee `operations.Fingerprint` already gives); a failed verification MUST NOT be recorded or presented as `Verified` (Principle IX, spec FR-009); `internal capture-evidence` MUST NOT rewrite `tasks.md` (Principle VII — that stays `/mister-implement`'s own job, research.md #4).
**Scale/Scope**: `internal/vcs` (two new functions: `HeadCommit`, `IsWorkingTreeDirty`); `internal/operations` (`Fingerprint` internals factored into a shared `digestBytes` helper; new `evidence.go` implementing `CaptureEvidence`); new `internal/evidence` package (the `Evidence-*:` field parser + `EvidenceState` derivation, mirroring `internal/prepare/fields.go`'s own shape); `internal/prepare/scan.go` (`taskStatus` redefined to consult `internal/evidence`); `internal/validation` (`findings.go` gains three Codes; new `task_evidence.go`, wired into `validator.go` at `specCoverageFindings`'s own two call sites); `internal/cli/internalcmd` (new `capture-evidence.go` command). No new package layer beyond the one small, single-purpose `internal/evidence` package (parsing/state only — computation stays in `operations`/`vcs`, which already own fingerprinting and Git access, per Principle VI).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I (Semantic/Deterministic Separation)**: PASS. The binary computes fingerprint/Git-state/command-result facts mechanically; deciding *what* command verifies a requirement, and *whether* a passing command actually covers that requirement (spec's own stated Limit), stays the agent's/Skill's judgment — never inferred or executed automatically from prose.
- **Principle II (Deterministic Operations Are the Only Mutation Primitive)**: PASS. The Task-content fingerprint and Git revision snapshot — exactly the class of fact an LLM must never approximate — are produced solely by `internal capture-evidence`/`vcs`, never fabricated by a Skill.
- **Principle III (Filesystem Is the Single Source of Truth)**: PASS. Evidence is plain `Evidence-*:` Markdown lines plus a plain log file, both git-tracked and fully reconstructable; no new database or authoritative counter.
- **Principle IV (Simplicity First — YAGNI)**: PASS with an explicit boundary (research.md #4, #5, #6, #7): no new artifact type, no new sandboxing/permission mechanism, no repository-scope-aware dirty-tree filtering, no cross-artifact fingerprinting beyond the Task's own content — every one of these was considered and deliberately deferred or rejected as more than this feature's own gap requires.
- **Principle V (Test-First Discipline)**: Applies — fingerprint scoping, Git-state capture, evidence-state derivation, the redefined `taskStatus`, and the new validation Codes are exactly the deterministic logic this principle requires unit/integration tests for, written alongside implementation.
- **Principle VI (Clean Code & SOLID)**: PASS. Fingerprint computation stays in `internal/operations` (which already owns `Fingerprint`); Git facts stay in `internal/vcs` (which already owns every Git shell-out); evidence *parsing/state* is its own small package rather than bloating `internal/prepare` or `internal/validation` with a second responsibility neither currently has; both of those packages *consume* `internal/evidence` rather than reimplementing it.
- **Principle VII (Explicit Mutation Boundaries)**: PASS, and load-bearing for this feature's whole design (research.md #4): `internal capture-evidence` writes only the one new log file it creates; it never rewrites `tasks.md`, `spec.md`, or any other existing artifact — recording the returned evidence into the Task's own body remains `/mister-implement`'s own explicitly-assigned responsibility.
- **Principle VIII (Safety by Construction)**: PASS. `--dir` resolution reuses `artifacts.RelativeWithinRoot` (path-traversal rejection, same guarantee `Fingerprint` already gives); the new log file is written via the existing `writeAtomic` helper (`internal/operations/atomic_write.go`), not a bespoke write path.
- **Principle IX (Transparent, Machine-Readable Contracts)**: Applies directly — `internal capture-evidence` returns a stable, structured JSON envelope (ok/evidence fields), never prose; a failed subprocess still returns `ok: true` with `result: "fail"` (the capture itself succeeded at capturing a failure) versus a genuine operational error (bad `--dir`, timeout misconfiguration) returning `ok: false` with a stable error code — the two are never conflated.

No violations requiring Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/041-task-evidence-fingerprint/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/            # Phase 1 output (/speckit-plan command)
└── tasks.md              # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/
├── vcs/
│   ├── vcs.go                 # New: HeadCommit, IsWorkingTreeDirty
│   └── vcs_test.go
├── operations/
│   ├── fingerprint.go          # digestBytes extracted from Fingerprint's own body; TaskContentFingerprint added
│   ├── fingerprint_test.go
│   ├── evidence.go              # New: CaptureEvidence — fingerprint + Git snapshot + optional subprocess + log write
│   └── evidence_test.go
├── evidence/                    # New package: Evidence-*: field parsing + EvidenceState derivation
│   ├── doc.go
│   ├── fields.go                  # ParseEvidenceFields (mirrors prepare/fields.go's shape)
│   ├── fields_test.go
│   ├── state.go                   # EvidenceState (Verified/Stale/Failed/Unverified) derivation
│   └── state_test.go
├── prepare/
│   ├── scan.go                  # taskStatus consults internal/evidence before returning "complete"
│   └── scan_test.go
├── validation/
│   ├── findings.go              # CodeUnverifiedTask, CodeStaleTaskEvidence, CodeFailedTaskEvidence
│   ├── task_evidence.go          # New: taskEvidenceFindings, mirroring requirements.go's specCoverageFindings shape
│   ├── task_evidence_test.go
│   └── validator.go              # Wires taskEvidenceFindings in at specCoverageFindings's own two call sites
└── cli/
    └── internalcmd/
        ├── capture_evidence.go    # New: "internal capture-evidence SPEC-### TASK-###" command
        └── capture_evidence_test.go
```

**Structure Decision**: Single Go project, one new small package
(`internal/evidence`, parsing/state only — the same scale as
`internal/prepare` itself). Every other change widens a package that
already owns the exact responsibility involved: `internal/vcs` for Git
facts, `internal/operations` for fingerprinting and the one new
mutating operation, `internal/prepare` for the one-line redefinition of
what "complete" means, `internal/validation` for the new Finding codes.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations. Table intentionally omitted.
