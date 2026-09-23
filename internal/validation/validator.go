package validation

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// projectEntityTypes are the standalone entity types ValidateProject runs
// full per-entity checks against.
var projectEntityTypes = []ids.EntityType{ids.Program, ids.Feature, ids.Spec, ids.Knowledge, ids.Learning}

// ValidateProject runs ValidateEntity's own per-entity checks (via
// validateFoundEntity) against every discovered Program/Feature/Spec/
// Knowledge/Learning, aggregating every Finding (FR-003, FR-004). It also
// detects duplicate Task IDs — the one check Task qualifies for "for
// free" via ids.Scan's existing heading-based discovery, without the
// body-content parsing this feature otherwise defers (research.md).
// error is reserved for a filesystem-level failure, never for a
// structural problem — those are always Findings (FR-011).
func ValidateProject(root string, cfg project.Configuration) ([]Finding, error) {
	var findings []Finding

	for _, t := range projectEntityTypes {
		result, err := ids.Scan(root, cfg, t)
		if err != nil {
			return nil, err
		}
		for _, number := range sortedNumbers(result.Paths) {
			paths := result.Paths[number]
			findings = append(findings, validateFoundEntity(root, cfg, t, number, paths)...)

			if t == ids.Spec && len(paths) > 0 {
				specID := ids.EntityID{Type: ids.Spec, Prefix: ids.Spec.Prefix(), Number: number, Width: cfg.IDWidth}
				covFindings, report, err := specCoverageFindings(root, cfg, specID, paths[0])
				if err != nil {
					return nil, err
				}
				findings = append(findings, covFindings...)

				specPath := paths[0] + "/spec.md"
				if meta, err := artifacts.ParseMetadata(filepath.Join(root, specPath)); err == nil {
					gate := computeSpecPhaseGate(specID, meta.Status, report)
					if gate.Blocked {
						findings = append(findings, phaseGateFinding(gate, specPath))
					}
				}

				depFindings, err := taskDependencyFindings(root, cfg, specID, paths[0])
				if err != nil {
					return nil, err
				}
				findings = append(findings, depFindings...)

				evFindings, err := taskEvidenceFindings(root, cfg, specID, paths[0])
				if err != nil {
					return nil, err
				}
				findings = append(findings, evFindings...)
			}
		}
	}

	taskScan, err := ids.ScanTasks(root, cfg)
	if err != nil {
		return nil, err
	}
	for _, dup := range taskScan.Duplicates {
		findings = append(findings, taskDuplicateFinding(dup, cfg))
	}

	graph, specPath, err := projectDependencyGraph(root, cfg)
	if err != nil {
		return nil, err
	}
	for _, cycle := range detectCycles(graph) {
		findings = append(findings, dependencyCycleFinding(cycle, cfg, specPath))
	}

	findings = append(findings, checkConstitution(root, cfg)...)

	return findings, nil
}

// checkConstitution structurally checks the project's Constitution file
// (docs/architecture-specification.md §24) — additive to, not folded
// into, projectEntityTypes' own loop, since the Constitution has no
// EntityType/ID of its own (027-constitution-frontmatter-task-deps,
// research.md). A Constitution that does not exist yet is not a
// structural problem (create-constitution/SKILL.md's own Preconditions
// already allow first-run creation) — only a present-but-malformed or
// present-but-incomplete one produces a Finding.
func checkConstitution(root string, cfg project.Configuration) []Finding {
	path := cfg.ConstitutionPath

	meta, err := artifacts.ParseMetadata(filepath.Join(root, path))
	if err != nil {
		if errors.Is(err, artifacts.ErrArtifactNotFound) {
			return nil
		}
		code := CodeFrontmatterMalformed
		if errors.Is(err, artifacts.ErrRequiredFieldMissing) {
			code = CodeRequiredFieldMissing
		}
		return []Finding{{Code: code, Severity: SeverityError, Path: path, Message: err.Error()}}
	}

	if meta.SchemaVersion == 0 {
		return []Finding{{
			Code:     CodeRequiredFieldMissing,
			Severity: SeverityError,
			Path:     path,
			Message:  fmt.Sprintf("%s: field \"schema_version\" is required for type %q", path, meta.Type),
		}}
	}

	return nil
}

// taskDuplicateFinding renders one ids.TaskDuplicate (already scoped to
// a single Spec by ids.ScanTasks) as a Finding naming that Spec
// explicitly, so the scope is legible without cross-referencing the path
// (031-canonical-task-identity spec.md FR-004,
// contracts/task-identity-resolution.md §3). Shared by ValidateProject
// and ValidateEntity so a duplicate's wording never drifts between the
// two call sites.
func taskDuplicateFinding(dup ids.TaskDuplicate, cfg project.Configuration) Finding {
	spec := ids.EntityID{Type: ids.Spec, Prefix: ids.Spec.Prefix(), Number: dup.Spec, Width: cfg.IDWidth}
	return Finding{
		Code:     CodeDuplicateID,
		Severity: SeverityError,
		Path:     dup.Paths[0],
		Message:  fmt.Sprintf("%s: %d Task headings claim the same number %d: %v", spec, len(dup.Paths), dup.Task, dup.Paths),
	}
}

