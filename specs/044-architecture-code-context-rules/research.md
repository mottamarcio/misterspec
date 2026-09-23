# Research: Regras de Arquitetura e Contexto de Código

Input: `specs/044-architecture-code-context-rules/spec.md`. This
feature covers two related facets (architecture-rule verification and
task-scoped code retrieval) in one Spec, per the backlog document's
own framing. Each decision below states whether it serves one facet or
both, and what existing capability it reuses.

## Decision 1 — One Spec, two new packages, sharing one Go-source-walking foundation; no split

**Decision**: This plan does **not** split into two Specs. Both facets
share the same first real need this project has for *reading Go source
files as structured data* (not just Markdown) — architecture rules
need each file's own import list; code-context retrieval needs each
file's own top-level declarations (signatures) and associated test
files. A single new leaf capability, a new package `internal/gosource`
(pure parsing: imports and top-level declaration signatures, no I/O
beyond reading the one file it's given, no dependency on any other
`internal/*` package — the same "leaf" role `internal/evidence` already
plays per 041's own precedent), is built once and consumed by both:
`internal/architecture` (rule declaration + evaluation, Facet A, new
package) and `internal/context/index` (code retrieval's own indexing,
Facet B, extending the existing disposable index).

**Rationale**: Splitting into two Specs now would either duplicate
this Go-source-walking foundation across two Specs, or force an
awkward dependency of one Spec's implementation on the other's
mid-flight — worse than one Spec with two independently-testable user
story clusters (already how spec.md is organized: US1/US2 for rules,
US3 for retrieval), matching the precedent 041/042/043 already
established for a multi-facet Spec (Constitution Principle VI, DRY).

**Alternatives considered**: Splitting per the backlog document's own
suggestion — rejected for now, specifically because the shared
foundation makes combined scope smaller, not larger, than two separate
Specs would be; revisit only if implementation reveals the two facets
don't actually share as much as expected (Assumptions in spec.md keep
this door open).

## Decision 2 — Go source parsing uses only the standard library (`go/parser`, `go/ast`, `go/token`); no new dependency

**Decision**: `internal/gosource` (Decision 1) parses `.go` files with
`go/parser`'s `ParseFile(fset, path, nil, parser.ImportsOnly)`
(imports — cheap, consumed by the Go Language Adapter, Facet A) and
`parser.ParseComments` (full AST — for top-level declaration
signatures, consumed by code indexing, Facet B), both already in the
Go standard library shipped with the toolchain this project already
requires (`go 1.23.4`, `go.mod`).

**Rationale**: Zero new third-party dependency (Constitution Principle
IV) for exactly the two things needed: import lists and top-level
declaration signatures. `golang.org/x/mod` (already a dependency, used
elsewhere for module-path handling) is reused for resolving a Go
import path against this project's own module path when checking
whether an import is "internal" to the project.

**Alternatives considered**: A third-party static-analysis library
(e.g. `golang.org/x/tools/go/packages`) — rejected; it type-checks and
loads the full build graph, far more than "list imports and top-level
signatures" needs, and would be MisterSpec's own first dependency on
a tool that itself requires a working Go toolchain/module cache at
analysis time — a much heavier requirement than parsing source text.

## Decision 3 — Architecture rules are declared in a new project config section, not a new artifact type

**Decision**: Declared rules (forbidden dependency, layer boundary,
required contract) live in a new, optional `architecture_rules`
section of `.misterspec/config.yaml` (`project.Configuration`) — a
list of rule declarations, each naming its own kind, the module/path
pattern(s) involved, and (for a forbidden-dependency rule) the
forbidden target pattern. Absent entirely, the check reports zero
rules configured (not an error, not `not_evaluated` — there is simply
nothing declared to evaluate).

