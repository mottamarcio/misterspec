package evidence

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
)

func testTask() ids.EntityID {
	return ids.EntityID{Type: ids.Task, Prefix: "TASK", Number: 3, Width: 3}
}

// TestParseEvidenceFields_AllFieldsPresent is
// 041-task-evidence-fingerprint T009 (Foundational).
func TestParseEvidenceFields_AllFieldsPresent(t *testing.T) {
	body := []byte(`- [x] Complete
Serves: SPEC-014:R1
Evidence-Result: pass
Evidence-Origin: automated
Evidence-By: go test
Evidence-CapturedAt: 2026-09-22T14:30:00Z
Evidence-Command: go test ./internal/foo/...
Evidence-GitRevision: a1b2c3d
Evidence-WorkingTree: clean
Evidence-Fingerprint: sha256:deadbeef
Evidence-Log: ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/evidence/TASK-003-x.log
`)

	got := ParseEvidenceFields(testTask(), body)

	want := EvidenceFields{
		Task:        testTask(),
		Result:      "pass",
		Origin:      "automated",
		By:          "go test",
		CapturedAt:  "2026-09-22T14:30:00Z",
		Command:     "go test ./internal/foo/...",
		GitRevision: "a1b2c3d",
		WorkingTree: "clean",
		Fingerprint: "sha256:deadbeef",
		Log:         "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/evidence/TASK-003-x.log",
	}
	if got != want {
		t.Errorf("ParseEvidenceFields() = %+v, want %+v", got, want)
	}
}

// TestParseEvidenceFields_NoEvidenceLinesAtAll is spec Assumptions: a
// Task authored before this feature exists has every field at its zero
// value, never an error.
func TestParseEvidenceFields_NoEvidenceLinesAtAll(t *testing.T) {
	body := []byte("- [x] Complete\nServes: SPEC-014:R1\nDepends on: none\nScope: internal/foo\nVerify: go test ./internal/foo/...\n")

	got := ParseEvidenceFields(testTask(), body)

	want := EvidenceFields{Task: testTask()}
	if got != want {
		t.Errorf("ParseEvidenceFields() = %+v, want zero-valued fields %+v", got, want)
	}
}

// TestParseEvidenceFields_DeclaredOriginHasNoCommandOrLog confirms a
// declared-origin record — which never runs a command — simply has no
// Command/Log lines, and parses that absence as empty strings.
func TestParseEvidenceFields_DeclaredOriginHasNoCommandOrLog(t *testing.T) {
	body := []byte(`- [x] Complete
Evidence-Result: pass
Evidence-Origin: declared
Evidence-By: reviewer:alice
Evidence-CapturedAt: 2026-09-22T14:30:00Z
Evidence-GitRevision: a1b2c3d
Evidence-WorkingTree: clean
Evidence-Fingerprint: sha256:deadbeef
`)

	got := ParseEvidenceFields(testTask(), body)

	if got.Command != "" {
		t.Errorf("Command = %q, want \"\" for declared origin", got.Command)
	}
	if got.Log != "" {
		t.Errorf("Log = %q, want \"\" for declared origin", got.Log)
	}
	if got.Origin != "declared" || got.Result != "pass" {
		t.Errorf("Origin/Result = %q/%q, want declared/pass", got.Origin, got.Result)
	}
}
