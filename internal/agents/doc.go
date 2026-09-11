// Package agents defines misterspec's coding-agent integration
// boundary: the Adapter interface every concrete agent integration
// implements, a Registry for discovering and selecting one, and the
// installation-record types/functions used to materialize Skills for a
// chosen agent and later report which one is currently installed
// (docs/architecture-specification.md §34-35).
//
// This package has no knowledge of any concrete adapter — see
// internal/agents/claude for the one this project currently ships, and
// internal/agents/builtin for where they're wired together (avoiding an
// import cycle; see specs/006-agent-adapter/research.md).
package agents
