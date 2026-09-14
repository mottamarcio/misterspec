package internalcmd_test

import (
	"encoding/json"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestBacklinksCmd_Success(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-011/spec.md", "---\nid: SPEC-011\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on: []\nsupersedes: []\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on:\n  - SPEC-011\nsupersedes: []\n---\n")

	cmd := internalcmd.NewBacklinksCmd()
	cmd.SetArgs([]string{"SPEC-011", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded struct {
		Target    string `json:"target"`
		Backlinks struct {
			Formal []struct {
				Relation string `json:"relation"`
				Source   string `json:"source"`
			} `json:"formal"`
			Semantic []any `json:"semantic"`
		} `json:"backlinks"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.Target != "SPEC-011" {
		t.Errorf("target = %q, want SPEC-011", decoded.Target)
	}
	if len(decoded.Backlinks.Formal) != 1 || decoded.Backlinks.Formal[0].Relation != "depends_on" || decoded.Backlinks.Formal[0].Source != "SPEC-014" {
		t.Errorf("backlinks.formal = %+v, want [{depends_on SPEC-014}]", decoded.Backlinks.Formal)
	}
	if decoded.Backlinks.Semantic == nil {
		t.Error("backlinks.semantic = null, want an empty array")
	}
}

func TestBacklinksCmd_EmptyResult(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-999-unused.md", "---\nid: KNOW-999\ntype: knowledge\nstatus: active\n---\n")

	cmd := internalcmd.NewBacklinksCmd()
	cmd.SetArgs([]string{"KNOW-999", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded struct {
		Backlinks struct {
			Formal   []any `json:"formal"`
			Semantic []any `json:"semantic"`
		} `json:"backlinks"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.Backlinks.Formal == nil || decoded.Backlinks.Semantic == nil {
		t.Errorf("backlinks = %+v, want both an empty array, not null", decoded.Backlinks)
	}
}

func TestBacklinksCmd_NotFound(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewBacklinksCmd()
	cmd.SetArgs([]string{"SPEC-999", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 3 {
		t.Fatalf("exitCode = %d, want 3 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "entity_not_found")
}

func TestBacklinksCmd_InvalidTarget(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do it\n\n- [ ] Complete\n")

	cmd := internalcmd.NewBacklinksCmd()
	cmd.SetArgs([]string{"TASK-001", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_target")
}
