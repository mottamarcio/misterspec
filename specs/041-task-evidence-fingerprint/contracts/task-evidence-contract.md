# Contract: Task Evidence & Fingerprint Validity

This documents the machine-readable/machine-checked contracts this
feature adds (Constitution Principle IX). None of this changes the
shape of any existing command's default output for a Task that never
uses `Evidence-*:` fields — a Task authored before this feature exists
is unaffected except that a checked-but-never-verified Task now
correctly reports `"pending"` instead of `"complete"` (research.md #1,
the one deliberate behavior change this feature exists to make).

## 1. `internal/vcs` (extended)

```go
package vcs

// HeadCommit returns the current commit's short SHA. available is
// false, with no error, when root is not a Git repository or has no
// commits yet — mirrors CommitsSinceFileAdded's own "bool separate
// from error" convention (research.md #7).
func HeadCommit(root string) (sha string, available bool, err error)

// IsWorkingTreeDirty reports whether `git status --porcelain` returns
// any output at all — repository-wide, not scoped to specific paths
// (research.md #7's deliberate, documented simplification).
func IsWorkingTreeDirty(root string) (bool, error)
```

## 2. `internal/operations` (extended)

```go
package operations

// TaskContentFingerprint hashes body (a Task's own already-extracted
// Section.Body) the same way Fingerprint hashes a file's bytes — pure,
// no I/O (data-model.md "FileFingerprint (extended)").
func TaskContentFingerprint(task ids.EntityID, body []byte) FileFingerprint

// EvidenceCaptureRequest / CaptureEvidence (data-model.md "Evidence
// Capture Request/Result"). Origin, By are always required. Exactly
// one of {Result} (Origin == "declared") or {Command, optionally Args/
// Dir/Timeout} (Origin == "automated") applies — CaptureEvidence
// returns an error naming the mismatch otherwise.
type EvidenceCaptureRequest struct {
	Spec    ids.EntityID
	Task    ids.EntityID
	Origin  string // "automated" | "declared"
	By      string
	Result  string // required iff Origin == "declared"
	Command string // required iff Origin == "automated"; forbidden otherwise
	Args    []string
	Dir     string
	Timeout time.Duration
}

// CaptureEvidence computes the Task's current content fingerprint and
// Git snapshot, runs Command (if Origin == "automated") with Dir
// resolved via artifacts.RelativeWithinRoot (rejecting an escape
// attempt, Principle VIII) and bounded by Timeout, writes its combined
// output to <specDir>/evidence/<Task>-<RFC3339-compact>.log via the
// existing writeAtomic helper, and returns the populated
// evidence.EvidenceFields — it never writes into tasks.md itself
// (research.md #4). A command that runs and exits non-zero still
// returns (fields, nil) with Result == "fail" — that is a successful
// capture of a failure, not an operational error; a genuine
// operational error (bad Dir, malformed request, I/O failure writing
// the log) returns a non-nil error.
func CaptureEvidence(root string, cfg project.Configuration, req EvidenceCaptureRequest) (evidence.EvidenceFields, error)
```

## 3. `internal/evidence` (new package)

```go
package evidence

// EvidenceFields is the parsed Evidence-*: lines from one Task's own
// Section.Body (data-model.md "EvidenceFields"). Never errors — an
// absent field is simply its zero value.
type EvidenceFields struct {
	Task        ids.EntityID
	Result      string
	Origin      string
	By          string
	CapturedAt  string
	Command     string
	GitRevision string
	WorkingTree string
	Fingerprint string
	Log         string
}

func ParseEvidenceFields(task ids.EntityID, body []byte) EvidenceFields

// EvidenceState is the Task's real completeness classification
// (data-model.md "EvidenceState").
type EvidenceState string

const (
	Unverified EvidenceState = "unverified"
	Failed     EvidenceState = "failed"
	Stale      EvidenceState = "stale"
	Verified   EvidenceState = "verified"
)

// DeriveState applies data-model.md's precedence (Unverified, then
// Failed, then a Fingerprint comparison against currentFingerprint,
// supplied by the caller).
func DeriveState(fields EvidenceFields, currentFingerprint string) EvidenceState

// ContentFingerprint hashes body identically to
// operations.TaskContentFingerprint — "sha256:<hex>" — for callers that
// cannot import internal/operations. Correction found during
// implementation: internal/operations already imports
// internal/validation (status.go, for StatusSummary.StructuralErrors),
// so internal/validation cannot import internal/operations without an
// import cycle. internal/prepare (which has no such constraint) calls
// operations.TaskContentFingerprint directly; internal/validation calls
// this instead. Both compute the identical digest (cross-checked by
// TestTaskContentFingerprint_MatchesEvidenceContentFingerprint in
// internal/operations/fingerprint_test.go) — kept as two small,
// independent implementations rather than introducing a new
// operations<->validation dependency edge for one three-line primitive.
func ContentFingerprint(body []byte) string
```

