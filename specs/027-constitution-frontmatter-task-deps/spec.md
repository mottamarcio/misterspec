# Feature Specification: Constitution Frontmatter Guarantee, and Task Dependency/Parallelism Reporting

**Feature Branch**: `027-constitution-frontmatter-task-deps`
**Created**: 2026-09-16
**Status**: Draft
**Input**: User description: "unir as duas sugestões acima em uma só spec — frontmatter no create-constitution (independente do modelo usado), e no summary final de /create-tasks adicionar quais tasks têm dependência e quais podem ser implementadas em paralelo, para times trabalhando em paralelo. (Refined: the user separately asked whether frontmatter should be enforced for ALL generated Markdown artifacts, not just the Constitution. Investigation confirmed every other artifact type — Program, Feature, Spec, Knowledge, Learning, Plan, Tasks, Validation — is already created through misterspec's own deterministic template-rendering layer, which is already tested to always include correct frontmatter (`internal/templates`, `templates_test.go`). The Constitution is the sole exception: it has no entity ID and no deterministic creation operation, so it is the only artifact type an agent ever writes freehand — confirmed as the actual, and only, gap.)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The Constitution always has its required frontmatter (Priority: P1)

A developer runs `/create-constitution` (in any agent integration — Claude Code, Antigravity, or any other) to draft or amend the project's Constitution. Today, `docs/architecture-specification.md` §24 defines a required frontmatter block (`type: constitution`, `schema_version: 1`) for `ai/memory/constitution.md`, but `create-constitution`'s own Skill instructions never mention it — so whether the written file actually gets that frontmatter depends entirely on what the underlying model happens to do on its own, which has already been observed to differ between agent integrations.

