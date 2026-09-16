package artifacts_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestParseMetadata_WellFormed(t *testing.T) {
	root := testutil.Project(t)
	path := testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md", ""+
		"---\n"+
		"id: SPEC-014\n"+
		"type: spec\n"+
		"status: ready\n"+
		"parent: FEAT-004\n"+
		"depends_on:\n"+
		"  - SPEC-011\n"+
		"supersedes: []\n"+
		"---\n"+
		"# Some Spec\n")

	meta, err := artifacts.ParseMetadata(path)
	if err != nil {
		t.Fatalf("ParseMetadata() unexpected error: %v", err)
	}
	if meta.ID == nil || meta.ID.String() != "SPEC-014" {
		t.Errorf("meta.ID = %v, want SPEC-014", meta.ID)
	}
	if meta.Status != "ready" {
		t.Errorf("meta.Status = %q, want %q", meta.Status, "ready")
	}
	if meta.Parent == nil || meta.Parent.String() != "FEAT-004" {
		t.Errorf("meta.Parent = %v, want FEAT-004", meta.Parent)
	}
	if len(meta.DependsOn) != 1 || meta.DependsOn[0].String() != "SPEC-011" {
		t.Errorf("meta.DependsOn = %v, want [SPEC-011]", meta.DependsOn)
	}
}

func TestParseMetadata_SchemaVersion(t *testing.T) {
	root := testutil.Project(t)
	path := testutil.WriteFile(t, root, "ai/memory/constitution.md", ""+
		"---\n"+
		"type: constitution\n"+
		"schema_version: 1\n"+
		"---\n"+
		"# Project Constitution\n")

	meta, err := artifacts.ParseMetadata(path)
	if err != nil {
		t.Fatalf("ParseMetadata() unexpected error: %v", err)
	}
	if meta.Type != "constitution" {
		t.Errorf("meta.Type = %q, want %q", meta.Type, "constitution")
	}
	if meta.SchemaVersion != 1 {
		t.Errorf("meta.SchemaVersion = %d, want 1", meta.SchemaVersion)
	}
}

func TestParseMetadata_ArtifactNotFound(t *testing.T) {
	root := testutil.Project(t)

	_, err := artifacts.ParseMetadata(filepath.Join(root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md"))
	if !errors.Is(err, artifacts.ErrArtifactNotFound) {
		t.Fatalf("ParseMetadata() error = %v, want errors.Is(err, ErrArtifactNotFound)", err)
	}
}

func TestParseMetadata_FrontmatterMalformed(t *testing.T) {
	root := testutil.Project(t)

	// No frontmatter delimiters at all.
	path := testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md", "# Just a heading, no frontmatter\n")
	_, err := artifacts.ParseMetadata(path)
	if !errors.Is(err, artifacts.ErrFrontmatterMalformed) {
		t.Fatalf("ParseMetadata() error = %v, want errors.Is(err, ErrFrontmatterMalformed) for missing frontmatter", err)
	}

	// Frontmatter delimiters present but invalid YAML inside.
	path2 := testutil.WriteFile(t, root, "ai/knowledge/KNOW-002-y.md", "---\nid: [this is not valid\n---\nbody\n")
	_, err = artifacts.ParseMetadata(path2)
	if !errors.Is(err, artifacts.ErrFrontmatterMalformed) {
		t.Fatalf("ParseMetadata() error = %v, want errors.Is(err, ErrFrontmatterMalformed) for invalid YAML", err)
	}
}

func TestParseMetadata_RequiredFieldMissing(t *testing.T) {
	root := testutil.Project(t)
	path := testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\ntype: spec\nstatus: draft\n---\n# no id field\n")

	_, err := artifacts.ParseMetadata(path)
	if !errors.Is(err, artifacts.ErrRequiredFieldMissing) {
		t.Fatalf("ParseMetadata() error = %v, want errors.Is(err, ErrRequiredFieldMissing)", err)
	}
}
