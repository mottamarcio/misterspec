# Research: Feature-Level Git Branch Automation

## 1. How to talk to Git: shell out vs. a Go Git library

**Decision**: Shell out to the system `git` binary via `os/exec`, with no
new Go module dependency.

**Rationale**: The project's own architecture constraint already names
"History: Git" as a frozen dependency — every user of `misterspec` already
has `git` installed to manage the project this tool operates on. Adding a
pure-Go Git implementation (e.g. `go-git`) to avoid that already-guaranteed
prerequisite would be exactly the kind of speculative complexity Principle
IV (YAGNI) rules out: a heavier dependency, a larger binary, and a second
Git implementation to keep behaviorally identical to the `git` the user
already runs everywhere else in the same project.

**Alternatives considered**:
- `go-git/go-git` (pure Go): rejected — adds a non-trivial dependency and
  binary size increase for four simple, already-standard porcelain
  operations (`rev-parse`, `branch --show-current`, `checkout [-b]`), and
  risks subtle behavioral drift from the user's own installed `git`.
- Reimplementing the needed Git plumbing by hand (reading `.git/HEAD`
  directly): rejected — reinvents a well-tested tool for no benefit, and
  breaks the moment `.git` isn't a plain directory (worktrees, submodules).

## 2. Detecting a Git repository and current branch

**Decision**: `git rev-parse --is-inside-work-tree` (run with the target
project root as the working directory) to detect a usable Git repository;
`git branch --show-current` to read the current branch. Both are ordinary,
well-established porcelain-adjacent commands already used by this
project's own `.specify/extensions/git` tooling.

**Rationale**: `rev-parse --is-inside-work-tree` prints `true` and exits 0
inside a work tree (including one with zero commits), and exits non-zero
outside one — the exact binary signal FR-003's graceful-skip requirement
needs. `branch --show-current` prints an empty string on a detached HEAD,
which is treated as "no branch to compare against" (no mismatch warning is
possible without a named branch), never an error.

**Alternatives considered**:
- Checking for a `.git` directory directly: rejected — fails for worktrees
  and any non-standard `.git` layout `rev-parse` already handles correctly.
- `git status --porcelain` to also detect the branch: rejected — conflates
  two different questions (repo detection, current branch) into one command
  whose output shape would need extra parsing this project doesn't need.

## 3. Branch naming scheme (amended after initial validation)

**Decision**: `feat/<FEATURE-ID>`, optionally suffixed with a slugified,
caller-supplied human-readable summary — e.g. `feat/FEAT-007` or
`feat/FEAT-007-user-auth`. The `--slug` flag already accepted by
`internal create` becomes optional for `feature` (it stays required for
Knowledge/Learning, unchanged) and, when given, is sanitized
(lowercased, non-alphanumeric runs collapsed to a single `-`) before
being appended.

