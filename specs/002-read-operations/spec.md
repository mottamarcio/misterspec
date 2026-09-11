# Feature Specification: Read-Only Deterministic Operations

**Feature Branch**: `002-read-operations`
**Created**: 2026-09-11
**Status**: Draft
**Input**: User description: "considerando a sugestão acima" (proceed with
the read-only deterministic operations layer recommended after
001-core-foundation: `resolve`, `inspect`, `parent`, `children`,
`inventory`, `fingerprint`)

## User Scenarios & Testing *(mandatory)*

<!--
  Like 001-core-foundation, this feature has no end-user UI. Its "users"
  are the coding agent and misterspec's own Skills, which — per
  docs/architecture-specification.md §3, §9, §41-49 — must call these
  operations instead of guessing paths, inventing IDs, or fabricating
  hashes through LLM reasoning. Every story is independently testable
  against that consumer, on top of 001-core-foundation's Project/
  Configuration/EntityID/CanonicalPath/Metadata primitives, without any
  CLI/JSON surface existing yet (that is a later feature).
-->

### User Story 1 - Locate and Inspect Any Entity by ID (Priority: P1)

Given an entity's ID alone (e.g. "SPEC-014"), any consumer can find exactly
where that entity lives in the project and retrieve its full structured
metadata (status, parent, dependencies, supersedes) — without knowing, or
having to reconstruct, its filesystem location or ancestry chain first.

**Why this priority**: Nearly every canonical Skill's first deterministic
step is "resolve this ID" and/or "inspect this ID"
(`docs/architecture-specification.md` §41-49). Without this, an agent has
no reliable way to even begin acting on an entity a user or a prior Skill
referred to by ID.

**Independent Test**: Can be fully tested by populating a fixture project
with several entities (including two sharing the same ID under different
parents, and one Task ID that lives inside a shared `tasks.md`) and
confirming: a known ID resolves to its exact canonical path with complete
metadata; an unknown ID is reported as not-found; a duplicated ID is
reported as ambiguous rather than silently picking one.

**Acceptance Scenarios**:

1. **Given** a project containing a Spec with a known ID, **When** that ID
   is resolved, **Then** its exact canonical path is returned.
2. **Given** the same Spec ID, **When** it is inspected, **Then** its
   complete structured metadata (status, parent, depends_on, supersedes) is
   returned.
3. **Given** an ID with no matching artifact anywhere in the project,
   **When** it is resolved or inspected, **Then** a distinct "not found"
   result is returned.
4. **Given** two artifacts of the same type that were independently
   assigned the same ID number (e.g. by a race or manual edit), **When**
   that ID is resolved, **Then** a distinct "ambiguous" result naming both
   locations is returned — never a silent pick of one.
5. **Given** a Task ID declared as a heading inside a Spec's `tasks.md`,
   **When** that Task ID is resolved or inspected, **Then** the owning
   `tasks.md` and that specific heading are identified correctly.

---

### User Story 2 - Discover Structural Relationships (Priority: P2)

Given an entity, any consumer can discover its structural parent, and
given an entity, its direct structural children (optionally filtered to
one child type) — using only the project's fixed nesting rules, never
semantic judgment about what "belongs" to what.

**Why this priority**: Skills that create or review entities
(`/create-feature`, `/create-specs`, `/analyze`) need to see existing
siblings/children before acting, so they don't duplicate or miss coverage.
This builds directly on User Story 1's ability to inspect an entity's
declared parent, and is independently testable without it.

**Independent Test**: Can be fully tested against a fixture project with a
Program containing two Features, one of which contains two Specs —
confirming `parent` and `children` (with and without a type filter) each
return exactly the expected set, and that a leaf or parent-less entity
reports its case distinctly rather than erroring.

**Acceptance Scenarios**:

1. **Given** a Spec with a declared parent Feature, **When** its parent is
   requested, **Then** that Feature's ID and type are returned.
