// Package internalcmd implements the commands under misterspec's hidden
// "internal" command tree (docs/architecture-specification.md §9-18;
// references/backlinks added by 012-references-backlinks per
// docs/context-engine-implementation.md §7; context added by
// 017-internal-context-command per docs/context-engine-
// implementation.md §21), plus the shared JSON-envelope
// (WriteSuccess/WriteError) and error-classification (classify) helpers
// every one of them uses.
//
// Every command here is a thin adapter: argument/flag parsing, one call
// into an already-implemented package, and JSON shaping — no business
// logic of its own (specs/008-cli-cobra/spec.md FR-009). NewContextCmd
// is the one exception that orchestrates more than a single call: it
// sequences index.Open/Store.Sync (014) before Collect (015) and
// Rank/ApplyBudget (016), but still performs no ranking, budgeting, or
// indexing logic itself — see specs/017-internal-context-command/
// contracts/context-command.md.
package internalcmd
