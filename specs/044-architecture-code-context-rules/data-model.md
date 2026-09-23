# Data Model: Regras de Arquitetura e Contexto de Código

Three new packages: `internal/gosource` (shared leaf — Go source
parsing, research.md #1/#2), `internal/architecture` (Facet A — rule
declaration and evaluation), and additive changes to
`internal/context/index` plus `internal/prepare` (Facet B —
code-context retrieval), plus two new `project.Configuration` fields
shared by both facets (research.md #3/#4).

## ArchitectureRule (config)

Declared in `.misterspec/config.yaml`'s new, optional
`architecture_rules` list (research.md #3).

| Field | Type | Notes |
|---|---|---|
| `Kind` | `"forbidden_dependency" \| "layer_boundary" \| "required_contract"` | Which class of constraint this rule expresses. |
| `From` | `string` | A module/path pattern (e.g. `internal/artifacts/**`) naming what the rule constrains. |
| `To` | `string` | For `forbidden_dependency`: the forbidden import path pattern. For `layer_boundary`: the boundary's own "outer" pattern `From` MUST NOT import from. Unused (empty) for `required_contract`. |
| `Contract` | `string` | For `required_contract` only: the required import/symbol pattern `From` MUST import or implement. Empty otherwise. |

A malformed rule (unknown `Kind`, empty `From`) is itself a
configuration problem, reported the same way `project.Load` already
reports other malformed configuration — not silently skipped.

## CodeExclusions (config)

| Field | Type | Notes |
|---|---|---|
| `CodeExclusions` | `[]string` | Glob patterns, relative to the project root, applied identically by both facets (research.md #4). Empty by default. |

## Language Adapter (interface concept)

| Concern | Notes |
|---|---|
| Applicability | Detected per-project, not declared (research.md #8) — Go via `go.mod` presence at the project root. |
| Contract | Given the resolved rule set and the project root, returns one `Result` per (rule, applicable-location) pair it can evaluate, and reports every rule it cannot evaluate as `not_evaluated` rather than omitting it. |

## Result

One rule's own evaluation outcome (spec Key Entity "Resultado de
Arquitetura").

| Field | Type | Notes |
|---|---|---|
| `RuleIndex` | `int` | The rule's own position in the resolved `architecture_rules` list — stable within one run, used to correlate a `Result` back to its declaring rule. |
| `Status` | `"pass" \| "fail" \| "not_evaluated"` | Never a fourth value. `not_evaluated` MUST NOT be conflated with `pass` anywhere in rendering (spec FR-004/FR-005). |
| `Path` | `string` | The file where a `fail` was found, or the file/pattern a `not_evaluated` applies to. Empty for a project-wide `pass` with nothing more specific to name. |
| `Line` | `int` | 1-indexed, file-absolute — the exact import statement's own line for a `fail` on `forbidden_dependency`/`layer_boundary`. 0 when not applicable (e.g. `required_contract`, or `not_evaluated`). |
| `Message` | `string` | A specific, human-readable explanation — never generic, mirroring `validation.Finding.Message`'s own discipline. |
| `Reason` | `string` | Present only for `Status == "not_evaluated"` — `"no_adapter_for_project"` or `"rule_kind_unsupported"`, so a caller can tell the two `not_evaluated` causes apart. |

## ArchitectureCheckReport

The full response of one `check-architecture` run.

| Field | Type | Notes |
|---|---|---|
| `Adapter` | `string` | `"go"`, or `""` when no adapter applied to this project at all (every rule then reports `not_evaluated` with `reason: "no_adapter_for_project"`). |
| `Results` | `[]Result` | One entry per declared rule at minimum; a `forbidden_dependency`/`layer_boundary` rule may produce more than one `fail` entry (one per violating import site) alongside its own base evaluation. |

## CodeFile (index)

One indexed Go source file — the code-side sibling of the existing
`documents` table (research.md #6), in `internal/context/index`.

| Column | Type | Notes |
|---|---|---|
| `path` | `TEXT UNIQUE NOT NULL` | Project-root-relative. |
| `fingerprint` | `TEXT NOT NULL` | Whole-file content fingerprint (same `sha256:<hex>` convention `contextengine.Fingerprint` already establishes). |
| `indexed_at` | `INTEGER NOT NULL` | Unix seconds, mirroring `documents.indexed_at`. |

## CodeDeclaration (index)

One top-level declaration inside an indexed `CodeFile` — the code-side
sibling of `chunks` (research.md #6).

| Column | Type | Notes |
|---|---|---|
| `code_file_id` | `INTEGER NOT NULL REFERENCES code_files(id)` | |
| `name` | `TEXT NOT NULL` | The declaration's own identifier (function/type/const/var name). |
| `kind` | `TEXT NOT NULL` | `"func" \| "type" \| "const" \| "var"`. |
| `signature` | `TEXT NOT NULL` | The declaration's own signature text plus its doc comment — always populated, always the cheapest tier (research.md #7). |
| `body` | `TEXT` | The full declaration text, populated only when indexed at full-body tier; `NULL` otherwise — the signature/body distinction is a presence check, not a separate flag. |
| `start_line` / `end_line` | `INTEGER NOT NULL` | File-absolute, covering the full declaration (signature *and* body span the same range; `body` being `NULL` means only the range is known, not that it differs). |
| `is_test` | `INTEGER NOT NULL` (bool) | `true` when `code_file_id`'s own path is a `_test.go` file — lets retrieval group a declaration with its own associated tests (spec FR-007's "testes já associados"). |

Associated-test grouping (FR-007) is computed at query time: a
declaration in `foo.go` is associated with every `CodeDeclaration` row
whose own `CodeFile.path` is `foo_test.go` in the same directory — the
same file-naming convention Go's own toolchain already uses, not a new
one this feature invents.

## TaskPreparation (extended)

`internal/prepare`'s existing `TaskPreparation` (034/data-model.md)
gains one new field (research.md #6):

| Field | Type | Notes |
|---|---|---|
| `CodeContext` | `[]contextengine.PackageItem` | One entry per declaration resolved from the Task's own `Scope:` field — `Heading` carries the declaration's own name, `Content` carries `signature` (or `body` when promoted to full-body tier per Decision 7's fixed threshold), `Location`/`Fingerprint` exactly as 033 already defines them. Empty (not omitted) when `Scope:` is empty or names no `.go` file. |
| `CodeScopeNotFound` | `[]string` | Every `Scope:`-named path that does not resolve to an indexed file — surfaced explicitly (spec Edge Case: a Scope path that no longer exists), never silently dropped from the response. |

No new response type: code content reuses `contextengine.PackageItem`
verbatim, the same shape `--mode package` already returns for Markdown
content — a caller reading `internal prepare`'s own JSON learns one
item shape, not two.

## Reused, unmodified types (no new definitions)

- `contextengine.PackageItem` / `Fingerprint` (033) — the code
  response shape, unmodified.
- `artifacts.Estimator` (035) — reused for Decision 7's fixed
  signature-vs-body size threshold; no new estimation logic.
- `index.Store` (014, extended by 043) — `code_files`/
  `code_declarations` are new tables in the same disposable database,
  not a new store.
- `prepare.TaskFields.Scope` (034) — the existing `Scope:` parser is
  the input to Facet B's own file resolution; this feature adds no new
  Task-body syntax.

## New, additive changes to existing contracts

- `project.Configuration`: `ArchitectureRules []ArchitectureRule` and
  `CodeExclusions []string`, both optional, both empty by default
  (research.md #3/#4) — no existing field changes meaning.
- `internal/context/index`: `schemaVersion` bump; `code_files`/
  `code_declarations` tables added (research.md #6).
- `misterspec internal prepare`: response gains `code_context`/
  `code_scope_not_found` (data-model.md "TaskPreparation (extended)")
  — additive fields; a caller ignoring them sees no change.

## New, standalone contract

- `misterspec internal check-architecture` (new command,
  `internal/cli/internalcmd`) — full request/response documented in
  `contracts/check-architecture-contract.md`.
