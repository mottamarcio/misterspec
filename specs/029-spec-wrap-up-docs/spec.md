# Feature Specification: `/mister-wrap-up` — Spec Documentation for Future Official Docs

**Feature Branch**: `029-spec-wrap-up-docs`
**Created**: 2026-09-18
**Status**: Draft
**Input**: User description: "ao final do desenvolvimento das SPECs, é possivel dar o comando (opcional) 'mister-analyze SPEC-00X' para checar se está tudo correto. Gostaria de adicionar um novo comando opcional ao final da implementação da SPEC chamado '/mister-wrap-up SPEC-00X' para gerar a documentação daquela spec (unindo informações dos commits, do knowledge base e algum outro arquivo que for pertinente) e salvar o documento (escrito em markdown) dentro de uma pasta chamada 'cortex' na raiz do projeto. Esses arquivos markdown serão usados para gerar a documentação oficial do projeto (ou mesmo um gitbook) futuramente. (Clarified with the user: commit range is every commit on the Spec's own Feature branch from the commit that first added the Spec's `plan.md`, through the current HEAD — no commit-message convention required. 'Other pertinent files' means the Spec's own artifacts (`spec.md`, `plan.md`, `tasks.md`, `validation.md` if it exists), the project Constitution, and any related Learning artifacts. Output location is one file per Spec: `cortex/SPEC-###-<slug>.md`.)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Generate a Spec's own documentation after implementation (Priority: P1)

A developer has finished implementing a Spec (optionally already run `/mister-analyze` to verify it) and wants a single, readable Markdown document capturing what that Spec was, why it existed, what was actually built, and what happened along the way — suitable for later compiling into the project's official documentation or a GitBook, without the developer having to manually stitch together commit history, Knowledge, and the Spec's own artifacts by hand.

**Why this priority**: This is the entire feature — without it, there is no wrap-up document at all, and everything else (content sourcing, file naming, idempotent re-runs) only matters in service of producing this one artifact well.

**Independent Test**: Run `/mister-wrap-up SPEC-014` against a Spec whose implementation is complete (Plan, Tasks all attempted, optionally a Validation artifact); confirm a new Markdown file appears under `cortex/` and reads as a coherent, self-contained account of that Spec — not a raw dump of its inputs.

**Acceptance Scenarios**:

1. **Given** a Spec with a complete Plan, Tasks, and Validation, **When** a developer runs `/mister-wrap-up SPEC-014`, **Then** a new Markdown file is written under `cortex/` (named `SPEC-014-<slug>.md`) summarizing the Spec's purpose, what was implemented, and its outcome, drawing on the Spec's own artifacts, the commits made for it, the project's Knowledge base, the Constitution, and any related Learnings.
2. **Given** the same Spec, **When** the developer inspects the generated document, **Then** it reads as prose suitable for an eventual official documentation page or GitBook chapter — not a bullet-by-bullet restatement of `tasks.md`, not raw commit log output pasted verbatim.
3. **Given** a Spec whose implementation is only partially complete (some Tasks not yet attempted), **When** the developer runs `/mister-wrap-up` anyway, **Then** the command still produces a document, but clearly states that implementation is incomplete and which parts are still outstanding, rather than presenting the Spec as finished.

---

### User Story 2 - Commit history sourced correctly, without any new commit-message convention (Priority: P1)

A developer has been committing normally throughout a Spec's implementation, using whatever commit message style they already use — no special tag or reference to the Spec ID required.

**Why this priority**: Equal in importance to User Story 1 — if commit sourcing required a new authoring convention, every commit made before this feature existed (i.e. all prior history) would be silently excluded, and the feature would only work going forward under a discipline nothing enforces.

**Independent Test**: On a Feature branch with several commits made both before and after a Spec's `plan.md` was first added, run `/mister-wrap-up` for that Spec and confirm the generated document's commit-derived content reflects only commits from `plan.md`'s first appearance onward, regardless of what any commit message says.

**Acceptance Scenarios**:

1. **Given** a Feature branch with commits made before the Spec's `plan.md` existed and commits made after, **When** `/mister-wrap-up` runs, **Then** only the commits from the one that first added `plan.md` through the current state are considered part of this Spec's own history.
2. **Given** a Spec whose `plan.md` was only just created (no implementation commits yet), **When** `/mister-wrap-up` runs, **Then** the document reflects that little or nothing has been committed yet, rather than erroring or fabricating implementation history.

---

### User Story 3 - Re-running the command keeps the document current (Priority: P2)

A developer runs `/mister-wrap-up` for a Spec, continues working on it (more commits, an amended Plan, a later Validation), and runs the command again.

**Why this priority**: Lower priority than Stories 1–2 because a first-run-only version of this feature would still deliver the core value; but a wrap-up command that can't be safely re-run as a Spec evolves is a real usability gap for the "optional command developers run whenever they want a fresh snapshot" use case the user described.

**Independent Test**: Run `/mister-wrap-up` for a Spec, make further commits and re-run it; confirm the document is regenerated to reflect the current state rather than either failing on an already-existing file or silently leaving stale content.

