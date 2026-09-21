# Phase 0 Research: End-to-End Quality and Efficiency Evaluation

No `NEEDS CLARIFICATION` markers remained in the Technical Context —
every decision below was resolvable from this project's own established
precedent (019-dogfooding-evaluation, 033/035/036's JSON-contract
discipline, the Constitution). Each entry records the decision made and
why, so a later maintainer can see it was deliberate, not default.

## 1. Where does live agent task-execution telemetry come from?

**Decision**: The binary does not capture it directly. A `RunRecord`
file is a structured, hand-authored (or agent-authored, under the
running agent's own dogfooding discipline) artifact filled in after a
live session completes, using whatever usage figures that session's
own tooling reports (e.g. Claude Code's own reported token/cost
summary). `misterspec internal eval-compare` only ever *consumes*
already-written `RunRecord` files; it never launches a session.

**Rationale**: Spec 019 already proved this pattern works — the
assistant directly observed and recorded its own Skill invocation
behavior into `report.md` without any new tooling. Automating live
session capture would require the Go binary to reach into another
program's process (a different coding agent's CLI/IDE) — architecture
this project has never taken on, and Principle IV forbids adding a
machine command for what is fundamentally an external orchestration
decision, not a computable one.

**Alternatives considered**:
- A wrapper process that shells out to `claude` or similar and scrapes
  its output — rejected: fragile across agent versions, and the four
  other installed agents (018) have no common invocation surface to
  standardize against; would also violate the "no network in core
  commands" architecture constraint for commands other than the
  existing `--update` exception.
- Requiring provider API keys and calling providers' usage APIs
  directly — rejected: adds real external dependencies and
  credentials handling to a project whose core constraint is a
  network-free core binary; explicitly out of scope per the spec's
  own Assumptions ("this feature does not require live billing-account
  integration").

## 2. Where do retrieval evaluation cases live, and in what format?

**Decision**: YAML case files under `specs/037-eval-quality-efficiency/
fixture/cases/`, one file per case, each declaring `id`, `query` (free
text) or `task`/`intent`, `required` (paths or artifact IDs expected in
the pack), `forbidden` (optional), and `budget` (optional override).
The fixture corpus itself (a small multi-Spec project tree) lives
alongside, following the exact shape `internal/context/fixture_test.go`
and 019's own `fixture/` already use.

**Rationale**: Reuses an established, already-tested fixture pattern
(019, `internal/context`'s own test fixtures) instead of inventing a
new one; YAML matches how `misterspec` already represents structured,
hand-authored config (frontmatter, `config.yaml`).

**Alternatives considered**: JSON case files — rejected only as the
primary authoring format (harder to hand-edit/review in PRs than
YAML); the CLI's own output remains JSON per Principle IX, so this is
an authoring-format choice only, not a contract choice.

## 3. Where do recorded run records and baselines live?

**Decision**: A project-level `eval/` directory (`eval/runs/*.json`,
`eval/baselines/*.json`), committed to the repository like any other
durable record — not under `ai/` (reserved for product artifacts per
the Constitution's Architecture Constraints) and not under `specs/037-
.../` (that directory documents *this feature*, not every future
evaluation run other Specs will record after 037 ships).

**Rationale**: Baselines and run records need to outlive Spec 037
itself — a maintainer evaluating a 042-some-future-change proposal in
six months records against a `eval/baselines/retrieval-2026-09.json`
baseline, not against spec 037's own directory. Keeping them at the
project root, sibling to `internal/`/`specs/`, matches how `.misterspec/`
already holds project-level (not feature-level) state.

**Alternatives considered**: Storing run records inside each
proposing Spec's own directory — rejected: makes baseline comparison
across Specs (FR-012) awkward, since the "current baseline" would have
no single stable location to point at.

## 4. How is "correctly completed" determined for an agent task?

**Decision**: Exit code / pass-fail of a task-supplied, independent
acceptance-test command (e.g. `go test ./...` for a Go-code task, or a
specific test file), recorded as-is in the `RunRecord`. The harness
never infers correctness from the agent's own narrative output.

**Rationale**: Directly matches spec FR-002's own requirement and this
project's Constitution Principle IX distinction between `ok` (command
executed) and `valid`/`result` (semantic outcome) — the acceptance
test's exit code is the `result`, never the agent's self-report.

**Alternatives considered**: LLM-judged correctness (an "LLM as
judge") — explicitly rejected by the spec itself (FR-002: "never the
executing agent's own self-report"); an LLM judge is still a
self-report risk one layer removed, not an independent check.

## 5. How many repetitions, by default, for agent task-execution runs?

**Decision**: 3, configurable per run; not hard-coded as the only
allowed value.

**Rationale**: Matches the spec's own Assumptions section ("a
reasonable default repetition count... is small (on the order of 3)");
low enough to keep live-session cost bounded for routine use, high
enough to distinguish a real effect from single-run noise per FR-013.

**Alternatives considered**: A fixed, non-configurable count —
rejected: FR-004 requires the harness to *support* running each case
multiple times, implying the count is a run parameter, not a constant.

## 6. How does a `RunRecord`/comparison represent an "estimated" figure?

**Decision**: Every numeric effort field (`input_tokens`,
`output_tokens`, `cached_tokens`, `cost`, etc.) is paired with a
sibling boolean `*_estimated` field (or a single `estimated: true` at
the metric-group level when an entire run used a non-telemetry
provider), never a separate, easy-to-drop side-channel note.

**Rationale**: Directly satisfies FR-008 and SC-004 — the label must
be structurally attached to the figure itself so it cannot be silently
dropped when figures are aggregated or copied into a comparison
report, continuing this project's existing "structured field over
prose note" contract discipline (Principle IX).

**Alternatives considered**: A free-text `notes` field describing
which figures are estimated — rejected: not machine-checkable, and
this project's contracts consistently prefer structured fields
(`ok`/`valid`, `tier`/`reasons`) over prose for exactly this reason.
