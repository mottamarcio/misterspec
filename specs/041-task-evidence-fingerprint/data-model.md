# Data Model: Evidências de Execução e Validade por Fingerprint

## EvidenceFields (new — `internal/evidence`)

The parsed, raw `Evidence-*:` lines from one Task's own `Section.Body`
— mirrors `internal/prepare.TaskFields`'s exact shape and parsing idiom
(one regex per label, absence yields a zero value, never an error).

| Field | Type | Notes |
|---|---|---|
| `Task` | `ids.EntityID` | The owning Task. |
| `Result` | `string` | `"pass"`, `"fail"`, or `""` if no `Evidence-Result:` line exists at all (spec's "nunca verificado" state). |
| `Origin` | `string` | `"automated"` or `"declared"`; `""` if `Result` is also `""`. |
| `By` | `string` | Free-text tool/person identifier. |
| `CapturedAt` | `string` | RFC 3339 timestamp, verbatim. |
| `Command` | `string` | Present only for `Origin == "automated"`. |
| `GitRevision` | `string` | Short commit SHA, or `""` if the project wasn't a Git repository at capture time. |
| `WorkingTree` | `string` | `"clean"` or `"dirty"`; `""` if `GitRevision` is also `""`. |
| `Fingerprint` | `string` | `"sha256:<hex>"`, the Task's own content digest at capture time. |
| `Log` | `string` | Project-relative path to the full capture log; present only for `Origin == "automated"`. |

**Validation rules**: `ParseEvidenceFields` never errors — a Task
authored before this feature exists simply has every field at its zero
value (spec Assumptions, same precedent as `TaskFields`). Exactly one
evidence record exists per Task at a time; a re-capture's returned
fields replace the prior ones when `/mister-implement` writes them
(research.md #2) — `ParseEvidenceFields` itself has no opinion about
history, it only ever reads whatever is currently written.

## EvidenceState (new — `internal/evidence`)

The Task's real completeness classification (spec FR-012, Key Entity
"Estado de Completude") — always exactly one of four values, derived
from `EvidenceFields` plus the Task's own *current* content fingerprint
(never the recorded one, which is only the point of comparison).

| Value | Meaning | Derivation |
|---|---|---|
| `Unverified` | No evidence ever captured. | `EvidenceFields.Result == ""`. |
| `Failed` | Most recent capture's own result was a failure. | `Result == "fail"`. |
| `Stale` | A passing result exists, but the Task's content has changed since capture. | `Result == "pass"` and `Fingerprint` != the Task's current `TaskContentFingerprint`. |
| `Verified` | A passing result exists and the fingerprint still matches current content. | `Result == "pass"` and `Fingerprint` == current `TaskContentFingerprint`. |

**Validation rule**: exactly one state applies — the checks above are
evaluated in the listed order (`Failed` first, then any `Result` other
than exactly `"pass"` is `Unverified`, then the fingerprint
comparison), so precedence is unambiguous, never a combination.
**Correction found by code review, after initial implementation**: the
first implementation of `DeriveState` checked only `Result == ""`
(`Unverified`) and `Result == "fail"` (`Failed`), letting any other
string — a typo, a different casing, hand-typed malformed text —
fall through to the fingerprint comparison and potentially be reported
`Verified`. `Unverified` is now the outcome for `Result == ""` **and**
for any value that isn't exactly `"pass"` or `"fail"` — an
unrecognized `Result` is never silently trusted as a pass.

## TaskInfo (extended — `internal/prepare`)

Existing type (`scan.go`); this feature adds one field and redefines
one existing field's own meaning.

| Field | Type | Notes |
|---|---|---|
| `Evidence` | `evidence.EvidenceState` | New. Computed once per `ScanSpecTasks` call, alongside `Coverage`/`Dependency`/`Fields`. |
| `Status` (redefined) | `string` | Still `"pending"`/`"complete"` — **no shape change**, only meaning: `"complete"` now additionally requires `Evidence == evidence.Verified` (spec FR-008/FR-011). A checked-but-`Unverified`/`Stale`/`Failed` Task now reports `Status == "pending"`. |

**Note**: `readiness.go`/`selection.go` read `Status` exactly as they
do today — this redefinition is the entire mechanism by which spec
User Story 2/FR-011 gets satisfied, with no change to either file
(research.md #1).

**Correction found by code review, after initial implementation**: the
"checked AND `Evidence == Verified`" decision itself was originally
inlined separately inside both `internal/prepare/scan.go`'s
`taskStatus` and (unnoticed at first) left un-updated in
`internal/operations/inspect.go`'s own, older `taskCheckboxStatus` —
the command behind `internal inspect`, which still derived `"complete"`
from the checkbox alone. The same Task could report `"complete"` from
`internal inspect` while `internal prepare`/`internal validate`
correctly treated it as not ready/flagged. Fixed by extracting the
decision into `evidence.TaskStatus(checked bool, state EvidenceState)
string` — the one leaf package both `internal/prepare` and
`internal/operations` can import without a cycle — and rewriting
`internal/operations/inspect.go`'s Task-status derivation to use
`artifacts.ParseDocument` (for the Task's full `Section.Body`, needed
for evidence parsing/fingerprinting) and call this same shared
function, so a third divergent definition can't reappear.

## FileFingerprint (extended — `internal/operations`)

Existing type (`fingerprint.go`); no field changes. A new function
alongside the existing `Fingerprint`:

| Function | Notes |
|---|---|
| `TaskContentFingerprint(task ids.EntityID, body []byte) FileFingerprint` | Hashes `body` (the Task's own already-extracted `Section.Body`) using the same SHA-256 algorithm as `Fingerprint`, via a shared internal `digestBytes` helper (research.md #6). Pure function — no I/O, unlike `Fingerprint` itself. |

**Correction found during implementation**: `internal/validation` needs
this same digest to detect staleness (`taskEvidenceFindings`), but
cannot import `internal/operations` — `internal/operations` already
imports `internal/validation` (`status.go`), so the reverse import
would cycle. The identical primitive is therefore also exposed as
`evidence.ContentFingerprint(body []byte) string` in the new
`internal/evidence` package (a true leaf, safe for both
`internal/prepare` and `internal/validation` to import), cross-checked
against `operations.TaskContentFingerprint` by a dedicated test so the
two never silently diverge. `internal/prepare` (no such constraint)
still calls `operations.TaskContentFingerprint` directly.

## Evidence Capture Request/Result (new — `internal/operations`)

Not a stored type — the input/output shape of the new `CaptureEvidence`
operation.

| Field | Type | Notes |
|---|---|---|
| `Spec` | `ids.EntityID` | |
| `Task` | `ids.EntityID` | |
| `Origin` | `string` | `"automated"` or `"declared"` — required, no default (spec FR-002: always explicit). |
| `By` | `string` | Required. |
| `Result` | `string` | Required when `Origin == "declared"` (no command to derive it from); ignored (derived from exit code) when `Origin == "automated"`. |
| `Command` | `string` | Required when `Origin == "automated"`; rejected when `Origin == "declared"` (spec FR-003: never blur the two). |
| `Args` | `[]string` | Optional, `Origin == "automated"` only. |
| `Dir` | `string` | Optional (defaults to the project root), `Origin == "automated"` only; MUST resolve inside the project root (Principle VIII). |
| `Timeout` | `time.Duration` | Optional (a fixed, documented default), `Origin == "automated"` only. |

Result: the populated `EvidenceFields` above (`Task` is implied, not
repeated), returned as JSON — never written to `tasks.md` (research.md
#4).

## Validation Codes (new — `internal/validation/findings.go`)

| Code | Constant | Raised when |
|---|---|---|
| `unverified_task` | `CodeUnverifiedTask` | A Task's checkbox is checked but `EvidenceState == Unverified`. |
| `stale_task_evidence` | `CodeStaleTaskEvidence` | A Task's checkbox is checked but `EvidenceState == Stale`. |
| `failed_task_evidence` | `CodeFailedTaskEvidence` | A Task's checkbox is checked but `EvidenceState == Failed`. |

A Task whose checkbox is unchecked never raises any of these three —
they exist specifically to catch the "checked but not actually
verified" gap (spec FR-008/FR-009), not to demand evidence for work
still in progress.
