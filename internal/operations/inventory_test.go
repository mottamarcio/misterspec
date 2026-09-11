package operations_test

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestInventory_PopulatedDirectory(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/raw/prd.md", "# PRD\n")
	testutil.WriteFile(t, root, "ai/raw/architecture.md", "# Architecture\n")

	files, err := operations.Inventory(root, "ai/raw")
	if err != nil {
		t.Fatalf("Inventory() unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("Inventory() = %d files, want 2: %+v", len(files), files)
	}
	for _, f := range files {
		if f.Extension != ".md" {
			t.Errorf("Extension = %q, want %q for %s", f.Extension, ".md", f.Path)
		}
		if f.Size == 0 {
			t.Errorf("Size = 0 for %s, want > 0", f.Path)
		}
	}
}

func TestInventory_EmptyDirectory(t *testing.T) {
	root := testutil.Project(t)

	files, err := operations.Inventory(root, "ai/raw")
	if err != nil {
		t.Fatalf("Inventory() unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Inventory() = %+v, want empty", files)
	}
}

func TestInventory_NotYetCreatedDirectoryIsEmptyNotError(t *testing.T) {
	root := testutil.Project(t)

	files, err := operations.Inventory(root, "ai/knowledge")
	if err != nil {
		t.Fatalf("Inventory() unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Inventory() = %+v, want empty", files)
	}
}

func TestInventory_RejectsTraversalOutsideRoot(t *testing.T) {
	root := testutil.Project(t)

	_, err := operations.Inventory(root, "../../etc")
	if !errors.Is(err, artifacts.ErrPathOutsideProject) {
		t.Fatalf("Inventory() error = %v, want errors.Is(err, artifacts.ErrPathOutsideProject)", err)
	}
}

func TestInventory_EntityOwnDirectory(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md", "---\nid: SPEC-014\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/plan.md", "---\ntype: plan\nfor: SPEC-014\n---\n")

	files, err := operations.Inventory(root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014")
	if err != nil {
		t.Fatalf("Inventory() unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("Inventory() = %d files, want 2: %+v", len(files), files)
	}
}
