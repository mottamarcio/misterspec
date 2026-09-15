package operations_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestCreate_Program(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	result, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if result.ID.String() != "PRG-001" {
		t.Errorf("ID = %v, want PRG-001", result.ID)
	}
	if result.Path != "ai/programs/PRG-001/program.md" {
		t.Errorf("Path = %q, unexpected", result.Path)
	}

	inspected, err := operations.Inspect(root, cfg, "PRG-001")
	if err != nil {
		t.Fatalf("newly created Program failed Inspect(): %v", err)
	}
	if inspected.Metadata.Status != "draft" {
		t.Errorf("Status = %q, want %q", inspected.Metadata.Status, "draft")
	}
}

func TestCreate_FeatureUnderValidProgram(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	prg, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("Create(Program) unexpected error: %v", err)
	}

	feat, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Feature, Parent: prg.ID.String()})
	if err != nil {
		t.Fatalf("Create(Feature) unexpected error: %v", err)
	}
	if feat.ID.String() != "FEAT-001" {
		t.Errorf("ID = %v, want FEAT-001", feat.ID)
	}

	if _, err := operations.Inspect(root, cfg, "FEAT-001"); err != nil {
		t.Fatalf("newly created Feature failed Inspect(): %v", err)
	}
}

func TestCreate_SequentialSpecsUnderSameFeature(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	prg, _ := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	feat, _ := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Feature, Parent: prg.ID.String()})

	want := []string{"SPEC-001", "SPEC-002", "SPEC-003"}
	for _, w := range want {
		spec, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Spec, Parent: feat.ID.String()})
		if err != nil {
			t.Fatalf("Create(Spec) unexpected error: %v", err)
		}
		if spec.ID.String() != w {
			t.Fatalf("Create(Spec) = %v, want %v", spec.ID, w)
		}
	}
}

func TestCreate_InvalidParentRejectedNothingWritten(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	_, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Feature, Parent: "PRG-999"})
	if !errors.Is(err, operations.ErrInvalidParent) {
		t.Fatalf("Create() error = %v, want errors.Is(err, ErrInvalidParent)", err)
	}

	if _, statErr := os.Stat(filepath.Join(root, "ai/programs")); !os.IsNotExist(statErr) {
		t.Errorf("ai/programs unexpectedly exists after a rejected Create(): %v", statErr)
	}
}

func TestCreate_KnowledgeWithSlug(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	result, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Knowledge, Slug: "authentication"})
	if err != nil {
		t.Fatalf("Create(Knowledge) unexpected error: %v", err)
	}
	if result.Path != "ai/knowledge/KNOW-001-authentication.md" {
		t.Errorf("Path = %q, unexpected", result.Path)
	}

	learning, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Learning, Slug: "retry-strategy"})
	if err != nil {
		t.Fatalf("Create(Learning) unexpected error: %v", err)
	}
	if learning.Path != "ai/memory/learnings/LRN-001-retry-strategy.md" {
		t.Errorf("Path = %q, unexpected", learning.Path)
	}
}

// Note: Create's "target already exists" defensive check (FR-008) is not
// naturally reachable through black-box testing for auto-allocated types
// (Program/Feature/Spec/Knowledge/Learning) — by construction, NextID is
// always computed from what Scan already found, so a freshly allocated ID
// can never collide with an existing, Scan-visible artifact in normal
// operation. The check exists as a safety net (and its real, naturally
// reachable test coverage lives in create_artifact_test.go, where a
// second Plan/Tasks/Validation for the same Spec — a fixed, non-allocated
// path — genuinely can and does collide).

func TestCreate_UnsupportedType(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	_, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Task})
	if !errors.Is(err, operations.ErrUnsupportedType) {
		t.Fatalf("Create() error = %v, want errors.Is(err, ErrUnsupportedType)", err)
	}
}

// --- 022-feature-branch-automation: Feature creation and Git branches ---

func TestCreate_Feature_CreatesAndChecksOutNewBranch(t *testing.T) {
	root := testutil.Project(t)
	testutil.InitGitRepo(t, root)
	cfg := testConfig()

	prg, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("Create(Program) unexpected error: %v", err)
	}

	feat, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Feature, Parent: prg.ID.String()})
	if err != nil {
		t.Fatalf("Create(Feature) unexpected error: %v", err)
	}

	wantBranch := "feat/" + feat.ID.String()
	if feat.GitBranch != wantBranch {
		t.Errorf("GitBranch = %q, want %q", feat.GitBranch, wantBranch)
	}
	if !feat.GitBranchCreated {
		t.Error("GitBranchCreated = false for a brand-new Feature, want true")
	}
	if feat.GitSkippedReason != "" {
		t.Errorf("GitSkippedReason = %q, want empty", feat.GitSkippedReason)
	}

	current := gitCurrentBranch(t, root)
	if current != wantBranch {
		t.Errorf("current branch after Create(Feature) = %q, want %q", current, wantBranch)
	}
}

