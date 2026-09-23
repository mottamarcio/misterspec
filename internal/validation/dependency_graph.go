package validation

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// DependencyGraph is the project-wide Spec depends_on graph, rebuilt
// fresh on every validation run — nothing here is persisted (Constitution
// Principle III, 032/data-model.md "DependencyGraph").
type DependencyGraph struct {
	// Nodes is every Spec number found project-wide.
	Nodes []int
	// Edges maps a Spec number to the Spec numbers it depends_on,
	// filtered to entries of Type == ids.Spec (a non-Spec depends_on
	// entry is already flagged elsewhere by checkDependencyList).
	Edges map[int][]int
}

// DependencyCycle is one detected cycle, including the degenerate
// self-reference case (spec FR-009, data-model.md "DependencyCycle").
type DependencyCycle struct {
	// Path is the full ordered cycle, starting and ending at the same
	// Spec number (e.g. [1, 2, 3, 1]); a self-reference is [1, 1].
	Path []int
}

// buildDependencyGraph constructs the project-wide DependencyGraph from
// every Spec's own Metadata.DependsOn, already resolved by the caller
// into a specNumber -> []ids.EntityID lookup (032/data-model.md
// "DependencyGraph").
func buildDependencyGraph(specNumbers []int, dependsOn func(specNumber int) []ids.EntityID) DependencyGraph {
	g := DependencyGraph{Nodes: append([]int(nil), specNumbers...), Edges: map[int][]int{}}
	for _, n := range specNumbers {
		for _, entry := range dependsOn(n) {
			if entry.Type == ids.Spec {
				g.Edges[n] = append(g.Edges[n], entry.Number)
			}
		}
	}
	return g
}

// color is a node's DFS state for cycle detection (research.md
// Decision 5's white/gray/black walk).
type color int

const (
	white color = iota
	gray
	black
)

// detectCycles walks g with a DFS/three-color algorithm, returning one
// DependencyCycle per distinct cycle found — a back-edge to a node
// currently on the recursion stack (gray) closes a cycle, reported as
// the full stack slice from that node to the current one, inclusive
// (032/research.md Decision 5). Deterministic: Nodes and each node's own
// edge list are walked in sorted order, so the same graph always
// produces cycles in the same order.
func detectCycles(g DependencyGraph) []DependencyCycle {
	colors := map[int]color{}
	var stack []int
	var cycles []DependencyCycle

	nodes := append([]int(nil), g.Nodes...)
	sort.Ints(nodes)

	var visit func(n int)
	visit = func(n int) {
		colors[n] = gray
		stack = append(stack, n)

		edges := append([]int(nil), g.Edges[n]...)
		sort.Ints(edges)
		for _, next := range edges {
			switch colors[next] {
			case white:
				visit(next)
			case gray:
				// Found a back-edge to `next`, still on the stack —
				// the cycle is everything from `next`'s position to
				// here, plus `next` again to close the loop.
				idx := indexOf(stack, next)
				path := append([]int(nil), stack[idx:]...)
				path = append(path, next)
				cycles = append(cycles, DependencyCycle{Path: path})
			case black:
				// Already fully explored via another path — not a
				// cycle back to the current recursion stack.
			}
		}

		stack = stack[:len(stack)-1]
		colors[n] = black
	}

	for _, n := range nodes {
		if colors[n] == white {
			visit(n)
		}
	}

	return cycles
}

func indexOf(list []int, n int) int {
	for i, v := range list {
		if v == n {
			return i
		}
	}
	return -1
}

// projectDependencyGraph builds the project-wide DependencyGraph from
// every discovered Spec's own spec.md Metadata.DependsOn, and returns
// each Spec number's own spec.md path alongside it — both are needed to
// render a dependency_cycle Finding (contracts §2 step 5).
func projectDependencyGraph(root string, cfg project.Configuration) (DependencyGraph, map[int]string, error) {
	result, err := ids.Scan(root, cfg, ids.Spec)
	if err != nil {
		return DependencyGraph{}, nil, err
	}

	numbers := sortedNumbers(result.Paths)
	specPath := map[int]string{}
	for _, n := range numbers {
		paths := result.Paths[n]
		if len(paths) == 0 {
			continue
		}
		specPath[n] = paths[0] + "/spec.md"
	}

	g := buildDependencyGraph(numbers, func(n int) []ids.EntityID {
		p, ok := specPath[n]
		if !ok {
			return nil
		}
		meta, err := artifacts.ParseMetadata(filepath.Join(root, p))
		if err != nil {
			return nil
		}
		return meta.DependsOn
	})

	return g, specPath, nil
}

// dependencyCycleFinding renders one DependencyCycle as a Finding, per
// contracts §1's "dependency_cycle" row — Path is the first Spec in the
// cycle's own spec.md; Message lists the full ordered cycle.
func dependencyCycleFinding(cycle DependencyCycle, cfg project.Configuration, specPath map[int]string) Finding {
	rendered := make([]string, len(cycle.Path))
	for i, n := range cycle.Path {
		rendered[i] = ids.EntityID{Type: ids.Spec, Prefix: ids.Spec.Prefix(), Number: n, Width: cfg.IDWidth}.String()
	}
	return Finding{
		Code:     CodeDependencyCycle,
		Severity: SeverityError,
		Path:     specPath[cycle.Path[0]],
		Message:  fmt.Sprintf("dependency cycle: %s", strings.Join(rendered, " -> ")),
	}
}
