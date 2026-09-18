package prepare

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
)

func taskID(n int) ids.EntityID {
	return ids.EntityID{Type: ids.Task, Prefix: "TASK", Number: n, Width: 3}
}

// TestComputeTaskReadiness_EmptyDependsOnIsReady covers data-model.md
// "TaskReadiness": a Task with no declared dependency is trivially
// Ready.
func TestComputeTaskReadiness_EmptyDependsOnIsReady(t *testing.T) {
	got := computeTaskReadiness(taskID(1), nil, func(ids.EntityID) string { return "pending" })
	if !got.Ready {
		t.Errorf("Ready = false, want true for an empty DependsOn")
	}
	if len(got.Blockers) != 0 {
		t.Errorf("Blockers = %v, want none", got.Blockers)
	}
}

// TestComputeTaskReadiness_AllDependenciesComplete covers the normal
// ready case.
func TestComputeTaskReadiness_AllDependenciesComplete(t *testing.T) {
	deps := []ids.EntityID{taskID(1), taskID(2)}
	got := computeTaskReadiness(taskID(3), deps, func(ids.EntityID) string { return "complete" })
	if !got.Ready {
		t.Errorf("Ready = false, want true when every dependency is complete")
	}
}

// TestComputeTaskReadiness_IncompleteDependencyBlocks covers spec
// FR-006/FR-007: Blockers names exactly the incomplete dependencies.
func TestComputeTaskReadiness_IncompleteDependencyBlocks(t *testing.T) {
	deps := []ids.EntityID{taskID(1), taskID(2)}
	statusOf := func(id ids.EntityID) string {
		if id == taskID(1) {
			return "complete"
		}
		return "pending"
	}
	got := computeTaskReadiness(taskID(3), deps, statusOf)
	if got.Ready {
		t.Fatalf("Ready = true, want false — TASK-002 is still pending")
	}
	if len(got.Blockers) != 1 || got.Blockers[0] != taskID(2) {
		t.Errorf("Blockers = %+v, want exactly [TASK-002]", got.Blockers)
	}
}
