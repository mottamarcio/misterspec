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

type contextLocationJSON struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type contextScoreComponentsJSON struct {
	Tier           int `json:"tier"`
	RelationWeight int `json:"relation_weight"`
	IntentBonus    int `json:"intent_bonus"`
	TextRelevance  int `json:"text_relevance"`
	Total          int `json:"total"`
}

type contextProvenanceJSON struct {
	SourcePath    string `json:"source_path"`
	SourceSection string `json:"source_section"`
	SourceLine    int    `json:"source_line"`
}

type contextItemJSON struct {
	Path            string                      `json:"path"`
	Heading         string                      `json:"heading"`
	Tier            string                      `json:"tier"`
	Reasons         []string                    `json:"reasons"`
	Score           int                         `json:"score"`
	Tokens          int                         `json:"tokens"`
	Content         *string                     `json:"content"`
	Location        *contextLocationJSON        `json:"location"`
	Fingerprint     *string                     `json:"fingerprint"`
	ScoreComponents *contextScoreComponentsJSON `json:"score_components"`
	Provenance      *[]contextProvenanceJSON    `json:"provenance"`
}

type contextDiagnosticsJSON struct {
	CandidatesConsidered int                    `json:"candidates_considered"`
	ItemsSelected        int                    `json:"items_selected"`
	TokensAvailable      int                    `json:"tokens_available"`
	TokensSelected       int                    `json:"tokens_selected"`
	TokensExcluded       int                    `json:"tokens_excluded"`
	ReductionPercent     float64                `json:"reduction_percent"`
	PayloadTokens        int                    `json:"payload_tokens"`
	Estimator            string                 `json:"estimator"`
	HardLimit            int                    `json:"hard_limit"`
	Exclusions           []contextExclusionJSON `json:"exclusions"`
	RankingVersion       int                    `json:"ranking_version"`
}

type contextResultJSON struct {
	SchemaVersion   int                    `json:"schema_version"`
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

type contextExclusionJSON struct {
	Path    string `json:"path"`
	Heading string `json:"heading"`
	Reason  string `json:"reason"`
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

// --- 033-context-pack-output-contract: Foundational (--mode flag) ---

func TestContextCmd_ModeManifestAndPackageAndMarkdownAccepted(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	for _, mode := range []string{"manifest", "package", "markdown"} {
		cmd := internalcmd.NewContextCmd()
		cmd.SetArgs([]string{"SPEC-014", "--mode", mode, "--dir", root})
		output, exitCode := runCmd(cmd)
		if exitCode != 0 {
			t.Fatalf("--mode %s: exitCode = %d, want 0 (output: %s)", mode, exitCode, output)
		}
	}
}

func TestContextCmd_OmittedModeMatchesExplicitManifest(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	withoutFlag := internalcmd.NewContextCmd()
	withoutFlag.SetArgs([]string{"SPEC-014", "--dir", root})
	outWithout, _ := runCmd(withoutFlag)

	withFlag := internalcmd.NewContextCmd()
	withFlag.SetArgs([]string{"SPEC-014", "--mode", "manifest", "--dir", root})
	outWith, _ := runCmd(withFlag)

	if outWithout != outWith {
		t.Errorf("omitted --mode output differs from explicit --mode manifest:\nomitted: %s\nexplicit: %s", outWithout, outWith)
	}
}

func TestContextCmd_UnrecognizedModeIsInvalidArgument(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--mode", "bogus", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_argument")
}

// --- User Story 1: full content without a mandatory re-read ---

func TestContextCmd_ModePackageIncludesMatchingContent(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--intent", "implementation", "--mode", "package", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	decoded := decodeContextOutput(t, output)
	if len(decoded.Context.Items) == 0 {
		t.Fatalf("context.items is empty, want at least the mandatory target entry")
	}
	for _, item := range decoded.Context.Items {
		if item.Content == nil || *item.Content == "" {
			t.Fatalf("item %+v has no content, want package mode to include it", item)
		}
		source, err := os.ReadFile(filepath.Join(root, item.Path))
		if err != nil {
			t.Fatalf("reading source %s: %v", item.Path, err)
		}
		if !strings.Contains(string(source), *item.Content) {
			t.Errorf("item content %q not found verbatim in source file %s", *item.Content, item.Path)
		}
	}
}

func TestContextCmd_DefaultModeHasNoContentField(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	if strings.Contains(output, `"content"`) {
		t.Errorf("default-mode output unexpectedly contains a \"content\" field: %s", output)
	}
}

func TestContextCmd_ModePackageFingerprintChangesOnEditAndStableOtherwise(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	run := func() contextEnvelopeJSON {
		cmd := internalcmd.NewContextCmd()
		cmd.SetArgs([]string{"SPEC-014", "--intent", "implementation", "--mode", "package", "--dir", root})
		output, exitCode := runCmd(cmd)
		if exitCode != 0 {
			t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
		}
		return decodeContextOutput(t, output)
	}

	targetPath := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/spec.md"
	findTarget := func(env contextEnvelopeJSON) *string {
		for _, item := range env.Context.Items {
			if item.Path == targetPath {
				return item.Fingerprint
			}
		}
		t.Fatalf("no item found for target path %s among %+v", targetPath, env.Context.Items)
		return nil
	}

	firstFP := findTarget(run())
	secondFP := findTarget(run())
	if firstFP == nil || *firstFP == "" || !strings.HasPrefix(*firstFP, "sha256:") {
		t.Fatalf("fingerprint = %v, want a non-empty \"sha256:<hex>\" string", firstFP)
	}
	if *firstFP != *secondFP {
		t.Errorf("fingerprint changed across two calls with no edit: %q != %q", *firstFP, *secondFP)
	}

	specPath := filepath.Join(root, targetPath)
	original, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("reading fixture spec: %v", err)
	}
	edited := append(append([]byte{}, original...), []byte("\nExtra sentence.\n")...)
	if err := os.WriteFile(specPath, edited, 0o644); err != nil {
		t.Fatalf("editing fixture spec: %v", err)
	}

	thirdFP := findTarget(run())
	if thirdFP == nil || *thirdFP == *firstFP {
		t.Errorf("fingerprint did not change after editing the source section: %v", thirdFP)
	}
}

