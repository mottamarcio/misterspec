# Phase 0 Research: Project Bootstrap

As with the prior features, the frozen envelope
(`docs/architecture-specification.md` §21, §36, §58, §70 and the
ratified constitution) leaves no open `NEEDS CLARIFICATION`. This
document records the package-placement and reuse decisions made within
that envelope — including one pre-existing discrepancy this feature must
be consistent with, not "fix" unilaterally.

## Decision: New top-level package `internal/bootstrap`, not an extension of `internal/project`

- **Decision**: `Inspect`, `Bootstrap`, and `Verify` live in a new
  package, `internal/bootstrap`, which composes `internal/project`,
  `internal/installer`, and `internal/agents` — none of those three gain
  new cross-dependencies on each other.
- **Rationale**: `internal/project` is this project's foundation layer —
  deliberately dependency-light (only stdlib + `yaml.v3`) since
  001-core-foundation, consumed *by* every later package, never
  depending *on* one of them. Writing a brand-new project's
  configuration needs `installer.WriteAtomicFile` (005-embedded-kit) for
  the same atomic-write guarantee everything else in this project uses —
  but having `project` import `installer` would be exactly the
  architecturally-backwards dependency 005-embedded-kit's own research.md
  already rejected once (there, for `installer` almost depending on
  `operations`). A new top-level composition package — the same role
  `operations`, `validation`, and now `bootstrap` all play — is where
  "compose the foundation into the answer a caller wants" belongs.
- **Alternatives considered**: Adding a `project.Bootstrap` function that
  imports `installer` — rejected for the layering-inversion reason above.
  Putting this logic in `internal/operations` — rejected; `operations`
  composes primitives into *entity-level* answers (resolve, inspect,
  create an artifact); project-level bootstrapping is a different
  concern at a different scope, deserving its own package the same way
  `validation` didn't get folded into `operations` in
  004-structural-validation.

## Decision: New project configuration mirrors `project.Configuration`'s actual (flat) shape — not §21's illustrative nested YAML

- **Decision**: `bootstrap`'s config-writing logic constructs a
  `project.Configuration` value (reusing its already-exported
  `Default*` constants from 001-core-foundation) and marshals *that
  struct* directly, rather than hand-writing YAML text.
- **Rationale**: `docs/architecture-specification.md` §21 shows a
  *nested* YAML example (`agent: {id: ...}`, `project: {artifacts_dir:
  ...}`, ...), but `project.Configuration`'s actual struct tags
  (`schema_version`, `agent_id`, `artifacts_dir`, ... — all flat, no
  nesting) were established back in 001-core-foundation and are what
  every one of the 001-006 features' test fixtures
  (`testutil.DefaultConfigYAML`) already use and what `project.Load`
  actually parses. This feature must write a config `project.Load` can
  read back (FR-010's whole point — the bootstrapped project must be
  detectable, not detectable-in-theory) — so it follows the schema that
  is *actually implemented*, not the illustrative example, and marshaling
  the real struct directly (instead of a separately hand-written template
  string) makes drift between "what's written" and "what's read"
  structurally impossible.
- **Alternatives considered**: Writing nested YAML matching §21's
  literal example — rejected; it would produce a config
  `project.Load` cannot parse, directly breaking FR-010. Retroactively
  changing `project.Configuration` to a nested shape to match §21 exactly
  — rejected as out of scope for this feature and a breaking change to
  six already-shipped features' fixtures for no functional gain; noted
  here as a pre-existing, harmless documentation/implementation
  discrepancy from 001, not something this feature is positioned to fix.

## Decision: `Bootstrap` takes an `*agents.Registry` and a Skills `fs.FS` as parameters, not `builtin.Default()`/`kit.SkillsFS` hardcoded

- **Decision**: `Bootstrap(targetDir, agentID string, registry
  *agents.Registry, skills fs.FS) (BootstrapOutcome, error)` — the
  caller supplies both.
