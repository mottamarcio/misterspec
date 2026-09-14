package contextengine

import "sort"

// Tier classifies why a Candidate was included, in ascending priority
// order (docs/context-engine-implementation.md §14, research.md #10) —
// the CandidateSet's own primary sort key.
type Tier int

const (
	// TierMandatory is the Constitution and the target itself.
	TierMandatory Tier = iota
	// TierStructural is a formal relationship (parent/depends_on/
	// supersedes) directly on the target.
	TierStructural
	// TierSemantic is a direct wikilink or an incoming backlink,
	// directly on the target.
	TierSemantic
	// TierText is a free-text match from the local search index.
	TierText
	// TierSecondHop is an outgoing relationship one further hop from a
	// Tier 2/3 artifact.
	TierSecondHop
)

// Reason is one specific justification for including a Candidate.
// Relation is one of "constitution", "target", "parent", "depends_on",
// "supersedes", "wikilink", "backlink", or "text_match" (data-model.md).
type Reason struct {
	Tier     Tier
	Relation string
}

// Candidate is one piece of collected context. Identified, for
// deduplication, by (Path, StartLine, EndLine) — the same identity
// artifacts.Chunk already carries (research.md #8).
type Candidate struct {
	Path      string
	Heading   string
	Content   string
	StartLine int
	EndLine   int
	// Reasons is never empty (FR-009).
	Reasons []Reason
}

// CandidateSet is Collect's own output: deduplicated, deterministically
// ordered (research.md #10), not yet ranked or budgeted (FR-013).
type CandidateSet struct {
	Candidates []Candidate
}

// candidateKey is a Candidate's own deduplication identity.
type candidateKey struct {
	path      string
	startLine int
	endLine   int
}

// mergeAndSort merges candidates sharing the same (Path, StartLine,
// EndLine) into one Candidate carrying every distinct Reason (FR-008),
// drops a TierSecondHop reason wherever a direct (non-second-hop)
// reason already justifies the same chunk (FR-011, research.md #8), and
// returns the result ordered by each candidate's own best (lowest)
// Tier, then Path, then StartLine (research.md #10) — a total,
// deterministic order requiring no separate scoring step.
func mergeAndSort(candidates []Candidate) CandidateSet {
	merged := map[candidateKey]*Candidate{}
	var order []candidateKey

	for _, c := range candidates {
		key := candidateKey{c.Path, c.StartLine, c.EndLine}
		if existing, ok := merged[key]; ok {
			existing.Reasons = append(existing.Reasons, c.Reasons...)
			continue
		}
		cp := c
		cp.Reasons = append([]Reason(nil), c.Reasons...)
		merged[key] = &cp
		order = append(order, key)
	}

	result := make([]Candidate, 0, len(order))
	for _, key := range order {
		c := merged[key]
		c.Reasons = dedupeReasons(c.Reasons)
		result = append(result, *c)
	}

	sort.Slice(result, func(i, j int) bool {
		ti, tj := minTier(result[i].Reasons), minTier(result[j].Reasons)
		if ti != tj {
			return ti < tj
		}
		if result[i].Path != result[j].Path {
			return result[i].Path < result[j].Path
		}
		return result[i].StartLine < result[j].StartLine
	})

	return CandidateSet{Candidates: result}
}

// dedupeReasons removes exact duplicate Reasons, and — when any
// non-second-hop Reason is present — drops every TierSecondHop Reason
// for the same Candidate (FR-011): a chunk found both directly and via
// second-hop expansion is always classified as direct.
func dedupeReasons(reasons []Reason) []Reason {
	hasDirect := false
	for _, r := range reasons {
		if r.Tier != TierSecondHop {
			hasDirect = true
			break
		}
	}

	seen := map[Reason]bool{}
	out := make([]Reason, 0, len(reasons))
	for _, r := range reasons {
		if hasDirect && r.Tier == TierSecondHop {
			continue
		}
		if seen[r] {
			continue
		}
		seen[r] = true
		out = append(out, r)
	}
	return out
}

// minTier returns the lowest (highest-priority) Tier among reasons —
// a Candidate's own effective Tier for sorting purposes, since a merged
// Candidate may carry Reasons from more than one Tier.
func minTier(reasons []Reason) Tier {
	min := reasons[0].Tier
	for _, r := range reasons[1:] {
		if r.Tier < min {
			min = r.Tier
		}
	}
	return min
}
