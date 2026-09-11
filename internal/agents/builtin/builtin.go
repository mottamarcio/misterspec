package builtin

import (
	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/agents/claude"
)

// Default returns a Registry pre-populated with every concrete adapter
// this build knows about. Today: Claude Code only — additional adapters
// (Codex, Gemini CLI, ...) remain a later release decision
// (docs/architecture-specification.md §34; specs/006-agent-adapter/spec.md
// Assumptions).
func Default() *agents.Registry {
	return agents.NewRegistry(claude.New())
}
