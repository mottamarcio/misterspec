package internalcmd_test

import (
	"encoding/json"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

// writeCaptureEvidenceFixture builds a minimal Spec with one Task,
// reused by capture_evidence_test.go's own test cases.
func writeCaptureEvidenceFixture(t *testing.T, root string) {
	t.Helper()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Requirements\n\n### R1 — First\n\nSomething.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/tasks.md",
		"---\ntype: tasks\nfor: SPEC-014\n---\n# Tasks\n\n## TASK-001 — Rotate refresh tokens\n\n- [ ] Complete\n\nServes: SPEC-014:R1\n")
	testutil.InitGitRepo(t, root)
}

type captureEvidenceFieldsJSON struct {
	Task             string `json:"task"`
	Result           string `json:"result"`
	Origin           string `json:"origin"`
	By               string `json:"by"`
	CapturedAt       string `json:"captured_at"`
	Command          string `json:"command"`
	GitRevision      string `json:"git_revision"`
	WorkingTree      string `json:"working_tree"`
	InputFingerprint string `json:"input_fingerprint"`
	Log              string `json:"log"`
}

type captureEvidenceEnvelopeJSON struct {
	OK       bool                       `json:"ok"`
	Evidence *captureEvidenceFieldsJSON `json:"evidence"`
	Error    *cliErrorJSON              `json:"error"`
}

type cliErrorJSON struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func decodeCaptureEvidenceOutput(t *testing.T, output string) captureEvidenceEnvelopeJSON {
	t.Helper()
	var decoded captureEvidenceEnvelopeJSON
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	return decoded
}

// TestCaptureEvidenceCmd_DeclaredOriginGoldenEnvelope is
// 041-task-evidence-fingerprint T016 (User Story 1): the declared-
// origin golden-envelope shape from contracts §4's first example.
func TestCaptureEvidenceCmd_DeclaredOriginGoldenEnvelope(t *testing.T) {
	root := testutil.Project(t)
	writeCaptureEvidenceFixture(t, root)

	cmd := internalcmd.NewCaptureEvidenceCmd()
	cmd.SetArgs([]string{"SPEC-014", "--task", "TASK-001", "--dir", root,
		"--origin", "declared", "--by", "reviewer:alice", "--result", "pass"})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	decoded := decodeCaptureEvidenceOutput(t, output)
	if !decoded.OK {
		t.Fatalf("ok = false (output: %s)", output)
	}
	e := decoded.Evidence
	if e == nil {
		t.Fatalf("evidence is nil (output: %s)", output)
	}
	if e.Result != "pass" {
		t.Errorf("result = %q, want %q", e.Result, "pass")
	}
	if e.Origin != "declared" {
		t.Errorf("origin = %q, want %q", e.Origin, "declared")
	}
	if e.By != "reviewer:alice" {
		t.Errorf("by = %q, want %q", e.By, "reviewer:alice")
	}
	if e.CapturedAt == "" {
		t.Error("captured_at is empty, want an RFC 3339 timestamp")
	}
	if e.Command != "" {
		t.Errorf("command = %q, want empty for declared origin", e.Command)
	}
	if e.WorkingTree != "clean" {
		t.Errorf("working_tree = %q, want %q for a freshly-committed fixture", e.WorkingTree, "clean")
	}
	if e.InputFingerprint == "" {
		t.Error("input_fingerprint is empty, want a computed sha256 digest")
	}
}

// TestCaptureEvidenceCmd_AutomatedOriginGoldenEnvelope is
// 041-task-evidence-fingerprint T030 (User Story 4): the automated-
// origin golden-envelope shape from contracts §4's second example.
func TestCaptureEvidenceCmd_AutomatedOriginGoldenEnvelope(t *testing.T) {
	root := testutil.Project(t)
	writeCaptureEvidenceFixture(t, root)

	cmd := internalcmd.NewCaptureEvidenceCmd()
	cmd.SetArgs([]string{"SPEC-014", "--task", "TASK-001", "--dir", root,
		"--origin", "automated", "--by", "go test",
		"--command", "sh", "--args", "-c", "--args", "echo hello"})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	decoded := decodeCaptureEvidenceOutput(t, output)
	if !decoded.OK {
		t.Fatalf("ok = false (output: %s)", output)
	}
	e := decoded.Evidence
	if e == nil {
		t.Fatalf("evidence is nil (output: %s)", output)
	}
	if e.Result != "pass" {
		t.Errorf("result = %q, want %q", e.Result, "pass")
	}
	if e.Origin != "automated" {
		t.Errorf("origin = %q, want %q", e.Origin, "automated")
	}
	if e.Command != "sh -c echo hello" {
		t.Errorf("command = %q, want the exact rendered command", e.Command)
	}
	if e.Log == "" {
		t.Error("log is empty, want a path under <specDir>/evidence/")
	}
}

