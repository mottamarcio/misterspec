package artifacts

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// CanonicalPath is the single, deterministic filesystem location an
// entity's artifact must occupy, as paths relative to the project root
// (using forward slashes, regardless of host OS).
type CanonicalPath struct {
	// Directory is the entity's canonical directory.
	Directory string
	// File is the entity's canonical file within Directory. It is empty
	// for entity types whose filename includes an agent-chosen slug
	// (Knowledge, Learning) — see ResolvePath's documentation.
	File string
}

// ErrPathOutsideProject is returned instead of any path — computed or
// supplied — that would resolve outside the project root (FR-008).
var ErrPathOutsideProject = errors.New("artifacts: path resolves outside project root")

// ResolvePath computes the canonical directory (and, where deterministic,
// file) for an entity of type t with id, given its ancestry in parents,
// ordered from outermost to innermost. Feature requires parents[0] = its
// Program's ID; Spec requires parents[0] = its Program's ID and
// parents[1] = its Feature's ID — the full ancestor chain is required
// because the canonical layout nests Feature and Spec directories under
// their Program and Feature respectively (docs/architecture-specification.md
// §20), and ResolvePath is a pure function of its arguments: it never
// reads the filesystem to discover an ancestor on its own (FR-007,
// SC-003).
//
// For Knowledge and Learning, whose canonical filename includes an
// agent-chosen slug (docs/architecture-specification.md §61) that
// ResolvePath is not given, only Directory is populated; discovering the
// exact filename for an existing artifact is a directory-scan operation
// (ids.Scan / a future resolve operation), not pure computation.
//
// ResolvePath rejects (ErrPathOutsideProject) before ever constructing a
// path that would escape root (FR-008, SC-004).
func ResolvePath(root string, cfg project.Configuration, t ids.EntityType, id ids.EntityID, parents ...ids.EntityID) (CanonicalPath, error) {
	switch t {
	case ids.Program:
		rel := filepath.Join(cfg.ProgramsRoot, id.String())
		return resolvedFile(root, rel, "program.md")

	case ids.Feature:
		if len(parents) < 1 {
			return CanonicalPath{}, fmt.Errorf("artifacts: resolving a feature path requires its parent program ID")
		}
		rel := filepath.Join(cfg.ProgramsRoot, parents[0].String(), "features", id.String())
		return resolvedFile(root, rel, "feature.md")

	case ids.Spec:
		if len(parents) < 2 {
			return CanonicalPath{}, fmt.Errorf("artifacts: resolving a spec path requires its parent program and feature IDs")
		}
		rel := filepath.Join(cfg.ProgramsRoot, parents[0].String(), "features", parents[1].String(), "specs", id.String())
		return resolvedFile(root, rel, "spec.md")

	case ids.Knowledge:
		return resolvedDirOnly(root, cfg.KnowledgeDir)

	case ids.Learning:
		return resolvedDirOnly(root, cfg.LearningsDir)

	default:
		return CanonicalPath{}, fmt.Errorf("artifacts: unsupported entity type %v", t)
	}
}

func resolvedFile(root, relDir, filename string) (CanonicalPath, error) {
	dir, err := RelativeWithinRoot(root, relDir)
	if err != nil {
		return CanonicalPath{}, err
	}
	dir = filepath.ToSlash(dir)
	return CanonicalPath{Directory: dir, File: dir + "/" + filename}, nil
}

func resolvedDirOnly(root, relDir string) (CanonicalPath, error) {
	dir, err := RelativeWithinRoot(root, relDir)
	if err != nil {
		return CanonicalPath{}, err
	}
	return CanonicalPath{Directory: filepath.ToSlash(dir)}, nil
}

// RelativeWithinRoot resolves path (absolute, or relative to root) to a
// path relative to root, rejecting it with ErrPathOutsideProject if the
// resolved absolute location does not lie inside root. The rejection
// happens before the caller ever uses the resolved path for anything else
// (FR-008). Exported in 002-read-operations for reuse by
// internal/operations's Inventory and Fingerprint, which need the exact
// same containment guarantee ResolvePath and ClassifyPath already rely on
// here (Constitution Principle VI, DRY; Principle VIII, one proven
// implementation instead of a second that could diverge).
func RelativeWithinRoot(root, path string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absRoot = filepath.Clean(absRoot)

	var absPath string
	if filepath.IsAbs(path) {
		absPath = filepath.Clean(path)
	} else {
		absPath = filepath.Clean(filepath.Join(absRoot, path))
	}

	if absPath != absRoot && !strings.HasPrefix(absPath, absRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %s", ErrPathOutsideProject, path)
	}

	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return "", err
	}
	return rel, nil
}
