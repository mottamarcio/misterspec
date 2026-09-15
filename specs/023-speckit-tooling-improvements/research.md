# Research: Spec-Kit Tooling Improvements

## 1. Hook-invocation boilerplate (User Story 1)

**Decision**: Append one sentence after every mandatory-hook block in all 9
skills: *"After emitting the block above you MUST actually invoke the hook
and wait for it to finish before continuing. Run it the same way you would
run the command yourself in this agent/session. Emitting the block alone
does not run the hook."* Also change every "skip hook checking silently and
continue normally" line (both pre- and post-execution checks, all 9 skills)
to instead report the parse failure explicitly, naming which hooks (if any
are known to be mandatory) could not be checked.

**Rationale**: This is upstream's own fix for the exact bug this project's
own history already demonstrated — copied near-verbatim since the wording
was already precise and battle-tested. No adaptation needed beyond
replacing upstream's generic `{command}`/`__SPECKIT_COMMAND_*__` templating
placeholders with our own already-resolved slash-command names (our skills
don't use that multi-agent templating layer — each `SKILL.md` is already
written for one specific agent).

**Alternatives considered**: A single shared "hooks" doc referenced by all
9 skills instead of repeating the text 9 times: rejected — every other
Spec-Kit convention in this repo already duplicates the hook-checking
boilerplate per-skill (each `SKILL.md` must stand alone, since a coding
agent loads one skill at a time with no guarantee of cross-file context),
and introducing a new "shared doc" indirection here would be exactly the
kind of new layer Constitution Principle IV rules out without demonstrated
need.

## 2. Checklists as a read-only gate during implementation (User Story 2)

**Decision**: Add to `speckit-implement/SKILL.md`'s existing checklist-status
step: *"Treat checklist markers as a read-only gate: scan checkbox state,
report status, and ask before proceeding when needed; do NOT modify
checklist files or markers."* Add to `speckit-checklist/SKILL.md`: *"This
command generates or appends checklist items; it MUST NOT mark generated
items `[x]`."* Add to `.specify/templates/checklist-template.md`: the
`Review Ownership` and `Marker Semantics` lines from upstream, plus the
note that `/speckit-implement` reads checklist state as a gate and must not
modify markers.

**Rationale**: Direct port — upstream's own wording already distinguishes
`checklists/requirements.md` (spec-quality, maintained by
`speckit-specify`/`speckit-clarify`) from custom checklists (maintained by
`speckit-checklist`, reviewer-owned) without inventing new terminology;
reusing it keeps this repo's own checklist vocabulary consistent with the
upstream project it's derived from, which matters since future upstream
syncs stay diffable.

**Alternatives considered**: Enforcing this mechanically (e.g. a
pre-commit hook that diffs checklist files before/after an
`speckit-implement` run): rejected as disproportionate — this is
instruction text read by a coding agent, not a CI-enforced invariant; a
clear, explicit instruction is the same enforcement mechanism every other
constraint in these skills already relies on (e.g. "MUST NOT rewrite
spec.md" in `speckit-plan`).

## 3. `speckit-taskstoissues` deduplication (User Story 3)

**Decision**: Port upstream's dedup algorithm verbatim: before creating any
issue, call `list_issues` (GitHub MCP server) with no `state` filter
(returns both open and closed), `perPage: 100`, paginating via `after`
using the previous response's `endCursor`; match each returned issue title
against `\bT\d{3,}\b` (accepts our own already-3-digit-padded IDs and any
future wider ones without a ceiling); mark any task ID that matches as
already-issued; skip creating an issue for it and report
`"<ID> already has an issue, skipping"`; stop paginating once every task ID
in the current `tasks.md` is accounted for, or no pages remain.

**Rationale**: This exact regex and pagination strategy is already proven
upstream and directly closes a reproducible bug (duplicate issues on
re-run) — no reason to design a different mechanism. The `\bT\d{3,}\b`
pattern (not `\bT\d{3}\b`) matters specifically for this project since IDs
are zero-padded to a *configurable* width (`project.Configuration.IDWidth`
in the actual Go product, though irrelevant to this repo's own `tasks.md`
IDs, which are also `T`-prefixed 3-digit) — using `{3,}` rather than a
fixed `{3}` means a future wider ID format still matches correctly.

**Alternatives considered**: Storing a local mapping of task ID → issue
number (e.g. in a dotfile) instead of querying GitHub live: rejected —
reintroduces exactly the persisted-secondary-state pattern Constitution
Principle III forbids for the Go product, and the same reasoning applies
here: GitHub's own issue list is already the authoritative record: no
reason to duplicate it locally and risk drift.

## 4. `speckit-constitution` Scope Guard (User Story 4)

**Decision**: Port upstream's `## Scope Guard` section verbatim (adjusted
only to remove the `__SPECKIT_COMMAND_*__` templating placeholder, naming
`/speckit-specify` directly): classify every part of the user's input as
either constitution content or a separate, non-governance intent; never
execute a feature-implementation/code-gen/refactor/deploy request found
mixed in; extract such requests as deferred intents; surface them in a
`Next Actions` section (omitted entirely when there are none) suggesting
the appropriate follow-up command without invoking it; ask for
clarification rather than guessing when classification is genuinely
ambiguous.

**Rationale**: A near-exact match for Constitution Principle VII (Explicit
Mutation Boundaries) already governing this repo — adopting upstream's own
wording is simpler and more consistent than drafting new phrasing for a
principle this project already has.

