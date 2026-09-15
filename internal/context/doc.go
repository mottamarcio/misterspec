// Package contextengine implements misterspec's Context Engine
// collector (docs/context-engine-implementation.md §13-18): given an
// explicit request naming a target artifact, deterministically collect
// a deduplicated, tier-labeled candidate set from this project's
// already-existing wikilink (011), reference-graph (012), document/
// chunk (013), and disposable-search-index (014) capabilities.
//
// Declared as package contextengine, not the directory-derived default
// of package context, specifically to avoid shadowing the standard
// library's own context package for every future caller
// (specs/015-context-collector/research.md #11) — the import path
// itself is unchanged (internal/context), matching 014's own
// internal/context/index package living alongside it.
//
// Collect never ranks or numerically scores its own output beyond the
// Tier-level distinctions it already labels every Candidate with, and
// applies no token budget itself — this package's own Rank and
// ApplyBudget (Phase 7, docs/context-engine-implementation.md §17-20)
// are the next, later step in the same pipeline: Rank assigns every
// Candidate a Score meaningful only within its own Tier — Tier always
// dominates Score, never the reverse (§17.1's own "important ranking
// invariant") — and ApplyBudget fits a ranked list into a token
// budget, filling tier by tier, always preserving mandatory content in
// full. Neither performs any I/O of its own; both operate purely on
// values this package already produces.
//
// Render (Phase 8, 017-internal-context-command) is this package's own
// last, purely presentational step: it converts an already-computed
// Result into a well-formed Markdown context pack, grouped by Tier,
// without altering selection, ranking, or budgeting in any way. The
// actual CLI entry point — "misterspec internal context" — lives in
// internal/cli/internalcmd, which orchestrates Collect, Rank,
// ApplyBudget, and (optionally) Render in sequence; this package
// remains unaware of any CLI, JSON, or index-lifecycle concern.
//
// See specs/015-context-collector/contracts/collector.md,
// specs/016-ranking-budgeting/contracts/ranking-budgeting.md, and
// specs/017-internal-context-command/contracts/context-command.md for
// this package's exported contract, and each feature's own
// data-model.md for its entity definitions and algorithms.
package contextengine
