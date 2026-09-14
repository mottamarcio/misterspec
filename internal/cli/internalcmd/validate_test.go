package internalcmd_test

import (
	"encoding/json"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func decodeValidate(t *testing.T, output string) (valid bool, findingsCount int) {
	t.Helper()
	var decoded struct {
		OK       bool             `json:"ok"`
		Valid    bool             `json:"valid"`
		Findings []map[string]any `json:"findings"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if !decoded.OK {
		t.Error("ok = false, want true — validate ran successfully even when the project is invalid")
	}
	if decoded.Findings == nil {
		t.Error("findings = null, want an array (empty or populated)")
	}
	return decoded.Valid, len(decoded.Findings)
}

func TestValidateCmd_ValidProject(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: draft\n---\n")

	cmd := internalcmd.NewValidateCmd()
	cmd.SetArgs([]string{"--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	valid, count := decodeValidate(t, output)
	if !valid || count != 0 {
		t.Errorf("valid=%v count=%d, want true/0", valid, count)
	}
}

func TestValidateCmd_InvalidProject_ExitsFour(t *testing.T) {
	root := testutil.Project(t)
	// A Feature declaring a parent Program that does not exist —
	// missing_parent (internal/validation's own established check).
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md",
		"---\nid: FEAT-004\ntype: feature\nstatus: draft\nparent: PRG-999\n---\n")

	cmd := internalcmd.NewValidateCmd()
	cmd.SetArgs([]string{"--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 4 {
		t.Fatalf("exitCode = %d, want 4 (output: %s)", exitCode, output)
	}
	valid, count := decodeValidate(t, output)
	if valid || count == 0 {
		t.Errorf("valid=%v count=%d, want false/>0", valid, count)
	}
}

func TestValidateCmd_SingleEntity(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: draft\n---\n")

	cmd := internalcmd.NewValidateCmd()
	cmd.SetArgs([]string{"PRG-001", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	valid, _ := decodeValidate(t, output)
	if !valid {
		t.Error("valid = false, want true")
	}
}
