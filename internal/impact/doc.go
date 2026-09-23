// Package impact computes a change-triggered impact report: given two
// Git revisions, it builds the set of artifact- and Requirement-level
// elements that changed between them (ChangeSet), then walks the
// reverse formal (depends_on/parent/supersedes), Requirement-coverage
// (Serves:), evidence (041-task-evidence-fingerprint), and wikilink
// (038-wikilink-chunk-provenance) relations already computed by
// internal/operations, internal/validation, and internal/evidence to
// find every artifact or Task with a known relation to that change —
// each reported with its full propagation path, a specific reason, a
// classification (deterministic invalidation vs. suggested review),
// and a severity (042-impact-analysis-review plan.md "Summary").
//
// This package is layered like internal/prepare: it imports
// internal/operations, internal/validation, internal/evidence,
// internal/vcs, internal/ids, and internal/artifacts, and is itself
// imported by internal/cli/internalcmd — never the reverse, avoiding
// the codebase's one import-cycle risk (internal/operations already
// imports internal/validation).
package impact
