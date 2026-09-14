package bootstrap_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/bootstrap"
	"github.com/mottamarcio/misterspec/internal/installer"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

// fakeAdapter is a minimal agents.Adapter used only to test Bootstrap —
// never the real claude package, keeping this feature's tests
// independent of a concrete adapter's own behavior (research.md). It
// still performs a real, atomic install (reusing internal/installer and
// agents.RecordInstall directly) so Bootstrap's composition is exercised
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

func fixtureSkills() fstest.MapFS {
	return fstest.MapFS{
		"skill-one.md": {Data: []byte("# skill-one\n")},
	}
}

func fixtureRegistry() *agents.Registry {
	return agents.NewRegistry(fakeAdapter{id: "fake-agent", target: ".fake/skills"})
}

func TestBootstrap_FreshNonExistentDirectory(t *testing.T) {
	target := filepath.Join(t.TempDir(), "newproject")

	outcome, err := bootstrap.Bootstrap(target, "fake-agent", fixtureRegistry(), fixtureSkills())
	if err != nil {
		t.Fatalf("Bootstrap() unexpected error: %v", err)
	}

	if outcome.ProjectRoot != target {
		t.Errorf("outcome.ProjectRoot = %q, want %q", outcome.ProjectRoot, target)
	}
	if !outcome.ConfigWritten {
		t.Error("outcome.ConfigWritten = false, want true")
	}
	if len(outcome.TemplateOutcomes) == 0 {
		t.Fatal("outcome.TemplateOutcomes is empty, want at least one kit resource")
	}
	for _, o := range outcome.TemplateOutcomes {
		if o.Status != installer.Installed {
			t.Errorf("template outcome for %q Status = %v, want Installed (Err: %v)", o.Resource.Name, o.Status, o.Err)
		}
	}
	if outcome.AgentInstall.AdapterID != "fake-agent" {
		t.Errorf("outcome.AgentInstall.AdapterID = %q, want %q", outcome.AgentInstall.AdapterID, "fake-agent")
	}
	if len(outcome.AgentInstall.Outcomes) != 1 || outcome.AgentInstall.Outcomes[0].Status != installer.Installed {
		t.Errorf("outcome.AgentInstall.Outcomes = %+v, want one Installed entry", outcome.AgentInstall.Outcomes)
	}

	// FR-010: the result must be detectable via the project's own
	// existing detection capability.
	if _, err := project.Detect(target); err != nil {
		t.Errorf("project.Detect(target) after Bootstrap() unexpected error: %v", err)
	}
}

func TestBootstrap_RejectsAlreadyInitializedTarget(t *testing.T) {
	root := testutil.Project(t)
	before, err := os.ReadFile(filepath.Join(root, ".misterspec", "config.yaml"))
	if err != nil {
		t.Fatalf("reading fixture config.yaml: %v", err)
	}

	_, err = bootstrap.Bootstrap(root, "fake-agent", fixtureRegistry(), fixtureSkills())
	if !errors.Is(err, bootstrap.ErrAlreadyInitialized) {
		t.Fatalf("Bootstrap() error = %v, want errors.Is(err, bootstrap.ErrAlreadyInitialized)", err)
	}

	after, err := os.ReadFile(filepath.Join(root, ".misterspec", "config.yaml"))
	if err != nil {
		t.Fatalf("reading config.yaml after rejected Bootstrap(): %v", err)
	}
	if string(before) != string(after) {
		t.Error("config.yaml was modified by a rejected Bootstrap() call, want untouched")
	}
	if _, err := os.Stat(filepath.Join(root, ".fake", "skills")); !os.IsNotExist(err) {
		t.Error("Bootstrap() installed Skills into an already-initialized target, want nothing written")
	}
}

func TestBootstrap_RejectsUnregisteredAgent(t *testing.T) {
	target := filepath.Join(t.TempDir(), "newproject2")

	_, err := bootstrap.Bootstrap(target, "no-such-agent", fixtureRegistry(), fixtureSkills())
	if !errors.Is(err, bootstrap.ErrUnknownAgent) {
		t.Fatalf("Bootstrap() error = %v, want errors.Is(err, bootstrap.ErrUnknownAgent)", err)
	}

	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Error("Bootstrap() created target directory before rejecting an unregistered agent, want nothing written")
	}
}
