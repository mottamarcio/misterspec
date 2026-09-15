# Phase 0 Research: Embedded Kit and Resource Installer

As with the prior features, the frozen envelope
(`docs/architecture-specification.md` §32-33, §57-58, §70 and the
ratified constitution) leaves no open `NEEDS CLARIFICATION`. This
document records the package-placement and reuse decisions made within
that envelope.

## Decision: `kit` is a top-level package, not under `internal/`

- **Decision**: Create `kit/kit.go` (package `kit`) at the repository
  root, holding the `go:embed` directive(s), with `kit/templates/*.tmpl`
  as the embedded files.
- **Rationale**: `docs/architecture-specification.md` §32 places `kit/`
  as a sibling of `cmd/` and `internal/` at the module root, not nested
  inside `internal/`. There's no reason to restrict it via Go's
  `internal/` visibility rules — nothing outside this module needs to
  import misterspec's Go packages at all (it's a compiled CLI tool, not a
  library), so the placement is purely about matching the architecture's
  intended resource layout, not about access control.
- **Alternatives considered**: `internal/kit` — rejected as contradicting
  §32's explicit layout without a concrete benefit; the module isn't
  consumed as a library, so `internal/`'s enforcement doesn't add
  anything here.

## Decision: Relocate 003-entity-creation's templates into `kit/templates/`, `internal/templates` reads from `kit`

- **Decision**: Move the 8 `.tmpl` files from `internal/templates/files/`
  to `kit/templates/`. `internal/templates/templates.go` drops its own
  `//go:embed files/*.tmpl` and reads file content via
  `kit.TemplatesFS.ReadFile(...)` instead. `internal/templates`'s public
  API (`Render`, `Kind`, the `*Data` structs) is unchanged — no caller in
  `internal/operations` needs to change at all.
- **Rationale**: This is spec.md's User Story 3 made concrete — one
  source of truth for template content (FR-008), matching §32's intended
  layout, verified by 003-entity-creation's existing test suite passing
  unmodified (SC-005).
- **Alternatives considered**: Leaving templates embedded privately in
  `internal/templates` and having `kit/templates/` hold a *second*,
  parallel copy for installation purposes — rejected outright; two copies
  of the same content is exactly what FR-008 exists to prevent, and it's
  a live drift risk (one gets edited, the other doesn't) that the
  project's own DRY discipline (Constitution Principle VI) rules out.

## Decision: New package `internal/installer`, matching §32's own file list exactly

- **Decision**: `internal/installer/{installer.go, filesystem.go}` — this
  is the one part of this feature's structure `docs/architecture-specification.md`
  §32 names outright (unlike prior features' occasional deviations from
  the sketch, here the doc's own layout is followed as-is).
  `installer.go` holds `List`/`Install`/the `Resource`/`Outcome` types;
  `filesystem.go` holds the low-level atomic-write helper.
- **Rationale**: Matches the architecture spec's own package sketch
  directly — no cycle or layering conflict this time requires deviating
  from it (unlike `internal/validation`'s situation in
  004-structural-validation).

## Decision: `filesystem.go` gets its own atomic-write helper, not a reuse of `operations`'s

- **Decision**: `internal/installer/filesystem.go` implements the same
  temp-file-then-`fsync`-then-`os.Rename` recipe
  `internal/operations/atomic_write.go` already has, as its own small
  (~30 line) unexported helper — not imported from `operations`.
- **Rationale**: `operations`'s `writeAtomic` is unexported (package-
  private) specifically because 002-read-operations and
  003-entity-creation never needed it outside that package. Exporting it
  now to be reused by `installer` — a sibling primitive package, not a
  consumer of `operations`'s higher-level compositions — would create an
  architecturally backwards dependency (a lower-level materialization
  primitive depending on the higher-level operations-composition
  package for a three-line file-writing recipe). Duplicating this one
  small, stable, already-proven recipe is a smaller cost than that
  coupling, and mirrors 004-structural-validation's identical reasoning
  for its own small duplicated parent-type table.
- **Alternatives considered**: Exporting `operations.WriteAtomic` and
  importing `operations` from `installer` — rejected for the layering
  reason above. Extracting a brand-new shared `internal/atomicfile`
  package for two current call sites — rejected as premature
  generalization (Constitution Principle IV); revisit if a third
  consumer needs it.

## Decision: Reuse `artifacts.RelativeWithinRoot` for installation containment

- **Decision**: `Install` computes each resource's destination path
  relative to the target directory and checks it via the already-
  exported, already-tested `artifacts.RelativeWithinRoot` (exported in
  002-read-operations) before writing anything (FR-006).
- **Rationale**: This is the identical containment guarantee every prior
  feature's writes and reads already rely on — one proven implementation,
  reused again (Constitution Principle VI, VIII), not a fourth
  reimplementation of the same boundary check.
- **Alternatives considered**: A separate containment check specific to
  `installer` — rejected as unjustified duplication of a check already
  proven correct three features running.

## Decision: `Install` materializes raw resource files, not rendered content

- **Decision**: `Install` copies each embedded resource's raw bytes to
  its target path — it does not execute `internal/templates`'s
  `text/template` rendering. Rendering requires entity-specific data (an
  ID, a parent) that doesn't exist at install time; installation is about
  making the framework's resources available on disk (e.g., for a future
  `.misterspec/` mirror `misterspec init` will populate), not about
  creating any project artifact.
- **Rationale**: Conflating "install the framework's resources" with
  "create a project entity" would blur exactly the boundary spec.md's
  Assumptions draws between this feature and 003-entity-creation's
  `Create`/`CreateArtifact` — those remain the only things that write
  into a project's `ai/` tree; `Install` writes framework resources
  elsewhere entirely.
- **Alternatives considered**: Having `Install` render each template with
  placeholder data — rejected as meaningless output with no real
  consumer request behind it.

## Output

All unknowns resolved. No `NEEDS CLARIFICATION` markers remain.
