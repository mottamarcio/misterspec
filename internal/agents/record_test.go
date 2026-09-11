package agents_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/agents"
)

func TestRecordInstall_WritesExpectedSchema(t *testing.T) {
	root := t.TempDir()

	err := agents.RecordInstall(root, agents.InstallResult{
		AdapterID:       "claude-code",
		IntegrationPath: ".claude/skills",
	})
	if err != nil {
		t.Fatalf("RecordInstall() unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".misterspec", "install.json"))
	if err != nil {
		t.Fatalf("reading install.json: %v", err)
	}
	var record agents.InstallRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("unmarshaling install.json: %v", err)
	}
	if record.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d, want 1", record.SchemaVersion)
	}
	if record.MisterspecVersion != agents.FrameworkVersion {
		t.Errorf("MisterspecVersion = %q, want %q", record.MisterspecVersion, agents.FrameworkVersion)
	}
	if record.Agent.ID != "claude-code" || record.Agent.IntegrationPath != ".claude/skills" {
		t.Errorf("Agent = %+v, unexpected", record.Agent)
	}
}

func TestCurrentInstall_AfterRealInstall(t *testing.T) {
	root := t.TempDir()

	if err := agents.RecordInstall(root, agents.InstallResult{AdapterID: "claude-code", IntegrationPath: ".claude/skills"}); err != nil {
		t.Fatalf("RecordInstall() unexpected error: %v", err)
	}

	record, installed, err := agents.CurrentInstall(root)
	if err != nil {
		t.Fatalf("CurrentInstall() unexpected error: %v", err)
	}
	if !installed {
		t.Fatal("CurrentInstall() installed = false, want true")
	}
	if record.Agent.ID != "claude-code" || record.Agent.IntegrationPath != ".claude/skills" {
		t.Errorf("CurrentInstall() record.Agent = %+v, unexpected", record.Agent)
	}
}

func TestCurrentInstall_NeverInstalled(t *testing.T) {
	root := t.TempDir()

	_, installed, err := agents.CurrentInstall(root)
	if err != nil {
		t.Fatalf("CurrentInstall() unexpected error: %v", err)
	}
	if installed {
		t.Fatal("CurrentInstall() installed = true, want false for a project that was never installed")
	}
}
