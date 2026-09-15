**Reconciled against the actual implementation (T040)** — six things
discovered or refined during implementation, not present in the
original draft below:

1. **`WriteSuccess` takes `fields map[string]any`, not `key string,
   value any`.** The original draft assumed every command's success JSON
   has exactly one field beyond `"ok"`, but `validate` needs `"valid"` +
   `"findings"` and `status` needs `"counts"` + `"specs"` +
   `"structural_errors"` — both multiple sibling keys. A map keeps the
   envelope helper genuinely shared across all eleven commands rather
   than `validate`/`status` needing a second, parallel marshaling path.
2. **`WriteError` returns `error` (an `*ExitCodeError`), not `int`.**
   Every command's `RunE` can then simply `return
   internalcmd.WriteError(w, err)` on any failure — the exit code rides
   inside the returned error (extracted via `errors.As` in
   `cli.Execute`) instead of needing a second explicit wrapping step at
   every one of the eleven call sites.
3. **`classify` also matches `validation.ErrInvalidTarget`.**
   `internal/validation` defines its own `ErrInvalidTarget` sentinel,
   distinct from `operations.ErrInvalidTarget` (same name, different
   package, different `errors.New` value) — both are real return paths
   (`validate`'s `ValidateEntity` uses the former), and both classify to
   `"invalid_target"`, exit `2`.
4. **`kit.SkillsFS`'s placeholder is `kit/skills/README.md`, not a
   hidden `.gitkeep`.** `go:embed` cannot compile against a directory
   containing only hidden (`.`/`_`-prefixed) files unless the pattern is
   `all:`-prefixed — confirmed by direct experiment during
   implementation — and `all:` would instead materialize a stray hidden
   file on every bootstrap. A plainly named, self-documenting
   `README.md` is the one file present until Phase 6 authors real
   Skills; `kit.SkillsFS` itself is `fs.Sub`'d so its root matches what
   `Adapter.Install` already expects (research.md).
5. **`init`'s `--agent` requirement is enforced manually inside `RunE`,
   not via Cobra's `MarkFlagRequired`.** A `MarkFlagRequired` failure
   short-circuits before `RunE` ever runs, bypassing the JSON-envelope
   path entirely (Cobra would print its own plain-text error) — an
   explicit `if agent == "" { return
   internalcmd.WriteError(...ErrInvalidArgument...) }` check keeps every
   failure, including this one, going through the same envelope, both
   when `init` runs standalone in tests and through the full root tree.
6. **`installer.Outcome` and `agents.InstallResult` JSON field names**
   (not specified in the original draft): each entry under
   `bootstrap.templates`/`bootstrap.agent.outcomes` is `{"name",
   "kind", "status", "path"}` plus an optional `"error"` when
   `status == "failed"` — `status` renders `installer.OutcomeStatus` as
   `"installed"`/`"skipped"`/`"failed"`.

# Phase 1 Contracts: CLI Command Layer

Unlike 001-007, this feature's contract *is* an external interface — a
terminal command tree — so this document specifies observable CLI
behavior (commands, flags, stdout JSON, exit codes) rather than only Go
function signatures. The Go signatures beneath each command are also
given, since they are what tests call directly (fast, in-process) before
a slower subprocess-level check confirms the same behavior end to end
(plan.md's Testing strategy).

## Command tree

```text
misterspec                          # root command; --help shows only this and "init"
├── init                            # public (User Story 2)
└── internal                        # hidden from --help (User Story 3); FR-005
    ├── resolve <id>
    ├── inspect <id>
    ├── parent <id>
    ├── children <id> [--type T]
    ├── create <type> [--parent P] [--slug S]
    ├── create-artifact <kind> --for <specID>
    ├── fingerprint <path>
    ├── inventory <dir>
    ├── validate [<id>]
    └── status
```

Every command accepts an optional `--dir` flag (default `.`) naming the
target project's root or a directory within it — every internal command
resolves this to a project the same way `project.Detect` already does
(no separate "must be run from inside the project" constraint).

## `internal/cli` — root and public command

```go
package cli

// Execute is cmd/misterspec/main.go's entire body: build the root
// command tree and run it, returning the process's exit code.
func Execute() int

// newRootCmd builds the root *cobra.Command. --help shows only "init"
// (and Cobra's own built-in help/completion) — the "internal" command
// is registered but Hidden: true (FR-005, User Story 3).
func newRootCmd() *cobra.Command

// newInternalCmd builds the "internal" parent *cobra.Command
// (Hidden: true) and attaches every internalcmd.NewXxxCmd() constructor
// as a subcommand — the single place the ten operation commands are
// wired into the tree (User Story 1).
func newInternalCmd() *cobra.Command
```

```go
package cli

