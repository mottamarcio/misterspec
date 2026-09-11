// Package validation implements misterspec's structural validation
// layer: checking entities and the whole project against the fixed
// structural rules frozen in docs/architecture-specification.md, and
// reporting every problem found as data (a Finding) rather than as a Go
// error that stops at the first one.
//
// This package deliberately does not depend on internal/operations —
// see specs/004-structural-validation/research.md for why: an import
// cycle with the new operations/status.go, and because
// operations.Resolve/Inspect's fail-fast error semantics are a poor fit
// for validation's "collect every anomaly" philosophy. It depends only
// on internal/ids, internal/artifacts, and internal/project.
package validation
