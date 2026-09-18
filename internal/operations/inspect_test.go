package operations_test

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

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