**Rationale**: A rule is project-wide configuration, not project
content with its own lifecycle/identity — it does not need an ID,
frontmatter, a canonical directory, or any of the five entity types'
own machinery. `project.Configuration` already exists exactly for
"values that genuinely vary between projects" (Constitution Principle
IV's own criterion) — architecture rules vary per project by
definition, unlike e.g. `id_width`.

**Alternatives considered**: A new Knowledge-like artifact type for
rules — rejected; rules are not content an agent reads for meaning,
they are declarative input to a mechanical check, exactly what
Constitution Principle I reserves for the binary, not a new
artifact-lifecycle concept. A dedicated `ai/architecture-rules.yaml`
file outside `.misterspec/config.yaml` — rejected; introduces a second
project-configuration file/format for no benefit over one additional,
optional section of the existing one.

## Decision 4 — Exclusions are a new, small, additive `Configuration` field — the first of its kind in this project

**Decision**: A new optional field, `code_exclusions []string`
(glob patterns, relative to the project root — e.g. `vendor/**`,
`**/*_generated.go`), is added to `project.Configuration`. Empty by
default (no exclusions). Applied identically by both facets: an
excluded path is never walked for rule evaluation and never indexed
for code-context retrieval, regardless of whether it would otherwise
match a declared rule or a Task's own Scope.

**Rationale**: Spec FR-010 explicitly requires respecting "exclusões
já declaradas pelo projeto" — research confirmed no such mechanism
exists yet anywhere in this codebase (`project.Configuration` has no
exclusion-related field at all today). This is a genuinely new,
narrowly-scoped capability this Spec must introduce, not something to
merely "reuse" — documented here rather than silently assumed, since
the spec's own Assumptions section did not anticipate this gap.

**Alternatives considered**: Silently reusing `.gitignore` — rejected;
`.gitignore` governs what Git tracks, a different concern from what
this feature should analyze (a project may intentionally track
generated code in Git while still wanting it excluded from
architecture checks, or vice versa) — conflating the two would surprise
a maintainer relying on `.gitignore` for its own, unrelated purpose
(Constitution Principle IV: a config field's meaning should not be
inferred from an unrelated file with its own separate contract).

## Decision 5 — Architecture results are a new, dedicated tri-state type, not `validation.Finding`

**Decision**: A new type, `architecture.Result` (`Rule`, `Status`
[`"pass" | "fail" | "not_evaluated"`], `Path`, `Line`, `Message`),
is returned by a new deterministic operation and a new CLI command
(`internal check-architecture`) — not folded into
`internal/validation`'s existing `Finding`/`ValidateProject`/
`ValidateEntity` machinery.

**Rationale**: `validation.Finding` is binary by construction (a
problem exists, or it is never reported at all) — there is no existing
"I checked and it was fine" affirmative signal, and definitely no
`not_evaluated` distinct from silence, which spec FR-004/FR-005 make
non-negotiable. Retrofitting a third state onto `Finding` would change
what every existing `Finding` consumer (`internal validate`'s own
JSON contract) is allowed to assume about an absent Finding —
exactly the kind of casual cross-feature contract change Constitution
Principle VII warns against. Architecture rules also check *source
code*, not SDD artifacts — a different domain from everything
`internal/validation` currently owns (Constitution Principle VI, SRP).

**Alternatives considered**: Emitting a synthetic "pass" `Finding` with
a new `Severity` value — rejected; every existing `internal validate`
consumer today reasonably assumes a `Finding` names a problem, and
introducing a severity that means "no problem" would be a confusing,
backwards-incompatible reinterpretation of an already-shipped,
documented field (033's own precedent: additive, never
reinterpreting).

## Decision 6 — Code-context retrieval is a new field on `internal prepare`'s own response, not a new `internal context` target; storage extends the existing disposable index with a `code_files`/`code_declarations` pair reusing 033's `PackageItem` shape

**Decision**: `contextengine.Request.Target` today "must resolve to one
of the five entity types `operations.References` already covers"
(`internal/context/request.go`) — a Task is not, and was never meant
to be, a valid `internal context` target (034 deliberately built a
*separate* command, `internal prepare`, for Task-shaped questions,
rather than teaching `internal context` a sixth target type). Code-
context retrieval therefore does not hook into `internal context` at
all: `TaskPreparation` (034/data-model.md) gains one new field,
`CodeContext []contextengine.PackageItem`, populated by parsing the
Task's own already-existing `Scope:` field into file paths and
resolving each against a new `code_files`/`code_declarations` table
pair (mirroring `documents`/`chunks`' own shape,
`internal/context/index/schema.go`) that stores, per indexed Go file,
its own content fingerprint, and one row per top-level declaration
(function/type/const/var) carrying its own signature text, doc
comment, file-absolute line range, and — only when a body was actually
needed (Decision 7) — the full declaration text.

**Rationale**: `internal/context/index` already owns exactly this
class of problem — a disposable, reconstructable, fingerprinted,
line-addressable content store — for Markdown; extending it with a
second content kind (code) reuses the whole existing lifecycle
(`ensureSchema`'s version-mismatch-triggers-rebuild path, `Store.Sync`)
rather than inventing a second index (Constitution Principle III/VI).
Reusing `PackageItem` as the response shape means `internal prepare`'s
own JSON gains code content in a shape any existing `--mode package`
consumer (any Skill already built against 033) already knows how to
read — `content`/`location`/`fingerprint`, no new shape to learn.
Attaching it to `internal prepare` rather than `internal context`
respects 034's own deliberate separation (Task-shaped questions go to
`prepare`, artifact-shaped questions go to `context`) instead of
overloading `context.Request.Target` with a sixth, special-cased
target type it was never designed for (implementation-time correction
to this plan's own original draft, which had proposed exactly that).

**Alternatives considered**: Teaching `contextengine.Request.Target`
to also accept a Task ID — rejected; `internal prepare` already exists
specifically so a Task's own context assembly (Requirements, Plan
sections, and now code) happens in one dedicated response, and
`internal context`'s own contract (`Target` resolves via
`operations.References`) would need a special-cased carve-out for
exactly one non-`References`-covered type, contradicting its own
documented precondition. A wholly separate code-index database —
rejected; two independently-opened SQLite lifecycles in the same CLI
process is unjustified complexity this feature does not need
(Principle IV, same reasoning 043's own Decision 3 already applied to
the `packs` table). Inventing a new "code item" response shape
distinct from `PackageItem` — rejected; no expressive benefit
`PackageItem`'s existing fields don't already cover (`Content`,
`Location`, `Fingerprint`, `Heading` doubling as "declaration name").

