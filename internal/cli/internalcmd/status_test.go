package internalcmd_test

import (
	"encoding/json"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestStatusCmd_Success(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: draft\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: draft\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\n")

	cmd := internalcmd.NewStatusCmd()
	cmd.SetArgs([]string{"--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		OK               bool           `json:"ok"`
		Counts           map[string]int `json:"counts"`
		Specs            map[string]int `json:"specs"`
		StructuralErrors int            `json:"structural_errors"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if !decoded.OK {
		t.Error("ok = false, want true")
	}
	if decoded.Counts["program"] != 1 || decoded.Counts["feature"] != 1 || decoded.Counts["spec"] != 1 {
		t.Errorf("counts = %+v, unexpected", decoded.Counts)
	}
	if decoded.Specs["ready"] != 1 {
		t.Errorf("specs = %+v, want ready:1", decoded.Specs)
	}
	if decoded.StructuralErrors != 0 {
		t.Errorf("structural_errors = %d, want 0", decoded.StructuralErrors)
	}
}
