# Quickstart: Validate Constitution Frontmatter Guarantee, and Task Dependency/Parallelism Reporting

## Prerequisites

- A local checkout of this repo on branch `027-constitution-frontmatter-task-deps` (or later, once merged).
- Go 1.23.4+ installed (`go version`).

## Automated validation

```sh
go test ./internal/artifacts/...
go test ./internal/validation/...
go test ./internal/example/... -run TestSkillsContent
```

Expected: all pass, including:
- New `SchemaVersion` field parsing tests in `internal/artifacts`.
- New `checkConstitution` tests in `internal/validation`: absent file → no finding; malformed frontmatter → `CodeFrontmatterMalformed`; missing `schema_version` → `CodeRequiredFieldMissing`; valid frontmatter → no finding. Existing `004-structural-validation` tests for Program/Feature/Spec/Knowledge/Learning/Task unchanged and still passing.
- Extended `skills_content_test.go` assertions covering both Skill-content additions.

## Manual validation (end-to-end scenario)

1. In a scratch project with an existing Knowledge base and no Constitution yet, run `/create-constitution`. Read `ai/memory/constitution.md` and confirm it opens with `---\ntype: constitution\nschema_version: 1\n---`.
2. Run `misterspec internal validate` against that same project. Confirm no Constitution-related Finding appears (frontmatter is correct).
3. Manually edit `ai/memory/constitution.md` to remove its frontmatter block entirely (or just the `schema_version` line). Run `misterspec internal validate` again and confirm a Finding now appears with `path: "ai/memory/constitution.md"` and the expected code (`frontmatter_malformed` or `required_field_missing`).
4. Run `/create-constitution` again against that broken file. Confirm the frontmatter is repaired (added back) without disturbing the existing body sections.
5. Generate Tasks for a Spec whose Plan implies at least one dependency chain and one independent pair, via `/create-tasks`. Confirm the completion summary explicitly names the dependency (e.g. "Task B depends on Task A") and the parallel-safe group, distinctly from each other.
6. Re-run `/create-tasks` on the same Spec after it already has Tasks from a prior run (extending, not replacing). Confirm the dependency/parallel reporting reflects the full current `tasks.md`, not only the newly added Tasks.

## Expected outcome

Every acceptance scenario in `spec.md` (User Stories 1–2) passes as described, `004-structural-validation`'s own existing behavior is unaffected, and `go test ./...` passes with no regressions.
