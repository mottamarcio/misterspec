package prepare

import (
	"regexp"
	"strings"

	"github.com/mottamarcio/misterspec/internal/ids"
)

// TaskFields is a Task's own "Verify:"/"Scope:" declarations, parsed
// the same way "Serves:"/"Depends on:" already are
// (034/data-model.md "TaskFields", research.md Decision 1).
type TaskFields struct {
	Task   ids.EntityID
	Verify string
	Scope  string
}

var (
	verifyLinePattern = regexp.MustCompile(`^Verify:\s*(.+)$`)
	scopeLinePattern  = regexp.MustCompile(`^Scope:\s*(.+)$`)
)

// parseTaskFields scans body (one Task's own Section.Body) for its own
// "Verify:" and "Scope:" lines, each captured verbatim. Either line's
// absence yields an empty string, never an error — a Task authored
// before this feature exists has neither (spec FR-003, Assumptions).
func parseTaskFields(task ids.EntityID, body []byte) TaskFields {
	fields := TaskFields{Task: task}
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if sub := verifyLinePattern.FindStringSubmatch(line); sub != nil {
			fields.Verify = sub[1]
			continue
		}
		if sub := scopeLinePattern.FindStringSubmatch(line); sub != nil {
			fields.Scope = sub[1]
		}
	}
	return fields
}
