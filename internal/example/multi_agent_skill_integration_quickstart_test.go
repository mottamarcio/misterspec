// This file compiles and runs specs/018-multi-agent-skill-integration/
// quickstart.md's end-to-end flow (listing all six registered adapters
// and their TargetPath()s; installing for cursor-agent; installing
// both agy and codex into one project and confirming both survive;
// asserting the four updated SKILL.md files' own content contains
// "internal context" with their respective correct intents) against
// internal/agents/builtin and kit.SkillsFS.
package example

import (
	"context"
	"io/fs"
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/agents/builtin"
	"github.com/mottamarcio/misterspec/internal/testutil"
	"github.com/mottamarcio/misterspec/kit"
)

func TestMultiAgentSkillIntegrationQuickstart_EndToEnd(t *testing.T) {
	// 1. Discover the six now-supported agents (quickstart.md §1).
	registry := builtin.Default()
	wantTargets := map[string]string{
		"agy":          ".agents/skills",
		"claude-code":  ".claude/skills",
		"codex":        ".agents/skills",
		"copilot":      ".github/skills",
		"cursor-agent": ".cursor/skills",
		"devin":        ".devin/skills",
	}
	for id, target := range wantTargets {
		a, ok := registry.Get(id)
		if !ok {
			t.Fatalf("registry.Get(%q) ok = false, want true", id)
		}
		if a.TargetPath() != target {
			t.Errorf("Get(%q).TargetPath() = %q, want %q", id, a.TargetPath(), target)
		}
	}

	// 2. Install for a new agent (Cursor) (quickstart.md §2).
	root := testutil.Project(t)
	ctx := context.Background()
	cursorAdapter, _ := registry.Get("cursor-agent")
	result, err := cursorAdapter.Install(ctx, agents.InstallRequest{
		ProjectRoot: root,
		Skills:      kit.SkillsFS,
		Overwrite:   false,
	})
	if err != nil {
		t.Fatalf("cursor-agent Install() unexpected error: %v", err)
	}
	if len(result.Outcomes) == 0 {
		t.Fatal("cursor-agent Install() produced no outcomes")
	}

	// 3. Install for two agents that share a target directory
	// (quickstart.md §3): agy and codex both scan .agents/skills.
	codexAdapter, _ := registry.Get("codex")
	if _, err := codexAdapter.Install(ctx, agents.InstallRequest{ProjectRoot: root, Skills: kit.SkillsFS}); err != nil {
		t.Fatalf("codex Install() unexpected error: %v", err)
	}
	agyAdapter, _ := registry.Get("agy")
	if _, err := agyAdapter.Install(ctx, agents.InstallRequest{ProjectRoot: root, Skills: kit.SkillsFS}); err != nil {
		t.Fatalf("agy Install() unexpected error: %v", err)
	}

	// 4-5. The four updated Skills each request a Context Pack with
	// their own correct intent (quickstart.md §4-5).
	wantIntent := map[string]string{
		"mister-implement": "implementation",
		"mister-plan":      "planning",
		"mister-tasks":     "tasks",
		"mister-analyze":   "validation",
	}
	for skill, intent := range wantIntent {
		data, err := fs.ReadFile(kit.SkillsFS, skill+"/SKILL.md")
		if err != nil {
			t.Fatalf("reading kit/skills/%s/SKILL.md: %v", skill, err)
		}
		content := string(data)

		wantOp := "internal context SPEC-### --intent " + intent
		if !strings.Contains(content, wantOp) {
			t.Errorf("%s/SKILL.md does not contain %q", skill, wantOp)
		}
		if !strings.Contains(strings.ToLower(content), "free to") {
			t.Errorf("%s/SKILL.md does not state the agent remains free to explore further (FR-007)", skill)
		}
		if !strings.Contains(content, "If the request fails, proceed") &&
			!strings.Contains(content, "If this fails, proceed") {
			t.Errorf("%s/SKILL.md does not state a fallback on a failed Context Pack request (FR-008)", skill)
		}
	}
}
