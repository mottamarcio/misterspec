# Content Contract: Project Documentation Site and README

Not a code contract — the acceptance checklist `/speckit-tasks` and
implementation must satisfy in full before this feature is considered
complete. Reconciled against the actual published content once written
(same discipline as every prior feature's own contract).

## README (FR-001–FR-004)

- [X] States what misterspec is and the problem it solves, in the
      first paragraph, no prior context assumed.
- [X] Installation instructions that alone produce a working
      `misterspec` binary.
- [X] A minimal quickstart (first commands to run).
- [X] Exactly one clear link to the published documentation site.
- [X] No leftover placeholder content from the pre-feature one-line
      README.

## Site-wide (FR-005, FR-010, FR-011)

- [X] Every page reachable from every other page via the shared
      sidebar (`nav.js`).
- [X] Every page renders correctly at a 400px viewport width, no
      horizontal scroll.
- [X] The published URL requires no authentication, no local build.

## Concept coverage (FR-006)

- [X] `index.html` explains the deterministic/semantic split (the
      `misterspec` binary vs. the coding agent) before any
      feature-specific detail.
- [X] `index.html` explains the artifact lifecycle (Constitution →
      Program → Feature → Spec → Plan → Tasks → Implementation →
      Validation).

## Feature coverage (FR-007) — data-model.md's own inventory table

- [X] All 19 Specs (001-019) appear on exactly one page each, per
      data-model.md's own Feature-to-page inventory.
- [X] Every feature page states what/why/how-it-fits/example for every
      Spec it covers (data-model.md's own content model).
- [X] `context-engine.html` presents its seven Specs (011-017) as one
      connected pipeline narrative, not seven disconnected blurbs.

## Command coverage (FR-008, FR-009) — data-model.md's own inventory table

- [X] All 14 currently-registered commands appear in `commands.html`.
- [X] Every command's documented flags match its own current source
      exactly (data-model.md's own command inventory table).
- [X] Every command has at least one real, copyable example invocation
      and a real (or realistically-shaped) expected result.

## Non-goals (explicitly out of scope, per spec.md's own Assumptions)

- No mechanism to keep the site automatically synchronized with future
  features — this is a written-once snapshot as of Specs 001-019.
- No change to any Go source, template, Skill, or Context Engine
  behavior (FR-012).
- No dark-mode toggle, search, or versioning — not requested, not
  required by any FR/SC.
