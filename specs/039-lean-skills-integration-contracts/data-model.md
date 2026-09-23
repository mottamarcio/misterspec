# Data Model: Skills Enxutas e Contratos de Integração Testáveis

This feature adds development-time authoring/verification structures,
not new project artifact types. No new entity gets an ID, a canonical
path under `ai/`, or a lifecycle stage (Constitution's "Canonical
lifecycle" is unaffected).

## Fragment

A named, canonical block of Skill instruction text that recurs
verbatim (or near-verbatim, unified into one wording) across two or
more of the 10 canonical Skills today (research.md #1).

| Field | Type | Notes |
|---|---|---|
| `Name` | string | Stable identifier, e.g. `resolve-preamble`, `mechanical-steps-note`, `fallback-tolerance-note`. Referenced by manifests, never duplicated inline. |
| `Section` | string | Which of the 26 §39 headings this fragment belongs under (a fragment never spans sections). |
| `Body` | string | The literal Markdown text, authored once. |

**Validation rules**: `Name` MUST be unique. `Body` MUST NOT itself
reference another `Name` (no nested composition — keeps generation a
single flat substitution pass, Principle IV).

## SkillManifest

One per canonical Skill (`mister-implement`, `mister-plan`, …),
describing how its final `SKILL.md` is composed.

| Field | Type | Notes |
|---|---|---|
| `SkillName` | string | Matches the `kit/skills/<name>/` directory and the `SKILL.md` frontmatter `name:`. |
| `Sections` | ordered list of `SectionEntry` | One entry per §39 heading, in the required order (matches `requiredSkillHeadings` already asserted by `skills_content_test.go`). |

### SectionEntry

| Field | Type | Notes |
|---|---|---|
| `Heading` | string | One of the 26 required headings. |
| `Parts` | ordered list of `Part` | Concatenated (each on its own paragraph) to form the section body. |

### Part

A tagged union: exactly one of:
- `FragmentRef string` — the `Fragment.Name` to insert verbatim.
- `Bespoke string` — Skill-specific Markdown authored only for this
  Skill (the common case — most section bodies are 100% bespoke, per
  research.md #1).

**Validation rules**: every `FragmentRef` MUST resolve to a known
`Fragment` whose `Section` matches the enclosing `SectionEntry.Heading`
(a fragment cannot be silently reused under the wrong heading). A
`SkillManifest`'s `Sections` list MUST cover exactly the 26 required
headings, in order — the same invariant `skills_content_test.go`
already enforces on the *generated output*; the manifest enforces it
one level earlier, at authoring time.

## GeneratedSkill

The pure-function output of composing one `SkillManifest` against the
current `Fragment` set: the exact bytes written to
`kit/skills/<name>/SKILL.md`. Not a stored type — a return value of
`internal/skillgen.Generate(manifest, fragments) ([]byte, error)`,
asserted byte-for-byte equal to the committed file by the new
drift-check test (`internal/example/skillgen_drift_test.go`).

## CapabilityClaim (test fixture, not runtime data)

The FR-006 allowlist entry: maps one Skill's textual verification
promise to the real capability backing it.

| Field | Type | Notes |
|---|---|---|
| `SkillName` | string | |
| `Claim` | string | Short label matching prose in that Skill's "Validation Rules"/"Deterministic Operations" section (e.g. `"requirement coverage"`). |
| `BackingCode` | string | One of `internal/validation`'s `Code*` constants, when the claim is a structural-validation promise. |
| `BackingOp` | string | One of `knownInternalCommands` (already defined in `skills_content_test.go`), when the claim is a non-validation deterministic capability (e.g. fingerprinting). |

**Validation rule**: exactly one of `BackingCode`/`BackingOp` MUST be
set per claim — a claim backed by neither fails the test (FR-006).

## Metrics (extended — `internal/eval`)

Existing type (`internal/eval/record.go`); this feature adds one field.

| Field (new) | Type | Notes |
|---|---|---|
| `ContextFallbacks` | `int` | Count of times, during this `TaskResult`'s run, the Context Pack/prepared context was insufficient and the agent fell back to further reads (spec FR-008/SC-005). A plain count, not estimate-prone (unlike tokens/cost) — no `*_estimated` sibling required, consistent with `Calls`/`ExtraReads`/`Rework`, which also have none today. |

**Validation rule**: `ContextFallbacks` MUST be `>= 0` (added to
`Metrics.Validate()` alongside the existing estimated-sibling checks).
No change to `RunRecord`, `TaskResult`, `Baseline`, or `compare.go`'s
public shape beyond this one additive field. Correction found during
implementation: `compare.go`'s `compareTaskResults` diffs only
`TaskResult.Outcome` — it has never diffed individual `Metrics` fields
(`ExtraReads`/`Rework` are not surfaced by `Compare()` either, both
pre-dating this feature), so there was no existing per-field diffing
path for `ContextFallbacks` to "join." The field still satisfies
FR-008/SC-005 by being a durable, structured, JSON-serialized figure
any consumer of two `RunRecord`s can read and diff directly — adding a
dedicated comparison-report line for it would be new scope beyond what
this feature needs (Principle IV).

## SmokeTestScenario (test fixture, not runtime data)

One representative-integration run for FR-007/SC-004.

| Field | Type | Notes |
|---|---|---|
| `AdapterID` | string | One of the three fixed representative adapters: `claude-code`, `cursor-agent`, `copilot`. |
| `TargetPath` | string | The adapter's install target (`.claude/skills`, `.cursor/skills`, `.github/skills` respectively) — asserted distinct across the fixed set. |
| `CommandChain` | ordered list of string | The real `internal` command sequence a representative Skill (`mister-implement`) documents: `resolve`, `context` or `prepare`, `validate`. |
| `Outcome` | `pass` \| `fail` | Recorded per adapter; a `fail` names the adapter and the failing step in the chain (spec Acceptance Scenario US3.1). |

No persistence — this is the shape of one test-table row in
`internal/example/skill_smoke_test.go`, not a stored artifact.
