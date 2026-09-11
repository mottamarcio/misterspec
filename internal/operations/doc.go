// Package operations composes internal/project, internal/artifacts, and
// internal/ids into the read-only answers a coding agent or Skill
// actually asks for: locate and inspect an entity by ID alone (Resolve,
// Inspect), walk its structural parent/children (Parent, Children), and
// discover files or verify content integrity (Inventory, Fingerprint).
//
// Every operation in this package is read-only: none creates, modifies,
// or deletes any file. Mutating operations (create, create-artifact) are
// a distinct, later feature.
//
// See specs/002-read-operations/contracts/operations.md for this
// package's exported contract and specs/002-read-operations/data-model.md
// for its result and error types.
package operations
