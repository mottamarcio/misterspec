package internalcmd_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

// runGitTest runs a Git subcommand in root, failing the test on error
// (022-feature-branch-automation).
func runGitTest(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

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

// --- 022-feature-branch-automation ---

func TestCreateCmd_Program_NoGitKey(t *testing.T) {
	root := testutil.Project(t)
	testutil.InitGitRepo(t, root)

	cmd := internalcmd.NewCreateCmd()
	cmd.SetArgs([]string{"program", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	if strings.Contains(output, `"git"`) {
		t.Errorf("output contains a \"git\" key for a Program, want none: %s", output)
	}
}

func TestCreateCmd_Feature_GitBranchCreated(t *testing.T) {
	root := testutil.Project(t)
	testutil.InitGitRepo(t, root)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: draft\n---\n")

	cmd := internalcmd.NewCreateCmd()
	cmd.SetArgs([]string{"feature", "--parent", "PRG-001", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded struct {
		Created struct {
			Git struct {
				Branch  string `json:"branch"`
				Created bool   `json:"created"`
			} `json:"git"`
		} `json:"created"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.Created.Git.Branch != "feat/FEAT-001" || !decoded.Created.Git.Created {
		t.Errorf("created.git = %+v, want branch=feat/FEAT-001 created=true", decoded.Created.Git)
	}
}

func TestCreateCmd_Spec_GitWarningOnMismatch(t *testing.T) {
	root := testutil.Project(t)
	testutil.InitGitRepo(t, root)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: draft\n---\n")

	featCmd := internalcmd.NewCreateCmd()
	featCmd.SetArgs([]string{"feature", "--parent", "PRG-001", "--dir", root})
	if _, exitCode := runCmd(featCmd); exitCode != 0 {
		t.Fatalf("creating the parent Feature failed with exitCode = %d", exitCode)
	}

	// Feature creation just checked out feat/FEAT-001 — switch away
	// before creating the Spec, to exercise the mismatch path.
	runGitTest(t, root, "checkout", "-q", "-")

	specCmd := internalcmd.NewCreateCmd()
	specCmd.SetArgs([]string{"spec", "--parent", "FEAT-001", "--dir", root})
	output, exitCode := runCmd(specCmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded struct {
		Created struct {
			Git struct {
				Branch  string `json:"branch"`
				Warning string `json:"warning"`
			} `json:"git"`
		} `json:"created"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.Created.Git.Branch != "feat/FEAT-001" || decoded.Created.Git.Warning == "" {
		t.Errorf("created.git = %+v, want branch=feat/FEAT-001 and a non-empty warning", decoded.Created.Git)
	}
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