**Acceptance Scenarios**:

1. **Given** a Spec whose wrap-up document already exists from an earlier run, **When** the developer runs `/mister-wrap-up` again after further commits, **Then** the existing document is replaced with a fresh one reflecting the Spec's current state — not duplicated, not left stale.

### Edge Cases

- What happens when the named Spec has no `plan.md` yet? The command cannot determine a commit range or draw on Plan content; it must stop and recommend generating a Plan first, rather than guessing a starting point.
- What happens when the project is not a Git repository, or the Spec's own commit history can't be determined (e.g. `plan.md` was never committed)? The document is still generated from the Spec's own artifacts, Knowledge, Constitution, and Learnings, but clearly notes that commit history could not be included, rather than failing outright.
- What happens when a Spec has no related Knowledge, Constitution content, or Learnings at all? Those sections are omitted or noted as "none found" rather than fabricated.
- What happens when `cortex/` doesn't exist yet in the project? It is created automatically on first use.
- What happens when two different Specs would produce the same slug (e.g. very similar titles)? The Spec ID prefix (`SPEC-014-...` vs `SPEC-021-...`) already guarantees the filename itself never collides, even if the human-readable slug portion happens to match.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide an optional command, `/mister-wrap-up SPEC-###`, usable at any point after a Spec has a Plan, independent of whether `/mister-analyze` has been run.
- **FR-002**: The generated document MUST be written as a single Markdown file under a `cortex/` directory at the project root, named with the Spec's own ID as a prefix (`SPEC-###-<slug>.md`), creating `cortex/` automatically if it does not yet exist.
- **FR-003**: The document's content MUST be synthesized prose covering the Spec's purpose, what was implemented, and its outcome — not a verbatim concatenation or restatement of its source files.
- **FR-004**: The document MUST draw on: the Spec's own artifacts (`spec.md`, `plan.md`, `tasks.md`, and `validation.md` if it exists), the commits made for it (per FR-005), the project's Knowledge base, the Constitution, and any Learning artifacts relevant to that Spec.
- **FR-005**: The commit history included MUST be every commit on the Spec's own Feature branch from the commit that first added that Spec's `plan.md` (inclusive) through the current state (inclusive) — determined by the file's own git history, never by requiring a commit-message convention.
- **FR-006**: If the named Spec has no Plan yet, the command MUST stop and recommend creating one first, without producing a document.
- **FR-007**: If the project is not a Git repository, or the Spec's commit range cannot be determined, the command MUST still produce a document from its other sources, explicitly noting that commit history could not be included.
- **FR-008**: If the Spec's implementation is incomplete (not every Task attempted), the document MUST state this plainly and name what remains outstanding, rather than presenting the Spec as finished.
- **FR-009**: Re-running the command for a Spec that already has a wrap-up document MUST replace it with a freshly generated one reflecting the Spec's current state, not duplicate it or leave it stale.
- **FR-010**: When a relevant source (Knowledge, Constitution content, or Learnings) has nothing applicable to the Spec, the document MUST say so rather than fabricating content to fill that section.

### Key Entities

- **Wrap-Up Document**: A Markdown file under `cortex/`, one per Spec, named `SPEC-###-<slug>.md`, synthesizing that Spec's purpose, implementation, and outcome from its own artifacts, its commit history, Knowledge, Constitution, and Learnings — intended as raw material for a future official documentation site or GitBook, not itself a canonical project-state artifact like a Spec or Plan.
- **Spec Commit Range**: The commits on a Spec's own Feature branch from the one that first added its `plan.md` through the current state — derived live from Git history, not stored anywhere.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can produce a readable, self-contained documentation page for any Spec with a Plan, in one command, without manually gathering commit history, Knowledge, or other source material themselves.
- **SC-002**: 100% of generated wrap-up documents accurately reflect the Spec's actual implementation completeness — a Spec with outstanding Tasks is never presented as finished.
- **SC-003**: Re-running the command for the same Spec after further work always yields a single, current document — never duplicate or stale files for the same Spec.
- **SC-004**: The commit history reflected in a wrap-up document is correct regardless of what commit messages say, verified against Git's own record of when `plan.md` was first added.

## Assumptions

- `cortex/` sits at the project root, parallel to `ai/` — it is not part of the existing `ai/` semantic-state hierarchy, since its own content (documentation prose) is a downstream by-product of project state, not the state itself.
- "The Spec's own Feature branch" relies on the existing Feature-branch automation (`022-feature-branch-automation`) already associating a Feature (and everything under it, including this Spec) with one dedicated branch; this feature does not change how that association is made.
- The wrap-up document is explicitly not a canonical, validated project artifact the way a Spec/Plan/Tasks/Validation is — it is downstream, human/documentation-facing output, so it is reasonable for it to be regenerated wholesale on every run rather than amended in place.
- Exact prose style, section headings, and length of the generated document are planning-level detail, not fixed by this specification, beyond the requirement that it read as synthesized prose (FR-003) suitable for eventual official docs/GitBook use, not a raw dump of its sources.
