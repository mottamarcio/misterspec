package tui

import (
	"context"
	"path/filepath"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/installer"
)

// fakeAdapter is a minimal agents.Adapter used only by this package's
// own tests — never the real claude package (mirrors
// 007-project-bootstrap's bootstrap_test.go fakeAdapter). It still
// performs a real, atomic install (reusing internal/installer and
// agents.RecordInstall directly) so a full happy-path flow is exercised
// end to end, not just with a stub that records a call.
type fakeAdapter struct {
	id, target string
}

func (f fakeAdapter) ID() string         { return f.id }
func (f fakeAdapter) Name() string       { return "Fake Agent" }
func (f fakeAdapter) TargetPath() string { return f.target }
func (f fakeAdapter) Install(ctx context.Context, req agents.InstallRequest) (agents.InstallResult, error) {
	outcomes, err := installer.InstallFS(req.Skills, ".", "skill", filepath.Join(req.ProjectRoot, f.target), req.Overwrite)
	if err != nil {
		return agents.InstallResult{}, err
	}
	result := agents.InstallResult{AdapterID: f.id, IntegrationPath: f.target, Outcomes: outcomes}
	if err := agents.RecordInstall(req.ProjectRoot, result); err != nil {
		return agents.InstallResult{}, err
	}
	return result, nil
}
