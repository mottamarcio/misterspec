package bootstrap

import (
	"context"
	"errors"
	"io/fs"
	"os"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/installer"
)

// ErrAlreadyInitialized is returned by Bootstrap when targetDir (or an
// ancestor) is already a misterspec project (FR-004, SC-001). Nothing is
// written when this is returned.
var ErrAlreadyInitialized = errors.New("bootstrap: already initialized")

// ErrUnknownAgent is returned by Bootstrap when agentID is not
// registered in registry (FR-007, SC-002). Nothing is written when this
// is returned.
var ErrUnknownAgent = errors.New("bootstrap: unknown agent")

// BootstrapOutcome is the result of one Bootstrap call — the specific
// outcome of its configuration, kit-resource, and agent-install parts,
// never a single pass/fail flag (FR-008, SC-004).
type BootstrapOutcome struct {
	// ProjectRoot is the bootstrapped project's root (targetDir,
	// created if it did not yet exist).
	ProjectRoot string
	// ConfigWritten is whether .misterspec/config.yaml was written
	// successfully.
	ConfigWritten bool
	// TemplateOutcomes is the per-kit-resource result, reused directly
	// from internal/installer.Install.
	TemplateOutcomes []installer.Outcome
	// AgentInstall is the chosen agent's own Install result, reused
	// directly — includes its own per-Skill Outcomes.
	AgentInstall agents.InstallResult
}

// Bootstrap creates a new misterspec project at targetDir (creating the
// directory itself if it does not yet exist), for the given,
// already-chosen agentID: it writes a default project configuration,
// installs the framework's kit resources, and installs Skills for that
// agent via its own registered Adapter.
//
// Bootstrap checks Inspect and registry.Get(agentID) before writing
// anything (FR-004, FR-007, SC-001, SC-002) — an already-initialized
// target or an unregistered agent ID is always rejected before any file
// is touched.
func Bootstrap(targetDir, agentID string, registry *agents.Registry, skills fs.FS) (BootstrapOutcome, error) {
	inspected, err := Inspect(targetDir)
	if err != nil {
		return BootstrapOutcome{}, err
	}
	if inspected.Initialized {
		return BootstrapOutcome{}, ErrAlreadyInitialized
	}

	adapter, ok := registry.Get(agentID)
	if !ok {
		return BootstrapOutcome{}, ErrUnknownAgent
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return BootstrapOutcome{}, err
	}

	outcome := BootstrapOutcome{ProjectRoot: targetDir}

	if err := writeDefaultConfig(targetDir, agentID); err != nil {
		return outcome, err
	}
	outcome.ConfigWritten = true

	templateOutcomes, err := installer.Install(targetDir, false)
	if err != nil {
		return outcome, err
	}
	outcome.TemplateOutcomes = templateOutcomes

	installResult, err := adapter.Install(context.Background(), agents.InstallRequest{
		ProjectRoot: targetDir,
		Skills:      skills,
		Overwrite:   false,
	})
	if err != nil {
		return outcome, err
	}
	outcome.AgentInstall = installResult

	return outcome, nil
}