// newInitCmd builds the public "init" command: --agent (required,
// checked explicitly inside RunE rather than via Cobra's
// MarkFlagRequired — item 5 above), --dir (default "."). Calls
// bootstrap.Bootstrap(dir, agent, builtin.Default(), kit.SkillsFS)
// directly and reports the result through internalcmd's envelope helper
// (User Story 2, research.md).
func newInitCmd() *cobra.Command
```

**Guarantees**: `misterspec --help` and `misterspec` (no args) list only
`init` (plus Cobra's own `help`/`completion`/`--version`) — never any
`internal` subcommand name (FR-005, SC-003). `misterspec internal
<command>` still runs correctly when invoked directly by name — hidden
from discovery, not removed (Acceptance Scenario 2, User Story 3).

## `internal/cli/internalcmd` — shared envelope and error classification

```go
package internalcmd

// ExitCodeError carries the process exit code a failed command should
// terminate with; cli.Execute extracts it via errors.As (item 2 above).
type ExitCodeError struct{ Code int }

func (e *ExitCodeError) Error() string

// WriteSuccess marshals {"ok": true, <fields...>} to w as JSON, one
// object, newline-terminated — fields supplies every key beyond "ok"
// (item 1 above). Every command's RunE calls this exactly once on
// success (research.md).
func WriteSuccess(w io.Writer, fields map[string]any) error

// WriteError classifies err (via classify), marshals {"ok": false,
// "error": {"code", "message"}} to w, and returns a non-nil
// *ExitCodeError carrying the matching exit code (item 2 above). Every
// command's RunE calls this exactly once on failure, simply
// `return internalcmd.WriteError(w, err)`.
func WriteError(w io.Writer, err error) error

// classify maps a returned error to (code, exitCode) via errors.Is/As
// against every sentinel 001-007 export (data-model.md's Error Code
// table, including both operations.ErrInvalidTarget and
// validation.ErrInvalidTarget — item 3 above). Unrecognized errors
// classify as ("unexpected_failure", 1).
func classify(err error) (code string, exitCode int)
```

**Guarantees**: No command constructs its own JSON error object or
picks its own exit code inline (research.md) — `classify` is the single
place every sentinel-to-code mapping lives, so a new sentinel is wired
in once and every command that can produce it is correctly classified
automatically.

## `internal/cli/internalcmd` — the ten operation commands

Each is a `func New<Name>Cmd() *cobra.Command` returning a fully
configured Cobra command whose `RunE`: (1) resolves `--dir` to a
`project.Configuration` via `project.Detect` — a failure here is always
`project_not_initialized` before any operation-specific logic runs; (2)
calls exactly one existing operation/validation function; (3) shapes its
result per data-model.md and calls `WriteSuccess`, or calls `WriteError`
on failure.

```go
func NewResolveCmd() *cobra.Command       // operations.Resolve
func NewInspectCmd() *cobra.Command       // operations.Inspect
func NewParentCmd() *cobra.Command        // operations.Parent
func NewChildrenCmd() *cobra.Command      // operations.Children (+ --type)
func NewCreateCmd() *cobra.Command        // operations.Create (+ --parent, --slug)
func NewCreateArtifactCmd() *cobra.Command // operations.CreateArtifact (+ --for)
func NewFingerprintCmd() *cobra.Command   // operations.Fingerprint
func NewInventoryCmd() *cobra.Command     // operations.Inventory
func NewValidateCmd() *cobra.Command      // validation.ValidateProject / ValidateEntity
func NewStatusCmd() *cobra.Command        // operations.Status
```

**Guarantees** (FR-001, FR-002, FR-009): every one of these ten
functions calls exactly one existing 001-007 function — no command
contains business logic beyond argument/flag parsing and JSON shaping.
`validate` with no ID argument calls `ValidateProject`; with one, calls
`ValidateEntity` — never both, never a third implementation.

## Exit codes (§8, data-model.md's Error Code table)

| Code | Meaning |
|---|---|
| 0 | success |
| 1 | unexpected failure |
| 2 | invalid invocation (`invalid_target`, `unsupported_type`, `invalid_argument`) |
| 3 | target not found (`entity_not_found`, `entity_ambiguous`) |
| 4 | structural validation failure — `validate` reporting `valid: false` (research.md); note `ok` stays `true` in the JSON body even at this exit code, per §16's explicit "ok = executed successfully; valid = project structure is valid" distinction |
| 5 | mutation rejected (`invalid_parent`, `already_exists`, `path_outside_project`, `already_initialized`, `unknown_agent`) |
| 6 | project not initialized (`project_not_initialized`) |

`validate`'s exit code is the one case not driven by `classify` (which
only ever runs on an *error* return) — `NewValidateCmd`'s `RunE` checks
`len(findings) > 0` on its own successful return and exits `4` directly
(research.md).

## Cross-cutting: no output styling (FR-010)

No command writes ANSI escape codes, spinners, or progress output to
stdout or stderr. `internal/cli` depends on no terminal-styling library.
