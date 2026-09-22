package operations_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func evidenceFixture(t *testing.T, root string) (spec ids.EntityID, specDir string) {
	t.Helper()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	specDir = "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014"
	testutil.WriteFile(t, root, specDir+"/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Requirements\n\n### R1 — First\n\nSomething.\n")
	testutil.WriteFile(t, root, specDir+"/tasks.md",
		"---\ntype: tasks\nfor: SPEC-014\n---\n# Tasks\n\n## TASK-001 — Rotate refresh tokens\n\n- [ ] Complete\n\nServes: SPEC-014:R1\n")
	testutil.InitGitRepo(t, root)
	return ids.EntityID{Type: ids.Spec, Prefix: "SPEC", Number: 14, Width: 3}, specDir
}

func taskEntityID() ids.EntityID {
	return ids.EntityID{Type: ids.Task, Prefix: "TASK", Number: 1, Width: 3}
}

// TestCaptureEvidence_DeclaredOriginPass is
// 041-task-evidence-fingerprint T013 (User Story 1).
func TestCaptureEvidence_DeclaredOriginPass(t *testing.T) {
	root := testutil.Project(t)
	spec, specDir := evidenceFixture(t, root)

	fields, err := operations.CaptureEvidence(root, testConfig(), operations.EvidenceCaptureRequest{
		Spec: spec, Task: taskEntityID(), SpecDir: specDir,
		Origin: "declared", By: "reviewer:alice", Result: "pass",
	})
	if err != nil {
		t.Fatalf("CaptureEvidence() unexpected error: %v", err)
	}

	if fields.Result != "pass" {
		t.Errorf("Result = %q, want %q", fields.Result, "pass")
	}
	if fields.Origin != "declared" {
		t.Errorf("Origin = %q, want %q", fields.Origin, "declared")
	}
	if fields.By != "reviewer:alice" {
		t.Errorf("By = %q, want %q", fields.By, "reviewer:alice")
	}
	if _, err := time.Parse(time.RFC3339, fields.CapturedAt); err != nil {
		t.Errorf("CapturedAt = %q, not a valid RFC3339 timestamp: %v", fields.CapturedAt, err)
	}
	if fields.GitRevision == "" {
		t.Error("GitRevision is empty, want a real commit SHA (fixture is a real repo)")
	}
	if fields.WorkingTree != "clean" {
		t.Errorf("WorkingTree = %q, want %q for a freshly-committed fixture", fields.WorkingTree, "clean")
	}
	if fields.Fingerprint == "" || !strings.HasPrefix(fields.Fingerprint, "sha256:") {
		t.Errorf("Fingerprint = %q, want a sha256:<hex> digest", fields.Fingerprint)
	}
	if fields.Command != "" {
		t.Errorf("Command = %q, want empty for declared origin", fields.Command)
	}
	if fields.Log != "" {
		t.Errorf("Log = %q, want empty for declared origin", fields.Log)
	}
}

// TestCaptureEvidence_DeclaredOriginFail covers the negative Result.
func TestCaptureEvidence_DeclaredOriginFail(t *testing.T) {
	root := testutil.Project(t)
	spec, specDir := evidenceFixture(t, root)

	fields, err := operations.CaptureEvidence(root, testConfig(), operations.EvidenceCaptureRequest{
		Spec: spec, Task: taskEntityID(), SpecDir: specDir,
		Origin: "declared", By: "reviewer:alice", Result: "fail",
	})
	if err != nil {
		t.Fatalf("CaptureEvidence() unexpected error: %v", err)
	}
	if fields.Result != "fail" {
		t.Errorf("Result = %q, want %q", fields.Result, "fail")
	}
}

// TestCaptureEvidence_DeclaredOriginRejectsCommand covers contracts
// §2's "forbidden otherwise": Command set with Origin declared is a
// malformed request, not silently ignored.
func TestCaptureEvidence_DeclaredOriginRejectsCommand(t *testing.T) {
	root := testutil.Project(t)
	spec, specDir := evidenceFixture(t, root)

	_, err := operations.CaptureEvidence(root, testConfig(), operations.EvidenceCaptureRequest{
		Spec: spec, Task: taskEntityID(), SpecDir: specDir,
		Origin: "declared", By: "reviewer:alice", Result: "pass", Command: "go",
	})
	if err == nil {
		t.Fatal("CaptureEvidence() error = nil, want an error for declared origin combined with Command")
	}
}

