// Package bootstrap composes misterspec's foundational primitives —
// internal/project's detection/configuration, internal/installer's kit
// materialization, and internal/agents's registry/adapter/install-record
// — into the deterministic "inspect, then bootstrap, then verify" core
// a future misterspec init will orchestrate
// (docs/architecture-specification.md §4, §36-37, §68).
//
// None of the three composed packages gains a new cross-dependency on
// another here: internal/project stays dependency-light, and this
// package's own config-writing logic (writeDefaultConfig) reuses
// project.Configuration's exported Default* constants and
// internal/installer's exported WriteAtomicFile rather than either
// package importing the other (see specs/007-project-bootstrap/research.md).
package bootstrap
