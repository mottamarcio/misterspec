# Feature Specification: Disposable SQLite Index

**Feature Branch**: `014-sqlite-index`
**Created**: 2026-09-14
**Status**: Draft
**Input**: User description: "seguir com a sugestão 'phase 4: disposable
sqlite index'" (proceed with Phase 4 — Disposable SQLite Index — of
`docs/context-engine-implementation.md`, itself
`docs/architecture-specification.md`'s Phase 7 "Second Brain", building
on 011's wikilinks, 012's reference graph, and 013's document/chunk
model)

## User Scenarios & Testing *(mandatory)*

<!--
  As with 001-013, this feature's "users" are misterspec's own
  deterministic layer and future callers — Phase 5's incremental sync
  scheduling and Phase 6's context collector, rather than an end-user UI
  of its own. Per context-engine-implementation.md's own Phase 4
  scoping, this covers building, incrementally updating, searching, and
  rebuilding a local, disposable index of the project's own content — no
  ranking/budgeting, no context assembly, no internal CLI command yet
  (Phase 8's own later job).
-->

### User Story 1 - Build a Searchable Index From the Project's Own Content (Priority: P1)

Given a project's existing knowledge and artifact content, build a
local, disposable index of it, so a full-text search over that content
returns the matching pieces — without having to re-scan and re-parse
every artifact from scratch on every single question.

**Why this priority**: This is the foundational capability — nothing
else in this feature (incremental updates, relationship lookups) means
anything until a first, complete index actually exists and can be
searched.

**Independent Test**: Starting from a project with no index at all,
request a build; confirm every eligible artifact's content appears in
the index, and a search for a term known to appear in exactly one piece
of content returns that piece.

**Acceptance Scenarios**:

1. **Given** a project with several artifacts and no existing index,
   **When** a build is requested, **Then** every eligible artifact's
   content appears in the index.
2. **Given** a freshly built index, **When** a search is made for a
   term that appears in exactly one indexed piece of content, **Then**
   that piece is returned.
3. **Given** a project containing only source code files and no
   eligible knowledge/artifact content, **When** a build is requested,
   **Then** none of the source code is indexed.
4. **Given** the index is deleted entirely, **When** a build is
   requested again, **Then** a complete, correct index is produced from
   the project's current state, with nothing lost.

---

### User Story 2 - Keep the Index Cheaply Up to Date (Priority: P2)

Given an index that already reflects the project's prior state,
re-synchronize it after the project has changed — touching only
artifacts that are actually new, changed, or removed since the last
synchronization, leaving everything else untouched — so keeping the
index current after a small edit stays fast regardless of how large the
project has grown.

**Why this priority**: A search index nobody can afford to keep current
is not actually useful in practice — this is what makes User Story 1's
index a living, trustworthy reflection of the project rather than a
one-time snapshot. Depends on User Story 1's index already existing.

**Independent Test**: Build an index, then change one artifact, add a
new one, and delete another, leaving the rest of the project untouched;
re-synchronize and confirm only the changed, new, and deleted artifacts'
indexed content was affected — everything else is provably untouched.

**Acceptance Scenarios**:

1. **Given** an existing index and a project where nothing has changed
   since the last synchronization, **When** synchronization runs again,
   **Then** no artifact's indexed content is re-processed.
2. **Given** an existing index and one artifact whose content has
   changed, **When** synchronization runs, **Then** only that artifact's
   own indexed content is replaced with its current content.
3. **Given** an existing index and one brand-new artifact, **When**
   synchronization runs, **Then** the new artifact's content appears in
   the index.
4. **Given** an existing index and one previously-indexed artifact that
   has since been deleted, **When** synchronization runs, **Then** that
   artifact's indexed content is removed from the index.
5. **Given** a synchronization that is interrupted partway through,
   **When** the project or index is inspected afterward, **Then** no
   project artifact has been altered, and the index is left in either
   its state from before the interrupted run or a fully correct
   post-run state — never a partially-applied, inconsistent one.

---

### User Story 3 - Look Up an Artifact's Relationships From the Index (Priority: P3)

Given the index has been built, retrieve an artifact's already-known
outgoing and incoming relationships (012's own reference graph) directly
from the index, as a faster path — while the direct, index-independent
computation (012) remains fully correct and available whether or not an
index exists at all.

**Why this priority**: Purely a performance path over an already-proven
correctness guarantee (012) — valuable once the index exists (User
Story 1), but never a requirement for the graph itself to work.
Ordered last because nothing in this feature depends on it, and no
caller needs it yet (Phase 6's own context collector is the first real
consumer).

**Independent Test**: Build an index over a project with known formal
and semantic relationships between artifacts; look up one artifact's
relationships from the index and confirm the result matches exactly
what 012's own direct computation reports for the same artifact.

**Acceptance Scenarios**:

1. **Given** an indexed project with a known relationship between two
   artifacts, **When** that relationship is looked up from the index,
   **Then** it matches 012's own direct computation exactly.
2. **Given** an artifact with no relationships at all, **When** its
   relationships are looked up from the index, **Then** the result is a
   well-formed empty answer, not an error.

---

### Edge Cases

- The index does not exist yet at all (first-ever build) — handled the
  same as any other build, just with more to do; never an error
  condition of its own.
- An artifact that fails to be read or structured (malformed
  frontmatter, unreadable file) during synchronization — does not halt
  synchronization of the rest of the project; the problem is reported,
  not fatal to the whole run.
- The index reflects a completely different, unrelated, or corrupted
  prior state (for example, copied from elsewhere, or damaged) — a full
  rebuild recovers a correct index with zero loss of actual project
  information, since the index itself never held any authoritative data
  to lose.
- A search or relationship lookup is made against an index that has
  never been built at all — a well-defined, non-crashing result (an
  empty answer, or a clear signal that no index exists yet), never an
  unhandled failure.
- The same final project state reached two different ways (one full
  rebuild vs. an initial build plus several incremental
  synchronizations) — both must produce a search-equivalent index: the
  same terms return the same underlying content either way.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST build a local, disposable search index from
  the project's own knowledge and artifact content — never from
  arbitrary repository source code.
- **FR-002**: The index MUST be fully derivable from the project's own
  existing files at any time — nothing in it may ever become the
  authoritative source of any project fact; the project's own files
  remain that source, unconditionally.
- **FR-003**: System MUST support full-text search over indexed content,
  returning the matching pieces of content.
- **FR-004**: System MUST support incremental synchronization: given an
  existing index and the project's current state, only artifacts that
  are new, changed, or deleted since the last synchronization are
  re-processed — an unchanged artifact's indexed content is left
  untouched.
- **FR-005**: System MUST detect whether an artifact has changed using
  the same content-identity mechanism the project already uses
  elsewhere for this purpose — never a second, independently invented
  change-detection mechanism.
- **FR-006**: System MUST support discarding and fully rebuilding the
  index at any time, producing a complete, correct index reflecting the
  project's current state regardless of the index's prior condition.
- **FR-007**: A failed or interrupted synchronization MUST NOT corrupt
  any of the project's own authoritative artifacts, and MUST leave the
  index in either its state from before that run or a fully correct
  post-run state — never a partially-applied, inconsistent one.
- **FR-008**: System MUST be able to report, for an artifact already
  reflected in the index, both its outgoing and incoming relationships,
  sourced from the index.
- **FR-009**: The index-backed relationship lookup (FR-008) MUST report
  results consistent with this project's own direct, index-independent
  relationship computation, for the same artifact and project state.
- **FR-010**: An artifact that cannot be read or structured MUST NOT
  halt synchronization of the rest of the project — it is skipped and
  reported, never treated as a fatal failure of the whole run.
- **FR-011**: System MUST NOT modify any authoritative project artifact
  as a side effect of building, synchronizing, searching, or rebuilding
  the index — every one of these operations is strictly read-only with
  respect to the project's own files.
- **FR-012**: The index's storage location MUST be clearly separate from
  the project's own authoritative content, and safe to delete at any
  time without losing any actual project information.

### Key Entities

- **Indexed Document**: The index's own record of one artifact —
  identity (path, and entity ID where one exists), title, current
  content fingerprint, and when it was last indexed. Exists purely to
  let synchronization know what it has already processed; never
  authoritative over the artifact itself.
- **Indexed Chunk**: One already-produced piece of an artifact's content
  (013's own Chunk), stored so it can be found by a full-text search
  and traced back to its Indexed Document.
- **Indexed Link**: One already-computed relationship (012's own
  Reference/Backlink), stored so it can be looked up directly from the
  index rather than recomputed on the spot — always expected to agree
  with 012's own direct computation, never a competing or overriding
  source of truth about relationships.
- **Search Result**: One matching Indexed Chunk returned by a full-text
  query, identifying exactly where it came from.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A full build over a representative project indexes 100%
  of eligible artifacts, with zero source-code files ever included.
- **SC-002**: A full-text search for a term known to appear in exactly
  one indexed piece of content returns that piece, every time, across
  tested fixtures.
- **SC-003**: After a small, targeted change to one artifact, a
  subsequent synchronization updates only that artifact's own indexed
  content — every other already-indexed artifact is provably untouched.
- **SC-004**: Discarding the index entirely and rebuilding it from
  scratch always converges to a search-equivalent index — the same
  search terms return the same underlying content either way.
- **SC-005**: 100% of tested index-backed relationship lookups match
  this project's own direct, index-independent computation for the same
  artifacts.
- **SC-006**: Zero authoritative project files are ever modified by any
  index operation, verified across every tested scenario.

## Assumptions

- This feature is scoped to exactly
  `docs/context-engine-implementation.md`'s own Phase 4 ("Disposable
  SQLite Index"): building, incrementally synchronizing, searching, and
  rebuilding a local index, plus index-backed relationship lookups. It
  explicitly excludes that document's later phases — dedicated
  incremental-synchronization *scheduling* concerns (Phase 5, beyond
  the synchronization mechanism itself), the context collector and
  ranking (Phase 6-7), and the internal context command (Phase 8) —
  each remains its own, later feature, continuing this project's
  established one-phase-at-a-time discipline (every prior feature, 001
  through 013, has kept an equivalent boundary).
- The source document already commits, at the architecture level, to a
  specific local database technology with built-in full-text search for
  this index (`docs/context-engine-implementation.md` §10, chosen
  explicitly over alternatives for this exact workload) — this spec
  treats that as a settled architectural decision already made outside
  this feature, the same way this project already treats Go, Cobra, and
  Bubble Tea as frozen elsewhere, not as an open question for this
  feature's own scope.
- The index lives at a clearly non-authoritative, disposable cache
  location under this project's own tooling directory, is never
  committed to version control, and is always safe to delete — mirroring
  how this project already treats its own short-lived allocation lock
  (Constitution Principle III).
- What gets indexed follows `docs/context-engine-implementation.md`
  §12 precisely: broader than 011's/012's own five-entity-type scope
  (Program, Feature, Spec, Knowledge, Learning) — it also includes Plan,
  Tasks, Validation, and the Constitution, since all are genuine prose
  content a future retrieval capability will want, and 013's own
  chunking already operates on content and location alone, with no
  dependency on an entity ID existing.
- The index-backed relationship lookup (FR-008/FR-009) is an optional
  performance path only — 012's own direct, filesystem-computed
  reference graph remains fully correct and available whether or not an
  index exists at all (`docs/context-engine-implementation.md` §7.2's
  own explicit requirement, already honored by 012 and unaffected by
  this feature).
- This feature does not decide *when* synchronization runs (on every
  command, on a schedule, on demand before a search) — that operational
  question belongs to whichever later capability actually invokes it
  (Phase 8's own command, or a future dogfooding decision).
