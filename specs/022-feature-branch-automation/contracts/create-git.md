# Contract: `internal create`'s extended JSON payload

Reconciled against the actual implementation (022-feature-branch-automation)
— every shape below was verified against a real `misterspec internal create`
invocation, not just unit tests.

## `misterspec internal create feature --parent PRG-001` (Git-tracked project, automation enabled, branch newly created, no slug)

```json
{
  "ok": true,
  "created": {
    "id": "FEAT-007",
    "type": "feature",
    "path": "ai/programs/PRG-001/features/FEAT-007",
    "git": {
      "branch": "feat/FEAT-007",
      "created": true
    }
  }
}
```

## `misterspec internal create feature --parent PRG-001 --slug "User Auth"` (same, with a readable slug)

```json
{
  "ok": true,
  "created": {
    "id": "FEAT-007",
    "type": "feature",
    "path": "ai/programs/PRG-001/features/FEAT-007",
    "git": {
      "branch": "feat/FEAT-007-user-auth",
      "created": true
    }
  }
}
```

## `misterspec internal create feature ...` (branch already existed — resumed)

```json
{
  "ok": true,
  "created": {
    "...": "...",
    "git": {
      "branch": "feat/FEAT-007",
      "created": false
    }
  }
}
```

## `misterspec internal create feature ...` (not a Git repository)

```json
{
  "ok": true,
  "created": {
    "...": "...",
    "git": {
      "skipped_reason": "not_a_git_repo"
    }
  }
}
```

## `misterspec internal create feature ...` (`git_branch_automation: false` in config)

```json
{
  "ok": true,
  "created": {
    "...": "...",
    "git": {
      "skipped_reason": "disabled"
    }
  }
}
```

## `misterspec internal create spec --parent FEAT-007 ...` (current branch matches `feat/FEAT-007`)

```json
{
  "ok": true,
  "created": {
    "id": "SPEC-014",
    "type": "spec",
    "path": "ai/programs/PRG-001/features/FEAT-007/specs/SPEC-014",
    "git": {
      "branch": "feat/FEAT-007"
    }
  }
}
```

## `misterspec internal create spec --parent FEAT-007 ...` (current branch is `dev`, not `feat/FEAT-007`)

```json
{
  "ok": true,
  "created": {
    "...": "...",
    "git": {
      "branch": "feat/FEAT-007",
      "warning": "current branch \"dev\" does not match parent Feature FEAT-007's own branch \"feat/FEAT-007\""
    }
  }
}
```

## Field reference

| Field | Type | Meaning |
|---|---|---|
| `created.git.branch` | string | The relevant branch name (the one just ensured, for Feature; the parent Feature's own branch, for Spec). Omitted when `skipped_reason` is present. |
| `created.git.created` | bool | Present only for `Type == feature`: `true` if the branch was newly created, `false` if an existing branch of that name was resumed. |
| `created.git.skipped_reason` | string | `"not_a_git_repo"` or `"disabled"` — present only when Git automation did not run at all. Mutually exclusive with `branch`/`created`/`warning`. |
| `created.git.warning` | string | Present only for `Type == spec` when the current branch does not match the parent Feature's own branch. Never blocks the creation. |

Entity types other than `feature`/`spec` (`program`, `task`, `knowledge`,
`learning`) never gain a `git` object at all — unchanged from today's
payload shape.

## Behavioral guarantees carried over (not re-tested here, only extended)

- `operations.Create` still refuses an invalid/ambiguous parent and an
  already-existing target path before writing anything (existing FR-004/
  FR-008/FR-011/FR-012, unchanged).
- Git branch operations are additive metadata on an already-successful
  `create` call — they never turn a successful creation into a failure
  (FR-003, FR-005).
