package prepare

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/evidence"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

// TestScanSpecTasks_CheckedButUnverifiedTaskIsPending is
// 041-task-evidence-fingerprint T014 (User Story 1): a Task whose
// checkbox is checked but carries no Evidence-Result: line at all must
// report Status == "pending" — the exact gap this feature closes
// (contrary to this codebase's own pre-041 behavior, where the
// checkbox alone decided Status).
func TestScanSpecTasks_CheckedButUnverifiedTaskIsPending(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	specDir := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001"
	testutil.WriteFile(t, root, specDir+"/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\n## Requirements\n\n### R1 — First\n\nSomething.\n")
	testutil.WriteFile(t, root, specDir+"/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do the thing\n\n- [x] Complete\nServes: SPEC-001:R1\n")

	proj, err := project.Detect(root)
	if err != nil {
		t.Fatalf("project.Detect() unexpected error: %v", err)
	}

	spec := ids.EntityID{Type: ids.Spec, Prefix: "SPEC", Number: 1, Width: proj.Config.IDWidth}
	infos, found, err := ScanSpecTasks(proj.Root, proj.Config, spec, specDir)
	if err != nil {
		t.Fatalf("ScanSpecTasks() unexpected error: %v", err)
	}
	if !found {
		t.Fatal("found = false, want true")
	}
	if len(infos) != 1 {
		t.Fatalf("infos = %+v, want exactly 1 Task", infos)
	}
	if infos[0].Status != "pending" {
		t.Errorf("Status = %q, want %q — checked but no Evidence-Result: at all", infos[0].Status, "pending")
	}
	if infos[0].Evidence != evidence.Unverified {
		t.Errorf("Evidence = %q, want %q", infos[0].Evidence, evidence.Unverified)
	}
}

// TestScanSpecTasks_CheckedWithMatchingVerifiedEvidenceIsComplete
// covers the positive case: a checked Task carrying a passing result
// and a Fingerprint matching its own current content is genuinely
// complete.
func TestScanSpecTasks_CheckedWithMatchingVerifiedEvidenceIsComplete(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	specDir := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001"
	testutil.WriteFile(t, root, specDir+"/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\n## Requirements\n\n### R1 — First\n\nSomething.\n")

	// Write the Task body first with no Evidence-*: lines, then read it
	// back via the real artifacts parser to get the exact bytes
	// ScanSpecTasks will itself see as this Task's own Section.Body
	// (including the leading blank line after the heading) — hand-
	// computing that byte-for-byte would risk exactly the kind of
	// off-by-a-blank-line mismatch found during implementation
	// elsewhere in this feature (internal/cli/internalcmd/
	// prepare_test.go's markTaskComplete).
	tasksPath := testutil.WriteFile(t, root, specDir+"/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do the thing\n\n- [x] Complete\nServes: SPEC-001:R1\n")
	preEvidenceBody, err := artifacts.ReadBody(tasksPath)
	if err != nil {
		t.Fatalf("artifacts.ReadBody() unexpected error: %v", err)
	}
	var sectionBody string
	for _, section := range artifacts.ParseDocument(preEvidenceBody).Sections {
		if section.Heading == "TASK-001 — Do the thing" {
			sectionBody = section.Body
		}
	}
	fp := evidence.ContentFingerprint([]byte(sectionBody))

	testutil.WriteFile(t, root, specDir+"/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do the thing\n\n- [x] Complete\nServes: SPEC-001:R1\n"+
			"Evidence-Result: pass\nEvidence-Origin: declared\nEvidence-By: reviewer\nEvidence-Fingerprint: "+fp+"\n")

	proj, err := project.Detect(root)
	if err != nil {
		t.Fatalf("project.Detect() unexpected error: %v", err)
	}

	spec := ids.EntityID{Type: ids.Spec, Prefix: "SPEC", Number: 1, Width: proj.Config.IDWidth}
	infos, found, err := ScanSpecTasks(proj.Root, proj.Config, spec, specDir)
	if err != nil {
		t.Fatalf("ScanSpecTasks() unexpected error: %v", err)
	}
	if !found || len(infos) != 1 {
		t.Fatalf("infos = %+v found=%v, want exactly 1 Task", infos, found)
	}
	if infos[0].Evidence != evidence.Verified {
		t.Errorf("Evidence = %q, want %q", infos[0].Evidence, evidence.Verified)
	}
	if infos[0].Status != "complete" {
		t.Errorf("Status = %q, want %q", infos[0].Status, "complete")
	}
}

// TestScanSpecTasks_StaleEvidenceIsPending is
// 041-task-evidence-fingerprint T026 (User Story 3): a Task whose
// recorded Evidence-Fingerprint no longer matches its own current
// content reports Status == "pending", consistent with
// TestScanSpecTasks_CheckedButUnverifiedTaskIsPending's own pattern —
// a stale, no-longer-trustworthy pass is not treated as complete
// either.
func TestScanSpecTasks_StaleEvidenceIsPending(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	specDir := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001"
	testutil.WriteFile(t, root, specDir+"/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\n## Requirements\n\n### R1 — First\n\nSomething.\n")
	// A recorded fingerprint that deliberately does not match this
	// Task's own current content.
	testutil.WriteFile(t, root, specDir+"/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do the thing\n\n- [x] Complete\nServes: SPEC-001:R1\n"+
			"Evidence-Result: pass\nEvidence-Origin: declared\nEvidence-By: reviewer\nEvidence-Fingerprint: sha256:doesnotmatch\n")

	proj, err := project.Detect(root)
	if err != nil {
		t.Fatalf("project.Detect() unexpected error: %v", err)
	}
	spec := ids.EntityID{Type: ids.Spec, Prefix: "SPEC", Number: 1, Width: proj.Config.IDWidth}
	infos, found, err := ScanSpecTasks(proj.Root, proj.Config, spec, specDir)
	if err != nil {
		t.Fatalf("ScanSpecTasks() unexpected error: %v", err)
	}
	if !found || len(infos) != 1 {
		t.Fatalf("infos = %+v found=%v, want exactly 1 Task", infos, found)
	}
	if infos[0].Evidence != evidence.Stale {
		t.Errorf("Evidence = %q, want %q", infos[0].Evidence, evidence.Stale)
	}
	if infos[0].Status != "pending" {
		t.Errorf("Status = %q, want %q — stale evidence is not treated as complete", infos[0].Status, "pending")
	}
}
