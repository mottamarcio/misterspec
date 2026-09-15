// Package installer implements misterspec's generic, agent-agnostic
// resource-materialization primitive: listing what the embedded kit
// (package kit) provides, and safely copying it onto disk — atomically,
// never silently overwriting, never touching a project's ai/ artifact
// tree.
//
// This package is distinct from internal/operations's Create/
// CreateArtifact: those allocate entity IDs and write into a project's
// artifact tree; this package only materializes framework resources
// (templates today) into a caller-chosen target directory. See
// specs/005-embedded-kit/research.md for the reasoning.
package installer
