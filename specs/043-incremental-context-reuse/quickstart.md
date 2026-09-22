# Quickstart: Validating Incremental Context Pack Reuse

Prerequisites: a built `misterspec` binary from this branch, and an
initialized project with a Spec that has selectable content (at least
one Requirement).

## 1. First call returns a `pack_id` to reuse later (Foundational)

```sh
misterspec internal context SPEC-014 --mode package
```

**Expected**: `ok: true`; `context.pack_id` is present, a non-empty
`"sha256:..."` string; `context.items` is the full package, same shape
as before this feature existed.

## 2. Repeat call with no source change reuses everything (User Story 1)

```sh
PACK=$(misterspec internal context SPEC-014 --mode package | jq -r .context.pack_id)
misterspec internal context SPEC-014 --mode package --base "$PACK"
```

**Expected**: `context.diff.entries` contains only `"reuse"` entries —
`context.reuse.items_sent == 0`; `context.diff.base_pack_id` equals
`$PACK`; no `content` field appears anywhere in `diff.entries`.

## 3. A real content change appears as exactly one diff entry (User Story 1)

Edit one Requirement's text in `spec.md`, then:

```sh
misterspec internal context SPEC-014 --mode package --base "$PACK"
```

**Expected**: `context.diff.entries` contains exactly one `"modified"`
entry (for the changed Requirement) plus `"reuse"` entries for
everything else — `context.reuse.items_sent == 1`.

## 4. An unknown base recovers safely, never a partial result (User Story 2)

```sh
misterspec internal context SPEC-014 --mode package --base "sha256:0000000000000000000000000000000000000000000000000000000000000000"
```

**Expected**: `context.recovered == true`; `context.reason ==
"unknown_base"`; `context.items` is the full package — not `diff`.

## 5. A config change invalidates the base even if content didn't move (User Story 3)

```sh
misterspec internal context SPEC-014 --mode package --base "$PACK" --budget 100
```

**Expected**: `context.recovered == true`; `context.reason ==
"invalidated"` (the resolved `Budget` component of the pack's own
identity differs from `$PACK`'s).

## 6. Diff always reconstructs the exact full pack (User Story 1, FR-003)

After step 3, fetch a fresh full pack for the same target and diff the
two JSON payloads programmatically:

```sh
misterspec internal context SPEC-014 --mode package > /tmp/full-current.json
```

**Expected**: reconstructing the full item list from step 3's `diff`
(substituting each `"reuse"` entry with the corresponding item from
the `$PACK` response saved in step 1/2, in the diff's own listed
order) is byte-identical, item by item, to `/tmp/full-current.json`'s
own `context.items`.

## Related commands

- `misterspec internal context --mode package` (033) — the underlying
  full-content mode this feature diffs against; unaffected when
  `--base` is omitted.
- `misterspec internal analyze-impact` (042) — answers a related but
  different question ("what does a change affect"); this feature
  answers "how much of what I already fetched can I skip resending."
