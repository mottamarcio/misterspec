package ids_test

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func testConfig() project.Configuration {
	return project.Configuration{
		SchemaVersion:    1,
		AgentID:          "claude-code",
		ArtifactsDir:     "ai",
		RawDir:           "ai/raw",
		KnowledgeDir:     "ai/knowledge",
		ConstitutionPath: "ai/memory/constitution.md",
		LearningsDir:     "ai/memory/learnings",
		ProgramsRoot:     "ai/programs",
		IDWidth:          3,
	}
}

func TestScan_SpecsAcrossFeaturesAndPrograms(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md", "---\nid: SPEC-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md", "---\nid: SPEC-002\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-002/features/FEAT-005/specs/SPEC-009/spec.md", "---\nid: SPEC-009\n---\n")

	result, err := ids.Scan(root, cfg, ids.Spec)
	if err != nil {
		t.Fatalf("Scan() unexpected error: %v", err)
	}
	if len(result.IDs) != 3 {
		t.Fatalf("Scan() found %d IDs, want 3: %+v", len(result.IDs), result.IDs)
	}
	if len(result.Duplicates) != 0 {
		t.Errorf("Scan() found unexpected duplicates: %+v", result.Duplicates)
	}

	next := ids.NextID(result.IDs, ids.Spec, cfg.IDWidth)
	if next.Number != 10 {
		t.Errorf("NextID() = %v, want SPEC-010 (max existing + 1)", next)
	}
}

func TestScan_DetectsDuplicateAcrossDifferentParents(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md", "---\nid: FEAT-004\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-002/features/FEAT-004/feature.md", "---\nid: FEAT-004\n---\n")

	result, err := ids.Scan(root, cfg, ids.Feature)
	if err != nil {
		t.Fatalf("Scan() unexpected error: %v", err)
	}
	if len(result.Duplicates) != 1 {
		t.Fatalf("Scan() found %d duplicate groups, want 1: %+v", len(result.Duplicates), result.Duplicates)
	}
	if got := result.Duplicates[0].ID.Number; got != 4 {
		t.Errorf("Duplicates[0].ID.Number = %d, want 4", got)
	}
	if len(result.Duplicates[0].Paths) != 2 {
		t.Errorf("Duplicates[0].Paths = %v, want 2 entries", result.Duplicates[0].Paths)
	}
	// The scan must still complete and report the rest of the project,
	// not halt because of the duplicate (FR-012).
	if result.IDs == nil {
		t.Error("Scan() IDs unexpectedly nil after a duplicate was found")
	}
}

func TestScan_EmptyTypeReturnsEmptyNotError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	result, err := ids.Scan(root, cfg, ids.Program)
	if err != nil {
		t.Fatalf("Scan() unexpected error: %v", err)
	}
	if len(result.IDs) != 0 {
		t.Errorf("Scan() found %d IDs, want 0", len(result.IDs))
	}
	if len(result.Duplicates) != 0 {
		t.Errorf("Scan() found %d duplicates, want 0", len(result.Duplicates))
	}
}

func TestScan_KnowledgeFlatFilesWithSlugs(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-authentication.md", "---\nid: KNOW-001\n---\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-002-payments.md", "---\nid: KNOW-002\n---\n")

	result, err := ids.Scan(root, cfg, ids.Knowledge)
	if err != nil {
		t.Fatalf("Scan() unexpected error: %v", err)
	}
	if len(result.IDs) != 2 {
		t.Fatalf("Scan() found %d IDs, want 2: %+v", len(result.IDs), result.IDs)
	}
}

func TestResolveTarget_ResolvesToExactlyOneArtifact(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-x.md", "---\nid: KNOW-003\n---\n")

	id, paths, err := ids.ResolveTarget(root, cfg, "KNOW-003")
	if err != nil {
		t.Fatalf("ResolveTarget() unexpected error: %v", err)
	}
	if id.Type != ids.Knowledge || id.Number != 3 {
		t.Errorf("ResolveTarget() id = %+v, want Type=Knowledge Number=3", id)
	}
	if len(paths) != 1 {
		t.Fatalf("ResolveTarget() paths = %v, want exactly 1", paths)
	}
}

func TestResolveTarget_BrokenReturnsZeroPaths(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	id, paths, err := ids.ResolveTarget(root, cfg, "SPEC-999")
	if err != nil {
		t.Fatalf("ResolveTarget() unexpected error: %v", err)
	}
	if id.Type != ids.Spec || id.Number != 999 {
		t.Errorf("ResolveTarget() id = %+v, want Type=Spec Number=999", id)
	}
	if len(paths) != 0 {
		t.Errorf("ResolveTarget() paths = %v, want none for a nonexistent target", paths)
	}
}

func TestResolveTarget_AmbiguousReturnsEveryClaimingPath(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	testutil.WriteFile(t, root, "ai/knowledge/KNOW-005-a.md", "---\nid: KNOW-005\n---\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-005-b.md", "---\nid: KNOW-005\n---\n")

	id, paths, err := ids.ResolveTarget(root, cfg, "KNOW-005")
	if err != nil {
		t.Fatalf("ResolveTarget() unexpected error: %v", err)
	}
	if id.Number != 5 {
		t.Errorf("ResolveTarget() id.Number = %d, want 5", id.Number)
	}
	if len(paths) != 2 {
		t.Fatalf("ResolveTarget() paths = %v, want exactly 2 claiming paths", paths)
	}
}

func TestResolveTarget_MalformedTargetIsError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	_, _, err := ids.ResolveTarget(root, cfg, "nodash")
	if err == nil {
		t.Fatal("ResolveTarget() expected an error for a malformed target, got nil")
	}
}

func TestResolveTarget_DifferentZeroPaddingWidthStillResolves(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig() // cfg.IDWidth is 3

	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md", "---\nid: SPEC-002\n---\n")

	// Written with width 1, not the project's configured width 3 — must
	// still resolve, matching how a frontmatter field already tolerates
	// this (internal/artifacts's parseFieldID) and how 011's own wikilink
	// classification already does.
	id, paths, err := ids.ResolveTarget(root, cfg, "SPEC-2")
	if err != nil {
		t.Fatalf("ResolveTarget() unexpected error: %v", err)
	}
	if id.Number != 2 {
		t.Errorf("ResolveTarget() id.Number = %d, want 2", id.Number)
	}
	if len(paths) != 1 {
		t.Fatalf("ResolveTarget() paths = %v, want exactly 1", paths)
	}
}

func TestScan_TaskHeadingsInsideTasksFiles(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"# Tasks\n\n## TASK-001 — Add session persistence\n\n- [ ] Complete\n\n## TASK-002 — Add tests\n\n- [ ] Complete\n")

	result, err := ids.Scan(root, cfg, ids.Task)
	if err != nil {
		t.Fatalf("Scan() unexpected error: %v", err)
	}
	if len(result.IDs) != 2 {
		t.Fatalf("Scan() found %d task IDs, want 2: %+v", len(result.IDs), result.IDs)
	}
}
