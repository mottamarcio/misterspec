# Research: Skills Enxutas e Contratos de Integração Testáveis

## 1. Where does the actual duplication PROP-14 targets live?

**Investigation**: Compared the 10 canonical Skills (`kit/skills/*/SKILL.md`,
2378 lines total) section by section. Sections that share a heading
(all 10 Skills share the same 26 headings per
`docs/architecture-specification.md` §39, machine-checked by
`internal/example/skills_content_test.go`) are almost never
byte-identical prose — "Idempotency", "Forbidden Mutations",
"Interaction Rules" etc. are genuinely Skill-specific (each explains a
different mutation boundary or resume behavior). However, grepping
`## Deterministic Operations` bodies across all 10 files found real,
literal, multi-line duplication: the `internal resolve SPEC-###` preamble
paragraph appears verbatim in 4 Skills, "Use misterspec operations for
every mechanical repository step[.]( this Skill has one for.)" appears
(with two exact wordings) in 9 of 10 Skills, and the "using this Skill's
own Required/Optional Context above instead — it is never a Failure
Condition" fallback-tolerance sentence recurs 3 times.

**Decision**: The duplication is not whole-section but recurring
canonical *phrases/paragraphs* scattered inside otherwise bespoke
sections. Treat the "canonical source" (spec FR-001) as a small,
explicit set of named text fragments (e.g. `resolve-preamble`,
`mechanical-steps-note`, `fallback-tolerance-note`) authored once, with
each Skill's own Markdown composed from {shared fragments} + {its own
bespoke body} by a build-time generator — never a runtime include,
satisfying spec FR-001's "sem depender de includes que o agente não
suporte" directly, since composition happens before the file is
committed/embedded, not when an agent reads it.

**Alternatives considered**:
- *Full per-section templating* (treat every one of the 26 sections as
  parametrized template): rejected — most sections are genuinely
  bespoke; forcing them through a template adds indirection (Principle
  IV, YAGNI) without removing real duplication, and would make Skills
  harder to author/review as plain Markdown.
- *Do nothing automated, just a style guide*: rejected — spec FR-001
  requires that a shared-rule edit apply everywhere without manual
  per-file edits; a style guide alone can't guarantee that (US1
  Acceptance Scenario 1).
- *Per-agent content variants*: investigated and rejected — every
  currently supported adapter (`internal/agents/{claude,codex,agy,
  cursoragent,copilot,devin}`) installs byte-identical Skill content,
  differing only in `TargetPath()` (confirmed via
  `internal/example/multi_agent_skill_integration_quickstart_test.go`).
  There is no existing per-agent content divergence to consolidate;
  "each integration" in the original PROP-14 backlog text maps, in this
  codebase, to "each Skill", not "each agent". No per-agent variant
  mechanism is introduced.

## 2. How does consolidation stay verifiable, not just aspirational?

**Decision**: The generator is deterministic and its output is
committed (the same `kit/skills/<name>/SKILL.md` files `kit.SkillsFS`
already embeds — no change to the embed or to `Adapter.Install`). A new
test regenerates every Skill into a temp directory from the same
canonical fragments + manifests and byte-diffs the result against the
committed files, failing loudly on drift — the same golden-file
discipline Constitution Principle V already requires for adapter output
and `config.yaml`. A hand-edit that isn't reflected back into a fragment
or manifest is caught the same way an un-regenerated golden fixture is
caught today.

**Rationale**: Keeps `kit/skills/` as the actual source of truth
developers read and Git diffs show (Constitution Principle III's spirit
extended to authoring, not just runtime state), while still guaranteeing
FR-001/US1's "edit once, apply everywhere" property through a checked
invariant rather than developer discipline alone.

## 3. How does "verification promised == real capability" get checked?

**Investigation**: `internal/validation/findings.go` already defines a
closed, enumerable set of `Code*` constants (e.g.
`CodeUncoveredRequirement`, `CodeUnresolvedDependency`,
`CodeDuplicateID`) — every structural check the binary can actually
perform is one of these codes, used identically by callers and by
`internal/validation/*_test.go`. `internal/example/skills_content_test.go`
already cross-checks every `internal <op>` mention in a Skill against
`knownInternalCommands`, per-Skill, via
`skillOperationsAllowlist` (spec FR-004 — already delivered by Spec
009/031/032 work, not new).

**Decision**: Extend the same test file's approach with two more checks
(spec FR-005, FR-006), reusing its existing per-Skill-allowlist pattern
rather than inventing a new mechanism:
- **FR-005 (example validity)**: every fenced/backtick `misterspec …`
  invocation example in a Skill's body MUST parse against the real
  Cobra command tree (`internal/cli`) — flags and subcommand names MUST
  exist. Implemented by building the real `*cobra.Command` tree in-test
  and calling `Find()`/flag lookup per example, not a hand-maintained
  second copy of the CLI surface.
- **FR-006 (promised verification ⇄ real capability)**: a new, explicit
  per-Skill allowlist (mirroring `skillOperationsAllowlist`) mapping
  each Skill's textual verification claims (e.g. "valida cobertura de
  requisitos") to one or more real `validation.Code*` constants (or a
  real `internal <op>` capability for non-validation claims, e.g.
  `internal fingerprint`). A Skill whose claim has no allowlisted code
  fails the test — the same shape of failure as an unknown `internal`
  command today.