// TestCaptureEvidence_DeclaredOriginRequiresResult covers the mirror
// case: declared origin with no Result is also rejected.
func TestCaptureEvidence_DeclaredOriginRequiresResult(t *testing.T) {
	root := testutil.Project(t)
	spec, specDir := evidenceFixture(t, root)

	_, err := operations.CaptureEvidence(root, testConfig(), operations.EvidenceCaptureRequest{
		Spec: spec, Task: taskEntityID(), SpecDir: specDir,
		Origin: "declared", By: "reviewer:alice",
	})
	if err == nil {
		t.Fatal("CaptureEvidence() error = nil, want an error for declared origin with no Result")
	}
}

// TestCaptureEvidence_AutomatedOriginRunsExactCommand is
// 041-task-evidence-fingerprint T028 (User Story 4): Origin ==
// "automated" runs exactly the given Command/Args inside Dir, derives
// Result from the exit code, writes the command's combined output to a
// new log file under <specDir>/evidence/, and returns that path in Log.
func TestCaptureEvidence_AutomatedOriginRunsExactCommand(t *testing.T) {
	root := testutil.Project(t)
	spec, specDir := evidenceFixture(t, root)

	fields, err := operations.CaptureEvidence(root, testConfig(), operations.EvidenceCaptureRequest{
		Spec: spec, Task: taskEntityID(), SpecDir: specDir,
		Origin: "automated", By: "echo test",
		Command: "sh", Args: []string{"-c", "echo hello-from-capture-evidence"},
	})
	if err != nil {
		t.Fatalf("CaptureEvidence() unexpected error: %v", err)
	}
	if fields.Result != "pass" {
		t.Errorf("Result = %q, want %q for exit code 0", fields.Result, "pass")
	}
	if fields.Command != "sh -c echo hello-from-capture-evidence" {
		t.Errorf("Command = %q, want the exact rendered command", fields.Command)
	}
	if fields.Log == "" {
		t.Fatal("Log is empty, want a path under <specDir>/evidence/")
	}
	if !strings.HasPrefix(fields.Log, specDir+"/evidence/") {
		t.Errorf("Log = %q, want a path under %q", fields.Log, specDir+"/evidence/")
	}
	logContent, err := os.ReadFile(filepath.Join(root, fields.Log))
	if err != nil {
		t.Fatalf("reading log file %s: %v", fields.Log, err)
	}
	if !strings.Contains(string(logContent), "hello-from-capture-evidence") {
		t.Errorf("log content = %q, want it to contain the command's own stdout", logContent)
	}
}

// TestCaptureEvidence_AutomatedOriginNonZeroExitIsFail covers a
// command that runs but fails — a successful *capture* of a failure,
// not an operational error.
func TestCaptureEvidence_AutomatedOriginNonZeroExitIsFail(t *testing.T) {
	root := testutil.Project(t)
	spec, specDir := evidenceFixture(t, root)

	fields, err := operations.CaptureEvidence(root, testConfig(), operations.EvidenceCaptureRequest{
		Spec: spec, Task: taskEntityID(), SpecDir: specDir,
		Origin: "automated", By: "failing command",
		Command: "sh", Args: []string{"-c", "exit 1"},
	})
	if err != nil {
		t.Fatalf("CaptureEvidence() unexpected error: %v", err)
	}
	if fields.Result != "fail" {
		t.Errorf("Result = %q, want %q for a non-zero exit code", fields.Result, "fail")
	}
}

