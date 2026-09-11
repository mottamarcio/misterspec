package builtin_test

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/agents/builtin"
)

func TestDefault_IncludesClaudeCode(t *testing.T) {
	registry := builtin.Default()

	adapter, ok := registry.Get("claude-code")
	if !ok {
		t.Fatal(`Default().Get("claude-code") ok = false, want true`)
	}
	if adapter.TargetPath() != ".claude/skills" {
		t.Errorf("TargetPath() = %q, want %q", adapter.TargetPath(), ".claude/skills")
	}
}
