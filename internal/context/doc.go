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
// applies no token budget — both remain Phase 7's own job
// (specs/015-context-collector/research.md #2). No CLI command is added
// here either — Phase 8's own later "internal context" command is the
// actual entry point this capability eventually feeds.
//
// See specs/015-context-collector/contracts/collector.md for this
// package's exported contract and specs/015-context-collector/
// data-model.md for its entity definitions and Collect's own algorithm.
package contextengine
