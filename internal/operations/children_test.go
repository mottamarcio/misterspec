package operations_test

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestChildren_ProgramWithTwoFeatures(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: draft\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: draft\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-002/feature.md", "---\nid: FEAT-002\ntype: feature\nstatus: draft\nparent: PRG-001\n---\n")
	// A sibling Program that must NOT be included.
	testutil.WriteFile(t, root, "ai/programs/PRG-002/features/FEAT-003/feature.md", "---\nid: FEAT-003\ntype: feature\nstatus: draft\nparent: PRG-002\n---\n")

	children, err := operations.Children(root, cfg, "PRG-001", nil)
	if err != nil {
		t.Fatalf("Children() unexpected error: %v", err)
	}
	if len(children) != 2 {
		t.Fatalf("Children() = %d entries, want 2: %+v", len(children), children)
	}
	got := map[string]bool{}
	for _, c := range children {
		got[c.ID.String()] = true
	}
	if !got["FEAT-001"] || !got["FEAT-002"] {
		t.Errorf("Children() = %v, want FEAT-001 and FEAT-002", got)
	}
	if got["FEAT-003"] {
		t.Error("Children() unexpectedly included FEAT-003 from a sibling program")
	}
}

func TestChildren_FilteredByType(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: draft\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: draft\nparent: PRG-001\n---\n")

	featureType := ids.Feature
	children, err := operations.Children(root, cfg, "PRG-001", &featureType)
	if err != nil {
		t.Fatalf("Children() unexpected error: %v", err)
	}
	if len(children) != 1 || children[0].ID.String() != "FEAT-001" {
		t.Fatalf("Children() = %+v, want [FEAT-001]", children)
	}
}

func TestChildren_EmptyForTypeItHasNone(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: draft\n---\n")

	specType := ids.Spec
	children, err := operations.Children(root, cfg, "PRG-001", &specType)
	if err != nil {
		t.Fatalf("Children() unexpected error: %v", err)
	}
	if len(children) != 0 {
		t.Errorf("Children() = %+v, want empty (Program has no Spec children directly)", children)
	}
}
