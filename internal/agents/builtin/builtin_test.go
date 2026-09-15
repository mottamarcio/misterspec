package builtin_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/agents/builtin"
	"github.com/mottamarcio/misterspec/internal/installer"
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

// TestDefault_IncludesEveryAgent proves all six adapters
// (018-multi-agent-skill-integration) are registered, each with its own
// documented, research.md-verified TargetPath().
func TestDefault_IncludesEveryAgent(t *testing.T) {
	registry := builtin.Default()

	want := map[string]string{
		"claude-code":  ".claude/skills",
		"agy":          ".agents/skills",
		"codex":        ".agents/skills",
		"copilot":      ".github/skills",
		"cursor-agent": ".cursor/skills",
		"devin":        ".devin/skills",
	}
	for id, targetPath := range want {
		adapter, ok := registry.Get(id)
		if !ok {
			t.Errorf(`Default().Get(%q) ok = false, want true`, id)
			continue
		}
		if adapter.TargetPath() != targetPath {
			t.Errorf("Get(%q).TargetPath() = %q, want %q", id, adapter.TargetPath(), targetPath)
		}
	}

	list := registry.List()
	if len(list) != len(want) {
		t.Errorf("List() = %d adapters, want %d", len(list), len(want))
	}
}

// TestDefault_AgyAndCodexCoexist proves installing both agy and codex
// into the same project — which share the identical, verified
// .agents/skills target directory (research.md #1) — never corrupts
// either one's own installed Skill files (FR-004), even though
// install.json (§22's own single-agent schema, unchanged by this
// feature) only ever reflects whichever adapter was recorded most
// recently.
func TestDefault_AgyAndCodexCoexist(t *testing.T) {
	root := t.TempDir()
	skills := fstest.MapFS{"implement.md": {Data: []byte("# implement\n")}}
	registry := builtin.Default()
	ctx := context.Background()

	codexAdapter, _ := registry.Get("codex")
	if _, err := codexAdapter.Install(ctx, agents.InstallRequest{ProjectRoot: root, Skills: skills}); err != nil {
		t.Fatalf("codex Install() unexpected error: %v", err)
	}

	agyAdapter, _ := registry.Get("agy")
	result, err := agyAdapter.Install(ctx, agents.InstallRequest{ProjectRoot: root, Skills: skills})
	if err != nil {
		t.Fatalf("agy Install() unexpected error: %v", err)
	}
	// Both adapters target .agents/skills with identical source content
	// — agy's own call finds codex's already-installed file and skips
	// it (Overwrite defaults to false), which is correct, harmless
	// behavior, not corruption: the content is byte-identical either
	// way.
	for _, o := range result.Outcomes {
		if o.Status != installer.Skipped && o.Status != installer.Installed {
			t.Errorf("agy Outcome for %q Status = %v, want Skipped or Installed", o.Resource.Name, o.Status)
		}
	}

	got, err := os.ReadFile(filepath.Join(root, ".agents", "skills", "implement.md"))
	if err != nil {
		t.Fatalf("reading shared skill file: %v", err)
	}
	if string(got) != "# implement\n" {
		t.Errorf("shared skill file content = %q, want unchanged fixture content", got)
	}
}
