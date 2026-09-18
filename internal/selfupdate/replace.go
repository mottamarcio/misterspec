package selfupdate

import "os"

// Replace atomically swaps the binary at exePath for the already-verified
// content at tmpPath, given goos (an explicit parameter — never
// runtime.GOOS read internally — so tests can exercise the Windows path
// without a Windows runner; research.md).
//
// On every platform except Windows, a single os.Rename repoints the
// directory entry — safe even while the old binary is still executing,
// since the running process keeps its already-open reference to the
// old, now-unlinked inode until it exits.
//
// On Windows, the OS holds an exclusive lock on a running .exe and
// refuses to overwrite it directly, but permits renaming it aside, so
// the sequence is: rename the current exe to "<name>.old" first
// (best-effort removal of any stale .old left by a prior run), then
// rename the verified temp file into the original path.
func Replace(tmpPath, exePath, goos string) error {
	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return err
	}

	if goos == "windows" {
		oldPath := exePath + ".old"
		_ = os.Remove(oldPath) // best-effort; a prior run may have left this
		if err := os.Rename(exePath, oldPath); err != nil {
			return err
		}
	}

	return os.Rename(tmpPath, exePath)
}
