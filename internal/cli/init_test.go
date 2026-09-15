package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/misterspec/internal/agents"
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
			ProjectRoot   string   `json:"project_root"`
			ConfigWritten bool     `json:"config_written"`
			Directories   []string `json:"directories"`
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
	if decoded.Bootstrap.Agent.AdapterID != "claude-code" {
		t.Errorf("bootstrap.agent.adapter_id = %q, want %q", decoded.Bootstrap.Agent.AdapterID, "claude-code")
	}
	if len(decoded.Bootstrap.Agent.Outcomes) == 0 {
		t.Error("bootstrap.agent.outcomes is empty, want at least the README.md placeholder (research.md)")
	}
	if len(decoded.Bootstrap.Directories) != 6 {
		t.Errorf("bootstrap.directories = %v, want 6 entries (021-init-scaffold-distribution)", decoded.Bootstrap.Directories)
	}
	for _, want := range []string{"ai", "ai/raw", "ai/knowledge", "ai/memory", "ai/memory/learnings", "ai/programs"} {
		found := false
		for _, got := range decoded.Bootstrap.Directories {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("bootstrap.directories = %v, missing %q", decoded.Bootstrap.Directories, want)
		}
	}

	if _, err := project.Detect(target); err != nil {
		t.Errorf("project.Detect(target) after init unexpected error: %v", err)
	}

	// FR-001: the JSON payload no longer names a "templates" key at all
	// (021-init-scaffold-distribution/research.md #4 — removed, not
	// left as an empty array).
	if strings.Contains(output, `"templates"`) {
		t.Errorf("output still contains a \"templates\" key, want it removed entirely: %s", output)
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
	// No --agent and no interactive terminal — explicitly stubbed
	// (010-interactive-init-tui) so this test's result never depends
	// on go test's own actual TTY status.
	restore := stubInteractive(false)
	defer restore()

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

// stubInteractive overrides isInteractiveTerminal for the duration of a
// test, returning a func that restores the original
// (010-interactive-init-tui's overridable-var testability design,
// research.md).
func stubInteractive(interactive bool) (restore func()) {
	original := isInteractiveTerminal
	isInteractiveTerminal = func() bool { return interactive }
	return func() { isInteractiveTerminal = original }
}

// stubTUIRunInit overrides tuiRunInit for the duration of a test,
// recording the dir it was invoked with (empty string if never
// invoked), without ever launching a real Bubble Tea program.
func stubTUIRunInit(t *testing.T) (calledWith *string) {
	t.Helper()
	original := tuiRunInit
	var gotDir string
	tuiRunInit = func(dir string, registry *agents.Registry, skills fs.FS, opts ...tea.ProgramOption) error {
		gotDir = dir
		return nil
	}
	t.Cleanup(func() { tuiRunInit = original })
	return &gotDir
}

func TestInitCmd_InteractiveNoAgent_LaunchesTUI(t *testing.T) {
	restore := stubInteractive(true)
	defer restore()
	gotDir := stubTUIRunInit(t)

	target := t.TempDir()
	output, exitCode := runInitCmd([]string{"--dir", target})

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	if *gotDir != target {
		t.Errorf("tuiRunInit called with dir = %q, want %q", *gotDir, target)
	}
}

func TestInitCmd_NonInteractiveNoAgent_FailsWithSameShapeAsMissingAgent(t *testing.T) {
	restore := stubInteractive(false)
	defer restore()

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

func TestInitCmd_AgentProvided_NeverLaunchesTUI(t *testing.T) {
	// Even when interactive, --agent means "I already know what I
	// want" — the non-interactive path must run, unchanged, and
	// tuiRunInit must never be called.
	restore := stubInteractive(true)
	defer restore()
	gotDir := stubTUIRunInit(t)

	target := filepath.Join(t.TempDir(), "newproject")
	output, exitCode := runInitCmd([]string{"--agent", "claude-code", "--dir", target})

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	if *gotDir != "" {
		t.Errorf("tuiRunInit was invoked (dir=%q) despite --agent being provided", *gotDir)
	}
	if _, err := project.Detect(target); err != nil {
		t.Errorf("project.Detect(target) after --agent path unexpected error: %v", err)
	}
}
