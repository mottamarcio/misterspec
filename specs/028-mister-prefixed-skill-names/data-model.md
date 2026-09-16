# Phase 1 Data Model: `mister-`-Prefixed Skill Names

No new persisted entity, storage format, or Go type is introduced. Documented here for the one renaming concept involved.

## Skill Rename Mapping (fixed, exhaustive — the entire mapping this feature applies)

| Old Name | New Name |
|---|---|
| `analyze` | `mister-analyze` |
| `create-constitution` | `mister-constitution` |
| `create-feature` | `mister-features` |
| `create-knowledge-base` | `mister-knowledge-base` |
| `create-plan` | `mister-plan` |
| `create-program` | `mister-program` |
| `create-specs` | `mister-specify` |
| `create-tasks` | `mister-tasks` |
| `implement` | `mister-implement` |

This table is the single source of truth for every rename in this feature — the directory name, the Skill's own frontmatter `name:` field, its own `## Invocation` slash command, and every other Skill's cross-reference to it, all derive from this same mapping.

## Cross-Reference (not a stored entity — a content relationship)

One Skill's own file mentioning another Skill's name, in one of these forms:

| Form | Example (before) | Where it appears |
|---|---|---|
| Next-step guidance | `` `/implement SPEC-###` `` | `## Recommended Next Step` |
| Stated dependency | `` `/create-plan` — the Skill this one depends on directly. `` | `## Related Skills` |
| Failure-condition recommendation | `recommend revisiting the Plan (`/create-plan`)` | `## Failure Conditions` |
| Ordinary prose | `` `/implement` writes code. `` | Anywhere in body text |

Every one of the 9 renamed Skills contains at least one cross-reference to at least one other (confirmed: a full mesh, not a chain, per the pre-plan investigation) — every occurrence of an old name in any of these four forms, in any of the 9 files, must become the new name.

## Test-File Literal (the one Go-level artifact affected)

`internal/example/skills_content_test.go` holds three distinct kinds of hardcoded name literals, all derived from the same mapping table above:

| Structure | Shape | Count |
|---|---|---|
| `skillOperationsAllowlist` | `map[string][]string` keyed by Skill name | 9 keys |
| `canonicalSkillNames` | `[]string` | 9 elements |
| Feature-specific literal path reads | `fs.ReadFile(kit.SkillsFS, "<name>/SKILL.md")` plus matching error-message string literals | 3 tests (`TestSkillsContent_ImplementDualInvocation`, `TestSkillsContent_ConstitutionFrontmatterDocumented`, `TestSkillsContent_CreateTasksReportsDependenciesAndParallelism`) |
