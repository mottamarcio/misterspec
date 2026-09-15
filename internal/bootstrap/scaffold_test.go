package bootstrap

import (
	"os"
	"path/filepath"
	"testing"
)

// wantScaffoldedDirs is data-model.md's own fixed order.
var wantScaffoldedDirs = []string{
	"ai",
	"ai/raw",
	"ai/knowledge",
	"ai/memory",
	"ai/memory/learnings",
	"ai/programs",
}

// wantGitkeepDirs is the subset of wantScaffoldedDirs that are
// genuine leaves — "ai" and "ai/memory" are each the parent of
// another scaffolded directory, so once scaffolding completes they
// already contain real subdirectories of their own and correctly get
// no .gitkeep (see scaffold.go's own two-pass rationale).
var wantGitkeepDirs = []string{
	"ai/raw",
	"ai/knowledge",
	"ai/memory/learnings",
	"ai/programs",
}

func TestScaffoldDirectories_CreatesAllSixDirectories(t *testing.T) {
	root := t.TempDir()

	got, err := scaffoldDirectories(root)
	if err != nil {
		t.Fatalf("scaffoldDirectories() unexpected error: %v", err)
	}

	if len(got) != len(wantScaffoldedDirs) {
		t.Fatalf("scaffoldDirectories() = %v, want %d entries", got, len(wantScaffoldedDirs))
	}
	for i, want := range wantScaffoldedDirs {
		if got[i] != want {
			t.Errorf("scaffoldDirectories()[%d] = %q, want %q (fixed order)", i, got[i], want)
		}
		info, statErr := os.Stat(filepath.Join(root, want))
		if statErr != nil {
			t.Fatalf("directory %q not created: %v", want, statErr)
		}
		if !info.IsDir() {
			t.Errorf("%q exists but is not a directory", want)
		}
	}
}

func TestScaffoldDirectories_WritesGitkeepOnlyInEmptyDirectories(t *testing.T) {
	root := t.TempDir()

	if _, err := scaffoldDirectories(root); err != nil {
		t.Fatalf("scaffoldDirectories() unexpected error: %v", err)
	}

	for _, dir := range wantGitkeepDirs {
		gitkeep := filepath.Join(root, dir, ".gitkeep")
		info, err := os.Stat(gitkeep)
		if err != nil {
			t.Errorf("%q: .gitkeep not found: %v", dir, err)
			continue
		}
		if info.Size() != 0 {
			t.Errorf("%q/.gitkeep size = %d, want 0 (zero-byte placeholder)", dir, info.Size())
		}
	}

	// "ai" and "ai/memory" are each the parent of another scaffolded
	// directory — by the time scaffolding completes they already
	// contain real subdirectories, so they correctly get no .gitkeep
	// of their own (it would be redundant clutter, not a bug).
	for _, dir := range []string{"ai", "ai/memory"} {
		if _, err := os.Stat(filepath.Join(root, dir, ".gitkeep")); !os.IsNotExist(err) {
			t.Errorf("%q unexpectedly has a .gitkeep — it already contains real subdirectories", dir)
		}
	}
}

func TestScaffoldDirectories_DoesNotDisturbExistingRealContent(t *testing.T) {
	root := t.TempDir()

	// Pre-populate one of the six directories with real content, as if
	// a Knowledge artifact already existed there.
	knowledgeDir := filepath.Join(root, "ai", "knowledge")
	if err := os.MkdirAll(knowledgeDir, 0o755); err != nil {
		t.Fatalf("seeding fixture: %v", err)
	}
	realFile := filepath.Join(knowledgeDir, "KNOW-001-real.md")
	if err := os.WriteFile(realFile, []byte("---\nid: KNOW-001\n---\nReal content.\n"), 0o644); err != nil {
		t.Fatalf("seeding fixture: %v", err)
	}

	if _, err := scaffoldDirectories(root); err != nil {
		t.Fatalf("scaffoldDirectories() unexpected error: %v", err)
	}

	data, err := os.ReadFile(realFile)
	if err != nil {
		t.Fatalf("real content missing after scaffoldDirectories(): %v", err)
	}
	if string(data) != "---\nid: KNOW-001\n---\nReal content.\n" {
		t.Errorf("real content was modified: %q", data)
	}
	if _, err := os.Stat(filepath.Join(knowledgeDir, ".gitkeep")); !os.IsNotExist(err) {
		t.Error("scaffoldDirectories() wrote .gitkeep into a non-empty directory, want none (FR-006)")
	}
}

func TestScaffoldDirectories_IsIdempotent(t *testing.T) {
	root := t.TempDir()

	first, err := scaffoldDirectories(root)
	if err != nil {
		t.Fatalf("first scaffoldDirectories() unexpected error: %v", err)
	}

	second, err := scaffoldDirectories(root)
	if err != nil {
		t.Fatalf("second scaffoldDirectories() unexpected error: %v", err)
	}

	if len(first) != len(second) {
		t.Fatalf("second call returned %d entries, want %d (same as first)", len(second), len(first))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Errorf("entry %d differs between calls: %q vs %q", i, first[i], second[i])
		}
	}

	for _, dir := range wantGitkeepDirs {
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			t.Fatalf("reading %q: %v", dir, err)
		}
		if len(entries) != 1 || entries[0].Name() != ".gitkeep" {
			t.Errorf("%q contains %v after two calls, want exactly one .gitkeep", dir, entries)
		}
	}
}
