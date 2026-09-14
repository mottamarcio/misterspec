# Quickstart: CLI Command Layer (Cobra)

Unlike 001-007's quickstarts, this one is real terminal usage — the
first feature in this project a human or script actually runs, not only
a future Go caller.

## 1. The public surface stays small

```bash
$ misterspec --help
Usage:
  misterspec [command]

Available Commands:
  init        Bootstrap a new misterspec project
  help        Help about any command
  completion  Generate the autocompletion script for the specified shell

$ misterspec internal --help
Error: unknown command "internal" for "misterspec"
```

(The `internal` tree is `Hidden: true` — Cobra refuses to show it in
help, but it still runs when invoked directly, next.)

## 2. Bootstrap a new project, non-interactively

```bash
$ misterspec init --agent claude-code --dir ./my-project
{"ok":true,"bootstrap":{"project_root":"./my-project","config_written":true,"templates":[...],"agent":{"adapter_id":"claude-code","integration_path":".claude/skills","outcomes":[{"resource":{"name":"README.md",...},"status":0,"path":"README.md"}]}}}
$ echo $?
0
```

(`outcomes` has exactly one entry — `kit/skills/README.md`, a
placeholder explaining canonical Skill content is Phase 6 future work —
until that phase replaces it with real Skills; research.md. The project
itself is still fully created and detectable either way.)

## 3. Rejections are structured, not crashes

```bash
$ misterspec init --agent claude-code --dir ./my-project
{"ok":false,"error":{"code":"already_initialized","message":"..."}}
$ echo $?
5

$ misterspec init --agent no-such-agent --dir ./fresh
{"ok":false,"error":{"code":"unknown_agent","message":"..."}}
$ echo $?
5
```

## 4. Invoke a deterministic operation directly

```bash
$ misterspec internal resolve SPEC-014 --dir ./my-project
{"ok":true,"entity":{"id":"SPEC-014","type":"spec","path":"ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md"}}

$ misterspec internal resolve SPEC-999 --dir ./my-project
{"ok":false,"error":{"code":"entity_not_found","message":"..."}}
$ echo $?
3
```

## 5. `create` allocates an ID and scaffolds the artifact

```bash
$ misterspec internal create feature --parent PRG-001 --dir ./my-project
{"ok":true,"created":{"id":"FEAT-002","type":"feature","path":"ai/programs/PRG-001/features/FEAT-002/feature.md"}}
```

## 6. `validate`'s `ok`/`valid` stay distinct — but the exit code still tells a script what happened

```bash
$ misterspec internal validate --dir ./my-project
{"ok":true,"valid":true,"findings":[]}
$ echo $?
0

$ misterspec internal validate --dir ./broken-project
{"ok":true,"valid":false,"findings":[{"code":"missing_parent","severity":"error","path":"...","message":"..."}]}
$ echo $?
4
```

## 7. `status` for a Skill deciding what to do next

```bash
$ misterspec internal status --dir ./my-project
{"ok":true,"counts":{"program":1,"feature":2,"spec":3,"task":0,"knowledge":0,"learning":0},"specs":{"draft":1,"ready":2},"structural_errors":0}
```

## Validation

Validated by this feature's Go-level tests (calling `internalcmd`'s
`RunE` functions and Cobra's own `Execute` in-process, capturing
stdout/exit code without a real subprocess — fast) plus a small set of
true subprocess integration tests building the real `misterspec` binary
and running it against a fixture project directory, confirming the
exact JSON and exit codes shown above end to end.
