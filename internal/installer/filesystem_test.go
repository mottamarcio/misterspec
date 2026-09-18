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

	if err := WriteAtomicFile(target, content, DefaultFileMode); err != nil {
		t.Fatalf("WriteAtomicFile() unexpected error: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content = %q, want %q", got, content)
	}
}

func TestWriteAtomicFile_RequestedModeApplied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not meaningful on Windows")
	}

	dir := t.TempDir()
	target := filepath.Join(dir, "binary")

	if err := WriteAtomicFile(target, []byte("bin"), 0o755); err != nil {
		t.Fatalf("WriteAtomicFile() unexpected error: %v", err)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat written file: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("mode = %v, want %v", info.Mode().Perm(), os.FileMode(0o755))
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
	err := WriteAtomicFile(target, []byte("should never land"), DefaultFileMode)
	if err == nil {
		t.Fatal("WriteAtomicFile() into a read-only directory: expected an error, got nil")
	}

	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Errorf("target file unexpectedly exists after a failed write: stat err = %v", statErr)
	}

	entries, err := os.ReadDir(roDir)
	if err == nil && len(entries) != 0 {
		t.Errorf("read-only dir unexpectedly contains entries after a failed write: %v", entries)
	}
}
