package bootstrap

import (
	"errors"
	"os"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/project"
)

// InspectResult is the outcome of inspecting a target directory, without
// writing anything (FR-001, FR-002).
type InspectResult struct {
	// Initialized is false for a directory that is not yet a misterspec
	// project (FR-001); true if targetDir (or an ancestor —
	// project.Detect's own upward-walking semantics, research.md) is
	// already one.
	Initialized bool
	// ProjectRoot is set only when Initialized — the detected project's
	// root, which may differ from targetDir if it is a nested
	// subdirectory of an existing project.
	ProjectRoot string
	// InstalledAgent is set only when Initialized and an agent has
	// actually been installed; empty otherwise (Acceptance Scenario 3).
	InstalledAgent string
	// AgentInstalled mirrors agents.CurrentInstall's own second return
	// value directly — true iff InstalledAgent is meaningful.
	AgentInstalled bool
	// Empty is true when !Initialized and targetDir contains no
	// entries (or does not exist yet at all) — false otherwise,
	// including whenever Initialized is true (an initialized project
	// is never reported empty). Powers the interactive init flow's
	// non-empty-directory warning (010-interactive-init-tui's
	// research.md); added additively — every existing caller is
	// unaffected.
	Empty bool
}

// Inspect determines whether targetDir is already a misterspec project
// and, if so, which agent (if any) is currently installed there. It
// performs no filesystem writes (FR-001, FR-002).
//
// Inspect returns (result, nil) for both "not yet a project" and
// "already a project" — those are not error conditions. It returns
// (InspectResult{}, err) with errors.Is(err, project.ErrInvalidConfiguration)
// only when a project marker exists but its configuration is malformed
// (Edge Case) — kept distinct from "not initialized" the same way
// project.Detect itself keeps ErrNotInitialized and
// ErrInvalidConfiguration distinct.
func Inspect(targetDir string) (InspectResult, error) {
	proj, err := project.Detect(targetDir)
	if err != nil {
		if errors.Is(err, project.ErrNotInitialized) {
			empty, statErr := isEmptyDir(targetDir)
			if statErr != nil {
				return InspectResult{}, statErr
			}
			return InspectResult{Initialized: false, Empty: empty}, nil
		}
		return InspectResult{}, err
	}

	result := InspectResult{
		Initialized: true,
		ProjectRoot: proj.Root,
	}

	record, installed, err := agents.CurrentInstall(proj.Root)
	if err != nil {
		return InspectResult{}, err
	}
	if installed {
		result.AgentInstalled = true
		result.InstalledAgent = record.Agent.ID
	}

	return result, nil
}

// isEmptyDir reports whether dir contains no entries — true also when
// dir does not exist yet at all, since there is nothing there to be
// non-empty. Any other stat/read error is returned as-is.
func isEmptyDir(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	return len(entries) == 0, nil
}
