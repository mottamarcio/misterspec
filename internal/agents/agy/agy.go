package agy

import (
	"context"
	"path/filepath"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/installer"
)

// id is this adapter's stable identifier — the user's own explicit
// choice for this feature (018-multi-agent-skill-integration/
// research.md #1), not docs/architecture-specification.md §34's own
// illustrative "antigravity" example ID.
const id = "agy"

// name is this adapter's human-readable name.
const name = "Antigravity"

// targetPath is Antigravity's own verified project-level Skills
// directory (research.md #1: Google Antigravity Docs' own Skills
// codelab, "<project-root>/.agents/skills/").
const targetPath = ".agents/skills"

// Adapter implements agents.Adapter for Antigravity.
type Adapter struct{}

// New constructs the Antigravity adapter.
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
// <req.ProjectRoot>/.agents/skills byte-for-byte, identical to
// claude.Adapter.Install — Antigravity already scans the same
// directory-per-skill SKILL.md "Agent Skills" convention misterspec's
// own canonical Skills already use, so no format transformation is
// needed (research.md #1, #2).
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