// TestCaptureEvidence_AutomatedOriginTimeoutIsFail covers a command
// that exceeds its own Timeout — terminated and reported as "fail",
// never left hanging.
func TestCaptureEvidence_AutomatedOriginTimeoutIsFail(t *testing.T) {
	root := testutil.Project(t)
	spec, specDir := evidenceFixture(t, root)

	fields, err := operations.CaptureEvidence(root, testConfig(), operations.EvidenceCaptureRequest{
		Spec: spec, Task: taskEntityID(), SpecDir: specDir,
		Origin: "automated", By: "slow command",
		Command: "sleep", Args: []string{"5"}, Timeout: 100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("CaptureEvidence() unexpected error: %v", err)
	}
	if fields.Result != "fail" {
		t.Errorf("Result = %q, want %q for a command that exceeded its timeout", fields.Result, "fail")
	}
}

// TestCaptureEvidence_AutomatedOriginRejectsDirEscapingRoot covers
// Principle VIII: a Dir that would resolve outside the project root is
// rejected before any subprocess ever starts.
func TestCaptureEvidence_AutomatedOriginRejectsDirEscapingRoot(t *testing.T) {
	root := testutil.Project(t)
	spec, specDir := evidenceFixture(t, root)

	_, err := operations.CaptureEvidence(root, testConfig(), operations.EvidenceCaptureRequest{
		Spec: spec, Task: taskEntityID(), SpecDir: specDir,
		Origin: "automated", By: "test",
		Command: "echo", Args: []string{"should never run"},
		Dir: "../../../../etc",
	})
	if !errors.Is(err, artifacts.ErrPathOutsideProject) {
		t.Fatalf("CaptureEvidence() error = %v, want errors.Is(err, artifacts.ErrPathOutsideProject)", err)
	}
}

// TestCaptureEvidence_AutomatedOriginWithoutCommandIsRejected covers
// the mirror case of the declared-origin checks: automated origin with
// no Command is a malformed request.
func TestCaptureEvidence_AutomatedOriginWithoutCommandIsRejected(t *testing.T) {
	root := testutil.Project(t)
	spec, specDir := evidenceFixture(t, root)

	_, err := operations.CaptureEvidence(root, testConfig(), operations.EvidenceCaptureRequest{
		Spec: spec, Task: taskEntityID(), SpecDir: specDir,
		Origin: "automated", By: "test",
	})
	if err == nil {
		t.Fatal("CaptureEvidence() error = nil, want an error for automated origin with no Command")
	}
}

// TestCaptureEvidence_AutomatedOriginReflectsDirtyWorkingTree is
// 041-task-evidence-fingerprint T029 (quickstart.md §5): capturing
// automated evidence while an unrelated file has an uncommitted change
// yields WorkingTree == "dirty"; after reverting, "clean".
func TestCaptureEvidence_AutomatedOriginReflectsDirtyWorkingTree(t *testing.T) {
	root := testutil.Project(t)
	spec, specDir := evidenceFixture(t, root)

	uncommitted := filepath.Join(root, "untracked.txt")
	if err := os.WriteFile(uncommitted, []byte("local change"), 0o644); err != nil {
		t.Fatalf("writing uncommitted file: %v", err)
	}

	fields, err := operations.CaptureEvidence(root, testConfig(), operations.EvidenceCaptureRequest{
		Spec: spec, Task: taskEntityID(), SpecDir: specDir,
		Origin: "automated", By: "test", Command: "true",
	})
	if err != nil {
		t.Fatalf("CaptureEvidence() unexpected error: %v", err)
	}
	if fields.WorkingTree != "dirty" {
		t.Errorf("WorkingTree = %q, want %q with an uncommitted file present", fields.WorkingTree, "dirty")
	}

	if err := os.Remove(uncommitted); err != nil {
		t.Fatalf("removing uncommitted file: %v", err)
	}
	// The first capture above wrote its own new evidence log file,
	// which is itself now an untracked file — commit it (and anything
	// else pending) first, so this second capture's own "clean" check
	// reflects only untracked.txt's own removal, not the first
	// capture's own real, correct side effect (found during
	// implementation: the initial test omitted this and failed for the
	// right underlying reason — the working tree genuinely was still
	// dirty because of the log CaptureEvidence itself had just written).
	commitAll(t, root)

	fields, err = operations.CaptureEvidence(root, testConfig(), operations.EvidenceCaptureRequest{
		Spec: spec, Task: taskEntityID(), SpecDir: specDir,
		Origin: "automated", By: "test", Command: "true",
	})
	if err != nil {
		t.Fatalf("CaptureEvidence() unexpected error: %v", err)
	}
	if fields.WorkingTree != "clean" {
		t.Errorf("WorkingTree = %q, want %q after reverting the uncommitted change and committing the prior capture's own log", fields.WorkingTree, "clean")
	}
}

// commitAll stages and commits every pending change in root — a
// small test-only helper for tests that need a genuinely clean working
// tree partway through, after CaptureEvidence's own log-writing side
// effect.
func commitAll(t *testing.T, root string) {
	t.Helper()
	add := exec.Command("git", "add", "-A")
	add.Dir = root
	if out, err := add.CombinedOutput(); err != nil {
		t.Fatalf("git add -A: %v\n%s", err, out)
	}
	commit := exec.Command("git", "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-q", "-m", "commit pending evidence log")
	commit.Dir = root
	if out, err := commit.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
}
