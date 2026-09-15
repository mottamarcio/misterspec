package operations

import (
	"errors"
	"fmt"
	"sort"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// ErrEntityNotFound is returned when no artifact anywhere in the project
// matches the requested ID (FR-002).
var ErrEntityNotFound = errors.New("operations: entity not found")

// ErrEntityAmbiguous is returned (wrapped inside a *AmbiguousIDError) when
// more than one artifact claims the same ID (FR-003).
var ErrEntityAmbiguous = errors.New("operations: entity ID is ambiguous")

// ErrInvalidTarget is returned when a raw ID string is not syntactically
// valid for any entity type, or does not match the project's configured
// ID width — distinct from ErrEntityNotFound, which means the syntax was
// fine but nothing matched.
var ErrInvalidTarget = errors.New("operations: invalid target ID")

// AmbiguousIDError wraps ErrEntityAmbiguous, naming every location that
// claimed the ambiguous ID (FR-003) — a bare sentinel would lose that
// data.
type AmbiguousIDError struct {
	ID        string
	Locations []string
}

func (e *AmbiguousIDError) Error() string {
	return fmt.Sprintf("operations: %q is ambiguous: claimed by %v", e.ID, e.Locations)
}

func (e *AmbiguousIDError) Unwrap() error {
	return ErrEntityAmbiguous
}

// ResolvedLocation is the outcome of resolving exactly one ID (FR-001).
type ResolvedLocation struct {
	ID   ids.EntityID
	Type artifacts.ArtifactType
	// Path is relative to the project root. For a Task, this is
	// "<tasks.md path>#TASK-NNN" — Task has no independent file of its
	// own (FR-016).
	Path string
}

// canonicalFilename returns the fixed filename for entity types whose
// canonical location is "a directory + a known filename" (Program,
// Feature, Spec). Knowledge and Learning are flat files already named by
// Scan; Task is handled separately (its "path" is already file+heading).
func canonicalFilename(t ids.EntityType) (string, bool) {
	switch t {
	case ids.Program:
		return "program.md", true
	case ids.Feature:
		return "feature.md", true
	case ids.Spec:
		return "spec.md", true
	default:
		return "", false
	}
}

// artifactTypeFor maps an ids.EntityType to the artifacts.ArtifactType of
// the file it resolves to. Task maps to TypeTasks: an individual Task's
// location is inside a Tasks artifact file, which is the closest
// ArtifactType describing where it lives.
func artifactTypeFor(t ids.EntityType) artifacts.ArtifactType {
	switch t {
	case ids.Program:
		return artifacts.TypeProgram
	case ids.Feature:
		return artifacts.TypeFeature
	case ids.Spec:
		return artifacts.TypeSpec
	case ids.Knowledge:
		return artifacts.TypeKnowledge
	case ids.Learning:
		return artifacts.TypeLearning
	case ids.Task:
		return artifacts.TypeTasks
	default:
		return artifacts.TypeConstitution // unreachable for a valid EntityType
	}
}

// Resolve finds the one artifact matching rawID, searching the project's
// structure — the agent must never guess or construct this path itself
// (FR-001, FR-002, FR-003).
func Resolve(root string, cfg project.Configuration, rawID string) (ResolvedLocation, error) {
	// First pass: learn the type from the prefix alone, tolerant of any
	// width, purely to classify what kind of ID this claims to be.
	loose, err := ids.ParseAny(rawID)
	if err != nil {
		return ResolvedLocation{}, fmt.Errorf("%w: %v", ErrInvalidTarget, err)
	}

	// Second pass: enforce the project's configured canonical width
	// strictly, the same syntax rule applied everywhere else in this
	// project — a wrong-width ID (e.g. "SPEC-14" when the project uses
	// 3-digit IDs) is invalid, not silently matched by numeric value.
	id, err := ids.Parse(loose.Type, rawID, cfg.IDWidth)
	if err != nil {
		return ResolvedLocation{}, fmt.Errorf("%w: %v", ErrInvalidTarget, err)
	}

	result, err := ids.Scan(root, cfg, id.Type)
	if err != nil {
		return ResolvedLocation{}, err
	}

	paths := result.Paths[id.Number]
	switch len(paths) {
	case 0:
		return ResolvedLocation{}, fmt.Errorf("%w: %s", ErrEntityNotFound, rawID)
	case 1:
		return ResolvedLocation{
			ID:   id,
			Type: artifactTypeFor(id.Type),
			Path: resolvedPath(id.Type, paths[0]),
		}, nil
	default:
		sorted := append([]string(nil), paths...)
		sort.Strings(sorted)
		return ResolvedLocation{}, &AmbiguousIDError{ID: rawID, Locations: sorted}
	}
}

// resolvedPath turns a Scan-reported location into the ResolvedLocation's
// final Path: for Program/Feature/Spec, Scan reports the entity's
// directory, so the fixed filename is appended; for Knowledge/Learning
// and Task, Scan already reports the exact file (or file#heading).
func resolvedPath(t ids.EntityType, scanPath string) string {
	if filename, ok := canonicalFilename(t); ok {
		return scanPath + "/" + filename
	}
	return scanPath
}
