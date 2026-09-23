package validation

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
)

// The tests below reuse the same unexported helpers (specID/taskID/
// reqTestConfig) already defined in requirements_test.go, in this same
// package.

// TestParseTaskDependsOn_SingleEntry covers data-model.md
// "TaskDependency" and research.md Decision 1.
func TestParseTaskDependsOn_SingleEntry(t *testing.T) {
	body := []byte("- [ ] Complete\n\nDepends on: TASK-002\n")
	got := ParseTaskDependsOn(taskID(1), body, reqTestConfig())
	if len(got.DependsOn) != 1 || got.DependsOn[0] != taskID(2) {
		t.Fatalf("DependsOn = %+v, want [TASK-002]", got.DependsOn)
	}
}

// TestParseTaskDependsOn_MultipleEntries covers the comma-separated
// list form, mirroring Serves:'s own grammar.
func TestParseTaskDependsOn_MultipleEntries(t *testing.T) {
	body := []byte("Depends on: TASK-002, TASK-003\n")
	got := ParseTaskDependsOn(taskID(1), body, reqTestConfig())
	if len(got.DependsOn) != 2 {
		t.Fatalf("DependsOn = %+v, want 2 entries", got.DependsOn)
	}
}

// TestParseTaskDependsOn_ExplicitNone covers the explicit "none" value.
func TestParseTaskDependsOn_ExplicitNone(t *testing.T) {
	body := []byte("Depends on: none\n")
	got := ParseTaskDependsOn(taskID(1), body, reqTestConfig())
	if len(got.DependsOn) != 0 {
		t.Errorf("DependsOn = %+v, want empty for an explicit \"none\"", got.DependsOn)
	}
}

// TestParseTaskDependsOn_AbsentLine covers spec FR-003: a Task authored
// before this feature exists, with no Depends on: line at all, is
// treated as having no dependencies.
func TestParseTaskDependsOn_AbsentLine(t *testing.T) {
	body := []byte("- [ ] Complete\n\nServes: SPEC-001:R1\n")
	got := ParseTaskDependsOn(taskID(1), body, reqTestConfig())
	if len(got.DependsOn) != 0 {
		t.Errorf("DependsOn = %+v, want empty when the line is absent", got.DependsOn)
	}
}

// TestDetectTaskDependencyCycles_TwoTaskCycle covers spec FR-008 and
// research.md Decision 2: two Tasks depending on each other produce
// exactly one cycle naming both.
func TestDetectTaskDependencyCycles_TwoTaskCycle(t *testing.T) {
	deps := []TaskDependency{
		{Task: taskID(1), DependsOn: []ids.EntityID{taskID(2)}},
		{Task: taskID(2), DependsOn: []ids.EntityID{taskID(1)}},
	}
	cycles := DetectTaskDependencyCycles(deps)
	if len(cycles) != 1 {
		t.Fatalf("cycles = %+v, want exactly 1", cycles)
	}
}

// TestDetectTaskDependencyCycles_SelfDependency covers the degenerate
// self-dependency case.
func TestDetectTaskDependencyCycles_SelfDependency(t *testing.T) {
	deps := []TaskDependency{
		{Task: taskID(1), DependsOn: []ids.EntityID{taskID(1)}},
	}
	cycles := DetectTaskDependencyCycles(deps)
	if len(cycles) != 1 {
		t.Fatalf("cycles = %+v, want exactly 1", cycles)
	}
}

// TestDetectTaskDependencyCycles_AcyclicNoFalsePositive covers that an
// acyclic Task graph produces none.
func TestDetectTaskDependencyCycles_AcyclicNoFalsePositive(t *testing.T) {
	deps := []TaskDependency{
		{Task: taskID(1), DependsOn: nil},
		{Task: taskID(2), DependsOn: []ids.EntityID{taskID(1)}},
		{Task: taskID(3), DependsOn: []ids.EntityID{taskID(2)}},
	}
	cycles := DetectTaskDependencyCycles(deps)
	if len(cycles) != 0 {
		t.Errorf("cycles = %+v, want none", cycles)
	}
}

// TestInvalidTaskDependencies_DanglingReference covers spec FR-009: a
// Depends on: entry naming a nonexistent Task number in the same Spec.
func TestInvalidTaskDependencies_DanglingReference(t *testing.T) {
	deps := []TaskDependency{
		{Task: taskID(1), DependsOn: []ids.EntityID{taskID(9)}},
	}
	invalid := InvalidTaskDependencies(deps)
	if len(invalid) != 1 || invalid[0].Missing != taskID(9) {
		t.Fatalf("invalid = %+v, want exactly one entry naming TASK-009", invalid)
	}
}

// TestInvalidTaskDependencies_CrossSpecRejectedAtParseTime covers spec
// FR-009's other case: a composite cross-Spec reference is rejected at
// parse time (captured in MalformedDependsOn), never reaching
// InvalidTaskDependencies as a "valid same-Spec reference."
func TestInvalidTaskDependencies_CrossSpecRejectedAtParseTime(t *testing.T) {
	body := []byte("Depends on: SPEC-002:TASK-001\n")
	got := ParseTaskDependsOn(taskID(1), body, reqTestConfig())
	if len(got.DependsOn) != 0 {
		t.Errorf("DependsOn = %+v, want none — cross-Spec is invalid", got.DependsOn)
	}
	if len(got.MalformedDependsOn) != 1 {
		t.Fatalf("MalformedDependsOn = %v, want exactly one entry", got.MalformedDependsOn)
	}
}

// TestParseTaskDependsOn_MalformedEntryCaptured covers that a malformed
// entry is captured, not silently dropped (mirrors TaskCoverage's own
// MalformedReferences).
func TestParseTaskDependsOn_MalformedEntryCaptured(t *testing.T) {
	body := []byte("Depends on: not-a-task\n")
	got := ParseTaskDependsOn(taskID(1), body, reqTestConfig())
	if len(got.DependsOn) != 0 {
		t.Errorf("DependsOn = %+v, want none for a malformed entry", got.DependsOn)
	}
	if len(got.MalformedDependsOn) != 1 || got.MalformedDependsOn[0] != "not-a-task" {
		t.Fatalf("MalformedDependsOn = %v, want [\"not-a-task\"]", got.MalformedDependsOn)
	}
}