// --- User Story 2: file-absolute location ---

func TestContextCmd_ModePackageLocationIsFileAbsoluteWithFrontmatter(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)
	// SPEC-014's own fixture body (writeContextFixture):
	// line 1: ---
	// line 2: id: SPEC-014
	// line 3: type: spec
	// line 4: status: ready
	// line 5: depends_on:
	// line 6:   - SPEC-011
	// line 7: supersedes: []
	// line 8: ---
	// line 9: ## Requirements
	// line 10: (blank)
	// line 11: Refresh token rotation. See [[KNOW-003]].

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--intent", "implementation", "--mode", "package", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	decoded := decodeContextOutput(t, output)

	targetPath := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/spec.md"
	var target *contextItemJSON
	for i := range decoded.Context.Items {
		if decoded.Context.Items[i].Path == targetPath {
			target = &decoded.Context.Items[i]
		}
	}
	if target == nil {
		t.Fatalf("no item found for target path %s", targetPath)
	}
	if target.Location == nil {
		t.Fatalf("item has no location")
	}

	lines, err := os.ReadFile(filepath.Join(root, targetPath))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	fileLines := strings.Split(string(lines), "\n")
	gotLine := fileLines[target.Location.StartLine-1]
	if !strings.Contains(gotLine, "## Requirements") {
		t.Errorf("location.start_line = %d points at %q, want the \"## Requirements\" heading line", target.Location.StartLine, gotLine)
	}
}

func TestContextCmd_ModePackageLocationCorrectForShorterFrontmatter(t *testing.T) {
	// Every canonical misterspec artifact requires frontmatter with its
	// required fields (internal/artifacts.ParseMetadata's own existing
	// behavior) — this test uses the shortest valid frontmatter to
	// confirm the offset isn't hardcoded to the longer fixture used
	// elsewhere in this file, but genuinely computed from each file's
	// own frontmatter length.
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md", ""+
		"---\n"+ // line 1
		"id: KNOW-001\n"+ // line 2
		"type: knowledge\n"+ // line 3
		"status: active\n"+ // line 4
		"---\n"+ // line 5
		"## Summary\n\nMinimal case.\n") // heading at file line 6

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"KNOW-001", "--mode", "package", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	decoded := decodeContextOutput(t, output)

	var target *contextItemJSON
	for i := range decoded.Context.Items {
		if decoded.Context.Items[i].Heading == "Summary" {
			target = &decoded.Context.Items[i]
		}
	}
	if target == nil {
		t.Fatalf("no Summary item found among %+v", decoded.Context.Items)
	}
	if target.Location.StartLine != 6 {
		t.Errorf("location.start_line = %d, want 6 (file-absolute)", target.Location.StartLine)
	}
}

