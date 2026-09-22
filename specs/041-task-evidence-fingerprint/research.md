# Research: Evidências de Execução e Validade por Fingerprint

## 1. Where does "complete" actually come from today, and what has to change?

**Investigation**: `internal/prepare/scan.go`'s `taskStatus(body string)
string` returns `"complete"` purely from the first `- [x]` checkbox
line in a Task's own `Section.Body` — no other signal. Every consumer
of that boolean (`internal/prepare/readiness.go`'s dependency-readiness
computation, `selection.go`'s "skip already-complete Tasks" logic) reads
`TaskInfo.Status == "complete"` directly. `TaskFields` (fields.go)
already parses a `Verify:` line from the same body — a declared
verification *method*, never checked against anything.

**Decision**: Redefine `taskStatus` so `"complete"` requires both the
checkbox *and* a `Verified` evidence state (research.md #5) — done
entirely inside `scan.go`'s own `TaskInfo` construction. `readiness.go`
and `selection.go` need **zero** code changes: they already gate on
`Status == "complete"`, so making that string correctly reflect
evidence, not just the checkbox, mechanically satisfies spec FR-011/
User Story 2 without touching either file (Constitution Principle IV —
the smallest change that actually closes the gap, not a parallel
"is it really ready" API next to the one that already exists).

## 2. Where does evidence live — new artifact type, or existing Task body?

**Investigation**: A Task has no file of its own — it's one heading-
bounded `Section` inside the Spec's shared `tasks.md`. The project
already has a precedent for a Task declaring structured facts about
itself as flat `Label: value` lines in its own body, parsed by regex:
`Serves:` (`requirements.go`), `Depends on:` (`task_dependencies.go`),
`Verify:`/`Scope:` (`prepare/fields.go`). None of these introduced a
sub-heading or a new file.

**Decision**: Evidence is a small, fixed set of new `Evidence-*:` flat
lines in the Task's own body, parsed the same way — not a nested
Markdown sub-heading (which `artifacts.ParseDocument` would incorrectly
treat as ending the Task's own `Section.Body` early, since a Section
ends at the *next heading of any level*) and not a new artifact type.
One evidence record per Task, latest-capture-wins (re-running
verification replaces the recorded summary, matching how a Task's
`Verify:`/`Scope:` are themselves singular, current-state fields, not
an accumulating log) — full history of every attempt, including
superseded ones, lives in the sibling log files research.md #3
describes, never discarded.

**Alternatives considered**: A separate `evidence.md` artifact per Spec
— rejected: adds a sixth artifact type for a fact that is really a
property of one specific Task, and would require its own cross-
reference back to `tasks.md` for no benefit over an inline field.

## 3. Where do the "extensive logs, kept out of the Context Pack" live?

**Investigation**: 037-eval-quality-efficiency put its own dev-time
data under a project-level `eval/` directory, deliberately *outside*
`ai/` because it is evaluation data, not a product artifact
(plan.md's own "Structure Decision"). Task evidence is the opposite
case — it is real, permanent SDD project state (spec FR-001), so it
belongs inside `ai/` like every other project artifact, git-tracked and
reconstructable (Constitution Principle III).

**Decision**: A full capture's raw output (stdout/stderr, or an
imported log for declared evidence) is written to
`<specDir>/evidence/<TASK-ID>-<RFC3339-compact-timestamp>.log`, a new
sibling directory next to that Spec's own `tasks.md`. The inline
`Evidence-Log:` field names this file's path, relative to the project
root, so a Context Pack consumer sees a short reference (FR-010) and
can read the full file only if it actually needs to.

## 4. What computes the evidence, and does it write into `tasks.md` itself?

**Investigation**: Constitution Principle VII (Explicit Mutation
Boundaries) explicitly assigns "Task evidence" edits to `/mister-
implement`: "`/implement` may modify code, tests, and task evidence."
No existing `internal …` operation rewrites a specific line/field
inside an already-existing artifact's body — `internal fingerprint`,
`internal resolve`, `internal context`, `internal prepare` are all
read/compute-only; the only writes any operation performs are whole-
new-file creation (`create`, `create-artifact`) or, with this feature,
a brand-new evidence log file.

**Decision**: The new deterministic operation, `internal capture-
evidence`, is compute-and-optionally-execute, never a `tasks.md`
rewriter. It: (a) computes the Task's own Section-content fingerprint
(research.md #6); (b) captures the current Git revision and whether the
working tree is dirty (research.md #7); (c) when given an explicit
`--command`, runs exactly that command with the given args/dir/timeout
and captures its exit code plus combined output; (d) writes that output
to the new log file (research.md #3); (e) returns every `Evidence-*:`
field's value as structured JSON. `/mister-implement` (the Skill) is
the one that writes those returned values into the Task's own body as
the `Evidence-*:` lines — exactly mirroring how it already reads
`internal fingerprint`'s output and writes normal prose/checkbox edits
itself, and exactly matching the Constitution's own assignment of task-
evidence writes to that Skill, not to the binary.

**Alternatives considered**: Having `internal capture-evidence` rewrite
`tasks.md` directly — rejected: this would be the first "edit an
existing artifact's specific field in place" mutation primitive in the
whole codebase, a materially larger and riskier addition (Constitution
Principle IV) than the actual gap this feature closes; the binary
computing untrustworthy-if-agent-supplied facts (fingerprint, Git
revision, execution result) and handing them to the Skill to record is
sufficient and strictly smaller.

## 5. Command execution safety (FR-003, FR-004)

**Decision**: `--command`/`--args`/`--dir`/`--timeout` are separate,
explicit flags — never a single opaque string parsed out of Markdown
prose, and never inferred from a Task's own `Verify:` text automatically
(that stays a human/Skill-read hint, not machine-executed input,
consistent with Constitution Principle I: deciding *what* verifies a
requirement is semantic judgment, the agent's job). `--dir` is resolved
and rejected if it would escape the project root, reusing
`artifacts.RelativeWithinRoot` exactly as `operations.Fingerprint`
already does (Principle VIII). Permission enforcement itself is not
reinvented here: the calling agent/session's own tool-approval layer is
already the trust boundary for every command that session runs,
`internal capture-evidence`'s subprocess included — this feature's
actual contribution is making the invocation explicit, single-purpose,
and auditable (a real log file, a real exit code), not adding a new,
separate sandboxing mechanism the rest of the binary doesn't have
either (YAGNI, Principle IV).

## 6. Fingerprint scope: whole file, or just the Task?

**Investigation**: `operations.Fingerprint(root, path)` hashes an
entire file, streamed via `io.Copy` — appropriate for a source file,
wrong for a Task, since two different Tasks in the same `tasks.md`
would spuriously invalidate each other's evidence on any unrelated
edit to the shared file.

**Decision**: Factor `Fingerprint`'s SHA-256-of-bytes core into a small
unexported helper (`digestBytes([]byte) FileFingerprint`) reused by
both the existing file-based `Fingerprint` (unchanged public behavior)
and a new `TaskContentFingerprint(task ids.EntityID, body []byte)
FileFingerprint`, hashing exactly the Task's own already-parsed
`Section.Body` — the same scope `ParseTaskCoverage`/`ParseTaskDependsOn`
already read from. A later edit anywhere else in `tasks.md` never
affects a Task's own recorded fingerprint; an edit to that Task's own
body (its scope, its Serves:/Depends on:, or its own Evidence-* lines
themselves once superseded) does — exactly matching spec FR-006/FR-007
and User Story 3's own Independent Test.

**Correction found during implementation**: `internal/validation` also
needs this digest (to detect staleness, research.md #8), but cannot
import `internal/operations` — `internal/operations` already imports
`internal/validation` (`status.go`, for `StatusSummary.
StructuralErrors`), so the reverse import would be a cycle. The
identical primitive is therefore duplicated as
`evidence.ContentFingerprint(body []byte) string` in the new
`internal/evidence` package (a true leaf with no outgoing dependency on
either `operations` or `validation`), cross-checked by a dedicated test
so the two implementations never silently diverge — see data-model.md's
own correction note.

**Alternatives considered**: Fingerprinting the Spec's own `spec.md`
requirements text instead of/in addition to the Task body — rejected
for this feature per spec Assumptions: broader cross-artifact impact is
PROP-11's separate concern; this feature only detects a change to the
Task's own directly-verified content.

## 7. Git revision and working-tree state

**Decision**: Add two small functions to the existing `internal/vcs`
package (which already owns every Git shell-out in this codebase):
`HeadCommit(root) (sha string, available bool, err error)` (mirrors
`CommitsSinceFileAdded`'s own `available bool, no error` convention for
"not a repo") and `IsWorkingTreeDirty(root) (bool, error)` (`git status
--porcelain`, non-empty output means dirty). Dirty is computed
repository-wide, not scoped to files relevant to the Task's own scope —
a deliberately conservative simplification: a false "dirty" costs
nothing (the evidence still records exactly what state it saw), while a
missed "actually dirty" would misrepresent what was verified. Scoping
"relevant" more precisely is future work if it ever proves necessary in
practice, not before (Principle IV).

## 8. Surfacing stale/missing/failed evidence in `internal validate`

**Investigation**: `internal/validation/requirements.go`'s
`specCoverageFindings(root, cfg, spec, specDir)` is called from both
`ValidateProject` (every Spec in the project) and `ValidateEntity`
(one Spec directly) at the exact same two call sites — the established
shape for "a Spec-scoped check that needs to run in both validation
modes."

**Decision**: Add `taskEvidenceFindings(root, cfg, spec, specDir)
([]Finding, error)` in a new `internal/validation/task_evidence.go`,
wired into `ValidateProject`/`ValidateEntity` at the same two call
sites as `specCoverageFindings`. It raises one of three new `Code*`
constants per checked-but-not-`Verified` Task: `unverified_task`
(checkbox checked, no `Evidence-Result:` at all), `stale_task_evidence`
(recorded fingerprint disagrees with the Task's current content), and
`failed_task_evidence` (checkbox checked, but the most recent
`Evidence-Result:` is `fail`) — spec FR-009/FR-012's own vocabulary,
directly reusable by anything that already reads `validate`'s Findings.
