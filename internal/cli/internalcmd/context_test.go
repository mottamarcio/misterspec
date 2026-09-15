package internalcmd_test

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

// writeContextFixture builds the shared 015/016-style fixture: SPEC-014
// depends on SPEC-011, links to KNOW-003, and the project has a
// Constitution — used across every User Story in this file.
func writeContextFixture(t *testing.T, root string) {
	t.Helper()
	testutil.WriteFile(t, root, "ai/memory/constitution.md",
		"---\ntype: constitution\n---\n## Principles\n\nFilesystem is the source of truth.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-011/spec.md",
		"---\nid: SPEC-011\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nThe dependency target.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-x.md",
		"---\nid: KNOW-003\ntype: knowledge\nstatus: active\n---\n## Summary\n\nAuthentication constraints.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-011\nsupersedes: []\n---\n## Requirements\n\nRefresh token rotation. See [[KNOW-003]].\n")
}

type contextItemJSON struct {
	Path    string   `json:"path"`
	Heading string   `json:"heading"`
	Tier    string   `json:"tier"`
	Reasons []string `json:"reasons"`
	Score   int      `json:"score"`
	Tokens  int      `json:"tokens"`
}

type contextDiagnosticsJSON struct {
	CandidatesConsidered int     `json:"candidates_considered"`
	ItemsSelected        int     `json:"items_selected"`
	TokensAvailable      int     `json:"tokens_available"`
	TokensSelected       int     `json:"tokens_selected"`
	TokensExcluded       int     `json:"tokens_excluded"`
	ReductionPercent     float64 `json:"reduction_percent"`
}

type contextResultJSON struct {
	Target          string                 `json:"target"`
	Intent          string                 `json:"intent"`
	Budget          int                    `json:"budget"`
	EstimatedTokens int                    `json:"estimated_tokens"`
	BudgetExceeded  bool                   `json:"budget_exceeded"`
	Overage         int                    `json:"overage"`
	Items           []contextItemJSON      `json:"items"`
	Diagnostics     contextDiagnosticsJSON `json:"diagnostics"`
	Rendered        *string                `json:"rendered"`
}

type contextEnvelopeJSON struct {
	OK      bool              `json:"ok"`
	Context contextResultJSON `json:"context"`
}

func decodeContextOutput(t *testing.T, output string) contextEnvelopeJSON {
	t.Helper()
	var decoded contextEnvelopeJSON
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	return decoded
}

func TestContextCmd_Success(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--intent", "implementation", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	decoded := decodeContextOutput(t, output)
	if !decoded.OK {
		t.Fatalf("ok = false, want true (output: %s)", output)
	}
	if decoded.Context.Target != "SPEC-014" {
		t.Errorf("context.target = %q, want SPEC-014", decoded.Context.Target)
	}
	if decoded.Context.Intent != "implementation" {
		t.Errorf("context.intent = %q, want implementation", decoded.Context.Intent)
	}
	if len(decoded.Context.Items) == 0 {
		t.Fatalf("context.items is empty, want at least the mandatory target/constitution entries")
	}
	foundTarget := false
	for _, item := range decoded.Context.Items {
		for _, r := range item.Reasons {
			if r == "target" {
				foundTarget = true
			}
		}
	}
	if !foundTarget {
		t.Errorf("context.items = %+v, want an item reasoned \"target\"", decoded.Context.Items)
	}
	if decoded.Context.Diagnostics.CandidatesConsidered == 0 {
		t.Errorf("context.diagnostics.candidates_considered = 0, want > 0")
	}
	if decoded.Context.Rendered != nil {
		t.Errorf("context.rendered = %v, want nil when --render was not passed", decoded.Context.Rendered)
	}
}

func TestContextCmd_TaskQueryBudgetPassThrough(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{
		"SPEC-014",
		"--intent", "implementation",
		"--task", "implement refresh token rotation",
		"--query", "refresh token rotation",
		"--budget", "500",
		"--dir", root,
	})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	decoded := decodeContextOutput(t, output)
	if decoded.Context.Budget != 500 {
		t.Errorf("context.budget = %d, want 500", decoded.Context.Budget)
	}
	if decoded.Context.EstimatedTokens > 500 && !decoded.Context.BudgetExceeded {
		t.Errorf("context.estimated_tokens = %d exceeds budget 500 but budget_exceeded = false", decoded.Context.EstimatedTokens)
	}
}

func TestContextCmd_OmittedFlagsUseDefaults(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	decoded := decodeContextOutput(t, output)
	if decoded.Context.Intent != "" {
		t.Errorf("context.intent = %q, want empty when --intent omitted", decoded.Context.Intent)
	}
	if decoded.Context.Budget != 6000 {
		t.Errorf("context.budget = %d, want 6000 (contextengine.DefaultBudget)", decoded.Context.Budget)
	}
}

func TestContextCmd_RepeatedRequestIsDeterministic(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	args := []string{"SPEC-014", "--intent", "implementation", "--query", "refresh token rotation", "--dir", root}

	cmd1 := internalcmd.NewContextCmd()
	cmd1.SetArgs(args)
	output1, exitCode1 := runCmd(cmd1)
	if exitCode1 != 0 {
		t.Fatalf("first invocation exitCode = %d, want 0 (output: %s)", exitCode1, output1)
	}

	cmd2 := internalcmd.NewContextCmd()
	cmd2.SetArgs(args)
	output2, exitCode2 := runCmd(cmd2)
	if exitCode2 != 0 {
		t.Fatalf("second invocation exitCode = %d, want 0 (output: %s)", exitCode2, output2)
	}

	if output1 != output2 {
		t.Errorf("output1 != output2, want byte-for-byte identical:\n1: %s\n2: %s", output1, output2)
	}
}

