// This file compiles and runs specs/004-structural-validation/quickstart.md's
// end-to-end flow (validate a clean entity, validate a deliberately
// broken one, validate the whole project, then confirm Status's
// StructuralErrors matches) against internal/validation and
// operations.Status, on top of the fixture project the other quickstart
// tests in this package use (plan.md Phase 6, T013).
package example

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
	"github.com/mottamarcio/misterspec/internal/validation"
)

func TestValidationQuickstart_EndToEnd(t *testing.T) {
	root := testutil.Project(t)

	proj, err := project.Detect(root)
	if err != nil {
		t.Fatalf("project.Detect() unexpected error: %v", err)
	}

	prg, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("operations.Create(Program) unexpected error: %v", err)
	}

	// 1. A freshly created entity validates clean.
	clean, err := validation.ValidateEntity(proj.Root, proj.Config, prg.ID.String())
	if err != nil {
		t.Fatalf("validation.ValidateEntity() unexpected error: %v", err)
	}
	if len(clean) != 0 {
		t.Fatalf("ValidateEntity(%v) findings = %v, want empty for a just-created entity", prg.ID, clean)
	}

	// 2. A deliberately broken entity (written directly, bypassing
	// Create) reports a specific finding.
	testutil.WriteFile(t, root, "ai/programs/PRG-999/program.md", "---\nid: PRG-999\ntype: program\nstatus: not-a-real-status\n---\n")
	broken, err := validation.ValidateEntity(proj.Root, proj.Config, "PRG-999")
	if err != nil {
		t.Fatalf("validation.ValidateEntity() unexpected error: %v", err)
	}
	if len(broken) == 0 {
		t.Fatal("ValidateEntity(PRG-999) findings unexpectedly empty for an invalid status")
	}

	// 3. Whole-project validation aggregates both entities' results.
	projectFindings, err := validation.ValidateProject(proj.Root, proj.Config)
	if err != nil {
		t.Fatalf("validation.ValidateProject() unexpected error: %v", err)
	}
	if len(projectFindings) != len(broken) {
		t.Fatalf("ValidateProject() findings = %d, want %d (only PRG-999 is broken)", len(projectFindings), len(broken))
	}

	// 4. Status's StructuralErrors matches a direct ValidateProject call.
	summary, err := operations.Status(proj.Root, proj.Config)
	if err != nil {
		t.Fatalf("operations.Status() unexpected error: %v", err)
	}
	if summary.StructuralErrors != len(projectFindings) {
		t.Fatalf("Status().StructuralErrors = %d, want %d", summary.StructuralErrors, len(projectFindings))
	}
}
