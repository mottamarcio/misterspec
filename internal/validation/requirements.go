package validation

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// requirementHeadingPattern matches a Requirement heading's text (e.g.
// "R1 — <Requirement>", already stripped of its leading "#"s by
// artifacts.ParseDocument), per kit/templates/spec.md.tmpl's own
// "### R1 — <Requirement>" convention (032/research.md Decision 1). No
// dash, no zero-padding — a deliberately different grammar from
// ids.EntityID's "PREFIX-NNN" convention.
var requirementHeadingPattern = regexp.MustCompile(`^R(\d+)\b`)

// SpecRequirements is one Spec's own declared Requirement set, parsed
// from its spec.md body (032/data-model.md "SpecRequirements").
type SpecRequirements struct {
	Spec       ids.EntityID
	Numbers    []int
	Duplicates []int
}

// parseSpecRequirements scans body (a Spec's own Markdown body, via
// artifacts.ParseDocument) for "R<N>" headings at any level, returning
// every distinct number in document order and any number declared more
// than once (spec FR-001, FR-002).
func parseSpecRequirements(spec ids.EntityID, body []byte) SpecRequirements {
	doc := artifacts.ParseDocument(body)

	seen := map[int]bool{}
	dup := map[int]bool{}
	var numbers []int
	for _, section := range doc.Sections {
		sub := requirementHeadingPattern.FindStringSubmatch(section.Heading)
		if sub == nil {
			continue
		}
		n, err := strconv.Atoi(sub[1])
		if err != nil {
			continue
		}
		if seen[n] {
			dup[n] = true
			continue
		}
		seen[n] = true
		numbers = append(numbers, n)
	}

	var duplicates []int
	for n := range dup {
		duplicates = append(duplicates, n)
	}
	sort.Ints(duplicates)

	return SpecRequirements{Spec: spec, Numbers: numbers, Duplicates: duplicates}
}

// RequirementSections scans body (a Spec's own spec.md body) the same
// way parseSpecRequirements already does, additionally returning each
// found "R<N>" heading's own Section — a new, additive sibling
// (042-impact-analysis-review contracts §2, research.md #3);
// parseSpecRequirements itself is unchanged. A duplicate "R<N>"
// heading keeps only its first occurrence, matching
// parseSpecRequirements's own Numbers (its later occurrence is instead
// reported via Duplicates, which this function does not surface — a
// caller needing that keeps using parseSpecRequirements/
// specCoverageFindings directly).
func RequirementSections(spec ids.EntityID, body []byte) map[int]artifacts.Section {
	doc := artifacts.ParseDocument(body)

	sections := make(map[int]artifacts.Section)
	for _, section := range doc.Sections {
		sub := requirementHeadingPattern.FindStringSubmatch(section.Heading)
		if sub == nil {
			continue
		}
		n, err := strconv.Atoi(sub[1])
		if err != nil {
			continue
		}
		if _, exists := sections[n]; exists {
			continue
		}
		sections[n] = section
	}
	return sections
}

// RequirementRef is a resolved reference to one Requirement, as named by
// a Task's own "Serves:" line (032/data-model.md "RequirementRef").
type RequirementRef struct {
	Spec   ids.EntityID
	Number int
}

// String renders ref in its canonical form (e.g. "SPEC-014:R3").
func (ref RequirementRef) String() string {
	return fmt.Sprintf("%s:R%d", ref.Spec, ref.Number)
}

// parseRequirementRef parses raw ("SPEC-###:R#") against cfg's
// configured Spec ID width, reusing ids.Parse for the Spec half and a
// dedicated unpadded "R<digits>" parser for the Requirement half
// (032/research.md Decision 1, contracts §1).
func parseRequirementRef(raw string, cfg project.Configuration) (RequirementRef, error) {
	idx := strings.LastIndex(raw, ":R")
	if idx <= 0 {
		return RequirementRef{}, fmt.Errorf("validation: %q is not a well-formed requirement reference (want SPEC-###:R#)", raw)
	}

	specPart, numPart := raw[:idx], raw[idx+2:]
	if numPart == "" {
		return RequirementRef{}, fmt.Errorf("validation: %q is missing its requirement number", raw)
	}
	for _, r := range numPart {
		if r < '0' || r > '9' {
			return RequirementRef{}, fmt.Errorf("validation: %q's requirement number %q is not all-numeric", raw, numPart)
		}
	}
	number, err := strconv.Atoi(numPart)
	if err != nil || number < 1 {
		return RequirementRef{}, fmt.Errorf("validation: %q's requirement number is invalid", raw)
	}

	spec, err := ids.ParseAny(specPart)
	if err != nil {
		return RequirementRef{}, fmt.Errorf("validation: %q's Spec half %q: %v", raw, specPart, err)
	}
	spec, err = ids.Parse(spec.Type, specPart, cfg.IDWidth)
	if err != nil {
		return RequirementRef{}, fmt.Errorf("validation: %q's Spec half %q: %v", raw, specPart, err)
	}
	if spec.Type != ids.Spec {
		return RequirementRef{}, fmt.Errorf("validation: %q's Spec half %q is not a Spec ID", raw, specPart)
	}

	return RequirementRef{Spec: spec, Number: number}, nil
}

