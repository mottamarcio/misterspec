# Phase 0 Research: CLI Command Layer (Cobra)

`docs/architecture-specification.md` §4-8, §9-18, and §32 give this
feature an unusually concrete envelope — exact command names, exact JSON
shapes, an exact recommended package layout, and an exact list of
allowed vs. forbidden commands. This document records the decisions
still needed within that envelope: which existing Go capability maps to
which command, how the frozen JSON contracts get produced without a
second implementation, and the one new external dependency this feature
introduces (Cobra itself — the project's first since `gopkg.in/yaml.v3`
in 001-core-foundation).

## Decision: `github.com/spf13/cobra` v1.10.2, the project's second external dependency

- **Decision**: Add `github.com/spf13/cobra` (latest stable, v1.10.2 at
  time of writing) as a direct dependency. `cmd/misterspec/main.go`
  becomes the module's first binary entrypoint.
- **Rationale**: The user's own request and every prior suggestion
  named Cobra specifically, matching
  `docs/architecture-specification.md` §32's own recommended layout
  (`internal/cli/{root.go, init.go, internal.go}`). Cobra is the
  de facto standard for exactly this shape of tool (a small public
  surface plus a large hidden subcommand tree with per-command flags),
  and hand-rolling flag parsing would duplicate work Cobra already
  solves well, violating Constitution Principle IV (YAGNI) in the
  opposite direction — reinventing infrastructure that isn't this
  project's concern.
- **Alternatives considered**: Standard library `flag` package only —
  rejected; §32's own illustrative structure already names Cobra, and a
  hidden `internal` command tree with per-command flags is exactly
  Cobra's designed use case. `urfave/cli` or other alternatives —
  rejected as not what was requested and not what the architecture
  document itself illustrates.

## Decision: One `internal/cli` package, one file per command family, mirroring §32 exactly

- **Decision**: `internal/cli/root.go` (the root command, hidden-tree
  wiring), `internal/cli/init.go` (the public `init` command), and
  `internal/cli/internal.go` (the hidden `internal` command tree and its
  subcommands) — plus one file per operation under a new
  `internal/cli/internalcmd/` subpackage where a single `internal.go`
  would otherwise grow unreasonably large (see next decision).
  `cmd/misterspec/main.go` is a minimal entrypoint that only calls
  `cli.Execute()`.
- **Rationale**: Matches `docs/architecture-specification.md` §32's own
  recommended structure file-for-file at the top level. Splitting the
  ten operation subcommands into their own small files (rather than one
  1000-line `internal.go`) follows Constitution Principle VI (Clean
  Code) the same way every prior feature has kept files single-purpose
  — `internal.go` itself stays limited to constructing the hidden
  `internal` parent command and registering each subcommand's own
  constructor.
- **Alternatives considered**: A single flat `internal.go` with all ten
  subcommands inline — rejected as the one place in this codebase where
  a single file would meaningfully exceed every prior feature's own
  file-size discipline.

## Decision: Every command is a thin adapter — JSON marshaling lives in one shared envelope helper, not duplicated per command

- **Decision**: A single, small helper package,
  `internal/cli/internalcmd/envelope.go`, provides two functions:
  `WriteSuccess(w io.Writer, key string, value any) error` (marshals
  `{"ok": true, "<key>": value}`) and
  `WriteError(w io.Writer, code cliError) int` (marshals
  `{"ok": false, "error": {"code": ..., "message": ...}}` and returns
  the matching exit code). Every one of the ten operation commands calls
  exactly one of these — never marshals JSON itself.
