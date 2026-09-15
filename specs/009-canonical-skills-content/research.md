# Phase 0 Research: Canonical Skills Content

Unlike 001-008, most of this feature's substance is content (nine
`SKILL.md` files), not Go behavior — but one real, load-bearing
technical gap surfaced while grounding `docs/architecture-specification.md`
§38-49 against what 005/006/008 actually built, and it must be resolved
before any Skill content can be installed correctly at all. This
document records that decision first, then the content-shape decisions.

## Decision: `internal/installer`'s `ListFS`/`InstallFS` become recursive — the central technical enabler

- **Decision**: `ListFS` and `InstallFS` walk their source `fs.FS`
  recursively (via `fs.WalkDir`) instead of a single-level
  `fs.ReadDir`, and `Resource.Name` becomes the resource's path
  *relative to `sourceDir`* (e.g. `"create-plan/SKILL.md"`), not
  necessarily a bare filename. `Install`/`List` (the template-specific
  wrappers) and every existing caller's exported signature are
  unchanged.
- **Rationale**: A real coding-agent Skill is a *directory* containing
  `SKILL.md` — confirmed by this very project's own installed Spec Kit
  skills (`.claude/skills/speckit-specify/SKILL.md`, etc.) and by
  006-agent-adapter's `claude.Install` already targeting
  `.claude/skills` as a directory. `internal/installer.ListFS`'s
  current implementation (005-embedded-kit) explicitly skips directory
  entries (`if entry.IsDir() { continue }`) — it was never asked to
  install anything nested before, because kit templates and every
  fixture used in 005/006/008's own tests happen to be flat. Nine real
  Skills, each needing `<skill-name>/SKILL.md`, are the first resource
  set that actually requires recursion. Without this change, every
  Skill directory would simply be silently skipped and nothing would
  install — a correctness gap, not a style preference.
- **Backward compatibility, verified**: `WriteAtomicFile`
  (006-agent-adapter) already `os.MkdirAll`s a target's parent
  directory before writing, so nested destinations need no further
  change there. `kit.TemplatesFS` is flat (no subdirectories), so for
  every existing template caller `Resource.Name` is unchanged in
  practice (a relative path with zero intermediate directories *is* a
  bare filename). Every existing fixture in
  `internal/installer/installer_test.go` and
  `internal/agents/claude/claude_test.go` (`fixtureSkillsFS()`,
  `fixtureSkills()`) is likewise flat. This is confirmed, not assumed —
  005/006/008's full suites are named regression gates for this change
  (tasks.md), and the recursive walk is expected to produce byte-for-byte
  identical `Resource` sets for every one of them.
- **Path handling**: a discovered path uses `fs.FS`'s own forward-slash
  convention (e.g. `"create-plan/SKILL.md"`); converting it to an OS
  filesystem path for the write destination uses
  `filepath.FromSlash(r.Name)` before `filepath.Join`, so the change is
  correct on a non-Unix host too, not just coincidentally correct on
  Linux via matching separators.
- **Containment unaffected**: `installOneFS`'s existing
  `artifacts.RelativeWithinRoot` check operates on whatever
  `Resource.Name` is, regardless of how it was discovered — a multi-segment
  relative name is checked exactly the same way a single-segment one
  already is (`containment_test.go`'s malicious-name guard needs no
  change).
- **Alternatives considered**: Keeping `ListFS`/`InstallFS` flat and
  instead flattening each Skill into a single file with an unconventional
  name (e.g. `kit/skills/create-plan.md`) — rejected; this would produce
  Skills the coding agent's own native discovery convention cannot
  actually find (FR-006, SC-001 would fail), defeating the feature's
  purpose. Writing a second, Skill-specific installer function alongside
  the existing flat one — rejected as exactly the kind of duplicated
  mechanism Constitution Principle VI (DRY) and 006/007's own precedent
  (reuse `installer`, never reimplement it) rule out.

## Decision: `kit/skills/README.md` (008's compile placeholder) is removed, not kept alongside real content

- **Decision**: The one placeholder file 008-cli-cobra added so
  `//go:embed skills` could compile is deleted; nine real skill
  directories replace it, and the embed now has real content to match
  against with no placeholder needed.
- **Rationale**: The placeholder's own doc comment (008's research.md)
  said explicitly it exists only until real Skills exist. Leaving it in
  place once real content exists would mean every `misterspec init`
  installs a stray, meaningless `README.md` into `.claude/skills/`
  alongside the nine real Skills — confusing to both the coding agent
  (which might mistake it for something to load) and a human browsing
  the directory.
