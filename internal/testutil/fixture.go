package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// Project builds a temporary, valid misterspec project tree under
// t.TempDir() with a .misterspec/config.yaml (using the frozen defaults
// from docs/architecture-specification.md §21, with a 3-digit ID width),
// and returns the project's root directory. Use WriteFile and WriteConfig
// to customize it further before running the code under test.
func Project(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	WriteConfig(t, root, DefaultConfigYAML)
	return root
}

// DefaultConfigYAML is a well-formed .misterspec/config.yaml matching the
// schema in docs/architecture-specification.md §21.
const DefaultConfigYAML = `schema_version: 1
agent_id: claude-code
artifacts_dir: ai
raw_dir: ai/raw
knowledge_dir: ai/knowledge
constitution_path: ai/memory/constitution.md
learnings_dir: ai/memory/learnings
programs_root: ai/programs
id_width: 3
`

// WriteConfig writes contents as root's .misterspec/config.yaml, creating
// the .misterspec directory if needed. Passing an empty or malformed
// contents string is a valid way to build an invalid-configuration
// fixture.
func WriteConfig(t *testing.T, root, contents string) {
	t.Helper()

	dir := filepath.Join(root, ".misterspec")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("testutil: creating .misterspec dir: %v", err)
	}
	WriteFile(t, root, filepath.Join(".misterspec", "config.yaml"), contents)
}

// WriteFile writes contents to relPath under root, creating any
// intermediate directories. relPath must be relative (e.g.
// "ai/programs/PRG-001/program.md").
func WriteFile(t *testing.T, root, relPath, contents string) string {
	t.Helper()

	full := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("testutil: creating parent dirs for %s: %v", relPath, err)
	}
	if err := os.WriteFile(full, []byte(contents), 0o644); err != nil {
		t.Fatalf("testutil: writing %s: %v", relPath, err)
	}
	return full
}

// Subdir creates and returns the absolute path of a directory nested
// depth levels below root, useful for exercising project detection from a
// nested working directory (e.g. depth=3 for
// root/a/b/c).
func Subdir(t *testing.T, root string, depth int) string {
	t.Helper()

	dir := root
	for i := 0; i < depth; i++ {
		dir = filepath.Join(dir, "nested")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("testutil: creating nested dir: %v", err)
	}
	return dir
}
