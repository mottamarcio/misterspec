// Package testutil provides shared, test-only fixture helpers for building
// temporary misterspec project trees (t.TempDir()-based), reused by the
// filesystem-integration tests of internal/project, internal/artifacts,
// and internal/ids alike, per the project constitution's Test-First
// discipline (Principle V).
//
// Nothing in this package is imported by non-test code.
package testutil
