package operations_test

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestParent_SpecHasFeatureParent(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: draft\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md", "---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\n---\n")

	result, err := operations.Parent(root, cfg, "SPEC-014")
	if err != nil {
		t.Fatalf("Parent() unexpected error: %v", err)
	}
	if !result.HasParent {
		t.Fatal("HasParent = false, want true")
	}
	if result.Parent.ID.String() != "FEAT-004" {
		t.Errorf("Parent.ID = %v, want FEAT-004", result.Parent.ID)
	}
}

func TestParent_ProgramHasNoParent(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: draft\n---\n")

	result, err := operations.Parent(root, cfg, "PRG-001")
	if err != nil {
		t.Fatalf("Parent() unexpected error: %v", err)
	}
	if result.HasParent {
		t.Fatalf("HasParent = true, want false: %+v", result)
	}
	if result.Parent != nil {
		t.Errorf("Parent = %v, want nil", result.Parent)
	}
}

func TestParent_TaskResolvesToOwningSpec(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md", "---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Add session persistence\n\n- [ ] Complete\n")

	result, err := operations.Parent(root, cfg, "TASK-001")
	if err != nil {
		t.Fatalf("Parent() unexpected error: %v", err)
	}
	if !result.HasParent {
		t.Fatal("HasParent = false, want true")
	}
	if result.Parent.ID.String() != "SPEC-001" {
		t.Errorf("Parent.ID = %v, want SPEC-001", result.Parent.ID)
	}
}
