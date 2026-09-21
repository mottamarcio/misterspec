package eval

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeCaseFile(t *testing.T, dir, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
		t.Fatalf("writing case file: %v", err)
	}
}

func TestLoadCases_RejectsQueryAndTaskBothSet(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture")
	if err := os.MkdirAll(fixture, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCaseFile(t, dir, "case1.yaml", `
id: c1
target: SPEC-001
query: hello
task: do something
dir: fixture
required: [SPEC-001]
`)

	_, err := LoadCases(dir)
	if !errors.Is(err, ErrInvalidCase) {
		t.Fatalf("LoadCases() error = %v, want ErrInvalidCase", err)
	}
}

func TestLoadCases_RejectsDuplicateID(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture")
	if err := os.MkdirAll(fixture, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `
id: dup
target: SPEC-001
dir: fixture
required: [SPEC-001]
`
	writeCaseFile(t, dir, "case1.yaml", body)
	writeCaseFile(t, dir, "case2.yaml", body)

	_, err := LoadCases(dir)
	if !errors.Is(err, ErrInvalidCase) {
		t.Fatalf("LoadCases() error = %v, want ErrInvalidCase (duplicate id)", err)
	}
}

func TestLoadCases_RejectsUnresolvableDir(t *testing.T) {
	dir := t.TempDir()
	writeCaseFile(t, dir, "case1.yaml", `
id: c1
target: SPEC-001
dir: does-not-exist
required: [SPEC-001]
`)

	_, err := LoadCases(dir)
	if !errors.Is(err, ErrInvalidCase) {
		t.Fatalf("LoadCases() error = %v, want ErrInvalidCase (unresolvable dir)", err)
	}
}

func TestLoadCases_RejectsEmptyRequiredWithoutNote(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture")
	if err := os.MkdirAll(fixture, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCaseFile(t, dir, "case1.yaml", `
id: c1
target: SPEC-001
dir: fixture
`)

	_, err := LoadCases(dir)
	if !errors.Is(err, ErrInvalidCase) {
		t.Fatalf("LoadCases() error = %v, want ErrInvalidCase (empty required, no note)", err)
	}
}

func TestLoadCases_AllowsEmptyRequiredWithNote(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture")
	if err := os.MkdirAll(fixture, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCaseFile(t, dir, "case1.yaml", `
id: c1
target: SPEC-001
dir: fixture
note: deliberately no required content
`)

	cases, err := LoadCases(dir)
	if err != nil {
		t.Fatalf("LoadCases() error = %v, want nil", err)
	}
	if len(cases) != 1 {
		t.Fatalf("LoadCases() returned %d cases, want 1", len(cases))
	}
}

func TestLoadCases_RejectsMissingTarget(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture")
	if err := os.MkdirAll(fixture, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCaseFile(t, dir, "case1.yaml", `
id: c1
dir: fixture
required: [SPEC-001]
`)

	_, err := LoadCases(dir)
	if !errors.Is(err, ErrInvalidCase) {
		t.Fatalf("LoadCases() error = %v, want ErrInvalidCase (missing target)", err)
	}
}

func TestLoadCases_ValidSetLoadsInSortedOrder(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture")
	if err := os.MkdirAll(fixture, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCaseFile(t, dir, "b.yaml", `
id: b
target: SPEC-002
dir: fixture
required: [SPEC-002]
`)
	writeCaseFile(t, dir, "a.yaml", `
id: a
target: SPEC-001
dir: fixture
required: [SPEC-001]
`)

	cases, err := LoadCases(dir)
	if err != nil {
		t.Fatalf("LoadCases() error = %v, want nil", err)
	}
	if len(cases) != 2 || cases[0].ID != "a" || cases[1].ID != "b" {
		t.Fatalf("LoadCases() = %+v, want [a, b] in that order", cases)
	}
	if cases[0].ResolvedDir != fixture {
		t.Errorf("ResolvedDir = %q, want %q", cases[0].ResolvedDir, fixture)
	}
}
