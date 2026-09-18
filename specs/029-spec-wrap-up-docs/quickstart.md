# Quickstart: Validate `/mister-wrap-up`

## Prerequisites

- A local checkout of this repo on branch `029-spec-wrap-up-docs` (or later, once merged).
- Go 1.23.4+ installed (`go version`).

## Automated validation

```sh
go test ./internal/vcs/...
go test ./internal/cli/internalcmd/...
go test ./internal/example/...
```

Expected: all pass, including:
- New `internal/vcs` tests for `CommitsSinceFileAdded`: file never committed → `available: false`; file added at the repository's own root commit (no parent) → still returns the full range correctly; ordinary case with commits before and after the file was added → range starts exactly at the "added" commit; not a Git repository at all → `available: false`, no error.
- New `internal commits-since-file` CLI JSON-contract tests.
- `skills_content_test.go`'s extended assertions: `mister-wrap-up` conforms to the same §39 structural contract as every other canonical Skill; the renamed "all ten canonical Skills installed" test passes with `mister-wrap-up` included.

## Manual validation (end-to-end scenario)

1. In a scratch Git-tracked project initialized with `misterspec init`, create a small Program → Feature → Spec → Plan → Tasks chain (a couple of Tasks is enough), committing normally at each step with no special message convention.
2. Implement and complete the Tasks via `/mister-implement`, committing as usual.
3. Run `/mister-wrap-up SPEC-0XX`. Confirm:
   - `cortex/SPEC-0XX-<slug>.md` now exists.
   - It reads as synthesized prose — purpose, what was built, outcome — not a raw dump of `tasks.md` or `git log` output.
   - It correctly reflects every commit made since `plan.md` was first added, even though no commit message mentioned the Spec ID.
3. Make one more commit, then run `/mister-wrap-up SPEC-0XX` again. Confirm the document is replaced (not duplicated) and reflects the new commit.
4. Run `/mister-wrap-up` against a Spec with Tasks still outstanding. Confirm the document states plainly that implementation is incomplete and names what remains.
5. Run `/mister-wrap-up` against a Spec whose `plan.md` was never committed (or in a non-Git-tracked copy of the project). Confirm a document still generates, with a clear note that commit history could not be included.

## Expected outcome

Every acceptance scenario in `spec.md` (User Stories 1–3) passes as described, and `go test ./...` passes with no regressions.
