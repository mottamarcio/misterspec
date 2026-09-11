// Package kit is misterspec's embedded resource root: the framework's
// templates (and, in later features, its canonical Skills and
// agent-integration resources) are compiled directly into the binary
// from this directory tree, so nothing at runtime requires a network
// connection to retrieve them
// (docs/architecture-specification.md §33).
//
// This is the single source of truth for template content: both
// internal/templates (rendering) and internal/installer (raw
// materialization) read from the same embed.FS declared here — see
// specs/005-embedded-kit/research.md.
package kit
