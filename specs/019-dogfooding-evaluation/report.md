# Dogfooding Report: Context Engine and Skill Integration

**Feature**: [spec.md](./spec.md) | **Fixture**: [fixture/](./fixture/)
**Date**: 2026-09-15

This report records every finding from User Stories 1-3 and the
resulting Tuning Decision (User Story 4), per
`contracts/evaluation-protocol.md`. Every finding below is traceable
to either a raw `internal context` JSON response (quoted or
summarized) or a directly-observed live Skill invocation.

---

## User Story 1 — Static Evaluation Against Real History

Protocol: `misterspec internal context <id> --intent planning --dir
fixture` for each of the 9 fixture Specs, compared against that
feature's own real, historical `depends_on` (data-model.md).

### Raw responses

Full JSON responses for all 9 Specs, plus two additional
budget-constrained runs on `SPEC-017` (`--budget 300` and `--budget
700`), were captured during implementation. Summarized per-Spec below;
full raw JSON is reproducible verbatim via `quickstart.md`'s own loop.

### Per-Spec results

| Spec | Real `depends_on` | All present in pack? | `candidates_considered` | `tokens_available` | `reduction_percent` |
|---|---|---|---|---|---|
| SPEC-006 | *(none)* | N/A (no dependency to check) | 11 | 918 | 0% |
| SPEC-011 | *(none)* | N/A | 12 | 762 | 0% |
| SPEC-012 | SPEC-011 | ✅ yes (`tier: structural`) | 12 | 762 | 0% |
| SPEC-013 | SPEC-011 | ✅ yes | 12 | 762 | 0% |
| SPEC-014 | SPEC-012, SPEC-013 | ✅ yes, both | 15 | 1155 | 0% |
| SPEC-015 | SPEC-011, 012, 013, 014 | ✅ yes, all four | 15 | 1155 | 0% |
| SPEC-016 | SPEC-015 | ✅ yes | 15 | 1155 | 0% |
| SPEC-017 | SPEC-014, 015, 016 | ✅ yes, all three | 18 | 1538 | 0% |
| SPEC-018 | SPEC-006, SPEC-017 | ✅ yes, both | 15 | 1346 | 0% |

**Zero omissions across all 9 Specs** — every real, historically-known
dependency is present in its own Context Pack, correctly classified as
`tier: structural` with `reasons: ["depends_on"]` (SC-001, SC-002).
Every direct wikilink (`SPEC-015`→`KNOW-001`, `SPEC-016`→`KNOW-001`,
`SPEC-017`→`KNOW-001`+`LRN-001`, `SPEC-018`→`KNOW-001`+`KNOW-002`) is
also present, correctly classified as `tier: semantic`.

### Findings

| # | Spec / Session | Kind | Description | Evidence |
|---|---|---|---|---|
| F1 | SPEC-006 | over-inclusion | Second-hop expansion pulled in `SPEC-017` (Internal Context Command) — a Spec built two features later, with no real conceptual bearing on `SPEC-006`'s own historical (Agent Adapter) work. Reached via `SPEC-006`'s own backlink `SPEC-018` → `SPEC-018`'s own `depends_on: SPEC-017`. | `SPEC-006` response: `{"heading":"Intent","path":".../SPEC-017/spec.md","reasons":["depends_on"],"tier":"second_hop", ...}` |
| F2 | SPEC-014 | over-inclusion | Second-hop expansion pulled in `SPEC-016` (Ranking and Budgeting) via `SPEC-014`'s own backlink `SPEC-015` → `SPEC-015`'s own `depends_on: SPEC-016`(sic, actually the reverse chain) — a Spec built after `SPEC-014`, no real historical bearing. | `SPEC-014` response: `{"heading":"Intent","path":".../SPEC-016/spec.md","reasons":["depends_on"],"tier":"second_hop", ...}` |
| F3 | All 9 Specs | sufficiency (budget) | With the **default** budget (6000), `reduction_percent` is 0% for every single Spec — the fixture's total content (762-1538 tokens per Spec) never approaches the default budget, so **budgeting was never actually exercised** by the default-budget runs. F1/F2's own "over-inclusion" is a lowest-priority (`TierSecondHop`) artifact that a real, tighter budget trims first by design (see F4) — not evidence of a ranking defect on its own. | `reduction_percent: 0` in all 9 raw responses |
| F4 | SPEC-017 (`--budget 700`) | sufficiency (budget) | Under a realistic, tighter budget, all three real structural dependencies (`SPEC-014`, `SPEC-015`, `SPEC-016`) survive in full; every semantic/second-hop item (including F1/F2-style over-inclusions) is trimmed first, exactly as 016's own tier-ordering guarantee promises. | `budget:700` response: 9 items selected, all `mandatory`/`structural`; `reduction_percent: 61.7%` |
| F5 | SPEC-017 (`--budget 300`) | sufficiency (budget) | When budget can't even fit mandatory content (307 tokens > 300), every optional item — including the three real structural dependencies — is correctly, entirely trimmed, `budget_exceeded: true`, `overage: 7` reported accurately; mandatory content itself is still returned in full. | `budget:300` response |
| F6 | All 9 Specs | sufficiency (query) | No `--query`/`--task` was supplied in any User Story 1 run (only `--intent planning`), so the `TierText`/BM25 lexical-relevance path was never exercised by this evaluation — this static pass validates the structural/semantic/second-hop graph traversal only, not free-text retrieval quality. | No `text_match` reason appears in any of the 9 raw responses |

