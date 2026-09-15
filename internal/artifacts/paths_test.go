package artifacts_test

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

func testConfig() project.Configuration {
	return project.Configuration{
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
}

func mustID(t *testing.T, typ ids.EntityType, number int) ids.EntityID {
	t.Helper()
	return ids.EntityID{Type: typ, Prefix: typ.Prefix(), Number: number, Width: 3}
}

func TestResolvePath_CanonicalLayout(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig()

	prg := mustID(t, ids.Program, 1)
	feat := mustID(t, ids.Feature, 4)
	spec := mustID(t, ids.Spec, 14)

	tests := []struct {
		name     string
		typ      ids.EntityType
		id       ids.EntityID
		parents  []ids.EntityID
		wantDir  string
		wantFile string
	}{
		{"program", ids.Program, prg, nil, "ai/programs/PRG-001", "ai/programs/PRG-001/program.md"},
		{"feature", ids.Feature, feat, []ids.EntityID{prg}, "ai/programs/PRG-001/features/FEAT-004", "ai/programs/PRG-001/features/FEAT-004/feature.md"},
		{"spec", ids.Spec, spec, []ids.EntityID{prg, feat}, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014", "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := artifacts.ResolvePath(root, cfg, tt.typ, tt.id, tt.parents...)
			if err != nil {
				t.Fatalf("ResolvePath() unexpected error: %v", err)
			}
			if got.Directory != tt.wantDir {
				t.Errorf("Directory = %q, want %q", got.Directory, tt.wantDir)
			}
			if got.File != tt.wantFile {
				t.Errorf("File = %q, want %q", got.File, tt.wantFile)
			}
		})
	}
}

func TestResolvePath_Deterministic(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig()
	spec := mustID(t, ids.Spec, 14)
	prg := mustID(t, ids.Program, 1)
	feat := mustID(t, ids.Feature, 4)

	first, err := artifacts.ResolvePath(root, cfg, ids.Spec, spec, prg, feat)
	if err != nil {
		t.Fatalf("ResolvePath() unexpected error: %v", err)
	}
	second, err := artifacts.ResolvePath(root, cfg, ids.Spec, spec, prg, feat)
	if err != nil {
		t.Fatalf("ResolvePath() unexpected error: %v", err)
	}
	if first != second {
		t.Errorf("ResolvePath() not deterministic: %+v != %+v", first, second)
	}
}

func TestResolvePath_KnowledgeAndLearningResolveDirectoryOnly(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig()

	know, err := artifacts.ResolvePath(root, cfg, ids.Knowledge, mustID(t, ids.Knowledge, 1))
	if err != nil {
		t.Fatalf("ResolvePath(Knowledge) unexpected error: %v", err)
	}
	if know.Directory != "ai/knowledge" || know.File != "" {
		t.Errorf("ResolvePath(Knowledge) = %+v, want Directory=ai/knowledge, File=\"\" (exact filename includes an agent-chosen slug, requires a directory scan)", know)
	}

	lrn, err := artifacts.ResolvePath(root, cfg, ids.Learning, mustID(t, ids.Learning, 1))
	if err != nil {
		t.Fatalf("ResolvePath(Learning) unexpected error: %v", err)
	}
	if lrn.Directory != "ai/memory/learnings" || lrn.File != "" {
		t.Errorf("ResolvePath(Learning) = %+v, want Directory=ai/memory/learnings, File=\"\"", lrn)
	}
}

func TestResolvePath_RejectsTraversalOutsideRoot(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig()
	cfg.ProgramsRoot = "../../etc" // a malformed/malicious configuration value

	_, err := artifacts.ResolvePath(root, cfg, ids.Program, mustID(t, ids.Program, 1))
	if !errors.Is(err, artifacts.ErrPathOutsideProject) {
		t.Fatalf("ResolvePath() error = %v, want errors.Is(err, ErrPathOutsideProject)", err)
	}
}

func TestResolvePath_MissingParent(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig()

	if _, err := artifacts.ResolvePath(root, cfg, ids.Feature, mustID(t, ids.Feature, 1)); err == nil {
		t.Error("ResolvePath(Feature, no parent) expected an error, got nil")
	}
	if _, err := artifacts.ResolvePath(root, cfg, ids.Spec, mustID(t, ids.Spec, 1), mustID(t, ids.Program, 1)); err == nil {
		t.Error("ResolvePath(Spec, only one parent) expected an error, got nil")
	}
}
