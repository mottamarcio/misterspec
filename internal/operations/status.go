package operations

import (
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/validation"
)

// StatusSummary is a fresh, computed-on-demand snapshot of the project's
// current state (FR-012, FR-013).
type StatusSummary struct {
	// Counts is the number of existing entities per type — includes
	// Task, discovered the same way ValidateProject's free duplicate-ID
	// check does (004-structural-validation's research.md).
	Counts map[ids.EntityType]int
	// SpecsByState is the number of Specs currently in each declared
	// lifecycle state.
	SpecsByState map[string]int
	// StructuralErrors is len(validation.ValidateProject(root, cfg)) for
	// this same project state — computed by calling it, never a
	// separate, potentially-drifting count (research.md).
	StructuralErrors int
}

// statusEntityTypes are every entity type Status counts.
var statusEntityTypes = []ids.EntityType{ids.Program, ids.Feature, ids.Spec, ids.Knowledge, ids.Learning, ids.Task}

// Status computes StatusSummary fresh from the current filesystem on
// every call — no caching, no stored state (FR-013).
func Status(root string, cfg project.Configuration) (StatusSummary, error) {
	summary := StatusSummary{
		Counts:       map[ids.EntityType]int{},
		SpecsByState: map[string]int{},
	}

	var specIDs []ids.EntityID
	for _, t := range statusEntityTypes {
		result, err := ids.Scan(root, cfg, t)
		if err != nil {
			return StatusSummary{}, err
		}
		summary.Counts[t] = len(result.IDs)
		if t == ids.Spec {
			specIDs = result.IDs
		}
	}

	for _, specID := range specIDs {
		inspected, err := Inspect(root, cfg, specID.String())
		if err != nil {
			// A structurally broken Spec is reported via
			// StructuralErrors below, not silently dropped nor allowed
			// to fail the whole summary.
			continue
		}
		summary.SpecsByState[inspected.Metadata.Status]++
	}

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		return StatusSummary{}, err
	}
	summary.StructuralErrors = len(findings)

	return summary, nil
}
