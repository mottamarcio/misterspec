package impact

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/validation"
	"github.com/mottamarcio/misterspec/internal/vcs"
)

// ChangeStatus is one ChangedElement's own kind of change (data-model.md
// "ChangeStatus").
type ChangeStatus string

const (
	Added    ChangeStatus = "added"
	Modified ChangeStatus = "modified"
	Removed  ChangeStatus = "removed"
)

// ChangedElement is one artifact- or Requirement-scoped element found
// different between two revisions (data-model.md "ChangedElement").
type ChangedElement struct {
	ID     ids.EntityID
	Path   string
	Status ChangeStatus
	// RequirementNumber is non-nil only for a Requirement-level entry
	// inside a modified/added/removed Spec (research.md #3); nil for a
	// whole-artifact entry.
	RequirementNumber *int
}

// ChangeSet is the full result of comparing two revisions (data-model.md
// "ChangeSet").
type ChangeSet struct {
	From              string
	To                string
	Elements          []ChangedElement
	UnmappedCodePaths int
}

// flatFileIDPattern matches a Knowledge/Learning flat file's own
// filename convention (e.g. "KNOW-003-authentication.md"), mirroring
// internal/ids/scan.go's own scanFlatFiles filePattern shape.
var flatFileIDPattern = regexp.MustCompile(`^([A-Z]+)-(\d+)(?:[-.].*)?$`)

// dirIDPattern matches a canonical entity directory name exactly (e.g.
// "SPEC-014"), mirroring internal/ids/scan.go's own unexported
// idDirPattern — duplicated rather than imported since ids does not
// export it (research.md's additive-export precedent extended here at
// implementation time: this feature needs to classify a diff path by
// name alone, without requiring the file to still exist on disk, which
// rules out reusing ids.Scan/artifacts.ParseMetadata directly for a
// removed path).
var dirIDPattern = regexp.MustCompile(`^[A-Z]+-\d+$`)