func TestCreate_Feature_SlugMakesBranchNameReadable(t *testing.T) {
	root := testutil.Project(t)
	testutil.InitGitRepo(t, root)
	cfg := testConfig()

	prg, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("Create(Program) unexpected error: %v", err)
	}

	feat, err := operations.Create(root, cfg, operations.CreateRequest{
		Type:   ids.Feature,
		Parent: prg.ID.String(),
		Slug:   "User Auth",
	})
	if err != nil {
		t.Fatalf("Create(Feature) unexpected error: %v", err)
	}

	wantBranch := "feat/" + feat.ID.String() + "-user-auth"
	if feat.GitBranch != wantBranch {
		t.Errorf("GitBranch = %q, want %q", feat.GitBranch, wantBranch)
	}
	if current := gitCurrentBranch(t, root); current != wantBranch {
		t.Errorf("current branch = %q, want %q", current, wantBranch)
	}
}

func TestCreate_Feature_ResumesExistingBranch(t *testing.T) {
	root := testutil.Project(t)
	testutil.InitGitRepo(t, root)
	cfg := testConfig()

	prg, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("Create(Program) unexpected error: %v", err)
	}

	// Simulate an interrupted prior attempt: the branch for the next
	// Feature ID (FEAT-001) already exists, but no artifact was written.
	runGitTest(t, root, "checkout", "-q", "-b", "feat/FEAT-001")
	initialBranch := gitInitialBranch(t, root)
	runGitTest(t, root, "checkout", "-q", initialBranch)

	feat, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Feature, Parent: prg.ID.String()})
	if err != nil {
		t.Fatalf("Create(Feature) unexpected error: %v", err)
	}

	if feat.GitBranchCreated {
		t.Error("GitBranchCreated = true for an already-existing branch, want false (resume, not recreate)")
	}
	if feat.GitBranch != "feat/FEAT-001" {
		t.Errorf("GitBranch = %q, want %q", feat.GitBranch, "feat/FEAT-001")
	}
}

func TestCreate_Feature_NonGitProject_SkipsWithoutFailing(t *testing.T) {
	root := testutil.Project(t) // no InitGitRepo — plain, non-Git directory
	cfg := testConfig()

	prg, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("Create(Program) unexpected error: %v", err)
	}

	feat, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Feature, Parent: prg.ID.String()})
	if err != nil {
		t.Fatalf("Create(Feature) unexpected error: %v", err)
	}
	if feat.GitSkippedReason != "not_a_git_repo" {
		t.Errorf("GitSkippedReason = %q, want %q", feat.GitSkippedReason, "not_a_git_repo")
	}
	if feat.GitBranch != "" || feat.GitBranchCreated {
		t.Errorf("GitBranch/GitBranchCreated = %q/%v, want empty/false when skipped", feat.GitBranch, feat.GitBranchCreated)
	}
}

func TestCreate_Feature_AutomationDisabled_SkipsWithoutFailing(t *testing.T) {
	root := testutil.Project(t)
	testutil.InitGitRepo(t, root)
	cfg := testConfig()
	cfg.GitBranchAutomation = false
	initialBranch := gitInitialBranch(t, root)

	prg, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("Create(Program) unexpected error: %v", err)
	}

	feat, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Feature, Parent: prg.ID.String()})
	if err != nil {
		t.Fatalf("Create(Feature) unexpected error: %v", err)
	}
	if feat.GitSkippedReason != "disabled" {
		t.Errorf("GitSkippedReason = %q, want %q", feat.GitSkippedReason, "disabled")
	}
	if current := gitCurrentBranch(t, root); current != initialBranch {
		t.Errorf("current branch = %q, want unchanged %q (automation disabled)", current, initialBranch)
	}
}

// --- 022-feature-branch-automation: Spec creation and branch mismatch ---

func TestCreate_Spec_OnFeatureBranch_NoNewBranchNoWarning(t *testing.T) {
	root := testutil.Project(t)
	testutil.InitGitRepo(t, root)
	cfg := testConfig()

	prg, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("Create(Program) unexpected error: %v", err)
	}
	feat, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Feature, Parent: prg.ID.String()})
	if err != nil {
		t.Fatalf("Create(Feature) unexpected error: %v", err)
	}
	// Create(Feature) already checked out feat.GitBranch.
	branchesBefore := gitLocalBranches(t, root)

	spec, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Spec, Parent: feat.ID.String()})
	if err != nil {
		t.Fatalf("Create(Spec) unexpected error: %v", err)
	}

	if spec.GitBranch != feat.GitBranch {
		t.Errorf("GitBranch = %q, want parent Feature's own branch %q", spec.GitBranch, feat.GitBranch)
	}
	if spec.GitWarning != "" {
		t.Errorf("GitWarning = %q, want empty when on the Feature's own branch", spec.GitWarning)
	}
	if current := gitCurrentBranch(t, root); current != feat.GitBranch {
		t.Errorf("current branch after Create(Spec) = %q, want unchanged %q", current, feat.GitBranch)
	}
	if branchesAfter := gitLocalBranches(t, root); branchesAfter != branchesBefore {
		t.Errorf("local branches changed after Create(Spec): before=%q after=%q", branchesBefore, branchesAfter)
	}
}

