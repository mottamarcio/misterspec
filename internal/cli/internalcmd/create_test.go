package internalcmd_test

import (
	"encoding/json"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func decodeCreated(t *testing.T, output string) (id, typ, path string) {
	t.Helper()
	var decoded struct {
		Created struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			Path string `json:"path"`
		} `json:"created"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	return decoded.Created.ID, decoded.Created.Type, decoded.Created.Path
}

func TestCreateCmd_Program(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewCreateCmd()
	cmd.SetArgs([]string{"program", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	id, typ, path := decodeCreated(t, output)
	if id != "PRG-001" || typ != "program" || path == "" {
		t.Errorf("created = {%q, %q, %q}, unexpected", id, typ, path)
	}
}

func TestCreateCmd_FeatureUnderProgram(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: draft\n---\n")

	cmd := internalcmd.NewCreateCmd()
	cmd.SetArgs([]string{"feature", "--parent", "PRG-001", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	id, typ, _ := decodeCreated(t, output)
	if id != "FEAT-001" || typ != "feature" {
		t.Errorf("created = {%q, %q}, unexpected", id, typ)
	}
}

func TestCreateCmd_KnowledgeWithSlug(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewCreateCmd()
	cmd.SetArgs([]string{"knowledge", "--slug", "architecture-notes", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	id, typ, _ := decodeCreated(t, output)
	if id != "KNOW-001" || typ != "knowledge" {
		t.Errorf("created = {%q, %q}, unexpected", id, typ)
	}
}

func TestCreateCmd_InvalidParent(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewCreateCmd()
	cmd.SetArgs([]string{"feature", "--parent", "PRG-999", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 5 {
		t.Fatalf("exitCode = %d, want 5 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_parent")
}

func TestCreateCmd_UnsupportedTypeName(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewCreateCmd()
	cmd.SetArgs([]string{"not-a-real-type", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_argument")
}

func TestCreateCmd_InvalidSlug(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewCreateCmd()
	cmd.SetArgs([]string{"knowledge", "--slug", "", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_argument")
}
