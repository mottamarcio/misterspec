# Phase 0 Research: Project Documentation Site and README

All unknowns spec.md's own Assumptions section deliberately deferred to
planning are resolved below. No `NEEDS CLARIFICATION` markers remain.

## 1. Site source location and build approach

**Decision**: A new top-level `site/` directory holding plain,
hand-authored static HTML files (one per page) plus a small
`site/assets/` folder (`nav.js`, `styles.css`) — no bundler, no Node
toolchain, no build step. TailwindCSS and Iconify are both loaded from
their own CDNs directly in each page's `<head>` (`https://cdn.tailwindcss.com`
and the Iconify Icon web component script).

**Rationale**: FR-005 requires "no local build step for a visitor to
read it," and this project has no existing JS/Node tooling to extend —
introducing one (webpack/vite/a static-site generator) purely for a
one-time, write-once documentation site (per spec.md's own Assumptions:
"written once... ongoing maintenance... out of scope") would be
speculative infrastructure with no demonstrated need (Constitution
Principle IV, applied here even though this feature ships no Go code).
A separate `site/` directory (not reusing the existing `docs/`, which
holds this project's own internal architecture specifications) keeps
the two concerns — internal design docs vs. a public-facing reference
— cleanly distinct.

**Alternatives considered**: A static-site generator (Hugo, MkDocs,
Docusaurus). Rejected — each adds a new toolchain/runtime dependency
this Go-only project doesn't otherwise have, for a site whose own
scope (spec.md's Assumptions) is explicitly a one-time snapshot, not a
continuously regenerated reference. Reusing `docs/` as the Pages
source. Rejected — `docs/` already has an established, different
purpose (internal architecture specifications consumed by planning),
and GitHub Pages serving raw `.md` files there would produce an
inconsistent, half-styled result alongside the new HTML pages.

## 2. Shared navigation via one small script, not duplicated markup

**Decision**: `site/assets/nav.js` holds one JavaScript array
describing the site's entire navigation tree (page title, path, and
group), and injects the sidebar `<nav>` markup into a fixed
`<div id="sidebar">` element every page includes. Each page's own
`<body>` is otherwise just its own content.

**Rationale**: With 9 static pages (research.md #3) and no templating
engine, duplicating a hand-written sidebar in every file risks drift
(a link added to one page's copy but not another's) — the exact class
of bug DRY (Constitution Principle VI) exists to prevent, achievable
here with nothing more than one small, dependency-free script.

**Alternatives considered**: Copy-pasting the same `<nav>` block into
every HTML file. Rejected — the single most likely source of visible
inconsistency (FR-010's own "persistent navigation... from any page")
as pages are added or reordered.

## 3. Site map — 9 pages, grouped by real feature relationships

**Decision**:

| Page | Covers | Purpose |
|---|---|---|
| `index.html` | — | Home: what misterspec is, the deterministic/semantic split, the artifact lifecycle diagram (FR-006) |
| `getting-started.html` | — | Install + a first end-to-end run |
| `workflow.html` | — | The full Spec-Driven lifecycle (Constitution → Program → Feature → Spec → Plan → Tasks → Implementation → Validation) with the actual Skill/command invoked at each stage — the concrete, illustrated usability example spec.md calls for |
| `foundation.html` | Specs 001-005 | Core repository foundation: deterministic ID/path/fingerprint primitives, read operations, entity creation, structural validation, the embedded kit |
| `agents-cli.html` | Specs 006-010 | The Agent Adapter layer, project bootstrap, the Cobra-based CLI, canonical Skill content, the interactive `init` TUI |
| `context-engine.html` | Specs 011-017 | The Context Engine's own seven phases (wikilinks → references/backlinks → document/chunk model → disposable index → collector → ranking/budgeting → the `internal context` command) as one connected narrative with a pipeline diagram, one anchored section per phase |
| `multi-agent-skills.html` | Spec 018 | The five additional coding-agent adapters and the four Context-Pack-aware Skills |
| `dogfooding.html` | Spec 019 | The evaluation methodology and its own real findings/outcome |
| `commands.html` | — | The complete command reference (FR-008), one anchored section per command |

**Rationale**: Matches spec.md's own Assumption that exact grouping is
a planning decision, while satisfying FR-007's own "how it relates to
the features immediately before and after it" by grouping Specs into
their own real, already-established narrative arcs (the same arcs each
feature's own spec.md "Input" line already described when it said
"building on X's own Y") rather than one page per Spec number, which
would fragment closely related work (e.g. the Context Engine's seven
phases) across seven disconnected pages with heavy cross-linking
instead of one coherent one.

**Alternatives considered**: One page per Spec (20 pages). Rejected —
directly contradicts FR-007's own "relates to the features around it"
requirement by presenting tightly sequential work as isolated units;
also meaningfully more navigation surface for a site whose own content
volume (a documentation snapshot, not a growing wiki) doesn't need it.
One single long page for the whole project. Rejected — directly
contradicts FR-010's own GitBook-style, one-topic-per-page requirement
and Edge Cases' own "obvious where a new entry belongs" test.

## 4. Command reference content is generated from the real, current command tree

**Decision**: `commands.html` documents exactly the 14 commands
`internal/cli/internal.go` and `internal/cli/root.go` currently
register — the public `init` command, and all 13 `internal` commands
(`resolve`, `inspect`, `parent`, `children`, `create`,
`create-artifact`, `fingerprint`, `inventory`, `validate`, `status`,
`references`, `backlinks`, `context`) — each with its own real flag
set (read directly from each command's own `internalcmd/*.go` file,
never guessed) and one real, runnable example built against a small,
already-initialized example project.

**Rationale**: FR-009 requires every example to reflect actual, shipped
behavior; the only way to guarantee that is reading each command's own
current source rather than working from memory of what it "should" do
— several commands (e.g. `internal context`'s own `--budget` flag)
have already gone through a real, documented mid-implementation
correction (017's own research.md), so working from an assumption
rather than the current file risks documenting stale behavior.

**Alternatives considered**: Documenting only the public `init`
surface, treating `internal` commands as a hidden implementation
detail. Rejected — spec.md's own Assumptions explicitly call this out:
"internal commands [are] a real, agent-facing contract (Constitution
Principle IX), not a hidden implementation detail to omit."

## 5. GitHub Pages publishing mechanism

**Decision**: A new GitHub Actions workflow,
`.github/workflows/deploy-docs.yml`, using the official
`actions/configure-pages`, `actions/upload-pages-artifact`, and
`actions/deploy-pages` actions to publish `site/`'s own contents
verbatim, triggered on push to `dev` — the branch this project's own
established practice (every prior feature, 001-019) actually merges
and pushes to.

**Rationale**: This is the current, GitHub-recommended Pages mechanism
(no branch/folder picker needed in repository settings beyond
selecting "GitHub Actions" as the Pages source once) and requires zero
new tooling beyond a workflow file — this repository has no existing
CI at all, so this is also the project's first workflow, kept minimal
and scoped to exactly this one job. Triggering on `dev` (not `main`)
matches how this project actually ships today; retargeting to `main`
later is a one-line change to the workflow's own `on.push.branches`
list if the project's own branching practice changes.

**Alternatives considered**: Serving directly from a `docs/` folder on
a branch (the legacy, no-Actions Pages mode). Rejected — would force
reusing or renaming the existing `docs/` directory (research.md #1)
and offers no meaningful simplicity gain over one small workflow file.
Triggering on `main`. Rejected for now — nothing in this project's own
real history has ever pushed to `main`; a workflow that never fires
would satisfy FR-005 in principle but never in practice.

## 6. Visual design: Tailwind's own default `slate` palette, no custom config

**Decision**: `slate-900` (and the rest of Tailwind's own built-in
`slate` scale) is used directly as utility classes
(`bg-slate-900`, `text-slate-900`, `border-slate-900`, etc.) — no
`tailwind.config` customization is needed, since `slate` already ships
as one of Tailwind's own default palettes.

**Rationale**: The user's own explicit accent choice is already a
first-class, zero-configuration Tailwind color — the simplest possible
way to honor it exactly (Principle IV: no config layer where none is
needed).

**Alternatives considered**: Defining a custom CSS variable / Tailwind
theme extension for the accent. Rejected — pure ceremony for a color
Tailwind already ships by name.

## 7. Icons via the Iconify web component, not a downloaded icon font

**Decision**: `<script src="https://code.iconify.design/iconify-icon/2.1.0/iconify-icon.min.js">`
loaded once per page, with icons used inline as
`<iconify-icon icon="lucide:terminal"></iconify-icon>`-style elements
— no local icon assets, no separate icon font subset build.

**Rationale**: The user asked for Iconify specifically; the web
component is Iconify's own current, officially recommended
zero-build integration path, consistent with research.md #1's own
no-build-step decision, and gives access to Iconify's full multi-set
icon catalog without choosing and vendoring one specific icon font.

**Alternatives considered**: Downloading a specific Iconify icon set as
static SVGs. Rejected — more moving parts (a fixed set of pre-chosen
SVG files to maintain) for no benefit over the CDN-served web
component, given this site has no offline-availability requirement.
