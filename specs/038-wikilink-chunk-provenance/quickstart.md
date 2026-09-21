# Quickstart: Validating Wikilinks with Chunk-Level Provenance

Prerequisites: a built `misterspec` binary from this branch, and an
indexed project (any repo with `misterspec init` already run). An
existing `.misterspec/cache/context.db` built before this feature is
fine — the first command below transparently rebuilds it (contract §6).

## 1. Occurrence detail on `references`/`backlinks` (additive, always on)

```sh
misterspec internal references SPEC-014
```

**Expected**: exits `0`; every entry whose `"relation": "wikilink"`
now includes non-empty `source_section`/`source_line` alongside the
existing `relation`/`target`; a formal entry (`"relation":
"depends_on"` etc.) shows `"source_section": ""`, `"source_line": 0`
(contract §3).

```sh
misterspec internal backlinks KNOW-003
```

**Expected**: same shape, from the referencing (source) side — each
entry names exactly where in the referencing artifact the link back to
`KNOW-003` was written.

## 2. Explain why an item appears in a Context Pack (User Story 1)

```sh
misterspec internal context KNOW-003 --provenance
```

**Expected**: `schema_version` is `4`; any item with real occurrence
data (`reasons` includes `"wikilink"` outgoing, or `"backlink"` when
the underlying relationship was itself a wikilink) carries a
`provenance` array naming the exact referencing artifact, section, and
line (contract §4). An item with no such occurrence data (e.g.
`"target"`, `"constitution"`, `"parent"`, `"depends_on"`, or a
`"backlink"` from a formal relation) has no `provenance` field at all.

Run without `--provenance` and confirm the response is otherwise
identical to before this feature, aside from `schema_version`:

```sh
misterspec internal context KNOW-003
```

## 3. Two distinct occurrences from the same source stay distinct (User Story 1, Acceptance Scenario 2)

Given a fixture Spec that wikilinks the same target from two different
sections:

```sh
misterspec internal context <target-id> --provenance | jq '.context.items[] | select(.reasons | index("wikilink")) | .provenance'
```

**Expected**: two separate `provenance` entries (or two entries across
items), each with its own distinct `source_section`/`source_line` —
never merged into one.

## 4. The preference policy is off by default, and explicit when enabled (User Story 2)

```sh
misterspec internal context <target-id>
misterspec internal context <target-id> --prefer-section
```

**Expected**: without `--prefer-section`, ordering matches today's
exact behavior. With it, an item whose wikilink occurrence came from a
"Requirements"/"Functional Requirements" section is scored ahead of an
otherwise-equivalent item whose occurrence came from elsewhere
(contract §5). Confirm via `--diagnostic-scores --prefer-section`
together that this is a deliberate, visible scoring difference, not an
unexplained reorder.

## 5. Cycles and hubs stay bounded (User Story 3)

```sh
misterspec internal context <id-in-a-reference-cycle>
```

**Expected**: exits `0` promptly; `diagnostics.candidates_considered`
stays within the same documented bound as before this feature (the
existing 2-hop cap — this feature does not change it, only records
provenance for what it already returns).

```sh
misterspec internal context <heavily-referenced-hub-id>
```

**Expected**: same — bounded, and no referenced content appears more
than once in `items` regardless of how many separate reference paths
lead to it.

## 6. A stale cache rebuilds transparently (contract §6)

Using an existing `.misterspec/cache/context.db` built by a
`misterspec` binary from before this feature:

```sh
misterspec internal context SPEC-014
```

**Expected**: exits `0`; no error, no manual step — the version
mismatch (`schemaVersion` 1 vs. 2) is detected and the index is
rebuilt automatically, after which `source_section`/`source_line` data
is present for every wikilink-derived entry going forward.

See `data-model.md` for the full field-level shape of
`ReferenceOccurrence` and the extended `ReferenceEntry`/`BacklinkEntry`/
`Reason` types, and `contracts/wikilink-provenance-contract.md` for the
complete request/response contract.
