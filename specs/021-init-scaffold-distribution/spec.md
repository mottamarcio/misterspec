# Feature Specification: Init Scaffolding and Binary Distribution

**Feature Branch**: `021-init-scaffold-distribution`
**Created**: 2026-09-15
**Status**: Draft
**Input**: User description: "precisamos fazer algumas alterações
(talvez considerando um versão v1.0.1). Primeiramente que o usuário
não precisa fazer o clone de toda a pasta do github. Nós podemos
gerar uma build e colocar o link para o usuario baixar o binário onde
quiser. E quando o usuário der o comando 'misterspec init' e escolher
o modelo de agent, gostaria que criasse automaticamente (alem da pasta
de skills), as hierarquias de pastas tbm com o arquivo '.gitkeep'
dentro. Reparei que tem uma pasta 'templates' sendo baixada, mas acho
que essa não precisa (se realmente não precisar, pode remover após o
setup inicial e configuração da hierarquia de pastas)" (three related
onboarding/distribution improvements for a v1.0.1 patch release, based
directly on the current shipped `misterspec init` behavior)

## User Scenarios & Testing *(mandatory)*

<!--
  This feature's users are people adopting misterspec for the first
  time — the exact same audience 020-docs-site-readme's Getting
  Started page addresses. All three stories target friction the user
  personally observed in that real, current onboarding path: cloning
  a full source repository just to get a binary, an initialized
  project's folder structure not being immediately visible/trackable
  in git, and an unused artifact appearing in every newly initialized
  project.
-->

### User Story 1 - Initialized Projects Contain No Unused Files (Priority: P1)

When a project owner runs `misterspec init`, the resulting project
contains only files that are actually used — nothing installed "just
in case," and nothing left over that the tool itself never reads
again.

**Why this priority**: This is a correctness fix, not a new
capability — the current, real behavior installs a `templates/`
directory into every initialized project whose own content the
`misterspec` binary never reads back from disk (verified: artifact
rendering reads its own copy embedded directly in the binary, never
the project's own copy). It is pure clutter today, with zero
functional dependency on it, making it the safest and most
independently verifiable of these three changes.

**Independent Test**: Initialize a fresh project and inspect its
resulting file tree; confirm no `templates/` directory (or equivalent
unused artifact) is present, and confirm every other misterspec
capability (creating a new entity, rendering a new artifact) still
works exactly as before.

**Acceptance Scenarios**:

1. **Given** a fresh, empty target directory, **When** `misterspec
   init` completes, **Then** the resulting project contains no
   `templates/` directory or any other file that misterspec itself
   never subsequently reads.
2. **Given** that same freshly initialized project, **When** a new
   entity or subordinate artifact is created afterward, **Then** it is
   scaffolded correctly and identically to how it was before this
   change — removing the unused directory changes nothing about actual
   artifact creation.

---

### User Story 2 - A New Project's Full Folder Structure Is Immediately Visible (Priority: P2)

When a project owner runs `misterspec init` and selects a coding
agent, the complete directory hierarchy the project's own
configuration expects — not only the chosen agent's Skills directory —
is created immediately, with each otherwise-empty directory tracked by
version control from the very first commit.

**Why this priority**: Today, a directory like the project's own
Knowledge or Learnings location only appears once the first artifact
of that kind is actually created — a new project owner has no visual
confirmation of the structure they're working within, and an empty
directory git normally refuses to track can quietly vanish from a
first commit. This depends on nothing from User Story 1, but is a
larger, more visible change to the initialized project's own shape.

**Independent Test**: Initialize a fresh project and inspect its
resulting file tree before creating any artifact; confirm every
directory the project's own configuration names already exists and is
committable to version control as-is.

**Acceptance Scenarios**:

1. **Given** a fresh, empty target directory, **When** `misterspec
   init` completes, **Then** every directory the project's own
   configuration names (raw sources, Knowledge, memory/Learnings,
   Programs) already exists on disk.
2. **Given** that same freshly initialized project, **When** its file
   tree is checked for files a version control system would actually
   track, **Then** every one of those otherwise-empty directories
   contains a placeholder file, so the empty structure survives being
   committed.
3. **Given** that same freshly initialized project, **When** the first
   real artifact of a given kind is later created inside one of these
   directories, **Then** it is created normally, unaffected by the
   placeholder file already present.

---

### User Story 3 - Get the Binary Without Cloning the Repository (Priority: P3)

Someone who wants to use misterspec can obtain a working binary for
their own operating system directly, without cloning the source
repository or having a Go toolchain installed.

**Why this priority**: This removes the single biggest barrier named
in the user's own request — today the documented install path
requires `git clone` plus a local Go build. It is independent of User
Story 1 and 2 (a packaging/distribution concern, not a change to
`init`'s own behavior) and is prioritized last only because it depends
on a release process existing at all, which the other two stories
don't.

**Independent Test**: Without cloning the repository or installing a
Go toolchain, follow only the published download instructions for a
given operating system and confirm a working `misterspec` binary
results, capable of running `misterspec init` successfully.

**Acceptance Scenarios**:

1. **Given** someone with no local clone of the repository and no Go
   toolchain, **When** they follow the published binary download
   instructions for their own operating system, **Then** they obtain a
   working `misterspec` binary.
2. **Given** that downloaded binary, **When** `misterspec init` is run
   with it, **Then** it behaves identically to a binary built from
   source at the same version.
3. **Given** a new version is released later, **When** someone
   downloads the binary again, **Then** the download instructions and
   link continue to work without requiring them to know or guess a new
   URL structure.

---

### Edge Cases

- A directory the project's own configuration names that already
  exists (for example, a project re-initialized after a partial prior
  attempt) — its own existing content is left untouched; only a
  missing placeholder is added if the directory is otherwise empty.
- A directory that is not empty (already contains a real artifact) —
  no placeholder file is added; only genuinely empty directories need
  one.
- A user who already has a local clone and prefers building from
  source — that path continues to work unchanged; this feature adds a
  second option, it does not remove the first.
- An operating system or architecture combination the published
  binaries don't cover — falls back to the existing build-from-source
  instructions, which remain valid and documented.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `misterspec init` MUST NOT install any file into the
  target project that misterspec itself never reads back from that
  project afterward.
- **FR-002**: Artifact creation and rendering (new entities, new
  subordinate artifacts) MUST continue to work identically after
  FR-001's own removal — no functional behavior may depend on the
  removed file(s).
- **FR-003**: `misterspec init` MUST create every directory the
  project's own configuration names (raw sources, Knowledge,
  memory/Learnings, Programs — beyond the chosen agent's own Skills
  location), regardless of whether any artifact exists in it yet.
- **FR-004**: Every directory created by FR-003 that is otherwise empty
  MUST contain a placeholder file so it can be tracked by version
  control before any real artifact exists in it.
- **FR-005**: Creating a real artifact inside a directory that already
  contains only a placeholder file (FR-004) MUST succeed normally,
  exactly as it would in a directory with no placeholder at all.
- **FR-006**: Re-running `init`-related scaffolding against a directory
  that already exists MUST NOT remove, overwrite, or otherwise disturb
  any real content already present in it.
- **FR-007**: A pre-built `misterspec` binary MUST be published at a
  stable, discoverable location for each of the project's supported
  operating systems, without requiring the recipient to clone the
  source repository or install a build toolchain.
- **FR-008**: A downloaded pre-built binary MUST behave identically to
  one built from source at the same version.
- **FR-009**: The project's own installation documentation (README and
  documentation site) MUST present the pre-built binary download as an
  install option, without removing the existing build-from-source
  instructions.
