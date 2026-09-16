# Phase 0 Research: Constitution Frontmatter Guarantee, and Task Dependency/Parallelism Reporting

No `[NEEDS CLARIFICATION]` markers remained in the Technical Context. This file records the design decisions made while grounding the plan in the actual codebase.

## Decision: Fix the gap at both the Skill-prompt layer and the deterministic-validation layer

- **Decision**: Don't rely solely on improved wording in `create-constitution/SKILL.md` (which is still, ultimately, a request to a model that may or may not follow it precisely). Also extend `internal validate` — which `create-constitution` already calls at the end of its own Procedure — with a real, deterministic check for the Constitution's required frontmatter.
- **Rationale**: The reported bug (Claude omitted frontmatter, Antigravity didn't) is *exactly* the failure mode Constitution Principle I/II exists to prevent: a mechanical, checkable fact ("does this file's frontmatter contain `type`/`schema_version`?") left to model judgment instead of the deterministic layer. `docs/architecture-specification.md` §24 already defines the required frontmatter; `internal validate`/`ValidateProject` already runs exactly this class of check for every other artifact type via `checkEntity` — Constitution is the sole, explicit exception (`validator.go`'s own comment: "Task, Plan, Tasks, Validation, Constitution — mirroring Create's own supported-type boundary"). Closing that one exception makes FR-001's "regardless of which underlying model" guarantee real rather than aspirational.
- **Alternatives considered**:
  - *Prompt-only fix*: rejected as insufficient on its own — it's the same category of fix that already exists today (the current `SKILL.md` already relies on the model to structure the body correctly; frontmatter is no different in kind, so a prompt-only fix doesn't actually close the "regardless of model" gap the bug report is about).
  - *Route Constitution through `checkEntity` by giving it a synthetic EntityType*: rejected — `checkEntity` assumes an ID (`meta.ID`), a `Parent`/status lifecycle, and an `EntityType`-derived canonical filename, none of which the Constitution has (it has no ID at all, per `internal/ids/types.go`'s own doc comment). Forcing it through that path would require special-casing `checkEntity` internally anyway, which is no simpler than one small dedicated function.

## Decision: `checkConstitution` is additive to `ValidateProject`, not inserted into `projectEntityTypes`'s loop

- **Decision**: A new `checkConstitution(root string) []Finding` function, called once directly from `ValidateProject` (alongside, not inside, the existing `for _, t := range projectEntityTypes` loop and the existing Task-duplicate-ID check).
- **Rationale**: `projectEntityTypes`/`ids.Scan`/`checkEntity` are all built around `ids.EntityType`-bearing entities that live at scannable, ID-numbered paths (`ids.Scan(root, cfg, t)` returns `{number: [paths]}`). The Constitution is a single fixed-path file with no ID and no number to scan for — trying to fit it into that shape would be a bigger, riskier change to already-stable, tested code (`004-structural-validation`) for no benefit over one small standalone function.
- **Alternatives considered**:
  - *Add Constitution to `projectEntityTypes`*: rejected — `ids.Scan` has no case for a Constitution-like "singleton, no ID" type, and `sortedNumbers`/`validateFoundEntity`'s whole shape assumes a `map[int][]string]` keyed by entity number, which doesn't exist for a fixed-path singleton.

## Decision: A missing Constitution file produces no Finding

- **Decision**: `checkConstitution` returns no findings at all when `ai/memory/constitution.md` doesn't exist yet.
- **Rationale**: `create-constitution/SKILL.md`'s own Preconditions already document that a project may legitimately have no Constitution yet (before its first run); `internal validate` is also called by other Skills at times when a Constitution may not exist. Treating "file absent" as a structural problem would produce false-positive Findings for perfectly ordinary, expected project states — spec.md's own Edge Cases only describe *malformed/incomplete* frontmatter as the case to catch, never absence.
- **Alternatives considered**:
  - *Emit a warning-level Finding when absent*: rejected — `findings.go`'s own `Severity` type currently only ever produces `SeverityError` ("the type exists so a later feature can add warnings without a breaking change to Finding's shape" — its own doc comment); introducing the first-ever warning severity for this one case is out of scope and unrelated to the actual bug being fixed.

## Decision: Reuse existing Finding codes, add one new `Metadata` field

- **Decision**: A malformed/unparseable frontmatter block reports `CodeFrontmatterMalformed` (already used by `checkEntity` for the same underlying condition); a frontmatter block missing the required `schema_version` field reports `CodeRequiredFieldMissing` (already used by `ParseMetadata` for a missing `id`). `internal/artifacts.Metadata`/`frontmatterYAML` gain one new `SchemaVersion` field, parsed the same way existing optional fields (e.g. `Parent`) already are.
- **Rationale**: `findings.go`'s own Code vocabulary already covers both failure shapes exactly — inventing new codes for the same underlying condition (a required frontmatter field missing, or the block itself unparseable) would fragment the contract an agent already knows how to read, contradicting Constitution Principle IX (one consistent, already-documented vocabulary).
- **Alternatives considered**:
  - *A dedicated `CodeConstitutionFrontmatterInvalid` code*: rejected — the existing codes already describe the condition precisely enough (malformed vs. missing-required-field); a Constitution-specific code would only be justified if the failure mode were genuinely different in kind, which it isn't.

## Decision: `create-tasks`'s dependency/parallel reporting requires no new deterministic operation

- **Decision**: The Completion Contract addition is composed entirely from data the Skill's own existing Outputs already require it to record per Task (`depends_on` reference, `[P]` marker) — no new `internal` command, no new field in `tasks.md`'s own schema.
- **Rationale**: The information already exists in the artifact the Skill just wrote; the gap is purely that the completion *summary* doesn't surface it. This is prose composition from already-known data, the same category of work `implement`'s own `Recommended Next Step` section already does (composing exact next-invocation text from data it already has).
- **Alternatives considered**:
  - *A new `internal inspect --dependencies` style operation to compute a dependency graph deterministically*: considered, but rejected as premature per Constitution Principle IV — `create-tasks` already holds every Task's dependency data in memory the moment it finishes authoring `tasks.md`; no cross-invocation computation or graph algorithm complex enough to warrant a dedicated deterministic operation is involved (a handful of Tasks per Spec, dependency edges already explicit per Task, not inferred).
