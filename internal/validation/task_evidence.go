package validation

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/evidence"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// taskCheckboxLinePattern mirrors internal/prepare's own (unexported)
// taskCheckboxPattern — duplicated rather than imported, since
// internal/validation cannot import internal/prepare (internal/prepare
// already imports internal/validation).
var taskCheckboxLinePattern = regexp.MustCompile(`^- \[([ xX])\]`)

// taskEvidenceFindings computes every evidence-related Finding for one
// Spec's own tasks.md — mirrors specCoverageFindings/
// taskDependencyFindings's own per-Task section scan (requirements.go,
// task_dependencies.go); a Spec with no tasks.md yet contributes no
// findings, the same "absence is fine" convention (041-task-evidence-
// fingerprint research.md #8). For every *checked* Task, computes its
// real evidence.EvidenceState against its own current content and
// raises exactly one of CodeUnverifiedTask/CodeStaleTaskEvidence/
// CodeFailedTaskEvidence when that state isn't Verified — an unchecked
// Task never raises any of them (data-model.md "Validation Codes").
func taskEvidenceFindings(root string, cfg project.Configuration, spec ids.EntityID, specDir string) ([]Finding, error) {
	tasksPath := specDir + "/tasks.md"
	tasksBody, err := artifacts.ReadBody(filepath.Join(root, tasksPath))
	if err != nil {
		if errors.Is(err, artifacts.ErrArtifactNotFound) {
			return nil, nil
		}
		return nil, err
	}

	doc := artifacts.ParseDocument(tasksBody)

	var findings []Finding
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
		body := []byte(section.Body)

		if !taskCheckboxChecked(body) {
			continue
		}

		fields := evidence.ParseEvidenceFields(task, body)
		currentFingerprint := evidence.ContentFingerprint(body)
		state := evidence.DeriveState(fields, currentFingerprint)

		taskLoc := fmt.Sprintf("%s#%s", tasksPath, task)
		switch state {
		case evidence.Unverified:
			findings = append(findings, Finding{
				Code:     CodeUnverifiedTask,
				Severity: SeverityError,
				Path:     taskLoc,
				Message:  fmt.Sprintf("%s is checked complete but has no Evidence-Result: at all", task),
			})
		case evidence.Failed:
			findings = append(findings, Finding{
				Code:     CodeFailedTaskEvidence,
				Severity: SeverityError,
				Path:     taskLoc,
				Message:  fmt.Sprintf("%s is checked complete but its most recent Evidence-Result: is \"fail\"", task),
			})
		case evidence.Stale:
			findings = append(findings, Finding{
				Code:     CodeStaleTaskEvidence,
				Severity: SeverityError,
				Path:     taskLoc,
				Message:  fmt.Sprintf("%s is checked complete but its recorded Evidence-Fingerprint no longer matches its own current content", task),
			})
		}
	}

	return findings, nil
}

// taskCheckboxChecked reports whether body's first checkbox line is
// checked — false otherwise, including when no checkbox line is found
// at all.
func taskCheckboxChecked(body []byte) bool {
	for _, line := range strings.Split(string(body), "\n") {
		if cb := taskCheckboxLinePattern.FindStringSubmatch(line); cb != nil {
			return cb[1] != " "
		}
	}
	return false
}
