package internalcmd_test

import (
	"encoding/json"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestParentCmd_HasParent(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md",
		"---\nid: FEAT-004\ntype: feature\nstatus: draft\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\n---\n")

	cmd := internalcmd.NewParentCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		Parent *struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"parent"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.Parent == nil {
		t.Fatal("parent = null, want FEAT-004")
	}
	if decoded.Parent.ID != "FEAT-004" || decoded.Parent.Type != "feature" {
		t.Errorf("parent = %+v, unexpected", decoded.Parent)
	}
}

func TestParentCmd_NoParent(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: draft\n---\n")

	cmd := internalcmd.NewParentCmd()
	cmd.SetArgs([]string{"PRG-001", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		OK     bool `json:"ok"`
		Parent *any `json:"parent"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if !decoded.OK {
		t.Error("ok = false, want true — no parent is not an error")
	}
	if decoded.Parent != nil {
		t.Errorf("parent = %v, want null", *decoded.Parent)
	}
}
