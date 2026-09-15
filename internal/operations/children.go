package operations

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// childTypesFor returns the entity types that can be direct structural
// children of parentType, per the project's fixed nesting rules
// (docs/architecture-specification.md §20). A type with no nestable
// children (Spec, Knowledge, Learning, Task) returns nil — Plan/Tasks/
// Validation nest under a Spec too, but have no independent EntityType to
// represent them as a ResolvedLocation here.
func childTypesFor(parentType ids.EntityType) []ids.EntityType {
	switch parentType {
	case ids.Program:
		return []ids.EntityType{ids.Feature}
	case ids.Feature:
		return []ids.EntityType{ids.Spec}
	default:
		return nil
	}
}

// Children enumerates parentRawID's direct structural children, optionally
// filtered to one child EntityType, based on the project's fixed nesting
// rules — never a semantic judgment about what "belongs" together
// (FR-008). It returns an empty (possibly nil) slice, not an error, when
// there are no children of the requested type (FR-009).
func Children(root string, cfg project.Configuration, parentRawID string, filterType *ids.EntityType) ([]ResolvedLocation, error) {
	parentLoc, err := Resolve(root, cfg, parentRawID)
	if err != nil {
		return nil, err
	}

	candidateTypes := childTypesFor(parentLoc.ID.Type)
	if filterType != nil {
		if !containsType(candidateTypes, *filterType) {
			return nil, nil
		}
		candidateTypes = []ids.EntityType{*filterType}
	}
	if len(candidateTypes) == 0 {
		return nil, nil
	}

	parentDir := parentDirectory(parentLoc)

	var children []ResolvedLocation
	for _, ct := range candidateTypes {
		result, err := ids.Scan(root, cfg, ct)
		if err != nil {
			return nil, err
		}
		for number, paths := range result.Paths {
			for _, p := range paths {
				if !isNestedUnder(p, parentDir) {
					continue
				}
				id := ids.EntityID{Type: ct, Prefix: ct.Prefix(), Number: number, Width: cfg.IDWidth}
				children = append(children, ResolvedLocation{
					ID:   id,
					Type: artifactTypeFor(ct),
					Path: resolvedPath(ct, p),
				})
			}
		}
	}

	sort.Slice(children, func(i, j int) bool {
		return children[i].Path < children[j].Path
	})
	return children, nil
}

func containsType(types []ids.EntityType, t ids.EntityType) bool {
	for _, candidate := range types {
		if candidate == t {
			return true
		}
	}
	return false
}

// parentDirectory returns the directory a Program/Feature/Spec's own
// canonical file lives in — e.g. "ai/programs/PRG-001" for
// "ai/programs/PRG-001/program.md" — the prefix children must nest under.
func parentDirectory(loc ResolvedLocation) string {
	return filepath.ToSlash(filepath.Dir(loc.Path))
}

// isNestedUnder reports whether candidatePath lies directly or
// transitively under parentDir, using a "/"-boundary check so
// "PRG-001" never falsely matches a sibling like "PRG-0010".
func isNestedUnder(candidatePath, parentDir string) bool {
	return strings.HasPrefix(candidatePath, parentDir+"/")
}
