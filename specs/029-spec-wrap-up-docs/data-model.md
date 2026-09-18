# Phase 1 Data Model: `/mister-wrap-up`

## Commit (new, returned by `internal/vcs.CommitsSinceFileAdded`)

| Field | Description |
|---|---|
| Hash | Full commit SHA |
| Subject | The commit's own first message line |
| AuthorDate | ISO-8601 date the commit was authored |

Not persisted — computed live from Git on each call, per Constitution Principle III.

## Spec Commit Range (derived, not a stored entity)

| Concept | Derivation |
|---|---|
| Range start | The commit that first added the named Spec's own `plan.md` (via `git log --diff-filter=A --follow`) |
| Range end | The current `HEAD` on whatever branch is checked out |
| Availability | A bool, separate from error: `false` when the project isn't a Git repository, or `plan.md` was never committed — the Skill still produces a document in that case, per spec.md FR-007 |

## Wrap-Up Document (the Skill's own output — not a canonical artifact)

| Field | Description |
|---|---|
| Path | `cortex/SPEC-###-<slug>.md`, at the project root |
| Slug | Derived from the Spec's own title, via the same slugification logic `internal/vcs`'s branch-name helper already uses (generalized, not duplicated) |
| Regeneration | Wholesale replace on every run — never amended in place, unlike every other canonical Skill's own output (spec.md FR-009, plan.md's own documented exception to Constitution Principle VII) |

Unlike every other artifact schema in `docs/architecture-specification.md` §23–31, the Wrap-Up Document has **no required frontmatter contract** and is **not validated by `internal validate`** — it is downstream documentation material, not project state.

## Wrap-Up Document Sources (what it draws on — spec.md FR-004)

| Source | How it's read |
|---|---|
| Spec's own artifacts | `spec.md`, `plan.md`, `tasks.md`, `validation.md` (if it exists) — via `internal resolve`/`internal inspect`, same as every other Skill |
| Commit history | The Spec Commit Range, above |
| Knowledge base | `internal inventory knowledge` + reading whichever Knowledge artifacts the agent judges relevant — a semantic call, not a structured query (Knowledge has no field linking it to a specific Spec) |
| Constitution | `ai/memory/constitution.md`, read directly, for invariants the Spec's own implementation touched |
| Learnings | `ai/memory/learnings/LRN-*.md`, read directly — relevance is a semantic judgment call, since Learning's own schema (§31) carries no `for`/`parent` field linking it to a Spec |

## State Transitions (Wrap-Up Document)

```text
no wrap-up document for this Spec ──(first /mister-wrap-up run)──> document created
document exists ──(/mister-wrap-up run again, any time)──> document fully replaced with a fresh one
```

No intermediate "amend" state — unlike Constitution/Knowledge/etc., this is by design (research.md's own decision).