**Why this priority**: This is the concrete, reported bug — the Constitution is the one artifact type with no deterministic creation path (Constitution itself has no entity ID, per `docs/architecture-specification.md` §24 and `create-constitution/SKILL.md`'s own Forbidden Mutations), so it is uniquely exposed to model-dependent omissions that every other artifact type is already immune to.

**Independent Test**: Run `/create-constitution` against a project with an existing Knowledge base and no Constitution yet; read the resulting `ai/memory/constitution.md` and confirm it opens with a YAML frontmatter block containing `type: constitution` and `schema_version: 1`, exactly as §24 requires.

**Acceptance Scenarios**:

1. **Given** a project with Knowledge but no Constitution yet, **When** a developer runs `/create-constitution` for the first time, **Then** the written `ai/memory/constitution.md` begins with a YAML frontmatter block containing at minimum `type: constitution` and `schema_version: 1`, followed by the required body sections.
2. **Given** a Constitution that already exists (with correct frontmatter), **When** a developer runs `/create-constitution` again to amend it, **Then** the existing frontmatter is preserved (not duplicated, not stripped) while the body is amended.
3. **Given** a Constitution file that exists but is missing its required frontmatter (e.g. written before this feature, or by a model that omitted it), **When** a developer runs `/create-constitution` again, **Then** the Skill detects the missing frontmatter and adds it, rather than treating the file as already fully correct.

---

### User Story 2 - create-tasks reports which Tasks depend on which, and which can run in parallel (Priority: P1)

A developer (or a team splitting work across multiple people) runs `/create-tasks` to decompose a Spec's Plan into Tasks. Today's completion summary reports how many Tasks were added and why, but says nothing about which Tasks block which others, or which ones are free to be picked up simultaneously by different people — even though `tasks.md` already records each Task's own dependency ("what it depends on... or 'none'") and parallel marker.

**Why this priority**: Equal in importance to User Story 1 — this is the other concrete request, and it's valuable specifically because a team wants this information immediately after generation, without having to read the whole `tasks.md` file themselves to work out an execution plan.

**Independent Test**: Run `/create-tasks` against a Spec whose Plan implies at least one dependency chain and at least one pair of independent Tasks; confirm the completion summary explicitly lists which Tasks depend on which, and which Tasks have no unmet dependency between them (safe to work on in parallel).

**Acceptance Scenarios**:

1. **Given** a Plan that decomposes into Tasks where Task B depends on Task A, **When** `/create-tasks` finishes, **Then** the completion summary states that dependency explicitly (e.g. "Task B depends on Task A"), not just implicitly via the Tasks file's own content.
2. **Given** a Plan that decomposes into two or more Tasks with no dependency relationship between them, **When** `/create-tasks` finishes, **Then** the completion summary explicitly names that group of Tasks as safe to implement in parallel.
3. **Given** a Plan with a mix of dependent and independent Tasks, **When** `/create-tasks` finishes, **Then** the summary distinguishes the two clearly — dependency chains and parallel-safe groups are both visible, not merged into one undifferentiated list.
4. **Given** `/create-tasks` is re-run and extends an existing `tasks.md` with new Tasks, **When** it finishes, **Then** the dependency/parallelism reporting covers the Tasks file's current full state, not only the newly added Tasks in isolation.

### Edge Cases

- What happens when a Constitution file already has frontmatter, but it's missing one of the two required fields (e.g. `type` present, `schema_version` absent)? The missing field is added; an already-correct field is left untouched, consistent with `create-constitution`'s own existing idempotency guarantee (amend, don't rewrite wholesale).
- What happens when `/create-tasks` produces only a single Task (no dependency, nothing to parallelize)? The summary states plainly that there is nothing to report for dependencies/parallelism, rather than showing an empty or confusing section.
- What happens when every Task in a Spec turns out to be fully sequential (a straight dependency chain, no parallel opportunities at all)? The summary reports the full chain order and explicitly states there is no parallel-safe grouping, rather than omitting the topic.
- What happens when a Task's declared dependency doesn't exist (an authoring mistake)? Correction (found during `/speckit-analyze`): `internal validate` does **not** currently check this — it only detects duplicate Task ID numbers (`CodeDuplicateID`); a Task's free-text "depends on" reference is never checked against the other Tasks that actually exist, unlike a Spec's own structured `depends_on`/`supersedes` fields, which are checked. This feature's dependency/parallelism reporting therefore describes whatever a Task's own text claims, unverified — a dependency naming a nonexistent Task would still be reported as a dependency relationship rather than flagged as broken. Verifying Task dependency targets exist is out of scope for this feature; it would need its own dedicated spec.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `/create-constitution` MUST write `ai/memory/constitution.md` with a YAML frontmatter block containing at minimum `type: constitution` and `schema_version: 1`, per `docs/architecture-specification.md` §24 — regardless of which underlying model or agent integration is running the Skill.
- **FR-002**: When amending an existing Constitution that already has correct frontmatter, `/create-constitution` MUST preserve it unchanged — never duplicating or rewriting an already-correct frontmatter block.
- **FR-003**: When amending an existing Constitution whose frontmatter is missing or incomplete, `/create-constitution` MUST add the missing required field(s) rather than leaving the file non-conformant to §24.
- **FR-004**: `/create-tasks`'s completion summary MUST explicitly state each dependency relationship between Tasks it just generated or extended (which Task depends on which), not only rely on `tasks.md`'s own content to convey it.
- **FR-005**: `/create-tasks`'s completion summary MUST explicitly identify groups of Tasks that have no dependency on each other (and are therefore safe to implement in parallel, e.g. by different team members).
- **FR-006**: The dependency/parallelism reporting in `/create-tasks`'s summary MUST reflect the Tasks file's current full state after the invocation, including Tasks from earlier invocations, not only Tasks newly added in this run.
- **FR-007**: When a Spec's Tasks have no dependency relationships and no parallel opportunities to report (e.g. exactly one Task, or every Task fully sequential with no parallel opportunity), the summary MUST state that plainly rather than omitting the topic or showing a misleadingly empty section.

### Key Entities

- **Constitution Frontmatter**: The required YAML metadata block (`type: constitution`, `schema_version: 1`) at the top of `ai/memory/constitution.md`, per `docs/architecture-specification.md` §24 — distinct from the Constitution's own body sections (Product/Architecture/Security/etc. Invariants).
- **Task Dependency**: An existing, already-recorded relationship in `tasks.md` (one Task naming another as a prerequisite) — this feature reports it more visibly, it does not change how or where it is recorded.
- **Parallel-Safe Group**: A set of Tasks with no dependency relationship among them, derivable from the Tasks file's own existing dependency records — not a new field, a new way of presenting existing information.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of Constitution files written or amended by `/create-constitution` (across any agent integration) contain the required `type`/`schema_version` frontmatter fields, verified by reading the file after the Skill completes.
- **SC-002**: 100% of `/create-tasks` completion summaries for a Spec with two or more Tasks explicitly name at least one of: a dependency relationship, or a parallel-safe group — never silently omitting both when either exists in the underlying Tasks file.
- **SC-003**: A developer splitting work across a team can identify which Tasks are safe to assign to different people immediately from `/create-tasks`'s own completion summary, without needing to open and manually parse `tasks.md`.

## Assumptions

- The Constitution frontmatter contract (`type: constitution`, `schema_version: 1`) is exactly `docs/architecture-specification.md` §24's own existing, already-defined schema — this feature does not change or extend that schema, it closes the gap between the schema's own definition and `create-constitution/SKILL.md`'s own instructions, which currently omit it entirely.
- No other artifact type (Program, Feature, Spec, Knowledge, Learning, Plan, Tasks, Validation) needs equivalent work: each is already created through `internal create`/`internal create-artifact`, backed by `internal/templates`' own tested rendering, which already guarantees correct frontmatter deterministically, independent of which model or agent is running the Skill.
- Task dependency and parallel-group data already exists in `tasks.md`'s own content today (`create-tasks/SKILL.md`'s own existing Outputs: each Task records "what it depends on... or 'none'", plus an existing `[P]` parallel marker) — this feature surfaces that existing data more visibly in the completion summary; it does not introduce a new dependency-tracking mechanism.
