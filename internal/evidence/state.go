package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

// EvidenceState is the Task's real completeness classification
// (041-task-evidence-fingerprint data-model.md "EvidenceState", spec
// FR-012, Key Entity "Estado de Completude") — always exactly one of
// these four values.
type EvidenceState string

const (
	Unverified EvidenceState = "unverified"
	Failed     EvidenceState = "failed"
	Stale      EvidenceState = "stale"
	Verified   EvidenceState = "verified"
)

// DeriveState applies data-model.md's exact precedence: Unverified
// first (no Result at all), then Failed (a captured failure always
// wins, even over a stale-looking fingerprint from a much older pass),
// then a Fingerprint comparison against currentFingerprint — Stale on
// mismatch, Verified on match. currentFingerprint is supplied by the
// caller (internal/prepare uses operations.TaskContentFingerprint;
// internal/validation uses ContentFingerprint below, since it cannot
// import internal/operations — see doc.go).
func DeriveState(fields EvidenceFields, currentFingerprint string) EvidenceState {
	if fields.Result == "fail" {
		return Failed
	}
	if fields.Result != "pass" {
		// Empty (never captured) and anything else — a malformed or
		// hand-typed value that is neither "pass" nor "fail" — are
		// both Unverified: an untrusted or unrecognized Result value
		// must never be silently treated as a pass just because it
		// isn't the literal string "fail" (code review finding: the
		// original `else` branch fell through to the fingerprint
		// comparison for ANY non-"fail" value, so a typo like "PASS"
		// or "passed" — with a Fingerprint that happened to match,
		// e.g. copy-pasted from a prior valid record — was silently
		// reported Verified).
		return Unverified
	}
	if fields.Fingerprint != currentFingerprint {
		return Stale
	}
	return Verified
}

// TaskStatus is the one real definition of a Task's "pending"/
// "complete" status shared by every caller: "complete" iff checked is
// true AND state == Verified — "pending" otherwise, including a
// checked-but-Unverified/Stale/Failed Task (data-model.md "TaskInfo
// (extended)"). Extracted here, in the one leaf package both
// internal/prepare and internal/operations can import without a cycle,
// specifically so those two packages can never again define Task
// completion two different ways. Code review finding (041-task-
// evidence-fingerprint): internal/operations/inspect.go's own
// taskCheckboxStatus originally still derived "complete" from the
// checkbox alone, contradicting internal/prepare's already-corrected
// definition for the very same Task.
func TaskStatus(checked bool, state EvidenceState) string {
	if checked && state == Verified {
		return "complete"
	}
	return "pending"
}

// evidenceLinePattern matches any "Evidence-*:" line, regardless of
// which specific label — used only to strip these lines before
// fingerprinting (StripEvidenceLines), never to parse their values
// (ParseEvidenceFields owns that, one pattern per label).
var evidenceLinePattern = regexp.MustCompile(`^Evidence-[A-Za-z]+:`)

// StripEvidenceLines removes every "Evidence-*:" line from body before
// fingerprinting. Correction found during implementation: without this,
// fingerprinting a Task's raw current body (which already contains its
// own previously-recorded Evidence-Fingerprint: line) would never equal
// the fingerprint recorded *before* those lines existed — Verified
// would be unreachable the instant evidence is written, since writing
// the fingerprint changes the very content being fingerprinted. Both
// ContentFingerprint (below) and operations.TaskContentFingerprint
// apply this same stripping, so "the Task's own content" consistently
// means its real work fields (checkbox, Serves:, Depends on:, Scope:,
// Verify:) — never its own evidence bookkeeping about past captures.
func StripEvidenceLines(body []byte) []byte {
	lines := strings.Split(string(body), "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if evidenceLinePattern.MatchString(strings.TrimSpace(line)) {
			continue
		}
		kept = append(kept, line)
	}
	return []byte(strings.Join(kept, "\n"))
}

// ContentFingerprint hashes body — after stripping any existing
// "Evidence-*:" lines (StripEvidenceLines) — the same way
// operations.TaskContentFingerprint does: "sha256:<hex>". For callers
// that cannot import internal/operations without an import cycle
// (internal/operations already imports internal/validation via
// status.go, so internal/validation cannot import internal/operations
// in turn; correction found during implementation, see doc.go). Both
// implementations compute the identical SHA-256 digest over the
// identically-stripped content; kept as two small, independent
// primitives rather than introducing a new shared dependency edge
// between operations and validation.
func ContentFingerprint(body []byte) string {
	sum := sha256.Sum256(StripEvidenceLines(body))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}
