package ids_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func nextTaskIDTestConfig() project.Configuration {
	return project.Configuration{ProgramsRoot: "ai/programs", IDWidth: 3}
}

// TestNextTaskID_EmptyProjectReturnsTaskOne is
// 045-global-task-numbering T001 (spec FR-007): no Task anywhere in the
// project yields TASK-001.
func TestNextTaskID_EmptyProjectReturnsTaskOne(t *testing.T) {
	root := testutil.Project(t)
	cfg := nextTaskIDTestConfig()

	got, err := ids.NextTaskID(root, cfg)
	if err != nil {
		t.Fatalf("NextTaskID() unexpected error: %v", err)
	}
	want := ids.EntityID{Type: ids.Task, Prefix: "TASK", Number: 1, Width: 3}
	if got != want {
		t.Errorf("NextTaskID() = %+v, want %+v", got, want)
	}
}

// TestNextTaskID_HighestNumberAnywhereInProjectWins is 045-global-task-
// numbering T001 (spec.md US1 Acceptance Scenario 1): the highest
// existing Task number found in a *different* Spec than the one
// currently empty still determines the next number — it never resets
// to TASK-001 just because the "current" Spec has no Tasks of its own
// yet.
func TestNextTaskID_HighestNumberAnywhereInProjectWins(t *testing.T) {
	root := testutil.Project(t)
	cfg := nextTaskIDTestConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"# Tasks\n\n## TASK-001 — First\n\n## TASK-047 — Highest so far\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-002/specs/SPEC-002/tasks.md",
		"# Tasks\n")

	got, err := ids.NextTaskID(root, cfg)
	if err != nil {
		t.Fatalf("NextTaskID() unexpected error: %v", err)
	}
	want := ids.EntityID{Type: ids.Task, Prefix: "TASK", Number: 48, Width: 3}
	if got != want {
		t.Errorf("NextTaskID() = %+v, want %+v (TASK-048, not TASK-001)", got, want)
	}
}

// TestNextTaskID_ContinuesWithinSameSpecWhenItHoldsTheMax is
// 045-global-task-numbering T001 (spec.md US1 Acceptance Scenario 3):
// when the Spec being decomposed already holds the project's highest
// Task numbers, the next number still continues from them.
func TestNextTaskID_ContinuesWithinSameSpecWhenItHoldsTheMax(t *testing.T) {
	root := testutil.Project(t)
	cfg := nextTaskIDTestConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"# Tasks\n\n## TASK-050 — First\n\n## TASK-051 — Second\n")

	got, err := ids.NextTaskID(root, cfg)
	if err != nil {
		t.Fatalf("NextTaskID() unexpected error: %v", err)
	}
	want := ids.EntityID{Type: ids.Task, Prefix: "TASK", Number: 52, Width: 3}
	if got != want {
		t.Errorf("NextTaskID() = %+v, want %+v", got, want)
	}
}

// TestNextTaskID_WidthMatchesConfig confirms the returned EntityID's
// Width always reflects cfg.IDWidth, not a hardcoded value.
func TestNextTaskID_WidthMatchesConfig(t *testing.T) {
	root := testutil.Project(t)
	cfg := nextTaskIDTestConfig()
	cfg.IDWidth = 4

	got, err := ids.NextTaskID(root, cfg)
	if err != nil {
		t.Fatalf("NextTaskID() unexpected error: %v", err)
	}
	if got.Width != 4 {
		t.Errorf("NextTaskID().Width = %d, want 4", got.Width)
	}
	if got.String() != "TASK-0001" {
		t.Errorf("NextTaskID().String() = %q, want %q", got.String(), "TASK-0001")
	}
}

// TestNextTaskID_ReadOnlyAndRepeatable is 045-global-task-numbering T001
// (contracts/next-task-id-contract.md §3): calling it twice with no
// intervening write returns the same value both times, and never
// modifies the filesystem.
func TestNextTaskID_ReadOnlyAndRepeatable(t *testing.T) {
	root := testutil.Project(t)
	cfg := nextTaskIDTestConfig()
	taskFile := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md"
	testutil.WriteFile(t, root, taskFile, "# Tasks\n\n## TASK-005 — First\n")

	before, err := os.ReadFile(filepath.Join(root, taskFile))
	if err != nil {
		t.Fatalf("reading fixture before calls: %v", err)
	}

	first, err := ids.NextTaskID(root, cfg)
	if err != nil {
		t.Fatalf("NextTaskID() first call unexpected error: %v", err)
	}
	second, err := ids.NextTaskID(root, cfg)
	if err != nil {
		t.Fatalf("NextTaskID() second call unexpected error: %v", err)
	}
	if first != second {
		t.Errorf("NextTaskID() first = %+v, second = %+v, want identical repeated calls", first, second)
	}

	after, err := os.ReadFile(filepath.Join(root, taskFile))
	if err != nil {
		t.Fatalf("reading fixture after calls: %v", err)
	}
	if string(before) != string(after) {
		t.Error("tasks.md content changed after calling NextTaskID — expected read-only")
	}
}
