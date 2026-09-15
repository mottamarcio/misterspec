# Evaluation Protocol: Dogfooding and Evaluation

Not a code contract (this feature ships no new interface) — the
step-by-step protocol every User Story follows, so the evaluation is
reproducible by anyone re-running it later.

## User Story 1 protocol (static, no live agent needed)

For each `SPEC-{006,011..018}` in the fixture:

```sh
misterspec internal context SPEC-0## --intent planning --dir specs/019-dogfooding-evaluation/fixture
```

1. Record the full JSON response in `report.md`.
2. For each ID in that Spec's own real `depends_on` (data-model.md's
   table), confirm it (or its own chunk) appears among `context.items`.
   Missing → an `omission` Finding.
3. Review every other returned item; anything judged irrelevant to
   that Spec's own real work → an `over-inclusion` Finding.
4. Note `context.diagnostics` (candidates considered, reduction
   percent) in the report regardless of outcome.

## User Story 2/3 protocol (live agent session)

1. Ensure the fixture has both `claude-code` and `agy` Skills
   installed (`internal/agents`' own adapters, 018).
2. Through Claude Code, invoke a Context-Pack-Aware Skill (e.g.
   `/create-plan SPEC-018`) against the fixture.
3. Directly observe and record: did it call `internal context` before
   other exploration? was the returned pack sufficient, or did the
   agent read further files? wall-clock elapsed time for the whole
   invocation.
4. Repeat step 2-3 through Antigravity, same or an equivalent Spec.
5. Record any material difference between the two agents' own
   behavior.

## User Story 4 protocol

1. List every Finding from User Story 1-3.
2. For each, judge: does it point at a specific ranking/budgeting
   behavior (016), or is it explained by something else (a missing
   wikilink in the fixture itself, a collection gap in 015, an
   agent-specific quirk unrelated to retrieval)?
3. Record the Tuning Decision (data-model.md) — a change only if at
   least one Finding is genuinely ranking-attributable (FR-008).

## Non-goals of this protocol

- Does not modify 011-018's own shipped code as part of running the
  protocol itself — only as a possible, separately-tracked, evidence-
  gated follow-up from the Tuning Decision.
- Does not measure index size or enforce a latency threshold
  (research.md #4).
- Does not attempt a live session through Codex CLI, GitHub Copilot,
  Cursor, or Devin for Terminal (spec.md FR-006).
