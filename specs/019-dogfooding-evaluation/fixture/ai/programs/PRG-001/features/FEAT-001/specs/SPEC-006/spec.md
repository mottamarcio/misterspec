---
id: SPEC-006
type: spec
status: ready
parent: FEAT-001
depends_on: []
supersedes: []
---

## Intent

Agent Adapter Layer: an `Adapter` interface, a `Registry`, and at
least one working adapter (Claude Code), so misterspec's canonical
Skills can be materialized into a coding agent's own integration
location without misterspec inventing a per-agent installation
mechanism each time (`docs/architecture-specification.md` §34-35).