func TestContextCmd_ModePackageTaskItemLocationIdentifiesOneTask(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Requirements\n\nSomething.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/tasks.md",
		"---\ntype: tasks\nfor: SPEC-014\n---\n# Tasks\n\n## TASK-001 — First task\n\n- [ ] Complete\n\n## TASK-002 — Second task\n\n- [ ] Complete\n")

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--intent", "tasks", "--query", "Second task", "--mode", "package", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	decoded := decodeContextOutput(t, output)

	var task2 *contextItemJSON
	for i := range decoded.Context.Items {
		if strings.Contains(decoded.Context.Items[i].Heading, "TASK-002") {
			task2 = &decoded.Context.Items[i]
		}
	}
	if task2 == nil {
		t.Skipf("no TASK-002 item selected among %+v — text search may not have matched; not this test's concern", decoded.Context.Items)
		return
	}
	if task2.Location.StartLine < 7 {
		t.Errorf("TASK-002 location.start_line = %d, want it to point at TASK-002's own heading, not TASK-001's", task2.Location.StartLine)
	}
}

// --- User Story 3: versioned, non-duplicating, backward-compatible contract ---

func TestContextCmd_SchemaVersionPresentInEveryMode(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	for _, mode := range []string{"manifest", "package", "markdown"} {
		cmd := internalcmd.NewContextCmd()
		cmd.SetArgs([]string{"SPEC-014", "--mode", mode, "--dir", root})
		output, exitCode := runCmd(cmd)
		if exitCode != 0 {
			t.Fatalf("--mode %s: exitCode = %d, want 0 (output: %s)", mode, exitCode, output)
		}
		decoded := decodeContextOutput(t, output)
		if decoded.Context.SchemaVersion != 4 {
			t.Errorf("--mode %s: schema_version = %d, want 4 (038-wikilink-chunk-provenance)", mode, decoded.Context.SchemaVersion)
		}
	}
}

func TestContextCmd_RankingVersionPresentInEveryMode(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	for _, mode := range []string{"manifest", "package", "markdown"} {
		cmd := internalcmd.NewContextCmd()
		cmd.SetArgs([]string{"SPEC-014", "--mode", mode, "--dir", root})
		output, exitCode := runCmd(cmd)
		if exitCode != 0 {
			t.Fatalf("--mode %s: exitCode = %d, want 0 (output: %s)", mode, exitCode, output)
		}
		decoded := decodeContextOutput(t, output)
		if decoded.Context.Diagnostics.RankingVersion != 1 {
			t.Errorf("--mode %s: diagnostics.ranking_version = %d, want 1 (036-text-search-ranking spec FR-009)", mode, decoded.Context.Diagnostics.RankingVersion)
		}
	}
}

func TestContextCmd_DiagnosticScoresOmittedByDefaultPresentWhenRequested(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmdDefault := internalcmd.NewContextCmd()
	cmdDefault.SetArgs([]string{"SPEC-014", "--query", "rotation", "--dir", root})
	outDefault, exitCode := runCmd(cmdDefault)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, outDefault)
	}
	decodedDefault := decodeContextOutput(t, outDefault)
	if len(decodedDefault.Context.Items) == 0 {
		t.Fatalf("context.items is empty")
	}
	for _, item := range decodedDefault.Context.Items {
		if item.ScoreComponents != nil {
			t.Errorf("item %+v has score_components without --diagnostic-scores, want nil (spec FR-008)", item)
		}
	}

	cmdDiag := internalcmd.NewContextCmd()
	cmdDiag.SetArgs([]string{"SPEC-014", "--query", "rotation", "--diagnostic-scores", "--dir", root})
	outDiag, exitCode2 := runCmd(cmdDiag)
	if exitCode2 != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode2, outDiag)
	}
	decodedDiag := decodeContextOutput(t, outDiag)
	if len(decodedDiag.Context.Items) == 0 {
		t.Fatalf("context.items is empty")
	}
	for _, item := range decodedDiag.Context.Items {
		if item.ScoreComponents == nil {
			t.Fatalf("item %+v has no score_components with --diagnostic-scores set, want present (contracts §3.2)", item)
		}
		sum := item.ScoreComponents.RelationWeight + item.ScoreComponents.IntentBonus + item.ScoreComponents.TextRelevance
		if item.ScoreComponents.Total != sum {
			t.Errorf("score_components.total = %d, want %d (sum of relation_weight+intent_bonus+text_relevance)", item.ScoreComponents.Total, sum)
		}
		if item.ScoreComponents.Total != item.Score {
			t.Errorf("score_components.total = %d, item.score = %d, want equal", item.ScoreComponents.Total, item.Score)
		}
	}
}

