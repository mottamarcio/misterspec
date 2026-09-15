---

description: "Task list template for feature implementation"
---

# Tasks: Project Documentation Site and README

**Input**: Design documents from `/specs/020-docs-site-readme/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/content-inventory.md, quickstart.md

**Tests**: Not applicable — this feature ships static content only (plan.md's own Technical Context). Verification is `contracts/content-inventory.md`'s own checklist, reconciled against the actual published content once written, plus a manual accuracy check of every documented command example against its own current source (FR-009).

**Organization**: Tasks are grouped by user story. **User Story 1** (README) has no dependency on Foundational or User Story 2 — it can be done first, last, or in parallel. **User Story 2** (the site) depends on Foundational (the shared `nav.js`/`styles.css`) but its own 9 pages are otherwise independent files.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2)
- Every task names its exact file path

## Path Conventions

```text
README.md                              # rewritten (US1)
site/                                  # new (US2)
.github/workflows/deploy-docs.yml      # new (US2)
```

---

## Phase 1: Setup

**Purpose**: Create the new directories this feature writes into. Nothing to configure — no dependency, no build tool.

- [X] T001 Create the `site/`, `site/assets/`, and `.github/workflows/` directories.

**Checkpoint**: Directories exist — proceed to Foundational and/or User Story 1 in parallel.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The one shared piece every site page (User Story 2) needs but none of them individually owns — the navigation data/renderer and the small non-Tailwind-utility CSS. Does not block User Story 1 (README) at all.

**⚠️ CRITICAL**: No User Story 2 page may be authored until this phase completes. User Story 1 has no dependency on this phase.

- [X] T002 [P] Create `site/assets/nav.js` implementing data-model.md's own `NAV` array and the `DOMContentLoaded` sidebar-render/current-page-highlight logic (research.md #2).
- [X] T003 [P] Create `site/assets/styles.css` with the minimal non-Tailwind-utility rules needed (sidebar layout/scroll behavior, code-block wrapping, any print-safe tweaks) — everything expressible as a Tailwind utility class stays inline in the HTML, not here (Principle IV: no parallel styling system).

**Checkpoint**: Foundation ready — `site/assets/nav.js` and `site/assets/styles.css` exist. User Story 2 page authoring can begin.

---

## Phase 3: User Story 1 - Understand and Adopt misterspec From the README (Priority: P1) 🎯 MVP

**Goal**: The root README alone answers what misterspec is, why it exists, how to install it, and where to go for more — reflecting the project's actual current, shipped state.

**Independent Test**: Read only `README.md`; confirm it answers "what," "why," "how do I install," and "what do I do first," and links to the site.

- [X] T004 [US1] Rewrite `README.md` per `contracts/content-inventory.md`'s own README checklist: what misterspec is and the problem it solves (FR-001, no prior context assumed); installation instructions sufficient on their own (FR-002); a minimal quickstart (FR-003); exactly one link to the published site (FR-004); remove all placeholder content from the pre-feature one-line version.

**Checkpoint**: User Story 1 is independently complete — `README.md` alone satisfies SC-001/SC-002, verifiable by reading it with no other file open.

---

## Phase 4: User Story 2 - Explore the Complete Project Reference on the Documentation Site (Priority: P2)

**Goal**: A 9-page, GitBook-style static site covering misterspec's own concept, every shipped feature (001-019), and every currently-registered command, each with a real, current example — published and reachable with no local build step.

**Independent Test**: Open the published site with no other context; navigate to any feature or command page and confirm it explains what/why/how-it-relates (features) or purpose/flags/example (commands).

### Concept and orientation pages

- [X] T005 [P] [US2] Create `site/index.html`: misterspec's own core concept (the deterministic/semantic split between the `misterspec` binary and the coding agent) and the artifact lifecycle (Constitution → Program → Feature → Spec → Plan → Tasks → Implementation → Validation), stated before any feature-specific detail (FR-006). Includes the shared `<head>` (Tailwind CDN, Iconify web-component script, `assets/styles.css`, `assets/nav.js`) and `<div id="sidebar">` every page shares. Depends on T002, T003.
- [X] T006 [P] [US2] Create `site/getting-started.html`: installation instructions and a first end-to-end run, consistent with (but more detailed than) the README's own quickstart. Depends on T002, T003.
- [X] T007 [P] [US2] Create `site/workflow.html`: the full Spec-Driven lifecycle illustrated end to end, naming the real Skill/command invoked at each stage (`/create-constitution`, `/create-program`, `/create-feature`, `/create-specs`, `/create-plan`, `/create-tasks`, `/implement`, `/analyze`) — the concrete, illustrated usability example spec.md calls for. Depends on T002, T003.

### Feature pages (data-model.md's own Feature-to-page inventory — FR-007)

- [X] T008 [P] [US2] Create `site/foundation.html` covering Specs 001-005 (Core Repository Foundation, Read-Only Deterministic Operations, Atomic Entity Creation, Structural Validation & Project Status, Embedded Kit & Resource Installer), each with what/why/how-it-fits/example (data-model.md's own content model). Depends on T002, T003.
- [X] T009 [P] [US2] Create `site/agents-cli.html` covering Specs 006-010 (Agent Adapter Layer, Project Bootstrap, CLI/Cobra & the `internal` command tree, Canonical Skills Content, Interactive `init` TUI). Depends on T002, T003.
- [X] T010 [P] [US2] Create `site/context-engine.html` covering Specs 011-017 as one connected pipeline narrative with a pipeline diagram and one anchored section per phase (Wikilink Graph Foundation, References and Backlinks, Document Model and Chunking, Disposable SQLite Index, Context Collector and Retrieval, Ranking and Budgeting, Internal Context Command) — research.md #3. Depends on T002, T003.
- [X] T011 [P] [US2] Create `site/multi-agent-skills.html` covering Spec 018 (the five additional coding-agent adapters and the four Context-Pack-aware Skills). Depends on T002, T003.
- [X] T012 [P] [US2] Create `site/dogfooding.html` covering Spec 019 (the evaluation methodology and its own real findings/outcome from `specs/019-dogfooding-evaluation/report.md`). Depends on T002, T003.

### Command reference

- [X] T013 [US2] Create `site/commands.html` documenting all 14 currently-registered commands (data-model.md's own Command inventory table) — public `init` plus the 13 `internal` commands — each with its real signature, its real flags read directly from its own current `internal/cli/{init.go,internalcmd/*.go}` source (research.md #4), and at least one real, copyable example invocation with expected result (FR-008, FR-009). Depends on T002, T003.

### Deployment

- [X] T014 [US2] Create `.github/workflows/deploy-docs.yml` using `actions/configure-pages`, `actions/upload-pages-artifact` (source: `site/`), and `actions/deploy-pages`, triggered on push to `dev` (research.md #5).

**Checkpoint**: User Story 2 is independently complete — every page exists, cross-links via `nav.js`, and the deploy workflow is in place (SC-003, SC-004, SC-005 verifiable once merged and Pages is enabled).

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final accuracy and usability verification across both stories.

- [X] T015 Reconciled `commands.html` against `contracts/content-inventory.md`'s own "Command coverage" checklist — re-read every command's own current `internal/cli/{init.go,internalcmd/*.go}` source directly (`grep`'d `Short:`/`Flags()`/`WriteSuccess` shapes for all 14), confirming every documented flag and JSON shape matches exactly (FR-009).
- [X] T016 [P] Verified 400px-width safety structurally (no local browser render available in this environment): confirmed zero hardcoded `min-w-[…]` pixel widths across all 9 pages, every wide element (tables, the pipeline diagrams) wrapped in `overflow-x-auto`, and the sidebar is `fixed -translate-x-full` off-canvas on mobile (never forcing page width) — `<div>` tag balance and `html.parser`-clean markup confirmed on every page (FR-011, SC-006).
- [X] T017 [P] Verified every sidebar link programmatically: extracted every `href="*.html*"` reference across all 9 pages and confirmed each resolves to a real file (`ls site/*.html`) and every anchor (`commands.html#internal-context`, `context-engine.html#ranking-budgeting`) resolves to a real heading `id` — zero broken links found.
- [X] T018 Final reconciliation of `contracts/content-inventory.md`'s full checklist against the actual, published `README.md` and `site/` content. **One real gap found and fixed**: the Feature coverage item "every feature page states what/why/how-it-fits/example for every Spec" was not yet true — `foundation.html` (Spec 005), `agents-cli.html` (Specs 006–009), and `context-engine.html` (Specs 011, 013–016) were each missing a concrete example. Added one real, accurate example to each rather than marking the checklist complete without merit; re-verified div-balance/HTML-validity after the additions.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup. Blocks User Story 2 only.
- **User Story 1 (Phase 3)**: Depends on Setup only — proceeds in parallel with Foundational and User Story 2.
- **User Story 2 (Phase 4)**: Depends on Foundational.
- **Polish (Phase 5)**: Depends on User Story 1 and User Story 2 both being complete.

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on User Story 2 — this feature's own MVP.
- **User Story 2 (P2)**: Depends only on Foundational; independent of User Story 1's own content (though the README does link to it).

### Parallel Opportunities

- T002/T003 (Foundational) can proceed in parallel.
- T004 (User Story 1) can proceed at any point in parallel with Foundational and User Story 2 — different file.
- T005-T012 (8 of User Story 2's 9 pages) can all be authored in parallel once Foundational completes — independent files.
- T016/T017 (Polish) can proceed in parallel.

---

## Parallel Example: User Story 2's Content Pages

```bash
# Eight independent files, once T002/T003 exist:
Task: "Create site/index.html"
Task: "Create site/getting-started.html"
Task: "Create site/workflow.html"
Task: "Create site/foundation.html"
Task: "Create site/agents-cli.html"
Task: "Create site/context-engine.html"
Task: "Create site/multi-agent-skills.html"
Task: "Create site/dogfooding.html"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 3: User Story 1 (`README.md`).
3. **STOP and VALIDATE**: the README alone already satisfies this
   feature's own smallest, fastest-to-deliver value (SC-001, SC-002) —
   reachable without the site existing at all.

### Incremental Delivery

1. Setup → directories exist.
2. User Story 1 → `README.md` rewritten (MVP).
3. Foundational → shared nav/styles ready.
4. User Story 2 → all 9 pages + deploy workflow.
5. Polish (Phase 5), including the final content-inventory reconciliation.

### Team Strategy

User Story 1 and Foundational can proceed in parallel from the start.
Once Foundational lands, User Story 2's 8 content pages (T005-T012)
are fully parallel — different authors could take one page each with
zero coordination beyond the shared `nav.js` each merely references.

---

## Notes

- No task in this feature touches `internal/`, `cmd/`, `kit/`, or `docs/` — everything lives in `README.md`, `site/`, and one new workflow file (FR-012).
- Every command example (T013, T015) must be verified against that command's own current source, never written from memory (research.md #4, FR-009) — several commands have already undergone real, documented corrections mid-implementation (e.g. `internal context`'s own `--budget` flag, 017's research.md).
- `site/assets/nav.js` is the single place a new page's link is registered (quickstart.md §4) — do not hand-duplicate sidebar markup into any page.
- Commit after each phase; the README (User Story 1) is safe to ship on its own even before User Story 2 completes.
