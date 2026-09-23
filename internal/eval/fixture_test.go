package eval

import (
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/context/index"
	"github.com/mottamarcio/misterspec/internal/project"
)

// testConfig mirrors internal/context's own fixture_test.go testConfig
// — the standard, frozen Configuration shape every fixture project in
// this repository's tests uses.
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

// openSyncedStore opens and synchronizes an index.Store rooted at root
// — the shared setup RunRetrievalCase's own Tier 4 (text search) reads
// from, mirroring internal/context/fixture_test.go's own helper.
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
