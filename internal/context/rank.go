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
	// Components breaks Score into its individual contributors,
	// computed alongside Score (036-text-search-ranking spec FR-008) —
	// meaningful for diagnostic display only; ordering always uses
	// Score, never Components directly.
	Components ScoreComponents
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
		comp := computeScoreComponents(c, req.Intent, terms)
		scored[i] = ScoredCandidate{Candidate: c, Score: comp.Total, Components: comp}
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

// ScoreComponents breaks one candidate's Score down into its individual
// contributors (036-text-search-ranking data-model.md ScoreComponents)
// — computed alongside Score but surfaced only when a caller explicitly
// requests diagnostic detail (spec FR-008). Total always equals the
// corresponding ScoredCandidate.Score.
type ScoreComponents struct {
	Tier           int
	RelationWeight int
	IntentBonus    int
	TextRelevance  int
	Total          int
}

// computeScoreComponents computes c's own Score, broken into its three
// additive contributors: the strongest relationWeight among its own
// Reasons, an intentWeight bonus, and a text-relevance bonus — all
// three simple, deterministic signals meaningful only for ordering
// candidates that already share one Tier (research.md #3).
func computeScoreComponents(c Candidate, intent Intent, terms []string) ScoreComponents {
	best := 0
	relations := make([]string, 0, len(c.Reasons))
	for _, r := range c.Reasons {
		relations = append(relations, r.Relation)
		if w := relationWeight(r.Relation); w > best {
			best = w
		}
	}
	intentBonus := intentWeight(intent, relations)
	textBonus := textRelevanceContribution(c, relations, terms)
	return ScoreComponents{
		Tier:           int(minTier(c.Reasons)),
		RelationWeight: best,
		IntentBonus:    intentBonus,
		TextRelevance:  textBonus,
		Total:          best + intentBonus + textBonus,
	}
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

// textRelevanceContribution returns c's own text-relevance score
// contribution: for a candidate carrying a text_match Reason (Tier 4),
// the real bm25() signal already computed by index.Search and carried
// on c.TextRank (036-text-search-ranking spec FR-005) — replacing the
// term-occurrence heuristic this package used before; for every other
// candidate, the unchanged term-overlap heuristic below (data-model.md
// Candidate validation rule: TextRank is only ever read here when
// text_match is present).
func textRelevanceContribution(c Candidate, relations []string, terms []string) int {
	if hasTextMatch(relations) {
		return bm25TextRelevance(c.TextRank)
	}
	return textRelevance(c.Heading, c.Content, terms)
}

// hasTextMatch reports whether relations includes "text_match".
func hasTextMatch(relations []string) bool {
	for _, r := range relations {
		if r == "text_match" {
			return true
		}
	}
	return false
}

// bm25TextRelevance converts FTS5's own bm25() value (rank; lower, more
// negative, is more relevant — index.Search's own convention) into this
// package's higher-is-better additive scoring space (036-text-search-
// ranking research.md #4), replacing the previous term-occurrence
// heuristic for text_match candidates. rank >= 0 — including the zero
// value, meaning "no real bm25 signal was ever recorded" (e.g. a
// Candidate built directly, not through Tier 4 collection) — always
// contributes 0. Otherwise scaled and clamped to [0,40], mirroring the
// previous heuristic's own 0..40 range.
func bm25TextRelevance(rank float64) int {
	if rank >= 0 {
		return 0
	}
	v := int(-rank * 10)
	if v > 40 {
		v = 40
	}
	return v
}

// textRelevance is a simple, deterministic term-overlap heuristic —
// not a real BM25 value (research.md #5) — counting how many of terms
// occur in heading+content, capped at 4 occurrences and scaled to
// mirror docs/context-engine-implementation.md §17's own "BM25
// contribution 0..40" range. Applied to every candidate that does not
// carry a text_match Reason (036-text-search-ranking research.md #4);
// text_match candidates use bm25TextRelevance instead.
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
