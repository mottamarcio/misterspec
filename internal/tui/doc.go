// Package tui implements misterspec's interactive "misterspec init" flow
// (docs/architecture-specification.md §36-37): a Bubble Tea Model
// walking through Inspect, a Non-Empty/Already-Initialized Warning,
// Agent Selection, Preview, Installing, Success, and Error screens.
//
// This package holds UI state only — never filesystem business rules
// (§37's own explicit rule). Every actual decision (is this target
// safe, is this agent valid, what will installing actually do) comes
// from internal/bootstrap, internal/agents, and internal/installer,
// already proven elsewhere; the one filesystem mutation anywhere in
// this package is a single bootstrap.Bootstrap call triggered only by
// explicit user confirmation on the Preview screen (see
// specs/010-interactive-init-tui/research.md).
package tui
