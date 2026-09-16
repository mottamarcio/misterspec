# Quickstart: Validate Dual-Mode Implement

## Prerequisites

- A local checkout of this repo on branch `026-implement-single-task` (or later, once merged).
- Go 1.23.4+ installed (`go version`).

## Automated validation

Run the existing Skill-content conformance suite, which will include the new Phase 2 assertion covering both invocation forms:

```sh
go test ./internal/example/... -run TestSkillsContent
```

Expected: all `TestSkillsContent_*` tests pass, including the new assertion that `kit/skills/implement/SKILL.md`'s `Invocation` section documents both `SPEC-###` and `SPEC-### TASK-NNN`.

Also run the full example suite to catch any cross-agent rendering regression:

```sh
go test ./internal/example/...
```

## Manual validation (end-to-end scenario)

1. In a scratch project initialized with `misterspec init` (or any project already using the kit), create a Spec with a Plan and at least three Tasks, e.g. via `/create-specs`, `/create-plan`, `/create-tasks` for a small dummy feature.
2. **All-tasks mode**: run `/implement SPEC-0XX` (the Spec ID only). Confirm:
   - Every currently executable Task is implemented, verified, and marked complete in the one invocation, without stopping to ask you to re-run the command between tasks.
   - The completion summary lists every Task it implemented (and any it skipped as already complete).
3. Reset (or create a fresh) Spec with unimplemented Tasks. **Named-task mode**: run `/implement SPEC-0XX TASK-002` naming one specific Task ID. Confirm:
   - Only `TASK-002` is implemented and marked complete; every other Task's checkbox is untouched.
   - The completion summary names the next eligible Task ID(s) and mentions that `/implement SPEC-0XX` alone will run all remaining Tasks.
4. **Error handling**: run `/implement SPEC-0XX TASK-999` (a nonexistent Task ID). Confirm the system reports the Task wasn't found in that Spec, lists valid Task IDs, and makes no file changes.
5. **Dependency gate**: name a Task whose declared dependency isn't complete yet. Confirm the system stops before implementing it and names the blocking dependency, in both the named-task and all-tasks forms.

## Expected outcome

Both invocation forms behave exactly as documented in `contracts/implement-invocation.md`, and `go test ./internal/example/...` passes with no regressions.
