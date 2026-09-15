package vcs

import (
	"os/exec"
	"strings"
	"testing"
)

// initFixtureRepo creates a real Git repository in a fresh temp
// directory, with one empty initial commit on its default branch, so
// CurrentBranch has something non-empty to report (022-feature-branch-
// automation's own research.md #2 — real `git`, not a hand-rolled
// fixture, since that's exactly what IsRepo/CurrentBranch/EnsureBranch
// themselves shell out to).
func initFixtureRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init", "-q")
	run("-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "--allow-empty", "-q", "-m", "init")
	return dir
}

func TestIsRepo_TrueInsideFixture(t *testing.T) {
	dir := initFixtureRepo(t)

	if !IsRepo(dir) {
		t.Error("IsRepo() = false inside a real Git repository, want true")
	}
}

func TestIsRepo_FalseOutsideAnyRepo(t *testing.T) {
	dir := t.TempDir()

	if IsRepo(dir) {
		t.Error("IsRepo() = true for a plain non-Git directory, want false")
	}
}

func TestBranchName_NoSlugIsBareID(t *testing.T) {
	got := BranchName("FEAT-007", "")
	want := "feat/FEAT-007"
	if got != want {
		t.Errorf("BranchName(%q, \"\") = %q, want %q", "FEAT-007", got, want)
	}

	// Called again, with no filesystem/Git access at all — still the
	// same result (pure function, research.md #3).
	if got2 := BranchName("FEAT-007", ""); got2 != want {
		t.Errorf("BranchName(%q, \"\") on second call = %q, want %q (must be pure)", "FEAT-007", got2, want)
	}
}

func TestBranchName_SlugIsSlugifiedAndAppended(t *testing.T) {
	got := BranchName("FEAT-007", "User Auth!!")
	want := "feat/FEAT-007-user-auth"
	if got != want {
		t.Errorf("BranchName(%q, %q) = %q, want %q", "FEAT-007", "User Auth!!", got, want)
	}
}

func TestBranchName_UnusableSlugFallsBackToBareID(t *testing.T) {
	got := BranchName("FEAT-007", "!!!")
	want := "feat/FEAT-007"
	if got != want {
		t.Errorf("BranchName(%q, %q) = %q, want %q (slug reduces to nothing usable)", "FEAT-007", "!!!", got, want)
	}
}