2. **Given** a Program (which has no parent), **When** its parent is
   requested, **Then** a distinct "no parent" result is returned, not an
   error.
3. **Given** a Program containing two Features, **When** its children are
   requested, **Then** exactly those two Features are returned, and no
   unrelated entity.
4. **Given** the same Program, **When** its children are requested filtered
   to a type it has none of, **Then** an empty list is returned, not an
   error.

---

### User Story 3 - Discover Files and Verify Content Integrity (Priority: P3)

Any consumer can list the files present in a named project area (raw
sources, knowledge, or a program/feature/spec's own directory) with each
file's path, extension, and size — and compute a deterministic content
fingerprint for any file in the project, of any type, without the system
interpreting or understanding that file's contents.

**Why this priority**: This is the discovery-and-integrity primitive
`/create-knowledge-base` and future staleness-detection depend on
(`docs/architecture-specification.md` §13-14, §62-63). It has no dependency
on User Story 1 or 2 — it operates on plain filesystem areas and paths.

**Independent Test**: Can be fully tested against a fixture directory with
a mix of file types (Markdown, a non-UTF8/binary stand-in for a PDF, an
empty subdirectory) — confirming inventory lists every file with correct
metadata and an empty area returns an empty list, and that fingerprinting
the same file twice always yields the identical digest.

**Acceptance Scenarios**:

1. **Given** a project area containing several files, **When** it is
   inventoried, **Then** every file's relative path, extension, and size
   are returned.
2. **Given** a project area that exists but is empty, or does not exist
   yet, **When** it is inventoried, **Then** an empty list is returned, not
   an error.
3. **Given** any file in the project, **When** its fingerprint is
   requested, **Then** a SHA-256 digest is returned, identical across
   repeated requests for the same content.
4. **Given** a path that does not exist, or that would resolve outside the
   project root, **When** a fingerprint or inventory is requested for it,
   **Then** a distinct, specific error is returned.

---

### Edge Cases

- Two artifacts of the same type independently claiming the same ID number
  under different parents (the "ambiguous" case, distinct from "not
  found").
- Resolving or inspecting a Task ID, which — unlike every other entity
  type — lives as a Markdown heading inside a shared `tasks.md` rather than
  its own file or directory.
- Requesting the parent of an entity type that structurally has none
  (Program).
- Requesting children filtered to a type that is not a valid structural
  child of the given parent type.
- Inventorying an area before any file has ever been placed there.
- Fingerprinting a large file, a binary file, or a symlink.
- Resolving/inspecting a syntactically invalid ID (distinct from a
  syntactically valid but nonexistent one).
- A fingerprint or inventory request for a path outside the project root.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST resolve an entity ID to its exact canonical
  artifact path by searching the project's structure — never by the agent
  guessing, constructing, or remembering that path.
- **FR-002**: System MUST return a distinct "not found" result when no
  artifact anywhere in the project matches the requested ID.
- **FR-003**: System MUST return a distinct "ambiguous" result — naming
  every matching location — when more than one artifact claims the same
  ID, rather than silently resolving to one of them.
- **FR-004**: System MUST retrieve an artifact's complete structured
  metadata (id, type, status, parent, depends_on, supersedes, as
  applicable to its type) given either its ID or its already-resolved
  path.
- **FR-005**: System MUST report a specific, distinguishable reason (not
  found, frontmatter malformed, required field missing) when metadata
  cannot be retrieved, never a single generic failure.
- **FR-006**: System MUST determine an entity's structural parent from its
  own declared metadata, resolved to that parent's canonical location.
- **FR-007**: System MUST return a distinct "no parent" result — not an
  error — for an entity type that structurally has none.
- **FR-008**: System MUST enumerate an entity's direct structural children,
  optionally filtered to one child entity type, based on the project's
  fixed nesting rules.
- **FR-009**: System MUST return an empty list — not an error — when an
  entity has no children of the requested type.
- **FR-010**: System MUST list the files present in a named project area
  (at minimum: raw sources, knowledge, and any given entity's own
  directory), reporting each file's path relative to the project root, its
  extension, and its size in bytes.
- **FR-011**: System MUST return an empty list — not an error — when
  inventorying an area that is empty or does not yet exist, as long as
  that area's location is itself valid.
- **FR-012**: System MUST compute a SHA-256 content fingerprint for any
  file in the project, regardless of file type, without interpreting or
  parsing that file's contents.
- **FR-013**: System MUST report a distinct, specific error — never a
  silent empty/zero result — when asked to inventory or fingerprint a path
  that does not exist or would resolve outside the project root.
- **FR-014**: Every operation in this feature MUST be read-only: none may
  create, modify, or delete any file, anywhere, under any circumstance.
- **FR-015**: Every result MUST distinguish "the operation completed
  successfully" from "what it found" — a successful inventory of an empty
  area, or a successful children query with zero results, is never
  reported as a failure.
- **FR-016**: System MUST resolve and inspect a Task ID correctly despite
  Task entities having no independent file or directory of their own —
  identifying the specific Spec's `tasks.md` and heading that declares it.

### Key Entities

- **Resolved Location**: The outcome of resolving an ID — either exactly
  one canonical path, a distinct not-found result, or a distinct ambiguous
  result naming every matching path.
- **Structural Relationship**: A parent or children edge between two
  entities, derived only from declared metadata and fixed canonical
  nesting — never a semantic judgment about what "belongs" together.
- **File Inventory Entry**: One file found while inventorying an area —
  its path (relative to project root), extension, and size.
- **Fingerprint**: A file's algorithm-tagged content digest (e.g.
  `sha256:...`), computed independently of that file's type or whether
  misterspec understands its format.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For any existing entity, resolving and then inspecting its
  ID returns its exact canonical path and complete metadata in two direct
  calls, with zero occurrences of a wrong or guessed path across all
  entity types.
- **SC-002**: Every artifact pair sharing a duplicated ID is reported as
  ambiguous, with 0% silently resolved to an arbitrary one of them.
- **SC-003**: For any Program or Feature, a children query returns exactly
  its full set of direct structural children — 100% of true children
  present, 0% of unrelated entities included.
- **SC-004**: Inventorying an empty or not-yet-created (but validly
  located) area returns an empty list every time, never an error.
- **SC-005**: The same file content always produces an identical
  fingerprint across repeated requests, independent of when or how many
  times it is requested.
- **SC-006**: Across all six operations in this feature, zero files are
  ever created, modified, or deleted as a side effect of any read
  operation.

## Assumptions

- This feature builds directly on `001-core-foundation`'s
  `internal/project`, `internal/artifacts`, and `internal/ids` packages;
  it introduces no new external dependency beyond the Go standard library
  (`crypto/sha256` for fingerprinting).
- Still no CLI or JSON output surface — this spec covers the internal Go
  operations layer only. Wrapping these operations as
  `misterspec internal ...` machine-readable commands
  (`docs/architecture-specification.md` §6-8) is a distinct, later
  feature.
- "Children" reflects only the fixed structural nesting frozen in
  `docs/architecture-specification.md` §20 (Program→Feature→Spec, and
  Knowledge/Learning as flat siblings under their own directories); it
  does not compute or infer any semantic relationship.
- Task entities are handled as a special case throughout this feature
  (resolve/inspect/parent), since — unlike every other entity type — they
  are declared as headings inside a shared `tasks.md` rather than having
  an independent file or directory (per 001-core-foundation's
  `ids.Scan` design for Task).
- The consumers ("users") of this feature are, in priority order: (1) the
  coding agent and the Skills it runs, indirectly via (2) misterspec's own
  future deterministic-operations CLI layer, which is this feature's
  direct caller.
