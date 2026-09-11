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
			findings = append(findings, validateFoundEntity(root, cfg, t, number, result.Paths[number])...)
		}
	}

	taskResult, err := ids.Scan(root, cfg, ids.Task)
	if err != nil {
		return nil, err
	}
	for _, dup := range taskResult.Duplicates {
		findings = append(findings, Finding{
			Code:     CodeDuplicateID,
			Severity: SeverityError,
			Path:     dup.Paths[0],
			Message:  fmt.Sprintf("%d Task headings claim the same number %d: %v", len(dup.Paths), dup.ID.Number, dup.Paths),
		})
	}

	return findings, nil
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

	return validateFoundEntity(root, cfg, id.Type, id.Number, scanResult.Paths[id.Number]), nil
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
