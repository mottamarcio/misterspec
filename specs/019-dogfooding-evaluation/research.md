# Phase 0 Research: Dogfooding and Evaluation

All unknowns spec.md's own Assumptions section deliberately deferred to
planning are resolved below, including one structural finding that
reshapes how User Story 1 is actually executed. No `NEEDS
CLARIFICATION` markers remain.

## 1. This repository has no real `ai/` corpus to run `internal context` against

**Finding**: `misterspec`'s own repository (this codebase) is
developed using GitHub Spec Kit's own `specs/NNN-name/{spec,plan,
tasks}.md` convention — the same one this very feature is being
planned through — not misterspec's own `ai/programs/.../specs/
SPEC-###/spec.md` artifact model. There is no `ai/` directory and no
`.misterspec/config.yaml` anywhere in this repository (verified:
`ls ai/` and `ls .misterspec/` both report nothing). `misterspec
internal context` (017) operates exclusively on that `ai/` model — it
has nothing to query here as-is.

**Decision**: Build one small, purpose-built fixture project under
`specs/019-dogfooding-evaluation/fixture/` — a real, filesystem-
backed misterspec project (`.misterspec/config.yaml` + `ai/...`) —
encoding this repository's own **real, already-known** feature
dependency graph (011 through 018, plus 006) as `ai/programs/PRG-001/
features/FEAT-001/specs/SPEC-011` through `SPEC-018`, each Spec's own
`depends_on` set to that feature's own real, historical dependency,
and each Spec's own body condensed from that feature's own real
`spec.md`/`plan.md` Summary — not fabricated content. Two Knowledge
entries and one Learning are added the same way, each wikilinked from
the Spec whose own real work actually referenced that kind of
information (e.g. 018's own research into each agent's real
integration convention). This fixture is what User Story 1's `internal
context` requests run against.

**Rationale**: The real dependency chain (below) and the real
Summary/Intent content of every one of these already-shipped features
already exist — using them, translated into the one artifact shape
`internal context` actually understands, is the closest faithful proxy
to "this project's own real development history" achievable without a
much larger, explicitly out-of-scope migration of misterspec's own
self-development workflow onto its own `ai/` model. It also means
every "did the pack surface the real dependency" check in User Story 1
has a genuinely verifiable right answer, not an invented one.

**Alternatives considered**: (a) Reusing the synthetic
SPEC-011/SPEC-014/KNOW-003 fixture already shared by every 015-018 unit
test. Rejected — that fixture is illustrative fiction (a "refresh
token rotation" example from `docs/context-engine-implementation.md`
itself), not this project's own real history, and reusing it would not
actually test anything spec.md's User Story 1 didn't already cover in
011-018's own unit tests. (b) Migrating misterspec's own real
`specs/001-018` directly into its own `ai/` format as a permanent,
committed dogfooding corpus. Rejected as a much larger, separate
undertaking (a real repository-wide workflow migration) not justified
by this evaluation feature's own narrow purpose — Principle IV.

**Real dependency graph encoded in the fixture** (from each feature's
own already-published spec.md/plan.md, not re-derived):

```text
SPEC-011 (wikilink-foundation)        — depends_on: []
SPEC-012 (references-backlinks)       — depends_on: [SPEC-011]
SPEC-013 (document-model-chunking)    — depends_on: [SPEC-011]
SPEC-014 (sqlite-index)               — depends_on: [SPEC-012, SPEC-013]
SPEC-015 (context-collector)          — depends_on: [SPEC-011, SPEC-012, SPEC-013, SPEC-014]
SPEC-016 (ranking-budgeting)          — depends_on: [SPEC-015]
SPEC-017 (internal-context-command)   — depends_on: [SPEC-014, SPEC-015, SPEC-016]
SPEC-018 (multi-agent-skill-integration) — depends_on: [SPEC-006, SPEC-017]
SPEC-006 (agent-adapter)              — depends_on: [] (referenced by SPEC-018 only)
```

## 2. The same fixture project is the live-dogfooding sandbox for User Story 2/3

**Decision**: The identical fixture project from #1 is also where
User Story 2 and User Story 3's own live Skill invocations happen —
`misterspec init`-equivalent state (a `.misterspec/config.yaml`
already present, both `claude-code` and `agy` adapters' Skills
installed via `internal/agents` into it) so `/create-plan SPEC-018` or
`/implement SPEC-017` can genuinely be invoked through Claude Code and
Antigravity against it.

**Rationale**: The same structural gap #1 identified (no real `ai/`
project in this repository) applies equally to a live Skill
invocation — `/implement`, `/create-plan`, etc. all require a real
target Spec to operate on. Reusing the same fixture avoids building
two separate throwaway projects and keeps User Story 1's static
findings and User Story 2/3's live observations directly comparable
(same Specs, same dependency graph, same Knowledge/Learnings).

**Alternatives considered**: A second, separate live-only fixture.
Rejected — no reason for the two to differ; one fixture serves both
purposes.

## 3. The fixture is a one-off evaluation artifact, not shipped content

**Decision**: `specs/019-dogfooding-evaluation/fixture/` is committed
for reproducibility (so a future maintainer can re-run this same
evaluation later) but is explicitly not part of `kit/` — it is never
embedded into the `misterspec` binary and is not a template or
canonical Skill.

**Rationale**: It exists purely to give this one evaluation something
real to query; it has no product role. Committing it (rather than
generating it ephemerally and discarding it) satisfies FR-009's own
durability requirement — the Dogfooding Report can point at exactly
what was queried, and the evaluation can be re-run verbatim later if
retrieval quality is ever questioned again.

**Alternatives considered**: A `t.TempDir()`-style ephemeral fixture
generated only during a `go test`/`go run` invocation, discarded after.
Rejected — this evaluation is explicitly not a unit test asserting a
fixed pass/fail outcome (011-018's own tests already do that for the
underlying mechanics); it is a human-reviewed judgment call recorded in
prose (the Dogfooding Report), which needs the fixture to remain
inspectable alongside that report, not regenerated and thrown away.

## 4. No latency/index-size instrumentation is added

**Decision**: Elapsed time for a live Skill invocation (User Story
2/3) is measured externally (wall-clock, by whoever runs the session),
not through any new code. Index size is not measured at all in this
feature (spec.md's own Assumptions already deferred it).

**Rationale**: Directly follows spec.md's own Assumptions — no
numeric latency threshold exists yet to justify building measurement
infrastructure around, and the fixture's own index is far too small
for size to be an informative signal. Adding either now would be
speculative surface with no consuming requirement (Principle IV).

**Alternatives considered**: Adding a `--timing` flag or similar to
`internal context` (017). Rejected — no requirement in spec.md asks
for machine-readable timing; a human noting wall-clock time around a
live session is sufficient for FR-004/FR-005's own recording
requirement.

## 5. The Dogfooding Report's location and structure

**Decision**: One Markdown file, `specs/019-dogfooding-evaluation/
report.md`, structured as: one subsection per evaluated Spec (011-018)
under User Story 1, one subsection each for the Claude Code and
Antigravity live observations (User Story 2/3), a findings table
(Dogfooding Finding entities, per data-model.md), and one closing
Tuning Decision section (User Story 4).

**Rationale**: Matches spec.md's own Assumptions ("a Markdown document
produced alongside this feature's own planning/implementation
artifacts") and keeps the report co-located with the fixture it
describes and the spec it satisfies — a future reader finds evidence,
methodology, and conclusion in one place.

**Alternatives considered**: A structured JSON/YAML report instead of
prose Markdown. Rejected — this evaluation's own outputs are
inherently qualitative judgment calls (is this omission acceptable?
was this pack sufficient?) better expressed in reviewable prose than a
rigid schema; the *inputs* to those judgments (diagnostics counts) are
already machine-readable via 017's own JSON output, quoted directly
into the report where relevant.

## 6. No code change is planned up front

**Decision**: This feature's own Project Structure plans zero Go
source changes. A ranking/budgeting change is added to scope only if
User Story 4 (T-gated by FR-007/FR-008) actually identifies one,
handled as a follow-up task at that point, not designed speculatively
now.

**Rationale**: Designing a ranking change before evidence exists would
directly violate FR-008 and this feature's own governing rule (§30
Phase 9: "tune ranking only from observed failures, not intuition
alone"). The correct planning posture for evidence-gated work is to
plan the evidence-gathering fully and leave the contingent work
genuinely contingent.

**Alternatives considered**: Pre-emptively drafting a set of candidate
ranking tweaks "in case they're needed." Rejected — this is exactly
the intuition-driven tuning this feature exists to prevent.
