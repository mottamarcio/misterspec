# Feature Specification: Wikilinks with Chunk-Level Provenance

**Feature Branch**: `038-wikilink-chunk-provenance`
**Created**: 2026-09-21
**Status**: Draft
**Input**: User description: "PROP-08 — Wikilinks com proveniência por trecho: evoluir o suporte existente a wikilinks para recuperar contexto com base no lugar e na finalidade da referência." (Backlog proposal from `tmp/misterspec-specs-sugeridas.md`: preserve the exact place a wikilink reference was written from — source artifact, location, and enclosing section — through indexing and retrieval, so a retrieved item can be explained by the specific reference occurrence that justified it, and so retrieval can eventually prefer references made where they matter most, evaluated through the project's own evidence-based promotion process before becoming default behavior.)

## User Scenarios & Testing *(mandatory)*

<!--
  Wikilinks already let one artifact point to another
  (011-wikilink-foundation) and already feed structural/semantic
  context retrieval (012-references-backlinks, 015-context-collector).
  What is still missing is provenance: today a retrieved item can only
  say "some artifact references this one," never "this specific Spec
  referenced it from its own Requirements section, on this line." This
  feature's users are maintainers and coding agents who need to trust
  and act on *why* something showed up in their context, and, later,
  to have retrieval favor references made in the most relevant part of
  a document over incidental mentions — evaluated with evidence
  (037-eval-quality-efficiency) before that preference ever changes
  default behavior, the same discipline already applied to ranking
  changes (036-text-search-ranking).
-->

### User Story 1 - Explain Exactly Why a Referenced Artifact Appeared (Priority: P1)

As a maintainer or coding agent reviewing a Context Pack, when an item
was included because another artifact references it, I want to see
exactly which artifact made that reference, from which section of that
artifact, and at which location — not just "referenced by SPEC-014
somewhere" — so I can judge whether the reference is actually relevant
to my current work without opening the source file myself.

**Why this priority**: This is the foundational capability every other
part of this feature depends on — without recording where a reference
was written from, there is nothing to prefer, explain, or evaluate.
It is also immediately useful on its own: today's retrieval already
includes referenced content, it just cannot explain it precisely.

**Independent Test**: Can be fully tested by requesting a Context Pack
for a target that another artifact references from a specific,
identifiable section, and confirming the returned explanation names
that exact referencing artifact, section, and location — not merely
the fact that a reference exists.

**Acceptance Scenarios**:

1. **Given** an artifact that references another artifact's ID from
   its own "Related Specs" section, **When** a Context Pack is
   requested for the referenced artifact, **Then** the returned
   explanation names the referencing artifact, identifies the "Related
   Specs" section as the reference's origin, and points to the
   reference's specific location within that artifact.
2. **Given** an artifact that references the same target from two
   different sections, **When** a Context Pack is requested for that
   target, **Then** both reference occurrences are distinguishable
   from each other, not collapsed into a single, generic "referenced"
   signal.
3. **Given** a reference written as `[[ID|Display Text]]`, **When**
   its provenance is reported, **Then** the display alias never
   changes which artifact or location the reference is attributed to.

---

### User Story 2 - Prefer References Made Where They Matter (Priority: P2)

As a maintainer relying on retrieval to prioritize the most useful
context first, I want references written inside an artifact's
Requirements content, or inside the section directly covering my
current task, to be preferred over references made incidentally
elsewhere in the same artifact — but only once this preference has
been evaluated and shown to actually help, never adopted just because
it seems reasonable.

**Why this priority**: This is the actual retrieval-quality payoff of
recording provenance (User Story 1), but it is explicitly a policy
change to ranking/retrieval behavior, which this project's own
established discipline requires evidence for before promoting to
default — so it depends on User Story 1 already existing, and it
cannot ship as an unconditional default on its own.

**Independent Test**: Can be fully tested by requesting a Context Pack
under this preference policy for a target referenced both from a
Requirements section and from an unrelated section elsewhere, and
confirming the Requirements-section reference is preferred, while
confirming this policy is never applied unless explicitly evaluated
and enabled through the project's evidence-based promotion process.

**Acceptance Scenarios**:

1. **Given** a target referenced once from a Requirements section and
   once from an unrelated section of the same artifact, **When** this
   preference policy is enabled, **Then** the Requirements-section
   reference is treated as more relevant than the other.
2. **Given** the same setup, **When** this preference policy has not
   been evaluated and promoted, **Then** retrieval behavior is
   unchanged from today's — both references are treated equivalently.
3. **Given** a candidate version of this preference policy, **When** a
   maintainer wants to adopt it as default, **Then** they can only do
   so after recording a comparison against prior behavior through the
   project's own evaluation harness.

---

### User Story 3 - Retrieval Stays Bounded and Non-Redundant (Priority: P3)

As a maintainer working in a project where artifacts reference each
other heavily, or even reference each other in a cycle, I want
retrieval to never load the same content twice for one request and
never expand without limit chasing references, so that requesting
context for any artifact remains fast and predictable regardless of
how densely cross-referenced the project becomes.

**Why this priority**: This is a safety/scalability guarantee rather
than new user-facing value — it protects User Stories 1 and 2 from
becoming a liability as a project's own reference graph grows, and can
be verified independently of whether preference-based retrieval
(User Story 2) is ever enabled.

**Independent Test**: Can be fully tested by requesting a Context Pack
for a target inside a reference cycle (A references B, B references
A) and for a target referenced from many other artifacts, and
confirming both requests complete, return bounded results, and never
include the same referenced content more than once.

**Acceptance Scenarios**:

1. **Given** two artifacts that reference each other, **When** a
   Context Pack is requested for either one, **Then** the request
   completes without following the cycle indefinitely.
2. **Given** a target referenced from many other artifacts, **When** a
   Context Pack is requested for it, **Then** the number of
   reference-driven items considered stays within a documented,
   predictable bound.
3. **Given** two separate reference paths that both lead to the same
   underlying content, **When** a Context Pack is assembled, **Then**
   that content appears only once, not once per path.

---

### Edge Cases

- A wikilink whose target no longer exists (a broken reference) — its
  provenance is simply not usable for retrieval; this is not a new
  failure mode beyond how broken references are already handled today.
- The section a reference was written under is later renamed or the
  content around it is restructured — the reference's recorded
  provenance reflects the section identity at the time it was last
  recorded, and is refreshed the next time the source artifact is
  reprocessed; it is not guaranteed to survive a rename until a
  reference is re-recorded.
- An artifact with no wikilinks at all — provenance simply has nothing
  to report for it; not an error condition.
- A reference recorded from a source artifact that is later deleted or
  its content substantially changed — the previously recorded
  provenance must not be presented as still valid without being
  refreshed against the artifact's current state.
- A single line containing more than one reference — each occurrence
  is tracked as its own, separately identifiable reference.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: For every wikilink-based reference used to justify
  including content in a Context Pack, the system MUST record the
  specific artifact it was written in, the location within that
  artifact, and the section it falls under.
- **FR-002**: A Context Pack item included because of a reference MUST
  be traceable back to the exact reference occurrence (source
  artifact, section, and location) that justified its inclusion —
  never only "some reference to this exists somewhere."
- **FR-003**: The system MUST continue supporting both a plain
  reference (`[[ID]]`) and an aliased reference (`[[ID|Display
  Text]]`) with identical resolution behavior; the display alias MUST
  remain presentation-only and MUST NOT affect which artifact or
  location a reference is attributed to.
- **FR-004**: When the same target is referenced from more than one
  location — including more than once within the same source artifact
  — each occurrence MUST remain separately identifiable rather than
  being collapsed into one generic signal.
- **FR-005**: The system MUST support a preference for references
  written within a Spec's own Requirements content, or within the
  section directly covering the currently active task, over references
  written elsewhere in the same artifact.
- **FR-006**: The preference described in FR-005 MUST NOT be applied
  as default retrieval behavior unless it has first been evaluated and
  compared against prior behavior through the project's own
  evidence-based evaluation process; until then, all reference
  occurrences MUST be treated equivalently regardless of the section
  they were written under.
- **FR-007**: The system MUST NOT load the same referenced content
  more than once within a single Context Pack request, regardless of
  how many separate reference paths lead to it.
- **FR-008**: The system MUST bound how many additional hops of
  references are followed from the original requested artifact, so
  that a reference cycle, or an artifact referenced by unusually many
  others, can never cause unbounded retrieval.
- **FR-009**: A reference recorded for retrieval/provenance purposes
  MUST NOT be interpreted, stored, or presented as a formal execution
  dependency — only an artifact's own explicitly declared dependency
  relationship establishes that, unchanged by this feature.
- **FR-010**: All stored reference/provenance information MUST be
  fully reconstructable from the project's current Markdown files at
  any time, and MUST be automatically rebuilt — without a manual
  migration step — whenever the information this feature records about
  a reference changes shape.
- **FR-011**: Reference/provenance information for a given artifact
  MUST be refreshed whenever that artifact's own content changes, so a
  reference is never retrieved or explained using stale information
  about where or how it was written.
- **FR-012**: A broken reference (one whose target does not resolve to
  a real artifact) MUST NOT cause a Context Pack request to fail; it
  simply contributes no usable provenance.

### Key Entities

- **Reference Occurrence**: one specific, identifiable place in an
  artifact where a wikilink to another artifact appears — captures
  which artifact it was written in, where within that artifact, which
  section it falls under, and which artifact it points to.
- **Context Pack Item Provenance**: the explanation attached to a
  retrieved Context Pack item, naming the specific Reference
  Occurrence(s) that justified including it.
- **Reference Index**: the complete, disposable, rebuildable record of
  every known Reference Occurrence across the project, used to answer
  both "what does this artifact reference, and from where" and "what
  references this artifact, and from where."

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For 100% of Context Pack items included because of a
  reference, a maintainer can identify the exact originating artifact,
  section, and location directly from the response, without opening
  the source file themselves.
- **SC-002**: 0 instances of unbounded or runaway retrieval occur when
  requesting context within a project structure containing a reference
  cycle or an artifact with an unusually high number of incoming
  references.
- **SC-003**: 0 changes to the reference-preference policy (User
  Story 2) are adopted as default behavior without a recorded,
  evidence-based comparison against prior behavior.
- **SC-004**: 100% of stored reference/provenance information can be
  regenerated from the project's current files with no information
  loss, verified by comparing the regenerated result against the
  original.
- **SC-005**: For a target referenced from more than one location,
  100% of those distinct reference occurrences remain individually
  identifiable in the response, rather than being merged into a single
  aggregate signal.

## Assumptions

- "Section" identity for a recorded reference is derived from the
  nearest enclosing heading at the time the reference is last
  processed, consistent with how the project already segments Markdown
  content for retrieval; making section identity resilient to renames
  (e.g. through stable, explicit anchors) is a separate, later
  proposal and out of scope here.
- The reference-preference policy in User Story 2 is an evaluable
  retrieval policy shipped behind the same evidence-based promotion
  discipline already used for ranking changes (036-text-search-
  ranking) — this feature does not itself change default retrieval
  behavior on release.
- This feature preserves the existing, already-established distinction
  between a formal dependency (`depends_on`) and a semantic wikilink
  reference; nothing here changes what counts as a dependency for any
  other purpose (validation, task readiness, etc.).
- The specific numeric bound on reference-hop expansion (FR-008) is a
  planning-level decision; this specification requires only that a
  documented, predictable bound exists, not a particular number.
- This feature depends on the Context Pack contract already
  established (033-context-pack-output-contract) as the surface
  through which provenance is reported to a caller.
