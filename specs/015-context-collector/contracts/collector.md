# Phase 6 Contracts: Context Collector and Retrieval

**Reconciled against the actual implementation (T018)** — zero drift
beyond the package-name correction already reflected below (`package
contextengine`, not the directory-derived `package context` —
research.md #11, caught in the very first file written during
implementation, before it could affect any other file). Every signature
(`Intent`, `Request`, `Tier`, `Reason`, `Candidate`, `CandidateSet`,
`Collect`) matches what shipped, verbatim. Unexported helpers
(`validateIntent`, `mergeAndSort`, `dedupeReasons`, `minTier`,
`candidateKey`, `chunkArtifact`, `connectedCandidates`,
`secondHopCandidates`, `textSearchLimit`) are internal decomposition
detail, not part of this feature's exported surface.

No HTTP/CLI surface is added by this feature (research.md) — its
contract is the new Go API in `internal/context`.

## `internal/context` (new files)

```go
package contextengine

// Intent names what kind of work a Request's context is for. The zero
// value ("") means no particular intent.
type Intent string

const (
    IntentPlanning       Intent = "planning"
    IntentTasks          Intent = "tasks"
    IntentImplementation Intent = "implementation"
    IntentValidation     Intent = "validation"
    IntentAnalysis       Intent = "analysis"
)

// Request is one caller's context question (FR-001). No Budget field —
// deliberately deferred to Phase 7 (research.md #2).
type Request struct {
    Target string // required
    Task   string
    Intent Intent
    Query  string
}

// Tier classifies why a Candidate was included, in ascending priority
// order (research.md #10).
type Tier int

const (
    TierMandatory Tier = iota
    TierStructural
    TierSemantic
    TierText
    TierSecondHop
)

// Reason is one specific justification for including a Candidate.
type Reason struct {
    Tier     Tier
    Relation string // "constitution" | "target" | "parent" | "depends_on" | "supersedes" | "wikilink" | "backlink" | "text_match"
}

// Candidate is one piece of collected context — identified, for
// deduplication, by (Path, StartLine, EndLine) (research.md #8).
type Candidate struct {
    Path      string
    Heading   string
    Content   string
    StartLine int
    EndLine   int
    Reasons   []Reason // never empty
}

// CandidateSet is Collect's own output: deduplicated, deterministically
// ordered, not yet ranked or budgeted (FR-013).
type CandidateSet struct {
    Candidates []Candidate
}

// Collect resolves req.Target (must be one of the five entity types
// operations.References already covers — research.md #3), gathers its
// mandatory (Tier 0/1), structural/semantic (Tier 2/3, via
// operations.References/Backlinks), text (Tier 4, via store.Search),
// and bounded second-hop (Tier 5, research.md #7) context, deduplicates
// it (FR-008), and returns it in deterministic order. Strictly
// read-only (FR-012).
func Collect(root string, cfg project.Configuration, store index.Store, req Request) (CandidateSet, error)
```

**Guarantees**:
- `Collect` never modifies any project artifact, the reference graph, or
  the search index (FR-012).
- The project's Constitution (if one exists) and the target artifact
  itself always appear in a successful result (FR-003, FR-004).
- Every `Candidate` carries at least one `Reason`; the same underlying
  chunk discovered more than once appears exactly once, with every
  distinct `Reason` it earned (FR-008, FR-009).
- Tier 0-3 are computed directly from project files — never from
  `internal/context/index` — so their correctness never depends on the
  disposable index existing (research.md #5, matching 012's/014's own
  guarantee). Only Tier 4 depends on `store`.
- Expansion never proceeds beyond one hop past a direct (Tier 2/3)
  connection (FR-010); a chunk found both directly and via second-hop
  expansion is always classified as direct (FR-011).
- `Collect` applies no ranking score and no budget — `CandidateSet` is
  the complete, deduplicated result (FR-013); Phase 7 is where scoring
  and trimming happen.

## Cross-cutting: no change to any existing contract

`internal/artifacts`, `internal/validation`, `internal/operations`,
`internal/context/index`, and `internal/cli`'s exported surfaces are
completely untouched — this feature only calls their already-existing,
already-tested functions. No new CLI command (research.md's own
"Summary of Go footprint") — Phase 8's own "internal context" command is
the later, actual CLI surface this capability feeds into.