// TestCaptureEvidenceCmd_AutomatedOriginRejectsDirEscapingRoot covers
// the error envelope for an --exec-dir that would resolve outside the
// project root.
func TestCaptureEvidenceCmd_AutomatedOriginRejectsDirEscapingRoot(t *testing.T) {
	root := testutil.Project(t)
	writeCaptureEvidenceFixture(t, root)

	cmd := internalcmd.NewCaptureEvidenceCmd()
	cmd.SetArgs([]string{"SPEC-014", "--task", "TASK-001", "--dir", root,
		"--origin", "automated", "--by", "test",
		"--command", "echo", "--exec-dir", "../../../../etc"})
	output, exitCode := runCmd(cmd)
	if exitCode == 0 {
		t.Fatalf("exitCode = 0, want non-zero for --exec-dir escaping the project root (output: %s)", output)
	}
	decoded := decodeCaptureEvidenceOutput(t, output)
	if decoded.OK {
		t.Errorf("ok = true, want false (output: %s)", output)
	}
	if decoded.Error == nil || decoded.Error.Code != "path_outside_project" {
		t.Errorf("error = %+v, want code %q", decoded.Error, "path_outside_project")
	}
}

// TestCaptureEvidenceCmd_AutomatedOriginWithoutCommandIsRejected
// covers the error envelope for --origin automated given without
// --command.
func TestCaptureEvidenceCmd_AutomatedOriginWithoutCommandIsRejected(t *testing.T) {
	root := testutil.Project(t)
	writeCaptureEvidenceFixture(t, root)

	cmd := internalcmd.NewCaptureEvidenceCmd()
	cmd.SetArgs([]string{"SPEC-014", "--task", "TASK-001", "--dir", root,
		"--origin", "automated", "--by", "test"})
	output, exitCode := runCmd(cmd)
	if exitCode == 0 {
		t.Fatalf("exitCode = 0, want non-zero for --origin automated with no --command (output: %s)", output)
	}
	decoded := decodeCaptureEvidenceOutput(t, output)
	if decoded.OK {
		t.Errorf("ok = true, want false (output: %s)", output)
	}
	if decoded.Error == nil || decoded.Error.Code != "invalid_argument" {
		t.Errorf("error = %+v, want code %q", decoded.Error, "invalid_argument")
	}
}

// TestCaptureEvidenceCmd_DeclaredOriginRejectsCommand covers contracts
// §2's "forbidden otherwise": Origin declared + Command set is an
// error, not a silently-ignored flag.
func TestCaptureEvidenceCmd_DeclaredOriginRejectsCommand(t *testing.T) {
	root := testutil.Project(t)
	writeCaptureEvidenceFixture(t, root)

	cmd := internalcmd.NewCaptureEvidenceCmd()
	cmd.SetArgs([]string{"SPEC-014", "--task", "TASK-001", "--dir", root,
		"--origin", "declared", "--by", "reviewer:alice", "--result", "pass", "--command", "go"})
	output, exitCode := runCmd(cmd)
	if exitCode == 0 {
		t.Fatalf("exitCode = 0, want non-zero for declared origin combined with --command (output: %s)", output)
	}

	decoded := decodeCaptureEvidenceOutput(t, output)
	if decoded.OK {
		t.Errorf("ok = true, want false (output: %s)", output)
	}
	if decoded.Error == nil || decoded.Error.Code != "invalid_argument" {
		t.Errorf("error = %+v, want code %q", decoded.Error, "invalid_argument")
	}
}
