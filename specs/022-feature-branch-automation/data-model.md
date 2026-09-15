# Data Model: Feature-Level Git Branch Automation

No new persistent entity is introduced (Constitution Principle III — the
filesystem/Git state itself is authoritative; nothing new is recorded in
`.misterspec/`). This documents the shapes that change.

## `project.Configuration` (extended)

| Field | YAML key | Type | Default | Notes |
|---|---|---|---|---|
| `GitBranchAutomation` | `git_branch_automation` | `bool` | `true` | New. Absent key defaults to `true`, following `Load`'s existing pointer-based "absent means default" pattern (matches every other optional field already in this struct). |

## `operations.CreateResult` (extended)

| Field | Type | Present when | Notes |
|---|---|---|---|
| `ID`, `Path` | (existing) | always | Unchanged. |
| `GitBranch` | `string` | a Git-tracked project, automation enabled, and `Type` is `Feature` or `Spec` | The relevant branch name — for `Feature`, the branch just ensured/created; for `Spec`, the parent Feature's own branch. Empty string when not applicable. |
| `GitBranchCreated` | `bool` | `Type == Feature`, a Git-tracked project, automation enabled | `true` only when the branch did not already exist and was newly created (FR-007's resume path reports `false`). |
| `GitSkippedReason` | `string` | automation would apply but did not run | One of `"not_a_git_repo"` or `"disabled"`; empty when automation ran normally or when the entity type never triggers Git behavior (Task, Knowledge, Learning, Program). |
| `GitWarning` | `string` | `Type == Spec`, a Git-tracked project, automation enabled, and the current branch differs from the parent Feature's own branch | Human-readable mismatch notice (FR-005); empty otherwise. |

## `internal/vcs` package surface (amended — slug support)

Pure orchestration inputs/outputs — no new stored state.

```go
// IsRepo reports whether root is inside a Git work tree.
func IsRepo(root string) bool

// CurrentBranch returns the currently checked-out branch name in root,
// or "" if HEAD is detached or root is not a Git repository.
func CurrentBranch(root string) (string, error)

// BranchName derives a branch name for a Feature ID, optionally with a
// slugified human-readable summary appended (e.g. "FEAT-007", "user
// auth" -> "feat/FEAT-007-user-auth"; "" slug -> "feat/FEAT-007").
// Pure function, no I/O.
func BranchName(featureID, slug string) string

// BranchForID returns the local branch already dedicated to featureID
// (matched by ID prefix, ignoring whatever slug it was created with),
// or "" if none exists yet.
func BranchForID(root, featureID string) (string, error)

// EnsureBranch checks out the named branch in root, creating it from
// the current HEAD first if it does not already exist locally. Returns
// whether the branch was newly created.
func EnsureBranch(root, name string) (created bool, err error)

// EnsureFeatureBranch creates-and-checks-out featureID's own branch
// (BranchName, using slug if given), or resumes the branch BranchForID
// already finds for that ID — an existing name always wins over a
// newly requested slug, so an ID never ends up with two branches.
func EnsureFeatureBranch(root, featureID, slug string) (branch string, created bool, err error)
```

## `operations.CreateRequest` (extended)

| Field | Type | Applies to | Notes |
|---|---|---|---|
| `Slug` | `string` | (existing) required for Knowledge/Learning; **now also accepted, optionally, for Feature** | For Feature, used only to make the Git branch name readable (via `vcs.BranchName`) — never validated the way Knowledge/Learning's slug is (no path-safety check needed, since `vcs.BranchName` sanitizes it for Git-ref safety regardless of content), and never part of the artifact's own path. |

## Relationships

- One Feature ↔ exactly one Git branch, identified by ID prefix
  (`feat/<Feature ID>` or `feat/<Feature ID>-<anything>`) — looked up
  live via `vcs.BranchForID`, never stored. The optional slug affects
  only what that branch is initially named, never how it is found
  afterward.
- Every Spec's parent Feature is already resolvable via the existing
  parent-ID relationship `operations.Create` already validates for Spec
  creation; the Spec's own "expected branch" for mismatch-warning
  purposes is `vcs.BranchForID(root, parentFeatureID)` — empty when the
  parent Feature has no dedicated branch yet (e.g. created before
  automation was enabled), in which case no warning is possible.
