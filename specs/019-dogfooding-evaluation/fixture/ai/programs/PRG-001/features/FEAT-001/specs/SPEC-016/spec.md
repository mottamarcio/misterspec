---
id: SPEC-016
type: spec
status: ready
parent: FEAT-001
depends_on:
  - SPEC-015
supersedes: []
---

## Intent

Ranking and Budgeting: assign every candidate a deterministic score
used only to order candidates already sharing a Tier — Tier always
dominates Score — then fit the ranked list into a token budget,
building directly on SPEC-015's own deduplicated, tier-labeled
candidate set. See [[KNOW-001]] for the Constitution principles this
budgeting process must hold to.
