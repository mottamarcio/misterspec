package operations

import (
	"sort"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// TaskCollisionReport is one Task number claimed by more than one Spec
// project-wide — the case that collided under the old global-scan
// behavior, now valid, expected state under the per-Spec identity model
// (031-canonical-task-identity spec.md User Story 3, FR-008,
// data-model.md "MigrationDiagnosticEntry"). It is purely informational:
// never a validation Finding, never mutated by anything that produces
// it.
type TaskCollisionReport struct {
	TaskNumber int
	Specs      []ids.EntityID
	Paths      []string
}

// TaskIdentityMigrationDiagnostic reports every Task number claimed by
// more than one Spec, without reading or modifying anything beyond the
// same scan ResolveTask and ValidateProject already perform
// (031-canonical-task-identity spec.md FR-007, FR-008). An empty result
// is a valid, common outcome — most projects will have no collisions —
// not an error.
func TaskIdentityMigrationDiagnostic(root string, cfg project.Configuration) ([]TaskCollisionReport, error) {
	taskScan, err := ids.ScanTasks(root, cfg)
	if err != nil {
		return nil, err
	}

	reports := make([]TaskCollisionReport, 0, len(taskScan.Collisions))
	for _, c := range taskScan.Collisions {
		specIDs := make([]ids.EntityID, 0, len(c.Specs))
		for _, s := range c.Specs {
			specIDs = append(specIDs, ids.EntityID{Type: ids.Spec, Prefix: ids.Spec.Prefix(), Number: s, Width: cfg.IDWidth})
		}
		reports = append(reports, TaskCollisionReport{TaskNumber: c.Task, Specs: specIDs, Paths: c.Paths})
	}
	sort.Slice(reports, func(i, j int) bool { return reports[i].TaskNumber < reports[j].TaskNumber })

	return reports, nil
}
