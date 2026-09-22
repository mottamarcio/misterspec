// Package vcs is a thin wrapper around the system `git` binary — the
// only Git integration point in misterspec (022-feature-branch-
// automation). It shells out via os/exec rather than embedding a Git
// implementation (research.md #1): every user of misterspec already
// has git installed to manage the project this tool operates on.
//
// Every operation here is non-destructive: EnsureBranch only ever
// creates-and-checks-out a new branch or checks out an existing one —
// never resets, force-pushes, or deletes a branch (Constitution
// Principle VIII, research.md #4).
package vcs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// IsRepo reports whether root is inside a Git work tree.
func IsRepo(root string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = root
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// CurrentBranch returns the currently checked-out branch name in root,
// or "" if HEAD is detached or root is not a Git repository.
func CurrentBranch(root string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// unsafeBranchChars matches every run of characters not safe to use
// verbatim in a Git branch name segment — a deliberately conservative
// allow-list (letters, digits, hyphen) rather than trying to enumerate
// every character `git check-ref-format` itself forbids. Reused as-is
// for filename slugs (Slugify, 029-spec-wrap-up-docs) — the safety
// requirements are identical for both use cases.
var unsafeBranchChars = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// Slugify turns free-text into a lowercase, hyphen-separated segment
// safe to use in a Git branch name or a filename — e.g. "User Auth!!"
// -> "user-auth". Returns "" for input that reduces to nothing usable,
// in which case the caller falls back to no slug (Constitution
// Principle VI — one implementation shared by both use cases).
func Slugify(s string) string {
	s = unsafeBranchChars.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return strings.ToLower(s)
}

// BranchName derives the branch name for a Feature ID, optionally with
// a human-readable slug appended for readability (research.md #3,
// amended) — e.g. BranchName("FEAT-007", "user auth") ->
// "feat/FEAT-007-user-auth". An empty or unusable slug falls back to
// the bare "feat/FEAT-007". Pure function, no I/O.
func BranchName(featureID, slug string) string {
	base := "feat/" + featureID
	if clean := Slugify(slug); clean != "" {
		return base + "-" + clean
	}
	return base
}

// BranchForID returns the local branch already dedicated to featureID
// — any branch named exactly "feat/<featureID>" or starting with
// "feat/<featureID>-" — or "" if none exists yet. Since a Feature ID is
// unique by construction, at most one such branch should ever exist;
// this looks it up live rather than storing it anywhere (Constitution
// Principle III), so a branch created with one slug is still found
// correctly even if a later call names a different (or no) slug.
func BranchForID(root, featureID string) (string, error) {
	cmd := exec.Command("git", "branch", "--format=%(refname:short)")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	base := "feat/" + featureID
	for _, name := range strings.Fields(string(out)) {
		if name == base || strings.HasPrefix(name, base+"-") {
			return name, nil
		}
	}
	return "", nil
}

// EnsureBranch checks out the named branch in root, creating it from
// the current HEAD first if it does not already exist locally
// (research.md #4). Returns whether the branch was newly created.
func EnsureBranch(root, name string) (created bool, err error) {
	verify := exec.Command("git", "rev-parse", "--verify", "--quiet", "refs/heads/"+name)
	verify.Dir = root
	exists := verify.Run() == nil

	args := []string{"checkout", "-q", name}
	if !exists {
		args = []string{"checkout", "-q", "-b", name}
	}

	checkout := exec.Command("git", args...)
	checkout.Dir = root
	if out, err := checkout.CombinedOutput(); err != nil {
		return false, fmt.Errorf("vcs: git %v: %w: %s", args, err, out)
	}

	return !exists, nil
}

// Commit is one entry in the range CommitsSinceFileAdded returns —
// never persisted, always derived live from Git (Constitution
// Principle III).
type Commit struct {
	Hash       string
	Subject    string
	AuthorDate string
}

// commitLogFormat is the exact "git log --format" string
// CommitsSinceFileAdded parses, and the separator its own parsing
// relies on — %x1f is the ASCII unit-separator, a character no commit
// subject or date could plausibly contain.
const commitLogFormat = "%H%x1f%s%x1f%as"

// CommitsSinceFileAdded returns every commit on the current branch in
// root from the one that first added path (inclusive) through HEAD
// (inclusive), newest first — matching git log's own default order.
// available is false, with no error, when history cannot be
// determined at all: root is not a Git repository, or path was never
// committed (029-spec-wrap-up-docs, research.md's own "bool separate
// from error" decision, mirroring 022's GitSkippedReason pattern) —
// this is never a Failure Condition for a caller, only a fact to note.
func CommitsSinceFileAdded(root, path string) (commits []Commit, available bool, err error) {
	if !IsRepo(root) {
		return nil, false, nil
	}

	// git log lists newest-first; --follow --diff-filter=A finds every
	// commit that added path (across renames), so the LAST line is the
	// original "added" commit.
	added := exec.Command("git", "log", "--diff-filter=A", "--follow", "--format=%H", "--", path)
	added.Dir = root
	out, err := added.Output()
	if err != nil {
		return nil, false, fmt.Errorf("vcs: git log --diff-filter=A -- %s: %w", path, err)
	}
	lines := strings.Fields(strings.TrimSpace(string(out)))
	if len(lines) == 0 {
		return nil, false, nil
	}
	addedCommit := lines[len(lines)-1]

	// <addedCommit>^..HEAD excludes addedCommit itself unless it has no
	// parent (the repository's own root commit), in which case Git
	// already includes it as the range's own start.
	rangeSpec := addedCommit + "^..HEAD"
	if !hasParent(root, addedCommit) {
		rangeSpec = "HEAD"
	}

	log := exec.Command("git", "log", "--format="+commitLogFormat, rangeSpec)
	log.Dir = root
	logOut, err := log.Output()
	if err != nil {
		return nil, false, fmt.Errorf("vcs: git log %s: %w", rangeSpec, err)
	}

	for _, line := range strings.Split(strings.TrimRight(string(logOut), "\n"), "\n") {
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, "\x1f", 3)
		if len(fields) != 3 {
			continue
		}
		commits = append(commits, Commit{Hash: fields[0], Subject: fields[1], AuthorDate: fields[2]})
	}
	return commits, true, nil
}

// hasParent reports whether commit has at least one parent — false
// only for a repository's own root commit.
func hasParent(root, commit string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", "--quiet", commit+"^")
	cmd.Dir = root
	return cmd.Run() == nil
}

// HeadCommit returns the current commit's short SHA. available is
// false, with no error, when root is not a Git repository or has no
// commits yet — mirrors CommitsSinceFileAdded's own "bool separate from
// error" convention (041-task-evidence-fingerprint research.md #7).
func HeadCommit(root string) (sha string, available bool, err error) {
	if !IsRepo(root) {
		return "", false, nil
	}

	// "No commits yet" (a brand-new, empty repository) is checked first,
	// by exit code alone — the same --verify --quiet technique
	// EnsureBranch already uses below — rather than by matching the
	// short-SHA command's own stderr text, which varies across Git
	// versions. Code review finding (041-task-evidence-fingerprint): the
	// original implementation treated *any* failure of the short-SHA
	// command as "not available, no error," silently masking a genuine
	// Git failure (corrupted repo, permission error, git binary issue)
	// as a normal, empty-repo absence.
	verify := exec.Command("git", "rev-parse", "--verify", "--quiet", "HEAD")
	verify.Dir = root
	if verify.Run() != nil {
		return "", false, nil
	}

	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", false, err
	}
	return strings.TrimSpace(string(out)), true, nil
}