func sortedNumbers(paths map[int][]string) []int {
	nums := make([]int, 0, len(paths))
	for n := range paths {
		nums = append(nums, n)
	}
	sort.Ints(nums)
	return nums
}

// ErrInvalidTarget is returned by ValidateEntity for a genuinely unusable
// request: a syntactically malformed rawID, or a type this feature does
// not validate as a standalone entity (Task, Plan, Tasks, Validation,
// Constitution — mirroring Create's own supported-type boundary).
// Every structural anomaly *about* a well-formed request's target is a
// Finding, never this error (research.md).
var ErrInvalidTarget = errors.New("validation: invalid target")

// validatableTypes are the entity types ValidateEntity accepts as a
// standalone target.
var validatableTypes = map[ids.EntityType]bool{
	ids.Program:   true,
	ids.Feature:   true,
	ids.Spec:      true,
	ids.Knowledge: true,
	ids.Learning:  true,
}

// canonicalFilename mirrors internal/operations/resolve.go's unexported
// helper of the same shape — duplicated rather than imported to avoid the
// import cycle recorded in research.md. Program/Feature/Spec resolve to a
// directory + fixed filename; Knowledge/Learning's Scan-reported path is
// already the exact file.
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

// ValidateEntity checks rawID's entity against every applicable
// structural rule (FR-001, FR-002), returning every Finding — never
// stopping at the first. error is reserved for a genuinely unusable
// request; see ErrInvalidTarget.
func ValidateEntity(root string, cfg project.Configuration, rawID string) ([]Finding, error) {
	loose, err := ids.ParseAny(rawID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTarget, err)
	}
	if !validatableTypes[loose.Type] {
		return nil, fmt.Errorf("%w: %v is not a standalone-validatable entity type", ErrInvalidTarget, loose.Type)
	}
	id, err := ids.Parse(loose.Type, rawID, cfg.IDWidth)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTarget, err)
	}

	scanResult, err := ids.Scan(root, cfg, id.Type)
	if err != nil {
		return nil, err
	}

	findings := validateFoundEntity(root, cfg, id.Type, id.Number, scanResult.Paths[id.Number])

	if id.Type == ids.Spec {
		// A single-Spec validation MUST surface that Spec's own Task
		// duplicates the same way ValidateProject would (spec FR-011) —
		// scoped to id.Number only, never another Spec's duplicates.
		taskScan, err := ids.ScanTasks(root, cfg)
		if err != nil {
			return nil, err
		}
		for _, dup := range taskScan.Duplicates {
			if dup.Spec != id.Number {
				continue
			}
			findings = append(findings, taskDuplicateFinding(dup, cfg))
		}

		if paths := scanResult.Paths[id.Number]; len(paths) > 0 {
			covFindings, report, err := specCoverageFindings(root, cfg, id, paths[0])
			if err != nil {
				return nil, err
			}
			findings = append(findings, covFindings...)

			specPath := paths[0] + "/spec.md"
			if meta, err := artifacts.ParseMetadata(filepath.Join(root, specPath)); err == nil {
				gate := computeSpecPhaseGate(id, meta.Status, report)
				if gate.Blocked {
					findings = append(findings, phaseGateFinding(gate, specPath))
				}
			}

			depFindings, err := taskDependencyFindings(root, cfg, id, paths[0])
			if err != nil {
				return nil, err
			}
			findings = append(findings, depFindings...)

			evFindings, err := taskEvidenceFindings(root, cfg, id, paths[0])
			if err != nil {
				return nil, err
			}
			findings = append(findings, evFindings...)
		}

		// A cycle is a fact about the project-wide graph, not just this
		// Spec's own file — check the same full graph checkDependencyList
		// already resolves against, and report only cycles this Spec
		// actually participates in (contracts §3).
		graph, specPath, err := projectDependencyGraph(root, cfg)
		if err != nil {
			return nil, err
		}
		for _, cycle := range detectCycles(graph) {
			if intSliceContains(cycle.Path, id.Number) {
				findings = append(findings, dependencyCycleFinding(cycle, cfg, specPath))
			}
		}
	}

	return findings, nil
}