// --- User Story 2: clear, actionable errors ---

func TestContextCmd_UnknownTarget(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-999", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 3 {
		t.Fatalf("exitCode = %d, want 3 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "entity_not_found")
}

func TestContextCmd_AmbiguousTarget(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: draft\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-002/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: draft\nparent: PRG-002\n---\n")

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"FEAT-004", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 3 {
		t.Fatalf("exitCode = %d, want 3 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "entity_ambiguous")
}

func TestContextCmd_UnsupportedIntent(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--intent", "bogus", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "unsupported_intent")
}

func TestContextCmd_NonNumericBudget(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--budget", "notanumber", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_argument")
}

func TestContextCmd_ErrorResponsesNeverLookLikeSuccess(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-999", "--dir", root})
	output, _ := runCmd(cmd)

	var decoded map[string]any
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, output)
	}
	if _, hasContext := decoded["context"]; hasContext {
		t.Errorf("error response unexpectedly has a \"context\" key: %s", output)
	}
}

// --- User Story 3: transparent index readiness ---

func TestContextCmd_NoCacheYetStillSucceeds(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cachePath := root + "/.misterspec/cache/context.db"
	if _, err := os.Stat(cachePath); err == nil {
		t.Fatalf("cache unexpectedly already exists at %s", cachePath)
	}

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	if _, err := os.Stat(cachePath); err != nil {
		t.Errorf("cache still does not exist at %s after invocation: %v", cachePath, err)
	}
}

func TestContextCmd_ArtifactChangeIsReflectedOnNextInvocation(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd1 := internalcmd.NewContextCmd()
	cmd1.SetArgs([]string{"SPEC-014", "--query", "rotation", "--dir", root})
	if _, exitCode := runCmd(cmd1); exitCode != 0 {
		t.Fatalf("first invocation exitCode = %d, want 0", exitCode)
	}

	testutil.WriteFile(t, root, "ai/knowledge/KNOW-009-new.md",
		"---\nid: KNOW-009\ntype: knowledge\nstatus: active\n---\n## Summary\n\nA brand new rotation constraint.\n")

	cmd2 := internalcmd.NewContextCmd()
	cmd2.SetArgs([]string{"SPEC-014", "--query", "rotation", "--dir", root})
	output2, exitCode2 := runCmd(cmd2)
	if exitCode2 != 0 {
		t.Fatalf("second invocation exitCode = %d, want 0 (output: %s)", exitCode2, output2)
	}

	decoded := decodeContextOutput(t, output2)
	found := false
	for _, item := range decoded.Context.Items {
		if item.Path == "ai/knowledge/KNOW-009-new.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("context.items = %+v, want the newly-added KNOW-009 reflected after Sync", decoded.Context.Items)
	}
}

func TestContextCmd_IncompatibleSchemaIsTransparentlyRebuilt(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	// Simulate a stale, incompatible schema left by an older version of
	// the index package (matching internal/context/index/schema_test.go's
	// own TestEnsureSchema_RecreatesOnVersionMismatch approach, but
	// against a real file so the CLI's own index.Open sees it).
	cachePath := filepath.Join(root, ".misterspec", "cache", "context.db")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("creating cache dir: %v", err)
	}
	staleDB, err := sql.Open("sqlite", cachePath)
	if err != nil {
		t.Fatalf("sql.Open() unexpected error: %v", err)
	}
	if _, err := staleDB.Exec("CREATE TABLE documents (id INTEGER)"); err != nil {
		t.Fatalf("seeding stale schema: %v", err)
	}
	if _, err := staleDB.Exec("PRAGMA user_version = 999"); err != nil {
		t.Fatalf("seeding stale user_version: %v", err)
	}
	if err := staleDB.Close(); err != nil {
		t.Fatalf("closing stale db: %v", err)
	}

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 even with a corrupt/incompatible cache file (output: %s)", exitCode, output)
	}
	decoded := decodeContextOutput(t, output)
	if len(decoded.Context.Items) == 0 {
		t.Errorf("context.items is empty, want a correct result despite the corrupt starting cache")
	}
}

func TestContextCmd_NeverModifiesAuthoritativeArtifacts(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	specPath := root + "/ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/spec.md"
	before, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("reading fixture spec: %v", err)
	}

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	if _, exitCode := runCmd(cmd); exitCode != 0 {
		t.Fatalf("exitCode != 0")
	}

	after, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("re-reading fixture spec: %v", err)
	}
	if string(before) != string(after) {
		t.Errorf("SPEC-014/spec.md content changed after a context request:\nbefore: %s\nafter: %s", before, after)
	}
}

// --- User Story 4: rendered Markdown pack ---

func TestContextCmd_RenderPopulatesRenderedField(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--intent", "implementation", "--render", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	decoded := decodeContextOutput(t, output)
	if decoded.Context.Rendered == nil || *decoded.Context.Rendered == "" {
		t.Fatalf("context.rendered = %v, want a non-empty Markdown string", decoded.Context.Rendered)
	}
	for _, item := range decoded.Context.Items {
		if !strings.Contains(*decoded.Context.Rendered, item.Path) {
			t.Errorf("context.rendered does not mention item path %q", item.Path)
		}
	}
}

func TestContextCmd_NoRenderLeavesRenderedNull(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	decoded := decodeContextOutput(t, output)
	if decoded.Context.Rendered != nil {
		t.Errorf("context.rendered = %v, want nil", decoded.Context.Rendered)
	}
}

func TestContextCmd_RenderStillErrorsCleanlyOnUnknownTarget(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-999", "--render", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 3 {
		t.Fatalf("exitCode = %d, want 3 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "entity_not_found")
}
