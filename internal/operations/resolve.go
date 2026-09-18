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

// ErrSpecContextRequired is returned (wrapped inside a
// *SpecContextRequiredError) when a bare "TASK-NNN" reference is given
// with no owning-Spec context and more than one Spec claims that number
// project-wide (031-canonical-task-identity spec.md FR-003). This is
// deliberately distinct from ErrEntityAmbiguous: the two (or more) Specs
// each legitimately own their own TASK-NNN — this is not the same
// identity claimed twice, just an under-specified reference to it.
var ErrSpecContextRequired = errors.New("operations: bare Task reference requires Spec context")

// SpecContextRequiredError wraps ErrSpecContextRequired, naming the Task
// number and every candidate Spec that claims it, so a caller can retry
// with the composite "SPEC-###:TASK-###" form instead of parsing prose.
type SpecContextRequiredError struct {
	TaskNumber int
	Candidates []ids.EntityID
}

func (e *SpecContextRequiredError) Error() string {
	width := 3
	if len(e.Candidates) > 0 {
		width = e.Candidates[0].Width
	}
	taskRef := fmt.Sprintf("TASK-%0*d", width, e.TaskNumber)
	return fmt.Sprintf("operations: %s is claimed by more than one Spec (%v); resolve it as SPEC-###:%s instead", taskRef, e.Candidates, taskRef)
}

func (e *SpecContextRequiredError) Unwrap() error {
	return ErrSpecContextRequired
}

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

	if id.Type == ids.Task {
		return ResolveTask(root, cfg, rawID, nil)
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

// ResolveTask finds the one Task matching raw, which may be a composite
// reference ("SPEC-014:TASK-003") or a bare local reference
// ("TASK-003") — the shared resolver every Task-identity consumer uses
// (031-canonical-task-identity spec.md FR-002, FR-003,
// contracts/task-identity-resolution.md §2).
//
// A composite raw resolves directly against its named Spec. A bare raw
// resolves against specCtx if given; otherwise it consults the
// project-wide cross-Spec index: exactly one claiming Spec resolves
// unambiguously (spec FR-009, preserving today's convenient case), zero
// is ErrEntityNotFound, and more than one is ErrSpecContextRequired —
// never an arbitrary pick, and never conflated with a genuine same-Spec
// duplicate (ErrEntityAmbiguous).
func ResolveTask(root string, cfg project.Configuration, raw string, specCtx *ids.EntityID) (ResolvedLocation, error) {
	ref, err := ids.ParseTaskRef(raw, cfg)
	if err != nil {
		return ResolvedLocation{}, fmt.Errorf("%w: %v", ErrInvalidTarget, err)
	}

	// Composite form: raw already names its own Spec.
	if ref.Spec.Number != 0 {
		if specCtx != nil && *specCtx != ref.Spec {
			return ResolvedLocation{}, fmt.Errorf("%w: composite reference %q disagrees with given Spec context %v", ErrInvalidTarget, raw, *specCtx)
		}
		return resolveTaskInSpec(root, cfg, ref.Spec, ref.Local, raw)
	}

	// Bare form with an explicit Spec context supplied by the caller.
	if specCtx != nil {
		return resolveTaskInSpec(root, cfg, *specCtx, ref.Local, raw)
	}

	// Bare form, no context at all: consult the project-wide index.
	taskScan, err := ids.ScanTasks(root, cfg)
	if err != nil {
		return ResolvedLocation{}, err
	}
	bySpec := taskScan.ByNumber[ref.Local.Number]
	switch len(bySpec) {
	case 0:
		return ResolvedLocation{}, fmt.Errorf("%w: %s", ErrEntityNotFound, raw)
	case 1:
		for specNum, paths := range bySpec {
			return taskLocationFromPaths(ref.Local, cfg, specNum, paths, raw)
		}
		panic("unreachable") // len(bySpec) == 1 guarantees exactly one iteration
	default:
		specNums := make([]int, 0, len(bySpec))
		for s := range bySpec {
			specNums = append(specNums, s)
		}
		sort.Ints(specNums)
		candidates := make([]ids.EntityID, 0, len(specNums))
		for _, s := range specNums {
			candidates = append(candidates, ids.EntityID{Type: ids.Spec, Prefix: ids.Spec.Prefix(), Number: s, Width: cfg.IDWidth})
		}
		return ResolvedLocation{}, &SpecContextRequiredError{TaskNumber: ref.Local.Number, Candidates: candidates}
	}
}

// resolveTaskInSpec resolves local within spec's own tasks.md, scoped to
// that one Spec only — zero claims is ErrEntityNotFound (whether spec
// itself doesn't exist or simply has no such Task), more than one claim
// inside that same Spec is a genuine ErrEntityAmbiguous (spec FR-004).
func resolveTaskInSpec(root string, cfg project.Configuration, spec, local ids.EntityID, raw string) (ResolvedLocation, error) {
	taskScan, err := ids.ScanTasks(root, cfg)
	if err != nil {
		return ResolvedLocation{}, err
	}
	paths := taskScan.BySpec[spec.Number][local.Number]
	return taskLocationFromPaths(local, cfg, spec.Number, paths, raw)
}

// taskLocationFromPaths turns the paths claiming (specNum, local) into a
// ResolvedLocation, or the appropriate not-found/ambiguous error.
func taskLocationFromPaths(local ids.EntityID, cfg project.Configuration, specNum int, paths []string, raw string) (ResolvedLocation, error) {
	switch len(paths) {
	case 0:
		return ResolvedLocation{}, fmt.Errorf("%w: %s", ErrEntityNotFound, raw)
	case 1:
		return ResolvedLocation{ID: local, Type: artifacts.TypeTasks, Path: paths[0]}, nil
	default:
		sorted := append([]string(nil), paths...)
		sort.Strings(sorted)
		spec := ids.EntityID{Type: ids.Spec, Prefix: ids.Spec.Prefix(), Number: specNum, Width: cfg.IDWidth}
		return ResolvedLocation{}, &AmbiguousIDError{ID: fmt.Sprintf("%s:%s", spec, local), Locations: sorted}
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
