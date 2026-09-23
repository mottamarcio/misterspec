package validation_test

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/evidence"
	"github.com/mottamarcio/misterspec/internal/testutil"
	"github.com/mottamarcio/misterspec/internal/validation"
)

func specFixture(t *testing.T, root string) {
	t.Helper()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\n## Requirements\n\n### R1 — First\n\nSomething.\n")
}

// TestValidateEntity_CheckedTaskWithNoEvidenceIsUnverified is
// 041-task-evidence-fingerprint T015 (User Story 1): a checked Task
// with no Evidence-*: lines at all produces CodeUnverifiedTask.
func TestValidateEntity_CheckedTaskWithNoEvidenceIsUnverified(t *testing.T) {
	root := testutil.Project(t)
	specFixture(t, root)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do the thing\n\n- [x] Complete\nServes: SPEC-001:R1\n")

	findings, err := validation.ValidateEntity(root, testConfig(), "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeUnverifiedTask) {
		t.Errorf("findings = %+v, want %q", findings, validation.CodeUnverifiedTask)
	}
	if containsCode(findings, validation.CodeFailedTaskEvidence) || containsCode(findings, validation.CodeStaleTaskEvidence) {
		t.Errorf("findings = %+v, want no Failed/Stale code for an Unverified Task", findings)
	}
}

// TestValidateEntity_CheckedTaskWithFailedEvidenceIsFlagged covers the
// Failed branch.
func TestValidateEntity_CheckedTaskWithFailedEvidenceIsFlagged(t *testing.T) {
	root := testutil.Project(t)
	specFixture(t, root)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do the thing\n\n- [x] Complete\nServes: SPEC-001:R1\nEvidence-Result: fail\nEvidence-Origin: automated\nEvidence-By: go test\n")

	findings, err := validation.ValidateEntity(root, testConfig(), "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeFailedTaskEvidence) {
		t.Errorf("findings = %+v, want %q", findings, validation.CodeFailedTaskEvidence)
	}
}

// TestValidateEntity_UncheckedTaskNeverRaisesEvidenceCodes covers
// data-model.md's "A Task whose checkbox is unchecked never raises any
// of these three" rule.
func TestValidateEntity_UncheckedTaskNeverRaisesEvidenceCodes(t *testing.T) {
	root := testutil.Project(t)
	specFixture(t, root)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do the thing\n\n- [ ] Complete\nServes: SPEC-001:R1\n")

	findings, err := validation.ValidateEntity(root, testConfig(), "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	for _, code := range []string{validation.CodeUnverifiedTask, validation.CodeFailedTaskEvidence, validation.CodeStaleTaskEvidence} {
		if containsCode(findings, code) {
			t.Errorf("findings = %+v, want no %q for an unchecked Task", findings, code)
		}
	}
}

// TestValidateEntity_EditingVerifiedContentBecomesStale is
// 041-task-evidence-fingerprint T025 (User Story 3): a Task with
// Verified evidence, whose body is then edited, produces
// CodeStaleTaskEvidence — no manual staleness check required.
// Re-capturing evidence against the new content clears the Finding.
func TestValidateEntity_EditingVerifiedContentBecomesStale(t *testing.T) {
	root := testutil.Project(t)
	specFixture(t, root)
	tasksPath := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md"

	// Step 1: write a checked Task with no evidence yet, read back its
	// real Section.Body, and fingerprint that — the same real-parser
	// approach used elsewhere in this feature to avoid an off-by-a-
	// blank-line mismatch.
	full := testutil.WriteFile(t, root, tasksPath,
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do the thing\n\n- [x] Complete\nServes: SPEC-001:R1\nScope: internal/foo\n")
	body, err := artifacts.ReadBody(full)
	if err != nil {
		t.Fatalf("artifacts.ReadBody() unexpected error: %v", err)
	}
	var originalFingerprint string
	for _, section := range artifacts.ParseDocument(body).Sections {
		if section.Heading == "TASK-001 — Do the thing" {
			originalFingerprint = evidence.ContentFingerprint([]byte(section.Body))
		}
	}

	// Step 2: record Verified evidence against that exact content.
	testutil.WriteFile(t, root, tasksPath,
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do the thing\n\n- [x] Complete\nServes: SPEC-001:R1\nScope: internal/foo\n"+
			"Evidence-Result: pass\nEvidence-Origin: declared\nEvidence-By: reviewer\nEvidence-Fingerprint: "+originalFingerprint+"\n")

	findings, err := validation.ValidateEntity(root, testConfig(), "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if containsCode(findings, validation.CodeStaleTaskEvidence) || containsCode(findings, validation.CodeUnverifiedTask) || containsCode(findings, validation.CodeFailedTaskEvidence) {
		t.Fatalf("findings = %+v, want none before any edit (Task is Verified)", findings)
	}

	// Step 3: edit the Task's own verified content (Scope: changes) —
	// the recorded Evidence-Fingerprint now disagrees with reality.
	testutil.WriteFile(t, root, tasksPath,
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do the thing\n\n- [x] Complete\nServes: SPEC-001:R1\nScope: internal/bar\n"+
			"Evidence-Result: pass\nEvidence-Origin: declared\nEvidence-By: reviewer\nEvidence-Fingerprint: "+originalFingerprint+"\n")

	findings, err = validation.ValidateEntity(root, testConfig(), "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeStaleTaskEvidence) {
		t.Errorf("findings = %+v, want %q after editing the Task's own verified content", findings, validation.CodeStaleTaskEvidence)
	}

	// Step 4: re-capture evidence against the new content — the
	// Finding must disappear.
	body, err = artifacts.ReadBody(full)
	if err != nil {
		t.Fatalf("artifacts.ReadBody() unexpected error: %v", err)
	}
	var newFingerprint string
	for _, section := range artifacts.ParseDocument(body).Sections {
		if section.Heading == "TASK-001 — Do the thing" {
			newFingerprint = evidence.ContentFingerprint([]byte(section.Body))
		}
	}
	testutil.WriteFile(t, root, tasksPath,
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do the thing\n\n- [x] Complete\nServes: SPEC-001:R1\nScope: internal/bar\n"+
			"Evidence-Result: pass\nEvidence-Origin: declared\nEvidence-By: reviewer\nEvidence-Fingerprint: "+newFingerprint+"\n")

	findings, err = validation.ValidateEntity(root, testConfig(), "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if containsCode(findings, validation.CodeStaleTaskEvidence) {
		t.Errorf("findings = %+v, want no %q after re-capturing evidence against the new content", findings, validation.CodeStaleTaskEvidence)
	}
}
