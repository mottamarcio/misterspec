package bootstrap_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/bootstrap"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestInspect_UninitializedDirectory(t *testing.T) {
	root := t.TempDir()

	result, err := bootstrap.Inspect(root)
	if err != nil {
		t.Fatalf("Inspect() unexpected error: %v", err)
	}
	if result.Initialized {
		t.Error("Inspect() Initialized = true, want false for an empty directory")
	}
	if !result.Empty {
		t.Error("Inspect() Empty = false, want true for an empty directory (010-interactive-init-tui)")
	}
}

func TestInspect_NonExistentDirectory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "does-not-exist-yet")

	result, err := bootstrap.Inspect(root)
	if err != nil {
		t.Fatalf("Inspect() unexpected error: %v", err)
	}
	if result.Initialized {
		t.Error("Inspect() Initialized = true, want false for a non-existent directory")
	}
	if !result.Empty {
		t.Error("Inspect() Empty = false, want true for a non-existent directory")
	}
}

func TestInspect_NonEmptyUninitializedDirectory(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root, "README.md", "# an existing, unrelated repository\n")

	result, err := bootstrap.Inspect(root)
	if err != nil {
		t.Fatalf("Inspect() unexpected error: %v", err)
	}
	if result.Initialized {
		t.Error("Inspect() Initialized = true, want false — no .misterspec/config.yaml here")
	}
	if result.Empty {
		t.Error("Inspect() Empty = true, want false — the directory already contains other files")
	}
}

func TestInspect_AlreadyInitializedDirectoryIsNeverReportedEmpty(t *testing.T) {
	root := testutil.Project(t)

	result, err := bootstrap.Inspect(root)
	if err != nil {
		t.Fatalf("Inspect() unexpected error: %v", err)
	}
	if !result.Initialized {
		t.Fatal("Inspect() Initialized = false, want true")
	}
	if result.Empty {
		t.Error("Inspect() Empty = true, want false — an initialized project is never reported empty")
	}
}

func TestInspect_AlreadyInitializedWithAgentInstalled(t *testing.T) {
	root := testutil.Project(t)
	if err := agents.RecordInstall(root, agents.InstallResult{
		AdapterID:       "claude-code",
		IntegrationPath: ".claude/skills",
	}); err != nil {
		t.Fatalf("RecordInstall() unexpected error: %v", err)
	}

	result, err := bootstrap.Inspect(root)
	if err != nil {
		t.Fatalf("Inspect() unexpected error: %v", err)
	}
	if !result.Initialized {
		t.Fatal("Inspect() Initialized = false, want true")
	}
	if result.ProjectRoot != root {
		t.Errorf("Inspect() ProjectRoot = %q, want %q", result.ProjectRoot, root)
	}
	if !result.AgentInstalled {
		t.Error("Inspect() AgentInstalled = false, want true")
	}
	if result.InstalledAgent != "claude-code" {
		t.Errorf("Inspect() InstalledAgent = %q, want %q", result.InstalledAgent, "claude-code")
	}
}

func TestInspect_AlreadyInitializedNoAgentInstalled(t *testing.T) {
	root := testutil.Project(t)

	result, err := bootstrap.Inspect(root)
	if err != nil {
		t.Fatalf("Inspect() unexpected error: %v", err)
	}
	if !result.Initialized {
		t.Fatal("Inspect() Initialized = false, want true")
	}
	if result.AgentInstalled {
		t.Error("Inspect() AgentInstalled = true, want false for a project with no install.json")
	}
	if result.InstalledAgent != "" {
		t.Errorf("Inspect() InstalledAgent = %q, want empty", result.InstalledAgent)
	}
}

func TestInspect_InvalidConfigurationIsDistinctFromUninitialized(t *testing.T) {
	root := t.TempDir()
	testutil.WriteConfig(t, root, "not: [valid yaml")

	_, err := bootstrap.Inspect(root)
	if err == nil {
		t.Fatal("Inspect() error = nil, want an error wrapping project.ErrInvalidConfiguration")
	}
	if !errors.Is(err, project.ErrInvalidConfiguration) {
		t.Errorf("Inspect() error = %v, want errors.Is(err, project.ErrInvalidConfiguration)", err)
	}
}
