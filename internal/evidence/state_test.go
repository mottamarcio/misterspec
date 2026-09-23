package evidence

import (
	"strings"
	"testing"
)

// TestContentFingerprint_StableAcrossWritingItsOwnRecordedFingerprint
// is the self-referential-invalidation bug found during
// implementation (see state.go's own StripEvidenceLines doc comment):
// fingerprinting a Task body that already contains its own previously
// recorded "Evidence-*:" lines must equal the fingerprint of that same
// body with no Evidence-*: lines at all — otherwise writing evidence
// would immediately make itself Stale.
func TestContentFingerprint_StableAcrossWritingItsOwnRecordedFingerprint(t *testing.T) {
	before := []byte("- [x] Complete\nServes: SPEC-014:R1\n")
	withEvidence := []byte(string(before) +
		"Evidence-Result: pass\nEvidence-Origin: declared\nEvidence-By: reviewer\n" +
		"Evidence-Fingerprint: sha256:whatever-was-recorded-before\n")

	got := ContentFingerprint(withEvidence)
	want := ContentFingerprint(before)
	if got != want {
		t.Errorf("ContentFingerprint(withEvidence) = %q, want %q (must match the pre-evidence body's own fingerprint)", got, want)
	}
}

func TestStripEvidenceLines_RemovesOnlyEvidenceLabeledLines(t *testing.T) {
	body := []byte("- [x] Complete\nServes: SPEC-014:R1\nEvidence-Result: pass\nScope: internal/foo\nEvidence-Log: some/path.log\n")

	got := string(StripEvidenceLines(body))
	if strings.Contains(got, "Evidence-") {
		t.Errorf("StripEvidenceLines() = %q, still contains an Evidence- line", got)
	}
	if !strings.Contains(got, "Serves: SPEC-014:R1") || !strings.Contains(got, "Scope: internal/foo") {
		t.Errorf("StripEvidenceLines() = %q, must keep non-Evidence lines untouched", got)
	}
}

// TestDeriveState_Precedence is 041-task-evidence-fingerprint T011
// (Foundational) — one case per data-model.md "EvidenceState" row, in
// the documented precedence order (Unverified, then Failed, then the
// fingerprint comparison).
func TestDeriveState_Precedence(t *testing.T) {
	cases := []struct {
		name              string
		fields            EvidenceFields
		currentFingerprint string
		want              EvidenceState
	}{
		{
			name:              "no result at all is Unverified",
			fields:            EvidenceFields{Task: testTask()},
			currentFingerprint: "sha256:current",
			want:              Unverified,
		},
		{
			name: "a fail result is Failed, even if the fingerprint would otherwise match",
			fields: EvidenceFields{
				Task:        testTask(),
				Result:      "fail",
				Fingerprint: "sha256:current",
			},
			currentFingerprint: "sha256:current",
			want:               Failed,
		},
		{
			name: "a pass result with a mismatched fingerprint is Stale",
			fields: EvidenceFields{
				Task:        testTask(),
				Result:      "pass",
				Fingerprint: "sha256:old",
			},
			currentFingerprint: "sha256:current",
			want:               Stale,
		},
		{
			name: "a pass result with a matching fingerprint is Verified",
			fields: EvidenceFields{
				Task:        testTask(),
				Result:      "pass",
				Fingerprint: "sha256:current",
			},
			currentFingerprint: "sha256:current",
			want:               Verified,
		},
		{
			// Code review finding: a Result value that is neither
			// "pass" nor "fail" (a typo, a differently-cased value, or
			// any other hand-typed text) must never be silently
			// treated as a pass just because it isn't literally "fail"
			// — even when its Fingerprint happens to match current
			// content (e.g. copy-pasted from a prior valid record).
			name: "a malformed result is Unverified even with a matching fingerprint",
			fields: EvidenceFields{
				Task:        testTask(),
				Result:      "PASS",
				Fingerprint: "sha256:current",
			},
			currentFingerprint: "sha256:current",
			want:               Unverified,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := DeriveState(c.fields, c.currentFingerprint)
			if got != c.want {
				t.Errorf("DeriveState(%+v, %q) = %q, want %q", c.fields, c.currentFingerprint, got, c.want)
			}
		})
	}
}
