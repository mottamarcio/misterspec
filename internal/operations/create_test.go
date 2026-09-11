package operations_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestCreate_Program(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	result, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if result.ID.String() != "PRG-001" {
		t.Errorf("ID = %v, want PRG-001", result.ID)
	}
	if result.Path != "ai/programs/PRG-001/program.md" {
		t.Errorf("Path = %q, unexpected", result.Path)
	}

	inspected, err := operations.Inspect(root, cfg, "PRG-001")
	if err != nil {
		t.Fatalf("newly created Program failed Inspect(): %v", err)
	}
	if inspected.Metadata.Status != "draft" {
		t.Errorf("Status = %q, want %q", inspected.Metadata.Status, "draft")
	}
}

func TestCreate_FeatureUnderValidProgram(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	prg, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("Create(Program) unexpected error: %v", err)
	}

	feat, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Feature, Parent: prg.ID.String()})
	if err != nil {
		t.Fatalf("Create(Feature) unexpected error: %v", err)
	}
	if feat.ID.String() != "FEAT-001" {
		t.Errorf("ID = %v, want FEAT-001", feat.ID)
	}

	if _, err := operations.Inspect(root, cfg, "FEAT-001"); err != nil {
		t.Fatalf("newly created Feature failed Inspect(): %v", err)
	}
}

func TestCreate_SequentialSpecsUnderSameFeature(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	prg, _ := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Program})
	feat, _ := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Feature, Parent: prg.ID.String()})

	want := []string{"SPEC-001", "SPEC-002", "SPEC-003"}
	for _, w := range want {
		spec, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Spec, Parent: feat.ID.String()})
		if err != nil {
			t.Fatalf("Create(Spec) unexpected error: %v", err)
		}
		if spec.ID.String() != w {
			t.Fatalf("Create(Spec) = %v, want %v", spec.ID, w)
		}
	}
}

func TestCreate_InvalidParentRejectedNothingWritten(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	_, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Feature, Parent: "PRG-999"})
	if !errors.Is(err, operations.ErrInvalidParent) {
		t.Fatalf("Create() error = %v, want errors.Is(err, ErrInvalidParent)", err)
	}

	if _, statErr := os.Stat(filepath.Join(root, "ai/programs")); !os.IsNotExist(statErr) {
		t.Errorf("ai/programs unexpectedly exists after a rejected Create(): %v", statErr)
	}
}

func TestCreate_KnowledgeWithSlug(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	result, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Knowledge, Slug: "authentication"})
	if err != nil {
		t.Fatalf("Create(Knowledge) unexpected error: %v", err)
	}
	if result.Path != "ai/knowledge/KNOW-001-authentication.md" {
		t.Errorf("Path = %q, unexpected", result.Path)
	}

	learning, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Learning, Slug: "retry-strategy"})
	if err != nil {
		t.Fatalf("Create(Learning) unexpected error: %v", err)
	}
	if learning.Path != "ai/memory/learnings/LRN-001-retry-strategy.md" {
		t.Errorf("Path = %q, unexpected", learning.Path)
	}
}

// Note: Create's "target already exists" defensive check (FR-008) is not
// naturally reachable through black-box testing for auto-allocated types
// (Program/Feature/Spec/Knowledge/Learning) — by construction, NextID is
// always computed from what Scan already found, so a freshly allocated ID
// can never collide with an existing, Scan-visible artifact in normal
// operation. The check exists as a safety net (and its real, naturally
// reachable test coverage lives in create_artifact_test.go, where a
// second Plan/Tasks/Validation for the same Spec — a fixed, non-allocated
// path — genuinely can and does collide).

func TestCreate_UnsupportedType(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	_, err := operations.Create(root, cfg, operations.CreateRequest{Type: ids.Task})
	if !errors.Is(err, operations.ErrUnsupportedType) {
		t.Fatalf("Create() error = %v, want errors.Is(err, ErrUnsupportedType)", err)
	}
}