- **FR-010**: The binary download location's own URL structure MUST
  remain stable and discoverable across future releases, without
  requiring a user to already know the new version number in advance.

### Key Entities

- **Initialized Project Structure**: Everything `misterspec init`
  creates in a target directory — the project configuration, the
  chosen agent's Skills, and now the full set of artifact-bearing
  directories the configuration names, each committable to version
  control from the first commit.
- **Placeholder File**: The marker (FR-004) that lets an otherwise-
  empty, meaningful directory be tracked by version control; never a
  real artifact, and never an obstacle to creating one later (FR-005).
- **Published Binary**: A pre-built `misterspec` executable for a
  given operating system, available for direct download at a stable
  location (FR-007), behaviorally identical to a source build (FR-008).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of freshly initialized projects contain zero files
  that misterspec itself never reads back afterward.
- **SC-002**: 100% of freshly initialized projects have every one of
  the project's own configured directories already present and
  committable to version control, before any artifact is created.
- **SC-003**: 100% of subsequent artifact-creation operations inside a
  placeholder-only directory succeed on the first attempt, with no
  behavior different from an equivalent directory with no placeholder.
- **SC-004**: A person with no existing clone of the repository and no
  Go toolchain can obtain a working binary and successfully run
  `misterspec init` in under 5 minutes, following only the published
  download instructions.
- **SC-005**: 0 regressions in existing build-from-source or
  already-initialized-project behavior, verified across the existing
  test suite plus this feature's own new coverage.

## Assumptions

- "The `templates` folder being downloaded" (User Story 1) refers to
  the already-shipped, real behavior where `misterspec init`
  materializes the framework's own artifact templates into the target
  project's `templates/` directory — verified, by reading the current
  source, to be entirely unused afterward: artifact rendering already
  reads its own copy embedded directly in the compiled binary, never a
  project's own on-disk copy. This specification requires only that
  the unused file(s) no longer appear in a newly initialized project
  (FR-001) — whether that is achieved by never writing them at all, or
  by writing and then removing them, is an implementation detail left
  to planning; not writing them at all is the simpler of the two and
  is expected to be preferred, but is not itself a requirement.
- "Hierarquia de pastas" (User Story 2) means the directories this
  project's own configuration already names as meaningful locations
  (raw sources, Knowledge, memory/Learnings, Programs) — not an
  arbitrary or expanded set invented for this feature. The exact
  placeholder file name/format (the user's own example: `.gitkeep`) is
  carried forward as the user's explicit choice, though any
  filename/format that is not itself a Markdown artifact or otherwise
  misclassified by misterspec's own artifact scanner is an acceptable
  implementation of the same requirement.
- "Gerar uma build e colocar o link" (User Story 3) is read as
  publishing pre-built binaries through this project's own existing
  distribution channel (GitHub Releases, already used for the
  `v1.0.0-alpha` tag) rather than standing up a separate hosting
  service — the exact set of operating systems/architectures covered,
  and the exact automation building them, are implementation details
  left to planning.
- This feature is scoped to exactly these three onboarding/
  distribution improvements — it does not change the Context Engine,
  Skill content, or any other already-shipped capability's own
  behavior beyond what FR-001–FR-006 describe for `init` itself.
- This feature's own three stories are independent enough to ship
  separately if needed (spec.md's own per-story "Independent Test"),
  but are bundled here as one v1.0.1-flavored patch release per the
  user's own framing, continuing this project's practice (e.g. 018) of
  bundling closely-related onboarding concerns raised together.
