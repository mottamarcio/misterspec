package copilot_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/agents/copilot"
	"github.com/mottamarcio/misterspec/internal/installer"
)

func fixtureSkills() fstest.MapFS {
	return fstest.MapFS{
		"create-knowledge-base.md": {Data: []byte("# create-knowledge-base\n")},
		"create-plan.md":           {Data: []byte("# create-plan\n")},
	}
}

func TestCopilot_Metadata(t *testing.T) {
	a := copilot.New()
	if a.ID() != "copilot" {
		t.Errorf("ID() = %q, want %q", a.ID(), "copilot")
	}
	if a.Name() == "" {
		t.Error("Name() is empty, want a human-readable name")
	}
	if a.TargetPath() != ".github/skills" {
		t.Errorf("TargetPath() = %q, want %q", a.TargetPath(), ".github/skills")
	}
}

func TestCopilot_InstallFreshDirectory(t *testing.T) {
	root := t.TempDir()
	a := copilot.New()

	result, err := a.Install(context.Background(), agents.InstallRequest{
		ProjectRoot: root,
		Skills:      fixtureSkills(),
		Overwrite:   false,
	})
	if err != nil {
		t.Fatalf("Install() unexpected error: %v", err)
	}
	if result.AdapterID != "copilot" {
		t.Errorf("result.AdapterID = %q, want %q", result.AdapterID, "copilot")
	}
	if result.IntegrationPath != ".github/skills" {
		t.Errorf("result.IntegrationPath = %q, want %q", result.IntegrationPath, ".github/skills")
	}
	if len(result.Outcomes) != 2 {
		t.Fatalf("result.Outcomes = %d entries, want 2", len(result.Outcomes))
	}
	for _, o := range result.Outcomes {
		if o.Status != installer.Installed {
			t.Errorf("Outcome for %q Status = %v, want Installed (Err: %v)", o.Resource.Name, o.Status, o.Err)
		}
		got, err := os.ReadFile(filepath.Join(root, ".github", "skills", o.Resource.Name))
		if err != nil {
			t.Fatalf("reading installed skill %s: %v", o.Resource.Name, err)
		}
		want, _ := fixtureSkills().ReadFile(o.Resource.Name)
		if string(got) != string(want) {
			t.Errorf("installed skill %s not byte-identical to fixture source", o.Resource.Name)
		}
	}

	recordPath := filepath.Join(root, ".misterspec", "install.json")
	data, err := os.ReadFile(recordPath)
	if err != nil {
		t.Fatalf("reading install.json: %v", err)
	}
	var record agents.InstallRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("unmarshaling install.json: %v", err)
	}
	if record.SchemaVersion != agents.SchemaVersion {
		t.Errorf("record.SchemaVersion = %d, want %d", record.SchemaVersion, agents.SchemaVersion)
	}
	if record.Agent.ID != "copilot" || record.Agent.IntegrationPath != ".github/skills" {
		t.Errorf("record.Agent = %+v, unexpected", record.Agent)
	}
}

func TestCopilot_InstallWithoutOverwriteSkipsAlreadyPresent(t *testing.T) {
	root := t.TempDir()
	a := copilot.New()
	ctx := context.Background()

	if _, err := a.Install(ctx, agents.InstallRequest{ProjectRoot: root, Skills: fixtureSkills(), Overwrite: false}); err != nil {
		t.Fatalf("first Install() unexpected error: %v", err)
	}

	result, err := a.Install(ctx, agents.InstallRequest{ProjectRoot: root, Skills: fixtureSkills(), Overwrite: false})
	if err != nil {
		t.Fatalf("second Install() unexpected error: %v", err)
	}
	for _, o := range result.Outcomes {
		if o.Status != installer.Skipped {
			t.Errorf("Outcome for %q Status = %v, want Skipped", o.Resource.Name, o.Status)
		}
	}
}
