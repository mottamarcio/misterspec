package gosource

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixture(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing fixture %s: %v", name, err)
	}
	return path
}

// TestImports_MultipleImports is 044-architecture-code-context-rules
// T003 (Foundational): Imports returns every import path and its own
// file-absolute line (contracts §1).
func TestImports_MultipleImports(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "multi.go", `package foo

import (
	"fmt"
	"os"

	"github.com/mottamarcio/misterspec/internal/ids"
)

func Bar() {}
`)

	got, err := Imports(path)
	if err != nil {
		t.Fatalf("Imports() unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("Imports() = %+v, want 3 entries", got)
	}

	want := map[string]int{
		"fmt": 4,
		"os":  5,
		"github.com/mottamarcio/misterspec/internal/ids": 7,
	}
	for _, ref := range got {
		wantLine, ok := want[ref.Path]
		if !ok {
			t.Errorf("unexpected import %q in %+v", ref.Path, got)
			continue
		}
		if ref.Line != wantLine {
			t.Errorf("import %q line = %d, want %d", ref.Path, ref.Line, wantLine)
		}
	}
}

// TestImports_NoImports is 044-architecture-code-context-rules T003:
// a file with no imports returns an empty, non-nil slice.
func TestImports_NoImports(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "noimports.go", "package foo\n\nfunc Bar() {}\n")

	got, err := Imports(path)
	if err != nil {
		t.Fatalf("Imports() unexpected error: %v", err)
	}
	if got == nil {
		t.Error("Imports() = nil, want an empty non-nil slice")
	}
	if len(got) != 0 {
		t.Errorf("Imports() = %+v, want none", got)
	}
}

// TestImports_InvalidSyntax is 044-architecture-code-context-rules
// T003: a syntactically invalid .go file returns an error.
func TestImports_InvalidSyntax(t *testing.T) {
	dir := t.TempDir()
	// ImportsOnly mode stops parsing after the import block, so a
	// broken function body downstream would not be caught — the
	// import block itself must be malformed to exercise this path.
	path := writeFixture(t, dir, "broken.go", "package foo\n\nimport (\n\t\"fmt\n)\n")

	if _, err := Imports(path); err == nil {
		t.Error("Imports() error = nil for a syntactically invalid import block, want non-nil")
	}
}
