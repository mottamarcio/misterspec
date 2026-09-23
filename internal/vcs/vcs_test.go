package vcs

import (
	"os"
	"os/exec"
	"path/filepath"
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

func TestCommitsSinceFileAdded_OrdinaryCaseExcludesCommitsBeforeFileAdded(t *testing.T) {
	dir := initFixtureRepo(t)

	// A commit before plan.md exists at all.
	writeFile(t, dir, "unrelated.txt", "first")
	runGit(t, dir, "add", "unrelated.txt")
	runGit(t, dir, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-q", "-m", "unrelated change")

	// The commit that adds plan.md itself.
	writeFile(t, dir, "plan.md", "the plan")
	runGit(t, dir, "add", "plan.md")
	runGit(t, dir, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-q", "-m", "add plan.md")

	// A commit after plan.md was added.
	writeFile(t, dir, "plan.md", "the plan, revised")
	runGit(t, dir, "add", "plan.md")
	runGit(t, dir, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-q", "-m", "revise plan.md")

	commits, available, err := CommitsSinceFileAdded(dir, "plan.md")
	if err != nil {
		t.Fatalf("CommitsSinceFileAdded() unexpected error: %v", err)
	}
	if !available {
		t.Fatal("available = false, want true")
	}
	if len(commits) != 2 {
		t.Fatalf("commits = %d, want 2 (the add + the revise, never the unrelated commit before it): %+v", len(commits), commits)
	}
	for _, c := range commits {
		if strings.Contains(c.Subject, "unrelated") {
			t.Errorf("commits includes a commit before plan.md was added: %+v", c)
		}
	}
}

func TestCommitsSinceFileAdded_FileAddedAtRootCommit(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	writeFile(t, dir, "plan.md", "the plan")
	runGit(t, dir, "add", "plan.md")
	runGit(t, dir, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-q", "-m", "root commit adds plan.md")

	commits, available, err := CommitsSinceFileAdded(dir, "plan.md")
	if err != nil {
		t.Fatalf("CommitsSinceFileAdded() unexpected error: %v", err)
	}
	if !available {
		t.Fatal("available = false, want true")
	}
	if len(commits) != 1 {
		t.Fatalf("commits = %d, want 1 (the root commit itself): %+v", len(commits), commits)
	}
}

func TestCommitsSinceFileAdded_FileNeverCommitted(t *testing.T) {
	dir := initFixtureRepo(t)

	commits, available, err := CommitsSinceFileAdded(dir, "never-committed.md")
	if err != nil {
		t.Fatalf("CommitsSinceFileAdded() unexpected error: %v", err)
	}
	if available {
		t.Error("available = true for a file never committed, want false")
	}
	if len(commits) != 0 {
		t.Errorf("commits = %d, want 0", len(commits))
	}
}

func TestCommitsSinceFileAdded_NotARepo(t *testing.T) {
	dir := t.TempDir()

	commits, available, err := CommitsSinceFileAdded(dir, "plan.md")
	if err != nil {
		t.Fatalf("CommitsSinceFileAdded() unexpected error: %v", err)
	}
	if available {
		t.Error("available = true outside a Git repository, want false")
	}
	if len(commits) != 0 {
		t.Errorf("commits = %d, want 0", len(commits))
	}
}

// writeFile is a small test-only helper writing a file's content
// directly, for commits this package's own tests need to construct.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", name, err)
	}
}

// TestHeadCommit_ReturnsCurrentShortSHA is 041-task-evidence-fingerprint
// T002 (Foundational): HeadCommit reports the current commit with
// available == true inside a real repository with commits.
func TestHeadCommit_ReturnsCurrentShortSHA(t *testing.T) {
	dir := initFixtureRepo(t)
	want := strings.TrimSpace(runGit(t, dir, "rev-parse", "--short", "HEAD"))

	sha, available, err := HeadCommit(dir)
	if err != nil {
		t.Fatalf("HeadCommit() unexpected error: %v", err)
	}
	if !available {
		t.Fatal("available = false inside a real repo with commits, want true")
	}
	if sha != want {
		t.Errorf("HeadCommit() sha = %q, want %q", sha, want)
	}
}

// TestHeadCommit_NotARepo mirrors CommitsSinceFileAdded's own "bool
// separate from error" convention (research.md #7).
func TestHeadCommit_NotARepo(t *testing.T) {
	dir := t.TempDir()

	sha, available, err := HeadCommit(dir)
	if err != nil {
		t.Fatalf("HeadCommit() unexpected error: %v", err)
	}
	if available {
		t.Error("available = true outside a Git repository, want false")
	}
	if sha != "" {
		t.Errorf("sha = %q, want \"\" when not available", sha)
	}
}

// TestIsWorkingTreeDirty_CleanRepo is 041-task-evidence-fingerprint T003
// (Foundational).
func TestIsWorkingTreeDirty_CleanRepo(t *testing.T) {
	dir := initFixtureRepo(t)

	dirty, err := IsWorkingTreeDirty(dir)
	if err != nil {
		t.Fatalf("IsWorkingTreeDirty() unexpected error: %v", err)
	}
	if dirty {
		t.Error("IsWorkingTreeDirty() = true for a freshly-committed repo with no local changes, want false")
	}
}

func TestIsWorkingTreeDirty_UncommittedFile(t *testing.T) {
	dir := initFixtureRepo(t)
	writeFile(t, dir, "untracked.txt", "local change")

	dirty, err := IsWorkingTreeDirty(dir)
	if err != nil {
		t.Fatalf("IsWorkingTreeDirty() unexpected error: %v", err)
	}
	if !dirty {
		t.Error("IsWorkingTreeDirty() = false after writing an uncommitted file, want true")
	}
}

func TestIsWorkingTreeDirty_NotARepo(t *testing.T) {
	dir := t.TempDir()

	if _, err := IsWorkingTreeDirty(dir); err == nil {
		t.Error("IsWorkingTreeDirty() error = nil outside a Git repository, want non-nil")
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

// commitAll is a small test-only helper: stage everything and commit,
// returning the new commit's short SHA (042-impact-analysis-review
// T002/T003 fixtures).
func commitAll(t *testing.T, dir, message string) string {
	t.Helper()
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-q", "-m", message)
	return strings.TrimSpace(runGit(t, dir, "rev-parse", "--short", "HEAD"))
}

// TestDiffNameStatus_AddedModifiedRemoved is 042-impact-analysis-review
// T002 (Foundational): DiffNameStatus reports an added, a modified, and
// a removed path correctly between two commits (contracts §1).
func TestDiffNameStatus_AddedModifiedRemoved(t *testing.T) {
	dir := initFixtureRepo(t)
	writeFile(t, dir, "keep.md", "v1")
	writeFile(t, dir, "remove.md", "v1")
	from := commitAll(t, dir, "base")

	writeFile(t, dir, "keep.md", "v2")
	writeFile(t, dir, "added.md", "new")
	if err := os.Remove(filepath.Join(dir, "remove.md")); err != nil {
		t.Fatalf("removing remove.md: %v", err)
	}
	to := commitAll(t, dir, "change")

	entries, err := DiffNameStatus(dir, from, to)
	if err != nil {
		t.Fatalf("DiffNameStatus() unexpected error: %v", err)
	}

	got := map[string]string{}
	for _, e := range entries {
		got[e.Path] = e.Status
	}
	if got["added.md"] != "A" {
		t.Errorf("added.md status = %q, want \"A\"", got["added.md"])
	}
	if got["keep.md"] != "M" {
		t.Errorf("keep.md status = %q, want \"M\"", got["keep.md"])
	}
	if got["remove.md"] != "D" {
		t.Errorf("remove.md status = %q, want \"D\"", got["remove.md"])
	}
}

// TestDiffNameStatus_EmptyToDiffsAgainstWorkingTree is 042-impact-
// analysis-review T002: to == "" diffs from against the working tree,
// so an uncommitted edit shows up (contracts §1).
func TestDiffNameStatus_EmptyToDiffsAgainstWorkingTree(t *testing.T) {
	dir := initFixtureRepo(t)
	writeFile(t, dir, "keep.md", "v1")
	from := commitAll(t, dir, "base")

	writeFile(t, dir, "keep.md", "v2 uncommitted")

	entries, err := DiffNameStatus(dir, from, "")
	if err != nil {
		t.Fatalf("DiffNameStatus() unexpected error: %v", err)
	}
	if len(entries) != 1 || entries[0].Path != "keep.md" || entries[0].Status != "M" {
		t.Errorf("DiffNameStatus() = %+v, want one modified entry for keep.md", entries)
	}
}

// TestDiffNameStatus_FromNotFound is 042-impact-analysis-review T002:
// an unresolvable from returns a non-nil error.
func TestDiffNameStatus_FromNotFound(t *testing.T) {
	dir := initFixtureRepo(t)

	if _, err := DiffNameStatus(dir, "does-not-exist", ""); err == nil {
		t.Error("DiffNameStatus() error = nil for an unresolvable revision, want non-nil")
	}
}

// TestFileAtRevision_ExistsAtCommit is 042-impact-analysis-review T003
// (Foundational): FileAtRevision returns a file's content and found ==
// true at a commit where it exists (contracts §1).
func TestFileAtRevision_ExistsAtCommit(t *testing.T) {
	dir := initFixtureRepo(t)
	writeFile(t, dir, "spec.md", "hello")
	rev := commitAll(t, dir, "add spec.md")

	content, found, err := FileAtRevision(dir, "spec.md", rev)
	if err != nil {
		t.Fatalf("FileAtRevision() unexpected error: %v", err)
	}
	if !found {
		t.Fatal("found = false at the commit that added the file, want true")
	}
	if string(content) != "hello" {
		t.Errorf("content = %q, want %q", content, "hello")
	}
}

// TestFileAtRevision_NotYetAdded mirrors CommitsSinceFileAdded's own
// "bool separate from error" convention: found is false, with no
// error, at a commit before the file existed.
func TestFileAtRevision_NotYetAdded(t *testing.T) {
	dir := initFixtureRepo(t)
	writeFile(t, dir, "other.md", "unrelated")
	before := commitAll(t, dir, "base without spec.md")
	writeFile(t, dir, "spec.md", "hello")
	commitAll(t, dir, "add spec.md")

	content, found, err := FileAtRevision(dir, "spec.md", before)
	if err != nil {
		t.Fatalf("FileAtRevision() unexpected error: %v", err)
	}
	if found {
		t.Error("found = true at a commit before the file existed, want false")
	}
	if content != nil {
		t.Errorf("content = %q, want nil", content)
	}
}

// TestFileAtRevision_EmptyRevReadsWorkingTree is 042-impact-analysis-
// review T003: rev == "" reads the working-tree file directly,
// reflecting an uncommitted edit.
func TestFileAtRevision_EmptyRevReadsWorkingTree(t *testing.T) {
	dir := initFixtureRepo(t)
	writeFile(t, dir, "spec.md", "committed")
	commitAll(t, dir, "add spec.md")
	writeFile(t, dir, "spec.md", "uncommitted edit")

	content, found, err := FileAtRevision(dir, "spec.md", "")
	if err != nil {
		t.Fatalf("FileAtRevision() unexpected error: %v", err)
	}
	if !found {
		t.Fatal("found = false for an existing working-tree file, want true")
	}
	if string(content) != "uncommitted edit" {
		t.Errorf("content = %q, want %q", content, "uncommitted edit")
	}
}
