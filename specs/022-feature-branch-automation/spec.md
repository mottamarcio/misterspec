# Feature Specification: Feature-Level Git Branch Automation

**Feature Branch**: `022-feature-branch-automation`
**Created**: 2026-09-15
**Status**: Draft
**Input**: User description: "estou validando aqui e parece estar funcionando quase 100%, só detectei um pequeno problema. ele não esta criando novos branches para 'feat' e criando tudo no branch que estiver selecionando. Acho importante adicionar essa criação de branch 'feat' para cada feature que for implementada (deixo para vc decidir se é melhor criar branch a nivel de 'feature' ou de 'spec')"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A new Feature starts on its own branch (Priority: P1)

A developer using misterspec to manage a Git-tracked project creates a
new Feature (`misterspec internal create feature --parent PRG-001`).
Today, that Feature — and everything created under it afterward — lands
on whatever branch happened to be checked out, silently mixing unrelated
work together. Instead, creating the Feature should put the developer on
a fresh, dedicated branch for that Feature's own work, isolated from
whatever else is in progress on the branch they started from.

**Why this priority**: This is the entire problem being reported —
without it, every Feature's work is at risk of being committed to the
wrong branch, which is exactly the gap validation surfaced.

**Independent Test**: Starting from a clean checkout of any branch,
create a new Feature; confirm a new branch now exists and is checked
out, and that the Feature's own artifact file is written there rather
than on the original branch.

**Acceptance Scenarios**:

1. **Given** a Git-tracked misterspec project checked out on branch
   `dev` with a clean working tree, **When** a developer creates a new
   Feature under an existing Program, **Then** a new branch dedicated
   to that Feature is created and checked out, and the Feature's
   artifact file exists on that new branch.
2. **Given** the same project, **When** the developer runs `git log
   dev` afterward, **Then** the new Feature's artifact commit does
   **not** appear in `dev`'s own history — it is isolated to the new
   branch.
3. **Given** a project directory that is not a Git repository, **When**
   a developer creates a new Feature, **Then** Feature creation
   succeeds exactly as it does today, and the response clearly states
   that branch creation was skipped rather than failing silently or
   loudly erroring out.

---

### User Story 2 - Work under a Feature stays on that Feature's branch (Priority: P2)

Once a Feature has its own branch, everything created underneath it —
Specs, and later Tasks and implementation — should continue to
accumulate on that same branch rather than each spawning a branch of
its own, so a single Feature's work reviews and merges as one coherent
unit.

**Why this priority**: Without this, User Story 1 alone would still
leave Specs scattered depending on whatever branch is active, which is
the same underlying problem one level down.

**Independent Test**: After completing User Story 1 (a Feature and its
branch exist), stay on that branch and create a Spec under the Feature;
confirm no additional branch is created and the Spec lands on the
Feature's own branch.

**Acceptance Scenarios**:

1. **Given** a Feature's own dedicated branch is currently checked out,
   **When** a developer creates a Spec under that Feature, **Then** no
   new branch is created, and the Spec's artifact file is added to the
   Feature's existing branch.
2. **Given** a developer has switched away from a Feature's own branch
   (deliberately or by mistake) back to `dev`, **When** they attempt to
   create a Spec under that Feature, **Then** the creation still
   succeeds (never blocked outright) but the response clearly flags
   that the current branch does not match the Feature's own branch, so
   the mismatch cannot go unnoticed.

---

### User Story 3 - Teams that manage branches themselves can opt out (Priority: P3)

Some teams already have their own branching conventions (trunk-based
development, a different naming scheme, branch creation handled by
another tool) and do not want misterspec creating branches on their
behalf. They can turn this behavior off entirely for their project.

**Why this priority**: Valuable for adoption breadth, but the tool is
still useful and correct without it — today's existing behavior is the
fallback, not a broken state.

**Independent Test**: Disable the setting in a project's own
configuration; create a Feature; confirm no branch is created and the
Feature still lands on whichever branch was already checked out.

**Acceptance Scenarios**:

1. **Given** a project whose configuration has automatic branch
   creation disabled, **When** a developer creates a new Feature,
   **Then** no new branch is created, and the Feature's artifact file
   is written to the currently checked-out branch — identical to
   today's behavior.

---

### Edge Cases

- What happens when the branch a new Feature would use already exists
  (for example, a previous Feature-creation attempt was interrupted
  after the branch was made but before the artifact file was written)?
  The existing branch is reused (checked out) rather than treated as a
  failure — Feature IDs are already guaranteed unique by the ID
  allocator, so a name collision only ever means "resuming."
