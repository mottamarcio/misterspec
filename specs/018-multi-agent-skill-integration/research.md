# Phase 0 Research: Multi-Agent Skill Integration

All unknowns spec.md's own Assumptions section deliberately deferred to
planning are resolved below. Facts about each of the five agents' own
integration conventions were verified against each agent's current
public documentation and, decisively, against
[GitHub Spec Kit's own integrations reference](https://github.github.io/spec-kit/reference/integrations.html)
— an actively maintained, real-world multi-agent framework installing
the same kind of Markdown Skill content this project does, into these
exact same five agents, today. No `NEEDS CLARIFICATION` markers
remain.

## 1. Per-agent integration convention (verified) — one shared format, five directories

**Decision**: All five agents, like Claude Code already, support the
same open **Agent Skills** convention: a directory per skill containing
a `SKILL.md` file with `name`/`description` YAML frontmatter followed
by a Markdown body — byte-identical in shape to misterspec's own
already-existing canonical Skill files
(`kit/skills/<name>/SKILL.md`). Only the target root directory differs
per agent:

| Agent | ID | Target directory | Verified against |
|---|---|---|---|
| Codex CLI | `codex` | `.agents/skills/<name>/SKILL.md` | [Build skills — ChatGPT/Codex docs](https://learn.chatgpt.com/docs/build-skills): "Codex scans `.agents/skills` in every directory from your current working directory up to the repository root." |
| Google Antigravity | `agy` | `.agents/skills/<name>/SKILL.md` | [Authoring Google Antigravity Skills — Google Codelabs](https://codelabs.developers.google.com/getting-started-with-antigravity-skills): project-level Skills directory is `<project-root>/.agents/skills/`, `SKILL.md` with `name` (optional, defaults to directory name) + `description` (mandatory) frontmatter; confirmed independently by [Spec Kit's own integrations reference](https://github.github.io/spec-kit/reference/integrations.html) ("Skills-based integration; skills are installed automatically"). |
| GitHub Copilot | `copilot` | `.github/skills/<name>/SKILL.md` | [Adding agent skills for GitHub Copilot — GitHub Docs](https://docs.github.com/en/copilot/how-tos/copilot-on-github/customize-copilot/customize-cloud-agent/add-skills): "create a .github/skills ... directory ... each skill should have its own directory ... with a SKILL.md file ... name (required) ... description (required)"; matches Spec Kit's own default (`speckit-<command>/SKILL.md` under `.github/skills/`) rather than its optional, secondary `--commands` flat-prompt-file mode, which this feature does not need. |
| Cursor | `cursor-agent` | `.cursor/skills/<name>/SKILL.md` | [Spec Kit's own integrations reference](https://github.github.io/spec-kit/reference/integrations.html): "`cursor-agent` \| `.cursor/skills` \| Skills-based integration; installs skills into `.cursor/skills`." |
| Devin for Terminal | `devin` | `.devin/skills/<name>/SKILL.md` | [Rules & AGENTS.md — Devin Docs](https://docs.devin.ai/cli/extensibility/rules): "project-scoped skills, use `.devin/skills/<name>/SKILL.md`"; confirmed by Spec Kit's own reference. |

Adapter IDs are exactly what the user specified (`agy`, `codex`,
`copilot`, `cursor-agent`, `devin`) — not
`docs/architecture-specification.md` §34's own illustrative
`antigravity` example ID, since the user's explicit choice for this
feature overrides an illustrative example, and matches Spec Kit's own
choice of the same five ID strings for the same five agents.

**Rationale**: Each is the current, documented directory that agent's
own tooling actually scans for Skills — an adapter pointed anywhere
else would install files that agent silently never discovers,
defeating User Story 1 entirely. Cross-checking against Spec Kit's own
integrations reference (a second, independently-verified source
installing this exact same class of content into these exact same
five agents in production today) resolved every case where a single
source's documentation was ambiguous or where an agent supports more
than one legacy mechanism (e.g. Copilot's older flat `.prompt.md`
files, Cursor's older flat `.cursor/commands/`) alongside its newer,
directory-per-skill convention.

**Alternatives considered**: An earlier research pass (before
consulting Spec Kit's own reference) found each agent's *other*,
non-Skills mechanism — Codex's deprecated flat "custom prompts",
Copilot's flat `.prompt.md` "prompt files", Cursor's flat
`.cursor/commands/*.md` "custom commands", Antigravity's flat
`.agent/workflows/*.md` "workflows" — and would have required a
transforming adapter for three of the five, flattening `SKILL.md` into
each agent's own bespoke single-file dialect. Rejected once the
Skills-based convention above was confirmed as each agent's own
currently-supported, more-capable mechanism (it also lets a Skill ship
bundled scripts/references/assets, which none of the flat-file
mechanisms support) — using it instead means zero content
transformation is needed anywhere in this feature.

## 1.1 The existing Claude Code adapter was re-verified, not just assumed correct

**Decision**: `internal/agents/claude/claude.go`'s own existing
`targetPath = ".claude/skills"` is confirmed still correct and is left
unmodified by this feature.

**Rationale**: Per the request to also double-check the already-shipped
integration while researching the five new ones: Spec Kit's own
integrations reference independently confirms the identical convention
— "Claude Code: Skills-based integration; installs skills in
`.claude/skills`" — matching 006-agent-adapter's own already-shipped
`targetPath` byte-for-byte. No drift was found between what 006 shipped
and what Claude Code currently scans.

**Alternatives considered**: N/A — this is a verification, not a
design decision; it is recorded here because the review was explicitly
requested and its outcome (no change needed) is itself a fact worth
documenting rather than leaving silently implicit.

## 2. One adapter shape for all six agents — a copy, not a transform

**Decision**: Every one of the five new adapters (`codex`, `devin`,
`agy`, `copilot`, `cursor-agent`) is implemented exactly like
`claude`'s own existing `Adapter.Install`
(`internal/agents/claude/claude.go`): call the already-existing
`installer.InstallFS(req.Skills, ".", "skill", target, req.Overwrite)`
unchanged, record the result via the already-existing
`agents.RecordInstall`, and return it. Only each adapter's own `id`,
`name`, and `targetPath` constants differ. Each new file is a
~20-line, near-verbatim copy of `claude.go`.

**Rationale**: Finding #1 means there is no format difference left to
bridge — 006-agent-adapter's own `Install` implementation, `Adapter`
interface, and `InstallFS`/`RecordInstall` machinery already fully
support this. Building any new transformation, parsing, or rendering
infrastructure now would be pure speculative surface with zero
demonstrated need (Constitution Principle IV) — the single biggest
simplification finding #1's own research unlocked.

**Alternatives considered**: The shared "flattening helper" design
from this research's own earlier pass (a new `SkillMeta`/
`ParseSkillMeta`/`InstallFlattened` trio in `internal/agents`).
Rejected outright once finding #1 confirmed no agent in this feature's
scope actually needs it — keeping it would have been unused
complexity violating Principle IV.

## 3. `builtin.Default()` registers all six adapters

**Decision**: `internal/agents/builtin.Default()` is extended to
register `claude.New()`, `agy.New()`, `codex.New()`, `copilot.New()`,
`cursoragent.New()`, `devin.New()` — one line each, matching the
existing pattern exactly.

**Rationale**: This is the one integration point that actually makes
the five new adapters selectable by `misterspec init` at all; no other
change is needed for discovery (`agents.Registry.List`/`Get` already
work generically over whatever adapters are registered, per
006-agent-adapter).

**Alternatives considered**: A build tag or configuration flag gating
which adapters are compiled in. Rejected — no demonstrated need
(Principle IV); the embedded kit and adapters are already a fixed part
of the single static binary, and five more small adapters add
negligible size.

## 4. Package layout mirrors `claude`'s own precedent

**Decision**: One new package per new adapter under `internal/agents/`
— `agy`, `codex`, `copilot`, `cursoragent`, `devin` — each exposing a
single `New() agents.Adapter` constructor, exactly `claude`'s own
existing shape. Package name `cursoragent` (not `cursor-agent`, not a
valid Go identifier) for the `cursor-agent` adapter ID.

**Rationale**: Matches the one convention already established by
006-agent-adapter; a reviewer who understands `claude/claude.go`
already understands the shape of all five new files, which is now
trivially true given finding #2's own single-shape conclusion.

**Alternatives considered**: One combined `internal/agents/otheradapters`
package holding all five. Rejected — breaks the established
one-package-per-adapter convention for no benefit, and would make a
future per-adapter change (e.g. an agent changing its own scanned
directory) touch a shared file instead of an isolated one.

## 5. Which Skills get the new "request a Context Pack" step, and with which intent

**Decision**: Exactly the four canonical Skills spec.md names, each
gaining one new required operation and one new early `Procedure` step
requesting `internal context <target> --intent <intent>` right after
its own existing `resolve`/`inspect` steps, before any Skill-specific
exploration:

| Skill | File | Intent |
|---|---|---|
| `/implement` | `kit/skills/implement/SKILL.md` | `implementation` |
| `/create-plan` | `kit/skills/create-plan/SKILL.md` | `planning` |
| `/create-tasks` | `kit/skills/create-tasks/SKILL.md` | `tasks` |
| `/analyze` | `kit/skills/analyze/SKILL.md` | `validation` |

**Rationale**: The first three map onto `contextengine.Intent`'s own
constant names directly. `/analyze` is mapped to `validation`, not
`analysis`, because `docs/context-engine-implementation.md` §15.4
itself groups "Validation / analysis" as one combined workflow, and
`/analyze` is misterspec's own concrete Skill for exactly that
combined stage (producing the Validation artifact) — there is no
second, distinct "analysis" Skill in this project's own canonical set
to which `IntentAnalysis` would separately apply.

**Alternatives considered**: Introducing a fifth Skill or splitting
`/analyze`'s own scope to use `IntentAnalysis` for some sub-step.
Rejected as unrelated new scope — spec.md's Assumptions explicitly
name exactly these four Skills, matching
`docs/context-engine-implementation.md` §30 Phase 10's own explicit
list.

## 6. Every updated Skill keeps its own existing "Deterministic Operations"/"Procedure" structure

**Decision**: For each of the four Skills, add exactly one line under
"Required operations" (`internal context <target> --intent <intent>`)
and exactly one new early numbered step in "Procedure" (immediately
after the existing `resolve`/`inspect` steps), plus one sentence
elsewhere in the file noting the pack informs but does not replace the
Skill's own judgment, and one sentence making explicit that further
exploration remains allowed (FR-007) — no other section of any of the
four files changes.

**Rationale**: Every one of misterspec's canonical Skills already
follows one fixed, load-bearing section structure (`Purpose`,
`Invocation`, ..., `Deterministic Operations`, `Procedure`, ...,
`Completion Contract`) that 009-canonical-skills-content established;
inserting the new step into that same, already-proven structure is the
smallest possible change consistent with FR-009/FR-010 (this feature
changes *how context is gathered*, nothing else about any Skill).

**Alternatives considered**: Rewriting each Skill's `Procedure` from
scratch around the Context Pack. Rejected — unnecessary churn against
FR-010's own explicit constraint, and higher risk of accidentally
changing what a Skill considers complete.

## 7. Fallback behavior on a failed Context Pack request

**Decision**: Each updated Skill's own instructions state: if the
`internal context` request fails for any reason, proceed exactly as
this Skill already did before this feature (its own existing Required
Context / exploration approach), rather than stopping.

**Rationale**: Directly satisfies FR-008. The Context Engine (011-017)
is read-only and side-effect-free by its own established contract, so
a failed request has nothing to roll back — the only correct response
is "ignore it and continue," matching `docs/context-engine-
implementation.md` §23's own explicit "bootstrap/retrieval
optimization, not a sandbox."

**Alternatives considered**: Treating a failed Context Pack request as
a new Failure Condition that stops the Skill. Rejected — directly
contradicts FR-008 and would make every one of these four Skills
strictly less reliable than before this feature, the opposite of the
intended effect.
