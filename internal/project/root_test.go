package project_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestDetect_ValidRoot(t *testing.T) {
	root := testutil.Project(t)

	proj, err := project.Detect(root)
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}

	wantRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	if proj.Root != wantRoot {
		t.Errorf("Root = %q, want %q", proj.Root, wantRoot)
	}
	if proj.Config.AgentID != "claude-code" {
		t.Errorf("Config.AgentID = %q, want %q", proj.Config.AgentID, "claude-code")
	}
}

func TestDetect_NestedSubdirectory(t *testing.T) {
	root := testutil.Project(t)
	nested := testutil.Subdir(t, root, 4)

	proj, err := project.Detect(nested)
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}

	wantRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	if proj.Root != wantRoot {
		t.Errorf("Root = %q, want %q (same as detecting from the root itself)", proj.Root, wantRoot)
	}
}

func TestDetect_NotInitialized(t *testing.T) {
	dir := t.TempDir()

	_, err := project.Detect(dir)
	if !errors.Is(err, project.ErrNotInitialized) {
		t.Fatalf("Detect() error = %v, want errors.Is(err, ErrNotInitialized)", err)
	}
}

func TestDetect_InvalidConfiguration(t *testing.T) {
	root := t.TempDir()
	testutil.WriteConfig(t, root, "agent_id: claude-code\n") // missing schema_version

	_, err := project.Detect(root)
	if !errors.Is(err, project.ErrInvalidConfiguration) {
		t.Fatalf("Detect() error = %v, want errors.Is(err, ErrInvalidConfiguration)", err)
	}
	// Must be reported distinctly from "not initialized" — a project *was*
	// found, it's just broken.
	if errors.Is(err, project.ErrNotInitialized) {
		t.Errorf("Detect() error unexpectedly also matches ErrNotInitialized: %v", err)
	}
}
