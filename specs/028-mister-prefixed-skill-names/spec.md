# Feature Specification: `mister-`-Prefixed Skill Names to Avoid Slash-Command Collisions

**Feature Branch**: `028-mister-prefixed-skill-names`
**Created**: 2026-09-16
**Status**: Draft
**Input**: User description: "eu tive um problema de conflito de nome dos 'slash commands' com outros frameworks. Acho que teremos que renomear as skills (e invocações). Por exemplo: 'analyze' para 'mister-analyze'; 'create-constitution' para 'mister-constitution'; 'create-feature' para 'mister-features'; 'create-knowledge-base' para 'mister-knowledge-base'; 'create-plan' para 'mister-plan'; 'create-program' para 'mister-program'; 'create-specs' para 'mister-specify'; 'create-tasks' para 'mister-tasks'; 'implement' para 'mister-implement'. (Confirmed with the user: the two names that don't follow the plain "strip create-, add mister-" pattern — `mister-features` and `mister-specify` — are intentional, not typos. Confirmed: full rename, old names stop working entirely, no alias/dual-naming period.)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A misterspec user's slash commands no longer collide with another framework's (Priority: P1)

A developer uses misterspec alongside one or more other agent frameworks in the same project or agent environment. Today, misterspec's 9 canonical Skills install as short, generic slash commands (`/analyze`, `/implement`, `/create-tasks`, etc.) — names common enough that another framework installed in the same environment can register an identically-named command, and the developer can no longer reliably invoke the one they mean.

**Why this priority**: This is the exact problem reported — without this, misterspec is unusable alongside any other framework that happens to claim the same generic command name, which is a real, already-encountered blocker to adoption.

**Independent Test**: Initialize a project with misterspec (`misterspec init`), inspect the installed slash commands, and confirm every one of the 9 begins with `mister-` and none of the old generic names (`analyze`, `create-tasks`, `implement`, etc.) exist anywhere in the installed set.

**Acceptance Scenarios**:

1. **Given** a fresh project, **When** a developer runs `misterspec init`, **Then** the 9 installed commands are exactly: `/mister-analyze`, `/mister-constitution`, `/mister-features`, `/mister-knowledge-base`, `/mister-plan`, `/mister-program`, `/mister-specify`, `/mister-tasks`, `/mister-implement` — no more, no fewer.
2. **Given** the same project, **When** the developer looks for any of the old command names (`/analyze`, `/create-constitution`, `/create-feature`, `/create-knowledge-base`, `/create-plan`, `/create-program`, `/create-specs`, `/create-tasks`, `/implement`), **Then** none of them exist as installed commands.

---

### User Story 2 - Renamed Skills still correctly guide the user to each other (Priority: P1)

A developer follows the "next step" guidance one Skill's own completion summary gives them (e.g. finishing `/mister-tasks` and being told what to run next). Every Skill's own instructions reference several *other* Skills by name — as the next step in the pipeline, as a dependency, or in an explanation of what another Skill does.

**Why this priority**: Equal in importance to User Story 1 — a rename that changes each Skill's own name but leaves its *references to other Skills* pointing at old, now-nonexistent commands would strand users mid-pipeline with instructions to run a command that no longer exists.

**Independent Test**: Read through all 9 renamed Skills' own content and confirm every mention of another Skill — in guidance about what to run next, in an explanation of a dependency, or in ordinary prose — uses that other Skill's new name; confirm no old name appears anywhere in any of the 9.

**Acceptance Scenarios**:

1. **Given** the renamed `/mister-tasks` Skill, **When** a developer finishes running it, **Then** the next-step guidance it gives names `/mister-implement`, not `/implement`.
2. **Given** any of the 9 renamed Skills, **When** its own content explains which other Skill it depends on or hands off to, **Then** every such reference uses that other Skill's new `mister-`-prefixed name.
3. **Given** the full set of 9 renamed Skills, **When** searched for any occurrence of an old command name (`/analyze`, `/create-constitution`, `/create-feature`, `/create-knowledge-base`, `/create-plan`, `/create-program`, `/create-specs`, `/create-tasks`, `/implement`), **Then** zero occurrences are found.

---

### User Story 3 - Old command names are fully retired, not kept as a fallback (Priority: P2)

A developer who has used misterspec before, or who has notes/scripts/muscle memory referencing an old command name, tries one of the old names after upgrading.

**Why this priority**: Lower priority than Stories 1–2 because it's a clarifying constraint on the *absence* of behavior, not new functionality — but worth stating explicitly so no one mistakenly builds or expects a compatibility shim.

**Independent Test**: After upgrading to the renamed Skill set, attempt to invoke an old command name and confirm it is not recognized (whatever "not recognized" ordinarily looks like in the agent tool being used) rather than silently working or redirecting.

