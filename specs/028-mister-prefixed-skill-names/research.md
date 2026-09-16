# Phase 0 Research: `mister-`-Prefixed Skill Names

No `[NEEDS CLARIFICATION]` markers remained in the Technical Context. This file records the decisions already made with the user (both clarifying questions, before `/speckit-specify` wrote the spec) plus what a pre-plan codebase investigation confirmed.

## Decision: No name-transformation layer needed — rename the directories directly

- **Decision**: The rename is exactly "rename 9 directories and fix their content," nothing more — no mapping table, no adapter change, no new abstraction.
- **Rationale**: A dedicated investigation (before this plan was written) confirmed `internal/installer/installer.go`'s `ListFS` uses each file's path relative to the source directory verbatim as the installed resource name; no agent adapter (Claude, Codex, Cursor, Copilot, Devin, agy) renames or prefixes anything. The installed slash-command name *is* the `kit/skills/` directory name, literally, everywhere. There is also no existing "short name" vs. "command name" vocabulary anywhere in the codebase to reuse or extend — today, directory name and command name are the same concept under two labels.
- **Alternatives considered**:
  - *Add a name-mapping/alias layer in the installer so the on-disk directory could keep its old name while installing under a new one*: rejected — pure unjustified complexity (Constitution Principle IV) for a problem that doesn't exist; renaming the directory achieves the identical outcome with zero new code.

## Decision: Full mesh cross-reference edit, not a scripted find-and-replace left unverified

- **Decision**: Every one of the 9 renamed `SKILL.md` files gets its own pass confirming every mention of another Skill (in `## Related Skills`, `## Recommended Next Step`, `## Failure Conditions`, and ordinary prose) uses the new name — verified by a final full-repo search for any remaining old name inside `kit/skills/`, not assumed correct just because a rename script ran.
- **Rationale**: The investigation found a full mesh, not a chain — every one of the 9 Skills references at least two others by name (e.g. `create-plan/SKILL.md` mentions `/implement` in ordinary prose, not just in its own "next step" block). A single missed occurrence would strand a user mid-pipeline with an instruction to run a command that no longer exists (User Story 2's own failure mode). Given ~72 total occurrences across 9 files per the investigation, a final grep-based verification step is cheap insurance against a missed spot.
- **Alternatives considered**:
  - *Trust a single global find-and-replace across `kit/skills/` without a follow-up verification pass*: rejected — plausible but not verified; a final search-for-old-names pass is one command and removes the risk entirely (spec.md's own Acceptance Scenario 2/3, User Story 2, explicitly calls for zero old-name occurrences, which deserves a real check, not an assumption).

## Decision: Update `internal/example/skills_content_test.go` last, using its own failures as the checklist

- **Decision**: Rename the 9 directories and fix their cross-references *first*; only then update the test file. In between, `go test ./internal/example/...` is expected to fail — every `assertSkillConformant(t, "<old-name>")` call, every `skillOperationsAllowlist` lookup by old key, and the three feature-specific tests' literal `fs.ReadFile(kit.SkillsFS, "<old-name>/SKILL.md")` calls will all fail to find their target once the directories no longer exist under the old names.
- **Rationale**: This is the same test-first-flavored discipline used for every prompt-content feature this session (specs 026, 027) — except here the "test written first, confirmed failing" step is free: the existing suite *becomes* the failing baseline the moment the rename happens, with no new test to author. `canonicalSkillNames` already drives `TestSkillsContent_AllNineInstalled`'s real end-to-end install-and-byte-compare check, so updating that one slice (plus the map keys and the three literal-path tests) is the entire remaining fix — and the suite going green afterward is direct proof the rename plus cross-reference edits are complete and correct.
- **Alternatives considered**:
  - *Update the test file first, then rename directories*: rejected — inverts the natural verification signal for no benefit; the test file's own edits are trivial (string literals only), so there's no reason to front-load them before the actual content work they're meant to verify.

## Decision: Historical Specs (009, 018, and all others) are never rewritten

- **Decision**: Specs already completed under the old names — their own `spec.md`, `plan.md`, `tasks.md`, etc. — are left exactly as they are, even where they mention an old command name like `/create-tasks`.
- **Rationale**: Those documents record what was actually specified, planned, and run *at the time* — rewriting them to reflect a later rename would falsify the historical record (the same principle already applied earlier this session when spec 027's own `checklists/requirements.md` recorded a correction as a new note rather than silently editing away what had been claimed before). Only currently-active `kit/skills/` content and its own automated verification are in scope, per spec.md's own Assumptions.
- **Alternatives considered**:
  - *Also update historical specs' own text for "consistency"*: rejected — those documents aren't currently-active product surface; touching them serves no user-facing purpose and actively destroys the accuracy of the project's own history.
