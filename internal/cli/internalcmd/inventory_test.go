package internalcmd_test

import (
	"encoding/json"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestInventoryCmd_Populated(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/raw/prd.pdf", "not-really-a-pdf")
	testutil.WriteFile(t, root, "ai/raw/architecture.md", "# Architecture\n")

	cmd := internalcmd.NewInventoryCmd()
	cmd.SetArgs([]string{"ai/raw", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		Files []struct {
			Path      string `json:"path"`
			Extension string `json:"extension"`
			Size      int64  `json:"size"`
		} `json:"files"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if len(decoded.Files) != 2 {
		t.Fatalf("len(files) = %d, want 2 (%+v)", len(decoded.Files), decoded.Files)
	}
}

func TestInventoryCmd_EmptyDirectory(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewInventoryCmd()
	cmd.SetArgs([]string{"ai/raw", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		OK    bool  `json:"ok"`
		Files []any `json:"files"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if !decoded.OK {
		t.Error("ok = false, want true — an empty/non-existent directory is not an error")
	}
	if decoded.Files == nil {
		t.Error("files = null, want an empty array")
	}
}

func TestInventoryCmd_PathOutsideProject(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewInventoryCmd()
	cmd.SetArgs([]string{"../../../../etc", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 5 {
		t.Fatalf("exitCode = %d, want 5 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "path_outside_project")
}