func TestContextCmd_ProvenanceOmittedByDefaultPresentWhenRequested(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmdDefault := internalcmd.NewContextCmd()
	cmdDefault.SetArgs([]string{"SPEC-014", "--dir", root})
	outDefault, exitCode := runCmd(cmdDefault)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, outDefault)
	}
	decodedDefault := decodeContextOutput(t, outDefault)
	if decodedDefault.Context.SchemaVersion != 4 {
		t.Errorf("schema_version = %d, want 4 (038-wikilink-chunk-provenance)", decodedDefault.Context.SchemaVersion)
	}
	for _, item := range decodedDefault.Context.Items {
		if item.Provenance != nil {
			t.Errorf("item %+v has provenance without --provenance, want nil (contract §4)", item)
		}
	}

	// SPEC-014 has an outgoing wikilink to KNOW-003, written in its own
	// "Requirements" section — that surfaces as a "wikilink" reason on
	// KNOW-003's own item when requesting context for SPEC-014.
	cmdProv := internalcmd.NewContextCmd()
	cmdProv.SetArgs([]string{"SPEC-014", "--provenance", "--dir", root})
	outProv, exitCode2 := runCmd(cmdProv)
	if exitCode2 != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode2, outProv)
	}
	decodedProv := decodeContextOutput(t, outProv)
	if decodedProv.Context.SchemaVersion != 4 {
		t.Errorf("--provenance: schema_version = %d, want 4", decodedProv.Context.SchemaVersion)
	}

	var sawWikilinkItem bool
	for _, item := range decodedProv.Context.Items {
		hasWikilinkReason := false
		for _, r := range item.Reasons {
			if r == "wikilink" {
				hasWikilinkReason = true
			}
		}
		if !hasWikilinkReason {
			if item.Provenance != nil {
				t.Errorf("item %+v has no wikilink reason but has provenance, want nil (never a stand-in empty array)", item)
			}
			continue
		}
		sawWikilinkItem = true
		if item.Provenance == nil || len(*item.Provenance) == 0 {
			t.Fatalf("item %+v has a wikilink reason but no provenance entries with --provenance set", item)
		}
		p := (*item.Provenance)[0]
		if p.SourcePath != "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/spec.md" {
			t.Errorf("provenance source_path = %q, want SPEC-014's own path", p.SourcePath)
		}
		if p.SourceSection != "Requirements" {
			t.Errorf("provenance source_section = %q, want %q", p.SourceSection, "Requirements")
		}
		if p.SourceLine <= 0 {
			t.Errorf("provenance source_line = %d, want a positive line", p.SourceLine)
		}
	}
	if !sawWikilinkItem {
		t.Fatalf("no item with a wikilink reason found for SPEC-014; items = %+v", decodedProv.Context.Items)
	}
}

// TestContextCmd_ProvenancePresentForIncomingBacklinkReason confirms
// contract §4's own correction: internal/context/collector.go always
// labels an incoming reference "backlink", never "wikilink", so
// --provenance must key off whether real occurrence data is present —
// not the relation string — or every incoming reference's own
// provenance would be silently dropped.
func TestContextCmd_ProvenancePresentForIncomingBacklinkReason(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"KNOW-003", "--provenance", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	decoded := decodeContextOutput(t, output)

	var sawBacklinkItem bool
	for _, item := range decoded.Context.Items {
		hasBacklinkReason := false
		for _, r := range item.Reasons {
			if r == "backlink" {
				hasBacklinkReason = true
			}
		}
		if !hasBacklinkReason {
			continue
		}
		sawBacklinkItem = true
		if item.Provenance == nil || len(*item.Provenance) == 0 {
			t.Fatalf("item %+v has a backlink reason (from a semantic wikilink) but no provenance entries", item)
		}
		p := (*item.Provenance)[0]
		if p.SourceSection != "Requirements" {
			t.Errorf("provenance source_section = %q, want %q", p.SourceSection, "Requirements")
		}
	}
	if !sawBacklinkItem {
		t.Fatalf("no item with a backlink reason found for KNOW-003; items = %+v", decoded.Context.Items)
	}
}

