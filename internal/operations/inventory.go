package operations

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/mottamarcio/misterspec/internal/artifacts"
)

// FileEntry is one file found while inventorying a directory (FR-010).
type FileEntry struct {
	// Path is relative to the project root.
	Path string
	// Extension includes the leading dot (e.g. ".md"), or "" for an
	// extensionless file.
	Extension string
	// Size is the file's size in bytes.
	Size int64
}

// Inventory lists every file reachable under dir (relative to root, or an
// absolute path within it — e.g. cfg.RawDir, cfg.KnowledgeDir, or a
// ResolvedLocation's own directory), reporting each one's path, extension,
// and size (FR-010). It returns an empty, non-nil-error slice — never an
// error — for a directory that is empty or does not exist yet, as long as
// its location is itself valid (FR-011). It rejects a dir that would
// resolve outside root (FR-013), reusing artifacts.RelativeWithinRoot
// rather than a second containment check.
func Inventory(root, dir string) ([]FileEntry, error) {
	rel, err := artifacts.RelativeWithinRoot(root, dir)
	if err != nil {
		return nil, err
	}
	full := filepath.Join(root, rel)

	info, err := os.Stat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return []FileEntry{{
			Path:      filepath.ToSlash(rel),
			Extension: filepath.Ext(full),
			Size:      info.Size(),
		}}, nil
	}

	var files []FileEntry
	walkErr := filepath.WalkDir(full, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		entryInfo, err := d.Info()
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, FileEntry{
			Path:      filepath.ToSlash(relPath),
			Extension: filepath.Ext(path),
			Size:      entryInfo.Size(),
		})
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}
