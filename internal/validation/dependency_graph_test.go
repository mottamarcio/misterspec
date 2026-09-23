package validation

import "testing"

// TestDetectCycles_ThreeNodeCycle covers 032/data-model.md
// "DependencyGraph"/"DependencyCycle" and research.md Decision 5: a
// 3-node cycle returns one DependencyCycle naming the full path.
func TestDetectCycles_ThreeNodeCycle(t *testing.T) {
	g := DependencyGraph{
		Nodes: []int{1, 2, 3},
		Edges: map[int][]int{1: {2}, 2: {3}, 3: {1}},
	}
	cycles := detectCycles(g)
	if len(cycles) != 1 {
		t.Fatalf("detectCycles() = %+v, want exactly 1 cycle", cycles)
	}
	path := cycles[0].Path
	if len(path) != 4 || path[0] != path[len(path)-1] {
		t.Fatalf("cycle Path = %v, want a closed loop of length 4 (start == end)", path)
	}
}

// TestDetectCycles_SelfReference covers spec FR-009's degenerate case.
func TestDetectCycles_SelfReference(t *testing.T) {
	g := DependencyGraph{
		Nodes: []int{1},
		Edges: map[int][]int{1: {1}},
	}
	cycles := detectCycles(g)
	if len(cycles) != 1 {
		t.Fatalf("detectCycles() = %+v, want exactly 1 cycle", cycles)
	}
	if got := cycles[0].Path; len(got) != 2 || got[0] != 1 || got[1] != 1 {
		t.Errorf("self-reference Path = %v, want [1 1]", got)
	}
}

// TestDetectCycles_AcyclicGraphNoFalsePositive covers spec Acceptance
// Scenario 3: a deep, branching acyclic graph produces zero cycles.
func TestDetectCycles_AcyclicGraphNoFalsePositive(t *testing.T) {
	g := DependencyGraph{
		Nodes: []int{1, 2, 3, 4, 5},
		Edges: map[int][]int{1: {2, 3}, 2: {4}, 3: {4}, 4: {5}},
	}
	cycles := detectCycles(g)
	if len(cycles) != 0 {
		t.Errorf("detectCycles() = %+v, want none", cycles)
	}
}

// TestDetectCycles_MultipleIndependentCycles covers that separate
// cycles in the same graph are each detected.
func TestDetectCycles_MultipleIndependentCycles(t *testing.T) {
	g := DependencyGraph{
		Nodes: []int{1, 2, 3, 4},
		Edges: map[int][]int{1: {2}, 2: {1}, 3: {4}, 4: {3}},
	}
	cycles := detectCycles(g)
	if len(cycles) != 2 {
		t.Fatalf("detectCycles() = %+v, want exactly 2 independent cycles", cycles)
	}
}
