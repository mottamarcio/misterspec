package cli_test

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

// buildMisterspecBinary compiles the real cmd/misterspec binary once per
// test, into a temp directory — proving the actual, shippable artifact
// behaves as documented (quickstart.md), not only the in-process RunE
// functions internalcmd's own tests already exercise.
func buildMisterspecBinary(t *testing.T) string {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "misterspec")
	build := exec.Command("go", "build", "-o", bin, "github.com/mottamarcio/misterspec/cmd/misterspec")
	var stderr bytes.Buffer
	build.Stderr = &stderr
	if err := build.Run(); err != nil {
		t.Fatalf("go build ./cmd/misterspec: %v\n%s", err, stderr.String())
	}
	return bin
}

func TestSubprocess_InternalResolve(t *testing.T) {
	bin := buildMisterspecBinary(t)
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\n---\n")

	cmd := exec.Command(bin, "internal", "resolve", "SPEC-014", "--dir", root)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("misterspec internal resolve: %v (output: %s)", err, out)
	}
	if cmd.ProcessState.ExitCode() != 0 {
		t.Fatalf("exit code = %d, want 0", cmd.ProcessState.ExitCode())
	}

	var decoded struct {
		OK     bool `json:"ok"`
		Entity struct {
			ID string `json:"id"`
		} `json:"entity"`
	}
	if jsonErr := json.Unmarshal(out, &decoded); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", jsonErr, out)
	}
	if !decoded.OK || decoded.Entity.ID != "SPEC-014" {
		t.Errorf("decoded = %+v, unexpected (raw: %s)", decoded, out)
	}
}

func TestSubprocess_InternalStatus(t *testing.T) {
	bin := buildMisterspecBinary(t)
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: draft\n---\n")

	cmd := exec.Command(bin, "internal", "status", "--dir", root)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("misterspec internal status: %v (output: %s)", err, out)
	}

	var decoded struct {
		OK     bool           `json:"ok"`
		Counts map[string]int `json:"counts"`
	}
	if jsonErr := json.Unmarshal(out, &decoded); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", jsonErr, out)
	}
	if !decoded.OK || decoded.Counts["program"] != 1 {
		t.Errorf("decoded = %+v, unexpected (raw: %s)", decoded, out)
	}
}

func TestSubprocess_InternalResolve_NotFound_ExitCode(t *testing.T) {
	bin := buildMisterspecBinary(t)
	root := testutil.Project(t)

	cmd := exec.Command(bin, "internal", "resolve", "SPEC-999", "--dir", root)
	out, err := cmd.Output()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected an *exec.ExitError, got %v (output: %s)", err, out)
	}
	if exitErr.ExitCode() != 3 {
		t.Errorf("exit code = %d, want 3", exitErr.ExitCode())
	}

	var decoded struct {
		OK    bool `json:"ok"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if jsonErr := json.Unmarshal(out, &decoded); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", jsonErr, out)
	}
	if decoded.OK || decoded.Error.Code != "entity_not_found" {
		t.Errorf("decoded = %+v, unexpected (raw: %s)", decoded, out)
	}
}

func TestSubprocess_Init(t *testing.T) {
	bin := buildMisterspecBinary(t)
	target := filepath.Join(t.TempDir(), "bootstrapped")

	cmd := exec.Command(bin, "init", "--agent", "claude-code", "--dir", target)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("misterspec init: %v (output: %s)", err, out)
	}
	if cmd.ProcessState.ExitCode() != 0 {
		t.Fatalf("exit code = %d, want 0", cmd.ProcessState.ExitCode())
	}

	var decoded struct {
		OK        bool `json:"ok"`
		Bootstrap struct {
			ConfigWritten bool `json:"config_written"`
			Agent         struct {
				AdapterID string `json:"adapter_id"`
			} `json:"agent"`
		} `json:"bootstrap"`
	}
	if jsonErr := json.Unmarshal(out, &decoded); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", jsonErr, out)
	}
	if !decoded.OK || !decoded.Bootstrap.ConfigWritten || decoded.Bootstrap.Agent.AdapterID != "claude-code" {
		t.Errorf("decoded = %+v, unexpected (raw: %s)", decoded, out)
	}

	if _, detectErr := project.Detect(target); detectErr != nil {
		t.Errorf("project.Detect(target) after real `misterspec init`: %v", detectErr)
	}
}

func TestSubprocess_HelpOutputExcludesInternal(t *testing.T) {
	bin := buildMisterspecBinary(t)

	out, err := exec.Command(bin, "--help").CombinedOutput()
	if err != nil {
		t.Fatalf("misterspec --help: %v (output: %s)", err, out)
	}
	if strings.Contains(string(out), "internal") {
		t.Errorf("`misterspec --help` output mentions \"internal\", want it hidden:\n%s", out)
	}
	if !strings.Contains(string(out), "init") {
		t.Errorf("`misterspec --help` output does not mention \"init\":\n%s", out)
	}
}
