package prepare

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/validation"
)

// TestSelectTask_LowestNumberedReadyTask covers research.md Decision 6:
// among multiple Ready Tasks, the lowest-numbered one is chosen.
func TestSelectTask_LowestNumberedReadyTask(t *testing.T) {
	all := []TaskInfo{
		{Task: taskID(2), Status: "pending", Dependency: validation.TaskDependency{Task: taskID(2)}},
		{Task: taskID(1), Status: "pending", Dependency: validation.TaskDependency{Task: taskID(1)}},
	}
	statusOf := func(ids.EntityID) string { return "complete" }

	got, ok := SelectTask(all, statusOf)
	if !ok {
		t.Fatalf("SelectTask() ok = false, want true")
	}
	if got.Task != taskID(1) {
		t.Errorf("SelectTask() = %+v, want TASK-001", got)
	}
}

// TestSelectTask_SkipsBlockedPrefersLowestReady covers that a blocked
// lower-numbered Task is skipped in favor of the lowest-numbered
// *Ready* Task.
func TestSelectTask_SkipsBlockedPrefersLowestReady(t *testing.T) {
	all := []TaskInfo{
		{Task: taskID(1), Status: "pending", Dependency: validation.TaskDependency{Task: taskID(1), DependsOn: []ids.EntityID{taskID(9)}}},
		{Task: taskID(2), Status: "pending", Dependency: validation.TaskDependency{Task: taskID(2)}},
	}
	statusOf := func(id ids.EntityID) string {
		if id == taskID(9) {
			return "pending"
		}
		return "complete"
	}

	got, ok := SelectTask(all, statusOf)
	if !ok {
		t.Fatalf("SelectTask() ok = false, want true")
	}
	if got.Task != taskID(2) {
		t.Errorf("SelectTask() = %+v, want TASK-002 (TASK-001 is blocked)", got)
	}
}

// TestSelectTask_SkipsAlreadyCompleteTask covers that a Task already
// marked complete is never re-selected as "the next Task to start,"
// even though its own dependency-readiness check would trivially pass.
func TestSelectTask_SkipsAlreadyCompleteTask(t *testing.T) {
	all := []TaskInfo{
		{Task: taskID(1), Status: "complete", Dependency: validation.TaskDependency{Task: taskID(1)}},
		{Task: taskID(2), Status: "pending", Dependency: validation.TaskDependency{Task: taskID(2)}},
	}
	statusOf := func(ids.EntityID) string { return "complete" }

	got, ok := SelectTask(all, statusOf)
	if !ok {
		t.Fatalf("SelectTask() ok = false, want true")
	}
	if got.Task != taskID(2) {
		t.Errorf("SelectTask() = %+v, want TASK-002 — TASK-001 is already complete", got)
	}
}

// TestSelectTask_NoneReadyReturnsFalse covers spec FR-012: every Task
// blocked or otherwise unready yields ok=false, never an arbitrary pick.
func TestSelectTask_NoneReadyReturnsFalse(t *testing.T) {
	all := []TaskInfo{
		{Task: taskID(1), Status: "pending", Dependency: validation.TaskDependency{Task: taskID(1), DependsOn: []ids.EntityID{taskID(2)}}},
	}
	statusOf := func(ids.EntityID) string { return "pending" }

	_, ok := SelectTask(all, statusOf)
	if ok {
		t.Errorf("SelectTask() ok = true, want false — no Task is ready")
	}
}

// TestSelectTask_ChecgkedButUnverifiedTaskIsOfferedAgain is
// 041-task-evidence-fingerprint T023 (User Story 2): a Task whose
// checkbox is checked but whose evidence state is Unverified is not
// actually done — SelectTask must offer it again rather than skipping
// it as already complete. No change to selection.go itself
// (research.md #1) — Status == "pending" for this Task already, per
// scan.go's own redefined taskStatus.
func TestSelectTask_CheckedButUnverifiedTaskIsOfferedAgain(t *testing.T) {
	all := []TaskInfo{
		// Status "pending" here simulates a checked-but-Unverified Task
		// — scan.go's own redefined taskStatus, not this test, decides
		// that string; TaskInfo itself has no opinion.
		{Task: taskID(1), Status: "pending", Dependency: validation.TaskDependency{Task: taskID(1)}},
	}
	statusOf := func(ids.EntityID) string { return "complete" }

	got, ok := SelectTask(all, statusOf)
	if !ok {
		t.Fatalf("SelectTask() ok = false, want true — a checked-but-unverified Task is still selectable")
	}
	if got.Task != taskID(1) {
		t.Errorf("SelectTask() = %+v, want TASK-001 offered again", got)
	}
}
