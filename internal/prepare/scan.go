package prepare

import (
	"errors"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/evidence"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/validation"
)

// TaskInfo bundles everything one Task section of a Spec's tasks.md
// carries, parsed once and reused across readiness computation, the
// explicit-Task path, and automatic selection.
type TaskInfo struct {
	Task       ids.EntityID
	Heading    string
	// Status is "pending" or "complete" — redefined by
	// 041-task-evidence-fingerprint: "complete" now additionally
	// requires Evidence == evidence.Verified, not just a checked
	// checkbox (data-model.md "TaskInfo (extended)"). readiness.go and
	// selection.go are unaffected by this redefinition — they already
	// gate on this one string (research.md #1).
	Status     string
	// Evidence is this Task's real completeness classification, new in
	// 041-task-evidence-fingerprint (data-model.md "TaskInfo
	// (extended)").
	Evidence   evidence.EvidenceState
	Coverage   validation.TaskCoverage
	Dependency validation.TaskDependency
	Fields     TaskFields
}

// taskSectionHeadingPattern matches a Task section's heading text (e.g.
// "TASK-003 — Add session persistence"), mirroring
// internal/validation's own (unexported) pattern of the same shape.
var taskSectionHeadingPattern = regexp.MustCompile(`^TASK-(\d+)\b`)

// taskCheckboxPattern matches a Task's own completion checkbox line,
// mirroring internal/operations's own (unexported) pattern of the same
// shape.
var taskCheckboxPattern = regexp.MustCompile(`^- \[([ xX])\]`)

// ScanSpecTasks parses every Task section of spec's own tasks.md into
// one TaskInfo each — the single shared pass every prepare code path
// (explicit Task, automatic selection) builds on, mirroring 031/032's
// own "one shared parser" precedent. A Spec with no tasks.md yet
// returns an empty, non-nil-error slice — the CLI layer decides what
// that means for the caller's own request (contracts §2).
func ScanSpecTasks(root string, cfg project.Configuration, spec ids.EntityID, specDir string) (tasks []TaskInfo, found bool, err error) {
	tasksPath := specDir + "/tasks.md"
	body, err := artifacts.ReadBody(filepath.Join(root, tasksPath))
	if err != nil {
		if errors.Is(err, artifacts.ErrArtifactNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	doc := artifacts.ParseDocument(body)

	out := []TaskInfo{}
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
		sectionBody := []byte(section.Body)

		evidenceState := taskEvidenceState(task, sectionBody)

		out = append(out, TaskInfo{
			Task:       task,
			Heading:    section.Heading,
			Status:     taskStatus(section.Body, evidenceState),
			Evidence:   evidenceState,
			Coverage:   validation.ParseTaskCoverage(task, sectionBody, cfg),
			Dependency: validation.ParseTaskDependsOn(task, sectionBody, cfg),
			Fields:     parseTaskFields(task, sectionBody),
		})
	}

	return out, true, nil
}

// taskEvidenceState computes task's real evidence.EvidenceState from
// its own already-extracted Section.Body (041-task-evidence-
// fingerprint data-model.md "TaskInfo (extended)").
func taskEvidenceState(task ids.EntityID, body []byte) evidence.EvidenceState {
	fields := evidence.ParseEvidenceFields(task, body)
	currentFingerprint := operations.TaskContentFingerprint(task, body).String()
	return evidence.DeriveState(fields, currentFingerprint)
}

// taskStatus returns "complete" only if body's first checkbox line is
// checked AND state is evidence.Verified — "pending" otherwise
// (including when no checkbox line is found, or when the checkbox is
// checked but the evidence state is Unverified/Stale/Failed). Redefined
// by 041-task-evidence-fingerprint (data-model.md "TaskInfo
// (extended)"): a checked checkbox alone is no longer sufficient —
// this is the single change that makes readiness.go/selection.go
// evidence-aware with no code changes of their own (research.md #1),
// since both already gate on Status == "complete".
func taskStatus(body string, state evidence.EvidenceState) string {
	for _, line := range strings.Split(body, "\n") {
		if cb := taskCheckboxPattern.FindStringSubmatch(line); cb != nil {
			return evidence.TaskStatus(cb[1] != " ", state)
		}
	}
	return "pending"
}
