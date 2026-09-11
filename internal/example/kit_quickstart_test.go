// This file compiles and runs specs/005-embedded-kit/quickstart.md's
// end-to-end flow (List, Install fresh, Install again without overwrite,
// Install with overwrite, then confirm Create still works from the same
// kit) against internal/installer and internal/operations
// (plan.md Phase 6, T015).
package example

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/installer"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestKitQuickstart_EndToEnd(t *testing.T) {
	// 1. Discover what the kit provides — no filesystem writes.
	resources := installer.List()
	if len(resources) != 8 {
		t.Fatalf("installer.List() = %d resources, want 8", len(resources))
	}

	// 2. Install into a fresh directory.
	kitTarget := t.TempDir()
	outcomes, err := installer.Install(kitTarget, false)
	if err != nil {
		t.Fatalf("installer.Install() unexpected error: %v", err)
	}
	for _, o := range outcomes {
		if o.Status != installer.Installed {
			t.Fatalf("Install() outcome for %q = %v, want Installed", o.Resource.Name, o.Status)
		}
	}

	// 3. Re-running without overwrite changes nothing.
	outcomes, err = installer.Install(kitTarget, false)
	if err != nil {
		t.Fatalf("installer.Install() (no overwrite) unexpected error: %v", err)
	}
	for _, o := range outcomes {
		if o.Status != installer.Skipped {
			t.Fatalf("Install() (no overwrite) outcome for %q = %v, want Skipped", o.Resource.Name, o.Status)
		}
	}

	// 4. Re-running with overwrite replaces everything.
	outcomes, err = installer.Install(kitTarget, true)
	if err != nil {
		t.Fatalf("installer.Install() (overwrite) unexpected error: %v", err)
	}
	for _, o := range outcomes {
		if o.Status != installer.Installed {
			t.Fatalf("Install() (overwrite) outcome for %q = %v, want Installed", o.Resource.Name, o.Status)
		}
	}

	// 5. Artifact creation still works, rendered from the same kit
	// Install materialized raw copies of (FR-008).
	root := testutil.Project(t)
	proj, err := project.Detect(root)
	if err != nil {
		t.Fatalf("project.Detect() unexpected error: %v", err)
	}
	result, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("operations.Create() unexpected error: %v", err)
	}
	if result.ID.String() != "PRG-001" {
		t.Fatalf("Create().ID = %v, want PRG-001", result.ID)
	}
}
