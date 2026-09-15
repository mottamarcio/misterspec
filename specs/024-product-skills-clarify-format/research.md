# Research: Interactive Clarification and Richer Output for Product Skills

## 1. Interactive clarification mechanics for `create-specs`/`create-plan`

**Decision**: Reuse, near-verbatim, the same question-quality and
interaction mechanics already validated for this project's own dev-tooling
in spec 023 (`speckit-clarify`): a full-sentence question ending in `?`;
never a bare label/ID; 2-4 concrete options rendered as a Markdown table
with one row marked "(Recommended)" plus a one-sentence reason; a bounded
number of questions per invocation (reusing the existing prioritization
order already documented in each Skill's own "For unclear aspects" default-
handling guidance: scope > security/privacy > user experience > technical
detail); disambiguation-on-mismatch rather than guessing.

**Rationale**: This exact mechanic is already proven inside this same
repository (spec 023, merged to `dev`) — reusing it avoids inventing a
second, subtly different interaction pattern for the product's own Skills,
and directly satisfies the user's own explicit ask ("dando uma lista de
opções... adicionando 'recomendado'").

**Alternatives considered**: A lighter-weight "just ask one open-ended
question" approach instead of structured options: rejected — the user
explicitly asked for a list of options with a recommended pick, and
unstructured questions are harder for a human to answer quickly than a
short table.

## 2. Where "genuine ambiguity" is judged (no new deterministic operation)

**Decision**: The judgment of *whether* a requirement/strategy is
"genuinely ambiguous" (materially different implications) versus "has an
obvious default" stays exactly where it already is — inside the Skill's own
`Interaction Rules`/default-handling prose, evaluated by the agent. This
feature does not add a new `internal` command or heuristic to detect
ambiguity mechanically.

