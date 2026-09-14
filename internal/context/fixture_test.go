package contextengine

import (
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/context/index"
	"github.com/mottamarcio/misterspec/internal/project"
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

// openSyncedStore opens an index.Store rooted at root and synchronizes
// it once against the project's current state — the shared setup every
// Collect test in this package needs, since Tier 4 (text search) reads
// from the index.
func openSyncedStore(t *testing.T, root string, cfg project.Configuration) index.Store {
	t.Helper()
	store, err := index.Open(filepath.Join(root, ".misterspec/cache/context.db"))
	if err != nil {
		t.Fatalf("index.Open() unexpected error: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}
	return store
}
