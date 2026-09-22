# Quickstart: Validating Task Evidence & Fingerprint Validity

Prerequisites: a built `misterspec` binary from this branch, and an
initialized project (`misterspec init`) with a Spec that already has a
`tasks.md` containing at least one Task section.

## 1. A checked box alone is never "complete" (User Story 1)

Check a Task's own box (`- [x]`) by hand, with no `Evidence-*:` lines:

```sh
misterspec internal prepare SPEC-014
```

**Expected**: that Task is not treated as complete — it's still
eligible for (re-)selection, and any Task depending on it is not
reported ready. Confirm via validation too:

```sh
misterspec internal validate SPEC-014
```

**Expected**: a Finding with `"code": "unverified_task"` naming that
Task.

## 2. Capture real evidence, then the Task is genuinely complete

```sh
misterspec internal capture-evidence SPEC-014 --task TASK-003 \
  --origin automated --by "go test" \
  --command go --args test --args ./internal/foo/...
```

**Expected**: `ok: true`; the response's `evidence.result` reflects the
command's real exit code; `evidence.log` names a new file under
`.../SPEC-014/evidence/`. Write the returned fields into the Task's own
body as `Evidence-*:` lines (the step `/mister-implement` performs),
then:

```sh
misterspec internal prepare SPEC-014
misterspec internal validate SPEC-014
```

**Expected**: if `evidence.result` was `"pass"`, the Task is now
reported complete and `validate` raises no evidence-related Finding for
it; if it was `"fail"`, `validate` raises `"code":
"failed_task_evidence"` instead, and the Task is still not complete.

## 3. Dependency readiness honors evidence, not the checkbox (User Story 2)

With an upstream Task checked but only carrying `unverified_task`
(step 1's state):

```sh
misterspec internal prepare SPEC-014 --task TASK-004
```

**Expected**: `ready: false`, naming the upstream Task's own unverified
state as the blocker — never treated as satisfied just because its box
is checked. Capture valid evidence for the upstream Task (step 2) and
re-run:

**Expected**: `ready: true`.

## 4. Editing the verified content makes evidence stale (User Story 3)

With TASK-003 fully `Verified` from step 2, edit that Task's own
`Scope:` line (or any other text in its body) and re-run:

```sh
misterspec internal validate SPEC-014
```

**Expected**: a Finding with `"code": "stale_task_evidence"` for
TASK-003 — no manual staleness check required. Capture new evidence
against the edited content (step 2 again) and confirm the Finding
disappears and the Task is `Verified` again.

## 5. Automated execution is explicit and records real Git state (User Story 4)

```sh
echo "some local change" >> README.md
misterspec internal capture-evidence SPEC-014 --task TASK-003 \
  --origin automated --by "go test" --command go --args test --args ./...
```

**Expected**: `evidence.working_tree` is `"dirty"`, not just a bare
commit SHA — the uncommitted `README.md` change is reflected. Revert
it and commit whatever is currently pending — including any earlier
capture's own new `evidence/*.log` file, itself an untracked change
until committed (found during implementation's own manual walkthrough:
re-running immediately after only reverting `README.md` still reports
`"dirty"`, correctly, because of that log file) — and re-run:

**Expected**: `evidence.working_tree` is `"clean"`.

Confirm the command actually run is exactly what was named — no
Markdown-embedded text is ever executed implicitly:

```sh
misterspec internal capture-evidence SPEC-014 --task TASK-003 \
  --origin automated --by "test" --command go --args test --exec-dir ../outside-project
```

**Expected**: `ok: false`, a stable error naming the `--exec-dir` path-
traversal rejection — no subprocess runs.
