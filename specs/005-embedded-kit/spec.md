# Feature Specification: Embedded Kit and Resource Installer

**Feature Branch**: `005-embedded-kit`
**Created**: 2026-09-11
**Status**: Draft
**Input**: User description: "seguir com a sugestão da fase 3" (proceed
with the Phase 3 suggestion: the Embedded Kit — templates, skills,
`go:embed`, resource installer — per
`docs/architecture-specification.md` §33, §68)

## User Scenarios & Testing *(mandatory)*

<!--
  As with the prior four features, this has no end-user UI. Its "users"
  are the coding agent and misterspec's own future `misterspec init`
  command, which — per docs/architecture-specification.md §33 — must be
  able to retrieve every framework resource (templates, and later Skills
  and agent integrations) from inside the compiled binary, with zero
  network access required. Every story builds on 003-entity-creation's
  already-embedded artifact templates rather than starting over.
-->

### User Story 1 - Discover What the Embedded Kit Provides (Priority: P1)

Any consumer can list every resource the embedded kit currently contains
— at minimum, every artifact template and which entity/artifact type it
belongs to — without installing anything or touching the filesystem at
all.

**Why this priority**: This is the foundation everything else in this
feature builds on: before materializing anything anywhere, a consumer
(eventually `misterspec init`'s own preview step) needs to know what's
available. It's also the smallest possible slice — pure, side-effect-free
reading of what's compiled into the binary.

**Independent Test**: Can be fully tested by calling the discovery
capability against the compiled binary's embedded kit and confirming it
lists exactly the known template set, with no fixture project, no target
directory, and no filesystem writes involved at all.

**Acceptance Scenarios**:

1. **Given** the compiled binary's embedded kit, **When** its templates
   are listed, **Then** every embedded template is returned, named and
   classified by the entity/artifact type it belongs to.
2. **Given** the same listing, **When** it is requested a second time,
   **Then** the result is identical — discovery never depends on prior
   state or a previous installation having happened.

---

### User Story 2 - Install Embedded Resources Into a Target Directory (Priority: P2)

Given a target directory, any consumer can materialize the embedded kit's
resources onto disk as a set of atomic, per-file operations: each file
either arrives complete and correct, or isn't written at all — never
partially.

**Why this priority**: This is the actual "make the framework's resources
available in this project" capability a future `misterspec init` will
orchestrate. It depends on User Story 1 only in spirit (installing
requires knowing what to install), not in implementation — it can be
built and tested against the embedded kit directly.

**Independent Test**: Can be fully tested against a fresh temporary
directory (every file newly written), and against a directory where every
target file already exists (nothing overwritten unless explicitly
requested) — confirming file-by-file outcomes and byte-for-byte content
correctness in both cases, independent of User Story 1's discovery
capability.

**Acceptance Scenarios**:

1. **Given** an empty target directory, **When** the embedded kit is
   installed into it, **Then** every resource is written, byte-for-byte
   identical to its embedded source.
2. **Given** a target directory where every resource already exists,
   **When** installation is run again without requesting overwrite,
   **Then** every resource is reported as skipped and nothing on disk
   changes.
3. **Given** the same already-populated target directory, **When**
   installation is run with overwrite explicitly requested, **Then**
   every resource is replaced, still written atomically.
4. **Given** a target path that would resolve outside the intended
   project root, **When** installation is attempted, **Then** it is
   rejected before any file is written.

---

### User Story 3 - Templates Have One Source of Truth (Priority: P3)

The artifact templates `003-entity-creation`'s `Create`/`CreateArtifact`
already use are the same embedded resources this feature's kit exposes —
not a separate, independently maintained copy that could drift from it.

**Why this priority**: This is a consolidation, not new user-facing
capability — it closes the gap between 003-entity-creation's
feature-scoped template embedding and this feature's shared embedded kit
root, so the project has exactly one place templates live, matching
`docs/architecture-specification.md` §32's intended layout. It is
prioritized last because it's a refactor of already-correct, already-
tested behavior, not new risk.

**Independent Test**: Can be fully tested by re-running
003-entity-creation's own existing test suite unmodified after this
feature's consolidation and confirming every test still passes — proving
the artifact content produced by `Create`/`CreateArtifact` is unchanged.

**Acceptance Scenarios**:

1. **Given** 003-entity-creation's existing `Create`/`CreateArtifact`
   test suite, **When** it is run after this feature's changes, **Then**
   every test passes unmodified — the rendered artifact content is
   byte-for-byte identical to before.
2. **Given** the embedded kit's template listing (User Story 1), **When**
   it is compared against what `Create`/`CreateArtifact` actually render
   from, **Then** they are provably the same underlying resource, not two
   copies that happen to currently agree.

---

### Edge Cases

- Installing into a target directory that does not exist yet (it must be
  created as part of installation).
- Re-running installation when every target file already exists and
  overwrite is not requested — every resource must be reported as
  skipped, not silently ignored without any report at all.
- Re-running installation with overwrite requested — every file is
  replaced, still with no partially written result even under
  interruption.
- A target path — or an individual resource's computed destination —
  that would traverse outside the intended root.
- Requesting kit discovery with no target directory or project involved
  at all (a bare, project-independent query).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST embed the framework's artifact templates
  directly into the compiled binary, requiring no network access or
  external file to retrieve them at runtime.
- **FR-002**: System MUST allow a consumer to list every resource the
  embedded kit currently provides — at minimum, each template's name and
  the entity/artifact type it belongs to — without writing anything to
  disk.
- **FR-003**: System MUST materialize embedded resources into a target
  directory as one atomic operation per resource: a reader must never
  observe a partially written file.
- **FR-004**: System MUST NOT overwrite a file that already exists at an
  installation target unless the caller explicitly requests overwrite.
- **FR-005**: System MUST report, per resource, whether it was installed,
  skipped (already present, overwrite not requested), or failed — never a
  single pass/fail flag for the whole operation.
- **FR-006**: System MUST reject installing to, or a resource resolving
  to, a target location outside the intended project root, before any
  file is written.
- **FR-007**: A failed or interrupted installation MUST NOT leave any
  file partially written at its target path.
- **FR-008**: The embedded resource set MUST be the single source of
  truth for template content: `003-entity-creation`'s artifact-creation
  templates MUST be sourced from this same embedded set, not a separate,
  independently maintained copy.
- **FR-009**: Discovery (FR-002) MUST be self-describing from the
  embedded kit itself — a consumer MUST NOT need to already know the list
  of embedded resources in advance to enumerate them.

### Key Entities

- **Kit Resource**: One embedded item available for installation — its
  name, the entity/artifact type it belongs to, and its content.
- **Installation Outcome**: The per-resource result of an install
  operation — installed, skipped, or failed, with a specific reason where
  applicable.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The compiled binary can list and materialize every embedded
  resource with zero network requests, 100% of the time.
- **SC-002**: Re-running installation without requesting overwrite never
  modifies an already-present file — verified across every embedded
  resource.
- **SC-003**: Every installed file, once written, is byte-for-byte
  identical to its embedded source — no truncation, no partial content —
  verified across both fresh and already-populated target directories.
- **SC-004**: 100% of install attempts whose target would resolve outside
  the intended root are rejected before any write occurs.
- **SC-005**: `003-entity-creation`'s existing test suite passes
  unmodified after this feature's template consolidation — zero
  regression in artifact content produced by `Create`/`CreateArtifact`.

## Assumptions

- This feature's embedded resource set is **templates only** — the 8
  artifact templates already built in `003-entity-creation`, relocated
  under the shared `kit/` root
  `docs/architecture-specification.md` §32 lays out. Canonical Skill
  definitions (`kit/skills/`) and agent-specific integration resources
  (`kit/integrations/`) are explicitly **out of scope**: they will be
  embedded under this same `kit/` root once Phase 6 (canonical Skills)
  and Phase 4 (agent adapters) produce real content to embed. Building
  placeholder content now, before either exists, would be speculative
  scope (Constitution Principle IV, YAGNI).
- "Install" in this feature means the generic, agent-agnostic
  materialization primitive only. Choosing a specific coding agent,
  transforming a canonical Skill into that agent's format, and the full
  interactive `misterspec init` flow (inspect → collect decisions → plan
  → preview → confirm → apply → verify,
  `docs/architecture-specification.md` §36) remain out of scope — those
  are Phase 4 and Phase 5 respectively, building on top of this
  feature's primitive.
- This feature builds directly on `001-core-foundation`'s and
  `003-entity-creation`'s packages (path containment, atomic-write
  discipline); no new external dependency, still no CLI/JSON surface.
- The consumers ("users") of this feature are, in priority order: (1) the
  coding agent and the Skills it runs, indirectly via (2) misterspec's
  own future `misterspec init` command and deterministic-operations CLI
  layer — unchanged from the prior four features.
