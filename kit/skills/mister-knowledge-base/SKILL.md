---
name: mister-knowledge-base
description: Turn raw source documents into structured, durable Knowledge artifacts.
---

## Purpose

Convert the unprocessed material sitting under a project's raw-sources
directory into structured, addressable Knowledge artifacts — the
project's first durable memory, and the foundation every later Skill in
the pipeline eventually draws on.

## Invocation

`/mister-knowledge-base`

Takes no required argument. Operates on whatever the project's raw
sources directory currently contains.

## Responsibility

Read every raw source document not yet accounted for, extract the facts,
constraints, and open questions they contain, and record them as one
Knowledge artifact per coherent topic — never a single undifferentiated
dump, never fabricated content the sources don't actually support.

## Inputs

- The raw-sources directory's current contents (`internal inventory
  raw`'s own output).
- Any Knowledge artifacts already present, so this invocation extends
  rather than duplicates them.

## Outputs

- One Knowledge artifact per identified topic
  (`ai/knowledge/KNOW-NNN-<slug>.md`), each with the required frontmatter
  (`id`, `type`, `status`, `sources` — path plus fingerprint per source)
  and body (`Summary`, `Known Facts`, `Constraints`, `Unknowns`,
  `Conflicts`, `Provenance`, optionally `Related Topics`).
- A completion summary per the Completion Contract below.

## Preconditions

The project must already be initialized (`misterspec init` has already
run). At least one file must exist under the raw-sources directory —
otherwise there is nothing to process (see Edge Cases in Failure
Conditions).

## Required Context

The raw-sources directory's inventory and the content of every raw
source file not yet fingerprinted into an existing Knowledge artifact's
`sources` list.

## Optional Context

Any existing Knowledge artifacts, to detect overlap with a topic already
covered and extend it instead of creating a duplicate.

## Conditional Context

If a source document is a PDF or other non-Markdown format, its content
must be extracted through whatever capability is available before
semantic analysis — this Skill does not mandate a specific extraction
tool.

## Unnecessary Context

The project's Constitution, Programs, Features, or Specs — none exist
yet at this pipeline stage, and none are needed to extract facts from
raw sources.

## Authority

This Skill reasons about raw source content and decides topic
boundaries, fact extraction, and conflict identification. It does not
decide project scope, architecture, or priorities — those belong to
later Skills.

## Allowed Reads

Every file under the raw-sources directory; every existing Knowledge
artifact.

## Allowed Creates

Knowledge artifacts only, via `internal create knowledge`.

## Allowed Modifications

An existing Knowledge artifact's own body, when this invocation extends
a topic it already covers with newly found facts — never a Program,
Feature, Spec, or any other artifact type.

## Forbidden Mutations

Creating or modifying any artifact type other than Knowledge. Editing a
raw source file. Fabricating a source's fingerprint instead of computing
it via `internal fingerprint`.

## Deterministic Operations

Use misterspec operations for every mechanical repository step —
never hand-authored IDs or guessed paths.

Required operations:

- `internal inventory raw` — discover what raw sources exist.
- `internal fingerprint <source>` — compute each source's content digest
  before recording it in a Knowledge artifact's `sources` field.
- `internal create knowledge --slug <slug>` — allocate the next
  Knowledge ID and scaffold its canonical file; never construct a
  `KNOW-NNN` identifier or its file path by hand.
- `internal inspect <knowledge-id>` — confirm a newly created (or
  extended) Knowledge artifact's metadata is well-formed before
  reporting completion.
- `internal validate` — confirm the project's overall structure is still
  valid after this invocation's writes.

## Procedure

1. Run `internal inventory raw` to discover every raw source file.
2. For each source not yet fingerprinted into an existing Knowledge
   artifact's `sources` list, run `internal fingerprint <source>` and
   read its content.
3. Group related facts into coherent topics — each topic becomes one
   Knowledge artifact, or extends an existing one that already covers
   the same topic.
4. For each new topic, run `internal create knowledge --slug <slug>`,
   then write its body: `Summary`, `Known Facts`, `Constraints`,
   `Unknowns`, `Conflicts`, `Provenance` (naming which source each fact
   came from), and `Related Topics` where useful.
5. Run `internal inspect <knowledge-id>` for every artifact touched, to
   confirm its frontmatter is well-formed.
6. Run `internal validate` once, at the end, across the whole project.
7. Report completion per the Completion Contract below.

## Decision Rules

- A fact belongs in exactly one Knowledge artifact's `Known Facts`;
  a fact relevant to two topics is recorded once and cross-referenced
  via `Related Topics`, not duplicated.
- A statement the sources leave ambiguous or contradictory belongs in
  `Unknowns` or `Conflicts`, never silently resolved by guessing.
- A source already fully accounted for in an existing Knowledge
  artifact's `sources` list is not reprocessed.

## Interaction Rules

If raw sources are numerous or ambiguous, prefer creating more, narrower
Knowledge artifacts over one broad one — narrow topics keep later Skills
(especially `/mister-constitution`) able to cite specific Knowledge
precisely.

## Validation Rules

Every Knowledge artifact this Skill creates or modifies must pass
`internal validate` with zero new structural findings attributable to
it. Every `sources` entry must name a real, fingerprinted raw file.

## Failure Conditions

- The raw-sources directory is empty or does not exist: report that
  nothing was found to process — never fabricate Knowledge content.
- `internal create knowledge` reports `invalid_argument` for an empty or
  unsafe slug: choose a different, filesystem-safe slug and retry.
- The project is not initialized (`project_not_initialized`): stop and
  report that `misterspec init` must run first.

## Stop Conditions

Stop once every raw source has been accounted for in some Knowledge
artifact's `sources` list, or once no further meaningful topic
boundaries remain to extract.

## Success Criteria

Every raw source is referenced by at least one Knowledge artifact's
`sources` list; every created or modified Knowledge artifact passes
`internal validate` cleanly; no fact is fabricated beyond what the
sources actually state.

## Postconditions

The project's Knowledge base reflects every currently available raw
source. `/mister-constitution` can now draw on it.

## Idempotency

Re-running this Skill after new raw sources are added only processes
the new sources — already-fingerprinted sources are not reprocessed,
and existing Knowledge artifacts are extended, not duplicated.

## Resume Behavior

If interrupted partway through, re-running the Skill picks up exactly
where it left off: sources already fingerprinted into an existing
artifact's `sources` list are skipped; unfingerprinted sources are
processed again.

## Completion Contract

Every invocation ends with a concise operational summary naming:

Render this summary using structured formatting, not prose paragraphs: present **Artifacts** as a Markdown table when more than one artifact is involved (columns matching what's relevant — ID, path/type, and status or a one-line summary), or a single bullet when there is exactly one; present **Important findings** and **Attention** as bullet lists. This applies equally to a failure/stop report.

- **Outcome** — succeeded, partially succeeded, or found nothing to do.
- **Artifacts** — every Knowledge artifact created or extended, by ID.
- **Important findings** — notable facts or constraints surfaced.
- **Attention** — unresolved conflicts or unknowns preserved in the
  Knowledge artifacts themselves, named here so they aren't missed.
- **Recommended next step** — naming the exact next Skill invocation.

## Recommended Next Step

Once at least one Knowledge artifact exists, recommend:

```text
/mister-constitution
```

If no raw sources were found, recommend populating the raw-sources
directory first, and do not recommend `/mister-constitution` yet — a
Constitution distilled from nothing is not durable memory.

## Related Skills

`/mister-constitution` — the next Skill in the pipeline, which draws
directly on the Knowledge base this Skill produces.
