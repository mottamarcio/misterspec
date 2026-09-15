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
	"os/exec"
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
// every character `git check-ref-format` itself forbids.
var unsafeBranchChars = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// slugifyForBranch turns free-text (an agent-provided slug) into a
// lowercase, hyphen-separated segment safe to append to a branch name —
// e.g. "User Auth!!" -> "user-auth". Returns "" for input that reduces
// to nothing usable, in which case the caller falls back to no slug.
func slugifyForBranch(s string) string {
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
	if clean := slugifyForBranch(slug); clean != "" {
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
