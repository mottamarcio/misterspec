package prepare

import "github.com/mottamarcio/misterspec/internal/ids"

// TaskReadiness is the computed answer to "can this Task start now"
// (034/data-model.md "TaskReadiness").
type TaskReadiness struct {
	Task     ids.EntityID
	Ready    bool
	Blockers []ids.EntityID
}

// computeTaskReadiness reports task's own readiness given dependsOn
// (its own "Depends on:" declaration) and statusOf, a lookup returning
// each dependency's own Status ("complete"/"pending") — a pure function
// so it can be tested without any filesystem access, and so a detected
// dependency cycle (034-task-oriented-context-preparation research.md
// Decision 2) can be folded in later by the caller supplying a statusOf
// that never reports "complete" for a cycling Task, rather than this
// function needing its own graph-walking logic. An empty dependsOn is
// trivially Ready (spec FR-003).
func computeTaskReadiness(task ids.EntityID, dependsOn []ids.EntityID, statusOf func(ids.EntityID) string) TaskReadiness {
	readiness := TaskReadiness{Task: task, Ready: true}
	for _, dep := range dependsOn {
		if statusOf(dep) != "complete" {
			readiness.Ready = false
			readiness.Blockers = append(readiness.Blockers, dep)
		}
	}
	return readiness
}
