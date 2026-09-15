package builtin

import (
	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/agents/agy"
	"github.com/mottamarcio/misterspec/internal/agents/claude"
	"github.com/mottamarcio/misterspec/internal/agents/codex"
	"github.com/mottamarcio/misterspec/internal/agents/copilot"
	"github.com/mottamarcio/misterspec/internal/agents/cursoragent"
	"github.com/mottamarcio/misterspec/internal/agents/devin"
)

// Default returns a Registry pre-populated with every concrete adapter
// this build knows about: Claude Code, Antigravity (agy), Codex CLI,
// GitHub Copilot, Cursor (cursor-agent), and Devin for Terminal
// (018-multi-agent-skill-integration).
func Default() *agents.Registry {
	return agents.NewRegistry(
		claude.New(),
		agy.New(),
		codex.New(),
		copilot.New(),
		cursoragent.New(),
		devin.New(),
	)
}
