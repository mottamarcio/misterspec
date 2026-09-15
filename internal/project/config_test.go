package project_test

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestLoad_Valid(t *testing.T) {
	root := testutil.Project(t)

	cfg, err := project.Load(root)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	want := project.Configuration{
		SchemaVersion:    1,
		AgentID:          "claude-code",
		ArtifactsDir:     "ai",
		RawDir:           "ai/raw",
		KnowledgeDir:     "ai/knowledge",
		ConstitutionPath: "ai/memory/constitution.md",
		LearningsDir:     "ai/memory/learnings",
		ProgramsRoot:     "ai/programs",
		IDWidth:          3,
	}
	if cfg != want {
		t.Errorf("Load() = %+v, want %+v", cfg, want)
	}
}

func TestLoad_DefaultsAppliedWhenFieldsOmitted(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteConfig(t, root, "schema_version: 1\nagent_id: claude-code\n")

	cfg, err := project.Load(root)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.ArtifactsDir != project.DefaultArtifactsDir {
		t.Errorf("ArtifactsDir = %q, want default %q", cfg.ArtifactsDir, project.DefaultArtifactsDir)
	}
	if cfg.IDWidth != project.DefaultIDWidth {
		t.Errorf("IDWidth = %d, want default %d", cfg.IDWidth, project.DefaultIDWidth)
	}
}

func TestLoad_MissingSchemaVersion(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteConfig(t, root, "agent_id: claude-code\n")

	_, err := project.Load(root)
	assertInvalidField(t, err, "schema_version")
}

func TestLoad_EmptyRequiredDirectoryField(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteConfig(t, root, "schema_version: 1\nagent_id: claude-code\nartifacts_dir: \"\"\n")

	_, err := project.Load(root)
	assertInvalidField(t, err, "artifacts_dir")
}

func TestLoad_NonPositiveIDWidth(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteConfig(t, root, "schema_version: 1\nagent_id: claude-code\nid_width: 0\n")

	_, err := project.Load(root)
	assertInvalidField(t, err, "id_width")
}

func TestLoad_MalformedYAML(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteConfig(t, root, "schema_version: [this is not valid yaml\n")

	_, err := project.Load(root)
	if !errors.Is(err, project.ErrInvalidConfiguration) {
		t.Fatalf("Load() error = %v, want errors.Is(err, ErrInvalidConfiguration)", err)
	}
}

func assertInvalidField(t *testing.T, err error, field string) {
	t.Helper()

	if !errors.Is(err, project.ErrInvalidConfiguration) {
		t.Fatalf("error = %v, want errors.Is(err, ErrInvalidConfiguration)", err)
	}
	var cfgErr *project.ConfigError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("error = %v, want a *project.ConfigError", err)
	}
	if cfgErr.Field != field {
		t.Errorf("ConfigError.Field = %q, want %q", cfgErr.Field, field)
	}
}
