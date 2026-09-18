package internalcmd_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

// TestMigrationCheckTasksCmd_ReportsCollision covers
// 031-canonical-task-identity contracts/task-identity-resolution.md §4:
// the JSON shape (ok, collisions[].task_number/specs/paths) for a
// project with a genuine cross-Spec collision.
func TestMigrationCheckTasksCmd_ReportsCollision(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"# Tasks\n\n## TASK-001 — First in Spec 1\n\n- [ ] Complete\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/tasks.md",
		"# Tasks\n\n## TASK-001 — First in Spec 2\n\n- [ ] Complete\n")

	cmd := internalcmd.NewMigrationCheckTasksCmd()
	cmd.SetArgs([]string{"--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		OK         bool `json:"ok"`
		Collisions []struct {
			TaskNumber int      `json:"task_number"`
			Specs      []string `json:"specs"`
			Paths      []string `json:"paths"`
		} `json:"collisions"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if !decoded.OK {
		t.Fatalf("ok = false, want true (output: %s)", output)
	}
	if len(decoded.Collisions) != 1 {
		t.Fatalf("collisions = %+v, want exactly 1", decoded.Collisions)
	}
	if decoded.Collisions[0].TaskNumber != 1 {
		t.Errorf("collisions[0].task_number = %d, want 1", decoded.Collisions[0].TaskNumber)
	}
	if len(decoded.Collisions[0].Specs) != 2 {
		t.Errorf("collisions[0].specs = %v, want 2 entries", decoded.Collisions[0].Specs)
	}
}

// TestMigrationCheckTasksCmd_NoCollisionsIsEmptyArray covers spec.md
// User Story 3: an unaffected project returns an empty (not null)
// collisions array, still ok:true.
func TestMigrationCheckTasksCmd_NoCollisionsIsEmptyArray(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewMigrationCheckTasksCmd()
	cmd.SetArgs([]string{"--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		Collisions []any `json:"collisions"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.Collisions == nil {
		t.Error("collisions = null, want an empty array")
	}
}

// TestMigrationCheckTasksCmd_NeverWritesToTheFilesystem covers spec.md
// FR-007/FR-008: the diagnostic must never modify any artifact, even
// when it finds a genuine collision to report.
func TestMigrationCheckTasksCmd_NeverWritesToTheFilesystem(t *testing.T) {
	root := testutil.Project(t)
	taskFile := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md"
	testutil.WriteFile(t, root, taskFile, "# Tasks\n\n## TASK-001 — First in Spec 1\n\n- [ ] Complete\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/tasks.md",
		"# Tasks\n\n## TASK-001 — First in Spec 2\n\n- [ ] Complete\n")

	before, err := os.ReadFile(filepath.Join(root, taskFile))
	if err != nil {
		t.Fatalf("reading fixture before run: %v", err)
	}
	beforeInfo, err := os.Stat(filepath.Join(root, taskFile))
	if err != nil {
		t.Fatalf("statting fixture before run: %v", err)
	}

	cmd := internalcmd.NewMigrationCheckTasksCmd()
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
		t.Errorf("tasks.md content changed after running the diagnostic")
	}
	if !beforeInfo.ModTime().Equal(afterInfo.ModTime()) {
		t.Errorf("tasks.md mtime changed (%v -> %v) after a read-only diagnostic", beforeInfo.ModTime(), afterInfo.ModTime())
	}
}
