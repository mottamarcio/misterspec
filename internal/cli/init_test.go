package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

// runInitCmd executes newInitCmd() directly (white-box — newInitCmd is
// unexported, mirroring newRootCmd/newInternalCmd), returning its output
// and the exit code its RunE ultimately produced.
func runInitCmd(args []string) (output string, exitCode int) {
	cmd := newInitCmd()
	cmd.SetArgs(args)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		return buf.String(), 0
	}
	var exitErr *internalcmd.ExitCodeError
	if errors.As(err, &exitErr) {
		return buf.String(), exitErr.Code
	}
	return buf.String(), -1
}

func TestInitCmd_Success(t *testing.T) {
	target := filepath.Join(t.TempDir(), "newproject")

	output, exitCode := runInitCmd([]string{"--agent", "claude-code", "--dir", target})
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded struct {
		OK        bool `json:"ok"`
		Bootstrap struct {
			ProjectRoot   string `json:"project_root"`
			ConfigWritten bool   `json:"config_written"`
			Templates     []any  `json:"templates"`
			Agent         struct {
				AdapterID       string `json:"adapter_id"`
				IntegrationPath string `json:"integration_path"`
				Outcomes        []any  `json:"outcomes"`
			} `json:"agent"`
		} `json:"bootstrap"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if !decoded.OK {
		t.Error("ok = false, want true")
	}
	if !decoded.Bootstrap.ConfigWritten {
		t.Error("bootstrap.config_written = false, want true")
	}
	if len(decoded.Bootstrap.Templates) == 0 {
		t.Error("bootstrap.templates is empty, want at least one kit resource")
	}
	if decoded.Bootstrap.Agent.AdapterID != "claude-code" {
		t.Errorf("bootstrap.agent.adapter_id = %q, want %q", decoded.Bootstrap.Agent.AdapterID, "claude-code")
	}
	if len(decoded.Bootstrap.Agent.Outcomes) == 0 {
		t.Error("bootstrap.agent.outcomes is empty, want at least the README.md placeholder (research.md)")
	}

	if _, err := project.Detect(target); err != nil {
		t.Errorf("project.Detect(target) after init unexpected error: %v", err)
	}
}

func TestInitCmd_AlreadyInitialized(t *testing.T) {
	root := testutil.Project(t)

	output, exitCode := runInitCmd([]string{"--agent", "claude-code", "--dir", root})
	if exitCode != 5 {
		t.Fatalf("exitCode = %d, want 5 (output: %s)", exitCode, output)
	}

	var decoded struct {
		OK    bool `json:"ok"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.OK || decoded.Error.Code != "already_initialized" {
		t.Errorf("decoded = %+v, unexpected", decoded)
	}
}

func TestInitCmd_UnknownAgent(t *testing.T) {
	target := filepath.Join(t.TempDir(), "fresh")

	output, exitCode := runInitCmd([]string{"--agent", "no-such-agent", "--dir", target})
	if exitCode != 5 {
		t.Fatalf("exitCode = %d, want 5 (output: %s)", exitCode, output)
	}

	var decoded struct {
		OK    bool `json:"ok"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.OK || decoded.Error.Code != "unknown_agent" {
		t.Errorf("decoded = %+v, unexpected", decoded)
	}
}

func TestInitCmd_MissingAgentFlag(t *testing.T) {
	target := t.TempDir()

	output, exitCode := runInitCmd([]string{"--dir", target})
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}

	var decoded struct {
		OK    bool `json:"ok"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if decoded.OK || decoded.Error.Code != "invalid_argument" {
		t.Errorf("decoded = %+v, unexpected", decoded)
	}
}