func TestCreate_Spec_OnDifferentBranch_WarnsButSucceeds(t *testing.T) {
	root := testutil.Project(t)
	testutil.InitGitRepo(t, root)
	cfg := testConfig()
	devBranch := gitCurrentBranch(t, root)

	prg, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("Create(Program) unexpected error: %v", err)
	}
	feat, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Feature, Parent: prg.ID.String()})
	if err != nil {
		t.Fatalf("Create(Feature) unexpected error: %v", err)
	}

	runGitTest(t, root, "checkout", "-q", devBranch)

	spec, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Spec, Parent: feat.ID.String()})
	if err != nil {
		t.Fatalf("Create(Spec) unexpected error: %v", err)
	}

	if spec.GitBranch != feat.GitBranch {
		t.Errorf("GitBranch = %q, want parent Feature's own branch %q", spec.GitBranch, feat.GitBranch)
	}
	if spec.GitWarning == "" {
		t.Error("GitWarning is empty, want a mismatch notice naming both branches")
	}
	if !strings.Contains(spec.GitWarning, devBranch) || !strings.Contains(spec.GitWarning, feat.GitBranch) {
		t.Errorf("GitWarning = %q, want it to name both %q and %q", spec.GitWarning, devBranch, feat.GitBranch)
	}
	if current := gitCurrentBranch(t, root); current != devBranch {
		t.Errorf("current branch after Create(Spec) = %q, want unchanged %q (never auto-switched)", current, devBranch)
	}
}

func TestCreate_Spec_NonGitOrDisabled_SkipsLikeFeature(t *testing.T) {
	root := testutil.Project(t) // no InitGitRepo
	cfg := testConfig()

	prg, _ := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	feat, _ := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Feature, Parent: prg.ID.String()})

	spec, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Spec, Parent: feat.ID.String()})
	if err != nil {
		t.Fatalf("Create(Spec) unexpected error: %v", err)
	}
	if spec.GitSkippedReason != "not_a_git_repo" {
		t.Errorf("GitSkippedReason = %q, want %q", spec.GitSkippedReason, "not_a_git_repo")
	}
	if spec.GitWarning != "" || spec.GitBranch != "" {
		t.Errorf("GitWarning/GitBranch = %q/%q, want both empty when skipped", spec.GitWarning, spec.GitBranch)
	}
}

// --- 022-feature-branch-automation, User Story 3: full opt-out ---

// TestCreate_GitBranchAutomationDisabled_EndToEnd pins down the whole
// opt-out experience (not just an individual GitSkippedReason value):
// with automation disabled, creating a Feature and then a Spec under it
// must leave zero Git side effects of any kind — no new local branch,
// no branch switch — identical to how Create behaved before this
// feature existed.
func TestCreate_GitBranchAutomationDisabled_EndToEnd(t *testing.T) {
	root := testutil.Project(t)
	testutil.InitGitRepo(t, root)
	cfg := testConfig()
	cfg.GitBranchAutomation = false

	initialBranch := gitCurrentBranch(t, root)
	branchesBefore := gitLocalBranches(t, root)

	prg, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("Create(Program) unexpected error: %v", err)
	}
	feat, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Feature, Parent: prg.ID.String()})
	if err != nil {
		t.Fatalf("Create(Feature) unexpected error: %v", err)
	}
	if _, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Spec, Parent: feat.ID.String()}); err != nil {
		t.Fatalf("Create(Spec) unexpected error: %v", err)
	}

	if current := gitCurrentBranch(t, root); current != initialBranch {
		t.Errorf("current branch = %q, want unchanged %q across both calls", current, initialBranch)
	}
	if branchesAfter := gitLocalBranches(t, root); branchesAfter != branchesBefore {
		t.Errorf("local branches changed: before=%q after=%q, want zero new branches", branchesBefore, branchesAfter)
	}
}

func gitLocalBranches(t *testing.T, root string) string {
	t.Helper()
	return runGitTest(t, root, "branch", "--list")
}

func gitCurrentBranch(t *testing.T, root string) string {
	t.Helper()
	out := runGitTest(t, root, "branch", "--show-current")
	return strings.TrimSpace(out)
}

func gitInitialBranch(t *testing.T, root string) string {
	t.Helper()
	return gitCurrentBranch(t, root)
}

func runGitTest(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}
