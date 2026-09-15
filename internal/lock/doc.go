// Package lock implements misterspec's short-lived, non-authoritative
// project lock (.misterspec/.lock), used to serialize the deterministic
// operations that mutate the filesystem (Create, CreateArtifact) so
// concurrent requests never collide and an interrupted prior attempt is
// always safely recoverable — never a permanent block.
//
// Per the project constitution (Principle III), this lock is operational
// only: it holds no semantic project state and is never treated as an
// authoritative source of truth. See
// specs/003-entity-creation/contracts/creation.md for this package's
// exported contract.
package lock
