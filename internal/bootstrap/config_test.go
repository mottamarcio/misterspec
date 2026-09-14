package bootstrap

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/project"
)

func TestWriteDefaultConfig_ReadableByProjectLoad(t *testing.T) {
	root := t.TempDir()

	if err := writeDefaultConfig(root, "claude-code"); err != nil {
		t.Fatalf("writeDefaultConfig() unexpected error: %v", err)
	}

	cfg, err := project.Load(root)
	if err != nil {
		t.Fatalf("project.Load() unexpected error: %v", err)
	}

	if cfg.AgentID != "claude-code" {
		t.Errorf("cfg.AgentID = %q, want %q", cfg.AgentID, "claude-code")
	}
	if cfg.SchemaVersion != 1 {
		t.Errorf("cfg.SchemaVersion = %d, want 1", cfg.SchemaVersion)
	}
	if cfg.ArtifactsDir != project.DefaultArtifactsDir {
		t.Errorf("cfg.ArtifactsDir = %q, want %q", cfg.ArtifactsDir, project.DefaultArtifactsDir)
	}
	if cfg.RawDir != project.DefaultRawDir {
		t.Errorf("cfg.RawDir = %q, want %q", cfg.RawDir, project.DefaultRawDir)
	}
	if cfg.KnowledgeDir != project.DefaultKnowledgeDir {
		t.Errorf("cfg.KnowledgeDir = %q, want %q", cfg.KnowledgeDir, project.DefaultKnowledgeDir)
	}
	if cfg.ConstitutionPath != project.DefaultConstitutionPath {
		t.Errorf("cfg.ConstitutionPath = %q, want %q", cfg.ConstitutionPath, project.DefaultConstitutionPath)
	}
	if cfg.LearningsDir != project.DefaultLearningsDir {
		t.Errorf("cfg.LearningsDir = %q, want %q", cfg.LearningsDir, project.DefaultLearningsDir)
	}
	if cfg.ProgramsRoot != project.DefaultProgramsRoot {
		t.Errorf("cfg.ProgramsRoot = %q, want %q", cfg.ProgramsRoot, project.DefaultProgramsRoot)
	}
	if cfg.IDWidth != project.DefaultIDWidth {
		t.Errorf("cfg.IDWidth = %d, want %d", cfg.IDWidth, project.DefaultIDWidth)
	}
}

func TestWriteDefaultConfig_DetectableAfterWrite(t *testing.T) {
	root := t.TempDir()

	if err := writeDefaultConfig(root, "claude-code"); err != nil {
		t.Fatalf("writeDefaultConfig() unexpected error: %v", err)
	}

	proj, err := project.Detect(root)
	if err != nil {
		t.Fatalf("project.Detect() unexpected error: %v", err)
	}
	if proj.Root != root {
		t.Errorf("proj.Root = %q, want %q", proj.Root, root)
	}
}
