# Quickstart: Validating the canonical task identity model

This guide proves the feature works end-to-end once implemented. It exercises the three user stories from `spec.md` against a disposable temp project — it does not modify the real repository.

## Prerequisites

- A built `misterspec` binary (`go build ./...` from repo root) reflecting this feature's changes.
- A scratch directory outside the repo to hold the temp project (or use `t.TempDir()` inside the Go integration tests that cover this same scenario — this guide mirrors what those tests assert).

## Setup: two Specs sharing a Task number

```sh
mkdir -p /tmp/ms-quickstart && cd /tmp/ms-quickstart
misterspec init --here   # or the project's normal bootstrap path

# Create two Specs (existing `create` operation — outside this feature's scope)
# ... SPEC-001 and SPEC-002 now exist, each with its own tasks.md

# Author, by hand, a "## TASK-001" heading in each Spec's tasks.md — this is
# the exact scenario spec.md's User Story 1/2 describe: two Specs, each
# legitimately numbering their first task TASK-001.
```

## Scenario 1 — Composite resolution is unambiguous (User Story 1)

```sh
misterspec internal inspect SPEC-001:TASK-001
# Expected: JSON result naming SPEC-001's TASK-001 only.

misterspec internal inspect SPEC-002:TASK-001
# Expected: JSON result naming SPEC-002's TASK-001 only — a different
# location than the previous command, even though the local number matches.
```

**Pass condition**: the two commands above return two different `Path` values, each scoped to its own Spec. Neither command's output ever mentions the other Spec's task.

## Scenario 2 — Bare reference without context, when ambiguous (User Story 1, spec FR-003)

```sh
misterspec internal inspect TASK-001
# Expected: a non-zero exit, a stable JSON error naming both SPEC-001 and
# SPEC-002 as candidates, and an explicit ask for the composite form —
# never a silent pick of one of the two, never a duplicate-ID error.
```

**Pass condition**: the error output is structured (parseable JSON), names both candidate Specs, and the process does not exit `0` claiming success against one arbitrarily chosen match.

## Scenario 3 — Validation does not false-positive across Specs (User Story 2)

```sh
misterspec internal validate
# Expected: no CodeDuplicateID finding for TASK-001, since SPEC-001 and
# SPEC-002 each legitimately own their own TASK-001.
```

**Pass condition**: the JSON findings list contains zero `CodeDuplicateID` entries referencing Task number 1.

## Scenario 4 — Validation still catches a real duplicate (User Story 2)

```sh
# Hand-edit SPEC-001's tasks.md to add a second "## TASK-001 — ..." heading.
misterspec internal validate
# Expected: exactly one CodeDuplicateID finding, scoped to SPEC-001,
# naming both heading locations within that one file.
```

**Pass condition**: the finding's message/path identifies SPEC-001 specifically (not "somewhere in the project") and lists both colliding locations inside that one `tasks.md`.

## Scenario 5 — Migration diagnostic is read-only (User Story 3, spec FR-007/FR-008)

```sh
misterspec internal migration-check tasks   # exact subcommand name TBD in tasks.md
# Expected: JSON listing the TASK-001 collision between SPEC-001 and
# SPEC-002 as an informational entry (not an error), with ok:true.

git status --short
# Expected: no changes — the diagnostic must not have modified any
# tasks.md file or renumbered anything.
```

**Pass condition**: the diagnostic lists the SPEC-001/SPEC-002 collision, exits successfully, and `git status` shows zero modified files.

## Scenario 6 — Single-Spec validation agrees with project-wide validation (spec FR-011)

```sh
misterspec internal validate SPEC-001
# Expected: the same CodeDuplicateID finding from Scenario 4 (if still
# present), and no finding at all if Scenario 4's edit was reverted first.
```

**Pass condition**: running `validate SPEC-001` alone and `validate` (whole project) filtered to SPEC-001's findings produce an identical set of Task-related findings.

## Cleanup

```sh
rm -rf /tmp/ms-quickstart
```
