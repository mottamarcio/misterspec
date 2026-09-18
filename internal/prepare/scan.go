package prepare

import (
	"errors"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/validation"
)

// TaskInfo bundles everything one Task section of a Spec's tasks.md
// carries, parsed once and reused across readiness computation, the
// explicit-Task path, and automatic selection.
type TaskInfo struct {
	Task       ids.EntityID
	Heading    string
	Status     string // "pending" or "complete"
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

		out = append(out, TaskInfo{
			Task:       task,
			Heading:    section.Heading,
			Status:     taskStatus(section.Body),
			Coverage:   validation.ParseTaskCoverage(task, sectionBody, cfg),
			Dependency: validation.ParseTaskDependsOn(task, sectionBody, cfg),
			Fields:     parseTaskFields(task, sectionBody),
		})
	}

	return out, true, nil
}

// taskStatus returns "complete" if body's first checkbox line is
// checked, "pending" otherwise (including when no checkbox line is
// found — a malformed Task section is never silently treated as done).
func taskStatus(body string) string {
	for _, line := range strings.Split(body, "\n") {
		if cb := taskCheckboxPattern.FindStringSubmatch(line); cb != nil {
			if cb[1] == " " {
				return "pending"
			}
			return "complete"
		}
	}
	return "pending"
}