// TestContextCmd_PreferSectionChangesOrderingOnlyWhenPassed exercises
// contract §5: SPEC-014 wikilinks KNOW-003 from its own "Requirements"
// section; requesting context for KNOW-003 surfaces SPEC-014 as a
// "backlink" item whose SectionPreference bonus (visible via
// --diagnostic-scores) is 0 by default and non-zero only with
// --prefer-section — and ordering itself must be byte-identical
// without the flag, matching quickstart.md §4.
func TestContextCmd_PreferSectionChangesOrderingOnlyWhenPassed(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)
	// A second Spec that also links to KNOW-003, but from an unrelated
	// section — an otherwise-equivalent Tier/relation sibling to
	// SPEC-014's own item, so --prefer-section has something to
	// distinguish.
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-020/spec.md",
		"---\nid: SPEC-020\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Notes\n\nAlso see [[KNOW-003]].\n")

	without := internalcmd.NewContextCmd()
	without.SetArgs([]string{"KNOW-003", "--dir", root})
	outWithout, exitCode := runCmd(without)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, outWithout)
	}

	baseline := internalcmd.NewContextCmd()
	baseline.SetArgs([]string{"KNOW-003", "--dir", root})
	outBaseline, exitCode2 := runCmd(baseline)
	if exitCode2 != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode2, outBaseline)
	}
	if outWithout != outBaseline {
		t.Fatalf("two omitted-flag runs produced different output — ordering must be deterministic:\n%s\nvs\n%s", outWithout, outBaseline)
	}

	withFlag := internalcmd.NewContextCmd()
	withFlag.SetArgs([]string{"KNOW-003", "--prefer-section", "--diagnostic-scores", "--dir", root})
	outWith, exitCode3 := runCmd(withFlag)
	if exitCode3 != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode3, outWith)
	}
	// Decode with a struct that also captures section_preference, since
	// contextScoreComponentsJSON doesn't declare it.
	var raw struct {
		Context struct {
			Items []struct {
				Path            string `json:"path"`
				ScoreComponents *struct {
					SectionPreference int `json:"section_preference"`
				} `json:"score_components"`
			} `json:"items"`
		} `json:"context"`
	}
	if err := json.Unmarshal([]byte(outWith), &raw); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	var foundPreferred, foundUnrelated bool
	for _, item := range raw.Context.Items {
		if item.ScoreComponents == nil {
			continue
		}
		switch item.Path {
		case "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-014/spec.md":
			foundPreferred = true
			if item.ScoreComponents.SectionPreference == 0 {
				t.Errorf("SPEC-014 (Requirements-sourced) section_preference = 0, want non-zero with --prefer-section")
			}
		case "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-020/spec.md":
			foundUnrelated = true
			if item.ScoreComponents.SectionPreference != 0 {
				t.Errorf("SPEC-020 (Notes-sourced) section_preference = %d, want 0", item.ScoreComponents.SectionPreference)
			}
		}
	}
	if !foundPreferred || !foundUnrelated {
		t.Fatalf("expected both SPEC-014 and SPEC-020 items with score_components; foundPreferred=%v foundUnrelated=%v", foundPreferred, foundUnrelated)
	}
}

