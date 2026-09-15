package operations_test

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
	"github.com/mottamarcio/misterspec/internal/validation"
)

func TestStatus_CountsMatchDirectEnumeration(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-002/program.md", "---\nid: PRG-002\ntype: program\nstatus: draft\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md", "---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md", "---\nid: SPEC-002\ntype: spec\nstatus: draft\nparent: FEAT-001\n---\n")

	summary, err := operations.Status(root, cfg)
	if err != nil {
		t.Fatalf("Status() unexpected error: %v", err)
	}
	if summary.Counts[ids.Program] != 2 {
		t.Errorf("Counts[Program] = %d, want 2", summary.Counts[ids.Program])
	}
	if summary.Counts[ids.Feature] != 1 {
		t.Errorf("Counts[Feature] = %d, want 1", summary.Counts[ids.Feature])
	}
	if summary.Counts[ids.Spec] != 2 {
		t.Errorf("Counts[Spec] = %d, want 2", summary.Counts[ids.Spec])
	}
}

func TestStatus_SpecsByState(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md", "---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md", "---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-003/spec.md", "---\nid: SPEC-003\ntype: spec\nstatus: draft\nparent: FEAT-001\n---\n")

	summary, err := operations.Status(root, cfg)
	if err != nil {
		t.Fatalf("Status() unexpected error: %v", err)
	}
	if summary.SpecsByState["ready"] != 2 {
		t.Errorf("SpecsByState[ready] = %d, want 2", summary.SpecsByState["ready"])
	}
	if summary.SpecsByState["draft"] != 1 {
		t.Errorf("SpecsByState[draft] = %d, want 1", summary.SpecsByState["draft"])
	}
}

func TestStatus_StructuralErrorsMatchesValidateProject(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	// A Feature with a missing parent — one structural problem.
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-999\n---\n")

	summary, err := operations.Status(root, cfg)
	if err != nil {
		t.Fatalf("Status() unexpected error: %v", err)
	}
	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if summary.StructuralErrors != len(findings) {
		t.Errorf("StructuralErrors = %d, want %d (ValidateProject's finding count)", summary.StructuralErrors, len(findings))
	}
}

func TestStatus_EmptyProject(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	summary, err := operations.Status(root, cfg)
	if err != nil {
		t.Fatalf("Status() unexpected error: %v", err)
	}
	if summary.Counts[ids.Program] != 0 || summary.Counts[ids.Feature] != 0 || summary.Counts[ids.Spec] != 0 {
		t.Errorf("Counts = %v, want all zero", summary.Counts)
	}
	if len(summary.SpecsByState) != 0 {
		t.Errorf("SpecsByState = %v, want empty", summary.SpecsByState)
	}
	if summary.StructuralErrors != 0 {
		t.Errorf("StructuralErrors = %d, want 0", summary.StructuralErrors)
	}
}