**Alternatives considered**: Parsing Skill prose with an LLM at test
time to auto-detect "promises" — rejected: non-deterministic, and this
is exactly the class of decision Constitution Principle I reserves for
compile-time/mechanical checks, not semantic judgment, when a closed
enumeration (the Code constants) already exists.

## 4. What are "representative integrations" for smoke tests (FR-007/SC-004)?

**Investigation**: The six registered adapters' `TargetPath()`s
(`internal/agents/builtin`) collapse to four distinct install
directories: `.claude/skills`, `.agents/skills` (shared by `codex` and
`agy`), `.cursor/skills`, `.github/skills` (copilot), `.devin/skills`.
`internal/example/multi_agent_skill_integration_quickstart_test.go`
already exercises install-time behavior (file placement, overwrite
handling) across several of these; it does not exercise any Skill's
*content* being followed end-to-end (i.e., the deterministic operations
a Skill instructs actually chain together correctly for an agent
reading that file).

**Decision**: Fix the representative set to three adapters covering
three distinct target directories: `claude-code` (`.claude/skills`),
`cursor-agent` (`.cursor/skills`), and `copilot` (`.github/skills`) —
satisfying SC-004's "destinos de instalação distintos" with the minimum
set that already covers every distinct target shape (a fourth,
`.agents/skills`-sharing pair, adds no new target-path coverage).
"Smoke test" here means: install the real generated Skills into a
temp project for each of the three adapters, then run the literal
`misterspec internal` command sequence one representative Skill
documents (`mister-implement`'s `resolve → context/prepare → validate`
chain) against a fixture repo, asserting the sequence completes and
produces the JSON shape that Skill's own text says it will get back.
This is Skill-content-and-binary integration, not agent-behavior
simulation — no LLM is invoked (Constitution: no network in core
commands/tests).

**Alternatives considered**: Smoke-testing all six adapters on every
change — rejected as disproportionate cost for the same target-path
coverage (Principle IV); the fixed three-adapter set is documented so a
future Spec can widen it if a target-path-independent regression is
ever found.

## 5. How does Context Engine fallback get recorded (FR-008)?

**Investigation**: `internal/eval.Metrics` (037-eval-quality-efficiency)
already has an `ExtraReads int` field, populated by hand into a
`RunRecord` after a live run, with an explicit `estimated` sibling
convention for other figures. Several Skills already document a
fallback-tolerant path in prose today (e.g. mister-implement's Context
Pack section: "This Skill remains free to read further repository
files… whenever the pack alone is insufficient" — an *un-tracked*
fallback). There is no existing automatic instrumentation inside
`internal context`/`internal prepare` that detects "the agent had to
fall back"; that decision is made by the agent reading the Skill, not by
the binary, matching Principle I (semantic judgment stays with the
agent).

**Decision**: Do not add a new automatic telemetry pipeline (no new
persistent state, Principle III). Instead: (a) add one new field to
`eval.Metrics`, `ContextFallbacks int` (count of times a Context Pack
was insufficient and the agent read further/re-queried during the
task), following the exact `ExtraReads`-adjacent convention already
established; (b) update the relevant Skills' own Deterministic
Operations / canonical fragment text to instruct the agent to note a
fallback occurrence for its own subsequent `RunRecord`, rather than
silently absorbing the extra read as if the primary path had succeeded.
`internal/eval`'s existing `Validate()`/comparison logic (`compare.go`)
already treats `ExtraReads` as a compared dimension — `ContextFallbacks`
joins the same comparison path with no new comparison mechanism needed.

**Alternatives considered**: Auto-detecting fallback inside `internal
context` itself (e.g. flagging when a candidate set is thin) — rejected:
a "the pack was insufficient" judgment depends on what the agent
actually needed, which the binary cannot know (Principle I); the
binary already reports its own diagnostic scores (`--diagnostic-scores`
from Spec 036) for the agent to reason from, that is as far as
mechanical inference can go.

## 6. Sizing the reduction (SC-001's "at least 20%")

**Investigation**: Current combined Skill body size is 2378 lines
across 10 files. The literal recurring fragments found in #1 above
total roughly 6-9 lines repeated 3-9 times each — extracting them
removes duplication but is a modest fraction of 2378 lines on its own.
Reaching a clearly *measurable* reduction additionally requires
trimming genuinely redundant restatement within a single Skill's own
body where multiple sections currently re-explain the same task-oriented
preparation / Context Pack consumption flow now that `internal prepare`
(034) and full-pack consumption (033) already exist — several Skills
still describe the pre-034 multi-call flow in more than one section.

**Decision**: SC-001's 20% is measured on total canonical Skill body
line count (`kit/skills/*/SKILL.md`, excluding frontmatter), compared
before/after this feature's changes, checked by a simple line-count
assertion in the same test that already parses these files
structurally — not a separate tokenizer, consistent with Constitution
Principle IV (no new estimator dependency for something this coarse).
