package operations_test

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

// TestTaskIdentityMigrationDiagnostic_ReportsCrossSpecCollision covers
// 031-canonical-task-identity spec.md User Story 3 / FR-008: a Task
// number claimed by more than one Spec project-wide (the case that
// collided under the old global-scan behavior) must be listed, naming
// every claiming Spec and path.
func TestTaskIdentityMigrationDiagnostic_ReportsCrossSpecCollision(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"# Tasks\n\n## TASK-001 — First in Spec 1\n\n- [ ] Complete\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/tasks.md",
		"# Tasks\n\n## TASK-001 — First in Spec 2\n\n- [ ] Complete\n")

	reports, err := operations.TaskIdentityMigrationDiagnostic(root, cfg)
	if err != nil {
		t.Fatalf("TaskIdentityMigrationDiagnostic() unexpected error: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("TaskIdentityMigrationDiagnostic() = %+v, want exactly 1 collision", reports)
	}
	report := reports[0]
	if report.TaskNumber != 1 {
		t.Errorf("report.TaskNumber = %d, want 1", report.TaskNumber)
	}
	if len(report.Specs) != 2 || report.Specs[0].String() != "SPEC-001" || report.Specs[1].String() != "SPEC-002" {
		t.Errorf("report.Specs = %v, want [SPEC-001 SPEC-002]", report.Specs)
	}
	if len(report.Paths) != 2 {
		t.Errorf("report.Paths = %v, want 2 entries", report.Paths)
	}
}

// TestTaskIdentityMigrationDiagnostic_NoCollisionsIsEmptyNotNil covers
// spec.md User Story 3: a project with no cross-Spec collisions is a
// valid, common outcome — an empty slice, not an error.
func TestTaskIdentityMigrationDiagnostic_NoCollisionsIsEmptyNotNil(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"# Tasks\n\n## TASK-001 — Only one\n\n- [ ] Complete\n")

	reports, err := operations.TaskIdentityMigrationDiagnostic(root, cfg)
	if err != nil {
		t.Fatalf("TaskIdentityMigrationDiagnostic() unexpected error: %v", err)
	}
	if len(reports) != 0 {
		t.Errorf("TaskIdentityMigrationDiagnostic() = %+v, want empty", reports)
	}
}

// TestTaskIdentityMigrationDiagnostic_SameSpecDuplicateIsNotACollision
// covers spec.md FR-004 vs FR-008: a genuine same-Spec duplicate is a
// validation problem (validation.ValidateProject's job), not a migration
// collision — the diagnostic must not double-report it.
func TestTaskIdentityMigrationDiagnostic_SameSpecDuplicateIsNotACollision(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"# Tasks\n\n## TASK-001 — First\n\n- [ ] Complete\n\n## TASK-001 — Also first, oops\n\n- [ ] Complete\n")

	reports, err := operations.TaskIdentityMigrationDiagnostic(root, cfg)
	if err != nil {
		t.Fatalf("TaskIdentityMigrationDiagnostic() unexpected error: %v", err)
	}
	if len(reports) != 0 {
		t.Errorf("TaskIdentityMigrationDiagnostic() = %+v, want empty — a same-Spec duplicate is not a cross-Spec collision", reports)
	}
}
