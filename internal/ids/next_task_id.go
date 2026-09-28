package ids

import "github.com/mottamarcio/misterspec/internal/project"

// NextTaskID computes the next available Task number, project-wide,
// exactly as NextID already does for Program/Feature/Spec — the
// highest existing Task number found in *any* Spec's tasks.md, plus
// one, or 1 if none exist yet (045-global-task-numbering data-model.md
// "NextTaskID", spec FR-001/FR-003/FR-007). It is a pure read: no
// filesystem write, no persisted counter, and no reservation of the
// returned number (research.md Decision 1) — mirroring how Task has
// never had an independent "create" operation of its own
// (kit/skills/mister-tasks/SKILL.md).
func NextTaskID(root string, cfg project.Configuration) (EntityID, error) {
	result, err := Scan(root, cfg, Task)
	if err != nil {
		return EntityID{}, err
	}
	return NextID(result.IDs, Task, cfg.IDWidth), nil
}