// TaskCoverage is one Task's own set of "Serves:" references, parsed
// from its Section.Body in tasks.md (032/data-model.md "TaskCoverage").
type TaskCoverage struct {
	Task                ids.EntityID
	References          []RequirementRef
	MalformedReferences []string
}

// servesLinePattern matches a "Serves: ..." line inside a Task's own
// body (kit/skills/mister-analyze/SKILL.md's literal "Serves:
// SPEC-###:R#" convention).
var servesLinePattern = regexp.MustCompile(`^Serves:\s*(.+)$`)

// ParseTaskCoverage scans body (one Task's own Section.Body) for every
// "Serves:" line, splitting each on commas and unioning across repeated
// lines (032/research.md Decision 3). A malformed entry is recorded in
// MalformedReferences rather than silently dropped.
func ParseTaskCoverage(task ids.EntityID, body []byte, cfg project.Configuration) TaskCoverage {
	coverage := TaskCoverage{Task: task}

	for _, line := range strings.Split(string(body), "\n") {
		sub := servesLinePattern.FindStringSubmatch(strings.TrimSpace(line))
		if sub == nil {
			continue
		}
		for _, entry := range strings.Split(sub[1], ",") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			ref, err := parseRequirementRef(entry, cfg)
			if err != nil {
				coverage.MalformedReferences = append(coverage.MalformedReferences, entry)
				continue
			}
			coverage.References = append(coverage.References, ref)
		}
	}

	return coverage
}

// taskSectionHeadingPattern matches a Task section's heading text (e.g.
// "TASK-003 — Add session persistence", already stripped of its leading
// "#"s by artifacts.ParseDocument) — the same "TASK-NNN" shape
// internal/ids/scan.go's own taskHeadingPattern matches against a raw
// line, applied here to a Section.Heading instead (research.md
// Decision 2).
var taskSectionHeadingPattern = regexp.MustCompile(`^TASK-(\d+)\b`)

// RequirementCoverageReport is one Spec's requirement-coverage facts,
// computed by specCoverageFindings (032/data-model.md
// "RequirementCoverageReport").
type RequirementCoverageReport struct {
	Spec                  ids.EntityID
	UncoveredRequirements []int
	TasksWithoutCoverage  []ids.EntityID
}