- **Alternatives considered**: Keeping it as general kit documentation
  — rejected; `kit/skills/`'s own package-level Go doc comment
  (`kit/kit.go`) is the right place for that, not a file that gets
  installed into every project.

## Decision: Each `SKILL.md` is real Claude Code frontmatter plus §39's required body structure

- **Decision**: Every `kit/skills/<name>/SKILL.md` starts with a YAML
  frontmatter block carrying at least `name` and `description` (the
  minimum this project's own installed Spec Kit skills already use for
  Claude Code's own Skill discovery), followed by a body organized
  under exactly the H2 headings `docs/architecture-specification.md`
  §39 requires, in that order.
- **Rationale**: The frontmatter is what makes a `SKILL.md` a *loadable
  Claude Code Skill* at all (FR-006) — without it, installing the file
  materializes bytes but the agent never discovers it as an invocable
  capability. §39's H2 structure is misterspec's own, richer,
  deterministic-operations-aware skill contract (FR-002) — layering it
  as the body under a minimal, compatible frontmatter satisfies both
  without inventing a third format.
- **Alternatives considered**: Following only §39's structure with no
  frontmatter — rejected; would not be discoverable by Claude Code at
  all (SC-001 would fail). Following only Claude Code's own convention
  with no §39 structure — rejected; loses the deterministic-operations
  transparency and authority boundaries §39/§40 exist for, and this
  project's constitution (Principle IX) already commits to that
  transparency.

## Decision: Each Skill's "Deterministic Operations" section names only commands 008-cli-cobra actually built

- **Decision**: `docs/architecture-specification.md` §41-49's
  illustrative per-Skill operation contracts are followed exactly
  *except* where they name an operation that does not exist
  (`internal references`, standalone Task ID allocation) — those two
  gaps are resolved per spec.md's own Assumptions:
  - Everywhere §41-49 lists `internal references SPEC-###`, the Skill
    instead uses `internal inspect SPEC-###`'s own `depends_on`/
    `supersedes` fields, already present in its JSON output.
  - `/create-tasks` does not call a Task-ID-allocation command (none
    exists); its Procedure instead instructs the agent to author
    `## TASK-NNN` headings directly into the Spec's `tasks.md` (the
    same human/agent-authored boundary 003-entity-creation's own
    `Create` already drew: Task has no independent file/directory of
    its own), relying on `internal validate`'s existing duplicate-Task-ID
    detection (§17) as the safety net instead of a dedicated allocator.
  - `internal project` (§9.1) is never called — 008-cli-cobra's own
    research.md deliberately did not build a `project` command (no
    `internal/operations` function backs it), and no Skill needs it:
    every other command already fails with `project_not_initialized`
    on an uninitialized target, which is the only thing a Skill would
    otherwise use `internal project` to check for up front.
  - `/create-plan`'s deterministic operations list every other item
    from §46; `/implement` and `/analyze` likewise, minus their own
    `internal references` line.
- **Rationale**: FR-003/SC-004 make this a verifiable requirement, not
  a suggestion — a Skill instructing an agent to run a command that
  does not exist would fail at the moment of use, exactly the kind of
  broken contract this project's whole discipline (frozen JSON shapes,
  stable error codes, no aspirational documentation) exists to prevent.
- **Alternatives considered**: Building `internal references` and a
  Task-ID-allocation command as part of this feature so §41-49 could be
  followed literally — rejected; spec.md's own Assumptions already
  rule this out (new deterministic Go behavior is out of scope for a
  content feature), and neither gap blocks any of the nine Skills from
  being fully usable today.

## Decision: Structural conformance is machine-checked, not only reviewed by eye

- **Decision**: A test (`internal/example` or a small dedicated
  package) parses every installed `SKILL.md`, asserting: every §39
  heading is present, in order; the frontmatter has non-empty `name`/
  `description`; every `misterspec internal <word>` token found in its
  "Deterministic Operations" section names one of the ten commands
  008-cli-cobra actually registered (FR-003/SC-004, machine-verified,
  not eyeballed); the "Completion Contract" section names all five of
  §50's required concepts.
- **Rationale**: Nine hand-written documents drifting from their own
  required shape over time is a realistic risk this project's existing
  test-first discipline (Constitution Principle V) already exists to
  catch early, the same way 004-structural-validation's `Finding`
  codes are checked against `docs/architecture-specification.md` §7's
  list rather than trusted by inspection alone.
- **Alternatives considered**: Manual review only — rejected; every
  other feature in this project has machine-verified its own
  documented contract (JSON shapes, error codes, exit codes); content
  correctness deserves the same discipline, especially since content
  here doubles as the actual product surface end users interact with.

## Output

All unknowns resolved. No `NEEDS CLARIFICATION` markers remain.
