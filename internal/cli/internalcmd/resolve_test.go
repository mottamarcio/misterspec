package internalcmd_test

import (
	"encoding/json"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestResolveCmd_Success(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\n---\n")

	cmd := internalcmd.NewResolveCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		OK     bool `json:"ok"`
		Entity struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			Path string `json:"path"`
		} `json:"entity"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if !decoded.OK {
		t.Error("ok = false, want true")
	}
	if decoded.Entity.ID != "SPEC-014" || decoded.Entity.Type != "spec" {
		t.Errorf("entity = %+v, unexpected", decoded.Entity)
	}
}

func TestResolveCmd_NotFound(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewResolveCmd()
	cmd.SetArgs([]string{"SPEC-999", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 3 {
		t.Fatalf("exitCode = %d, want 3 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "entity_not_found")
}

func TestResolveCmd_Ambiguous(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-002/features/FEAT-005/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-005\n---\n")

	cmd := internalcmd.NewResolveCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 3 {
		t.Fatalf("exitCode = %d, want 3 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "entity_ambiguous")
}

func TestResolveCmd_ProjectNotInitialized(t *testing.T) {
	dir := t.TempDir()

	cmd := internalcmd.NewResolveCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", dir})
	output, exitCode := runCmd(cmd)

	if exitCode != 6 {
		t.Fatalf("exitCode = %d, want 6 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "project_not_initialized")
}
