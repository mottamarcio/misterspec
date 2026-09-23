# Implementation Plan: Skills Enxutas e Contratos de Integração Testáveis

**Branch**: `039-lean-skills-integration-contracts` | **Date**: 2026-09-21 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/039-lean-skills-integration-contracts/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Today's 10 canonical Skills (`kit/skills/*/SKILL.md`, 2378 lines) share
the same 26-section contract but are hand-authored independently;
research.md's investigation found the real duplication is not whole
sections (those are genuinely Skill-specific) but a handful of literal,
multi-line canonical phrases repeated 3-9 times each, plus several
Skills still describing pre-034 multi-call context flows now superseded
by `internal prepare`/full-pack consumption. The technical approach:
(1) extract the actually-duplicated phrases into a small set of named
canonical fragments composed into each Skill's committed `SKILL.md` by a
new build-time generator (`internal/skillgen`), verified by a
drift-check test — never a runtime include (spec FR-001/FR-002); (2)
update Skills whose flow still bypasses `internal prepare`/full-pack
consumption to use it, trimming the superseded restatement (FR-003,
contributing to SC-001's size reduction); (3) extend the existing
`internal/example/skills_content_test.go` machine checks — which
already verify every `internal <op>` mention against a real command
allowlist — with two new checks: real Cobra-example-syntax validation
and a promised-verification-to-`validation.Code*`-capability allowlist
(FR-005/FR-006); (4) add a new smoke test exercising the real
`resolve → context/prepare → validate` command chain a Skill documents,
run against three representative adapters covering three distinct
install target paths (`claude-code`, `cursor-agent`, `copilot` — FR-007,
SC-004); (5) add one new field, `ContextFallbacks`, to
`internal/eval.Metrics` (037-eval-quality-efficiency), following its
existing `ExtraReads`/estimated-sibling convention, and update Skill
text to instruct recording a fallback occurrence rather than silently
absorbing it (FR-008).

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: `github.com/spf13/cobra` v1.10.2 (CLI — reused in-test to validate Skill example syntax against the real command tree, no new dependency); existing `internal/agents`, `internal/validation`, `internal/eval` packages. No new third-party dependency.
**Storage**: None new. `kit/skills/*/SKILL.md` remain plain Markdown, embedded via `kit.SkillsFS` (unchanged embed mechanism) — the new generator produces committed files, not a runtime-read source (Constitution Principle III: filesystem/Markdown stays the source of truth developers read and Git diffs show).
**Testing**: `go test` — unit tests for the new `internal/skillgen` fragment-composition logic; a golden/drift-check test regenerating all 10 Skills into a temp dir and byte-diffing against committed `kit/skills/*/SKILL.md`; extensions to `internal/example/skills_content_test.go` (example-syntax validation, verification-capability allowlist); a new filesystem-integration smoke test in `internal/example` running the real `resolve → context/prepare → validate` chain against three representative adapters; unit tests for `eval.Metrics.ContextFallbacks` following the existing `Metrics.Validate()` pattern.
**Target Platform**: cross-platform CLI binary (Linux/macOS/Windows), no network access required; smoke tests use only local temp fixtures, no live agent/LLM invocation.
**Project Type**: single Go project (CLI + internal libraries + embedded kit content)
**Performance Goals**: no new runtime performance target — `internal/skillgen` runs at development/CI time (via `go generate` and its drift-check test), never in the product's runtime path.
**Constraints**: generated Skill content MUST remain byte-identical to what a developer could hand-author and review as plain Markdown (Constitution Principle III/IX — no opaque build artifact); no per-agent content variance is introduced (research.md #1 — all adapters already receive identical content, only `TargetPath()` differs); no new persistent/authoritative state (Principle III) — `ContextFallbacks` is a hand-recorded `RunRecord` field like its `ExtraReads`/`estimated` siblings, not new automatic telemetry; SC-001's 20% size reduction is measured by line count over `kit/skills/*/SKILL.md` bodies, checked in-test, not by a new token estimator (Principle IV).
**Scale/Scope**: new `internal/skillgen` package (fragment types, per-Skill manifest, composition, drift-check); edits to all 10 `kit/skills/*/SKILL.md` files (converted to generator output, content trimmed per FR-002/FR-003); extensions to `internal/example/skills_content_test.go`; one new `internal/example/*_smoke_test.go` file; `internal/eval/record.go` (`Metrics.ContextFallbacks` field + `Validate()` update) and its test. No new package layer beyond `internal/skillgen`, no new CLI command (generation is a `go generate` + test-checked step, not a product-facing `misterspec` subcommand — Constitution "Public command surface" stays frozen).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I (Semantic/Deterministic Separation)**: PASS. Fragment composition, command-existence checking, and example-syntax validation are all fully mechanical (string composition, Cobra command-tree lookup). Deciding *whether* a Context Pack was insufficient (a fallback occurred) stays a judgment the agent makes and records, per research.md #5 — the binary never infers it.
- **Principle II (Deterministic Operations Are the Only Mutation Primitive)**: N/A. No new entity, ID, or canonical path is created; `internal/skillgen` produces development-time Markdown content, not project artifacts, and creates no IDs.
- **Principle III (Filesystem Is the Single Source of Truth)**: PASS. Generated `SKILL.md` files are committed and remain the actual source Git diffs show and `kit.SkillsFS` embeds unchanged; the drift-check test guarantees no divergence between fragments/manifests and committed output, so there is no hidden authoritative state.
- **Principle IV (Simplicity First — YAGNI)**: PASS with an explicit boundary, documented in research.md #1/#4: no per-agent variant mechanism (none is needed — investigated and confirmed unnecessary), no new `misterspec` subcommand for generation, no new token estimator for SC-001, no automatic fallback-detection telemetry (FR-008 stays a hand-recorded field, matching 037's existing precedent).
- **Principle V (Test-First Discipline)**: Applies — fragment composition, the drift-check, the two new `skills_content_test.go` checks, the smoke test, and `Metrics.ContextFallbacks` are exactly the deterministic-logic and contract-conformance surfaces this principle requires tests for, written alongside implementation.
- **Principle VI (Clean Code & SOLID)**: PASS. `internal/skillgen` owns fragment/manifest composition only; it does not duplicate `internal/agents` install logic or `internal/validation` rule logic — it reads `validation.Code*` constants and the CLI command tree as read-only references for the new test checks.
- **Principle VII (Explicit Mutation Boundaries)**: PASS. `internal/skillgen` writes only `kit/skills/*/SKILL.md`; it does not touch Spec/Plan/Tasks artifacts or any project-owned file. Skill text updates (FR-003) touch only the Skills' own files.
- **Principle VIII (Safety by Construction)**: PASS. No new mutating operation exposed to end users; generation is a development-time step over files already inside the repository, guarded by the same test suite as any other Go change.
- **Principle IX (Transparent, Machine-Readable Contracts)**: Applies — `eval.Metrics.ContextFallbacks` follows the exact JSON-field/estimated-sibling convention `Metrics` already uses; the smoke test asserts the real JSON envelope shape a Skill's own text promises, same discipline as existing `context`/`prepare` contract tests.

No violations requiring Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/039-lean-skills-integration-contracts/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/            # Phase 1 output (/speckit-plan command)
└── tasks.md              # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/
├── skillgen/                       # New package: canonical fragments + per-Skill manifests → SKILL.md
│   ├── fragment.go                  # Named fragment type + the actual fragment text as Go string constants (resolve-preamble, mechanical-steps-note, fallback-tolerance-note, …) — plain Go values, not a separate embedded filesystem, so kit.SkillsFS's existing "//go:embed skills" + mustSub keeps embedding only final composed Skill output, never raw fragment sources
│   ├── fragment_test.go
│   ├── manifest.go                  # Per-Skill manifest: ordered list of {fragment reference | bespoke inline body} per section
│   ├── manifest_test.go
│   ├── generate.go                  # Composes one Skill's manifest + fragments into final SKILL.md bytes
│   ├── generate_test.go
│   └── doc.go
├── eval/
│   ├── record.go                    # Metrics gains ContextFallbacks int; Validate() note updated (no new estimated-sibling needed — it's a count, not an estimate-prone figure)
│   └── record_test.go
├── example/
│   ├── skills_content_test.go       # Extended: example-syntax-against-real-Cobra-tree check (FR-005); verification-claim → validation.Code*/internal-op allowlist check (FR-006); Skill body line-count assertion (SC-001)
│   ├── skillgen_drift_test.go       # New: regenerates all Skills into a temp dir via internal/skillgen, byte-diffs against committed kit/skills/*/SKILL.md
│   └── skill_smoke_test.go          # New: installs generated Skills for claude-code/cursor-agent/copilot into temp projects; runs the real resolve → context/prepare → validate chain from mister-implement's documented flow against a fixture repo; asserts completion and JSON shape per adapter
└── cli/                             # Unchanged — no new command; Cobra tree read read-only by the new example-syntax check

kit/
└── skills/
    └── mister-*/SKILL.md            # Existing 10 files: content trimmed (superseded pre-034 flow restatement removed, FR-003) and now generator output (regenerate via `go generate ./internal/skillgen/...` after any manifest/fragment edit). No new subdirectory here — kit.SkillsFS's "//go:embed skills" must keep embedding only real, installable Skills.
```

**Structure Decision**: Single Go project, no new top-level directory.
`internal/skillgen` follows the existing package-per-capability
convention (`internal/eval`, `internal/validation`); it is a
development-time authoring tool, not a product command, so it lives
under `internal/` without CLI wiring (`internal/cli` is untouched) —
consistent with the frozen "Public command surface" constraint.
Fragment sources live as Go values inside `internal/skillgen` itself,
never under `kit/skills/`, so `kit.SkillsFS`'s existing
`//go:embed skills` + `mustSub` pattern (`kit/kit.go`) continues to
embed only final, installable Skill content — unchanged.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations. Table intentionally omitted.