**Rationale**: Constitution Principle I — this is semantic judgment, not a
mechanical computation the Go binary could reliably perform. The existing
`create-specs`/`create-plan` prose already names the judgment call
("ambiguous," "more than one reasonable option with materially different
tradeoffs"); this feature only changes what happens *after* that judgment
(ask vs. silently resolve), not how the judgment itself is made.

**Alternatives considered**: None seriously — introducing a deterministic
"ambiguity score" would be exactly the kind of semantic/deterministic
boundary violation Principle I forbids.

## 3. Recording deferred/unasked questions

**Decision**: Reuse the Spec artifact's own existing `## Unresolved
Questions` section (confirmed present in `kit/templates/spec.md.tmpl`) for
any ambiguity that goes unasked because the per-invocation question quota
was reached. No equivalent section exists on the Plan artifact template
(verified: `create-plan`'s own Interaction Rules already route an unaskable
planning fork into `Risks`/`Assumptions`, which the Plan template already
has) — User Story 2 reuses that existing home unchanged for the same
"deferred" case, applying the new "ask when genuinely ambiguous" behavior
only to what actually gets asked.

**Rationale**: Corrects an inaccuracy carried over from the prior
comparison report (it claimed only the Feature artifact has an
open-questions field) — verified directly against the actual `.tmpl` files
before designing this decision, per this project's own "verify against
current state before recommending" discipline. No template change is
needed.

**Alternatives considered**: Adding a new, separate "Deferred
Clarifications" section: rejected — would duplicate a field that already
exists and already serves this exact purpose.

## 4. Constraint-verbatim rule for `create-tasks`

**Decision**: Add one Decision Rule to `create-tasks/SKILL.md`: when a
Requirement's own text states an explicit constraint (a specific limit,
required format, or measurable threshold), the Task description quotes
that constraint's exact wording, in addition to the existing `SPEC-###:R#`
reference — not instead of it.

**Rationale**: Direct application of the same fix already validated
upstream (spec-kit's own `tasks.md` constraint-verbatim rule) and already
adopted for this repo's dev-tooling in spec 023 — the same drift risk
exists here (a Task's own author paraphrasing a constraint differently from
how the Spec worded it), and the fix is a one-sentence rule addition with
no new mechanism.

**Alternatives considered**: Requiring `internal inspect` to expose
constraints as a separate, structured field the agent must copy from:
rejected — over-engineering for a one-line instruction fix; the agent
already reads the full Requirement text via `internal inspect SPEC-###` in
`create-tasks`'s own existing Procedure step 1, so the constraint text is
already in context — it just needs to actually be quoted.

## 5. `analyze`'s narrow, confirmed remediation-Task capability

**Decision**: Add a new step to `analyze/SKILL.md`'s own Procedure,
immediately after a `Result: fail` is recorded whose responsible layer
(per its own existing §53 branching rule) is "implementation incomplete":
present the gap and ask the user whether to append a tracking Task for it,
using the same recommended-option interaction pattern as decision #1
(e.g. "(Recommended) Yes — track this so `/implement` picks it up next" vs.
"No — leave it as a reported finding only"). On explicit "yes," append
exactly one `## TASK-NNN` entry directly to the existing Tasks artifact
(scanning existing `TASK-NNN` headings for the next number, the same
convention `create-tasks` itself already uses — confirmed in its own
`Allowed Modifications`), naming the specific failing requirement and the
evidence `analyze` already recorded for it. Widen `analyze/SKILL.md`'s own
`Allowed Modifications` accordingly (from "The Validation artifact's own
content, on a re-run" to also include this one narrow, gated Task-append),
and add a matching line to its `Forbidden Mutations` making explicit that
this is the *only* new permitted write — the Spec, Plan, and every
pre-existing Task remain forbidden to touch, unchanged from today.

**Rationale**: This is the single most constitution-sensitive change in
this feature (Principle VII exists precisely to keep a Skill's own mutation
boundary narrow), so every constraint from spec.md's own User Story 4 is
carried through literally: gap-type-scoped (only "implementation
incomplete"), user-confirmed (never automatic), and append-only (never
touching existing content). This mirrors the exact same append-only
contract already validated for the dev-tooling `speckit-converge` skill in
spec 023, applied to a narrower single-Task case rather than a whole new
"Convergence phase."

**Alternatives considered**:
- Making `analyze` always append the Task automatically (no confirmation):
  rejected outright — the user explicitly asked for confirmation with
  options, and an automatic write here would be a materially larger
  widening of `analyze`'s own mutation boundary than the narrow, gated
  version.
- Extending this capability to Spec/Plan-layer findings too (e.g. letting
  `analyze` also flag "the Spec itself needs rewriting" as an append-able
  Task): rejected — spec.md's own Edge Cases explicitly scope this to
  "implementation incomplete" only; a Spec/Plan-layer gap still routes back
  to the responsible Skill in prose, exactly as `analyze` already does
  today.

## 6. Richer completion-summary formatting across all 9 Skills

**Decision**: Every `kit/skills/*/SKILL.md` file's `Completion Contract`
section already opens with the identical sentence "Every invocation ends
with a concise operational summary naming:" (confirmed byte-identical
across all 9 files). Insert one new instruction immediately after that
sentence, in all 9 files: *"Render this summary using structured
formatting, not prose paragraphs: present **Artifacts** as a Markdown
table when more than one artifact is involved (columns matching what's
relevant — ID, path/type, and status or a one-line summary), or a single
bullet when there is exactly one; present **Important findings** and
**Attention** as bullet lists. This applies equally to a failure/stop
report."*

**Rationale**: Every Skill's own `Completion Contract` already names five
distinct informational buckets — the gap is purely that nothing instructs
the agent to *render* that structure at runtime rather than flattening it
into prose. A single, identical instruction at an already-identical anchor
point across all 9 files is the smallest possible fix, requiring no
per-Skill customization.

**Alternatives considered**: Redesigning each Skill's own `Completion
Contract` bullets individually to bake in formatting per-field: rejected —
unnecessary duplication of effort for a rule that applies identically to
every Skill; a single added sentence at the shared anchor point achieves
the same outcome with far less text churn (and lower risk of the 9 files
drifting from each other over time).
