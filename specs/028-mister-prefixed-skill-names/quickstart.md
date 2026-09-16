# Quickstart: Validate `mister-`-Prefixed Skill Names

## Prerequisites

- A local checkout of this repo on branch `028-mister-prefixed-skill-names` (or later, once merged).
- Go 1.23.4+ installed (`go version`).

## Automated validation

```sh
go test ./internal/example/...
```

Expected: all `TestSkillsContent_*` tests pass, including:
- `TestSkillsContent_AllNineInstalled` — confirms exactly the 9 new names exist under `kit.SkillsFS`, and a real end-to-end install via the Claude Code adapter lands each one at `.claude/skills/mister-.../SKILL.md`, byte-identical to `kit.SkillsFS`'s own content.
- Every structural-conformance test (`skillOperationsAllowlist`-driven) resolving correctly under the new names.
- The three feature-specific tests from specs 026/027 (`TestSkillsContent_ImplementDualInvocation`, `TestSkillsContent_ConstitutionFrontmatterDocumented`, `TestSkillsContent_CreateTasksReportsDependenciesAndParallelism`) reading `mister-implement/SKILL.md`, `mister-constitution/SKILL.md`, and `mister-tasks/SKILL.md` respectively.

## Full-repo search verification (no old name remains in active content)

```sh
grep -rn '/analyze\b\|/create-constitution\b\|/create-feature\b\|/create-knowledge-base\b\|/create-plan\b\|/create-program\b\|/create-specs\b\|/create-tasks\b\|/implement\b' kit/skills/
```

Expected: zero matches. (Deliberately scoped to `kit/skills/` only — historical Specs under `specs/` and this repo's own separate `.claude/skills/speckit-*` tooling are out of scope and expected to still contain old-style references, per `research.md`'s own decision.)

## Manual validation (end-to-end scenario)

1. In a scratch Git-tracked project, run `misterspec init --dir <scratch> --agent claude-code` (non-interactive path) using a binary built from this branch.
2. List the installed commands (`ls <scratch>/.claude/skills/`) and confirm exactly the 9 new names from `data-model.md`'s mapping table — no old names present.
3. Open `mister-tasks/SKILL.md`'s own `## Recommended Next Step` and confirm it names `/mister-implement`, not `/implement`.
4. Open `mister-plan/SKILL.md` (or any other Skill known from the investigation to mention `/implement` in ordinary prose, not just its own next-step block) and confirm that prose mention is also updated.
5. Attempt to reference an old command name (e.g. `/create-tasks`) in the same project — confirm it is not a recognized installed command.

## Expected outcome

Every acceptance scenario in `spec.md` (User Stories 1–3) passes as described, and `go test ./...` passes with no regressions anywhere in the repository.