## 4. `misterspec internal capture-evidence SPEC-### --task TASK-###` (new command)

**Args**: `SPEC-###` positional (required, resolved via
`operations.Resolve` exactly as `internal prepare` already does);
`--task TASK-###` (required — Task IDs are Spec-local, matching
`internal prepare`'s own `SPEC-### --task TASK-###` shape rather than a
second, independently-resolvable positional argument).

**Flags**: `--origin automated|declared` (required), `--by <string>`
(required), `--result pass|fail` (required iff `--origin declared`),
`--command <string>` (required iff `--origin automated`; error if
supplied with `--origin declared`), `--args <string>` (repeatable),
`--exec-dir <string>` (optional, default `.`), `--timeout <duration>`
(optional, default `2m`); plus the `--dir <string>` every `internal`
command already has, for the target *project* directory.
**Correction found during implementation**: this contract originally
named the automated-command's own working directory `--dir`, colliding
with the pre-existing, unrelated `--dir` convention every other
`internal` command uses for "target project directory" (e.g. `internal
prepare SPEC-### --dir <path>`). Renamed to `--exec-dir` to avoid two
different meanings for the same flag name on this one command.

**Declared-origin example**:

```sh
misterspec internal capture-evidence SPEC-014 --task TASK-003 \
  --origin declared --by "reviewer:alice" --result pass
```

```json
{
  "ok": true,
  "evidence": {
    "task": "TASK-003",
    "result": "pass",
    "origin": "declared",
    "by": "reviewer:alice",
    "captured_at": "2026-09-22T14:30:00Z",
    "git_revision": "a1b2c3d",
    "working_tree": "clean",
    "input_fingerprint": "sha256:9f2c..."
  }
}
```

**Automated-origin example**:

```sh
misterspec internal capture-evidence SPEC-014 --task TASK-003 \
  --origin automated --by "go test" \
  --command go --args test --args ./internal/foo/... \
  --exec-dir . --timeout 90s
```

```json
{
  "ok": true,
  "evidence": {
    "task": "TASK-003",
    "result": "fail",
    "origin": "automated",
    "by": "go test",
    "captured_at": "2026-09-22T14:31:12Z",
    "command": "go test ./internal/foo/...",
    "git_revision": "a1b2c3d",
    "working_tree": "dirty",
    "input_fingerprint": "sha256:9f2c...",
    "log": "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/evidence/TASK-003-20260922T143112Z.log"
  }
}
```

**Error conditions** (`ok: false`, stable codes, Principle IX): a
`--exec-dir` resolving outside the project root; `--origin declared`
given alongside `--command`; `--origin automated` given without
`--command`; `--origin declared` given without `--result`; the named
Spec/Task not resolving to a real, existing Task.

## 5. `internal/validation` (extended)

```go
package validation

const (
	// ... existing codes unchanged ...
	CodeUnverifiedTask    = "unverified_task"
	CodeStaleTaskEvidence = "stale_task_evidence"
	CodeFailedTaskEvidence = "failed_task_evidence"
)
```

`taskEvidenceFindings(root, cfg, spec, specDir) ([]Finding, error)`
(new, `internal/validation/task_evidence.go`) is wired into
`ValidateProject`/`ValidateEntity` at `specCoverageFindings`'s own two
existing call sites (`validator.go`). For every checked Task in the
Spec's `tasks.md`, it computes `evidence.DeriveState` against that
Task's *current* content and raises exactly one of the three new
Codes when the state isn't `Verified` — an unchecked Task never raises
any of them.

## 6. `internal/prepare` (extended, behavior only — no new API)

`TaskInfo` gains `Evidence evidence.EvidenceState` (data-model.md
"TaskInfo (extended)"); `scan.go`'s `taskStatus` now returns
`"complete"` only when the checkbox is checked *and*
`Evidence == evidence.Verified`. `readiness.go` and `selection.go`
are unchanged — this is the entire mechanism satisfying spec FR-011/
User Story 2 (research.md #1).