// validateFoundEntity runs checks against whatever ids.Scan found for
// (t, number): zero paths (not found), more than one (duplicate — still
// checked further against the first, since a duplicate doesn't excuse its
// own separate problems from being reported too), or exactly one.
func validateFoundEntity(root string, cfg project.Configuration, t ids.EntityType, number int, paths []string) []Finding {
	if len(paths) == 0 {
		return []Finding{{
			Code:     CodeNotFound,
			Severity: SeverityError,
			Message:  fmt.Sprintf("no %s artifact with number %d exists", t, number),
		}}
	}

	var findings []Finding
	if len(paths) > 1 {
		findings = append(findings, Finding{
			Code:     CodeDuplicateID,
			Severity: SeverityError,
			Path:     paths[0],
			Message:  fmt.Sprintf("%d artifacts claim the same %s number %d: %v", len(paths), t, number, paths),
		})
	}
	findings = append(findings, checkEntity(root, cfg, t, number, paths[0])...)
	return findings
}

// checkEntity runs the frontmatter/required-field/ID-location/status/
// parent/dependency checks for one entity already known to exist at
// scanPath.
func checkEntity(root string, cfg project.Configuration, t ids.EntityType, number int, scanPath string) []Finding {
	filePath := scanPath
	if filename, ok := canonicalFilename(t); ok {
		filePath = scanPath + "/" + filename
	}

	meta, err := artifacts.ParseMetadata(filepath.Join(root, filePath))
	if err != nil {
		code := CodeFrontmatterMalformed
		if errors.Is(err, artifacts.ErrRequiredFieldMissing) {
			code = CodeRequiredFieldMissing
		}
		return []Finding{{Code: code, Severity: SeverityError, Path: filePath, Message: err.Error()}}
	}

	var findings []Finding

	if want := (ids.EntityID{Type: t, Prefix: t.Prefix(), Number: number, Width: cfg.IDWidth}); meta.ID == nil || *meta.ID != want {
		findings = append(findings, Finding{
			Code:     CodeIDLocationMismatch,
			Severity: SeverityError,
			Path:     filePath,
			Message:  fmt.Sprintf("declared id %v does not match canonical location %v", meta.ID, want),
		})
	}

	if states, ok := allowedStates(t); ok && !stringIn(meta.Status, states) {
		findings = append(findings, Finding{
			Code:     CodeInvalidStatus,
			Severity: SeverityError,
			Path:     filePath,
			Message:  fmt.Sprintf("status %q is not one of the allowed states for %v: %v", meta.Status, t, states),
		})
	}

	if wantParentType, needsParent := requiredParentType(t); needsParent {
		findings = append(findings, checkParent(root, cfg, filePath, meta.Parent, wantParentType)...)
	}

	if t == ids.Spec {
		findings = append(findings, checkDependencyList(root, cfg, filePath, "depends_on", meta.DependsOn)...)
		findings = append(findings, checkDependencyList(root, cfg, filePath, "supersedes", meta.Supersedes)...)
	}

	findings = append(findings, checkWikilinks(root, cfg, filePath)...)
	findings = append(findings, checkAnchors(root, filePath)...)

	return findings
}

func checkParent(root string, cfg project.Configuration, filePath string, parent *ids.EntityID, wantType ids.EntityType) []Finding {
	if parent == nil {
		return []Finding{{
			Code:     CodeMissingParent,
			Severity: SeverityError,
			Path:     filePath,
			Message:  fmt.Sprintf("no parent declared; a %v parent is required", wantType),
		}}
	}
	if parent.Type != wantType {
		return []Finding{{
			Code:     CodeInvalidParentType,
			Severity: SeverityError,
			Path:     filePath,
			Message:  fmt.Sprintf("parent %v is a %v, want a %v", parent, parent.Type, wantType),
		}}
	}

	result, err := ids.Scan(root, cfg, wantType)
	if err != nil {
		return []Finding{{Code: CodeMissingParent, Severity: SeverityError, Path: filePath, Message: err.Error()}}
	}
	if len(result.Paths[parent.Number]) == 0 {
		return []Finding{{
			Code:     CodeMissingParent,
			Severity: SeverityError,
			Path:     filePath,
			Message:  fmt.Sprintf("declared parent %v does not exist", parent),
		}}
	}
	return nil
}

func checkDependencyList(root string, cfg project.Configuration, filePath, field string, entries []ids.EntityID) []Finding {
	if len(entries) == 0 {
		return nil
	}
	result, err := ids.Scan(root, cfg, ids.Spec)
	if err != nil {
		return []Finding{{Code: CodeUnresolvedDependency, Severity: SeverityError, Path: filePath, Message: err.Error()}}
	}

	var findings []Finding
	for _, entry := range entries {
		if entry.Type != ids.Spec || len(result.Paths[entry.Number]) == 0 {
			findings = append(findings, Finding{
				Code:     CodeUnresolvedDependency,
				Severity: SeverityError,
				Path:     filePath,
				Message:  fmt.Sprintf("%s entry %v does not resolve to an existing Spec", field, entry),
			})
		}
	}
	return findings
}

func stringIn(s string, list []string) bool {
	for _, v := range list {
		if s == v {
			return true
		}
	}
	return false
}
