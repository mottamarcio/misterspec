# Implementation Plan: Project Documentation Site and README

**Branch**: `020-docs-site-readme` | **Date**: 2026-09-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/020-docs-site-readme/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

A documentation-only feature (no Go code, no product behavior change):
rewrite the root `README.md` into a real front door (what/why,
install, quickstart, a link to the full site), and publish a 9-page,
GitBook-style static HTML site under a new `site/` directory —
TailwindCSS and Iconify both loaded from CDN, no build step, a small
shared `nav.js` driving one consistent sidebar across every page — via
a new GitHub Actions workflow that deploys `site/` to GitHub Pages on
every push to `dev`. The site covers misterspec's own core concept,
every shipped feature (001-019, grouped into their own real narrative
arcs), and a complete reference for all 14 currently-registered
commands, each with a real, current, runnable example.

## Technical Context

**Language/Version**: Plain HTML5, vanilla JavaScript (ES2020+, no transpilation), CSS — no new Go code. TailwindCSS (CDN, Play CDN build) and the Iconify Icon web component (CDN) for styling/icons.
**Primary Dependencies**: None added to `go.mod`. Two external CDN scripts referenced by every HTML page (`cdn.tailwindcss.com`, `code.iconify.design`) — no local `node_modules`, no package manager, no lockfile.
**Storage**: N/A — static files only.
**Testing**: No automated test suite (this is static content); verification is a manual link/render check (quickstart.md) plus confirming every command documented in `commands.html` matches its own current `internalcmd/*.go` flag set and JSON shape at time of writing (FR-009). 001-019's own full Go test suites are unaffected and re-run unmodified as the standing regression gate (nothing here touches `internal/`, `cmd/`, or `kit/`).
**Target Platform**: Any modern browser, served as a static site via GitHub Pages; must remain fully usable at a 400px viewport (FR-011).
**Project Type**: Documentation/static-site addition to an existing single-module Go CLI project. New content lives under `site/` (the published site) and `.github/workflows/` (the one new deploy workflow); `README.md` is rewritten in place.
**Performance Goals**: N/A beyond "loads and renders as a normal static page" — no dynamic data, no client-side routing framework.
**Constraints**: No local build step for a visitor (FR-005) — every page must be viewable by opening the deployed URL directly. No change to any existing Go source, template, Skill, or Context Engine behavior (FR-012). Every documented example must reflect real, current, verified behavior (FR-009) — verified by reading each command's own current source, not from memory.
**Scale/Scope**: `README.md` (rewritten); `site/` — 9 HTML pages (research.md #3) + `site/assets/{nav.js,styles.css}`; `.github/workflows/deploy-docs.yml` (new, this repository's first CI workflow).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass (N/A).** No code, no operation, no agent judgment boundary exists in this feature — it is prose and markup describing already-shipped behavior. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass (N/A).** Nothing here allocates an ID, path, or hash. |
| III. Filesystem Is Single Source of Truth | **Pass (N/A).** No new persisted state of any kind; the published site is a static rendering of already-committed content. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised.** No Node toolchain, no static-site generator, no custom Tailwind config, no local icon asset pipeline — every one of these was considered and rejected for lack of demonstrated need (research.md #1, #6, #7). One small shared `nav.js` is the only new "infrastructure," justified by a real, demonstrated DRY risk (research.md #2). |
| V. Test-First Discipline | **N/A, documented.** No deterministic logic is introduced; verification is manual content accuracy (Technical Context, Testing) against each command's own real, current source. |
| VI. Clean Code & SOLID | **Pass (N/A new Go code).** Applied in spirit to the site itself: `nav.js` centralizes navigation once (DRY) rather than duplicating it across 9 files (research.md #2). |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** This feature writes only `README.md`, `site/`, and one new workflow file — it never touches `internal/`, `cmd/`, `kit/`, or any prior feature's own artifacts (FR-012). |
| VIII. Safety by Construction | **Pass (N/A new surface).** No new path-handling code; the deploy workflow uses GitHub's own first-party Pages actions unmodified. |
| IX. Transparent, Machine-Readable Contracts | **Pass, directly honored.** The command reference (`commands.html`) documents the `internal` surface as the real, agent-facing contract it already is (Constitution's own framing), not a hidden detail — matching spec.md's own explicit Assumption. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/020-docs-site-readme/
├── plan.md                          # This file (/speckit-plan command output)
├── research.md                      # Phase 0 output (/speckit-plan command)
├── data-model.md                    # Phase 1 output (/speckit-plan command)
├── quickstart.md                    # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── content-inventory.md         # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md              # /speckit-specify quality checklist
└── tasks.md                         # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
misterspec/
├── README.md                                  # REWRITTEN — real front door (FR-001–FR-004)
├── .github/
│   └── workflows/
│       └── deploy-docs.yml                      # NEW — this repo's first CI workflow (research.md #5)
├── site/                                          # NEW — the published GitHub Pages site
│   ├── index.html                                   # Home / concept (FR-006)
│   ├── getting-started.html                           # Install + first run
│   ├── workflow.html                                    # The Spec-Driven lifecycle, illustrated end to end
│   ├── foundation.html                                    # Specs 001-005
│   ├── agents-cli.html                                      # Specs 006-010
│   ├── context-engine.html                                   # Specs 011-017, one flagship page
│   ├── multi-agent-skills.html                                 # Spec 018
│   ├── dogfooding.html                                           # Spec 019
│   ├── commands.html                                               # Full command reference (FR-008)
│   └── assets/
│       ├── nav.js                                                    # Shared sidebar navigation (research.md #2)
│       └── styles.css                                                 # Small non-Tailwind-utility CSS (if any)
├── go.mod, go.sum, internal/, cmd/, kit/          # unchanged — no product code touched (FR-012)
└── docs/                                          # unchanged — remains internal architecture specs, not the public site (research.md #1)
```

**Structure Decision**: Everything this feature adds lives under three
new locations — `README.md` (rewritten in place), `site/` (the new
published content), and one new workflow file — with zero changes
anywhere under `internal/`, `cmd/`, `kit/`, or `docs/`. This is the
smallest footprint consistent with the spec's own explicit
documentation-only scope (FR-012).

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