- **Rationale**: `kit.SkillsFS` doesn't exist yet (Phase 6, unchanged
  from 006-agent-adapter's own scoping) — `skills` must still be
  caller-supplied for the same reason it was in `InstallRequest`.
  Accepting `*agents.Registry` rather than importing
  `internal/agents/builtin` directly keeps `bootstrap` testable against
  a fixture registry (a fake adapter, the same pattern
  006-agent-adapter's own `registry_test.go` used) without depending on
  the real Claude adapter's behavior — the eventual real caller (a
  future CLI) passes `builtin.Default()` itself.
- **Alternatives considered**: `bootstrap` importing
  `internal/agents/builtin` and calling `builtin.Default()` internally
  — rejected; it would make every `Bootstrap` test exercise the real
  Claude adapter's `Install` (network-free, but still unnecessarily
  coupling this package's tests to another package's concrete behavior)
  instead of a minimal fixture, the same reasoning 006 used to keep
  `Registry` tests adapter-agnostic.

## Decision: No new locking for `Bootstrap` — `Inspect`-then-reject is sufficient

- **Decision**: `Bootstrap` calls `Inspect` first and rejects
  (`ErrAlreadyInitialized`) if the target is already a project — no
  `internal/lock` (003-entity-creation) protection is added around the
  bootstrap write sequence itself.
- **Rationale**: `docs/architecture-specification.md` §58 scopes
  concurrency protection to "one active repository workflow" — the
  scenario `003-entity-creation`'s lock protects against (two *ordinary*
  `Create` calls racing during normal, expected usage, e.g. two Skills
  creating Specs back-to-back) has no equivalent here: bootstrapping a
  project is a one-time, explicitly-confirmed operation a future
  interactive flow will gate behind user confirmation, not something
  ordinary usage triggers repeatedly or concurrently. Adding a lock here
  without an identified concurrent-usage scenario would be exactly the
  speculative infrastructure Constitution Principle IV (YAGNI) rules out.
- **Alternatives considered**: Reusing `internal/lock` defensively anyway
  — rejected as unjustified for a scenario this project's own
  architecture spec doesn't identify as needing it; revisit only if a
  real concurrent-bootstrap scenario is identified later (e.g. a future
  `misterspec init` invoked twice in parallel by tooling, not a human).

## Decision: `Verify` reuses `Inspect` rather than a parallel check

- **Decision**: `Verify(targetDir, wantAgentID string) (VerifyResult,
  error)` calls `Inspect` internally and compares its
  `InstalledAgent` against `wantAgentID` — it does not re-derive project
  detection or installation-record reading a second way.
- **Rationale**: Constitution Principle VI (DRY) — "confirm the project
  detects and the right agent is installed" is exactly what `Inspect`
  already answers; a second, parallel implementation risks drifting from
  it exactly like 004-structural-validation's research.md warned against
  for `Status`/`ValidateProject`.
- **Alternatives considered**: A separate verification-specific check
  reading `.misterspec/install.json` directly — rejected as duplicating
  `Inspect`'s (and, transitively, `agents.CurrentInstall`'s) already-
  correct logic.

## Decision: `Inspect` calls `project.Detect(targetDir)` as-is — including its upward-walking ancestor search

- **Decision**: `Inspect(targetDir)` calls `project.Detect(targetDir)`
  unmodified rather than a bootstrap-specific "does exactly this
  directory contain `.misterspec/config.yaml`" check.
- **Rationale**: `Detect`'s upward walk (001-core-foundation) is
  deliberate — a project's root can be detected from any nested
  subdirectory. Reusing it here means bootstrapping correctly refuses not
  only a target that is itself already a project, but also one nested
  *inside* an existing project's tree — which is the right outcome (you
  do not want a second, nested misterspec project accidentally created
  inside an existing one), and it falls out of reusing `Detect` exactly
  as-is rather than needing a special case invented for this feature.
- **Alternatives considered**: A bootstrap-local check that only inspects
  `targetDir` itself (ignoring ancestors) — rejected; it would allow the
  nested-project footgun above and duplicate logic `Detect` already gets
  right, violating Constitution Principle VI (DRY) for no benefit.

## Output

All unknowns resolved. No `NEEDS CLARIFICATION` markers remain.
