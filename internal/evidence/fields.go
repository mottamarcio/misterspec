package evidence

import (
	"regexp"
	"strings"

	"github.com/mottamarcio/misterspec/internal/ids"
)

// EvidenceFields is the parsed "Evidence-*:" lines from one Task's own
// Section.Body (041-task-evidence-fingerprint data-model.md
// "EvidenceFields"). Never errors — an absent field is simply its zero
// value, the same precedent internal/prepare.TaskFields already
// established for "Verify:"/"Scope:".
type EvidenceFields struct {
	Task        ids.EntityID
	Result      string
	Origin      string
	By          string
	CapturedAt  string
	Command     string
	GitRevision string
	WorkingTree string
	Fingerprint string
	Log         string
}

var (
	evidenceResultPattern      = regexp.MustCompile(`^Evidence-Result:\s*(.+)$`)
	evidenceOriginPattern      = regexp.MustCompile(`^Evidence-Origin:\s*(.+)$`)
	evidenceByPattern          = regexp.MustCompile(`^Evidence-By:\s*(.+)$`)
	evidenceCapturedAtPattern  = regexp.MustCompile(`^Evidence-CapturedAt:\s*(.+)$`)
	evidenceCommandPattern     = regexp.MustCompile(`^Evidence-Command:\s*(.+)$`)
	evidenceGitRevisionPattern = regexp.MustCompile(`^Evidence-GitRevision:\s*(.+)$`)
	evidenceWorkingTreePattern = regexp.MustCompile(`^Evidence-WorkingTree:\s*(.+)$`)
	evidenceFingerprintPattern = regexp.MustCompile(`^Evidence-Fingerprint:\s*(.+)$`)
	evidenceLogPattern         = regexp.MustCompile(`^Evidence-Log:\s*(.+)$`)
)

// ParseEvidenceFields scans body (one Task's own Section.Body) for each
// "Evidence-*:" line, one regex per label — mirroring
// internal/prepare/fields.go's exact idiom (041-task-evidence-
// fingerprint contracts §3). Absence of every such line yields the
// zero-valued EvidenceFields, never an error (spec Assumptions).
func ParseEvidenceFields(task ids.EntityID, body []byte) EvidenceFields {
	fields := EvidenceFields{Task: task}
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case evidenceResultPattern.MatchString(line):
			fields.Result = evidenceResultPattern.FindStringSubmatch(line)[1]
		case evidenceOriginPattern.MatchString(line):
			fields.Origin = evidenceOriginPattern.FindStringSubmatch(line)[1]
		case evidenceByPattern.MatchString(line):
			fields.By = evidenceByPattern.FindStringSubmatch(line)[1]
		case evidenceCapturedAtPattern.MatchString(line):
			fields.CapturedAt = evidenceCapturedAtPattern.FindStringSubmatch(line)[1]
		case evidenceCommandPattern.MatchString(line):
			fields.Command = evidenceCommandPattern.FindStringSubmatch(line)[1]
		case evidenceGitRevisionPattern.MatchString(line):
			fields.GitRevision = evidenceGitRevisionPattern.FindStringSubmatch(line)[1]
		case evidenceWorkingTreePattern.MatchString(line):
			fields.WorkingTree = evidenceWorkingTreePattern.FindStringSubmatch(line)[1]
		case evidenceFingerprintPattern.MatchString(line):
			fields.Fingerprint = evidenceFingerprintPattern.FindStringSubmatch(line)[1]
		case evidenceLogPattern.MatchString(line):
			fields.Log = evidenceLogPattern.FindStringSubmatch(line)[1]
		}
	}
	return fields
}
