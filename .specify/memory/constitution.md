<!--
Sync Impact Report
===================
Version change: 1.0.0 → 1.1.0
Rationale: MINOR — expanding an existing Architecture Constraint's own
scope (Public command surface), not redefining or removing a principle.
Per the user's explicit request, adding two new top-level CLI flags
(`--version`, `--update`) that would otherwise conflict with the
frozen "misterspec init only" constraint — resolved here via a
deliberate amendment, not by silently overriding it.

Modified principles: N/A — no Core Principle (I–IX) changed. Only the
"Architecture Constraints" section's "Public command surface" and
"Embedded kit" bullets were amended.

Added sections: none (existing bullets expanded in place)

Removed sections: none

Templates requiring updates:
  - ✅ .specify/templates/plan-template.md — "Constitution Check" gate is
    generic and derives from this file at plan time; no edit needed.
  - ✅ .specify/templates/spec-template.md — no constitution-coupled content;
    no edit needed.
  - ✅ .specify/templates/tasks-template.md — no constitution-coupled
    content specific to this amendment; no edit needed.
  - ✅ No template or command file references "Public command surface" or
    "misterspec init" directly (grep-confirmed) — nothing else to sync.

Follow-up TODOs: none. The actual `--version`/`--update` feature work is
deferred to `/speckit-specify` (see Next Actions below) — this amendment
only clears the constitutional path for it.
-->

<!--
Sync Impact Report (1.0.0, superseded)
===================
Version change: TEMPLATE (unratified) → 1.0.0
Rationale: Initial ratification. The file previously held only unfilled
[PLACEHOLDER] tokens, so this is a MAJOR (first) version, not an amendment.

Modified principles: N/A (initial ratification)

Added sections:
  - Core Principles I–IX (semantic/deterministic separation, deterministic
    operations as the only mutation path, filesystem as source of truth,
    YAGNI/minimal configuration, test-first discipline, SOLID & clean code,
    explicit mutation boundaries, safety by construction, transparent
    machine-readable contracts)
  - Architecture Constraints (frozen core stack, from
    docs/architecture-specification.md §70)
  - Development Workflow & Quality Gates
  - Governance

Removed sections: none (all template placeholders resolved)

