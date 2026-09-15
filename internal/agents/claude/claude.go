package claude

import (
	"context"
	"path/filepath"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/installer"
)

// id is this adapter's stable identifier
// (docs/architecture-specification.md §34's example IDs).
const id = "claude-code"

// name is this adapter's human-readable name.
const name = "Claude Code"

// targetPath is Claude Code's Skill integration location, matching this
// project's own already-established convention (.claude/skills, visible
// in this repository's own .claude/skills/ directory).
const targetPath = ".claude/skills"

// Adapter implements agents.Adapter for Claude Code.
type Adapter struct{}

// New constructs the Claude Code adapter.
func New() agents.Adapter {
	return Adapter{}
}

// ID implements agents.Adapter.
func (Adapter) ID() string { return id }

// Name implements agents.Adapter.
func (Adapter) Name() string { return name }

// TargetPath implements agents.Adapter.
func (Adapter) TargetPath() string { return targetPath }

// Install implements agents.Adapter: materializes req.Skills into
// <req.ProjectRoot>/.claude/skills via internal/installer's proven
// atomic-write, no-silent-overwrite mechanism (no format transformation
// — misterspec's canonical Skill format is already Claude-compatible,
// see specs/006-agent-adapter/spec.md's Assumptions), then records the
// result via agents.RecordInstall.
func (a Adapter) Install(ctx context.Context, req agents.InstallRequest) (agents.InstallResult, error) {
	target := filepath.Join(req.ProjectRoot, targetPath)

	outcomes, err := installer.InstallFS(req.Skills, ".", "skill", target, req.Overwrite)
	if err != nil {
		return agents.InstallResult{}, err
	}

	result := agents.InstallResult{
		AdapterID:       a.ID(),
		IntegrationPath: targetPath,
		Outcomes:        outcomes,
	}

	if err := agents.RecordInstall(req.ProjectRoot, result); err != nil {
		return agents.InstallResult{}, err
	}

	return result, nil
}
