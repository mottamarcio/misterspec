# Quickstart: Validating Stable Section Anchors

Prerequisites: a built `misterspec` binary from this branch, and an
indexed project (any repo with `misterspec init` already run). An
existing `.misterspec/cache/context.db` built before this feature is
fine — the first command below transparently rebuilds it (contract §7).

## 1. Reference just one section, not the whole artifact (User Story 1)

Given a Knowledge artifact `KNOW-003` with a heading declaring an
explicit anchor:

```markdown
## Client Retry Policy {#retry-policy}

Retry with exponential backoff, capped at 5 attempts.
```

And a referencing artifact containing `[[KNOW-003#retry-policy|Política
de retries]]`:

```sh
misterspec internal context KNOW-003 --provenance
```

**Expected**: `schema_version` is `5`; the item for the referenced
section carries `"anchor": "retry-policy"` and a `heading_path` array
(its ancestor headings' titles, outermost first — empty `[]` when the
anchored heading has none, still present rather than omitted) — its
`content` is that section alone, not `KNOW-003`'s other sections, even
if `KNOW-003` has many.

Confirm a non-anchor wikilink to the same artifact is unaffected:

```sh
misterspec internal context KNOW-003
```

**Expected**: response shape/content identical to before this feature
aside from `schema_version`.

## 2. Rename the heading title, keep the anchor (User Story 2)

Edit only the heading text, keeping `{#retry-policy}`:

```markdown
## Backoff and Retry Policy for Clients {#retry-policy}
```

```sh
misterspec internal validate
misterspec internal context KNOW-003 --provenance
```

**Expected**: `validate` reports no new Finding; `context` still
resolves `[[KNOW-003#retry-policy]]` to the (now differently titled)
section — the referencing wikilink required no edit.

## 3. Duplicate anchors are caught (User Story 2, Acceptance Scenario 2)

Add a second heading in the same artifact also declaring
`{#retry-policy}`:

```sh
misterspec internal validate KNOW-003
```

**Expected**: a Finding with `"code": "duplicate_anchor"` and
`"path": "ai/knowledge/KNOW-003-x.md"` (or wherever KNOW-003 actually
lives), naming the repeated anchor in its message.

## 4. A broken anchor is never silently a whole-document link (User Story 3)

Remove the `{#retry-policy}` suffix from its heading, leaving the
referencing wikilink unchanged:

```sh
misterspec internal validate
```

**Expected**: a Finding with `"code": "unknown_anchor"` on the
referencing artifact — distinct from `"code": "broken_wikilink"`, which
only ever means the target *artifact* itself doesn't exist:

```sh
misterspec internal validate # with the wikilink instead pointing at a nonexistent artifact ID + an anchor
```

**Expected**: `"code": "broken_wikilink"` this time, not
`"unknown_anchor"` — the two stay distinguishable per the target
resolution outcome (spec FR-009).

## 5. Empty-body anchor still resolves (Edge Case)

Given a heading immediately followed by a subheading (empty own Body):

```markdown
## Networking Policies {#networking}
### Client Retry Policy {#retry-policy}

Retry with exponential backoff.
```

```sh
misterspec internal context KNOW-003 --provenance
```

Reference `[[KNOW-003#networking]]` from another artifact and request
context for it.

**Expected**: resolves successfully with empty `content` for that
item, not a broken-anchor Finding — an anchor on a heading with no own
body text is still a valid, resolvable anchor.
