// Package cli builds misterspec's Cobra command tree: the small public
// surface ("init") plus the hidden "internal" tree
// (docs/architecture-specification.md §4-5). It contains no business
// logic of its own — every command is a thin adapter over 001-007's
// already-implemented Go packages, composed via internal/cli/internalcmd
// for the ten deterministic operations and directly (bootstrap.Bootstrap)
// for "init" (see specs/008-cli-cobra/research.md).
package cli
