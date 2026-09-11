# Feature Specification: Core Repository Foundation

**Feature Branch**: `001-core-foundation`
**Created**: 2026-09-11
**Status**: Draft
**Input**: User description: "podemos começar com o foundation do projeto" (let's start with the project's foundation)

## User Scenarios & Testing *(mandatory)*

<!--
  This foundation feature has no end-user UI. Its "users" are the coding
  agent and the deterministic operations/Skills that consume it (per
  docs/architecture-specification.md §3, §32, §68 Phase 1). Every story
  below is independently testable against that consumer, without any of
  the later CLI surface (`misterspec internal ...`, `misterspec init`)
  existing yet.
-->

### User Story 1 - Reliable Project Detection & Configuration (Priority: P1)

Any consumer (an agent, a Skill, or a later deterministic operation) can be
pointed at a directory — the project root itself, or any directory nested
inside it — and reliably learn whether it is inside an initialized
misterspec project, where that project's root is, and what its resolved
configuration is (artifact locations, knowledge/raw locations, memory and
constitution locations, programs root, and ID numbering width).

**Why this priority**: Every other capability in misterspec — every
deterministic operation, every Skill — first needs to know "am I in a
project, and where does everything live?" Without this, nothing else in the
system can be trusted to operate on the right files.

**Independent Test**: Can be fully tested by pointing detection at a set of
fixture directories (a valid project root, a nested subdirectory inside one,
an uninitialized directory, and a directory with a malformed configuration)
and verifying the correct project root, configuration values, or
not-initialized/invalid signal is returned for each — with no other
capability required.

**Acceptance Scenarios**:

1. **Given** a directory that is a valid, initialized misterspec project
   root, **When** detection runs from that directory, **Then** the project
   root and fully resolved configuration are returned.
2. **Given** a directory nested several levels below a valid project root,
   **When** detection runs from that nested directory, **Then** the same
   project root and configuration are returned as if run from the root.
3. **Given** a directory with no misterspec project anywhere in its
   ancestry, **When** detection runs, **Then** a distinct "not initialized"
   result is returned rather than an error that looks like any other
   failure.
4. **Given** a project whose configuration file exists but is malformed or
   missing a required field, **When** detection runs, **Then** a distinct
   "invalid configuration" result is returned, naming the problem, instead
   of silently falling back to defaults.

---

### User Story 2 - Canonical Path Resolution & Artifact Typing (Priority: P2)

Given an entity type and ID (and, where applicable, its parent chain), any
consumer can compute the single canonical directory and file path that
entity's artifact must occupy — and, given an arbitrary artifact file path,
can determine which entity type it represents. No two callers ever compute
a different path for the same entity, and no computed or supplied path can
ever fall outside the project.

**Why this priority**: This is the layer that guarantees "the LLM never
guesses a path" (a core project invariant). It depends only on User Story 1
(knowing the project root and artifacts directory) and can be fully
exercised before any metadata parsing or ID discovery exists.

**Independent Test**: Can be fully tested by feeding a fixed project root
and a table of (entity type, ID, parent) inputs and confirming the returned
canonical paths match the frozen filesystem layout exactly, plus feeding
path-traversal inputs (e.g. containing `..`) and confirming they are
rejected — independent of any other story.

**Acceptance Scenarios**:

1. **Given** a project root and a (type, ID, parent) tuple for each
   supported entity type, **When** the canonical path is requested,
   **Then** the returned directory and file path match the fixed filesystem
   layout exactly, every time.
2. **Given** an artifact file path inside the project, **When** it is
   classified, **Then** its entity type is correctly identified from its
   canonical location and/or declared type.
3. **Given** a path or ID input that would resolve outside the project
   root, **When** resolution is attempted, **Then** it is rejected rather
   than resolved.

---

### User Story 3 - Metadata Parsing & ID Discovery (Priority: P3)

Given an artifact file, any consumer can parse its structured metadata
(frontmatter fields such as id, type, status, parent, depends_on,
supersedes) into a typed result, with malformed frontmatter and missing
required fields reported as distinct, specific problems. Given an entity
type, any consumer can scan the project tree and get back every
syntactically valid existing ID of that type, with malformed or duplicate
IDs flagged rather than silently included or silently dropped.

