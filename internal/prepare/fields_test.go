package prepare

import "testing"

// TestParseTaskFields_BothPresent covers data-model.md "TaskFields"
// (research.md Decision 1): Verify:/Scope: lines are captured verbatim.
func TestParseTaskFields_BothPresent(t *testing.T) {
	body := []byte("- [ ] Complete\n\nScope: internal/auth/refresh.go\n\nVerify: go test ./internal/auth/... -run TestRefreshRotation\n")
	got := parseTaskFields(taskID(1), body)
	if got.Scope != "internal/auth/refresh.go" {
		t.Errorf("Scope = %q, want %q", got.Scope, "internal/auth/refresh.go")
	}
	if got.Verify != "go test ./internal/auth/... -run TestRefreshRotation" {
		t.Errorf("Verify = %q, want the exact verification command", got.Verify)
	}
}

// TestParseTaskFields_BothAbsent covers spec FR-003/Assumptions: a Task
// authored before this feature exists has neither line, and that is
// not an error.
func TestParseTaskFields_BothAbsent(t *testing.T) {
	body := []byte("- [ ] Complete\n\nServes: SPEC-001:R1\n")
	got := parseTaskFields(taskID(1), body)
	if got.Scope != "" || got.Verify != "" {
		t.Errorf("got = %+v, want both empty when absent", got)
	}
}
