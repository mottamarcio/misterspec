package operations_test

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func testConfig() project.Configuration {
	return project.Configuration{
		SchemaVersion:       1,
		AgentID:             "claude-code",
		ArtifactsDir:        "ai",
		RawDir:              "ai/raw",
		KnowledgeDir:        "ai/knowledge",
		ConstitutionPath:    "ai/memory/constitution.md",
		LearningsDir:        "ai/memory/learnings",
		ProgramsRoot:        "ai/programs",
		IDWidth:             3,
		GitBranchAutomation: true,
	}
}

func TestResolve_UniqueMatch(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\n---\n")

	loc, err := operations.Resolve(root, cfg, "SPEC-014")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	wantPath := "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md"
	if loc.Path != wantPath {
		t.Errorf("Path = %q, want %q", loc.Path, wantPath)
	}
	if loc.Type != artifacts.TypeSpec {
		t.Errorf("Type = %v, want TypeSpec", loc.Type)
	}
	if loc.ID.Number != 14 {
		t.Errorf("ID.Number = %d, want 14", loc.ID.Number)
	}
}

func TestResolve_NotFound(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	_, err := operations.Resolve(root, cfg, "SPEC-999")
	if !errors.Is(err, operations.ErrEntityNotFound) {
		t.Fatalf("Resolve() error = %v, want errors.Is(err, ErrEntityNotFound)", err)
	}
}

func TestResolve_Ambiguous(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: draft\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-002/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: draft\nparent: PRG-002\n---\n")

	_, err := operations.Resolve(root, cfg, "FEAT-004")
	if !errors.Is(err, operations.ErrEntityAmbiguous) {
		t.Fatalf("Resolve() error = %v, want errors.Is(err, ErrEntityAmbiguous)", err)
	}
	var amb *operations.AmbiguousIDError
	if !errors.As(err, &amb) {
		t.Fatalf("Resolve() error = %v, want a *operations.AmbiguousIDError", err)
	}
	if amb.ID != "FEAT-004" {
		t.Errorf("AmbiguousIDError.ID = %q, want %q", amb.ID, "FEAT-004")
	}
	if len(amb.Locations) != 2 {
		t.Errorf("AmbiguousIDError.Locations = %v, want 2 entries", amb.Locations)
	}
}

func TestResolve_TaskResolvesToOwningTasksFileAndHeading(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Add session persistence\n\n- [ ] Complete\n")

	loc, err := operations.Resolve(root, cfg, "TASK-001")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	wantPath := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md#TASK-001"
	if loc.Path != wantPath {
		t.Errorf("Path = %q, want %q", loc.Path, wantPath)
	}
	if loc.Type != artifacts.TypeTasks {
		t.Errorf("Type = %v, want TypeTasks", loc.Type)
	}
}

func TestResolve_InvalidSyntax(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	tests := []string{"not-an-id", "SPEC", "SPEC-abc"}
	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			_, err := operations.Resolve(root, cfg, raw)
			if !errors.Is(err, operations.ErrInvalidTarget) {
				t.Fatalf("Resolve(%q) error = %v, want errors.Is(err, ErrInvalidTarget)", raw, err)
			}
		})
	}
}

func TestResolve_WrongWidthIsInvalidNotNotFound(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig() // IDWidth: 3
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\n---\n")

	// "SPEC-14" is syntactically self-consistent but does not match the
	// project's configured 3-digit width — must be rejected as invalid,
	// not silently matched to SPEC-014 by numeric value alone.
	_, err := operations.Resolve(root, cfg, "SPEC-14")
	if !errors.Is(err, operations.ErrInvalidTarget) {
		t.Fatalf("Resolve(\"SPEC-14\") error = %v, want errors.Is(err, ErrInvalidTarget)", err)
	}
}

func TestResolve_KnowledgeFlatFile(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-authentication.md", "---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n")

	loc, err := operations.Resolve(root, cfg, "KNOW-001")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if loc.Path != "ai/knowledge/KNOW-001-authentication.md" {
		t.Errorf("Path = %q, want the exact flat file path", loc.Path)
	}
	if loc.Type != artifacts.TypeKnowledge {
		t.Errorf("Type = %v, want TypeKnowledge", loc.Type)
	}
}
