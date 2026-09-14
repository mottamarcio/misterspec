# Quickstart: Interactive Init TUI

This is what a human at a terminal actually sees — `misterspec init`,
no flags, for the first time in this project's history.

## 1. Happy path — fresh, empty directory

```text
$ misterspec init

  Inspecting ./my-project...

  ✓ Not yet a misterspec project — safe to bootstrap.

  Select a coding agent:
  > Claude Code

  [Enter to select, Esc to cancel]
```

```text
  This will create:

    .misterspec/config.yaml
    ai/programs/, ai/knowledge/, ai/memory/  (kit templates: 8 files)
    .claude/skills/  (9 Skills for Claude Code)

  Proceed? [y/N]
```

```text
  ✓ Project bootstrapped.

    Configuration written
    8 kit resources installed
    9 Skills installed for claude-code

  Next: place a document under ai/raw/ and run /create-knowledge-base
  inside Claude Code.
```

## 2. Already-initialized target

```text
$ misterspec init

  Inspecting ./my-project...

  ⚠ Already a misterspec project (agent: claude-code).

  [C]ontinue anyway    [Esc] Cancel
```

Continuing past this warning still reaches the same rejection
`misterspec init --agent ...` (008-cli-cobra) already guarantees —
shown as a clear Error screen, never a silent no-op or an overwrite.

## 3. Non-empty, but not a misterspec project

```text
$ misterspec init

  Inspecting ./existing-repo...

  ⚠ Directory is not empty (not a misterspec project).

  [C]ontinue anyway    [Esc] Cancel
```

Continuing here *can* succeed — misterspec only ever writes its own
`.misterspec/`, kit, and agent-Skills subtrees, never touching
unrelated existing files.

## 4. Declining at the preview

```text
  Proceed? [y/N]
  > n

  Nothing was written.
```

## 5. No `--agent`, no terminal (e.g. piped, CI)

```text
$ misterspec init < /dev/null
{"ok":false,"error":{"code":"invalid_argument","message":"--agent is required outside an interactive terminal"}}
$ echo $?
2
```

The exact same shape a missing `--agent` already produced before this
feature existed (research.md) — machine callers keep getting one
consistent JSON contract.

## 6. The scriptable path is completely unchanged

```text
$ misterspec init --agent claude-code --dir ./ci-project
{"ok":true,"bootstrap":{...}}
```

No interactive screens appear — `--agent` still means "I already know
what I want" (008-cli-cobra, byte-for-byte unchanged).

## Validation

Validated by: `internal/tui`'s own `Model.Update`/`View` unit tests
(synthetic key sequences through every screen transition in
data-model.md's table, no real terminal needed — research.md); a
`bootstrap.Inspect`/`Empty` regression check against 007's existing
suite; and 008-cli-cobra's own `TestInitCmd_*` suite re-run unmodified,
confirming the non-interactive path is untouched.
