package internalcmd_test

import (
	"encoding/json"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func seedSpec(t *testing.T) string {
	t.Helper()
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\n---\n")
	return root
}

func TestCreateArtifactCmd_Plan(t *testing.T) {
	root := seedSpec(t)

	cmd := internalcmd.NewCreateArtifactCmd()
	cmd.SetArgs([]string{"plan", "--for", "SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		Created struct {
			Type string `json:"type"`
			Path string `json:"path"`
			For  string `json:"for"`
		} `json:"created"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.Created.Type != "plan" || decoded.Created.For != "SPEC-014" || decoded.Created.Path == "" {
		t.Errorf("created = %+v, unexpected", decoded.Created)
	}
}

func TestCreateArtifactCmd_UnrecognizedKind(t *testing.T) {
	root := seedSpec(t)

	cmd := internalcmd.NewCreateArtifactCmd()
	cmd.SetArgs([]string{"not-a-real-kind", "--for", "SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_argument")
}
