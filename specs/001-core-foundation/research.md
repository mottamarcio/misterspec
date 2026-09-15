# Phase 0 Research: Core Repository Foundation

All Technical Context values below come from decisions already frozen in
`docs/architecture-specification.md` (§32, §70) and the ratified
`.specify/memory/constitution.md`, not from open exploration — so there are
no `NEEDS CLARIFICATION` markers to resolve. This document records the
choices and the specific implementation decisions this feature still had to
make within that frozen envelope.

## Decision: Go module path & toolchain

- **Decision**: Module path `github.com/mottamarcio/misterspec`; target Go
  1.23+ (module `go` directive set to `1.23`).
- **Rationale**: Matches the project's GitHub remote
  (`git@github.com:mottamarcio/misterspec.git`) so `go get`/imports resolve
  correctly for any future consumer; Go 1.23 is a current stable toolchain
  with generics, `slices`/`maps` stdlib packages, and `testing/fstest`
  available — all useful for this feature's ID/path logic and fixture-based
  tests, with no third-party toolchain requirement.
- **Alternatives considered**: Pinning to an older Go version — rejected,
  no compatibility constraint exists yet (§70 doesn't pin a version) and an
  older baseline would only remove available stdlib conveniences.

## Decision: YAML frontmatter parsing library

- **Decision**: `gopkg.in/yaml.v3` for parsing the frontmatter block of
  artifact files.
- **Rationale**: It is the de facto standard Go YAML library, supports
  strict/typed unmarshaling (`KnownFields(true)`) which this feature needs
  to distinguish "malformed YAML" from "unknown/extra field" from "missing
  required field" (FR-009), and has no further transitive dependencies of
  concern.
- **Alternatives considered**: `sigs.k8s.io/yaml` (JSON-tag based, adds an
  indirection through JSON that complicates strict-field detection) —
  rejected as unnecessary indirection for this use case. Hand-rolled YAML
  subset parser — rejected as reinventing a well-solved problem and a
  direct violation of Principle IV (YAGNI).

## Decision: Frontmatter delimiter & extraction strategy

- **Decision**: Frontmatter is the content between the first two lines
  consisting solely of `---`, at the very start of the file; everything
  after the closing `---` is the Markdown body (not parsed by this
  feature — body structure validation belongs to a later Validation
  feature).
- **Rationale**: Matches every schema example in
  `docs/architecture-specification.md` §22–31 exactly.
- **Alternatives considered**: Supporting arbitrary frontmatter positions —
  rejected; the architecture spec's schemas are frozen at "frontmatter
  first," so supporting anything else is speculative scope (Principle IV).

## Decision: Filesystem traversal for project detection & ID scanning

- **Decision**: Use Go's standard `io/fs` / `path/filepath` walking
  (`filepath.WalkDir` or `fs.WalkDir` over an `os.DirFS`) rather than a
  third-party directory-walking library.
- **Rationale**: The standard library is sufficient for both "walk upward
  from cwd looking for `.misterspec/config.yaml`" (project detection) and
  "walk downward under `ai/programs/**` (or the configured artifacts root)
  collecting entity IDs" (ID scanning); no additional dependency is
  justified (Principle IV).
- **Alternatives considered**: A third-party glob/walk library — rejected,
  no capability gap exists that the stdlib doesn't already cover.

## Decision: Error modeling for distinct failure conditions

- **Decision**: Each distinguishable failure named by the functional
  requirements (not-initialized, invalid-configuration, artifact-not-found,
  frontmatter-malformed, missing-required-field, invalid-id-syntax,
  duplicate-id, path-outside-root) is a typed Go sentinel/wrapped error
  (`errors.Is`-compatible), not a bare `error` built from `fmt.Errorf`
  string matching.
- **Rationale**: FR-003, FR-005, and FR-009 explicitly require distinct,
  reportable conditions rather than a single generic error. Typed errors
  also let this feature's error vocabulary map cleanly, one-to-one, onto
  the stable JSON error codes a later CLI feature will expose (§7 of the
  architecture spec: `project_not_initialized`, `invalid_metadata`,
  `invalid_reference`, `path_outside_project`, etc.) without renaming
  anything at that point.
- **Alternatives considered**: String-matching on error messages — rejected,
  brittle and exactly what Principle IX's "reason about error codes, not
  prose" exists to prevent, even before the JSON layer exists.

## Decision: Scope boundary — what this feature does NOT build

- **Decision**: This feature builds read-only detection/resolution/parsing/
  scanning only. It does **not** build: the `misterspec internal ...` CLI
  surface, JSON output formatting, the atomic `create` operation, the
  allocation lock (`.misterspec/.lock`), or any file-writing code path.
- **Rationale**: Matches `docs/architecture-specification.md` §68's own
  phase split (Phase 1 vs. Phase 2) and Principle IV — building the write
  path before the read/detection path it depends on is unjustified,
  speculative sequencing.
- **Alternatives considered**: Building `create`/allocation now since it's
  "just one more function" — rejected; the spec (001-core-foundation) and
  its Assumptions section explicitly scope this feature to the read-only
  foundation layer.

## Output

All unknowns resolved. No `NEEDS CLARIFICATION` markers remain in Technical
Context or elsewhere in this plan.