**Acceptance Scenarios**:

1. **Given** a project using the renamed Skill set, **When** a developer types an old command name (e.g. `/implement`), **Then** it is not a recognized command — no alias, redirect, or dual-registration keeps it working.

### Edge Cases

- What happens to a project that already ran `misterspec init` under the old (pre-rename) Skill names, and now upgrades to a misterspec version carrying the renamed Skills? The old-named Skill files already installed on disk (e.g. under that agent's own skills directory) are not automatically removed by this rename alone — running `misterspec init` again installs the new `mister-`-prefixed Skills alongside whatever old-named files are already present from before. Cleaning up stale old-named installs from a prior version is out of scope for this feature (see Assumptions).
- What happens to specs already written and completed under the old command names (e.g. a Spec whose own `plan.md` mentions running `/create-tasks`)? Those are historical records of what was actually run at the time — they are not rewritten by this feature; only the currently-active Skill content and its own automated checks change.
- What happens to the two names that don't follow the plain "strip `create-`, add `mister-`" pattern (`create-feature` → `mister-features`, `create-specs` → `mister-specify`)? They are deliberate, confirmed exceptions — not inconsistencies to "fix" during implementation.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST rename each of the 9 canonical Skills exactly per this mapping: `analyze` → `mister-analyze`; `create-constitution` → `mister-constitution`; `create-feature` → `mister-features`; `create-knowledge-base` → `mister-knowledge-base`; `create-plan` → `mister-plan`; `create-program` → `mister-program`; `create-specs` → `mister-specify`; `create-tasks` → `mister-tasks`; `implement` → `mister-implement`.
- **FR-002**: Each renamed Skill's own declared name and its own invocation instructions MUST reflect its new name, not the old one.
- **FR-003**: Every reference any of the 9 Skills makes to any of the other 8 — as a next-step recommendation, a stated dependency, or ordinary explanatory prose — MUST use that other Skill's new name.
- **FR-004**: No installed command MUST continue to answer to any of the 9 old names after this change — there is no alias, redirect, or dual-registration of any kind.
- **FR-005**: The system's own authoritative list of canonical Skill names (wherever the full set of 9 is enumerated as current, active documentation) MUST be updated to the new names, so it does not describe a set of commands that no longer exist.
- **FR-006**: Every automated check that verifies a Skill's own content against its name (structural conformance, cross-reference checks, or any other machine-checked assertion tied to a specific Skill's name) MUST be updated to check the new names, so these checks continue verifying the Skills that actually exist rather than silently passing on stale names or failing to catch real problems.

### Key Entities

- **Skill**: One of the 9 canonical product capabilities (e.g. "generate tasks from a plan"), each identified by a name that doubles as its installed slash-command name.
- **Cross-Reference**: One Skill's own mention of another Skill by name, appearing in its guidance about what to run next, what it depends on, or explanatory prose — not a separate stored record, just content within each Skill's own file.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of the 9 canonical Skills install under their new `mister`-prefixed command names, verified after a fresh `misterspec init`.
- **SC-002**: Zero occurrences of any old command name remain anywhere within the currently-active Skill content or the authoritative canonical-name listing.
- **SC-003**: A developer using misterspec alongside another framework in the same environment can distinguish misterspec's own commands from any other framework's at a glance, by the shared `mister-` prefix, eliminating the specific collision class reported.
- **SC-004**: Every "what to run next" instruction a Skill gives, across all 9, names a command that actually exists after the rename — verified by confirming no Skill's own guidance references a retired name.

## Assumptions

- This rename applies only to misterspec's own native, canonical product Skills (the 9 shipped to end-user projects) — it does not touch this repository's own separate internal development tooling (the unrelated Skill set this repo itself uses to manage its own specs, which has different names and a different origin, and is not shipped to users).
- No backward-compatible alias, redirect, or transition period is provided for the 9 old names, per the user's explicit decision — this is a clean, one-time rename, not a phased migration.
- Cleaning up stale old-named Skill files left behind on disk by a prior version's `misterspec init`, in a project that upgrades to this renamed Skill set, is out of scope — this feature changes what gets installed going forward, not what a previous install already wrote.
- Already-completed Specs (and their own artifacts: `plan.md`, `tasks.md`, etc.) that reference an old command name as part of their own historical record are not rewritten by this feature — only currently-active Skill content and its own automated verification are in scope.
- `mister-features` (not `mister-feature`) and `mister-specify` (not `mister-specs`) are intentional, confirmed exceptions to the otherwise uniform "strip `create-`, add `mister-`" pattern the other 7 names follow.
