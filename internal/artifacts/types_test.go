package artifacts_test

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestClassifyPath_ByCanonicalLocation(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	tests := []struct {
		relPath string
		want    artifacts.ArtifactType
	}{
		{"ai/programs/PRG-001/program.md", artifacts.TypeProgram},
		{"ai/programs/PRG-001/features/FEAT-004/feature.md", artifacts.TypeFeature},
		{"ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md", artifacts.TypeSpec},
		{"ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/plan.md", artifacts.TypePlan},
		{"ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/tasks.md", artifacts.TypeTasks},
		{"ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/validation.md", artifacts.TypeValidation},
		{"ai/knowledge/KNOW-001-auth.md", artifacts.TypeKnowledge},
		{"ai/memory/learnings/LRN-001-topic.md", artifacts.TypeLearning},
		{"ai/memory/constitution.md", artifacts.TypeConstitution},
	}

	for _, tt := range tests {
		t.Run(tt.relPath, func(t *testing.T) {
			got, err := artifacts.ClassifyPath(root, cfg, tt.relPath)
			if err != nil {
				t.Fatalf("ClassifyPath(%q) unexpected error: %v", tt.relPath, err)
			}
			if got != tt.want {
				t.Errorf("ClassifyPath(%q) = %v, want %v", tt.relPath, got, tt.want)
			}
		})
	}
}

func TestClassifyPath_RejectsTraversalOutsideRoot(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	_, err := artifacts.ClassifyPath(root, cfg, "../outside.md")
	if !errors.Is(err, artifacts.ErrPathOutsideProject) {
		t.Fatalf("ClassifyPath() error = %v, want errors.Is(err, ErrPathOutsideProject)", err)
	}
}

func TestClassifyPath_UnrecognizedLocation(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	_, err := artifacts.ClassifyPath(root, cfg, "ai/somewhere/unexpected.md")
	if err == nil {
		t.Fatal("ClassifyPath() expected an error for an unrecognized canonical location, got nil")
	}
}

func TestArtifactType_HasEntityID(t *testing.T) {
	withID := []artifacts.ArtifactType{artifacts.TypeProgram, artifacts.TypeFeature, artifacts.TypeSpec, artifacts.TypeKnowledge, artifacts.TypeLearning}
	withoutID := []artifacts.ArtifactType{artifacts.TypePlan, artifacts.TypeTasks, artifacts.TypeValidation, artifacts.TypeConstitution}

	for _, typ := range withID {
		if !typ.HasEntityID() {
			t.Errorf("%v.HasEntityID() = false, want true", typ)
		}
	}
	for _, typ := range withoutID {
		if typ.HasEntityID() {
			t.Errorf("%v.HasEntityID() = true, want false", typ)
		}
	}
}