**Alternatives considered**: None seriously considered — this is a narrow,
clearly-scoped instruction addition with an obvious single correct shape.

## 5. `## Done When` self-check gate (User Story 5)

**Decision**: Add a `## Done When` section (a short checklist of that
skill's own required outputs) to `speckit-specify`, `speckit-plan`,
`speckit-tasks`, `speckit-implement`, and `speckit-clarify`, ported
near-verbatim from upstream's own per-command lists (e.g. `speckit-plan`:
"Plan workflow executed and design artifacts generated"; "Extension hooks
dispatched or skipped..."; "Completion reported to user with branch, plan
path, and generated artifacts").

**Rationale**: Cheap, direct port; each command's own existing completion-
report step already names roughly the same information — this makes the
self-check explicit and checkable rather than implicit in prose.

**Alternatives considered**: A single shared "Done When" checklist
referenced by all five: rejected for the same reason as decision #1 — each
`SKILL.md` must remain self-contained.

## 6. Adopting `speckit-converge` (User Story 6)

**Decision**: Port upstream's `converge.md` into a new
`.claude/skills/speckit-converge/SKILL.md`, adapted to this repo's own
skill conventions: our own frontmatter shape (`name`/`description`/
`argument-hint`/`compatibility`/`metadata`, matching every other
`speckit-*` skill) instead of upstream's generic `scripts:`/`tools:`
frontmatter; our own already-resolved script path
(`.specify/scripts/bash/check-prerequisites.sh --json --require-tasks
--include-tasks` — our local script has no `--require-spec` flag, so
`spec.md`'s own existence is checked as an explicit extra step) instead
of the `{SCRIPT}` placeholder;
our own hook-boilerplate text (per decision #1) instead of
`__SPECKIT_COMMAND_*__` placeholders. The command's own operating
constraints (append-only, never rewrite `spec.md`/`plan.md`/existing
tasks; byte-for-byte-unchanged `tasks.md` when converged; Constitution
violations always CRITICAL and listed first) are ported unchanged — they
are already exactly right for this project's own philosophy.

Two new `extensions.yml` hook keys are added — `before_converge` and
`after_converge` — populated with the same optional, disabled-by-default
`speckit.git.commit` entries every other command already has, so
convergence participates in the existing auto-commit hook system rather
than being a special case.

**Rationale**: The gap is real (documented in spec.md's own motivation:
every implementation task in this project's own history has ended with a
manually-remembered "regression checkpoint" task) and upstream's own
design already satisfies this repo's own safety posture (append-only,
Constitution-aware, no application-code writes) without modification —
porting it is lower-risk than designing an equivalent capability from
scratch.

**Alternatives considered**:
- Designing a bespoke convergence check tailored to this repo's own
  artifact numbering instead of porting upstream's: rejected — upstream's
  design already generalizes correctly to this repo's own `specs/NNN-*`
  layout with no adaptation needed beyond boilerplate/frontmatter, so a
  bespoke redesign would only add risk for no benefit.
- Making `speckit-converge` mandatory (e.g. required before
  `speckit-implement` can report done): rejected — it is an on-demand,
  optional check per spec.md's own User Story 6 framing ("a developer
  wants to know..."), not a blocking gate; over-enforcing it would be
  scope creep beyond what was asked.

## 7. `speckit-clarify` question quality and post-clarify checklist re-validation (User Story 7)

**Decision**: Port upstream's question-writing quality rules verbatim: a
question MUST be a full interrogative ending in `?`; the only permitted
suffix after `?` is an optional parenthesized requirement ID (never a bare
ID or topic label used as the question itself); a "Why it matters"
sentence must immediately follow. Port upstream's post-clarify checklist
re-validation step (§9 in `clarify.md`): after a clarification round
completes, if `checklists/requirements.md` exists, re-evaluate each
checkbox against the updated spec, toggle only items whose state actually
changed, and report newly-passing items and any regressions in the
completion report.

**Rationale**: Both are concrete, already-tested mechanics with clear
acceptance criteria (spec.md's own FR-011/FR-012) — no redesign needed.

**Alternatives considered**: Re-validating the *entire* checklist file
unconditionally (rewriting it fresh each time) instead of toggling only
changed markers: rejected — upstream's own approach (toggle only what
changed) avoids noisy diffs on a file humans may also be editing directly,
consistent with this project's own "no incidental rewrites" instinct
(Constitution Principle VII, User Story 2's own read-only-gate theme).

## 8. `speckit-plan`/`speckit-tasks` scope tightening (User Story 8)

**Decision**: Add to `speckit-plan/SKILL.md`'s quickstart-generation step:
*"Do not include full implementation code, model/service/controller
bodies, migrations, or complete test suites [in quickstart.md]. Keep this
artifact as a validation/run guide."* Add to `speckit-tasks/SKILL.md`'s
"From Data Model" task-generation rule: *"For each field with constraints
in data-model.md (max length, nullable/required, enum values, validation
rules), quote the constraint verbatim in the task description so it is not
left to implementation-time discretion."*

**Rationale**: Both are narrow, low-risk precision improvements with
obvious correct wording already validated upstream — no adaptation beyond
copying the text into this repo's own equivalent step.

**Alternatives considered**: None — these are the smallest, least
ambiguous items in this set.