// specCoverageFindings computes every requirement-coverage Finding for
// one Spec (duplicate Requirement IDs, uncovered Requirements, unknown/
// malformed/cross-Spec Serves: references, Tasks without any Serves:
// reference), reading specDir's own spec.md and, if present, tasks.md
// directly (032/contracts/requirement-coverage-and-dependency-validation.md
// §2). A Spec with no tasks.md yet is not an error — every Requirement
// is simply reported uncovered, matching ids.ScanTasks's own "absence is
// fine" behavior. A spec.md that fails to even parse is not
// double-reported here — internal/validation's own frontmatter checks
// (checkEntity) already surface that.
func specCoverageFindings(root string, cfg project.Configuration, spec ids.EntityID, specDir string) ([]Finding, RequirementCoverageReport, error) {
	report := RequirementCoverageReport{Spec: spec}

	specPath := specDir + "/spec.md"
	specBody, err := artifacts.ReadBody(filepath.Join(root, specPath))
	if err != nil {
		return nil, report, nil
	}
	specReq := parseSpecRequirements(spec, specBody)

	var findings []Finding
	for _, dup := range specReq.Duplicates {
		findings = append(findings, Finding{
			Code:     CodeDuplicateRequirementID,
			Severity: SeverityError,
			Path:     specPath,
			Message:  fmt.Sprintf("%s: requirement R%d is declared more than once", spec, dup),
		})
	}

	tasksPath := specDir + "/tasks.md"
	tasksBody, err := artifacts.ReadBody(filepath.Join(root, tasksPath))
	var taskCoverages []TaskCoverage
	switch {
	case err == nil:
		doc := artifacts.ParseDocument(tasksBody)
		for _, section := range doc.Sections {
			sub := taskSectionHeadingPattern.FindStringSubmatch(section.Heading)
			if sub == nil {
				continue
			}
			n, convErr := strconv.Atoi(sub[1])
			if convErr != nil || n < 1 {
				continue
			}
			task := ids.EntityID{Type: ids.Task, Prefix: ids.Task.Prefix(), Number: n, Width: cfg.IDWidth}
			taskCoverages = append(taskCoverages, ParseTaskCoverage(task, []byte(section.Body), cfg))
		}
	case errors.Is(err, artifacts.ErrArtifactNotFound):
		// No tasks.md yet — every Requirement below is simply uncovered.
	default:
		return nil, RequirementCoverageReport{}, err
	}

	covered := map[int]bool{}
	for _, tc := range taskCoverages {
		taskLoc := fmt.Sprintf("%s#%s", tasksPath, tc.Task)

		if len(tc.References) == 0 && len(tc.MalformedReferences) == 0 {
			report.TasksWithoutCoverage = append(report.TasksWithoutCoverage, tc.Task)
			findings = append(findings, Finding{
				Code:     CodeTaskWithoutRequirement,
				Severity: SeverityError,
				Path:     taskLoc,
				Message:  fmt.Sprintf("%s has no Serves: reference", tc.Task),
			})
		}

		for _, ref := range tc.References {
			if ref.Spec != spec {
				findings = append(findings, Finding{
					Code:     CodeCrossSpecRequirementReference,
					Severity: SeverityError,
					Path:     taskLoc,
					Message:  fmt.Sprintf("%s (owned by %s) references %s — coverage must reference the task's own Spec", tc.Task, spec, ref),
				})
				continue
			}
			if !intSliceContains(specReq.Numbers, ref.Number) {
				findings = append(findings, Finding{
					Code:     CodeUnknownRequirementReference,
					Severity: SeverityError,
					Path:     taskLoc,
					Message:  fmt.Sprintf("%s references %s, which does not exist", tc.Task, ref),
				})
				continue
			}
			covered[ref.Number] = true
		}

		for _, raw := range tc.MalformedReferences {
			findings = append(findings, Finding{
				Code:     CodeUnknownRequirementReference,
				Severity: SeverityError,
				Path:     taskLoc,
				Message:  fmt.Sprintf("%s has a malformed Serves: reference %q", tc.Task, raw),
			})
		}
	}

	for _, n := range specReq.Numbers {
		if !covered[n] {
			report.UncoveredRequirements = append(report.UncoveredRequirements, n)
			findings = append(findings, Finding{
				Code:     CodeUncoveredRequirement,
				Severity: SeverityError,
				Path:     specPath,
				Message:  fmt.Sprintf("%s: requirement R%d has no Task serving it", spec, n),
			})
		}
	}

	return findings, report, nil
}

// SpecPhaseGate is one Spec's computed phase-aware coverage gate
// (032/data-model.md "SpecPhaseGate", research.md Decision 6): a Spec
// still in "draft" is exempt from coverage completeness; any later
// status blocks on an incomplete RequirementCoverageReport.
type SpecPhaseGate struct {
	Spec    ids.EntityID
	Status  string
	Exempt  bool
	Blocked bool
}

// computeSpecPhaseGate applies the phase gate rule to report, given
// spec's own declared status (spec FR-011, FR-012).
func computeSpecPhaseGate(spec ids.EntityID, status string, report RequirementCoverageReport) SpecPhaseGate {
	gate := SpecPhaseGate{Spec: spec, Status: status, Exempt: status == "draft"}
	gate.Blocked = !gate.Exempt && (len(report.UncoveredRequirements) > 0 || len(report.TasksWithoutCoverage) > 0)
	return gate
}

// phaseGateFinding renders a blocked SpecPhaseGate as a Finding, emitted
// in addition to (never instead of) the coverage Finding(s) it escalates
// (contracts §1's "phase_gate_blocked" row).
func phaseGateFinding(gate SpecPhaseGate, specPath string) Finding {
	return Finding{
		Code:     CodePhaseGateBlocked,
		Severity: SeverityError,
		Path:     specPath,
		Message:  fmt.Sprintf("%s has status %q but incomplete requirement coverage — not ready for implementation", gate.Spec, gate.Status),
	}
}

func intSliceContains(list []int, n int) bool {
	for _, v := range list {
		if v == n {
			return true
		}
	}
	return false
}
