# Quickstart: Context Collector and Retrieval

Builds on 011 (wikilinks), 012 (reference graph), 013 (chunks), and 014
(disposable search index): SPEC-014 depends on SPEC-011 and links to
KNOW-003; the project has a Constitution.

## 1. Ask for a target with nothing else specified

```go
store, _ := index.Open(".misterspec/cache/context.db")
store.Sync(root, cfg)

set, _ := context.Collect(root, cfg, store, context.Request{Target: "SPEC-014"})
```

```text
set.Candidates includes:
  - every chunk of ai/memory/constitution.md,  Reasons: [{TierMandatory, "constitution"}]
  - every chunk of SPEC-014's own spec.md,     Reasons: [{TierMandatory, "target"}]
  - every chunk of SPEC-011 (its dependency),  Reasons: [{TierStructural, "depends_on"}]
  - every chunk of KNOW-003 (its wikilink),    Reasons: [{TierSemantic, "wikilink"}]
```

Nothing from Tier 4 (no query, no task) or Tier 5 beyond what SPEC-011/
KNOW-003 themselves reference.

## 2. Ask with a free-text question

```go
set, _ := context.Collect(root, cfg, store, context.Request{
    Target: "SPEC-014",
    Query:  "refresh token rotation",
})
```

Adds any chunk elsewhere in the project matching that text, labeled
`{TierText, "text_match"}` — even one nobody explicitly linked.

## 3. The same chunk found two ways appears once

If SPEC-011 (a direct dependency) *also* matches the free-text query:

```text
set.Candidates for SPEC-011's own chunk:
  Reasons: [{TierStructural, "depends_on"}, {TierText, "text_match"}]
```

One entry, two reasons — never two entries.

## 4. Second-hop expansion

If SPEC-011 itself depends on SPEC-009:

```text
set.Candidates includes SPEC-009's own chunks,
  Reasons: [{TierSecondHop, "depends_on"}]
```

If a project artifact three hops out exists, it never appears.

## 5. An out-of-scope or unknown target is rejected consistently

```go
_, err := context.Collect(root, cfg, store, context.Request{Target: "TASK-001"})
// errors.Is(err, operations.ErrInvalidTarget) == true
```

Same sentinel `References`/`Backlinks`/`Inspect` already use — nothing
new to learn.

## Validation

Validated by: `internal/context`'s new `Collect` tests covering
`docs/context-engine-implementation.md` §29.6's own list (target always
wins, formal/semantic connections labeled correctly, text matches
surfaced, deduplication across overlapping signals, bounded second-hop,
no third hop, out-of-scope target rejection); 011's, 012's, 013's, and
014's own full suites re-run unmodified, since this feature adds no
call site requiring any change to them.