**No true omissions were found** — every finding above is an
over-inclusion or a scope/coverage observation, never a missing real
dependency.

---

## User Story 2 — Claude Code Live Session

Protocol: following `create-plan`'s own installed instructions
(`fixture/.claude/skills/create-plan/SKILL.md`) against `SPEC-018`.

**Observation** (this assistant, running as Claude Code, followed the
installed Skill's own instructions verbatim against the fixture):

1. `internal resolve SPEC-018` and `internal inspect SPEC-018` were run
   first, per the Skill's own Procedure step 1.
2. `internal context SPEC-018 --intent planning` was run next
   (Procedure step 2, added by 018) — **before** any other repository
   exploration, confirming the Skill's own instructions are followed
   as written.
3. **Sufficiency**: the returned pack already contained `SPEC-006`'s
   and `SPEC-017`'s own real Intent (the two dependencies a Plan for
   `SPEC-018` genuinely needs), plus `KNOW-002`'s own real per-agent
   integration research and `KNOW-001`'s own relevant Constitution
   principles. For this fixture's own condensed content, the pack
   was **sufficient on its own** — no further repository exploration
   was genuinely needed to draft `SPEC-018`'s own Plan Summary and
   Requirement Coverage at this fixture's level of detail.
4. **Elapsed time**: the three-command sequence (`resolve`, `inspect`,
   `context`) completed in well under one second wall-clock (CLI
   round-trips against a 9-Spec fixture) — negligible relative to any
   human or agent reading/reasoning time.

---

## User Story 3 — Antigravity Live Session

**Status**: deferred, by explicit user decision, given User Story 2's
own Claude Code observation already confirmed the integration works
end-to-end (`internal context` requested first, pack sufficient) and
both Claude Code and Antigravity install and run the byte-identical
canonical Skill content (018) through the byte-identical `internal
context` command (017) — the source of any real cross-agent
divergence would be agent-specific instruction-following, not
anything this feature's own scope (the Context Engine, Skill content)
could differ on per agent. `fixture/.agents/skills/` remains installed
and ready (User Story 2's own Foundational step, T016) if this is
ever revisited later.

No finding is recorded for this story — it is explicitly out of scope
for this evaluation pass, not a gap in the evidence gathered.

---

## User Story 4 — Tuning Decision

### Classification of every finding

| Finding | Ranking-attributable? | Reasoning |
|---|---|---|
| F1, F2 (over-inclusion) | **No** | Both are `TierSecondHop` — 016's own tier-ordering guarantee already places this content strictly below mandatory/structural/semantic; F4 directly confirms a realistic, tighter budget trims exactly this content before touching any real dependency. The apparent over-inclusion is a property of the fixture being smaller than the default budget (F3), not of the ranking/budgeting algorithm itself. |
| F3 (budget never exercised at default) | **No** | This is a property of the fixture's own small scale relative to `DefaultBudget = 6000`, not a defect — 016's own budgeting logic is directly exercised and confirmed correct by F4/F5 once a realistic, tighter budget is supplied. |
| F4, F5 (budget-constrained behavior) | **N/A (positive confirmation)** | These findings *confirm* 016's own ranking/budgeting behaves exactly as designed (tier ordering respected, mandatory content never dropped, overage accurately reported) — supporting evidence for "no change needed," not evidence for a change. |
| F6 (text-tier untested) | **No** | A test-coverage gap in this evaluation's own protocol (no `--query` was supplied), not a finding about ranking behavior itself. Recorded so a future evaluation pass can extend User Story 1 with query-bearing requests. |

### Outcome (final)

**No ranking or budgeting change is warranted by the evidence
gathered.** Every dependency this project's own real feature history
actually used is retrieved, correctly tiered, and correctly preserved
under realistic budget pressure (F4/F5); the only "over-inclusion"
observed (F1/F2) is explained entirely by the evaluation's own default
budget being generous relative to a small fixture (F3), not by any
flaw in 016's tier-ordering or budgeting logic — F4 directly
demonstrates that a realistic budget already resolves it by design.

User Story 3 (a second live agent observation, through Antigravity)
was deliberately deferred by explicit user decision, since User Story
2 already confirmed the integration works correctly end-to-end and
both agents run the identical, agent-independent Context Engine and
Skill content this decision is actually about. This decision is
therefore finalized on User Story 1 and User Story 2's own evidence
alone — findings F1-F6 remain the complete evidentiary basis, with
zero findings attributable to ranking/budgeting behavior across all of
them (FR-007, FR-008 satisfied: no change is made because none is
justified by the evidence gathered, not because evidence-gathering was
skipped).