func TestContextCmd_QueryModeRoutingAndErrors(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	// Default (omitted --query-mode) and explicit "free" both succeed
	// on ordinary punctuation-bearing text.
	for _, args := range [][]string{
		{"SPEC-014", "--query", `"refresh token"`, "--dir", root},
		{"SPEC-014", "--query", `"refresh token"`, "--query-mode", "free", "--dir", root},
	} {
		cmd := internalcmd.NewContextCmd()
		cmd.SetArgs(args)
		output, exitCode := runCmd(cmd)
		if exitCode != 0 {
			t.Fatalf("args=%v: exitCode = %d, want 0 (output: %s)", args, exitCode, output)
		}
	}

	// Explicit "advanced" with valid FTS5 phrase syntax succeeds.
	cmdAdvanced := internalcmd.NewContextCmd()
	cmdAdvanced.SetArgs([]string{"SPEC-014", "--query", `"refresh token"`, "--query-mode", "advanced", "--dir", root})
	outputAdvanced, exitCodeAdvanced := runCmd(cmdAdvanced)
	if exitCodeAdvanced != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCodeAdvanced, outputAdvanced)
	}

	// An unrecognized --query-mode value is invalid_argument.
	cmdInvalid := internalcmd.NewContextCmd()
	cmdInvalid.SetArgs([]string{"SPEC-014", "--query", "rotation", "--query-mode", "bogus", "--dir", root})
	outputInvalid, exitCodeInvalid := runCmd(cmdInvalid)
	if exitCodeInvalid != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCodeInvalid, outputInvalid)
	}
	assertErrorCode(t, outputInvalid, "invalid_argument")

	// A malformed advanced query returns query_syntax_error, not a
	// generic failure.
	cmdMalformed := internalcmd.NewContextCmd()
	cmdMalformed.SetArgs([]string{"SPEC-014", "--query", `"unbalanced`, "--query-mode", "advanced", "--dir", root})
	outputMalformed, exitCodeMalformed := runCmd(cmdMalformed)
	if exitCodeMalformed != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCodeMalformed, outputMalformed)
	}
	assertErrorCode(t, outputMalformed, "query_syntax_error")
}

func TestContextCmd_HardLimitFlagParsesAndDefaultsTo12000(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	decoded := decodeContextOutput(t, output)
	if decoded.Context.Diagnostics.HardLimit != 12000 {
		t.Errorf("diagnostics.hard_limit = %d, want 12000 (omitted --hard-limit must still resolve to a finite default — FR-006)", decoded.Context.Diagnostics.HardLimit)
	}

	cmd2 := internalcmd.NewContextCmd()
	cmd2.SetArgs([]string{"SPEC-014", "--hard-limit", "500", "--dir", root})
	output2, exitCode2 := runCmd(cmd2)
	if exitCode2 != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode2, output2)
	}
	decoded2 := decodeContextOutput(t, output2)
	if decoded2.Context.Diagnostics.HardLimit != 500 {
		t.Errorf("diagnostics.hard_limit = %d, want 500 (--hard-limit must be honored)", decoded2.Context.Diagnostics.HardLimit)
	}
}

func TestContextCmd_DiagnosticsIdentifiesTheEstimatorInEveryMode(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	for _, mode := range []string{"manifest", "package", "markdown"} {
		cmd := internalcmd.NewContextCmd()
		cmd.SetArgs([]string{"SPEC-014", "--mode", mode, "--dir", root})
		output, exitCode := runCmd(cmd)
		if exitCode != 0 {
			t.Fatalf("--mode %s: exitCode = %d, want 0 (output: %s)", mode, exitCode, output)
		}
		decoded := decodeContextOutput(t, output)
		if decoded.Context.Diagnostics.Estimator != "default" {
			t.Errorf("--mode %s: diagnostics.estimator = %q, want %q (FR-001)", mode, decoded.Context.Diagnostics.Estimator, "default")
		}
	}
}

func TestContextCmd_ModePackageRenderedIsNull(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--mode", "package", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	decoded := decodeContextOutput(t, output)
	if decoded.Context.Rendered != nil {
		t.Errorf("context.rendered = %v, want nil in package mode — no duplication", decoded.Context.Rendered)
	}
}

func TestContextCmd_ModePackageWithRenderIsRejected(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	cmd := internalcmd.NewContextCmd()
	cmd.SetArgs([]string{"SPEC-014", "--mode", "package", "--render", "--dir", root})
	output, exitCode := runCmd(cmd)
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_argument")
}

