package internalcmd_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func decodeNextTaskID(t *testing.T, output string) (ok bool, nextTaskID string) {
	t.Helper()
	var decoded struct {
		OK         bool   `json:"ok"`
		NextTaskID string `json:"next_task_id"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	return decoded.OK, decoded.NextTaskID
}

// TestNextTaskIDCmd_EmptyProjectReturnsTaskOne is 045-global-task-
// numbering T003 (contracts/next-task-id-contract.md §2).
func TestNextTaskIDCmd_EmptyProjectReturnsTaskOne(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewNextTaskIDCmd()
	cmd.SetArgs([]string{"--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	ok, next := decodeNextTaskID(t, output)
	if !ok {
		t.Fatalf("ok = false, want true (output: %s)", output)
	}
	if next != "TASK-001" {
		t.Errorf("next_task_id = %q, want %q", next, "TASK-001")
	}
}

// TestNextTaskIDCmd_HighestNumberInDifferentSpecWins is 045-global-
// task-numbering T003 (spec.md US1 Acceptance Scenario 1): the Spec
// currently being decomposed has no Tasks of its own yet, but another
// Spec already holds the project's highest Task number.
func TestNextTaskIDCmd_HighestNumberInDifferentSpecWins(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"# Tasks\n\n## TASK-047 — Highest so far\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-002/specs/SPEC-002/tasks.md",
		"# Tasks\n")

	cmd := internalcmd.NewNextTaskIDCmd()
	cmd.SetArgs([]string{"--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	_, next := decodeNextTaskID(t, output)
	if next != "TASK-048" {
		t.Errorf("next_task_id = %q, want %q (never TASK-001)", next, "TASK-048")
	}
}

// TestNextTaskIDCmd_ContinuesWithinSameSpecWhenItHoldsTheMax is
// 045-global-task-numbering T003 (spec.md US1 Acceptance Scenario 3).
func TestNextTaskIDCmd_ContinuesWithinSameSpecWhenItHoldsTheMax(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"# Tasks\n\n## TASK-050 — First\n\n## TASK-051 — Second\n")

	cmd := internalcmd.NewNextTaskIDCmd()
	cmd.SetArgs([]string{"--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	_, next := decodeNextTaskID(t, output)
	if next != "TASK-052" {
		t.Errorf("next_task_id = %q, want %q", next, "TASK-052")
	}
}

// TestNextTaskIDCmd_ResolvesUnambiguouslyByBareNumberAcrossSpecs is
// 045-global-task-numbering T008 (spec.md US2 Acceptance Scenario 1,
// SC-004): two Tasks numbered using consecutive next-task-id values,
// authored in two different Specs, each resolve unambiguously by their
// bare TASK-NNN number alone — no collision, because the numbers were
// obtained project-wide rather than per-Spec.
func TestNextTaskIDCmd_ResolvesUnambiguouslyByBareNumberAcrossSpecs(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"# Tasks\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-002/specs/SPEC-002/tasks.md",
		"# Tasks\n")

	cmd := internalcmd.NewNextTaskIDCmd()
	cmd.SetArgs([]string{"--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	_, first := decodeNextTaskID(t, output)
	if first != "TASK-001" {
		t.Fatalf("first next_task_id = %q, want %q", first, "TASK-001")
	}
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"# Tasks\n\n## TASK-001 — In Spec 1\n\n- [ ] Complete\n")

	cmd2 := internalcmd.NewNextTaskIDCmd()
	cmd2.SetArgs([]string{"--dir", root})
	output2, exitCode2 := runCmd(cmd2)
	if exitCode2 != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode2, output2)
	}
	_, second := decodeNextTaskID(t, output2)
	if second != "TASK-002" {
		t.Fatalf("second next_task_id = %q, want %q", second, "TASK-002")
	}
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-002/specs/SPEC-002/tasks.md",
		"# Tasks\n\n## TASK-002 — In Spec 2\n\n- [ ] Complete\n")

	proj, err := project.Detect(root)
	if err != nil {
		t.Fatalf("project.Detect() unexpected error: %v", err)
	}
	locFirst, err := operations.ResolveTask(root, proj.Config, "TASK-001", nil)
	if err != nil {
		t.Fatalf("ResolveTask(TASK-001) unexpected error: %v", err)
	}
	locSecond, err := operations.ResolveTask(root, proj.Config, "TASK-002", nil)
	if err != nil {
		t.Fatalf("ResolveTask(TASK-002) unexpected error: %v", err)
	}
	if locFirst.Path == locSecond.Path {
		t.Errorf("TASK-001 and TASK-002 resolved to the same path %q, want distinct Specs", locFirst.Path)
	}
}

// TestNextTaskIDCmd_NeverWritesToTheFilesystem is 045-global-task-
// numbering T003 (contracts/next-task-id-contract.md §3), mirroring
// TestMigrationCheckTasksCmd_NeverWritesToTheFilesystem's own shape.
func TestNextTaskIDCmd_NeverWritesToTheFilesystem(t *testing.T) {
	root := testutil.Project(t)
	taskFile := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md"
	testutil.WriteFile(t, root, taskFile, "# Tasks\n\n## TASK-005 — First\n")

	before, err := os.ReadFile(filepath.Join(root, taskFile))
	if err != nil {
		t.Fatalf("reading fixture before run: %v", err)
	}
	beforeInfo, err := os.Stat(filepath.Join(root, taskFile))
	if err != nil {
		t.Fatalf("statting fixture before run: %v", err)
	}

	cmd := internalcmd.NewNextTaskIDCmd()
	cmd.SetArgs([]string{"--dir", root})
	if _, exitCode := runCmd(cmd); exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0", exitCode)
	}

	after, err := os.ReadFile(filepath.Join(root, taskFile))
	if err != nil {
		t.Fatalf("reading fixture after run: %v", err)
	}
	afterInfo, err := os.Stat(filepath.Join(root, taskFile))
	if err != nil {
		t.Fatalf("statting fixture after run: %v", err)
	}

	if string(before) != string(after) {
		t.Error("tasks.md content changed after running next-task-id")
	}
	if !beforeInfo.ModTime().Equal(afterInfo.ModTime()) {
		t.Errorf("tasks.md mtime changed (%v -> %v) after a read-only command", beforeInfo.ModTime(), afterInfo.ModTime())
	}
}