// entityIDForDiffPath classifies path (as reported by vcs.DiffNameStatus,
// relative to the project root) into the EntityID that owns it, purely
// from its own name — no filesystem access, so it works identically for
// an added, modified, or removed path. Two conventions are recognized,
// matching the same shapes internal/ids/scan.go's scanDirs/scanFlatFiles
// already enumerate: (1) a flat Knowledge/Learning file, whose own
// filename starts with "PREFIX-NNN"; (2) a directory-owned file
// (Program's "program.md", Feature's "feature.md", Spec's "spec.md",
// and — research.md #8 — a Spec's own "plan.md"/"tasks.md", all four
// attributed to the nearest ancestor directory named "PREFIX-NNN"). A
// path matching neither convention (e.g. Go source, or any file outside
// the five entity types' own canonical locations) returns ok == false.
func entityIDForDiffPath(path string) (id ids.EntityID, ok bool) {
	base := filepath.Base(path)
	if sub := flatFileIDPattern.FindStringSubmatch(base); sub != nil {
		if parsed, err := ids.ParseAny(sub[1] + "-" + sub[2]); err == nil {
			if parsed.Type == ids.Knowledge || parsed.Type == ids.Learning {
				return parsed, true
			}
		}
	}

	dir := filepath.Dir(path)
	for dir != "." && dir != "/" && dir != "" {
		seg := filepath.Base(dir)
		if dirIDPattern.MatchString(seg) {
			if parsed, err := ids.ParseAny(seg); err == nil {
				return parsed, true
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ids.EntityID{}, false
}

// changeStatusFor maps git diff --name-status's own letter to a
// ChangeStatus.
func changeStatusFor(gitStatus string) (ChangeStatus, bool) {
	switch gitStatus {
	case "A":
		return Added, true
	case "M":
		return Modified, true
	case "D":
		return Removed, true
	default:
		return "", false
	}
}

// BuildChangeSet computes the ChangeSet between from and to (to == ""
// meaning the working tree), scoped to scopePath when non-empty
// (contracts §3 "AnalyzeImpactRequest.Path"). A changed spec.md whose
// ChangeStatus is Modified is additionally expanded into one
// ChangedElement per Requirement number whose own Section.Body differs
// between the two revisions (research.md #3) — content is compared
// directly rather than via a separate hash, since the two full
// Section.Body strings are already in memory and a hash adds no
// distinguishing power for this in-process comparison.
func BuildChangeSet(root string, cfg project.Configuration, from, to, scopePath string) (ChangeSet, error) {
	if !vcs.IsRepo(root) {
		return ChangeSet{}, ErrNotARepository
	}

	toLabel := to
	if toLabel == "" {
		toLabel = "working-tree"
	}

	entries, err := vcs.DiffNameStatus(root, from, to)
	if err != nil {
		return ChangeSet{}, ErrRevisionNotFound
	}

	cs := ChangeSet{From: from, To: toLabel}
	for _, entry := range entries {
		if scopePath != "" && entry.Path != scopePath {
			continue
		}

		status, ok := changeStatusFor(entry.Status)
		if !ok {
			continue
		}

		id, matched := entityIDForDiffPath(entry.Path)
		if !matched {
			cs.UnmappedCodePaths++
			continue
		}

		cs.Elements = append(cs.Elements, ChangedElement{ID: id, Path: entry.Path, Status: status})

		if id.Type == ids.Spec && filepath.Base(entry.Path) == "spec.md" && status == Modified {
			reqElements, err := changedRequirements(root, id, entry.Path, from, to)
			if err != nil {
				return ChangeSet{}, err
			}
			cs.Elements = append(cs.Elements, reqElements...)
		}
	}

	sortChangedElements(cs.Elements)
	return cs, nil
}

// changedRequirements compares specPath's own Requirement sections
// between from and to, returning one ChangedElement per Requirement
// number that was added, removed, or whose own text changed
// (research.md #3).
func changedRequirements(root string, spec ids.EntityID, specPath, from, to string) ([]ChangedElement, error) {
	fromBody, fromFound, err := vcs.FileAtRevision(root, specPath, from)
	if err != nil {
		return nil, err
	}
	toBody, toFound, err := vcs.FileAtRevision(root, specPath, to)
	if err != nil {
		return nil, err
	}

	var fromSections, toSections map[int]artifacts.Section
	if fromFound {
		fromSections = validation.RequirementSections(spec, fromBody)
	}
	if toFound {
		toSections = validation.RequirementSections(spec, toBody)
	}

	numbers := map[int]bool{}
	for n := range fromSections {
		numbers[n] = true
	}
	for n := range toSections {
		numbers[n] = true
	}

	var out []ChangedElement
	for n := range numbers {
		fromSec, hadFrom := fromSections[n]
		toSec, hasTo := toSections[n]

		var status ChangeStatus
		switch {
		case hadFrom && hasTo:
			// Trimmed comparison: Section.Body runs up to (not
			// including) the next heading, so appending or removing an
			// unrelated later Requirement shifts an untouched
			// Requirement's own trailing blank-line count without
			// changing its actual text — trimming avoids reporting
			// that as a false-positive change (implementation finding).
			if strings.TrimSpace(fromSec.Body) == strings.TrimSpace(toSec.Body) {
				continue
			}
			status = Modified
		case !hadFrom && hasTo:
			status = Added
		case hadFrom && !hasTo:
			status = Removed
		default:
			continue
		}

		number := n
		out = append(out, ChangedElement{ID: spec, Path: specPath, Status: status, RequirementNumber: &number})
	}

	return out, nil
}

// sortChangedElements orders els by path, then whole-artifact entries
// before Requirement-level entries, then Requirement number ascending
// (data-model.md "ChangeSet.Elements").
func sortChangedElements(els []ChangedElement) {
	sort.SliceStable(els, func(i, j int) bool {
		if els[i].Path != els[j].Path {
			return els[i].Path < els[j].Path
		}
		iWhole := els[i].RequirementNumber == nil
		jWhole := els[j].RequirementNumber == nil
		if iWhole != jWhole {
			return iWhole
		}
		if iWhole {
			return false
		}
		return *els[i].RequirementNumber < *els[j].RequirementNumber
	})
}
