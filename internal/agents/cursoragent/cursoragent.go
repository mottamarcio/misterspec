// Package cursoragent implements agents.Adapter for Cursor. Named
// cursoragent, not cursor-agent (not a valid Go identifier), for the
// cursor-agent adapter ID — matching the id constant below
// (018-multi-agent-skill-integration/research.md #4).
package cursoragent

import (
	"context"
	"path/filepath"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/installer"
)

// id is this adapter's stable identifier — the user's own explicit
// choice for this feature.
const id = "cursor-agent"

// name is this adapter's human-readable name.
const name = "Cursor"

// targetPath is Cursor's own verified Skills directory
// (018-multi-agent-skill-integration/research.md #1: GitHub Spec Kit's
// own integrations reference — "cursor-agent | .cursor/skills |
// Skills-based integration; installs skills into .cursor/skills").
const targetPath = ".cursor/skills"

// Adapter implements agents.Adapter for Cursor.
type Adapter struct{}

// New constructs the Cursor adapter.
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
// <req.ProjectRoot>/.cursor/skills byte-for-byte, identical to
// claude.Adapter.Install — Cursor already scans the same directory-
// per-skill SKILL.md "Agent Skills" convention misterspec's own
// canonical Skills already use, so no format transformation is needed
// (research.md #1, #2).
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
