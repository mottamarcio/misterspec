# Quickstart: Validating Impact Analysis & Incremental Review

Prerequisites: a built `misterspec` binary from this branch, inside a
Git repository (`misterspec internal analyze-impact` requires Git
history to diff), with a Spec that has at least one Requirement, one
Task in `tasks.md` covering it via `Serves:`, and that Task's
`Evidence-Fingerprint:` already captured (041-task-evidence-
fingerprint) against a known commit.

## 1. Requirement text changes → covering Task is reported invalidated (User Story 1, 2)

Note the current commit, then edit the text under one `### R<N>` heading
in `spec.md` (not just its title — real body content) and leave the
change uncommitted:

```sh
BASE=$(git rev-parse HEAD)
# ...edit spec.md's R2 body text...
misterspec internal analyze-impact --from "$BASE"
```

**Expected**: `ok: true`; `change_set.elements` includes one entry for
the Spec with `"requirement_number": 2`; `affected_items` includes the
Task that serves `SPEC-###:R2`, with
`"classification": "deterministic_invalidation"` and a `reason`
naming the changed Requirement — not a generic message.

## 2. A wikilink mention alone never becomes an invalidation (User Story 2)

Pick an artifact `A` that only mentions a second artifact `B` via
`[[B]]` (no `depends_on`, no coverage, no evidence relationship).
Commit that baseline, then edit `B`'s content and leave it uncommitted:

```sh
BASE=$(git rev-parse HEAD)
# ...edit B's content...
misterspec internal analyze-impact --from "$BASE"
```

**Expected**: `A` appears in `affected_items` with
`"classification": "suggested_review"`, never
`"deterministic_invalidation"` — confirm by grepping the JSON response
for `A`'s ID and checking its own `classification` field, not just
scanning for its presence.

## 3. A real dependency chain reports its full path (User Story 3)

With Spec `X` declaring `depends_on: [Y]` in frontmatter, and a Task in
`X` whose `Serves:`/evidence otherwise ties it to `Y`'s content (or
simply confirm the one-hop case first): change `Y`, then run
`analyze-impact` scoped from before that change.

**Expected**: the `AffectedItem` for `X` (or the deeper Task) carries a
non-empty `paths[0].hops` array whose first hop's `from` is `Y`'s ID
and whose last hop's `to` is the affected item's own ID — the full
chain, not only the nearest hop.

## 4. No known relation is stated explicitly, never implied

Change an artifact with genuinely no incoming formal/semantic/coverage
relation (e.g. a freshly created, unreferenced Knowledge note):

```sh
BASE=$(git rev-parse HEAD)
# ...edit the new Knowledge note...
misterspec internal analyze-impact --from "$BASE"
```

**Expected**: `affected_items` does not mention it, **and**
`no_known_relation_elements` explicitly includes its ID — confirming
the report states "nothing known" rather than merely omitting it.

## 5. Termination on a reference cycle

In a fixture with two artifacts that wikilink each other (`A ↔ B`),
change one of them and run `analyze-impact` against it.

**Expected**: the command returns promptly (`ok: true`) with a
bounded, non-repeating set of `affected_items` — `B` is not expanded
more than once even though the cycle offers more than one path back to
it.

## 6. Unmapped code changes are declared, not hidden

Change a `.go` source file alongside a Spec edit in the same diff:

```sh
misterspec internal analyze-impact --from "$BASE"
```

**Expected**: `change_set.unmapped_code_paths` is ≥ 1 — the code
change is counted as an explicit gap, never silently dropped from the
response nor misreported as "no impact from this code change."

## Related commands

- `misterspec internal capture-evidence` (041) — the operation named
  by a reported `reverification_candidate`, to re-verify an
  invalidated Task against the project's current state.
- `misterspec internal validate` (032/041) — still the source of
  static coverage/evidence Findings; `analyze-impact` is a change-
  triggered *report*, not a replacement for `validate`'s own
  point-in-time checks.
