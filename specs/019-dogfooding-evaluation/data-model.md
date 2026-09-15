# Phase 1 Data Model: Dogfooding and Evaluation

No new Go types — this feature's only "entities" are the fixture
project's own artifact content and the Dogfooding Report's own
structure, both plain files.

## Fixture project entities (`specs/019-dogfooding-evaluation/fixture/`)

Standard misterspec artifact shapes (unchanged schema), populated with
real, condensed content from each already-shipped feature's own
`spec.md`/`plan.md` (research.md #1) — never invented:

| Entity | ID | `depends_on` | Real content source |
|---|---|---|---|
| Program | `PRG-001` | — | "Context Engine" — this repository's own §17 architecture-specification.md effort |
| Feature | `FEAT-001` | parent: `PRG-001` | Same |
| Spec | `SPEC-006` | `[]` | `specs/006-agent-adapter/spec.md`'s own Summary |
| Spec | `SPEC-011` | `[]` | `specs/011-wikilink-foundation/spec.md`'s own Summary |
| Spec | `SPEC-012` | `[SPEC-011]` | `specs/012-references-backlinks/spec.md`'s own Summary |
| Spec | `SPEC-013` | `[SPEC-011]` | `specs/013-document-model-chunking/spec.md`'s own Summary |
| Spec | `SPEC-014` | `[SPEC-012, SPEC-013]` | `specs/014-sqlite-index/spec.md`'s own Summary |
| Spec | `SPEC-015` | `[SPEC-011, SPEC-012, SPEC-013, SPEC-014]` | `specs/015-context-collector/spec.md`'s own Summary |
| Spec | `SPEC-016` | `[SPEC-015]` | `specs/016-ranking-budgeting/spec.md`'s own Summary |
| Spec | `SPEC-017` | `[SPEC-014, SPEC-015, SPEC-016]` | `specs/017-internal-context-command/spec.md`'s own Summary |
| Spec | `SPEC-018` | `[SPEC-006, SPEC-017]` | `specs/018-multi-agent-skill-integration/spec.md`'s own Summary |
| Knowledge | `KNOW-001` | — | Condensed from `docs/architecture-specification.md`'s own Constitution Principles (III, IV, IX) actually cited across 011-018's own Constitution Check tables |
| Knowledge | `KNOW-002` | — | Condensed from 018's own real research into each agent's real Skills-directory convention (research.md #1 of 018) |
| Learning | `LRN-001` | — | The real budget-flag design correction discovered during 017's own implementation (string flag + manual `strconv.Atoi`, not a Cobra `IntVar`) |
| Constitution | (project-level) | — | A short excerpt of the real `.specify/memory/constitution.md` Principles III/IV/IX, since those are the ones this evaluation's own Tier 0 checks matter for |

Wikilinks placed where the real feature's own work actually referenced
that Knowledge/Learning:

- `SPEC-018`'s body links `[[KNOW-002]]` (it really did require that
  research).
- `SPEC-017`'s body links `[[LRN-001]]` (the real correction happened
  during its own implementation).
- Every Spec's body links `[[KNOW-001]]` only where that feature's own
  real Constitution Check table actually invoked Principle III, IV, or
  IX by name (not all of them do).

## Dogfooding Finding (report entry, not a Go type)

One row per finding in `report.md`'s own findings table:

| Field | Meaning |
|---|---|
| Spec / Session | Which fixture Spec (User Story 1) or which live session (User Story 2/3) this finding came from |
| Kind | `omission` \| `over-inclusion` \| `sufficiency` \| `request-behavior` \| `elapsed-time` |
| Description | The specific, concrete observation |
| Evidence | The exact `internal context` JSON excerpt, or the specific observed agent action, this finding is based on |

## Tuning Decision (report section, not a Go type)

One final section in `report.md`:

- **Outcome**: `no change warranted` or a specific, named
  ranking/budgeting change.
- **Findings cited**: the exact Dogfooding Finding rows this decision
  is based on (by row reference) — empty only if the outcome is "no
  change warranted" with zero ranking-attributable findings.
