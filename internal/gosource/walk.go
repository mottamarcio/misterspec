package gosource

import (
	"io/fs"
	"os"
	"path/filepath"
)

// WalkGoFiles finds every ".go" file under root not matching one of
// exclusions' own glob patterns (044-architecture-code-context-rules
// research.md #4), as project-root-relative slash-separated paths in
// deterministic filesystem-walk order. Shared by internal/context/index
// (path exclusions) and internal/architecture (rule From/To patterns)
// rather than duplicated in each (Constitution Principle VI). A
// missing root is not an error — it simply has no files.
func WalkGoFiles(root string, exclusions []string) ([]string, error) {
	var found []string
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if MatchAnyGlob(exclusions, rel) {
			return nil
		}
		found = append(found, rel)
		return nil
	})
	if walkErr != nil {
		if os.IsNotExist(walkErr) {
			return nil, nil
		}
		return nil, walkErr
	}
	return found, nil
}
