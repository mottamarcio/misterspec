# Contract: Modified/New Skill Behavior

Reconciled against the actual implementation during `/speckit-implement`
(matching every prior feature's own contract convention). This project's
"contracts" for a dev-tooling feature are behavioral guarantees a coding
agent following these skills must satisfy — not an HTTP/CLI schema.

## Hook invocation (all 9 modified skills — User Story 1)

- A mandatory (`optional: false`) hook's own command MUST be invoked as a
  real step of the session — verifiable by the hook's own observable
  effect existing afterward (e.g. `git branch --show-current` shows the
  branch a `speckit.git.feature` hook was supposed to create).
- An unparseable `.specify/extensions.yml` MUST produce a visible message
  naming the parse error and stating that hook checking was skipped —
  never a silent continuation.

## `speckit-implement` (User Story 2)

- Given a `checklists/*.md` file with mixed checked/unchecked items before
  a run, its content (including every checkbox marker) MUST be
  byte-for-byte identical after the run.

## `speckit-checklist` (User Story 2)

- Every newly generated checklist item MUST be written as `- [ ] ...`,
  never `- [x] ...`, regardless of how confident the assessment is.

## `speckit-taskstoissues` (User Story 3)

- Given a `tasks.md` where every task already has a matching open or
  closed GitHub issue (title matches `\bT\d{3,}\b` for that task's ID),
  re-running the command MUST create zero new issues.
- Given a `tasks.md` with N already-issued tasks and M new tasks, a
  re-run MUST create exactly M new issues (one per new task) and zero for
  the N already-issued ones.

## `speckit-constitution` (User Story 4)

- Given a prompt containing both a governance change and an unrelated
  implementation/code/deploy request, the response MUST apply the
  governance change, MUST NOT perform the unrelated request, and MUST
  list it under a `Next Actions` section naming the appropriate follow-up
  command (without invoking it).
- Given a prompt containing only a governance change, behavior MUST be
  unchanged from before this feature — no `Next Actions` section is
  added when there is nothing to defer.

## `speckit-specify`/`speckit-plan`/`speckit-tasks`/`speckit-implement`/`speckit-clarify` (User Story 5)

- Each command's completion report MUST include an explicit `Done When`
  self-check listing that command's own required outputs, each marked
  against whether it was actually satisfied in this run.

## `speckit-converge` (new — User Story 6)

**Preconditions**: `spec.md`, `plan.md`, and `tasks.md` all exist for the
target feature — `plan.md`/`tasks.md` via `check-prerequisites.sh
--require-tasks --include-tasks`, `spec.md` via an explicit extra check
(the local script has no `--require-spec` flag); missing any of the
three stops the command with a message naming the specific prerequisite
command to run instead (`speckit-specify`, `speckit-plan`, or
`speckit-tasks`).

**Converged outcome** (no gap found):

```text
✅ Converged — the implementation satisfies the spec, plan, and tasks.
```

- `tasks.md` MUST be byte-for-byte unchanged (no empty `## Phase N:
  Convergence` header is ever written).

**Findings outcome** (one or more gaps found): a Markdown table is
presented in-session before any write:

```markdown
## Convergence Findings

| ID | Gap Type | Severity | Source | Evidence | Remaining Work |
|----|----------|----------|--------|----------|----------------|
| F1 | missing  | HIGH     | FR-008 | ...      | ...            |
```

then a single new section is appended to the end of `tasks.md`:

```markdown
## Phase N: Convergence

- [ ] T042 <imperative description> per <source-ref> (<gap-type>)
```

- `N` MUST be the highest existing phase number + 1.
- Task IDs MUST continue from the highest existing task ID + 1 —
  never reused, never renumbered.
- A Constitution-violation finding MUST be `CRITICAL` and MUST appear
  first among the appended tasks.
- `spec.md`, `plan.md`, and every task that existed before this run MUST
  be unchanged — the appended section is the command's only write.
- A second `speckit-converge` run against the same, still-unimplemented
  findings MUST NOT re-append the same findings as new tasks (its own
  intent-inventory step re-reads `tasks.md`'s own existing tasks, so
  already-tracked remaining work is recognized as already-tracked, not
  duplicated).

## `speckit-clarify` (User Story 7)

- Every presented clarification question MUST be a complete interrogative
  sentence ending in `?`, MUST NOT be a bare requirement ID or topic
  label, and MUST be immediately preceded or followed by a one-sentence
  "why it matters" statement.
- After a clarification round completes, if `checklists/requirements.md`
  exists, it MUST be re-validated against the updated spec; only checkbox
  markers whose pass/fail state actually changed MUST be toggled; the
  completion report MUST list newly-passing items and any regressions.

## `speckit-plan`/`speckit-tasks` (User Story 8)

- A generated `quickstart.md` MUST NOT contain full implementation code,
  complete model/service/controller bodies, database migrations, or a
  complete test suite.
- A generated task touching a field with a documented constraint in
  `data-model.md` MUST quote that constraint verbatim in the task's own
  description.
