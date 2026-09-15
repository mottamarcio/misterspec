# Quickstart: Spec-Kit Tooling Improvements

Every scenario below is manual verification against this repository's own
`.specify/` tooling — there is no unit-test framework for instruction text.
Run each on a disposable feature branch/spec, never against a real feature
in progress.

## 1. Hook is actually invoked, not just described

```sh
git checkout dev
/speckit-specify "throwaway test feature for hook verification"
git branch --show-current
# should print a new branch — NOT "dev" — confirming the mandatory
# before_specify hook (speckit.git.feature) actually ran, not just
# printed its own description
```

Then corrupt `.specify/extensions.yml` (e.g. an unbalanced quote), run any
`/speckit-*` command, and confirm the response explicitly names the parse
error rather than silently proceeding.

## 2. Checklist stays untouched during implementation

```sh
# Pick a feature with an existing checklists/requirements.md that has
# both checked and unchecked items.
md5sum specs/<feature>/checklists/requirements.md
/speckit-implement
md5sum specs/<feature>/checklists/requirements.md
# both hashes must be identical
```

## 3. Re-running taskstoissues creates no duplicates

```sh
/speckit-taskstoissues
# note the issue numbers created
/speckit-taskstoissues
# re-run with no changes to tasks.md — must create zero new issues,
# and report "<ID> already has an issue, skipping" for each
```

## 4. Constitution command defers unrelated work

```sh
/speckit-constitution "Add a new principle about X, and also please implement the login page while you're at it"
# must update the constitution with the new principle, must NOT touch
# any application code, and must list "implement the login page" under
# a Next Actions section recommending /speckit-specify
```

## 5. Done When gate appears in completion reports

```sh
/speckit-plan
# completion report must include a "Done When" checklist naming this
# command's own required outputs, each confirmed
```

## 6. Convergence reports gaps without rewriting existing artifacts

```sh
# Against a feature that is deliberately missing one piece of its own
# spec (e.g. comment out an implementation detail satisfying an FR):
md5sum specs/<feature>/spec.md specs/<feature>/plan.md
/speckit-converge
md5sum specs/<feature>/spec.md specs/<feature>/plan.md
# both must be unchanged; tasks.md must have exactly one new
# "## Phase N: Convergence" section appended, naming the missing FR

# Re-run with nothing missing:
md5sum specs/<feature>/tasks.md
/speckit-converge
md5sum specs/<feature>/tasks.md
# must be unchanged; report must say "✅ Converged"
```

## 7. Clarify questions are real questions, checklist re-validates

```sh
/speckit-clarify
# every question presented must read as a full sentence ending in "?",
# never a bare requirement ID or label; each must be preceded by a
# one-sentence "why it matters"
# after answering: completion report must show the spec quality
# checklist's before/after pass counts
```

## 8. Plan/tasks stay appropriately scoped

```sh
/speckit-plan
grep -c '```' specs/<feature>/quickstart.md
# quickstart.md must contain no full implementation code blocks —
# only short example invocations/verification commands

/speckit-tasks
grep -A2 '<field-with-a-constraint>' specs/<feature>/tasks.md
# the task touching that field must quote its data-model.md constraint
# verbatim, not paraphrase it
```
