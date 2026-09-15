package copilot

import (
	"context"
	"path/filepath"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/installer"
)

// id is this adapter's stable identifier.
const id = "copilot"

// name is this adapter's human-readable name.
const name = "GitHub Copilot"

// targetPath is GitHub Copilot's own verified Skills directory
// (018-multi-agent-skill-integration/research.md #1: GitHub Docs'
// "Adding agent skills for GitHub Copilot" — "create a .github/skills
// ... directory ... each skill should have its own directory ... with
// a SKILL.md file"), matching Spec Kit's own default choice rather
// than its optional, secondary flat-prompt-file mode.
const targetPath = ".github/skills"

// Adapter implements agents.Adapter for GitHub Copilot.
type Adapter struct{}

// New constructs the GitHub Copilot adapter.
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
// <req.ProjectRoot>/.github/skills byte-for-byte, identical to
// claude.Adapter.Install — GitHub Copilot already scans the same
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
