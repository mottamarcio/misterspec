package index

import (
	"testing"

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

// setupFullFixture writes exactly one artifact of each
// artifacts.ArtifactType value (research.md #3's own full indexing
// scope), wired together where a canonical parent relationship is
// required.
func setupFullFixture(t *testing.T, root string) {
	t.Helper()
	testutil.WriteFile(t, root, "ai/memory/constitution.md",
		"---\ntype: constitution\n---\n## Principles\n\nFilesystem is the source of truth.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nRefresh token rotation.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/plan.md",
		"---\ntype: plan\nfor: SPEC-001\n---\n## Summary\n\nThe implementation plan.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do it\n\n- [ ] Complete\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/validation.md",
		"---\ntype: validation\nfor: SPEC-001\n---\n## Results\n\nAll green.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-auth.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nAuthentication model facts.\n")
	testutil.WriteFile(t, root, "ai/memory/learnings/LRN-001-x.md",
		"---\nid: LRN-001\ntype: learning\nstatus: active\n---\n## Summary\n\nA learning.\n")
}

// setupUnrelatedFixture writes only content under ai/ that ClassifyPath
// cannot classify as any real artifact — proving such content is simply
// skipped, never indexed (research.md #3).
func setupUnrelatedFixture(t *testing.T, root string) {
	t.Helper()
	testutil.WriteFile(t, root, "ai/notes.md", "# Not a real artifact\n\nJust a stray file.\n")
}

func openTestStore(t *testing.T, root string) *sqliteStore {
	t.Helper()
	store, err := Open(root + "/.misterspec/cache/context.db")
	if err != nil {
		t.Fatalf("Open() unexpected error: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	ss, ok := store.(*sqliteStore)
	if !ok {
		t.Fatalf("Open() returned %T, want *sqliteStore", store)
	}
	return ss
}
