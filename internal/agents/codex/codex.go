package codex

import (
	"context"
	"path/filepath"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/installer"
)

// id is this adapter's stable identifier.
const id = "codex"

// name is this adapter's human-readable name.
const name = "Codex CLI"

// targetPath is Codex CLI's own verified Skills directory
// (018-multi-agent-skill-integration/research.md #1: "Codex scans
// .agents/skills in every directory from your current working
// directory up to the repository root" — learn.chatgpt.com/docs/
// build-skills). Note this is the same directory the agy adapter uses
// (research.md #1) — a verified fact, not a bug: both agents currently
// scan the identical project-relative path.
const targetPath = ".agents/skills"

// Adapter implements agents.Adapter for Codex CLI.
type Adapter struct{}

// New constructs the Codex CLI adapter.
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
// claude.Adapter.Install — Codex CLI already scans the same directory-
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
