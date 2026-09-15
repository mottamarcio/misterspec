// This file compiles and runs specs/006-agent-adapter/quickstart.md's
// end-to-end flow (list adapters, select by ID, install a fixture Skills
// set for Claude Code, then read back what's currently installed)
// against internal/agents on top of the generalized internal/installer
// (plan.md Phase 6, T018).
package example

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/agents/builtin"
	"github.com/mottamarcio/misterspec/internal/installer"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestAgentsQuickstart_EndToEnd(t *testing.T) {
	// 1. Discover available adapters.
	registry := builtin.Default()
	list := registry.List()
	if len(list) == 0 {
		t.Fatal("builtin.Default().List() is empty, want at least Claude Code")
	}

	// 2. Select one by ID.
	adapter, ok := registry.Get("claude-code")
	if !ok {
		t.Fatal(`registry.Get("claude-code") ok = false, want true`)
	}

	// An unknown ID is a distinct, non-error absence. "codex" is now a
	// real, registered adapter (018-multi-agent-skill-integration) — a
	// truly nonexistent ID is used here instead to keep proving this
	// same absence guarantee.
	if _, ok := registry.Get("not-a-real-agent"); ok {
		t.Fatal(`registry.Get("not-a-real-agent") ok = true, want false`)
	}

	// 3. Install a fixture Skills set for the selected adapter.
	root := testutil.Project(t)
	skills := fstest.MapFS{
		"create-plan.md": {Data: []byte("# create-plan\n")},
	}
	result, err := adapter.Install(context.Background(), agents.InstallRequest{
		ProjectRoot: root,
		Skills:      skills,
		Overwrite:   false,
	})
	if err != nil {
		t.Fatalf("adapter.Install() unexpected error: %v", err)
	}
	if result.AdapterID != "claude-code" || result.IntegrationPath != ".claude/skills" {
		t.Fatalf("result = %+v, unexpected", result)
	}
	for _, o := range result.Outcomes {
		if o.Status != installer.Installed {
			t.Fatalf("Outcome for %q Status = %v, want Installed", o.Resource.Name, o.Status)
		}
	}

	// 4. Read back what's currently installed.
	record, installed, err := agents.CurrentInstall(root)
	if err != nil {
		t.Fatalf("agents.CurrentInstall() unexpected error: %v", err)
	}
	if !installed {
		t.Fatal("CurrentInstall() installed = false, want true")
	}
	if record.Agent.ID != "claude-code" {
		t.Fatalf("record.Agent.ID = %q, want %q", record.Agent.ID, "claude-code")
	}

	// A project that was never installed reports "not installed", not an error.
	fresh := testutil.Project(t)
	_, installed, err = agents.CurrentInstall(fresh)
	if err != nil {
		t.Fatalf("agents.CurrentInstall() on a fresh project unexpected error: %v", err)
	}
	if installed {
		t.Fatal("CurrentInstall() on a fresh project installed = true, want false")
	}
}
