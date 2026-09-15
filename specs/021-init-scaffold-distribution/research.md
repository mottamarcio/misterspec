# Phase 0 Research: Init Scaffolding and Binary Distribution

All unknowns spec.md's own Assumptions section deliberately deferred to
planning are resolved below. No `NEEDS CLARIFICATION` markers remain.

## 1. Remove the templates-install call, not the underlying primitive

**Decision**: Delete `bootstrap.Bootstrap`'s own single call to
`installer.Install(targetDir, false)` (and the `TemplateOutcomes`
field it populated) and the TUI's own preview call to
`installer.List()`. `internal/installer.Install`/`List` themselves —
the generic, already-tested primitives materializing
`kit.TemplatesFS` — are left untouched.

**Rationale**: FR-001 requires that `init` stop writing the unused
directory into a target project; it does not require deleting a
generic, already-shipped, already-tested library capability that has
no problem of its own and could plausibly serve a future caller (e.g.
a future "show template" or docs-generation command). Removing the
one real caller is the smallest change that satisfies FR-001; deleting
`internal/installer.Install`/`List` and their own dedicated tests
would be speculative cleanup beyond this feature's own scope
(Constitution Principle IV — YAGNI cuts both ways: don't build what
isn't needed, and don't delete what isn't causing a problem).

**Alternatives considered**: Writing the templates then deleting them
immediately after ("remove after initial setup," per spec.md's own
Assumptions, which explicitly allows either). Rejected — strictly more
work (a write, then a delete, both needing error handling) for the
identical end state; never writing them is simpler and is what
spec.md's own Assumptions already names as the expected, simpler
option.

## 2. Directory scaffolding lives in `internal/bootstrap`, as one new file

**Decision**: One new file, `internal/bootstrap/scaffold.go`, exporting
a single function conceptually:

```go
func scaffoldDirectories(targetDir string) ([]string, error)
```

Called from `Bootstrap` right after `writeDefaultConfig` (config
exists before the structure it describes is scaffolded) and before the
agent's own `Install`. It creates every directory
`project.Configuration`'s own already-existing `Default*` constants
name — `ArtifactsDir`, `RawDir`, `KnowledgeDir`, the parent directory
of `ConstitutionPath` (i.e. `ai/memory`), `LearningsDir`,
`ProgramsRoot` — via `os.MkdirAll` (idempotent, tolerant of a directory
that already exists), and writes a placeholder file (research.md #3)
into any of them that is genuinely empty afterward, never into one
that already has real content (FR-006).

**Rationale**: This is bootstrap-specific, one-time setup logic — the
same package that already owns `writeDefaultConfig` (also
bootstrap-specific, one-time) and already reads
`project.DefaultArtifactsDir`, etc. (config.go's own existing import).
No new package is justified for one small function (Principle IV).

**Alternatives considered**: A new `internal/scaffold` package.
Rejected — no second caller exists or is anticipated; this is squarely
"one more thing `misterspec init` sets up," exactly `internal/
bootstrap`'s own existing responsibility.

## 3. Placeholder file: a literal, empty `.gitkeep`

**Decision**: The placeholder file (FR-004) is named `.gitkeep`,
written as a zero-byte file, exactly as the user's own request named
it — a long-established, widely recognized convention (the file has no
special meaning to Git itself; its only purpose is being a real file
inside an otherwise-empty directory so Git — which never tracks empty
directories — has something to commit).

**Rationale**: Directly honors the user's own explicit naming choice;
zero-byte content avoids any risk of the file being misinterpreted as
real project content (a `.md`/`.yaml` file with actual bytes could
theoretically confuse a naive directory scan, though misterspec's own
artifact classifier already keys off frontmatter and filename
patterns a bare `.gitkeep` will never match).

**Alternatives considered**: A `README.md` placeholder explaining the
directory's purpose. Rejected — heavier than necessary for a marker
whose only job is "exist," and a `.md` file sitting in, say,
`ai/knowledge/` risks a future scan treating it as a real (if
malformed) Knowledge artifact — `.gitkeep` carries no such ambiguity.

## 4. `BootstrapOutcome` gains a field; the CLI JSON envelope's `templates` key is removed, not emptied

**Decision**: `BootstrapOutcome.TemplateOutcomes` is replaced with
`DirectoriesScaffolded []string` (the relative paths actually
created/verified). `internal/cli/init.go`'s own JSON success payload
drops the `"templates"` key entirely and adds a `"directories"` key
listing them; the TUI's Preview and Success screens are updated the
same way (research.md #6).

**Rationale**: An empty `"templates": []` array would be a leftover
signal of a capability that no longer exists — actively misleading
under Constitution Principle IX ("transparent... contracts"). This
project is still in its `v1.0.0-alpha` phase (pre-1.0 stability
guarantee), making this the correct moment to correct the shape rather
than preserve a now-meaningless field for compatibility with nothing.

**Alternatives considered**: Keeping `"templates": []` for
"backward compatibility." Rejected — there is no real external
consumer of this alpha-stage JSON shape yet to be compatible with, and
an empty array documenting a removed capability is worse than no key
at all.

## 5. TUI Preview/Success screens show the new directories, not the old templates

**Decision**: `previewContent` (currently `templates
[]resourceSummary`) becomes `directories []string`; `buildPreview`
computes it via the same `project.Default*` constants
`scaffoldDirectories` uses (research.md #2) rather than a filesystem
call, since nothing exists on disk yet at Preview time. `viewPreview`
and `viewSuccess` render the new field in place of the old "kit
templates"/"kit resources installed" lines.

**Rationale**: The Preview screen's own stated contract (010's own
research.md) is "shows exactly what a confirmed install would create"
— once `Bootstrap` no longer creates `templates/`, showing it in
Preview would be actively wrong, not merely outdated.

**Alternatives considered**: Leaving the Preview screen unchanged
(showing zero templates, silently). Rejected — the user would see a
confusing "kit templates: 0 files" line with no indication anything
about the tool changed; showing the new directories in the same slot
is directly informative instead.

## 6. Binary releases: a GitHub Actions workflow cross-compiling with plain `go build`, no new build tool

**Decision**: A new workflow, `.github/workflows/release.yml`,
triggered on push of a tag matching `v*` (and `workflow_dispatch` with
a tag input, so it can also be run retroactively for the already-tagged
`v1.0.0-alpha`). It builds `misterspec` for `linux/amd64`,
`linux/arm64`, `darwin/amd64`, `darwin/arm64`, and `windows/amd64`
using Go's own native cross-compilation (`GOOS`/`GOARCH`, no CGO — the
project's own SQLite dependency, `modernc.org/sqlite`, is already
pure-Go per 014-sqlite-index's own research, so no per-target C
toolchain is needed), and uploads each resulting binary directly as a
release asset via `softprops/action-gh-release`, named
`misterspec_<os>_<arch>` (`.exe` suffix on Windows).

**Rationale**: FR-007/FR-008 only require a working, source-identical
binary per OS to be downloadable — plain `go build` with `GOOS`/
`GOARCH` already does this natively, with zero new dependency
(Principle IV: no goreleaser or other packaging tool where the
standard toolchain already suffices). Raw binaries (not `.zip`/
`.tar.gz`) keep the download instructions to "download, `chmod +x`,
run" — one fewer step than "download, extract, then run."

**Alternatives considered**: `goreleaser`. Rejected — a real, capable
tool, but a new dependency/config surface for exactly the same outcome
plain `go build` in a 5-entry matrix already achieves; revisit only if
real packaging needs (checksums, install scripts, package-manager
manifests) emerge later. Compressed archives per platform. Rejected —
adds an extraction step to every install for negligible size savings
on a single static Go binary.

## 7. README and the docs site both link to the same Releases page, not a version-pinned URL

**Decision**: FR-010's "stable, discoverable" download location is
`https://github.com/mottamarcio/misterspec/releases/latest` — GitHub's
own permanent alias that always resolves to the most recent non-
prerelease release — linked from both `README.md`'s own Install
section and `site/getting-started.html`, alongside the existing
build-from-source instructions (FR-009), not replacing them.

**Rationale**: A `/releases/latest` link never needs updating by hand
as new versions ship, directly satisfying FR-010's own "without
requiring a user to already know the new version number in advance."
`v1.0.0-alpha` is a pre-release, so `/latest` will only start
resolving to it once a non-prerelease tag exists — noted as a known,
acceptable state during the alpha phase, not a defect of this design.

**Alternatives considered**: Linking to the specific `v1.0.0-alpha`
tag's own release page. Rejected — becomes stale the moment a new
version ships, requiring a manual doc update FR-010 explicitly rules
out needing.

## 8. This feature's git integration stops at `dev`; no `stg`/`main` promotion

**Decision**: This feature's own implementation and review happen
entirely on `dev` (via its own feature branch and PR, per the newly
established branch-protection workflow) — no PR to `stg` or `main` is
opened as part of this feature.

**Rationale**: Explicit, direct user instruction mid-session: validate
the resulting build on `dev` first before any further promotion.

**Alternatives considered**: N/A — direct instruction, not a design
tradeoff.