func TestCurrentBranch_ReturnsFixtureBranch(t *testing.T) {
	dir := initFixtureRepo(t)

	want := strings.TrimSpace(runGit(t, dir, "branch", "--show-current"))
	if want == "" {
		t.Fatal("fixture repo's own current branch is empty — fixture itself is broken")
	}

	got, err := CurrentBranch(dir)
	if err != nil {
		t.Fatalf("CurrentBranch() unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("CurrentBranch() = %q, want %q", got, want)
	}
}

func TestEnsureBranch_CreatesAndChecksOutNewBranch(t *testing.T) {
	dir := initFixtureRepo(t)

	created, err := EnsureBranch(dir, "feat/FEAT-999")
	if err != nil {
		t.Fatalf("EnsureBranch() unexpected error: %v", err)
	}
	if !created {
		t.Error("EnsureBranch() created = false for a brand-new branch name, want true")
	}

	current := strings.TrimSpace(runGit(t, dir, "branch", "--show-current"))
	if current != "feat/FEAT-999" {
		t.Errorf("current branch after EnsureBranch() = %q, want %q", current, "feat/FEAT-999")
	}
}

func TestEnsureBranch_ResumesExistingBranchWithoutError(t *testing.T) {
	dir := initFixtureRepo(t)
	initialBranch := strings.TrimSpace(runGit(t, dir, "branch", "--show-current"))

	if _, err := EnsureBranch(dir, "feat/FEAT-999"); err != nil {
		t.Fatalf("first EnsureBranch() unexpected error: %v", err)
	}

	// Switch away, then resume — the second EnsureBranch() call must
	// check the existing branch back out, not fail or recreate it.
	runGit(t, dir, "checkout", "-q", initialBranch)

	created, err := EnsureBranch(dir, "feat/FEAT-999")
	if err != nil {
		t.Fatalf("second EnsureBranch() unexpected error: %v", err)
	}
	if created {
		t.Error("EnsureBranch() created = true for an already-existing branch, want false (resume, not recreate)")
	}

	current := strings.TrimSpace(runGit(t, dir, "branch", "--show-current"))
	if current != "feat/FEAT-999" {
		t.Errorf("current branch after resuming EnsureBranch() = %q, want %q", current, "feat/FEAT-999")
	}
}

func TestBranchForID_FindsNoBranchWhenNoneExists(t *testing.T) {
	dir := initFixtureRepo(t)

	got, err := BranchForID(dir, "FEAT-007")
	if err != nil {
		t.Fatalf("BranchForID() unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("BranchForID() = %q, want \"\" when no branch exists for this ID", got)
	}
}

func TestBranchForID_FindsBranchRegardlessOfSlug(t *testing.T) {
	dir := initFixtureRepo(t)
	runGit(t, dir, "checkout", "-q", "-b", "feat/FEAT-007-user-auth")

	got, err := BranchForID(dir, "FEAT-007")
	if err != nil {
		t.Fatalf("BranchForID() unexpected error: %v", err)
	}
	if got != "feat/FEAT-007-user-auth" {
		t.Errorf("BranchForID() = %q, want %q", got, "feat/FEAT-007-user-auth")
	}
}

func TestBranchForID_DoesNotMatchDifferentID(t *testing.T) {
	dir := initFixtureRepo(t)
	runGit(t, dir, "checkout", "-q", "-b", "feat/FEAT-17")

	// "FEAT-1" must not prefix-match the unrelated branch "feat/FEAT-17".
	got, err := BranchForID(dir, "FEAT-1")
	if err != nil {
		t.Fatalf("BranchForID() unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("BranchForID(%q) = %q, want \"\" (must not match %q)", "FEAT-1", got, "feat/FEAT-17")
	}
}

func TestEnsureFeatureBranch_CreatesWithSlugWhenNoneExists(t *testing.T) {
	dir := initFixtureRepo(t)

	branch, created, err := EnsureFeatureBranch(dir, "FEAT-007", "User Auth")
	if err != nil {
		t.Fatalf("EnsureFeatureBranch() unexpected error: %v", err)
	}
	if !created {
		t.Error("created = false for a brand-new Feature, want true")
	}
	if branch != "feat/FEAT-007-user-auth" {
		t.Errorf("branch = %q, want %q", branch, "feat/FEAT-007-user-auth")
	}
	if current := strings.TrimSpace(runGit(t, dir, "branch", "--show-current")); current != branch {
		t.Errorf("current branch = %q, want %q", current, branch)
	}
}

func TestEnsureFeatureBranch_ResumesExistingBranchIgnoringNewSlug(t *testing.T) {
	dir := initFixtureRepo(t)
	initialBranch := strings.TrimSpace(runGit(t, dir, "branch", "--show-current"))

	if _, _, err := EnsureFeatureBranch(dir, "FEAT-007", "user auth"); err != nil {
		t.Fatalf("first EnsureFeatureBranch() unexpected error: %v", err)
	}
	runGit(t, dir, "checkout", "-q", initialBranch)

	// A second call for the same ID, with a different slug, must
	// resume the branch already named for FEAT-007 rather than
	// creating a second one.
	branch, created, err := EnsureFeatureBranch(dir, "FEAT-007", "totally different text")
	if err != nil {
		t.Fatalf("second EnsureFeatureBranch() unexpected error: %v", err)
	}
	if created {
		t.Error("created = true on resume, want false")
	}
	if branch != "feat/FEAT-007-user-auth" {
		t.Errorf("branch = %q, want the already-existing %q", branch, "feat/FEAT-007-user-auth")
	}
}

// runGit is a small test-only helper — production code never has a
// reason to run arbitrary Git subcommands beyond IsRepo/CurrentBranch/
// BranchName/EnsureBranch's own fixed set.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}
