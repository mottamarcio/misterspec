package internalcmd_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/evidence"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

// writePrepareFixture builds a Spec with two Requirements, a Plan
// mentioning both, and two Tasks — TASK-001 (serves R1, no dependency,
// its own scope/verify) and TASK-002 (serves R2, depends on TASK-001).
func writePrepareFixture(t *testing.T, root string) {
	t.Helper()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n"+
			"## Requirements\n\n"+
			"### R1 — Refresh tokens must rotate on every use.\n\nDetailed rule for R1.\n\n"+
			"### R2 — Sessions must expire after 30 minutes.\n\nDetailed rule for R2.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/plan.md",
		"---\ntype: plan\nfor: SPEC-014\nstatus: draft\n---\n"+
			"## Implementation Sequence\n\nFirst cover R1, then R2.\n\n"+
			"## Unrelated Notes\n\nNothing to do with any requirement.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/tasks.md",
		"---\ntype: tasks\nfor: SPEC-014\n---\n# Tasks\n\n"+
			"## TASK-001 — Rotate refresh tokens\n\n"+
			"- [ ] Complete\n\n"+
			"Serves: SPEC-014:R1\n\n"+
			"Scope: internal/auth/refresh.go\n\n"+
			"Verify: go test ./internal/auth/... -run TestRefreshRotation\n\n"+
			"## TASK-002 — Expire idle sessions\n\n"+
			"- [ ] Complete\n\n"+
			"Serves: SPEC-014:R2\n\n"+
			"Depends on: TASK-001\n\n"+
			"Scope: internal/auth/session.go\n\n"+
			"Verify: go test ./internal/auth/... -run TestSessionExpiry\n")
}

type preparePlanSectionJSON struct {
	Heading             string   `json:"heading"`
	Content             string   `json:"content"`
	Fingerprint         string   `json:"fingerprint"`
	MatchedRequirements []string `json:"matched_requirements"`
}

type prepareRequirementJSON struct {
	Ref         string `json:"ref"`
	Content     string `json:"content"`
	Fingerprint string `json:"fingerprint"`
}

type prepareCodeItemJSON struct {
	Path        string `json:"path"`
	Heading     string `json:"heading"`
	Content     string `json:"content"`
	Fingerprint string `json:"fingerprint"`
}

type preparationJSON struct {
	Task              string                   `json:"task"`
	Heading           string                   `json:"heading"`
	Ready             bool                     `json:"ready"`
	Blockers          []string                 `json:"blockers"`
	Requirements      []prepareRequirementJSON `json:"requirements"`
	Scope             string                   `json:"scope"`
	Verify            string                   `json:"verify"`
	PlanSections      []preparePlanSectionJSON `json:"plan_sections"`
	CodeContext       []prepareCodeItemJSON    `json:"code_context"`
	CodeScopeNotFound []string                 `json:"code_scope_not_found"`
}

type prepareEnvelopeJSON struct {
	OK          bool             `json:"ok"`
	Preparation *preparationJSON `json:"preparation"`
	Message     string           `json:"message"`
}

func decodePrepareOutput(t *testing.T, output string) prepareEnvelopeJSON {
	t.Helper()
	var decoded prepareEnvelopeJSON
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	return decoded
}

// --- User Story 1: one-call preparation ---

