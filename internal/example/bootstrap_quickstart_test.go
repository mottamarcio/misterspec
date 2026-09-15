// This file compiles and runs specs/007-project-bootstrap/quickstart.md's
// end-to-end flow (Inspect a fresh directory, Bootstrap it for a chosen
// agent, Inspect again confirming it's now initialized, then Verify
// confirming the installed agent matches) against internal/bootstrap on
// top of internal/agents/builtin's real Claude Code adapter
// (plan.md Phase 6, T012).
package example

import (
	"testing"
	"testing/fstest"

	"github.com/mottamarcio/misterspec/internal/agents/builtin"
	"github.com/mottamarcio/misterspec/internal/bootstrap"
)

func TestBootstrapQuickstart_EndToEnd(t *testing.T) {
	target := t.TempDir()
	registry := builtin.Default()
	skills := fstest.MapFS{
		"create-plan.md": {Data: []byte("# create-plan\n")},
	}

	// 1. Inspect before doing anything — a fresh directory is
	// uninitialized.
	before, err := bootstrap.Inspect(target)
	if err != nil {
		t.Fatalf("Inspect() unexpected error: %v", err)
	}
	if before.Initialized {
		t.Fatal("Inspect() on a fresh directory Initialized = true, want false")
	}

	// 2. Bootstrap it for a chosen, registered agent.
	outcome, err := bootstrap.Bootstrap(target, "claude-code", registry, skills)
	if err != nil {
		t.Fatalf("Bootstrap() unexpected error: %v", err)
	}
	if !outcome.ConfigWritten {
		t.Fatal("Bootstrap() outcome.ConfigWritten = false, want true")
	}
	if len(outcome.DirectoriesScaffolded) == 0 {
		t.Fatal("Bootstrap() outcome.DirectoriesScaffolded is empty, want the project's own directory hierarchy (021-init-scaffold-distribution)")
	}
	if outcome.AgentInstall.AdapterID != "claude-code" {
		t.Fatalf("Bootstrap() outcome.AgentInstall.AdapterID = %q, want %q", outcome.AgentInstall.AdapterID, "claude-code")
	}

	// 3. A second Bootstrap of the same target is rejected outright.
	if _, err := bootstrap.Bootstrap(target, "claude-code", registry, skills); err == nil {
		t.Fatal("second Bootstrap() of the same target error = nil, want bootstrap.ErrAlreadyInitialized")
	}

	// 4. Inspect again — now initialized, with the agent installed.
	after, err := bootstrap.Inspect(target)
	if err != nil {
		t.Fatalf("Inspect() after Bootstrap() unexpected error: %v", err)
	}
	if !after.Initialized || !after.AgentInstalled || after.InstalledAgent != "claude-code" {
		t.Fatalf("Inspect() after Bootstrap() = %+v, unexpected", after)
	}

	// 5. Verify confirms the bootstrap is genuinely usable.
	verify, err := bootstrap.Verify(target, "claude-code")
	if err != nil {
		t.Fatalf("Verify() unexpected error: %v", err)
	}
	if !verify.Detected || !verify.AgentMatches {
		t.Fatalf("Verify() = %+v, want Detected and AgentMatches both true", verify)
	}
}
