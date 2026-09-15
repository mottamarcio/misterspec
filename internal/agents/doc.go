// Package agents defines misterspec's coding-agent integration
// boundary: the Adapter interface every concrete agent integration
// implements, a Registry for discovering and selecting one, and the
// installation-record types/functions used to materialize Skills for a
// chosen agent and later report which one is currently installed
// (docs/architecture-specification.md §34-35).
//
// This package has no knowledge of any concrete adapter — see
// internal/agents/claude, internal/agents/agy, internal/agents/codex,
// internal/agents/copilot, internal/agents/cursoragent, and
// internal/agents/devin for the six this project currently ships
// (018-multi-agent-skill-integration), and internal/agents/builtin for
// where they're wired together (avoiding an import cycle; see
// specs/006-agent-adapter/research.md). Every one of the five added by
// 018 is a byte-for-byte copy of claude's own Install body — research
// confirmed all six agents scan the same directory-per-skill SKILL.md
// "Agent Skills" convention, differing only in which directory each
// one scans (specs/018-multi-agent-skill-integration/contracts/
// adapters-and-skills.md).
package agents
