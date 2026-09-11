package operations

import "github.com/mottamarcio/misterspec/internal/project"

// ParentResult is the outcome of a parent query — deliberately not an
// error for the legitimate "no parent" case (FR-006, FR-007).
type ParentResult struct {
	HasParent bool
	Parent    *ResolvedLocation
}

// Parent determines rawID's structural parent from its own declared
// metadata (already normalized for Task by Inspect, which sets
// Metadata.Parent from the owning tasks.md's `for:` field). It returns
// ParentResult{HasParent: false} — never an error — for an entity type
// with no parent concept (FR-006, FR-007).
func Parent(root string, cfg project.Configuration, rawID string) (ParentResult, error) {
	result, err := Inspect(root, cfg, rawID)
	if err != nil {
		return ParentResult{}, err
	}

	if result.Metadata.Parent == nil {
		return ParentResult{HasParent: false}, nil
	}

	parentLoc, err := Resolve(root, cfg, result.Metadata.Parent.String())
	if err != nil {
		return ParentResult{}, err
	}
	return ParentResult{HasParent: true, Parent: &parentLoc}, nil
}