**Why this priority**: This is the layer later ID allocation ("max existing
suffix + 1", never a stored counter) and structural validation depend on.
It builds on User Story 2's path/typing rules but can be tested on its own
against a fixture tree of artifact files.

**Independent Test**: Can be fully tested against a fixture directory
containing well-formed artifacts, one with missing/invalid frontmatter, and
two artifacts sharing a duplicate ID — verifying metadata parses correctly
where valid, errors are specific where invalid, and the duplicate is
flagged during a scan without the scan itself failing.

**Acceptance Scenarios**:

1. **Given** an artifact with well-formed YAML frontmatter, **When** its
   metadata is parsed, **Then** all declared fields are returned in
   structured form.
2. **Given** an artifact with missing or syntactically invalid frontmatter,
   **When** its metadata is parsed, **Then** a specific, distinguishable
   error is returned (missing vs. malformed vs. missing-required-field).
3. **Given** a project containing several valid IDs of one entity type,
   **When** that type is scanned, **Then** every valid ID is returned and
   the next available ID equals the highest existing numeric suffix plus
   one.
4. **Given** a project containing two artifacts of the same type with the
   same numeric ID, **When** that type is scanned, **Then** the duplicate
   is reported and the scan still completes for the rest of the project.
5. **Given** a project containing zero artifacts of a given entity type,
   **When** that type is scanned, **Then** an empty result is returned, not
   an error.

---

### Edge Cases

- Working directory nested many levels below the project root.
- No project markers found anywhere from the working directory up to the
  filesystem root.
- Configuration file present but not valid YAML, or valid YAML missing a
  required key.
- Artifact frontmatter block missing entirely, or present but not valid
  YAML.
- An ID with the correct prefix but wrong zero-padding width or a
  non-numeric suffix.
- Two artifacts of the same entity type sharing the same numeric ID.
- A path or ID input containing traversal segments (e.g. `..`) that would
  resolve outside the project root.
- An artifact whose declared `parent` ID does not exist anywhere in the
  project (flagged as discoverable data here; judging whether that's a
  hard structural error is a later Validation feature's concern).
- An entity type with zero existing artifacts, scanned for the first time.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST detect, from the current working directory or any
  directory nested inside a project, whether that directory tree is an
  initialized misterspec project.
- **FR-002**: System MUST report the detected project's root as an
  absolute, normalized path, identical regardless of which nested
  subdirectory detection was run from.
- **FR-003**: System MUST return a distinct "not initialized" result — never
  a generic error and never a false success — when no project is found.
- **FR-004**: System MUST load the project configuration and expose, at
  minimum: the artifacts directory, the knowledge/raw directories, the
  memory and constitution paths, the programs root, and the ID numbering
  width.
- **FR-005**: System MUST return a distinct "invalid configuration" result,
  naming the specific problem, when the configuration file is malformed or
  missing a required field — never silently substituting a default for a
  required field.
- **FR-006**: System MUST classify any given artifact file path by entity
  type (program, feature, spec, plan, tasks, validation, knowledge,
  learning, or constitution) using its canonical location and/or declared
  type.
- **FR-007**: System MUST compute the canonical directory and file path for
  any entity given its type, ID, and — where the type has one — its parent
  ID, following the project's fixed filesystem layout, deterministically
  and identically on every call.
- **FR-008**: System MUST reject any path or ID resolution whose computed
  result would fall outside the project root.
- **FR-009**: System MUST parse an artifact's frontmatter into structured
  metadata, distinguishing "artifact not found," "frontmatter not
  well-formed," and "required field missing" as separate, specifically
  reported conditions.
- **FR-010**: System MUST validate entity ID syntax per supported entity
  type (program, feature, spec, task, knowledge, learning), rejecting a
  wrong prefix, non-numeric suffix, or incorrect zero-padding width as
  invalid rather than accepting it.
- **FR-011**: System MUST scan the project tree and enumerate every
  syntactically valid, existing ID for a requested entity type.
- **FR-012**: System MUST detect and report duplicate IDs of the same
  entity type found during a scan, without the presence of a duplicate
  halting enumeration of the rest of the project.
- **FR-013**: System MUST determine the next available ID for an entity
  type by finding the highest existing valid numeric suffix and
  incrementing it — never by reading or writing a persisted counter.
- **FR-014**: System MUST return an empty result, not an error, when
  scanning an entity type that has zero existing artifacts.

### Key Entities

- **Project**: The root of a misterspec-managed repository; identified by
  its filesystem root, its resolved configuration, and its artifacts
  directory.
- **Configuration**: The resolved set of settings governing where
  artifacts, knowledge, memory, and programs live, plus ID formatting
  rules, for one project.
- **Entity ID**: A typed, prefixed, zero-padded identifier (e.g.
  `SPEC-014`) that uniquely names one artifact within its entity type.
- **Artifact**: Any canonical Markdown-with-frontmatter file representing a
  Program, Feature, Spec, Plan, Tasks, Validation, Knowledge item, Learning,
  or the Constitution.
- **Canonical Path**: The single, deterministic filesystem location an
  entity's artifact must occupy, derived from its type, ID, and parent
  chain.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Any consumer can determine, in a single check and with zero
  false positives or false negatives across directory nesting depths,
  whether it is running inside a valid misterspec project.
- **SC-002**: 100% of syntactically valid entity IDs present in a populated
  project are discovered during a scan of that type, and 100% of malformed
  or duplicate IDs encountered are flagged rather than silently accepted.
- **SC-003**: For any given entity type, ID, and parent, the canonical path
  computed is identical across repeated calls and across callers — with no
  dependency on stored or cached state.
- **SC-004**: Zero path or ID resolutions ever return a location outside
  the project root, across all boundary and traversal inputs exercised.
- **SC-005**: A project with zero artifacts of a given entity type reports
  "no entities found" rather than an error, for 100% of supported entity
  types.

## Assumptions

- The filesystem layout, entity ID formats, and artifact schemas frozen in
  `docs/architecture-specification.md` (§20–31, §60–61) are the contract
  this foundation implements; this spec does not renegotiate them.
- This is an enabling/infrastructure feature: it does not yet expose the
  `misterspec internal ...` command surface or `misterspec init` — later
  features build the CLI and TUI on top of this foundation.
- ID numbering width defaults to 3 digits, per `.misterspec/config.yaml`,
  unless a project's configuration overrides it.
- The consumers ("users") of this feature are, in priority order: (1) the
  coding agent and the Skills it runs, indirectly via (2) misterspec's own
  deterministic operations, which are the direct callers of everything
  specified here.
- Historical ID gaps (e.g. SPEC-001, SPEC-002, SPEC-004 existing) are not
  backfilled; the next allocated ID is always max-plus-one.