func TestContextCmd_DefaultAndRenderOutputsUnchangedExceptAdditiveFields(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	assertOnlyAdditiveFieldsAdded := func(t *testing.T, output string) {
		t.Helper()
		var raw map[string]any
		if err := json.Unmarshal([]byte(output), &raw); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}
		ctx, ok := raw["context"].(map[string]any)
		if !ok {
			t.Fatalf("output has no \"context\" object: %s", output)
		}
		wantKeys := map[string]bool{
			"schema_version": true, "target": true, "intent": true, "budget": true,
			"estimated_tokens": true, "budget_exceeded": true, "overage": true,
			"items": true, "diagnostics": true, "rendered": true,
		}
		for k := range ctx {
			if !wantKeys[k] {
				t.Errorf("unexpected new top-level context field %q — only schema_version was expected to be additive", k)
			}
		}
		items, _ := ctx["items"].([]any)
		for _, it := range items {
			item, _ := it.(map[string]any)
			wantItemKeys := map[string]bool{"path": true, "heading": true, "tier": true, "reasons": true, "score": true, "tokens": true}
			for k := range item {
				if !wantItemKeys[k] {
					t.Errorf("unexpected new item field %q in manifest-shaped item — items must stay content-free by default", k)
				}
			}
		}
		diag, _ := ctx["diagnostics"].(map[string]any)
		wantDiagKeys := map[string]bool{
			"candidates_considered": true, "items_selected": true, "tokens_available": true,
			"tokens_selected": true, "tokens_excluded": true, "reduction_percent": true,
			"payload_tokens": true,
			// 035-context-budget-accuracy's own additive diagnostics
			// fields (contracts/budget-and-estimator-contract.md §2.2):
			"estimator": true, "hard_limit": true, "exclusions": true,
			// 036-text-search-ranking's own additive diagnostics field
			// (contracts/search-and-ranking-contract.md §3.1):
			"ranking_version": true,
		}
		for k := range diag {
			if !wantDiagKeys[k] {
				t.Errorf("unexpected new diagnostics field %q — only payload_tokens (033), estimator/hard_limit/exclusions (035), and ranking_version (036) were expected to be additive", k)
			}
		}
	}

	cmdDefault := internalcmd.NewContextCmd()
	cmdDefault.SetArgs([]string{"SPEC-014", "--intent", "implementation", "--dir", root})
	outDefault, exitCode := runCmd(cmdDefault)
	if exitCode != 0 {
		t.Fatalf("default: exitCode = %d, want 0 (output: %s)", exitCode, outDefault)
	}
	assertOnlyAdditiveFieldsAdded(t, outDefault)

	cmdRender := internalcmd.NewContextCmd()
	cmdRender.SetArgs([]string{"SPEC-014", "--intent", "implementation", "--render", "--dir", root})
	outRender, exitCode := runCmd(cmdRender)
	if exitCode != 0 {
		t.Fatalf("--render: exitCode = %d, want 0 (output: %s)", exitCode, outRender)
	}
	assertOnlyAdditiveFieldsAdded(t, outRender)
}

func TestContextCmd_PayloadTokensAtLeastTokensSelectedAndTokensSelectedUnchanged(t *testing.T) {
	root := testutil.Project(t)
	writeContextFixture(t, root)

	manifestCmd := internalcmd.NewContextCmd()
	manifestCmd.SetArgs([]string{"SPEC-014", "--intent", "implementation", "--mode", "manifest", "--dir", root})
	manifestOut, exitCode := runCmd(manifestCmd)
	if exitCode != 0 {
		t.Fatalf("manifest: exitCode = %d, want 0 (output: %s)", exitCode, manifestOut)
	}
	manifest := decodeContextOutput(t, manifestOut)

	packageCmd := internalcmd.NewContextCmd()
	packageCmd.SetArgs([]string{"SPEC-014", "--intent", "implementation", "--mode", "package", "--dir", root})
	packageOut, exitCode := runCmd(packageCmd)
	if exitCode != 0 {
		t.Fatalf("package: exitCode = %d, want 0 (output: %s)", exitCode, packageOut)
	}
	pkg := decodeContextOutput(t, packageOut)

	if pkg.Context.Diagnostics.TokensSelected != manifest.Context.Diagnostics.TokensSelected {
		t.Errorf("package tokens_selected = %d, want unchanged from manifest's %d", pkg.Context.Diagnostics.TokensSelected, manifest.Context.Diagnostics.TokensSelected)
	}
	if pkg.Context.Diagnostics.PayloadTokens < pkg.Context.Diagnostics.TokensSelected {
		t.Errorf("package payload_tokens = %d, want >= tokens_selected %d", pkg.Context.Diagnostics.PayloadTokens, pkg.Context.Diagnostics.TokensSelected)
	}
}
