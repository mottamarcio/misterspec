package lock

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DefaultStaleAfter is how old an existing lock file must be before
// Acquire treats it as abandoned by an interrupted prior attempt and
// recovers it automatically (FR-006).
const DefaultStaleAfter = 10 * time.Second

// DefaultTimeout is the longest Acquire will wait, retrying, before
// giving up with ErrLockTimeout.
const DefaultTimeout = 5 * time.Second

// pollInterval is how often Acquire retries while waiting for a
// currently-held (non-stale) lock to be released.
const pollInterval = 20 * time.Millisecond

// ErrLockTimeout is returned by Acquire when the project lock is
// genuinely held by another in-progress creation and is not recovered
// within timeout — distinct from a stale lock, which is recovered
// automatically rather than reported as an error.
var ErrLockTimeout = errors.New("lock: could not acquire within timeout")

// lockFilePath is the fixed location of the project lock, relative to
// root, per docs/architecture-specification.md §58.
const lockFilePath = ".misterspec/.lock"

// Lock is a held project lock. It holds no semantic project state
// (Constitution Principle III) — it exists only to serialize the
// deterministic operations that write to the filesystem.
type Lock struct {
	path     string
	released bool
}

// Acquire obtains root's project lock, creating it via O_CREATE|O_EXCL —
// atomic and portable across Linux, macOS, and Windows with no
// OS-specific syscalls. If an existing lock file is older than
// staleAfter, it is treated as abandoned by an interrupted prior
// creation attempt and removed, then acquisition is retried — recovering
// automatically, with no manual intervention (FR-006). Otherwise Acquire
// polls until the lock is released or until timeout elapses, at which
// point it returns ErrLockTimeout.
func Acquire(root string, staleAfter, timeout time.Duration) (*Lock, error) {
	path := filepath.Join(root, lockFilePath)
	deadline := time.Now().Add(timeout)

	for {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			fmt.Fprintf(f, "acquired: %s\npid: %d\n", time.Now().Format(time.RFC3339), os.Getpid())
			if closeErr := f.Close(); closeErr != nil {
				return nil, closeErr
			}
			return &Lock{path: path}, nil
		}
		if !os.IsExist(err) {
			return nil, err
		}

		// The lock file exists. If it's stale, recover it and retry
		// immediately — this is not "waiting," so it doesn't count
		// against timeout.
		if info, statErr := os.Stat(path); statErr == nil {
			if time.Since(info.ModTime()) > staleAfter {
				// Best-effort removal: if another process wins this
				// race and removes it first (or recreates it), our
				// next O_EXCL attempt is still the sole source of
				// truth for who acquires the lock.
				os.Remove(path)
				continue
			}
		} else if !os.IsNotExist(statErr) {
			return nil, statErr
		}

		if time.Now().After(deadline) {
			return nil, ErrLockTimeout
		}
		time.Sleep(pollInterval)
	}
}

// Release removes the lock file. Safe to call more than once — a second
// call is a no-op, not an error.
func (l *Lock) Release() error {
	if l == nil || l.released {
		return nil
	}
	l.released = true

	if err := os.Remove(l.path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
