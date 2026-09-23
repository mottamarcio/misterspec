package validation

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// dependsOnLinePattern matches a "Depends on: ..." line inside a Task's
// own body — the new label 034-task-oriented-context-preparation
// introduces, mirroring Serves:'s exact grammar (research.md
// Decision 1).
var dependsOnLinePattern = regexp.MustCompile(`^Depends on:\s*(.+)$`)

// TaskDependency is one Task's own "Depends on:" declaration
// (034/data-model.md "TaskDependency").
type TaskDependency struct {
	Task               ids.EntityID
	DependsOn          []ids.EntityID
	MalformedDependsOn []string
}

// ParseTaskDependsOn scans body (one Task's own Section.Body) for every
// "Depends on:" line, splitting each on commas and unioning across
// repeated lines — the same shape ParseTaskCoverage already uses for
// Serves: (research.md Decision 1). An explicit "none" value and an
// absent line both yield an empty DependsOn — a Task authored before
// this feature exists is never treated as an error (spec FR-003).
func ParseTaskDependsOn(task ids.EntityID, body []byte, cfg project.Configuration) TaskDependency {
	dep := TaskDependency{Task: task}

	for _, line := range strings.Split(string(body), "\n") {
		sub := dependsOnLinePattern.FindStringSubmatch(strings.TrimSpace(line))
		if sub == nil {
			continue
		}
		for _, entry := range strings.Split(sub[1], ",") {
			entry = strings.TrimSpace(entry)
			if entry == "" || strings.EqualFold(entry, "none") {
				continue
			}
			id, err := parseTaskNumberOnly(entry, cfg)
			if err != nil {
				dep.MalformedDependsOn = append(dep.MalformedDependsOn, entry)
				continue
			}
			dep.DependsOn = append(dep.DependsOn, id)
		}
	}

	return dep
}

// parseTaskNumberOnly parses raw as a bare "TASK-###" against cfg's
// configured ID width — a Depends on: entry is always local to its own
// Spec (spec FR-009), so it never accepts the composite SPEC-###:TASK-###
// form the way a Serves: reference accepts SPEC-###:R#; a composite
// form here is deliberately rejected as malformed, since Task
// dependencies are local by construction (031-canonical-task-identity's
// own per-Spec Task identity).
// DetectTaskDependencyCycles returns every distinct cycle among deps'
// own "Depends on:" edges, scoped to whichever Spec deps' own Tasks
// belong to (the caller is responsible for passing only one Spec's own
// TaskDependency values at a time) — reusing the existing
// DependencyGraph/detectCycles this package already defines for
// Spec-level cycles (dependency_graph.go, 032), at Task-number
// granularity instead (research.md Decision 2). Exported so both this
// package's own ValidateProject/ValidateEntity wiring and
// internal/prepare's readiness computation share the identical answer
// (031/032's own "one shared parser" precedent).
func DetectTaskDependencyCycles(deps []TaskDependency) [][]ids.EntityID {
	numberToID := map[int]ids.EntityID{}
	var nodes []int
	edges := map[int][]int{}
	for _, d := range deps {
		nodes = append(nodes, d.Task.Number)
		numberToID[d.Task.Number] = d.Task
		for _, dep := range d.DependsOn {
			edges[d.Task.Number] = append(edges[d.Task.Number], dep.Number)
			numberToID[dep.Number] = dep
		}
	}

	graph := DependencyGraph{Nodes: nodes, Edges: edges}
	cycles := detectCycles(graph)

	out := make([][]ids.EntityID, 0, len(cycles))
	for _, c := range cycles {
		path := make([]ids.EntityID, 0, len(c.Path))
		for _, n := range c.Path {
			path = append(path, numberToID[n])
		}
		out = append(out, path)
	}
	return out
}

// DanglingTaskDependency is one Task's "Depends on:" entry naming a
// Task number that does not exist among the same Spec's own Tasks
// (spec FR-009).
type DanglingTaskDependency struct {
	Task    ids.EntityID
	Missing ids.EntityID
}

// InvalidTaskDependencies returns every entry across deps whose
// DependsOn names a Task number that does not exist among deps' own
// Tasks — a dangling reference within the same Spec. A cross-Spec
// reference is already rejected at parse time (parseTaskNumberOnly) and
// surfaces in MalformedDependsOn instead, not here.
func InvalidTaskDependencies(deps []TaskDependency) []DanglingTaskDependency {
	known := map[int]bool{}
	for _, d := range deps {
		known[d.Task.Number] = true
	}

	var out []DanglingTaskDependency
	for _, d := range deps {
		for _, dep := range d.DependsOn {
			if !known[dep.Number] {
				out = append(out, DanglingTaskDependency{Task: d.Task, Missing: dep})
			}
		}
	}
	return out
}

// taskDependencyFindings computes every Task-dependency Finding for one
// Spec (cycle, dangling reference), reading specDir's own tasks.md
// directly — mirrors specCoverageFindings's own per-Task section scan
// (requirements.go), a Spec with no tasks.md yet contributes no
// findings, matching that same "absence is fine" convention.
func taskDependencyFindings(root string, cfg project.Configuration, spec ids.EntityID, specDir string) ([]Finding, error) {
	tasksPath := specDir + "/tasks.md"
	tasksBody, err := artifacts.ReadBody(filepath.Join(root, tasksPath))
	if err != nil {
		if errors.Is(err, artifacts.ErrArtifactNotFound) {
			return nil, nil
		}
		return nil, err
	}

	doc := artifacts.ParseDocument(tasksBody)
	var deps []TaskDependency
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
		deps = append(deps, ParseTaskDependsOn(task, []byte(section.Body), cfg))
	}

	var findings []Finding
	for _, cycle := range DetectTaskDependencyCycles(deps) {
		rendered := make([]string, len(cycle))
		for i, id := range cycle {
			rendered[i] = fmt.Sprintf("%s:%s", spec, id)
		}
		findings = append(findings, Finding{
			Code:     CodeTaskDependencyCycle,
			Severity: SeverityError,
			Path:     fmt.Sprintf("%s#%s", tasksPath, cycle[0]),
			Message:  fmt.Sprintf("task dependency cycle: %s", strings.Join(rendered, " -> ")),
		})
	}
	for _, dangling := range InvalidTaskDependencies(deps) {
		findings = append(findings, Finding{
			Code:     CodeInvalidTaskDependency,
			Severity: SeverityError,
			Path:     fmt.Sprintf("%s#%s", tasksPath, dangling.Task),
			Message:  fmt.Sprintf("%s:%s depends on %s:%s, which does not exist", spec, dangling.Task, spec, dangling.Missing),
		})
	}

	return findings, nil
}

func parseTaskNumberOnly(raw string, cfg project.Configuration) (ids.EntityID, error) {
	if strings.Contains(raw, ":") {
		return ids.EntityID{}, fmt.Errorf("validation: %q names a different Spec — Depends on: is local to its own Spec", raw)
	}
	id, err := ids.Parse(ids.Task, raw, cfg.IDWidth)
	if err != nil {
		return ids.EntityID{}, err
	}
	return id, nil
}
