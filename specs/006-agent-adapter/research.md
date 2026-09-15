# Phase 0 Research: Agent Adapter Layer

As with the prior features, the frozen envelope
(`docs/architecture-specification.md` §22, §32, §34-35, §70 and the
ratified constitution) leaves no open `NEEDS CLARIFICATION`. This
document records the package-placement, cycle-avoidance, and reuse
decisions made within that envelope.

## Decision: Generalize 005-embedded-kit's `internal/installer` instead of a third atomic-materialization implementation

- **Decision**: Add `InstallFS(source fs.FS, sourceDir, targetDir string, overwrite bool) ([]Outcome, error)` and `ListFS(source fs.FS, sourceDir string) []Resource` to `internal/installer`. The existing `Install(targetDir, overwrite)` and `List()` become thin wrappers —
  `Install(targetDir, overwrite) { return InstallFS(kit.TemplatesFS, "templates", targetDir, overwrite) }` — so 005's exported API and every existing caller are unchanged.
- **Rationale**: `internal/agents`' Claude adapter needs the identical
  atomic-write, containment-checked, no-silent-overwrite materialization
  005-embedded-kit already built and proved — only the source (`fs.FS`)
  and destination differ (Skills into `.claude/skills`, not templates
  into an arbitrary target). Reimplementing that logic a third time
  (after `internal/operations`'s and `internal/installer`'s own) would
  directly violate Constitution Principle VI (DRY) for no benefit.
- **Alternatives considered**: A separate, agent-specific materialization
  helper inside `internal/agents` — rejected as exactly the duplication
  DRY forbids. Making `internal/installer` agent-aware directly (skip the
  generalization, hardcode a second source) — rejected as it would
  couple a generic resource-installation primitive to
  agent-specific concepts it has no reason to know about.
- **Verification**: 005-embedded-kit's existing test suite (`List`/
  `Install` against `kit.TemplatesFS`) must pass unmodified after this
  change — the same explicit regression discipline 005 itself used for
  003's suite.

## Decision: A separate `internal/agents/builtin` wiring package, to avoid an import cycle

- **Decision**: `internal/agents` holds only the `Adapter` interface,
  `InstallRequest`/`InstallResult`/`InstallRecord` types, and the
  `Registry` type/constructor — it has zero knowledge of any concrete
  adapter. `internal/agents/claude` implements `Adapter` (and so imports
  `internal/agents` for those types — one direction). A third, small
  package, `internal/agents/builtin`, imports *both* `internal/agents`
  and `internal/agents/claude` and exposes `Default() *agents.Registry`,
  pre-populated with every concrete adapter this build knows about.
- **Rationale**: `internal/agents` cannot both define the `Adapter`
  interface *and* construct `claude.New()` for a convenience "default
  registry" — that would require `agents` → `claude` (to construct it)
  and `claude` → `agents` (for the interface/types) simultaneously, an
  import cycle Go refuses to compile. This is the identical situation
  004-structural-validation hit between `validation` and `operations`,
  resolved the identical way: one side gives, and a separate wiring point
  (there: `operations` importing `validation`; here: a dedicated
  `builtin` package) breaks the cycle.
- **Alternatives considered**: Concrete adapters self-registering via a
  package-level `init()` side effect when blank-imported (`import _
  ".../agents/claude"`) — rejected as unjustified "magic" for exactly one
  adapter (Constitution Principle IV); a registry built once, explicitly,
  by a caller that knows which adapters exist is simpler and more
  legible with only one adapter to wire up. Revisit only if the number of
  adapters and callers grows enough that explicit wiring becomes
  repetitive.

## Decision: `Registry` is a constructed value, not global mutable state

- **Decision**: `agents.NewRegistry(adapters ...Adapter) *Registry`
  builds an immutable-after-construction lookup; there is no
  package-level mutable registry variable or `init()`-time registration.
- **Rationale**: Matches this project's existing style (`operations`,
  `validation` are all plain functions/constructed values, never hidden
  global state) and makes testing trivial — a test builds its own
  `Registry` from fixture adapters, never depending on the real Claude
  adapter or import-order side effects.
- **Alternatives considered**: A package-level `var registry = map[...]`
  mutated by an exported `Register` function — rejected as global mutable
  state with no concrete need for it yet (only one build-time-known set
  of adapters exists; nothing dynamically registers adapters at runtime).

