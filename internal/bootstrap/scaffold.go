package bootstrap

import (
	"os"
	"path/filepath"

	"github.com/mottamarcio/misterspec/internal/installer"
	"github.com/mottamarcio/misterspec/internal/project"
)

// gitkeepFilename is the placeholder file written into an otherwise-
// empty scaffolded directory (021-init-scaffold-distribution's own
// research.md #3) — a literal, zero-byte file, the user's own explicit
// naming choice, so Git (which never tracks empty directories) has
// something real to commit.
const gitkeepFilename = ".gitkeep"

// scaffoldedDirs is every directory project.Configuration's own
// Default* constants name, relative to a project root, in a fixed,
// deterministic order (data-model.md) — everything the project's own
// configuration considers a meaningful location, beyond whichever
// agent's own Skills directory Bootstrap's own adapter.Install call
// handles separately.
var scaffoldedDirs = []string{
	project.DefaultArtifactsDir,
	project.DefaultRawDir,
	project.DefaultKnowledgeDir,
	filepath.ToSlash(filepath.Dir(project.DefaultConstitutionPath)),
	project.DefaultLearningsDir,
	project.DefaultProgramsRoot,
}

// scaffoldDirectories creates every directory scaffoldedDirs names,
// relative to targetDir, via os.MkdirAll — idempotent, so a directory
// that already exists (from a prior run) is left untouched. All six
// are created first, in a first pass, before any is checked for
// emptiness — some of them are one another's own parent (e.g. "ai" is
// the parent of "ai/raw"), so checking emptiness only after every
// directory exists avoids writing a now-redundant .gitkeep into a
// directory that in fact already contains real, meaningful
// subdirectories of its own. Only a directory that is still genuinely
// empty after that first pass gets a zero-byte .gitkeep
// (installer.WriteAtomicFile); a directory that already contains real
// content — its own or a sibling's, from an earlier run — is never
// disturbed and never gets a .gitkeep (FR-006). Returns
// scaffoldedDirs' own relative paths, in the same fixed order,
// regardless of which ones were newly created versus already present.
func scaffoldDirectories(targetDir string) ([]string, error) {
	for _, dir := range scaffoldedDirs {
		if err := os.MkdirAll(filepath.Join(targetDir, filepath.FromSlash(dir)), 0o755); err != nil {
			return nil, err
		}
	}

	for _, dir := range scaffoldedDirs {
		abs := filepath.Join(targetDir, filepath.FromSlash(dir))
		entries, err := os.ReadDir(abs)
		if err != nil {
			return nil, err
		}
		if len(entries) != 0 {
			continue
		}

		if err := installer.WriteAtomicFile(filepath.Join(abs, gitkeepFilename), nil, installer.DefaultFileMode); err != nil {
			return nil, err
		}
	}

	return append([]string(nil), scaffoldedDirs...), nil
}