- **Rationale**: `docs/architecture-specification.md` §6-8 fixes the
  envelope shape and the exit-code categories once, globally — not
  per-command. Ten independent JSON-marshaling call sites would risk the
  exact drift Constitution Principle VI (DRY) warns against (the same
  reasoning 006-agent-adapter's research.md used for `RecordInstall`,
  007-project-bootstrap's for `writeDefaultConfig`). This also gives the
  error→exit-code mapping exactly one place to get right and one place
  every future command reuses automatically.
- **Alternatives considered**: Each command's `RunE` marshaling its own
  JSON inline — rejected; already-established DRY precedent from every
  prior feature's shared-helper pattern.

## Decision: A `cliError` table maps each package's existing sentinel errors to §7's stable codes and §8's exit codes — once, not per-command

- **Decision**: One file, `internal/cli/internalcmd/errors.go`, defines
  a `classify(err error) (code string, exitCode int)` function using
  `errors.Is`/`errors.As` against every sentinel error 001-007 already
  export (`project.ErrNotInitialized`, `operations.ErrEntityNotFound`,
  `operations.ErrEntityAmbiguous`, `operations.ErrInvalidTarget`,
  `operations.ErrInvalidParent`, `operations.ErrAlreadyExists`,
  `operations.ErrUnsupportedType`, `operations.ErrInvalidSlug`,
  `artifacts.ErrPathOutsideProject`, `artifacts.ErrArtifactNotFound`,
  `bootstrap.ErrAlreadyInitialized`, `bootstrap.ErrUnknownAgent`, and a
  final fallback for anything unrecognized). Every command's `RunE`
  calls `classify` on its operation's returned error and passes the
  result to `WriteError` — no command hand-picks its own error code.
- **Rationale**: §7's code list is explicitly extensible ("possible
  stable error codes include") — this feature adds two new codes
  (`already_initialized`, `unknown_agent`) for 007-project-bootstrap's
  two sentinels, which have no equivalent in §7's illustrative list,
  following its own stated pattern rather than forcing them into an
  existing code that doesn't actually match. Centralizing the mapping
  means every command automatically gets a correct, consistent
  classification the moment its underlying package's error is added
  here once — not reimplemented ten times with the risk of drifting
  mappings for the same sentinel error.
- **Alternatives considered**: Each command switching on its own
  expected errors inline — rejected for the DRY/drift reasoning above,
  and because several sentinels (e.g. `artifacts.ErrPathOutsideProject`)
  are reachable from more than one command, which a shared classifier
  handles once instead of N times.

## Decision: `misterspec init` is `internal/bootstrap.Bootstrap` behind flags — no interactive prompt, no TUI

- **Decision**: `internal/cli/init.go` defines `--agent` (required
  string) and `--dir` (string, default `.`) flags, calls
  `bootstrap.Bootstrap(dir, agent, builtin.Default(), kit.SkillsFS)`
  directly, and reports the result through the same envelope helper
  every internal command uses.
- **Rationale**: spec.md's Assumptions already scope this out explicitly
  — the interactive Bubble Tea flow (§36-37: agent selection, preview,
  confirm) is a distinct, later feature. `kit.SkillsFS` still does not
  exist (Phase 6's canonical Skill content, unchanged scoping from every
  prior feature) — see the next decision for how this feature handles
  that gap without inventing new Skill content itself.
- **Alternatives considered**: Deferring the public `init` command
  entirely until the interactive flow exists — rejected; spec.md's own
  User Story 2 explicitly scopes a non-interactive `init` as valuable on
  its own (a scriptable/CI-usable bootstrap), independent of when the
  TUI arrives, and 007-project-bootstrap's `Bootstrap` was built
  precisely to make this possible without waiting on Bubble Tea.

## Decision: `kit.SkillsFS` is introduced now as a real embed with one placeholder file, not deferred again

- **Decision**: `kit/skills/README.md` (a plain, visible file
  explaining canonical Skill content is Phase 6 future work) and
  `kit/kit.go` gains `//go:embed skills` → `fs.Sub`'d down to
  `SkillsFS fs.FS`, alongside the existing `TemplatesFS`. `misterspec
  init` passes this real, currently-one-file `fs.FS` to `Bootstrap`, not
  a fixture.
- **Rationale**: Every prior feature (005 through 007) deferred
  `kit.SkillsFS`'s *content* (Phase 6, canonical Skills) but always
  accepted a caller-supplied `fs.FS` specifically so a real embed could
  slot in later without an API change (006's and 007's own research.md
  decisions say so explicitly). This feature is that later point for the
  *public command*. The original plan for this decision (recorded here
  during Phase 1 design) assumed a hidden `.gitkeep` placeholder would
  make `go:embed skills` compile against an otherwise-empty directory —
  **implementation discovered this is false**: Go's `go:embed` silently
  excludes any file or directory whose name starts with `.` or `_`
  unless the pattern is prefixed `all:`, and an otherwise-empty
  directory (nothing left to match) is a hard compile error
  ("contains no embeddable files"), confirmed by direct experiment. Using
  `all:skills` to force-include a `.gitkeep` was considered but rejected:
  it would materialize as a stray hidden `.claude/skills/.gitkeep` file
  on every bootstrap — less honest and less clean than a single, plainly
  named, clearly self-documenting `README.md` placeholder that
  materializes exactly the same way any future canonical Skill file
  will. `Adapter.Install`'s reported `Outcomes` therefore contains one
  entry (`README.md`, `Installed`) today, not zero — quickstart.md is
  updated to match.
- **Alternatives considered**: A hidden `.gitkeep` placeholder with a
  plain (non-`all:`) embed pattern — rejected; empirically does not
  compile (`go:embed` requires at least one *matched*, non-hidden file).
  `all:skills` with a hidden placeholder — rejected as producing a
  stray, less-honest hidden file in the installed result, per above.
  Passing a hardcoded, hand-built `fstest.MapFS` from
  `internal/cli/init.go` itself instead of a real embed — rejected; that
  would mean the actual public binary ships test fixture data as if it
  were real Skill content, which is worse than one honestly-labeled
  placeholder file until Phase 6 authors real Skills.

## Decision: `references` and a dedicated `project` command are NOT added, even though §9.1 and §15 name them

- **Decision**: This feature exposes exactly the ten operations spec.md's
  FR-001 already enumerates (resolve, inspect, parent, children, create,
  create-artifact, fingerprint, inventory, validate, status). It adds no
  `misterspec internal project` or `misterspec internal references`
  command.
- **Rationale**: `internal/operations` has no `References` function —
  002-read-operations never built one (its own scope explicitly listed
  resolve/inspect/parent/children/fingerprint/inventory only), and no
  later feature added one either. Adding a CLI command for an operation
  that doesn't exist would violate this feature's own FR-009 ("no second
  implementation" — there would be no *first* implementation to adapt,
  meaning the CLI layer would have to invent the business logic itself,
  exactly the layering violation this feature exists to avoid). A
  `project` command is arguably free (it would only read
  `project.Detect`'s already-existing `Project`/`Configuration` struct),
  but spec.md's own FR-001 — already validated and approved before this
  plan — enumerates exactly ten operations and does not include it;
  expanding the command surface beyond an approved spec inside planning
  is exactly what `/speckit-plan`'s own gate exists to prevent.
- **Alternatives considered**: Adding `project` since it needs no new
  logic — rejected; still out of spec.md's approved scope. Building a
  `references` implementation as part of this feature to unblock the
  command — rejected; that would be new business logic inside a feature
  spec.md explicitly scoped as "no new business logic," and belongs in
  its own future, focused feature (a "009" candidate) with its own
  research and tests, the same way 002-006 each stayed narrowly scoped.

## Decision: `children`'s optional `--type` filter maps directly to `Children`'s existing `filterType *ids.EntityType` parameter

- **Decision**: `misterspec internal children <id> --type feature`
  parses `--type`'s string value against a small, local
  name-to-`ids.EntityType` table in `internal/cli/internalcmd/children.go`
  (`ids.EntityType.String()` already renders the inverse — "feature" for
  `ids.Feature` — but `ids` exports no name-to-type parser, only
  `TypeForPrefix`'s prefix-to-type one, since no prior feature needed
  one; this is CLI-boundary input parsing, not a second implementation
  of anything `ids` already does) and passes the result as `Children`'s
  `filterType`; an absent flag passes `nil` exactly as `Children`'s own
  "no filter" case already expects. An unrecognized `--type` value is an
  `invalid_argument` error before `Children` is even called.
- **Rationale**: §9.5's example shows this exact flag; `Children`'s
  Go signature already accepts an optional filter — direct mapping onto
  existing business logic, with only the flag-string-to-enum conversion
  (unavoidable at any CLI boundary) written new here.
- **Alternatives considered**: Adding a name-to-type parser to the `ids`
  package itself — rejected as speculative for a single CLI flag's sake;
  a package-local, unexported table in the one command that needs it is
  simpler and matches Constitution Principle IV (YAGNI).

## Decision: JSON field values reuse each Go type's own `String()` naming consistently, even where §9/§18's illustrative examples use a different form

- **Decision**: Every place a JSON field names an entity/artifact type
  (`"type"` on `resolve`/`inspect`/`children`/`create`/`create-artifact`,
  and `status`'s per-type counts) uses `ids.EntityType.String()` or
  `artifacts.ArtifactType.String()` verbatim — singular, lowercase
  (`"feature"`, `"spec"`, `"plan"`). `status`'s `counts` object is keyed
  the same way (`"program"`, `"feature"`, `"spec"`, `"task"`,
  `"knowledge"`, `"learning"`) rather than §18's illustrative
  `"programs"`/`"features"`/`"specs"` (plural).
- **Rationale**: `Counts` is `map[ids.EntityType]int` — `EntityType` is
  an int-kind type with no `MarshalText`, so it cannot marshal as JSON
  object keys directly; this feature must already build a
  `map[string]int` by hand from it. Given that hand-built step is
  unavoidable, reusing `EntityType.String()`/`ArtifactType.String()`
  (already used verbatim for every other command's `"type"` field) keeps
  one consistent naming convention across the entire CLI surface, rather
  than introducing a second, plural naming scheme that exists only for
  `status`'s illustrative example. §18 itself frames its JSON as "output
  may include" (not a frozen literal schema the way §6's error envelope
  is), so this is a reasonable, internally-consistent choice within that
  latitude — recorded here explicitly per this project's standing
  practice of documenting every deviation from an illustrative example.
- **Alternatives considered**: A separate, one-off plural-naming table
  just for `status`'s `counts` keys, matching §18 literally — rejected
  as inconsistent with every other command's singular `"type"` values
  for no functional benefit, and as introducing a second naming
  convention to maintain.

## Decision: `resolve`'s JSON omits §9.2's illustrative `"directory"` field

- **Decision**: `misterspec internal resolve` returns `{"id", "type",
  "path"}` — no separate `"directory"` field.
- **Rationale**: `ResolvedLocation` (002-read-operations) carries `Path`
  only — for a Task, `Path` is `"<tasks.md path>#TASK-NNN"`, which has
  no single well-defined "directory" the way a Program/Feature/Spec's
  path does. A caller needing a directory can already derive one from
  `Path` for the entity types where that concept applies. Introducing a
  second field here would mean either inventing new data `ResolvedLocation`
  doesn't carry (new business logic, against this feature's own FR-009)
  or leaving it null/wrong for Task — worse than omitting it.
- **Alternatives considered**: Computing `filepath.Dir(Path)` and adding
  it as `"directory"` for non-Task entities only, `null` for Task —
  rejected as an inconsistent, type-dependent field shape for
  questionable benefit; omitted entirely is simpler and matches this
  feature's "no new business logic" constraint more cleanly.

## Decision: `validate` reporting `valid: false` still exits `4`, even though `ok` stays `true`

- **Decision**: `misterspec internal validate`'s process exit code is `4`
  whenever the resulting `findings` slice is non-empty (`valid: false`)
  — even though its JSON `ok` field is `true` (the command itself ran
  successfully). `ok`/`valid` remain independent JSON fields; only the
  exit code additionally reflects `valid`.
- **Rationale**: §16 is explicit that `ok` and `valid` "must remain
  distinct" *concepts* — that is a statement about the JSON envelope,
  not a claim that the process exit code must ignore `valid` entirely.
  §8 separately reserves exit code `4` specifically for "structural
  validation failure" as its own category, distinct from `0`
  (success) — a reserved code with no use would be pointless to
  document. Standard Unix tooling convention for this exact shape
  (a linter/validator that ran correctly but found problems) is a
  non-zero exit distinct from a crash — `golangci-lint`, `eslint`, and
  similar tools all exit non-zero on findings while still having "run
  successfully." This lets a shell script or CI step fail-fast on
  `misterspec internal validate`'s exit code alone, while a Skill
  reading the JSON still gets `ok`/`valid` as independently meaningful
  fields per §16's own requirement.
- **Alternatives considered**: Exit `0` whenever `ok: true`, regardless
  of `valid`, leaving `4` permanently unused — rejected; it would make
  §8's own documented exit-code table incomplete without any winning
  argument in exchange (a caller parsing JSON does not lose anything by
  the exit code *also* reflecting validity — the JSON is unchanged
  either way), and it would leave the tool's own scriptability degraded
  for the single most common case where scripting on exit code matters
  (an install- or CI-hook running `validate`).

## Decision: `inventory <dir>`'s positional argument is a literal project-relative path, not §14's illustrative scope keywords

- **Decision**: `misterspec internal inventory <dir>` passes `<dir>`
  straight through to `operations.Inventory(root, dir)` as-is (e.g.
  `ai/raw`, `ai/knowledge`) — it does not translate a shorthand keyword
  like `"raw"` or `"knowledge"` into `cfg.RawDir`/`cfg.KnowledgeDir`
  first.
- **Rationale**: §14's example (`misterspec internal inventory raw`)
  illustrates a scope keyword, but `operations.Inventory`
  (002-read-operations) takes a directory path directly — "relative to
  root, or an absolute path within it," per its own doc comment — with
  no keyword-to-path translation table anywhere in `internal/operations`.
  Adding one here would be new business logic this feature's own FR-009
  rules out, for a convenience `cfg.RawDir`/`cfg.KnowledgeDir` (both
  already known, short, copy-pasteable strings from `internal
  project`'s output or the project's own `config.yaml`) doesn't
  meaningfully need. A caller passes the actual path.
- **Alternatives considered**: Adding a small keyword table
  (`"raw"`→`cfg.RawDir`, `"knowledge"`→`cfg.KnowledgeDir`,
  `"programs"`→`cfg.ProgramsRoot`) in the CLI layer only — rejected as
  unrequested convenience logic living in the wrong layer (a CLI
  ergonomics decision, not a deterministic operation), and as scope
  creep beyond what spec.md's FR-001 approved (expose the existing
  operation, not extend it).

## Output

All unknowns resolved. No `NEEDS CLARIFICATION` markers remain.