## Decision: `InstallRequest.Skills` is a caller-supplied `fs.FS`, not a hardcoded `kit.SkillsFS`

- **Decision**: `InstallRequest` carries a `Skills fs.FS` field. The
  Claude adapter's `Install` calls `installer.InstallFS(req.Skills, ".",
  target, req.Overwrite)` — it never references `kit` directly.
- **Rationale**: `kit/skills/` has no real content yet (Phase 6, spec.md
  Assumptions) — `go:embed` cannot even declare a pattern over an empty
  directory. Accepting the Skills source as a parameter means this
  feature's tests can supply a real, working fixture `fs.FS`
  (`fstest.MapFS` or a `t.TempDir()`-backed `os.DirFS`) today, and the
  eventual caller (a future `misterspec init`) passes `kit.SkillsFS` once
  Phase 6 adds it — a one-line change at that call site, no interface
  change required here.
- **Alternatives considered**: Hardcoding `kit.SkillsFS` now, with a
  placeholder file just so `go:embed` compiles — rejected; a placeholder
  Skill file would be exactly the speculative content
  005-embedded-kit's own Assumptions already ruled out for `kit/skills/`,
  and it would need to be deleted and replaced the moment Phase 6 lands
  rather than just wired in.

## Decision: `install.json` schema and version constants, per §22 exactly

- **Decision**: `InstallRecord` mirrors
  `docs/architecture-specification.md` §22's schema exactly:
  `schema_version` (int, `1`), `misterspec_version` (string, hardcoded
  `"0.1.0"` — this document's own stated MVP target release), and
  `agent.id`/`agent.integration_path`. Written to
  `<ProjectRoot>/.misterspec/install.json` via the same atomic-write
  helper `internal/installer` already has (reused, not reimplemented a
  third time within `agents`).
- **Rationale**: The schema is already frozen exactly in the architecture
  spec; no version-detection or build-metadata machinery is needed for a
  single hardcoded MVP version string.
- **Alternatives considered**: Deriving `misterspec_version` from a build
  flag or VCS tag — rejected as unbuilt infrastructure (`cmd/` doesn't
  exist yet) this feature has no need to invent.

## Decision: `Adapter.Install` both materializes and writes the record

- **Decision**: A single call to `Adapter.Install(ctx, req)` performs the
  full operation FR-005/FR-006 describe: materialize Skills, then write
  `install.json`, returning the completed `InstallResult`. The
  record-writing step is a small shared helper in `internal/agents`
  (`recordInstall`), which every adapter implementation calls internally
  — so a second adapter later reuses it rather than reimplementing
  record-writing.
- **Rationale**: Matches `docs/architecture-specification.md` §34's
  interface shape most literally (one `Install` call, one outcome) and
  keeps a caller's mental model simple: "install for this adapter" is
  one call, not an install-then-separately-record two-step (echoing why
  003-entity-creation's `Create` is one atomic call rather than
  allocate-then-scaffold).
- **Alternatives considered**: A separate `agents.InstallAndRecord`
  orchestration wrapper around a record-free `Adapter.Install` —
  rejected as an unnecessary extra public entry point for what should be
  one coherent operation from a caller's point of view.

## Decision: `context.Context` stays in `Install`'s signature, unused today

- **Decision**: `Adapter.Install(ctx context.Context, req InstallRequest)
  (InstallResult, error)` keeps the `ctx` parameter exactly as
  `docs/architecture-specification.md` §34 specifies, even though this
  feature's local-filesystem-only work has nothing cancellable to do
  with it yet.
- **Rationale**: This is matching an already-frozen upstream interface
  shape, not adding new speculative capability — the distinction
  Constitution Principle IV draws is between building unrequested
  capability and following a contract that's already fixed. A future
  adapter fetching something remote (unlikely per §33's "no network"
  rule, but not this feature's call to foreclose) would need it; removing
  it now would just mean re-adding it later and breaking every adapter's
  signature at that point.
- **Alternatives considered**: Dropping `ctx` since nothing uses it today
  — rejected; §34's interface is explicitly frozen, and every other
  feature in this project has treated frozen interface shapes from the
  architecture spec as non-negotiable unless a real conflict (like an
  import cycle) forces a documented deviation. No such conflict exists
  here.

## Output

All unknowns resolved. No `NEEDS CLARIFICATION` markers remain.
