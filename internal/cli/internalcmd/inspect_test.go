package internalcmd_test

import (
	"encoding/json"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestInspectCmd_Success(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on: [SPEC-011]\n---\n")

	cmd := internalcmd.NewInspectCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		Entity struct {
			ID         string   `json:"id"`
			Type       string   `json:"type"`
			Status     string   `json:"status"`
			Parent     *string  `json:"parent"`
			DependsOn  []string `json:"depends_on"`
			Supersedes []string `json:"supersedes"`
		} `json:"entity"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.Entity.Status != "ready" {
		t.Errorf("status = %q, want %q", decoded.Entity.Status, "ready")
	}
	if decoded.Entity.Parent == nil || *decoded.Entity.Parent != "FEAT-004" {
		t.Errorf("parent = %v, want FEAT-004", decoded.Entity.Parent)
	}
	if len(decoded.Entity.DependsOn) != 1 || decoded.Entity.DependsOn[0] != "SPEC-011" {
		t.Errorf("depends_on = %v, want [SPEC-011]", decoded.Entity.DependsOn)
	}
	if decoded.Entity.Supersedes == nil {
		t.Error("supersedes = nil, want an empty array, not null")
	}
}

func TestInspectCmd_NoParent(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: draft\n---\n")

	cmd := internalcmd.NewInspectCmd()
	cmd.SetArgs([]string{"PRG-001", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		Entity struct {
			Parent *string `json:"parent"`
		} `json:"entity"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.Entity.Parent != nil {
		t.Errorf("parent = %v, want null", *decoded.Entity.Parent)
	}
}

func TestInspectCmd_NotFound(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewInspectCmd()
	cmd.SetArgs([]string{"SPEC-999", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 3 {
		t.Fatalf("exitCode = %d, want 3 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "entity_not_found")
}

// TestInspectCmd_CompositeTaskReference covers
// 031-canonical-task-identity contracts/task-identity-resolution.md §5:
// "inspect SPEC-A:TASK-001" must succeed and name only Spec A's task,
// even though Spec B has its own unrelated TASK-001.
func TestInspectCmd_CompositeTaskReference(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Spec 1's task\n\n- [ ] Complete\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/tasks.md",
		"---\ntype: tasks\nfor: SPEC-002\n---\n# Tasks\n\n## TASK-001 — Spec 2's task\n\n- [ ] Complete\n")

	cmd := internalcmd.NewInspectCmd()
	cmd.SetArgs([]string{"SPEC-001:TASK-001", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		Entity struct {
			ID     string  `json:"id"`
			Parent *string `json:"parent"`
		} `json:"entity"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.Entity.Parent == nil || *decoded.Entity.Parent != "SPEC-001" {
		t.Errorf("parent = %v, want SPEC-001", decoded.Entity.Parent)
	}
}

// TestInspectCmd_AmbiguousBareTaskRequiresSpecContext covers
// contracts/task-identity-resolution.md §5: a bare "inspect TASK-001"
// claimed by two Specs must fail with a structured
// spec_context_required error, not an arbitrary pick.
func TestInspectCmd_AmbiguousBareTaskRequiresSpecContext(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"# Tasks\n\n## TASK-001 — Spec 1's task\n\n- [ ] Complete\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/tasks.md",
		"# Tasks\n\n## TASK-001 — Spec 2's task\n\n- [ ] Complete\n")

	cmd := internalcmd.NewInspectCmd()
	cmd.SetArgs([]string{"TASK-001", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 3 {
		t.Fatalf("exitCode = %d, want 3 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "spec_context_required")
}