Templates requiring updates:
  - ✅ .specify/templates/plan-template.md — "Constitution Check" gate is
    generic and derives from this file at plan time; no edit needed.
  - ✅ .specify/templates/spec-template.md — no constitution-coupled content;
    no edit needed.
  - ✅ .specify/templates/tasks-template.md — updated the "Tests" note to
    reflect Principle V (Test-First Discipline, NON-NEGOTIABLE): test tasks
    are now the required default rather than opt-in.
  - ✅ .specify/templates/commands/*.md — directory does not exist in this
    project; nothing to sync.
  - ✅ README.md / docs/quickstart.md — no constitution references found;
    nothing to sync.

Follow-up TODOs: none. No placeholder was deferred.
-->

# MisterSpec Constitution

## Core Principles

### I. Semantic/Deterministic Separation (NON-NEGOTIABLE)

If an answer depends on what project information *means*, the coding agent
MUST decide it. If an answer can be *computed* from repository structure, the
`misterspec` Go binary MUST compute it, and the agent MUST call that
operation instead of reproducing its logic through LLM reasoning. This
boundary MUST be visible in every package (`internal/operations`,
`internal/validation`, `internal/ids`, …) and in every `SKILL.md`'s
"Deterministic Operations" section.

**Rationale**: Mixing semantic judgment and mechanical computation inside
either layer produces two failure modes MisterSpec exists to prevent: an LLM
inventing IDs/paths/hashes it cannot reliably compute, and a Go binary making
product/architecture judgments it cannot reliably reason about.

### II. Deterministic Operations Are the Only Mutation Primitive

IDs, canonical paths, source fingerprints, and structural artifact scaffolds
MUST be produced exclusively by `misterspec internal …` operations (e.g.
`create`, `create-artifact`, `resolve`, `fingerprint`). Skills MUST NOT
invent entity IDs, guess canonical paths, or fabricate hashes. ID allocation
and initial artifact creation MUST be atomic (a single `create`, never a
split `next-id` → `scaffold` sequence) to avoid races and half-created
entities.

**Rationale**: Splitting allocation from creation, or letting an LLM
approximate deterministic output, reintroduces exactly the class of bugs
(duplicate IDs, wrong paths, drifted hashes) this project exists to
eliminate.

### III. Filesystem Is the Single Source of Truth

Project state MUST be fully reconstructable from repository files at any
time. There MUST be no persistent authoritative counter (e.g. a stored
"next ID"), no database, and no vector store as authoritative state. Next IDs
MUST be derived by scanning existing artifacts (max existing suffix + 1). A
short-lived, non-authoritative allocation lock (e.g. `.misterspec/.lock`) MAY
exist only to protect a single mutating operation and MUST be safely
recoverable if stale.

**Rationale**: Authoritative secondary state (counters, databases) can drift
from the filesystem and silently corrupt project history; a filesystem+Git
model keeps the project inspectable, diffable, and recoverable with
ordinary tools.

### IV. Simplicity First — YAGNI & Minimal Configuration

New package layers (`domain/`, `application/`, `repositories/`, `services/`,
`controllers/`, or equivalent) MUST NOT be introduced speculatively; they
require demonstrated implementation pressure, not anticipated need. New
machine commands for decisions that require semantic reasoning (e.g.
"find relevant knowledge", "decide feature boundaries") MUST NOT be added —
those stay with the agent. Configuration fields (`.misterspec/config.yaml`)
MUST be added only for values that genuinely vary between projects; values
that never need customization stay as implementation defaults.

**Rationale**: This is YAGNI and KISS applied at the architecture level:
every extra layer, config knob, or command is a maintenance and cognitive
cost that must be justified by an actual, not hypothetical, need.

### V. Test-First Discipline (NON-NEGOTIABLE)

Every change to deterministic logic (ID parsing/allocation, path resolution,
frontmatter parsing, fingerprinting, validation rules, operations) MUST be
covered by unit tests, written before or alongside the implementation, and
MUST fail before the implementation satisfies them when following TDD.
Beyond unit tests, the four test layers defined in
`docs/architecture-specification.md` §65–66 are mandatory as the codebase
grows:

1. **Unit tests** — pure logic (ID parsing/allocation, path resolution,
   frontmatter parsing, fingerprinting, validation rules).
2. **Filesystem integration tests** — using temporary repositories, covering
   entity creation, ID resolution, duplicate detection, and invalid-parent
   rejection.
3. **Adapter tests** — generated agent-integration output checked against
   golden fixtures.
4. **TUI model tests** — Bubble Tea state transitions tested as state, not
   only via visual snapshots.

Golden-file tests (templates, adapters, `config.yaml`, `install.json`) MUST
require deliberate, reviewed changes — a golden diff is a signal, not noise
to suppress.

**Rationale**: MisterSpec's entire value proposition is that deterministic
operations are *reliable*. Untested deterministic code is a silent
contradiction of the project's core premise.

### VI. Clean Code & SOLID in the Go Implementation

Go code in this repository MUST follow standard clean-code discipline:

- **SRP** — a package/type has one reason to change (e.g. `ids` allocates,
  `validation` validates, `operations` orchestrates; they do not blur).
- **OCP** — extend behavior by adding new implementations (a new `Adapter`,
  a new validation rule) rather than modifying stable, depended-upon
  contracts.
- **LSP/ISP/DIP** — interfaces (e.g. `Adapter`) MUST stay small and defined
  from the consumer's need; higher-level packages MUST depend on interfaces,
  not concrete agent- or format-specific implementations.
- **DRY** — duplicated logic (ID parsing, path canonicalization, frontmatter
  handling) MUST be factored into the owning package instead of copied
  across `operations`, `cli`, or `tui`.
- **KISS** — the simplest implementation that satisfies the current
  requirement wins over a more general or "future-proof" one (see
  Principle IV).

**Rationale**: These are not abstract ideals; they are how §32's package
boundaries (`ids`, `validation`, `operations`, `artifacts`, `agents`, …) stay
independently testable and safe to evolve without cross-package breakage.

### VII. Explicit Mutation Boundaries

Each Skill and each deterministic operation MUST mutate only the artifacts
it owns and MUST NOT casually rewrite upstream artifacts. Concretely (per
`docs/architecture-specification.md` §55): `/create-plan` may write
`plan.md` but must not rewrite `spec.md`; `/create-tasks` may write
`tasks.md` but must not rewrite `plan.md` without explicit user request;
`/implement` may modify code, tests, and task evidence but must not rewrite
Spec requirements; `/analyze` may write `validation.md` and candidate
Learnings but must never modify implementation to force validation to pass,
nor rewrite a Spec to match the code. Agent adapters MUST NOT change
artifact semantics, modify project requirements, or hold their own
authoritative state — canonical behavior always originates from
`.misterspec/skills/`.

**Rationale**: Silent cross-layer rewrites are how "the implementation now
matches the spec" and "the spec now matches the implementation" become
indistinguishable — destroying the audit trail the artifact lifecycle exists
to provide.

### VIII. Safety by Construction

Every mutating deterministic operation MUST guarantee resolved targets stay
inside the repository root, rejecting path traversal (e.g. `../../etc/passwd`)
and defending against symlink escape where applicable. Framework-managed
artifact writes MUST be atomic (temp file → complete write → fsync where
appropriate → atomic rename). Initialization (`misterspec init`) MUST build
a plan, preview it, and get user confirmation before mutating the
filesystem, and MUST leave partial failure detectable and safely rerunnable
rather than silently corrupt.

**Rationale**: A tool that scaffolds and rewrites files inside a user's
repository must be trustworthy by default — safety here is not optional
hardening, it is the product's baseline contract.

### IX. Transparent, Machine-Readable Contracts

Internal commands (`misterspec internal …`) MUST default to structured JSON
with no decorative/ANSI output, and MUST distinguish `ok` (command executed)
from `valid`/`result` (semantic or structural outcome) as separate concepts.
Failures MUST return a non-zero exit code, a stable JSON error code, and a
concise message; Skills reason about error codes, not prose. Every Skill
invocation MUST end with a completion summary (outcome, artifacts, findings,
attention items) and an explicit recommended next slash command.

**Rationale**: Agents and users both need dependable, parseable signals
about what happened and what to do next; ambiguous prose output undermines
the whole deterministic/semantic split this project is built on.

## Architecture Constraints

The following decisions are frozen for the MVP (see
`docs/architecture-specification.md` §70) and MUST NOT be casually
reopened without a constitution amendment:

- **Language**: Go. **CLI framework**: Cobra. **TUI framework**: Bubble Tea.
- **Public command surface**: `misterspec init`, `misterspec --version`,
  and `misterspec --update` (amended 1.1.0 — see Sync Impact Report);
  the internal deterministic API lives under a hidden
  `misterspec internal …` namespace and MUST NOT be advertised in normal
  `--help` output. `--version` and `--update` MUST remain simple,
  self-contained utility flags — they MUST NOT grow into a broader
  user-facing CLI vocabulary (Principle IV, "Primary product UX" below).
- **Primary product UX**: agent slash commands, not a large user-facing CLI
  vocabulary.
- **Artifacts**: stored as Markdown + YAML frontmatter under `ai/`; project
  memory lives at `ai/memory/constitution.md`.
- **Canonical Skills**: sourced from `.misterspec/skills/`.
- **History**: Git. **Secondary authoritative state**: none (no database, no
  vector store).
- **Embedded kit**: templates, skills, and integrations are embedded into the
  binary via `go:embed`; `misterspec init` MUST NOT require network access
  to retrieve standard Skills or templates. `misterspec --update` (amended
  1.1.0) is the sole deliberate exception to this no-network rule — its
  entire purpose is checking GitHub for a newer release and fetching it;
  every other command, including `init`, remains network-free.
- **Canonical lifecycle**: Raw → Knowledge → Constitution → Program →
  Feature → Spec → Plan → Tasks → Implementation → Validation. Skills MUST
  inspect current state rather than assume strict linear progression.

## Development Workflow & Quality Gates

- Pull requests and reviews MUST verify compliance with this constitution;
  any added complexity (new package layer, new config field, new internal
  command) MUST be explicitly justified against Principle IV.
- Generated/golden artifacts (templates, adapter output, `config.yaml`,
  `install.json`) require reviewed, deliberate diffs — never auto-accepted.
- Dogfooding (§67–68): once `misterspec init`, the canonical Skills, and a
  first agent adapter exist, further MisterSpec feature work SHOULD proceed
  through MisterSpec itself where practical.
- Framework operating rules (how Skills behave) belong to canonical Skill
  policy; project-specific invariants belong in this Constitution — the two
  MUST NOT be duplicated into a separate agent-policy document.

## Governance

This Constitution supersedes ad-hoc conventions and prior undocumented
practice for this repository. Amendments are made by editing this file via
pull request, MUST include an updated Sync Impact Report (as an HTML comment
at the top of this file), and MUST follow semantic versioning:

- **MAJOR** — backward-incompatible principle removal or redefinition.
- **MINOR** — a new principle added, or existing guidance materially
  expanded.
- **PATCH** — clarification, wording, or non-semantic refinement.

Dependent templates (`.specify/templates/plan-template.md`,
`spec-template.md`, `tasks-template.md`, and any command files) MUST be
reviewed for consistency whenever a principle is added, removed, or
materially changed, and updated in the same amendment. Use `CLAUDE.md` for
day-to-day runtime agent guidance; this Constitution remains the durable,
non-negotiable layer beneath it.

**Version**: 1.1.0 | **Ratified**: 2026-09-11 | **Last Amended**: 2026-09-18
