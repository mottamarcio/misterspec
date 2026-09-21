package eval

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTasks_RejectsMissingAcceptanceTest(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture")
	if err := os.MkdirAll(fixture, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCaseFile(t, dir, "task1.yaml", `
id: t1
description: do a thing
dir: fixture
`)

	_, err := LoadTasks(dir)
	if !errors.Is(err, ErrInvalidCase) {
		t.Fatalf("LoadTasks() error = %v, want ErrInvalidCase", err)
	}
}

func TestLoadTasks_RejectsUnresolvableDir(t *testing.T) {
	dir := t.TempDir()
	writeCaseFile(t, dir, "task1.yaml", `
id: t1
description: do a thing
dir: missing
acceptance_test: go test ./...
`)

	_, err := LoadTasks(dir)
	if !errors.Is(err, ErrInvalidCase) {
		t.Fatalf("LoadTasks() error = %v, want ErrInvalidCase", err)
	}
}

func TestLoadTasks_RejectsDuplicateID(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture")
	if err := os.MkdirAll(fixture, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `
id: dup
description: do a thing
dir: fixture
acceptance_test: go test ./...
`
	writeCaseFile(t, dir, "task1.yaml", body)
	writeCaseFile(t, dir, "task2.yaml", body)

	_, err := LoadTasks(dir)
	if !errors.Is(err, ErrInvalidCase) {
		t.Fatalf("LoadTasks() error = %v, want ErrInvalidCase (duplicate id)", err)
	}
}

func TestLoadTasks_ValidTaskLoads(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture")
	if err := os.MkdirAll(fixture, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCaseFile(t, dir, "task1.yaml", `
id: t1
description: do a thing
dir: fixture
acceptance_test: go test ./...
`)

	tasks, err := LoadTasks(dir)
	if err != nil {
		t.Fatalf("LoadTasks() error = %v, want nil", err)
	}
	if len(tasks) != 1 || tasks[0].ID != "t1" {
		t.Fatalf("LoadTasks() = %+v, want one task t1", tasks)
	}
}
