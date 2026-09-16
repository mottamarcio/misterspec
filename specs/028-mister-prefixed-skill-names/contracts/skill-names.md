# Contract: Canonical Skill Slash-Command Names

This documents the user-facing slash-command contract for misterspec's 9 canonical Skills, before and after this feature. Not an API in the code sense — the "contract" is: directory name under `kit/skills/` = installed slash-command name, for every supported agent integration, with no exceptions (confirmed by investigation: no adapter transforms it).

## Before this feature

```text
/analyze
/create-constitution
/create-feature
/create-knowledge-base
/create-plan
/create-program
/create-specs
/create-tasks
/implement
```

Short, generic verbs/nouns — collide with any other agent framework that happens to register a command of the same bare name in the same environment.

## After this feature

```text
/mister-analyze
/mister-constitution
/mister-features
/mister-knowledge-base
/mister-plan
/mister-program
/mister-specify
/mister-tasks
/mister-implement
```

Every one begins with `mister-`, uniquely identifying it as belonging to misterspec regardless of what other frameworks are installed in the same environment.

## Invariant

- No old name continues to resolve to anything after this feature — no alias, no redirect, no dual-registration (spec.md FR-004). An agent tool asked for `/implement` after this change behaves exactly as it would for any command that was never installed.
- The mapping in `data-model.md`'s table is exhaustive and fixed — this contract does not introduce any 10th name or drop any of the 9.
