// Package gosource is a pure, single-purpose leaf parsing one Go
// source file's own imports and top-level declaration signatures via
// the standard library (go/parser) — no I/O beyond reading its own
// input file, no dependency on any other internal/* package. Shared
// by internal/architecture (import lists, for rule evaluation) and
// internal/context/index (declaration signatures, for code-context
// indexing) — 044-architecture-code-context-rules research.md #1/#2.
package gosource