**Rationale**: The ID-only scheme (`feat/FEAT-007`) shipped first and was
validated end-to-end, but real usage surfaced that a bare ID is hard to
recognize at a glance in a `git branch --list` — exactly the readability
problem this whole feature exists to fix one level up (a dedicated branch
that still requires opening the Feature artifact to know what it's for
defeats part of the point). Feature IDs remain solely responsible for
uniqueness — the slug is cosmetic only, sanitized rather than trusted
verbatim (an agent could otherwise pass text containing characters Git
branch names forbid), and never itself looked up: **finding** a Feature's
branch (`vcs.BranchForID`, decision #3-amended's own new lookup) matches
by ID prefix, ignoring whatever slug text was used, so a branch created
with one slug is still found correctly by a later call that requests a
different (or no) slug for the same ID.

**Alternatives considered**:
- Bare `FEAT-007` (no prefix at all, not even `feat/`): rejected — reads
  ambiguously next to ordinary branch names like `dev`.
- Requiring the slug (making it non-optional for Feature creation):
  rejected — would force every caller (including existing automation) to
  supply one; optional preserves backward compatibility with the
  ID-only scheme already shipped and tested.
- Storing the chosen branch name in `.misterspec/` so a later Spec
  creation doesn't need to search for it: rejected — reintroduces
  exactly the persistent, filesystem-external state Constitution
  Principle III forbids; a live `git branch --list` lookup by ID prefix
  is cheap and keeps Git itself as the only source of truth for "what
  branch belongs to this Feature."

## 4. Idempotent branch creation (resuming after interruption)

**Decision**: Before creating, check whether the target branch name already
exists locally (`git rev-parse --verify --quiet refs/heads/<name>`); if it
does, `git checkout <name>`; if not, `git checkout -b <name>`.

**Rationale**: FR-007 requires resuming rather than failing when a prior,
interrupted `create feature` call already made the branch. Checking
existence first (rather than attempting `checkout -b` and pattern-matching
its failure message) keeps the two code paths explicit and avoids relying
on `git`'s own human-readable error text, which is not a stable contract
across Git versions/locales.

**Alternatives considered**:
- `git checkout -B <name>` (force-create/reset): rejected — `-B` resets an
  existing branch to the current `HEAD`, which would silently discard any
  commits already made on a resumed branch; this violates the
  non-destructive guarantee in FR-008/Constitution Principle VIII.

## 5. Where this logic lives in the package graph

**Decision**: A new leaf package, `internal/vcs`, exposing a small,
consumer-shaped surface (`IsRepo`, `CurrentBranch`, `EnsureBranch`,
`BranchName`), consumed only by `internal/operations`.

**Rationale**: Mirrors the existing separation between `ids` (allocates),
`validation` (validates), and `operations` (orchestrates) — Principle VI's
own SRP guidance. `operations.Create` gains a few lines calling into
`internal/vcs`; it does not grow Git-specific logic of its own.

**Alternatives considered**:
- Inlining `os/exec` calls directly inside `operations/create.go`: rejected
  — blurs orchestration with a distinct external-tool integration concern,
  and makes the Git behavior harder to unit-test in isolation from the
  rest of `Create`'s own logic.

## 6. Configuration toggle

**Decision**: Add `GitBranchAutomation bool` (YAML key
`git_branch_automation`) to `project.Configuration`, defaulting to `true`
when the key is absent, following the exact same pointer-based
"absent means default" pattern `Load` already uses for every other
optional field.

**Rationale**: Directly implements User Story 3/FR-006. Defaulting to `true`
means the fix ships active for everyone by default (the whole point of this
feature), while still respecting Principle IV by making the opt-out a
single boolean rather than a richer policy object no current requirement
asks for.

**Alternatives considered**:
- A separate `.misterspec/branching.yaml` file: rejected — a second
  configuration file for one boolean is exactly the kind of unjustified
  new surface Principle IV rules out.
- An environment variable instead of a config field: rejected — every
  other per-project setting already lives in `.misterspec/config.yaml`;
  splitting one setting into a different mechanism would be inconsistent
  with zero benefit.

## 7. Surfacing outcomes without failing the caller

**Decision**: Extend `operations.CreateResult` (and the CLI's JSON success
payload) with explicit fields — a nullable/omitted `git` object carrying
`branch`, `created` (bool), `skipped_reason` (string, e.g.
`"not_a_git_repo"` or `"disabled"`), and, only for Spec creation, `warning`
when the current branch doesn't match the parent Feature's own branch.
Nothing here becomes a new error path — every one of these is informational
metadata on an otherwise-successful `create` call (Principle IX).

**Rationale**: FR-003 and FR-005 both require that these conditions never
block or fail the underlying creation — they need to be visible, not fatal.
A dedicated `git` object (rather than flattening fields into the top level)
keeps the payload self-documenting and matches this project's existing
convention of nesting related outcome data (see `bootstrap`'s own `agent`
object in `misterspec init`'s payload).
