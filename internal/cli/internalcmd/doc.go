// Package internalcmd implements the ten commands under misterspec's
// hidden "internal" command tree (docs/architecture-specification.md
// §9-18), plus the shared JSON-envelope (WriteSuccess/WriteError) and
// error-classification (classify) helpers every one of them uses.
//
// Every command here is a thin adapter: argument/flag parsing, one call
// into an already-implemented 001-007 package, and JSON shaping — no
// business logic of its own (specs/008-cli-cobra/spec.md FR-009).
package internalcmd
