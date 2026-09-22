package operations_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/evidence"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

// appendValidEvidence appends a passing, matching Evidence-Result
// record to taskHeading's own Section in the tasks.md already written
// at relPath — the checkbox must already be "- [x]". Code review
// finding (041-task-evidence-fingerprint): internal/operations.Inspect
// originally derived a Task's Status from its checkbox alone,
// contradicting internal/prepare's own, already-corrected definition
// for the same Task; these pre-existing fixtures assumed the old,
// checkbox-only behavior and now need real evidence to still exercise
// a genuinely "complete" Task (same class of fixture fix already
// applied to internal/cli/internalcmd/prepare_test.go's own
// markTaskComplete, for the same reason).
func appendValidEvidence(t *testing.T, root, relPath, taskHeading string) {
	t.Helper()
	path := filepath.Join(root, relPath)
	body, err := artifacts.ReadBody(path)
	if err != nil {
		t.Fatalf("appendValidEvidence: reading body from %s: %v", path, err)
	}
	doc := artifacts.ParseDocument(body)
	var fp string
	for _, section := range doc.Sections {
		if strings.HasPrefix(section.Heading, taskHeading) {
			fp = evidence.ContentFingerprint([]byte(section.Body))
			break
		}
	}
	if fp == "" {
		t.Fatalf("appendValidEvidence: could not locate %q's own Section to fingerprint", taskHeading)
	}

	// Rewrite the raw file (frontmatter included) — artifacts.ReadBody
	// above only returned the post-frontmatter body, used solely to
	// compute fp against the exact bytes production fingerprinting
	// would see.
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("appendValidEvidence: reading raw file %s: %v", path, err)
	}
	full := string(raw)
	evidenceBlock := "Evidence-Result: pass\nEvidence-Origin: declared\nEvidence-By: test\nEvidence-Fingerprint: " + fp + "\n"
	idx := strings.Index(full, "## "+taskHeading)
	if idx == -1 {
		t.Fatalf("appendValidEvidence: heading %q not found in %s", taskHeading, path)
	}
	next := strings.Index(full[idx+1:], "\n## ")
	var updated string
	if next == -1 {
		// Last section in the file: append before the final trailing
		// newline, never after it — an extra blank line here would
		// itself become part of the re-parsed Section.Body and
		// disagree with fp above (same pitfall documented in
		// markTaskComplete).
		updated = strings.TrimSuffix(full, "\n") + "\n" + evidenceBlock
	} else {
		insertAt := idx + 1 + next + 1 // just after the blank line preceding "## "
		updated = full[:insertAt] + evidenceBlock + full[insertAt:]
	}
	testutil.WriteFile(t, root, relPath, updated)
}

func TestInspect_WellFormedSpec(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on:\n  - SPEC-011\n---\n")

	result, err := operations.Inspect(root, cfg, "SPEC-014")
	if err != nil {
		t.Fatalf("Inspect() unexpected error: %v", err)
	}
	if result.Metadata.Status != "ready" {
		t.Errorf("Status = %q, want %q", result.Metadata.Status, "ready")
	}
	if result.Metadata.Parent == nil || result.Metadata.Parent.String() != "FEAT-004" {
		t.Errorf("Parent = %v, want FEAT-004", result.Metadata.Parent)
	}
	if len(result.Metadata.DependsOn) != 1 || result.Metadata.DependsOn[0].String() != "SPEC-011" {
		t.Errorf("DependsOn = %v, want [SPEC-011]", result.Metadata.DependsOn)
	}
	if result.Location.Path != "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md" {
		t.Errorf("Location.Path = %q, unexpected", result.Location.Path)
	}
}

func TestInspect_NotFound(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	_, err := operations.Inspect(root, cfg, "SPEC-999")
	if !errors.Is(err, operations.ErrEntityNotFound) {
		t.Fatalf("Inspect() error = %v, want errors.Is(err, ErrEntityNotFound)", err)
	}
}

func TestInspect_AmbiguousPropagatesFromResolve(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: draft\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-002/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: draft\nparent: PRG-002\n---\n")

	_, err := operations.Inspect(root, cfg, "FEAT-004")
	if !errors.Is(err, operations.ErrEntityAmbiguous) {
		t.Fatalf("Inspect() error = %v, want errors.Is(err, ErrEntityAmbiguous)", err)
	}
}

func TestInspect_TaskNarrowedMetadata(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n"+
			"# Tasks\n\n"+
			"## TASK-001 — Add session persistence\n\n- [ ] Complete\n\n"+
			"## TASK-002 — Add tests\n\n- [x] Complete\n")
	appendValidEvidence(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md", "TASK-002")

	pending, err := operations.Inspect(root, cfg, "TASK-001")
	if err != nil {
		t.Fatalf("Inspect(TASK-001) unexpected error: %v", err)
	}
	if pending.Metadata.Status != "pending" {
		t.Errorf("TASK-001 Status = %q, want %q", pending.Metadata.Status, "pending")
	}
	if pending.Metadata.Parent == nil || pending.Metadata.Parent.String() != "SPEC-001" {
		t.Errorf("TASK-001 Parent = %v, want SPEC-001", pending.Metadata.Parent)
	}

	complete, err := operations.Inspect(root, cfg, "TASK-002")
	if err != nil {
		t.Fatalf("Inspect(TASK-002) unexpected error: %v", err)
	}
	if complete.Metadata.Status != "complete" {
		t.Errorf("TASK-002 Status = %q, want %q", complete.Metadata.Status, "complete")
	}
}

// TestInspect_CompositeTaskReferenceScopedToOwningSpec covers spec.md
// User Story 1: Inspect(SPEC-A:TASK-001) must return Spec A's own task,
// even though Spec B has an unrelated TASK-001 of its own.
func TestInspect_CompositeTaskReferenceScopedToOwningSpec(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Spec 1's task\n\n- [ ] Complete\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/tasks.md",
		"---\ntype: tasks\nfor: SPEC-002\n---\n# Tasks\n\n## TASK-001 — Spec 2's task\n\n- [x] Complete\n")
	appendValidEvidence(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/tasks.md", "TASK-001")

	resultA, err := operations.Inspect(root, cfg, "SPEC-001:TASK-001")
	if err != nil {
		t.Fatalf("Inspect(SPEC-001:TASK-001) unexpected error: %v", err)
	}
	if resultA.Metadata.Status != "pending" {
		t.Errorf("Inspect(SPEC-001:TASK-001) Status = %q, want %q", resultA.Metadata.Status, "pending")
	}
	if resultA.Metadata.Parent == nil || resultA.Metadata.Parent.String() != "SPEC-001" {
		t.Errorf("Inspect(SPEC-001:TASK-001) Parent = %v, want SPEC-001", resultA.Metadata.Parent)
	}

	resultB, err := operations.Inspect(root, cfg, "SPEC-002:TASK-001")
	if err != nil {
		t.Fatalf("Inspect(SPEC-002:TASK-001) unexpected error: %v", err)
	}
	if resultB.Metadata.Status != "complete" {
		t.Errorf("Inspect(SPEC-002:TASK-001) Status = %q, want %q", resultB.Metadata.Status, "complete")
	}
}

func TestInspect_InvalidTarget(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	_, err := operations.Inspect(root, cfg, "not-an-id")
	if !errors.Is(err, operations.ErrInvalidTarget) {
		t.Fatalf("Inspect() error = %v, want errors.Is(err, ErrInvalidTarget)", err)
	}
}