func TestPrepareCmd_OneCallReturnsEverything(t *testing.T) {
	root := testutil.Project(t)
	writePrepareFixture(t, root)

	cmd := internalcmd.NewPrepareCmd()
	cmd.SetArgs([]string{"SPEC-014", "--task", "TASK-001", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	decoded := decodePrepareOutput(t, output)
	if decoded.Preparation == nil {
		t.Fatalf("preparation is nil (output: %s)", output)
	}
	p := decoded.Preparation
	if !p.Ready {
		t.Errorf("ready = false, want true for a dependency-free Task")
	}
	if len(p.Requirements) != 1 || p.Requirements[0].Ref != "SPEC-014:R1" {
		t.Errorf("requirements = %+v, want exactly [SPEC-014:R1]", p.Requirements)
	}
	if p.Scope != "internal/auth/refresh.go" {
		t.Errorf("scope = %q, want the declared scope", p.Scope)
	}
	if p.Verify != "go test ./internal/auth/... -run TestRefreshRotation" {
		t.Errorf("verify = %q, want the declared verification command", p.Verify)
	}
	if len(p.PlanSections) == 0 {
		t.Errorf("plan_sections is empty, want the \"Implementation Sequence\" section (mentions R1)")
	}
}

func TestPrepareCmd_TwoTasksReceiveDifferentContext(t *testing.T) {
	root := testutil.Project(t)
	writePrepareFixture(t, root)
	// Unblock TASK-002 so both prepare successfully as ready.
	markTaskComplete(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/tasks.md", "TASK-001")

	cmd1 := internalcmd.NewPrepareCmd()
	cmd1.SetArgs([]string{"SPEC-014", "--task", "TASK-001", "--dir", root})
	out1, _ := runCmd(cmd1)
	p1 := decodePrepareOutput(t, out1).Preparation

	cmd2 := internalcmd.NewPrepareCmd()
	cmd2.SetArgs([]string{"SPEC-014", "--task", "TASK-002", "--dir", root})
	out2, _ := runCmd(cmd2)
	p2 := decodePrepareOutput(t, out2).Preparation

	if p1 == nil || p2 == nil {
		t.Fatalf("preparation nil: p1=%v p2=%v", p1, p2)
	}
	if p1.Scope == p2.Scope {
		t.Errorf("both Tasks report the same scope %q, want different declared scopes", p1.Scope)
	}
	if len(p1.Requirements) == 0 || len(p2.Requirements) == 0 || p1.Requirements[0].Ref == p2.Requirements[0].Ref {
		t.Errorf("requirements not distinct: p1=%+v p2=%+v", p1.Requirements, p2.Requirements)
	}
}

func TestPrepareCmd_NeverModifiesProjectFiles(t *testing.T) {
	root := testutil.Project(t)
	writePrepareFixture(t, root)
	tasksPath := filepath.Join(root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/tasks.md")

	before, err := os.ReadFile(tasksPath)
	if err != nil {
		t.Fatalf("reading fixture before run: %v", err)
	}

	cmd := internalcmd.NewPrepareCmd()
	cmd.SetArgs([]string{"SPEC-014", "--task", "TASK-001", "--dir", root})
	if _, exitCode := runCmd(cmd); exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0", exitCode)
	}

	after, err := os.ReadFile(tasksPath)
	if err != nil {
		t.Fatalf("reading fixture after run: %v", err)
	}
	if string(before) != string(after) {
		t.Errorf("tasks.md content changed after prepare — it must be strictly read-only")
	}
}

// --- User Story 2: never present a blocked Task as ready ---

func TestPrepareCmd_IncompleteDependencyBlocks(t *testing.T) {
	root := testutil.Project(t)
	writePrepareFixture(t, root)
	// TASK-001 (TASK-002's own dependency) is still unchecked.

	cmd := internalcmd.NewPrepareCmd()
	cmd.SetArgs([]string{"SPEC-014", "--task", "TASK-002", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 10 {
		t.Fatalf("exitCode = %d, want 10 (output: %s)", exitCode, output)
	}
	decoded := decodePrepareOutput(t, output)
	if decoded.Preparation == nil {
		t.Fatalf("preparation is nil (output: %s)", output)
	}
	p := decoded.Preparation
	if p.Ready {
		t.Errorf("ready = true, want false")
	}
	if len(p.Blockers) != 1 || p.Blockers[0] != "SPEC-014:TASK-001" {
		t.Errorf("blockers = %v, want exactly [SPEC-014:TASK-001]", p.Blockers)
	}
}

func TestPrepareCmd_CompletingDependencyUnblocks(t *testing.T) {
	root := testutil.Project(t)
	writePrepareFixture(t, root)
	markTaskComplete(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/tasks.md", "TASK-001")

	cmd := internalcmd.NewPrepareCmd()
	cmd.SetArgs([]string{"SPEC-014", "--task", "TASK-002", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	decoded := decodePrepareOutput(t, output)
	if decoded.Preparation == nil || !decoded.Preparation.Ready {
		t.Fatalf("preparation = %+v, want Ready true", decoded.Preparation)
	}
}

func TestPrepareCmd_NoDependencyLineIsReadyByDefault(t *testing.T) {
	root := testutil.Project(t)
	writePrepareFixture(t, root)
	// TASK-001 has no "Depends on:" line at all.

	cmd := internalcmd.NewPrepareCmd()
	cmd.SetArgs([]string{"SPEC-014", "--task", "TASK-001", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	decoded := decodePrepareOutput(t, output)
	if decoded.Preparation == nil || !decoded.Preparation.Ready {
		t.Fatalf("preparation = %+v, want Ready true by default", decoded.Preparation)
	}
}

func TestPrepareCmd_DependencyCycleReturnsPromptlyAsBlocked(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Requirements\n\n### R1 — Something\n\nText.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/tasks.md",
		"---\ntype: tasks\nfor: SPEC-014\n---\n# Tasks\n\n"+
			"## TASK-003 — First half of the cycle\n\n- [ ] Complete\n\nDepends on: TASK-004\n\n"+
			"## TASK-004 — Second half of the cycle\n\n- [ ] Complete\n\nDepends on: TASK-003\n")

	done := make(chan struct{})
	var output string
	var exitCode int
	go func() {
		cmd := internalcmd.NewPrepareCmd()
		cmd.SetArgs([]string{"SPEC-014", "--task", "TASK-003", "--dir", root})
		output, exitCode = runCmd(cmd)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("prepare did not return within 5s — likely an infinite loop on the dependency cycle")
	}

	if exitCode != 10 {
		t.Fatalf("exitCode = %d, want 10 (output: %s)", exitCode, output)
	}
	decoded := decodePrepareOutput(t, output)
	if decoded.Preparation == nil || decoded.Preparation.Ready {
		t.Fatalf("preparation = %+v, want Ready false for a Task inside a dependency cycle", decoded.Preparation)
	}
}

// --- User Story 3: automatic selection ---

func TestPrepareCmd_AutoSelectsLowestReadyTask(t *testing.T) {
	root := testutil.Project(t)
	writePrepareFixture(t, root)
	// TASK-001 has no dependency (Ready); TASK-002 depends on TASK-001
	// (blocked). Auto-selection with no --task must pick TASK-001.

	cmd := internalcmd.NewPrepareCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	decoded := decodePrepareOutput(t, output)
	if decoded.Preparation == nil {
		t.Fatalf("preparation is nil (output: %s)", output)
	}
	if decoded.Preparation.Task != "SPEC-014:TASK-001" {
		t.Errorf("task = %q, want SPEC-014:TASK-001", decoded.Preparation.Task)
	}
}

func TestPrepareCmd_AutoSelectionReflectsNewStateAfterCompletion(t *testing.T) {
	root := testutil.Project(t)
	writePrepareFixture(t, root)

	first := internalcmd.NewPrepareCmd()
	first.SetArgs([]string{"SPEC-014", "--dir", root})
	out1, _ := runCmd(first)
	p1 := decodePrepareOutput(t, out1).Preparation
	if p1 == nil || p1.Task != "SPEC-014:TASK-001" {
		t.Fatalf("first selection = %+v, want SPEC-014:TASK-001", p1)
	}

	markTaskComplete(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/tasks.md", "TASK-001")

	second := internalcmd.NewPrepareCmd()
	second.SetArgs([]string{"SPEC-014", "--dir", root})
	out2, _ := runCmd(second)
	p2 := decodePrepareOutput(t, out2).Preparation
	if p2 == nil || p2.Task != "SPEC-014:TASK-002" {
		t.Fatalf("second selection = %+v, want SPEC-014:TASK-002 now that TASK-001 is complete", p2)
	}
}

func TestPrepareCmd_NoTaskReadyReturnsExplicitMessage(t *testing.T) {
	root := testutil.Project(t)
	writePrepareFixture(t, root)
	markTaskComplete(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/tasks.md", "TASK-001")
	markTaskComplete(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/tasks.md", "TASK-002")

	cmd := internalcmd.NewPrepareCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	decoded := decodePrepareOutput(t, output)
	if decoded.Preparation != nil {
		t.Errorf("preparation = %+v, want nil — every Task is complete", decoded.Preparation)
	}
	if decoded.Message == "" {
		t.Errorf("message is empty, want a clear explanation that no Task is ready")
	}
}

// markTaskComplete flips the given Task's own checkbox to "[x]" and
// appends a matching, valid Evidence-Result: pass record, simulating
// the Task having been finished and verified.
//
// Correction found during implementation (041-task-evidence-
// fingerprint): checking the box alone no longer means "complete" —
// every pre-existing test using this helper to simulate a finished Task
// now also needs real evidence, or Status stays "pending" under the
// redefined taskStatus (data-model.md "TaskInfo (extended)").
func markTaskComplete(t *testing.T, root, relPath, taskHeading string) {
	t.Helper()
	path := filepath.Join(root, relPath)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("markTaskComplete: reading %s: %v", path, err)
	}
	lines := strings.Split(string(data), "\n")
	inTarget := false
	checkboxLine := -1
	sectionEnd := len(lines)
	for i, line := range lines {
		if strings.HasPrefix(line, "## "+taskHeading) {
			inTarget = true
			continue
		}
		if inTarget && strings.HasPrefix(line, "## ") {
			sectionEnd = i
			break
		}
		if inTarget && checkboxLine == -1 && strings.HasPrefix(line, "- [ ]") {
			lines[i] = strings.Replace(line, "- [ ]", "- [x]", 1)
			checkboxLine = i
		}
	}
	if checkboxLine == -1 {
		t.Fatalf("markTaskComplete: no unchecked checkbox found under %q", taskHeading)
	}

	// Fingerprint this Task's own Section.Body exactly as
	// operations.TaskContentFingerprint/evidence.ContentFingerprint
	// would, over the real, currently-written file — not a hand-
	// reconstructed approximation.
	withCheckboxFlipped := strings.Join(lines, "\n")
	if err := os.WriteFile(path, []byte(withCheckboxFlipped), 0o644); err != nil {
		t.Fatalf("markTaskComplete: writing checkbox flip to %s: %v", path, err)
	}
	body, err := artifacts.ReadBody(path)
	if err != nil {
		t.Fatalf("markTaskComplete: reading body back from %s: %v", path, err)
	}
	doc := artifacts.ParseDocument(body)
	var fp string
	for _, section := range doc.Sections {
		if strings.HasPrefix(section.Heading, taskHeading) {
			fp = evidence.ContentFingerprint([]byte(section.Body))
			break
		}
	}
	if fp == "" {
		t.Fatalf("markTaskComplete: could not locate %q's own Section to fingerprint", taskHeading)
	}

	// No leading blank line: the section body already ends with one
	// before the next "## " heading (or EOF) — an extra blank line
	// here would itself become part of the re-parsed Section.Body and
	// (unlike an "Evidence-*:" line) is never stripped before
	// fingerprinting, so it would make the freshly-computed fingerprint
	// disagree with fp above (found during implementation: this exact
	// mismatch made the Task register as Stale instead of Verified).
	evidenceBlock := []string{
		"Evidence-Result: pass", "Evidence-Origin: declared", "Evidence-By: test",
		"Evidence-Fingerprint: " + fp,
	}
	final := append([]string{}, lines[:sectionEnd]...)
	// When the target Task is the file's last section (sectionEnd ==
	// len(lines)), lines' own final element is strings.Split's trailing
	// empty-string artifact of the file's single trailing newline, not
	// a deliberate blank separator line — unlike the real blank line
	// that precedes an actual next "## " heading. Appending evidenceBlock
	// after it would introduce an extra blank line the originally-
	// fingerprinted body never had, disagreeing with fp above (found
	// during implementation, the same class of bug as the leading-blank
	// fix on evidenceBlock itself, just at EOF instead of mid-file).
	if sectionEnd == len(lines) && len(final) > 0 && final[len(final)-1] == "" {
		final = final[:len(final)-1]
	}
	final = append(final, evidenceBlock...)
	final = append(final, lines[sectionEnd:]...)
	if err := os.WriteFile(path, []byte(strings.Join(final, "\n")), 0o644); err != nil {
		t.Fatalf("markTaskComplete: writing %s: %v", path, err)
	}
}

// --- User Story 3 (044-architecture-code-context-rules): code context ---

// TestPrepareCmd_CodeContextResolvedFromScope is
// 044-architecture-code-context-rules T025 (US3): TASK-001's own
// Scope: internal/auth/refresh.go resolves to that file's own
// declaration(s) in code_context, without a separate call.
func TestPrepareCmd_CodeContextResolvedFromScope(t *testing.T) {
	root := testutil.Project(t)
	writePrepareFixture(t, root)
	testutil.WriteFile(t, root, "go.mod", "module fixture.example\n\ngo 1.23\n")
	testutil.WriteFile(t, root, "internal/auth/refresh.go",
		"package auth\n\n// RotateRefreshToken rotates the token.\nfunc RotateRefreshToken() error {\n\treturn nil\n}\n")

	cmd := internalcmd.NewPrepareCmd()
	cmd.SetArgs([]string{"SPEC-014", "--task", "TASK-001", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	decoded := decodePrepareOutput(t, output)
	if decoded.Preparation == nil {
		t.Fatalf("preparation is nil (output: %s)", output)
	}
	p := decoded.Preparation
	if len(p.CodeContext) == 0 {
		t.Fatalf("code_context is empty, want at least one declaration from internal/auth/refresh.go")
	}
	found := false
	for _, item := range p.CodeContext {
		if item.Path == "internal/auth/refresh.go" && item.Heading == "RotateRefreshToken" {
			found = true
			if item.Content == "" || item.Fingerprint == "" {
				t.Errorf("item %+v missing content/fingerprint", item)
			}
		}
	}
	if !found {
		t.Errorf("code_context = %+v, want an item for RotateRefreshToken", p.CodeContext)
	}
	if p.CodeScopeNotFound == nil {
		t.Error("code_scope_not_found is null, want a (possibly empty) array")
	}
}

// TestPrepareCmd_CodeScopeNotFoundReportedExplicitly is
// 044-architecture-code-context-rules T025 (US3, spec Edge Case): a
// Scope path with no indexed .go file is named in code_scope_not_found.
func TestPrepareCmd_CodeScopeNotFoundReportedExplicitly(t *testing.T) {
	root := testutil.Project(t)
	writePrepareFixture(t, root)
	testutil.WriteFile(t, root, "go.mod", "module fixture.example\n\ngo 1.23\n")
	// internal/auth/refresh.go deliberately not written — TASK-001's
	// own Scope names a file that doesn't exist in this fixture.

	cmd := internalcmd.NewPrepareCmd()
	cmd.SetArgs([]string{"SPEC-014", "--task", "TASK-001", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	decoded := decodePrepareOutput(t, output)
	if decoded.Preparation == nil {
		t.Fatalf("preparation is nil (output: %s)", output)
	}
	p := decoded.Preparation
	if len(p.CodeScopeNotFound) != 1 || p.CodeScopeNotFound[0] != "internal/auth/refresh.go" {
		t.Errorf("code_scope_not_found = %+v, want [\"internal/auth/refresh.go\"]", p.CodeScopeNotFound)
	}
	if len(p.CodeContext) != 0 {
		t.Errorf("code_context = %+v, want none", p.CodeContext)
	}
}
