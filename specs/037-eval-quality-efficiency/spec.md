# Feature Specification: End-to-End Quality and Efficiency Evaluation

**Feature Branch**: `037-eval-quality-efficiency`
**Created**: 2026-09-21
**Status**: Draft
**Input**: User description: "PROP-07 — Avaliação de qualidade e eficiência de ponta a ponta: demonstrar se o MisterSpec reduz esforço total mantendo qualidade, ampliando a avaliação atual baseada em uma fixture pequena." (Backlog proposal from `tmp/misterspec-specs-sugeridas.md`: build a repeatable evaluation harness — deterministic retrieval evaluation plus agent task-execution evaluation — that measures whether MisterSpec reduces total effort while preserving quality, replacing the one-off, small-fixture dogfooding exercise from Spec 019 with a reproducible, comparable, and reusable measurement capability.)

## User Scenarios & Testing *(mandatory)*

<!--
  This feature's "users" are misterspec's own maintainers and
  contributors who need durable, reproducible evidence — not a one-off
  exercise — before promoting a retrieval, ranking, or budgeting change
  (e.g. 036's BM25 weighting, or future proposals) to become the
  default behavior. Spec 019 already established that ranking changes
  must be justified by observed evidence rather than intuition; this
  feature turns that one-time exercise into a standing, repeatable
  measurement capability that any future proposal can run against.
-->

### User Story 1 - Run a Deterministic Retrieval Evaluation (Priority: P1)

As a maintainer preparing to change how context is selected or ranked,
I want to run a fixed, versioned set of retrieval cases against a
known repository state and get a pass/fail and score report for each
case, so I can confirm a proposed change does not silently drop
content that must be present or promote irrelevant content ahead of
it — without needing a live LLM session.

**Why this priority**: This is the only evaluation that requires no
LLM session, no cost, and no external variability — it can run in CI
on every change and is the foundation every other measurement in this
feature builds on. Without it, there is no cheap, fast gate at all.

**Independent Test**: Can be fully tested by running the retrieval
evaluation command against the fixture corpus and confirming it
reports, for each case, whether the required content was returned, in
what position, and with what score — entirely offline and
deterministic, reproducible byte-for-byte across repeated runs on the
same input.

**Acceptance Scenarios**:

1. **Given** a fixed evaluation case with known required content,
   **When** the retrieval evaluation is run twice against the same
   repository state, **Then** both runs report identical results.
2. **Given** an evaluation case whose required content is missing from
   the returned results, **When** the evaluation runs, **Then** that
   case is reported as failed with the specific missing item named,
   not merely an aggregate score.
3. **Given** a change that reorders results without dropping required
   content, **When** the evaluation runs before and after the change,
   **Then** the report shows the ranking-position difference for each
   affected case, not just pass/fail.

---

### User Story 2 - Run an Agent Task-Execution Evaluation (Priority: P2)

As a maintainer deciding whether MisterSpec actually reduces the total
effort an agent spends completing a real task, I want to run a fixed
set of representative tasks through a documented reference flow and
through MisterSpec, using independent acceptance tests to judge
correctness, and get back success rate and effort figures (tokens,
calls, extra reads, rework, latency, cost) for each, so I can compare
them on the same basis instead of relying on impressions from a single
session.

**Why this priority**: This is the measurement that actually answers
the feature's central question — whether MisterSpec reduces effort
while preserving quality — but it depends on a live LLM session, so it
costs more and runs less often than User Story 1. It builds directly
on User Story 1: a retrieval regression found there explains an
effort regression found here.

**Independent Test**: Can be fully tested by selecting one evaluation
task, running it to completion under both the reference flow and a
MisterSpec variant, checking the result against that task's
independent acceptance test, and confirming a report is produced with
success/fail and the measured effort figures for both runs.

**Acceptance Scenarios**:

1. **Given** an evaluation task with an independent acceptance test,
   **When** an agent run completes, **Then** the run is marked correct
   only if the acceptance test passes, never by the agent's own
   self-report.
2. **Given** a task run that fails partway or does not complete,
   **When** the results are compiled, **Then** that attempt is
   recorded (not discarded) and counted in the total-tokens-per-
   correctly-completed-task ratio's denominator context, alongside the
   reported success rate.
3. **Given** the same task and configuration run multiple times,
   **When** results vary between repetitions, **Then** the report
   shows the individual repetitions, not only an average that hides
   the variance.
4. **Given** a run using a provider that does not expose token
   telemetry, **When** effort figures are reported for that run,
   **Then** every such figure is explicitly labeled as estimated and
   is never combined with measured figures without that label.

---

### User Story 3 - Compare a Proposed Change Against the Recorded Baseline (Priority: P3)

As a maintainer who has just implemented a retrieval, ranking, or
budgeting change, I want to compare its evaluation results against the
last recorded baseline, changing only that one dimension, and see
whether it improves, regresses, or has no measurable effect on quality
and effort, so I can decide whether to promote the change with
evidence instead of intuition.

**Why this priority**: This is the governance payoff of Stories 1 and
2 — without a recorded baseline and a comparison view, every
individual run is just an isolated data point. It depends on both
prior stories already producing comparable, reproducible results to
compare against.

**Independent Test**: Can be fully tested by recording one baseline
run, applying a single isolated change, running the same evaluation
cases again, and confirming the comparison output names which cases
changed, in which direction, and by how much, distinguishing a real
regression from normal repetition variance.

**Acceptance Scenarios**:

1. **Given** a recorded baseline and a candidate run that changed
   exactly one dimension, **When** they are compared, **Then** the
   comparison explicitly lists every case whose outcome changed and in
   which direction.
2. **Given** a candidate run whose configuration differs from the
   baseline in more than the one intended dimension, **When** a
   comparison is requested, **Then** the extra difference is flagged
   so the comparison is not silently treated as isolating one variable
   when it is not.
3. **Given** a candidate run that regresses a case that was previously
   passing, **When** the comparison is produced, **Then** that
   regression is visibly surfaced and cannot be hidden by an improved
   aggregate average elsewhere.

---

### Edge Cases

- An evaluation case whose required content legitimately does not
  exist in the corpus at the configured budget (e.g. it was
  deliberately excluded by design) — must not be reported as a
  retrieval failure; the case's own expectation must reflect that.
- An agent run that is aborted or errors for a reason unrelated to
  MisterSpec (e.g. an unrelated environment failure) — recorded as
  inconclusive, not silently counted as either a pass or a failure.
- A comparison where the baseline itself is stale (recorded against a
  repository state or contract version that no longer matches) — the
  comparison must surface that the baseline needs refreshing rather
  than produce a misleading delta.
- A task whose acceptance test itself is ambiguous or flaky across
  repetitions — repetitions must expose this as inconsistent results
  for that case rather than averaging it away.
- Two evaluation runs using different model versions or provider
  configurations being compared directly — must be flagged, since the
  configuration is part of what makes a comparison valid.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The evaluation harness MUST provide a deterministic
  retrieval evaluation set, runnable without any live LLM session,
  where each case declares its required and (optionally) forbidden
  content and produces the same result on repeated runs against the
  same repository state.
- **FR-002**: The evaluation harness MUST provide an agent
  task-execution evaluation set, where each task specifies a fixed
  repository state, a task description, and an independent acceptance
  test used to judge correctness — never the executing agent's own
  self-report.
- **FR-003**: Every agent task-execution evaluation run MUST record
  which repository revision, task, model, and configuration were used,
  so the run can be reproduced or audited later.
- **FR-004**: The harness MUST support running each agent
  task-execution case multiple times and MUST report each repetition's
  own result, not only an aggregate.
- **FR-005**: The harness MUST support comparing a documented reference
  flow (a baseline not using MisterSpec's context selection) against
  one or more MisterSpec variants over the same tasks.
- **FR-006**: A comparison between two runs MUST change exactly one
  configuration dimension at a time to be treated as isolating that
  dimension's effect; the harness MUST flag when more than one
  dimension differs between the two runs being compared.
- **FR-007**: The harness MUST measure, for each agent
  task-execution run: success/failure per the independent acceptance
  test, input tokens, output tokens, cached tokens where the provider
  exposes them, number of calls, number of additional reads beyond
  what MisterSpec supplied, rework (repeated or corrective actions),
  latency, and cost.
- **FR-008**: Any effort figure produced for a run using a provider
  that does not expose the underlying telemetry MUST be explicitly
  labeled as estimated, and MUST NOT be combined with measured figures
  from another run without that label being visible in the resulting
  report.
- **FR-009**: The harness MUST record every attempt, including failed
  or incomplete ones, and MUST report both the overall success rate
  and the total tokens of the evaluated set divided by the number of
  correctly completed tasks, presented together — never the ratio
  alone.
- **FR-010**: The retrieval evaluation set (FR-001) MUST include cases
  drawn from a larger project shape than a single small fixture,
  including at least one large specification, free-text query cases,
  cases with deliberately irrelevant reference material present, and
  cases exercised under real context-budget pressure.
- **FR-011**: The harness MUST produce, for each run, a durable,
  reviewable report distinguishing case-level results from aggregate
  results, so a specific regression cannot be hidden inside an
  improved average.
- **FR-012**: The harness MUST support recording a run as a named
  baseline and comparing any later run against a named baseline,
  reporting which cases changed outcome and in which direction.
- **FR-013**: A comparison MUST distinguish a real change in outcome
  from normal repetition-to-repetition variance, using the multiple
  repetitions recorded under FR-004.
- **FR-014**: The deterministic retrieval evaluation (FR-001) MUST be
  runnable in continuous integration on every relevant change; the
  agent task-execution evaluation (FR-002) MUST be runnable as a
  separate, explicitly-triggered operation with its own resource
  budget, since it depends on live LLM sessions and incurs real cost.
- **FR-015**: The harness MUST NOT be used to justify a retrieval,
  ranking, or budgeting change being promoted to default behavior
  unless that change's evaluation results are compared against a
  recorded baseline under this feature — continuing the evidence-only
  rule already established for ranking changes.

### Key Entities

- **Evaluation Case**: One deterministic retrieval check — a fixed
  query/task context, its required and forbidden content, and the
  repository state it is checked against.
- **Evaluation Task**: One agent task-execution scenario — a fixed
  repository state, task description, and independent acceptance test
  used to judge whether an agent run completed it correctly.
- **Evaluation Run**: One execution of the retrieval or task-execution
  evaluation set under a specific, recorded configuration (repository
  revision, model, variant, repetition count).
- **Metric Record**: The set of measured or estimated effort and
  quality figures (success, tokens, calls, extra reads, rework,
  latency, cost) produced by one Evaluation Run, per case or task.
- **Baseline**: A named, recorded Evaluation Run designated as the
  comparison reference for later runs.
- **Comparison Report**: The output of comparing a candidate
  Evaluation Run against a Baseline, naming every case or task whose
  outcome changed, its direction, and whether the comparison isolates
  a single configuration dimension.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A maintainer can run the deterministic retrieval
  evaluation and receive a case-by-case report in under 2 minutes on
  the project's current corpus, with identical results across repeated
  runs on the same repository state.
- **SC-002**: A maintainer can record one baseline and compare a
  candidate run against it, receiving a report that names every
  changed case, in under 5 minutes of review time (excluding the time
  the underlying agent runs themselves take to execute).
- **SC-003**: 100% of agent task-execution attempts recorded by the
  harness — including failed or incomplete ones — appear in the
  resulting report; none are silently dropped from the totals.
- **SC-004**: 100% of effort figures derived from a provider without
  token telemetry are visibly labeled as estimated in every report
  that includes them.
- **SC-005**: A regression that drops previously-passing retrieval or
  task-execution content is visible in the comparison report 100% of
  the time it occurs, even when an unrelated aggregate metric improves
  in the same comparison.
- **SC-006**: At least one retrieval or ranking change proposal can be
  evaluated end-to-end through this harness (baseline recorded,
  candidate run, comparison produced) without any manual, ad hoc
  measurement step outside the harness.

## Assumptions

- This feature supersedes the one-off dogfooding exercise from Spec
  019 with a repeatable, reusable harness; it does not require
  re-running 019's own already-completed, one-time findings.
- Agent task-execution runs (User Story 2) are scoped, by default, to
  the coding agents the maintainer can personally exercise (Claude
  Code, and Antigravity where available), consistent with the access
  constraint already established in Spec 019; extending to other
  agents is future work, not blocked scope for this feature.
- "Correctly completed" for an agent task-execution run means the
  task's own independent acceptance test passes; a passing acceptance
  test is treated as the ground truth for correctness, not the agent's
  own report of success.
- A reasonable default repetition count for agent task-execution runs
  is small (on the order of 3) unless a specific comparison needs more
  to distinguish a real effect from variance; the harness does not
  mandate a single fixed count for every case.
- The deterministic retrieval evaluation set (FR-001/FR-010) is
  expected to grow incrementally as new retrieval-affecting features
  (e.g. 036's ranking work, future wikilink or budget changes) land;
  this feature establishes the harness and an initial case set, not a
  permanently fixed one.
- Cost figures (FR-007) are computed from published provider pricing
  applied to measured or estimated token counts; this feature does not
  require live billing-account integration.
- The Comparison Report and recorded Baselines (Key Entities) are
  durable artifacts read by maintainers directly; this feature does
  not require a dashboard or UI beyond a reviewable report format.
- This feature does not itself change any retrieval, ranking, or
  budgeting behavior — it is exclusively a measurement capability, per
  its own FR-015 evidence-gating rule.
