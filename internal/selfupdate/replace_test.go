package selfupdate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplace_Unix(t *testing.T) {
	dir := t.TempDir()
	exePath := filepath.Join(dir, "misterspec")
	if err := os.WriteFile(exePath, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	tmpPath := filepath.Join(dir, ".tmp-new")
	if err := os.WriteFile(tmpPath, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Replace(tmpPath, exePath, "linux"); err != nil {
		t.Fatalf("Replace() error = %v", err)
	}

	data, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Errorf("exePath content = %q, want %q", data, "new")
	}
	info, err := os.Stat(exePath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Errorf("exePath mode = %v, want executable", info.Mode())
	}
}

func TestReplace_WindowsRenameAside(t *testing.T) {
	dir := t.TempDir()
	exePath := filepath.Join(dir, "misterspec.exe")
	if err := os.WriteFile(exePath, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	tmpPath := filepath.Join(dir, ".tmp-new")
	if err := os.WriteFile(tmpPath, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Replace(tmpPath, exePath, "windows"); err != nil {
		t.Fatalf("Replace() error = %v", err)
	}

	data, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Errorf("exePath content = %q, want %q", data, "new")
	}

	oldPath := exePath + ".old"
	if oldData, err := os.ReadFile(oldPath); err == nil && string(oldData) != "old" {
		t.Errorf("%s content = %q, want %q", oldPath, oldData, "old")
	}
}
