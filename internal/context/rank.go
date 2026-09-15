package contextengine

import (
	"sort"
	"strings"
)

// ScoredCandidate is one Candidate carrying a deterministic relevance
// Score, meaningful only among candidates already known to share the
// same Tier (FR-001, FR-002, FR-003, 016-ranking-budgeting/
// research.md #3) — Score is never compared, or summed with anything,
// across a Tier boundary.
type ScoredCandidate struct {
	Candidate
	Score int
}

// Rank assigns every Candidate in cs a Score and returns them ordered
// by ascending Tier first — absolutely, never crossed by Score
// (§17.1's "important ranking invariant," FR-002, FR-003) — then by
// descending Score, then Path, then StartLine, matching 015's own
// mergeAndSort tie-break convention (research.md #3). Deterministic
// for the same inputs (FR-011). Pure — no I/O, no mutation (FR-012).
func Rank(cs CandidateSet, req Request) []ScoredCandidate {
	terms := queryTerms(req)

	scored := make([]ScoredCandidate, len(cs.Candidates))
	for i, c := range cs.Candidates {
		scored[i] = ScoredCandidate{Candidate: c, Score: scoreCandidate(c, req.Intent, terms)}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		ti, tj := minTier(scored[i].Reasons), minTier(scored[j].Reasons)
		if ti != tj {
			return ti < tj
		}
		if scored[i].Score != scored[j].Score {
			return scored[i].Score > scored[j].Score
		}
		if scored[i].Path != scored[j].Path {
			return scored[i].Path < scored[j].Path
		}
		return scored[i].StartLine < scored[j].StartLine
	})

	return scored
}

// scoreCandidate computes c's own Score: the strongest relationWeight
// among its own Reasons, plus an intentWeight bonus, plus a
// textRelevance bonus — all three are simple, additive, deterministic
// signals meaningful only for ordering candidates that already share
// one Tier (research.md #3).
func scoreCandidate(c Candidate, intent Intent, terms []string) int {
	best := 0
	relations := make([]string, 0, len(c.Reasons))
	for _, r := range c.Reasons {
		relations = append(relations, r.Relation)
		if w := relationWeight(r.Relation); w > best {
			best = w
		}
	}
	return best + intentWeight(intent, relations) + textRelevance(c.Heading, c.Content, terms)
}

// relationWeight approximates docs/context-engine-implementation.md
// §17's own illustrative weights — examples, not immutable
// requirements (research.md #3) — used only to break ties within a
// single Tier.
func relationWeight(relation string) int {
	switch relation {
	case "constitution", "target":
		return 100
	case "parent", "depends_on", "supersedes":
		return 90
	case "wikilink":
		return 80
	case "backlink":
		return 65
	case "text_match":
		return 0 // scored separately by textRelevance
	default:
		return 45
	}
}

// preferredRelationsByIntent maps each Intent onto the relation kinds
// this project's own vocabulary actually has (research.md #4) — a
// deliberately narrower, concrete translation of docs/context-engine-
// implementation.md §15's own more abstract, illustrative preferences.
var preferredRelationsByIntent = map[Intent]map[string]bool{
	IntentPlanning:       {"parent": true, "depends_on": true},
	IntentTasks:          {"depends_on": true, "parent": true},
	IntentImplementation: {"wikilink": true, "backlink": true, "text_match": true},
	IntentValidation:     {"backlink": true, "text_match": true},
	IntentAnalysis:       {"backlink": true, "text_match": true},
}

// intentWeight returns a small, fixed bonus when at least one of
// relations is preferred for intent (research.md #4), else 0. The
// zero-value Intent ("") has no preferences and always returns 0.
func intentWeight(intent Intent, relations []string) int {
	preferred, ok := preferredRelationsByIntent[intent]
	if !ok {
		return 0
	}
	for _, r := range relations {
		if preferred[r] {
			return 5
		}
	}
	return 0
}

// textRelevance is a simple, deterministic term-overlap heuristic —
// not a real BM25 value (research.md #5) — counting how many of terms
// occur in heading+content, capped at 4 occurrences and scaled to
// mirror docs/context-engine-implementation.md §17's own "BM25
// contribution 0..40" range.
func textRelevance(heading, content string, terms []string) int {
	if len(terms) == 0 {
		return 0
	}
	haystack := strings.ToLower(heading + "\n" + content)
	count := 0
	for _, term := range terms {
		count += strings.Count(haystack, term)
	}
	if count > 4 {
		count = 4
	}
	return count * 10
}

// queryTerms extracts lowercase, length-3-or-more words from req's own
// query text — Query verbatim, falling back to Task verbatim when
// Query is empty, the same resolution rule Collect's own Tier 4 uses
// (research.md #5) — never a fabricated question.
func queryTerms(req Request) []string {
	text := req.Query
	if text == "" {
		text = req.Task
	}
	if text == "" {
		return nil
	}
	fields := strings.Fields(strings.ToLower(text))
	terms := make([]string, 0, len(fields))
	for _, f := range fields {
		if len(f) >= 3 {
			terms = append(terms, f)
		}
	}
	return terms
}