## Decision 7 — Signature vs. full body is a fixed, documented size threshold — `internal prepare` has no budget mechanism to reuse

**Decision**: `internal prepare` (034) has no budget/estimator concept
at all today — research confirmed it returns Requirements text and
Plan sections unconditionally, in full, in one response (034's own
User Story 1: "reunir tudo isso... em uma única resposta
determinística"), unlike `internal context`'s own ranked-and-budgeted
pipeline. Code-context retrieval follows that same unconditional
philosophy for the size decision: every declaration named in the
Task's own Scope is always included at signature tier; its full body
is additionally included when the file's own total estimated size
(via the already-existing `artifacts.Estimator`, the same unit 035
already uses) is at or under a small, fixed, documented constant
(`internal/prepare`'s own `maxInlineCodeBodySize`, expressed in
estimated tokens — the package that owns this decision, per Decision
6's corrected integration point) — never a per-request negotiable
budget.

**Rationale**: Spec FR-008 requires the full body to never be "the
unconditional default" for every file regardless of size — a fixed
threshold satisfies that directly, without introducing a request-level
budget parameter into a command whose entire existing design
deliberately has none (Constitution Principle IV — no new
configuration surface where a simple, documented constant already
satisfies the requirement; implementation-time correction to this
plan's own original draft, which had proposed reusing 035's
`ApplyBudget` before research confirmed `internal prepare` never
calls it).

**Alternatives considered**: Reusing 035's `ApplyBudget`/backfill
machinery directly — rejected once research showed `internal prepare`
does not use it for anything else today; adopting it only for code
would make one response mix a budgeted sub-section inside an otherwise
unconditional command, a confusing, inconsistent contract for a caller
to reason about. A per-request `--code-budget` flag on `internal
prepare` — rejected as unjustified new surface (Principle IV) until a
real need for tuning it is demonstrated; the fixed constant is the
simplest thing that satisfies FR-008 today.

## Decision 8 — `internal check-architecture` is a new, read-only command; Go is the only adapter shipped now

**Decision**: One new deterministic command, `misterspec internal
check-architecture`, runs every declared rule against the current
project, using the Go Language Adapter when the project's own
predominant language is Go (detected via the presence of `go.mod` at
the project root — no separate, redundant "language" config field).
A project with no `go.mod` (or, in principle, a future non-Go project)
gets `not_evaluated` for every rule, per spec FR-004 — there is
exactly one adapter, and its own applicability check is explicit and
visible in the response, never inferred silently.

**Rationale**: Matches every other `internal` command's own contract
discipline (Constitution Principle IX) and mirrors `042`'s own
precedent of adding one new read-only command for one new
capability rather than overloading `internal validate`. `go.mod`
presence is already the standard, zero-configuration signal a Go
project provides for "this is a Go module" — no new detection
mechanism needed.

**Alternatives considered**: A `--language` flag the caller must pass
— rejected; the project's own `go.mod` already answers this
unambiguously for the one adapter this Spec ships, and a flag that
could contradict the project's own actual language would only invite
a caller-asserted mismatch this feature has no reason to trust over
the filesystem itself (mirrors 042/043's own "never trust a caller
assertion without verification" precedent).
