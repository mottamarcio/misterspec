# Phase 1 Data Model: Core Repository Foundation

Entities extracted from `spec.md` § Key Entities, expanded with fields,
relationships, and validation rules drawn from the Functional Requirements
and `docs/architecture-specification.md` §20–31, §60–61.

## Project

Represents the root of a misterspec-managed repository once detected.

| Field | Type | Notes |
|---|---|---|
| `Root` | absolute path | Normalized, no trailing separator. Identical regardless of the subdirectory detection started from (FR-002). |
| `Config` | `Configuration` | The resolved configuration for this project. |

**Validation rules**:
- A `Project` only exists in "detected" form; there is no partially-valid
  `Project` value — detection returns either a `Project`, a distinct
  not-initialized result, or a distinct invalid-configuration result
  (FR-001, FR-003, FR-005).

**Relationships**: A `Project` owns exactly one `Configuration` and is the
root against which every `CanonicalPath` is resolved.

## Configuration

The resolved settings governing where things live in one project, loaded
from `.misterspec/config.yaml` (schema per architecture spec §21).

| Field | Type | Default | Notes |
|---|---|---|---|
| `SchemaVersion` | int | — | Required; unrecognized values are an invalid-configuration condition. |
| `AgentID` | string | — | e.g. `claude-code`. |
| `ArtifactsDir` | relative path | `ai` | Root of all semantic project state. |
| `RawDir` | relative path | `ai/raw` | |
| `KnowledgeDir` | relative path | `ai/knowledge` | |
| `ConstitutionPath` | relative path | `ai/memory/constitution.md` | |
| `LearningsDir` | relative path | `ai/memory/learnings` | |
| `ProgramsRoot` | relative path | `ai/programs` | |
| `IDWidth` | int | `3` | Zero-padding width for numeric ID suffixes. |

**Validation rules** (FR-004, FR-005):
- Missing `SchemaVersion` or an unsupported value → invalid-configuration,
  naming the field.
- Any required directory field present but empty/malformed → invalid-
  configuration, naming the field. A required field is never silently
  defaulted; only genuinely optional fields fall back to their default.
- `IDWidth` must be a positive integer.

## Entity ID

A typed, prefixed, zero-padded identifier uniquely naming one artifact
within its entity type (e.g. `SPEC-014`).

| Field | Type | Notes |
|---|---|---|
| `Type` | `EntityType` enum | One of `program`, `feature`, `spec`, `task`, `knowledge`, `learning`. |
| `Prefix` | string | Derived from `Type` (`PRG`, `FEAT`, `SPEC`, `TASK`, `KNOW`, `LRN`). |
| `Number` | int | The numeric suffix; always ≥ 1. |
| `Width` | int | The zero-padding width in effect when parsed (from `Configuration.IDWidth`). |

**Validation rules** (FR-010):
- Prefix must exactly match the type's fixed prefix.
- Suffix must be all-numeric, matching `Width` exactly (no shorter, no
  longer) — a wrong-width suffix is invalid, not silently accepted.
- An `EntityID` is only ever constructed via `Parse`; there is no way to
  build an unvalidated one from raw user/agent input.

**Relationships**: An `EntityID` identifies exactly one `Artifact`. A
`Project`, scanned for a given `EntityType`, yields a set of `EntityID`
values (FR-011) plus any detected duplicates (FR-012) and a computed "next"
`EntityID` (FR-013, pure function: `max(existing.Number) + 1`, or `1` if
none exist).

## Artifact

Any canonical Markdown-with-frontmatter file representing a Program,
Feature, Spec, Plan, Tasks, Validation, Knowledge item, Learning, or the
Constitution.

| Field | Type | Notes |
|---|---|---|
| `Type` | `ArtifactType` enum | `program`, `feature`, `spec`, `plan`, `tasks`, `validation`, `knowledge`, `learning`, `constitution`. Superset of `EntityType` (plan/tasks/validation/constitution have no independent ID — FR-006). |
| `Path` | `CanonicalPath` | Where this artifact's file lives. |
| `Metadata` | `Metadata` | Parsed frontmatter, once successfully parsed. |

**Validation rules** (FR-006, FR-009):
- `Type` is determined from canonical location first, declared frontmatter
  `type` field second; a mismatch between the two is a discoverable
  condition (surfaced, not silently resolved one way).
- An `Artifact` value is only produced once frontmatter parsing succeeds;
  failures short-circuit into one of the distinct error conditions below —
  there is no "partially populated" `Artifact`.

## Metadata

The structured result of parsing one artifact's YAML frontmatter block.

| Field | Type | Notes |
|---|---|---|
| `ID` | `*EntityID` | `nil` for types without an independent ID (plan, tasks, validation, constitution). |
| `Type` | string | Raw declared `type` value, prior to cross-checking against location. |
| `Status` | string | Free-form per artifact type's allowed lifecycle states (validated by a later Validation feature, not this one — see spec Assumptions). |
| `Parent` | `*EntityID` | Present for feature/spec. |
| `DependsOn` | `[]EntityID` | Present for spec. |
| `Supersedes` | `[]EntityID` | Present for spec. |

Note (reconciled post-implementation, see `contracts/packages.md`): a
speculative `Extra map[string]any` field for artifact-type-specific
frontmatter (`for`, `result`, `sources`, …) was dropped from the shipped
`Metadata` struct — no requirement or test in this feature reads it
(Constitution Principle IV, YAGNI). Add it in whichever later feature
first needs those fields.

**Error conditions** (FR-009, distinct and reportable):
1. `ErrArtifactNotFound` — the file does not exist.
2. `ErrFrontmatterMalformed` — no frontmatter delimiters found, or the YAML
   between them fails to parse.
3. `ErrRequiredFieldMissing` — frontmatter parses, but a field required for
   that artifact's declared/located type is absent (e.g. `id` on a Spec).

## Canonical Path

The single, deterministic filesystem location an entity's artifact must
occupy, derived from its type, ID, and parent chain (FR-007).

| Field | Type | Notes |
|---|---|---|
| `Directory` | path, relative to `Project.Root` | The entity's canonical directory. |
| `File` | path, relative to `Project.Root` | The entity's canonical file within `Directory` (e.g. `spec.md`, `KNOW-004-session.md`). |

**Validation rules** (FR-008):
- Computing or resolving a `CanonicalPath` MUST reject any input (ID,
  parent, or raw path) whose result would lie outside `Project.Root` —
  this check happens before any filesystem access, not after.
- Computation is a pure function of `(Project.Root, Config, Type, ID,
  Parent)` — no caching, no stored state (SC-003).

## State / Flow Summary

```text
Project Detection (US1)
  cwd/subdir → walk upward → find .misterspec/config.yaml
    → not found:  NotInitialized
    → found, malformed:  InvalidConfiguration(field)
    → found, valid:  Project{Root, Config}

Path Resolution & Typing (US2)          [depends on Project]
  (Type, ID, Parent) → CanonicalPath     (rejects traversal)
  ArtifactFilePath → ArtifactType        (location and/or declared type)

Metadata & ID Discovery (US3)           [depends on US2's typing/paths]
  Artifact file → parse frontmatter → Metadata | typed error
  EntityType → scan tree → []EntityID + []DuplicateID
  []EntityID → NextID (max + 1, or 1)
```
