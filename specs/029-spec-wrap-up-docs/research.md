# Phase 0 Research: `/mister-wrap-up`

No `[NEEDS CLARIFICATION]` markers remained in the Technical Context — the open design questions were resolved directly with the user during `/speckit-specify` (see `checklists/requirements.md` Notes). This file records the resulting decisions plus what grounding the codebase itself provided.

## Decision: Commit range is a new deterministic `internal/vcs` operation, not agent-run `git log`

- **Decision**: Add `CommitsSinceFileAdded(root, path string)` to `internal/vcs`, exposed via a new `internal git commits-since-file --path <path>` CLI command. The Skill calls this instead of running `git log` itself.
- **Rationale**: `internal/vcs`'s own file-level doc comment already establishes it as "the only Git integration point" in misterspec; `022-feature-branch-automation` and `027-constitution-frontmatter-task-deps` both already extended it (or the deterministic layer generally) rather than letting a Skill run raw Git commands, specifically because "which commits count" is a mechanical fact, not a judgment call (Constitution Principle I/II). Leaving this to the agent would reintroduce exactly the inconsistency risk (different agents/models computing the range differently) this project's own established pattern exists to prevent.
- **Alternatives considered**:
  - *Have the Skill run `git log` directly, the way `mister-implement` runs `go test` directly*: rejected — `implement`'s own precedent for running things directly is scoped to the *code the Task's own verification method names* (inherently a moving target the Skill can't know in advance); a Git commit range, by contrast, is exactly the kind of fixed, repeatable computation the deterministic layer exists for.

## Decision: The "added" commit is found via `git log --diff-filter=A --follow`, not commit-message parsing

- **Decision**: The new function locates the commit that introduced `plan.md` using Git's own rename-tracking (`--follow --diff-filter=A`), then returns every commit from that one (inclusive) through `HEAD` (inclusive) on the current branch.
- **Rationale**: This directly satisfies spec.md FR-005 — "no commit-message convention required" — since it relies only on Git's own record of when the file itself first appeared, immune to whatever commit message style a developer already uses (including all history predating this feature).
- **Alternatives considered**:
  - *`git log --grep="SPEC-###"`*: the option the user explicitly declined during clarification — would silently exclude all prior commits made before this feature existed, since nothing enforced that convention.
  - *Commits since the Feature branch's own creation (`022`'s branch-creation commit)*: rejected — a Feature branch can carry multiple Specs; this would over-include commits belonging to a different Spec on the same branch, whereas anchoring on each Spec's own `plan.md` scopes the range correctly per Spec.

## Decision: Graceful degradation returns a bool, not just an error

- **Decision**: `CommitsSinceFileAdded` returns `([]Commit, bool, error)` — the bool explicitly reports whether history could be determined at all (`false` when the project isn't a Git repo, or `plan.md` was never committed), separate from `error`, which is reserved for a genuine operational failure.
- **Rationale**: Mirrors the existing `internal/vcs` convention (`EnsureFeatureBranch`'s own `skippedReason` pattern from `022`) of distinguishing "this legitimately doesn't apply here" from "something went wrong" — spec.md FR-007 requires the document to still generate, with a clear note, when commit history isn't available; conflating that case with a hard error would make the Skill's own Failure Conditions harder to write correctly.
- **Alternatives considered**:
  - *Return an error for the not-a-repo / never-committed case*: rejected — would force the Skill to distinguish "stop entirely" from "continue without this one section" by string-matching an error message, brittle and inconsistent with the project's own established `ok`/`skipped_reason` contract shape.

## Decision: Filename slugification is generalized from the existing branch-slug helper, not duplicated

- **Decision**: `cortex/SPEC-###-<slug>.md`'s slug is produced by the same underlying slugification logic `internal/vcs`'s own `slugifyForBranch` already implements (lowercase, hyphen-separated, unsafe-character-stripped) — generalized into a small reusable helper both call, rather than a second copy.
- **Rationale**: Constitution Principle VI (DRY) — the two use cases (a Git branch segment, a filename segment) have identical safety requirements (no path separators, no unsafe characters, readable at a glance).
- **Alternatives considered**:
  - *A separate, independent slug function for filenames*: rejected — no requirement differs between the two use cases enough to justify a second implementation.

## Decision: `TestSkillsContent_AllNineInstalled` is renamed, not left inaccurate

- **Decision**: Rename the existing test (and its own doc comment, which currently says "exactly §38's nine Skills exist... no more, no fewer") to reflect 10 canonical Skills once `mister-wrap-up` is added, rather than leaving a test whose own name and comment no longer match what it asserts.
- **Rationale**: The test's own logic (driven entirely by `canonicalSkillNames`) needs no behavioral change — only its name and doc comment currently hard-code "nine." Leaving a stale name/comment after this feature would be the same class of documentation drift already fixed elsewhere this session (the `022`-era doc comments earlier corrected during `026`/`027`/`028`'s own investigations).
- **Alternatives considered**:
  - *Leave the name as-is since the logic itself still passes*: rejected — a passing test whose own name asserts something false ("nine" when there are ten) is exactly the kind of stale-but-technically-working state this project's own conventions have consistently corrected rather than tolerated this session.

## Decision: No new `cortex/` configuration field

- **Decision**: `cortex/` at the project root is a fixed convention this feature's Skill creates directly — no new `.misterspec/config.yaml` field (unlike `ArtifactsDir`, `RawDir`, etc., which are all configurable).
- **Rationale**: Constitution Principle IV — configuration fields are added only for values that genuinely vary between projects; nothing in the user's request or the spec's own Assumptions states a need for `cortex/`'s location to vary, and it deliberately sits *outside* the existing `ai/`-rooted, already-configurable hierarchy (spec.md's own Assumptions: "not part of the existing `ai/` semantic-state hierarchy").
- **Alternatives considered**:
  - *Add `CortexDir` to `project.Configuration` for consistency with every other directory field*: considered, but rejected as premature per Principle IV — a legitimate future addition if a real need to relocate it ever appears, not before.
