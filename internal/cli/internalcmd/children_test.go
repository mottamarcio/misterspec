package internalcmd_test

import (
	"encoding/json"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func seedProgramWithFeatures(t *testing.T) string {
	t.Helper()
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: draft\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: draft\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-002/feature.md",
		"---\nid: FEAT-002\ntype: feature\nstatus: draft\nparent: PRG-001\n---\n")
	return root
}

func TestChildrenCmd_NoFilter(t *testing.T) {
	root := seedProgramWithFeatures(t)

	cmd := internalcmd.NewChildrenCmd()
	cmd.SetArgs([]string{"PRG-001", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		Children []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"children"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if len(decoded.Children) != 2 {
		t.Fatalf("len(children) = %d, want 2 (%+v)", len(decoded.Children), decoded.Children)
	}
}

func TestChildrenCmd_TypeFilter(t *testing.T) {
	root := seedProgramWithFeatures(t)

	cmd := internalcmd.NewChildrenCmd()
	cmd.SetArgs([]string{"PRG-001", "--type", "feature", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		Children []struct {
			Type string `json:"type"`
		} `json:"children"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	for _, c := range decoded.Children {
		if c.Type != "feature" {
			t.Errorf("child type = %q, want %q", c.Type, "feature")
		}
	}
}

func TestChildrenCmd_UnrecognizedTypeFilter(t *testing.T) {
	root := seedProgramWithFeatures(t)

	cmd := internalcmd.NewChildrenCmd()
	cmd.SetArgs([]string{"PRG-001", "--type", "not-a-real-type", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_argument")
}
