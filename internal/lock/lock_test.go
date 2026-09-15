package lock_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mottamarcio/misterspec/internal/lock"
)

func TestAcquireRelease_Basic(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".misterspec"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	l, err := lock.Acquire(root, time.Second, time.Second)
	if err != nil {
		t.Fatalf("Acquire() unexpected error: %v", err)
	}

	lockPath := filepath.Join(root, ".misterspec", ".lock")
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("lock file not created: %v", err)
	}

	if err := l.Release(); err != nil {
		t.Fatalf("Release() unexpected error: %v", err)
	}
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("lock file still exists after Release(): err=%v", err)
	}
}

func TestRelease_SafeToCallTwice(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".misterspec"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	l, err := lock.Acquire(root, time.Second, time.Second)
	if err != nil {
		t.Fatalf("Acquire() unexpected error: %v", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("first Release() unexpected error: %v", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("second Release() unexpected error: %v", err)
	}
}

func TestAcquire_SecondCallBlocksUntilFirstReleases(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".misterspec"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	first, err := lock.Acquire(root, 5*time.Second, time.Second)
	if err != nil {
		t.Fatalf("first Acquire() unexpected error: %v", err)
	}

	released := make(chan struct{})
	go func() {
		time.Sleep(100 * time.Millisecond)
		first.Release()
		close(released)
	}()

	start := time.Now()
	second, err := lock.Acquire(root, 5*time.Second, 2*time.Second)
	if err != nil {
		t.Fatalf("second Acquire() unexpected error: %v", err)
	}
	defer second.Release()

	<-released
	if elapsed := time.Since(start); elapsed < 50*time.Millisecond {
		t.Errorf("second Acquire() returned suspiciously fast (%v) — expected it to wait for the first Release()", elapsed)
	}
}

func TestAcquire_TimeoutWhenActivelyHeld(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".misterspec"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	first, err := lock.Acquire(root, 10*time.Second, time.Second)
	if err != nil {
		t.Fatalf("first Acquire() unexpected error: %v", err)
	}
	defer first.Release()

	_, err = lock.Acquire(root, 10*time.Second, 150*time.Millisecond)
	if !errors.Is(err, lock.ErrLockTimeout) {
		t.Fatalf("second Acquire() error = %v, want errors.Is(err, ErrLockTimeout)", err)
	}
}

func TestAcquire_StaleLockRecoveredAutomatically(t *testing.T) {
	root := t.TempDir()
	lockDir := filepath.Join(root, ".misterspec")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	lockPath := filepath.Join(lockDir, ".lock")
	if err := os.WriteFile(lockPath, []byte("leftover from a crashed process\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	// Backdate the lock file well past a short staleAfter threshold.
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(lockPath, old, old); err != nil {
		t.Fatalf("setup: %v", err)
	}

	start := time.Now()
	l, err := lock.Acquire(root, 200*time.Millisecond, 2*time.Second)
	if err != nil {
		t.Fatalf("Acquire() unexpected error: %v", err)
	}
	defer l.Release()

	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("Acquire() took %v to recover a stale lock — expected prompt recovery, not waiting out the full timeout", elapsed)
	}
}
