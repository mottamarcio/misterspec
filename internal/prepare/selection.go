package prepare

import (
	"sort"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/validation"
)

// ComputeReadinessWithCycleCheck reports task's own readiness, treating
// it as blocked — never infinite-looping — when it participates in any
// dependency cycle among all's own Tasks (research.md Decision 2), by
// reusing validation.DetectTaskDependencyCycles rather than a second,
// duplicated cycle-detection algorithm. Otherwise it defers to the
// ordinary per-dependency completion check (computeTaskReadiness).
func ComputeReadinessWithCycleCheck(task ids.EntityID, dependsOn []ids.EntityID, statusOf func(ids.EntityID) string, all []TaskInfo) TaskReadiness {
	deps := make([]validation.TaskDependency, 0, len(all))
	for _, info := range all {
		deps = append(deps, info.Dependency)
	}

	for _, cycle := range validation.DetectTaskDependencyCycles(deps) {
		for _, id := range cycle {
			if id == task {
				return TaskReadiness{Task: task, Ready: false, Blockers: dependsOn}
			}
		}
	}

	return computeTaskReadiness(task, dependsOn, statusOf)
}

// SelectTask returns the lowest-numbered Task among all whose computed
// readiness is Ready, and true — or the zero TaskInfo and false when no
// Task qualifies (034/research.md Decision 6, spec FR-011/FR-012).
// Recomputed fresh from all/statusOf every call; nothing is cached.
func SelectTask(all []TaskInfo, statusOf func(ids.EntityID) string) (TaskInfo, bool) {
	sorted := append([]TaskInfo(nil), all...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Task.Number < sorted[j].Task.Number })

	for _, info := range sorted {
		if info.Status == "complete" {
			// Already done — never re-selected as "the next Task to
			// start" (spec FR-011/FR-012 are about what to work on
			// next, not what is already finished).
			continue
		}
		readiness := ComputeReadinessWithCycleCheck(info.Task, info.Dependency.DependsOn, statusOf, all)
		if readiness.Ready {
			return info, true
		}
	}
	return TaskInfo{}, false
}
