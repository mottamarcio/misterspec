# Research: Referências a Seções com Âncoras Estáveis

## 1. Anchor declaration syntax in Markdown

**Decision**: Extend the existing ATX heading pattern
(`internal/artifacts/document.go`'s `headingPattern`) to recognize an
optional trailing explicit-anchor suffix, `{#slug}`, e.g.
`## Retry Policy {#retry-policy}`. The suffix is stripped from
`Section.Heading` (the displayed title stays clean) and captured into a
new `Section.Anchor` field. A heading with no `{#...}` suffix has
`Anchor == ""` — the overwhelming majority of headings today, and
exactly spec.md's "an anchor is opt-in, not required" assumption.

**Rationale**: This is an established, widely recognized Markdown
extension (Pandoc, kramdown, and several static-site generators use the
identical `{#id}` heading-suffix convention), so authors already have a
mental model for it, and it satisfies spec FR-002/FR-003 directly: the
identifier lives in the heading line itself, independent of the title
text before it, and editing the title text without touching `{#...}`
leaves the anchor unchanged.

**Alternatives considered**:
- *Derive the anchor from the heading text itself* (a GitHub-style
  slug): rejected — this is exactly what spec User Story 2 says must
  NOT happen; a title edit would silently change the anchor's identity.
- *A separate front-matter list mapping headings to anchor IDs*:
  rejected — disconnects the anchor from the heading it names, harder
  to keep in sync by eye in a long document, and duplicates information
  editors can put directly at the point of use with `{#...}`.

## 2. Wikilink anchor syntax

**Decision**: Extend `internal/artifacts/wikilink.go`'s
`extractLineLinks`/`splitTargetAlias` to recognize `TARGET#anchor` as
the pre-alias portion: `[[ID#anchor]]` and `[[ID#anchor|Alias]]`. Add a
new `WikiLink.Anchor string` field (empty when no `#` is present);
`WikiLink.Target` stays exactly the entity-ID token it is today — no
change to `ids.ResolveTarget`'s input shape or to any existing caller
that only reads `Target`.

**Rationale**: Keeps the change additive and localized to the one
lexical scanner that already owns target/alias splitting (`|`); adding
a second split (`#`) before the existing one is a small, pure addition
with no new parsing pass. Every existing `[[ID]]`/`[[ID|Alias]]`
wikilink contains no `#`, so `Anchor` stays `""` for 100% of them
(spec FR-008).

**Alternatives considered**: A distinct bracket syntax for anchored
links (e.g. `[[ID]]#anchor` outside the brackets) — rejected as a novel
syntax with no precedent in this project's own wikilink model or in
comparable tools, and harder to keep visually paired with its alias.

## 3. Where anchor existence/uniqueness gets validated

**Investigation**: `internal/validation/wikilinks.go`'s
`classifyWikilink` already resolves a link's `Target` via
`ids.ResolveTarget` and returns exactly one of `CodeInvalidWikilink`/
`CodeBrokenWikilink`/`CodeAmbiguousWikilink`/no-finding. There is no
existing per-artifact "declared-values-must-be-unique" check to model
anchor-uniqueness on; `internal/validation/findings.go` is a flat,
append-only enum of `Code*` string constants closed and read by
`internal/example/skills_content_test.go`'s capability allowlist
(039-lean-skills-integration-contracts) — any real new capability must
add a new constant there, not overload an existing one.

**Decision**: Add two new codes: `CodeUnknownAnchor` ("declared
target artifact exists, but the referenced anchor does not"), raised by
extending `classifyWikilink` with one more branch (only reached when
`matches == 1` and `link.Anchor != ""` — resolve the target artifact's
own `Document`/`Section`s and check the anchor exists), and
`CodeDuplicateAnchor` ("two Sections in the same artifact declare the
same explicit anchor"), raised by a new, separate per-artifact check
(`checkAnchors`, mirroring `checkWikilinks`'s own shape) that scans one
artifact's own `Section`s for a repeated non-empty `Anchor` value —
independent of whether anything currently references it (spec's
"unreferenced anchor is not itself an error" — duplication is the
error, not the anchor's existence).

**Alternatives considered**: Reusing `CodeBrokenWikilink` for a missing
anchor — rejected: spec FR-009/SC-003 explicitly requires a diagnostic
that distinguishes "target artifact doesn't exist" from "anchor doesn't
exist in an existing target," which collapsing the two codes would lose.

## 4. Retrieving only the referenced section (FR-006)

**Investigation**: `internal/context/collector.go`'s
`connectedCandidates` turns every outgoing/incoming reference into
candidates via `chunkArtifact(root, path, reason)`, which chunks the
*entire* target artifact (`artifacts.ChunksWithOffset`, one `Candidate`
per `Chunk`) — today, a wikilink to `KNOW-003` always makes every
`Chunk` of `KNOW-003` a candidate, regardless of which section (if any)
the author meant. This is the exact point where an anchor-qualified
wikilink must instead select only the one matching Chunk.

**Decision**: Add a new `chunkArtifactAnchor(root, path, anchor,
reason)` alongside `chunkArtifact`, used by `connectedCandidates`
whenever the driving `ReferenceEntry`/`BacklinkEntry` is a `"wikilink"`
relation carrying a non-empty anchor. It parses the same `Document` as
`chunkArtifact`, locates the `Section` whose `Anchor` matches, and
returns exactly one `Candidate` for that Section's own Chunk (Content
included even when the Section's own `Body` is empty — an anchor on a
heading immediately followed by a subheading is still a valid,
resolvable anchor, per spec Edge Cases; `artifacts.Chunks` today skips
empty-Body sections entirely, so this path must not reuse `Chunks`
verbatim, it must locate the Section directly), plus a `HeadingPath
[]string` field on `Candidate`/the context envelope's item shape — the
ordered titles of that Section's ancestor headings (Level 1..N-1 chain
by nesting, not sibling headings) — satisfying "mínimo de contexto
hierárquico necessário" without embedding the ancestors' own bodies
(FR-006/FR-007: a breadcrumb, never a content merge).

**Alternatives considered**: Returning the referenced Chunk plus its
full parent Chunk's Content — rejected: pulls in every sibling
subsection under the same parent too (since a parent Section's own Body
is only the text before its first child heading, but retrieving the
*chunk* users would expect as "context" would mean walking up
differently) and risks approaching whole-document size for a deeply
common structure, contradicting FR-006's "not the whole artifact."
A pure heading-title breadcrumb is the minimal, unambiguous
interpretation of "hierarchical context necessary to preserve meaning."

## 5. Schema/contract versioning

**Investigation**: `internal/context/index/schema.go`'s
`schemaVersion` is 2 (bumped from 1 by 038 for the `links` table's
`source_section`/`source_line` widening); `internal/cli/internalcmd/
context.go`'s `contextSchemaVersion` is 4 (bumped from 3 by 038 for
`--provenance`). Both already follow the "additive field bumps the
version" discipline this feature must continue.

**Decision**: `chunks` table gains an `anchor TEXT` column (nullable —
empty for a Chunk with no declared anchor); `links` table gains a
`target_anchor TEXT` column (nullable — empty for a non-anchor-
qualified reference). `schemaVersion` bumps 2→3, triggering the index's
already-existing version-mismatch rebuild path (no manual migration,
Constitution Principle III, same as 038's own precedent).
`contextSchemaVersion` bumps 4→5 for the new `heading_path`/`anchor`
fields appearing on context items and `Reason`. `references`/
`backlinks` command output gains the same additive, unversioned fields
038 already established (`target_anchor`), no version bump needed
there (they are unversioned envelopes by design).

## 6. Compatibility with existing whole-document links (FR-008)

**Decision**: No behavioral change for `WikiLink.Anchor == ""` anywhere
in the pipeline — `classifyWikilink`'s existing three-code
classification is unchanged for a non-anchor link;
`connectedCandidates` calls the existing `chunkArtifact` exactly as
before when there is no anchor. The two new code paths (validation's
anchor-existence check, the collector's `chunkArtifactAnchor`) are only
ever reached when `link.Anchor != ""`, making this feature purely
additive over 011/012/038's existing, already-tested behavior.