// IsWorkingTreeDirty reports whether `git status --porcelain` returns
// any output at all — repository-wide, not scoped to specific paths, a
// deliberate simplification (041-task-evidence-fingerprint research.md
// #7): a false "dirty" costs nothing, while a missed "actually dirty"
// would misrepresent what evidence was captured against.
func IsWorkingTreeDirty(root string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("vcs: git status --porcelain: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// DiffEntry is one changed path between two revisions (042-impact-
// analysis-review contracts §1).
type DiffEntry struct {
	Path string
	// Status is git's own --name-status letter: "A" (added), "M"
	// (modified), or "D" (deleted) — the letters this feature's callers
	// consume are exactly these three; a rename/copy letter (R/C) is not
	// expected here since this package never passes -M/-C to git diff.
	Status string
}

// DiffNameStatus reports every path that differs between from and to,
// via `git diff --name-status <from> [<to>]`. to == "" diffs against
// the working tree (plain `git diff <from>`), matching
// IsWorkingTreeDirty's own working-tree scope (042-impact-analysis-
// review research.md #2).
func DiffNameStatus(root, from, to string) ([]DiffEntry, error) {
	args := []string{"diff", "--name-status", from}
	if to != "" {
		args = append(args, to)
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("vcs: git %v: %w", args, err)
	}

	var entries []DiffEntry
	for _, line := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, "\t", 2)
		if len(fields) != 2 {
			continue
		}
		entries = append(entries, DiffEntry{Path: fields[1], Status: fields[0]})
	}
	return entries, nil
}

// FileAtRevision returns path's content at rev via `git show
// <rev>:<path>`. found is false, with no error, when path did not
// exist at rev (an added file — mirrors CommitsSinceFileAdded's own
// "bool separate from error" convention). rev == "" reads the
// working-tree file directly (os.ReadFile), not through git show
// (042-impact-analysis-review contracts §1).
func FileAtRevision(root, path, rev string) (content []byte, found bool, err error) {
	if rev == "" {
		data, readErr := os.ReadFile(filepath.Join(root, path))
		if readErr != nil {
			if os.IsNotExist(readErr) {
				return nil, false, nil
			}
			return nil, false, readErr
		}
		return data, true, nil
	}

	cmd := exec.Command("git", "show", rev+":"+path)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		// git show exits non-zero both for "path does not exist at rev"
		// and for a genuine error (bad rev, not a repo); the former is
		// the overwhelmingly common case for this feature's own caller
		// (a path added later than `from`), so it is treated as "not
		// found" rather than surfaced as an error, matching
		// CommitsSinceFileAdded's existing convention for a comparable
		// absence.
		return nil, false, nil
	}
	return out, true, nil
}

// EnsureFeatureBranch creates-and-checks-out featureID's own branch —
// deriving its name from featureID and, if given, slug — or resumes
// the branch already dedicated to that Feature ID if one exists
// (BranchForID), regardless of what slug this call was given: an
// already-named branch always wins over a newly requested name, so a
// Feature ID never ends up with two branches (FR-007).
func EnsureFeatureBranch(root, featureID, slug string) (branch string, created bool, err error) {
	existing, err := BranchForID(root, featureID)
	if err != nil {
		return "", false, err
	}

	name := existing
	if name == "" {
		name = BranchName(featureID, slug)
	}

	created, err = EnsureBranch(root, name)
	if err != nil {
		return "", false, err
	}
	return name, created, nil
}
