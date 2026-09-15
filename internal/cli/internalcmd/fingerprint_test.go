package internalcmd_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestFingerprintCmd_Success(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/raw/architecture.md", "# Architecture\n")

	cmd := internalcmd.NewFingerprintCmd()
	cmd.SetArgs([]string{"ai/raw/architecture.md", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	var decoded struct {
		Source struct {
			Path        string `json:"path"`
			Algorithm   string `json:"algorithm"`
			Fingerprint string `json:"fingerprint"`
		} `json:"source"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.Source.Algorithm != "sha256" {
		t.Errorf("algorithm = %q, want %q", decoded.Source.Algorithm, "sha256")
	}
	if !strings.HasPrefix(decoded.Source.Fingerprint, "sha256:") {
		t.Errorf("fingerprint = %q, want a sha256:-prefixed digest", decoded.Source.Fingerprint)
	}
}

func TestFingerprintCmd_PathOutsideProject(t *testing.T) {
	root := testutil.Project(t)

	cmd := internalcmd.NewFingerprintCmd()
	cmd.SetArgs([]string{"../../../../etc/passwd", "--dir", root})
	output, exitCode := runCmd(cmd)

	if exitCode != 5 {
		t.Fatalf("exitCode = %d, want 5 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "path_outside_project")
}
