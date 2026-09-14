# Quickstart: Ranking and Budgeting

Builds directly on 015's `Collect`. Continuing the same fixture: SPEC-014
depends on SPEC-011, links to KNOW-003, the project has a Constitution,
and a free-text query also matches an unrelated chunk in SPEC-002.

## 1. Rank a candidate set

```go
set, _ := context.Collect(root, cfg, store, context.Request{
    Target: "SPEC-014",
    Query:  "refresh token rotation",
})

ranked := context.Rank(set, context.Request{
    Target: "SPEC-014",
    Query:  "refresh token rotation",
})
```

```text
ranked's own order:
  1. Constitution chunk(s)      Tier 0 (TierMandatory)
  2. SPEC-014's own chunk(s)    Tier 0/1 (TierMandatory)
  3. SPEC-011 chunk(s)          Tier 2 (TierStructural) — depends_on
  4. KNOW-003 chunk(s)          Tier 3 (TierSemantic) — wikilink
  5. SPEC-002's own chunk       Tier 4 (TierText) — text_match, even
                                though its own textRelevance score is
                                high — it never moves above Tier 0-3.
```

## 2. Fit the ranked set into a budget

```go
budget := 500
result := context.ApplyBudget(ranked, context.Request{
    Target: "SPEC-014",
    Budget: &budget,
})
```

```text
result.Items:  every mandatory chunk (Constitution + SPEC-014), then as
               many of the remaining ranked chunks as fit in 500 tokens
               — SPEC-011 before KNOW-003 before SPEC-002, stopping the
               moment one doesn't fit.
result.BudgetExceeded: false (mandatory content alone fits in 500).
result.Diagnostics: CandidatesConsidered == len(ranked);
               TokensAvailable - TokensExcluded == TokensSelected.
```

## 3. No budget specified — the fixed default applies

```go
result := context.ApplyBudget(ranked, context.Request{Target: "SPEC-014"})
// same as passing Budget: &context.DefaultBudget-equivalent value (6000)
```

## 4. Mandatory content alone exceeds a deliberately tiny budget

```go
tiny := 1
result := context.ApplyBudget(ranked, context.Request{
    Target: "SPEC-014",
    Budget: &tiny,
})
```

```text
result.Items:          still contains every Constitution/target chunk,
                       in full.
result.BudgetExceeded: true
result.Overage:        mandatory tokens - 1, exactly.
result.Items:          contains no optional (Tier 2+) chunk at all.
```

## 5. Repeating the same call is byte-for-byte identical

```go
result2 := context.ApplyBudget(context.Rank(set, req), req)
// result2 == result, field for field, every time.
```

## Validation

Validated by: `internal/context`'s new `Rank`/`ApplyBudget` tests
covering §29.7's own Budget test list (result below budget, optional
content removed at the budget boundary, mandatory content retained,
mandatory content exceeding budget, deterministic output, token
estimates included) plus this feature's own ranking-invariant tests
(mandatory/structural never outranked by a strong text match,
same-tier ordering driven by score, intent influencing only within a
tier); 011's, 012's, 013's, 014's, and 015's own full suites re-run
unmodified, since this feature adds no call site requiring any change
to them beyond `Request`'s new, additive `Budget` field.
