package installer

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteAtomicFile_SuccessfulWrite(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "sub", "out.txt")
	content := []byte("hello, atomic world\n")

	if err := writeAtomicFile(target, content); err != nil {
		t.Fatalf("writeAtomicFile() unexpected error: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content = %q, want %q", got, content)
	}
}

func TestWriteAtomicFile_FailurePartwayLeavesNoFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("read-only directory permissions behave differently on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root ignores directory permission bits")
	}

	dir := t.TempDir()
	roDir := filepath.Join(dir, "readonly")
	if err := os.Mkdir(roDir, 0o500); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Cleanup(func() { os.Chmod(roDir, 0o755) }) // allow t.TempDir() cleanup

	target := filepath.Join(roDir, "out.txt")
	err := writeAtomicFile(target, []byte("should never land"))
	if err == nil {
		t.Fatal("writeAtomicFile() into a read-only directory: expected an error, got nil")
	}

	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Errorf("target file unexpectedly exists after a failed write: stat err = %v", statErr)
	}

	entries, err := os.ReadDir(roDir)
	if err == nil && len(entries) != 0 {
		t.Errorf("read-only dir unexpectedly contains entries after a failed write: %v", entries)
	}
}