- What happens to uncommitted changes already present in the working
  tree at the moment a new Feature is created? They must not be lost,
  discarded, or require the developer to stash them manually first —
  they carry forward onto the new branch, matching ordinary Git branch
  creation semantics.
- What happens if the working tree is in a detached-HEAD state when a
  Feature is created? Branch creation still succeeds; the new branch
  simply starts from whatever commit was checked out.
- What happens when creating a Program (the level above Feature) or a
  Knowledge/Learning artifact (which has no parent Feature at all)? No
  branch is created for these — only Feature creation triggers this
  behavior (see Assumptions).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When a new Feature is created inside a Git-tracked
  project, the system MUST create a new Git branch dedicated to that
  Feature and switch the working tree to it before the Feature's own
  artifact file is written.
- **FR-002**: The created branch's name MUST always include the new
  Feature's own entity ID, so branch names never collide with each
  other; it MAY additionally include a human-readable summary of the
  Feature, supplied at creation time, so the branch's purpose is
  visible at a glance without depending on that summary alone for
  uniqueness (amended — see Assumptions).
- **FR-003**: When the target project directory is not a Git repository,
  the system MUST NOT attempt branch creation and MUST NOT fail Feature
  creation on that account — Feature creation succeeds exactly as it
  does today, and the response clearly states that branch creation was
  skipped.
- **FR-004**: The system MUST NOT create a new branch when creating a
  Spec, Task, Knowledge, or Learning artifact — only Feature creation
  triggers branch creation.
- **FR-005**: When creating a Spec (or other artifact) whose parent
  Feature has its own dedicated branch, and the branch currently
  checked out is not that Feature's own branch, the system MUST include
  a clear notice of that mismatch in its response, without blocking the
  creation and without switching branches on the developer's behalf.
- **FR-006**: The system MUST let a project disable automatic branch
  creation via its own configuration; when disabled, Feature creation
  MUST behave exactly as it does today (writes to whatever branch is
  already checked out).
- **FR-007**: If the deterministically-derived branch name already
  exists, the system MUST check out the existing branch rather than
  failing.
- **FR-008**: Automatic branch creation MUST NOT discard, stash, or
  otherwise alter any uncommitted changes already present in the
  working tree when a new Feature is created.

### Key Entities

- **Feature**: The existing top-level unit of work under a Program;
  gains an implicit one-to-one association with a dedicated Git branch,
  derived from its own ID.
- **Branch automation setting**: A per-project configuration toggle
  controlling whether Feature creation automatically manages Git
  branches (on by default) or leaves branch management entirely to the
  developer (today's existing behavior).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After creating a Feature, 100% of that Feature's own
  artifact files (the Feature itself, its Specs, its Tasks) land on a
  single dedicated branch rather than the branch that was active before
  the Feature was created, whenever automatic branch creation is
  enabled.
- **SC-002**: Creating a Feature in a directory that is not a Git
  repository requires zero additional steps or error handling from the
  developer compared to today.
- **SC-003**: A developer can fully disable this behavior with a single
  configuration change, verified by creating a Feature afterward and
  confirming no branch was created.
- **SC-004**: Re-attempting Feature creation after an interruption that
  left the branch already created succeeds without any manual Git
  cleanup by the developer.

## Assumptions

- **Branch granularity is per-Feature, not per-Spec.** A Feature already
  groups every Spec created under it in the same directory subtree
  (`programs/*/features/*/specs/*`); branching at the Spec level would
  fragment one Feature's cohesive change across many branches for no
  benefit, where branching at the Feature level maps naturally onto "one
  branch reviewed and merged as one unit of work" — the same pattern
  this project's own development workflow already uses successfully.
- **Amended**: branch names are the Feature's own ID plus an optional,
  caller-supplied human-readable summary (e.g. `FEAT-007-user-auth`),
  not the bare ID alone — validation surfaced that a bare ID is hard to
  recognize at a glance in a branch list. The summary is cosmetic only:
  it never participates in uniqueness (the ID alone still guarantees
  that) and a branch is always found by its ID regardless of what
  summary text was used when it was created.
- Automatic branch creation is enabled by default, since it directly
  fixes the reported gap; opting out (User Story 3) is the escape hatch
  for teams with their own conventions, not the default.
- Branch creation reuses ordinary, non-destructive Git operations only
  (create-and-checkout, or checkout of an existing branch) — nothing in
  this feature ever deletes, force-pushes, or rewrites branch history.
