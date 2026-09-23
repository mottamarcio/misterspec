// Package evidence parses a Task's own "Evidence-*:" body lines
// (041-task-evidence-fingerprint data-model.md "EvidenceFields") and
// derives its real completeness state (EvidenceState) from them plus
// the Task's current content fingerprint — a development-time-scale
// package, mirroring internal/prepare's own shape (plan.md
// "Scale/Scope").
//
// This package is a leaf: it depends only on internal/ids and the
// standard library, deliberately, so both internal/prepare and
// internal/validation can import it without an import cycle —
// internal/operations already imports internal/validation (status.go),
// so internal/validation cannot import internal/operations, which is
// why the shared content-fingerprint primitive lives here rather than
// solely in internal/operations (correction found during
// implementation; see ContentFingerprint in state.go and
// research.md #6/#8).
package evidence
