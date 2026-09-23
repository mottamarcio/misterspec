# Implementation Plan: Context Pack completo e contrato de saída versionado

**Branch**: `033-context-pack-output-contract` | **Date**: 2026-09-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/033-context-pack-output-contract/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

`internal/cli/internalcmd/context.go`'s `renderContextItems` (context.go:123-144) never includes a `content` field, with or without `--render`; `--render` only adds one Markdown blob (`contextengine.Render`) that duplicates path/heading/content/reasons as prose, never per-item structured content — exactly the gap every consuming Skill documents ("treat the Context Pack as a starting point," then separately re-read files). Worse than the spec's own premise suggested: **every** Context Pack item's location is body-relative, not just wikilinks — `chunkArtifact` (collector.go:206-226) builds every `Candidate` from `artifacts.ReadBody` (post-frontmatter) + `artifacts.ParseDocument`, so `StartLine`/`EndLine` on every single item are relative to the post-frontmatter body, inherited unchanged through `Chunk` → `Candidate` → `ScoredCandidate` → `ResultItem`. This plan adds a new, additive `--mode=package` output (full per-item content, file-absolute location, per-item fingerprint), a `schema_version` envelope field, output-size diagnostics that reflect what was actually serialized, and leaves today's default (`--dir`/no flags) and `--render` behavior byte-for-byte unchanged for zero-migration backward compatibility.

## Technical Context

**Language/Version**: Go 1.23.4 (existing `go.mod`)
**Primary Dependencies**: existing internal packages only — `internal/context` (`contextengine`), `internal/artifacts` (`ReadBody`, `ParseDocument`, `Chunks`, `EstimateTokens`), `internal/cli/internalcmd`. No new external dependency.
**Storage**: N/A — the Context Pack is computed fresh per request from the filesystem and the disposable search index (`internal/context/index`, already rebuilt/synced on every `internal context` call); nothing new is persisted (Constitution Principle III).
**Testing**: Go `testing` — unit tests for the new file-absolute line-offset helper and per-item fingerprint (pure functions), and CLI-level tests for the new `--mode=package` output shape and the unchanged default/`--render` paths, matching `internal/cli/internalcmd`'s and `internal/context`'s existing `*_test.go` conventions.
**Target Platform**: `misterspec` CLI binary, via `internal context <id>`.
**Project Type**: Single Go module — extends `internal/context` and `internal/cli/internalcmd/context.go`; one small additive helper in `internal/artifacts`.
**Performance Goals**: No regression vs. today's single Collect→Rank→ApplyBudget pipeline; `--mode=package` reuses the same already-collected `Candidate.Content` (already held in memory) rather than re-reading files.
**Constraints**: MUST NOT change the shape or values of the existing default (no flags) or `--render` output (Constitution-adjacent product commitment: spec.md FR-008, SC-003) — every new field is additive only; MUST NOT introduce a new persisted cache/state (Principle III) — the per-item fingerprint is computed from already-in-memory content, not a new on-disk index; MUST NOT add a new package layer (Principle IV) — stays inside `internal/context` and `internal/cli/internalcmd`, with one small addition to `internal/artifacts` (the existing owner of frontmatter/body splitting).
**Scale/Scope**: Touches `internal/artifacts/markdown.go` (new line-offset-aware read), `internal/context/collector.go` (`chunkArtifact` applies the offset), `internal/context/result.go`/`budget.go` (absolute `StartLine`/`EndLine` flow through unchanged field names — no type change needed, only the values become correct), `internal/context/render.go` (unchanged — still consumes the same `Result` shape), new `internal/context/pack.go` (per-item fingerprint + full-package assembly), `internal/cli/internalcmd/context.go` (new `--mode` flag, `schema_version`, output-size diagnostics).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Semantic/Deterministic Separation** — PASS. Every new behavior (offset arithmetic, content fingerprinting, output-size estimation, mode selection) is purely mechanical; no semantic judgment about *which* content matters is added — that stays in `Collect`/`Rank`/`ApplyBudget`, untouched.
- **II. Deterministic Operations Are the Only Mutation Primitive** — PASS. This feature is entirely read-only output shaping; it adds no `create`/mutation path and allocates no IDs.
- **III. Filesystem Is the Single Source of Truth** — PASS. No new persisted state; the fingerprint and absolute location are recomputed from the same file read `Collect` already performs, every call.
- **IV. Simplicity First — YAGNI & Minimal Configuration** — PASS. No new package layer: the line-offset helper extends `internal/artifacts` (already owns frontmatter/body splitting — `splitFrontmatter`, `ReadBody`); the full-package assembly and fingerprint live in `internal/context` (already owns Context Pack shaping — `render.go`, `result.go`). No new config field — `--mode` is a request-scoped CLI flag, not a project-wide setting.
- **V. Test-First Discipline** — Applies. The new offset helper, per-item fingerprint, and `--mode=package` output MUST get unit/CLI tests before/alongside implementation; the unchanged default/`--render` paths MUST get an explicit regression test proving byte-for-byte parity with pre-change behavior (spec.md FR-008/SC-003).
- **VI. Clean Code & SOLID** — PASS. `internal/artifacts` keeps owning "how is a Markdown file structured" (frontmatter/body split, line accounting); `internal/context` keeps owning "what does a Context Pack response look like" — no responsibility blur.
- **VII. Explicit Mutation Boundaries** — N/A. No artifact-owning Skill or operation gains write access; this is read-only response shaping exactly like the rest of `internal/context`.
- **VIII. Safety by Construction** — N/A. No new filesystem write path.
- **IX. Transparent, Machine-Readable Contracts** — Applies directly: this is the feature that *adds* `schema_version` and a documented, versioned contract to `internal context`'s output — the clearest instance of Principle IX in this plan. `--mode=package` combined with `--render` MUST be rejected as a structured `invalid_argument` error (never silently pick one), consistent with `classify`'s existing sentinel-to-code mapping.

No violations requiring justification; Complexity Tracking is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/033-context-pack-output-contract/
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
├── artifacts/
│   ├── markdown.go        # ReadBody (existing, unchanged) — add ReadBodyWithOffset,
│   │                        # reusing splitFrontmatter to also report the body's own
│   │                        # 1-indexed starting line in the whole file
│   └── markdown_test.go    # extended
├── context/
│   ├── collector.go        # chunkArtifact: apply the new offset to every Chunk's
│   │                        # StartLine/EndLine before building Candidates — the one
│   │                        # place every Candidate's location is currently wrong
│   ├── pack.go              # NEW: per-item content fingerprint; full-package
│   │                        # ("package" mode) item assembly; output-size estimation
│   │                        # for the requested mode
│   ├── pack_test.go         # NEW
│   ├── result.go            # unchanged type shapes — StartLine/EndLine now simply
│   │                        # carry correct (file-absolute) values
│   └── collector_test.go    # extended (absolute-line regression cases)
└── cli/internalcmd/
    ├── context.go            # new --mode flag ("manifest" default | "package" |
    │                          # "markdown"), schema_version, output-size diagnostics;
    │                          # default/--render paths byte-for-byte unchanged
    └── context_test.go       # extended
```

**Structure Decision**: No new top-level package. One additive function in `internal/artifacts` (the existing frontmatter/body-splitting owner), one new file in `internal/context` (`pack.go`, alongside its existing `render.go`/`result.go`/`budget.go` siblings), and additive changes to `internal/cli/internalcmd/context.go` — matching Constitution Principle IV/VI's existing package boundaries.

## Complexity Tracking

*No Constitution Check violations — this section is intentionally empty.*
