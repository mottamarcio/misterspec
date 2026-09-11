package artifacts

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mottamarcio/misterspec/internal/project"
)

// ArtifactType identifies which kind of canonical Markdown-with-frontmatter
// file an artifact is. It is a superset of ids.EntityType: Plan, Tasks,
// Validation, and Constitution have no independent ID of their own (see
// docs/architecture-specification.md §22, §28-30), but are still artifacts
// this package classifies and resolves paths for.
type ArtifactType int

const (
	// TypeProgram identifies a Program artifact (program.md).
	TypeProgram ArtifactType = iota
	// TypeFeature identifies a Feature artifact (feature.md).
	TypeFeature
	// TypeSpec identifies a Spec artifact (spec.md).
	TypeSpec
	// TypePlan identifies a Plan artifact (plan.md). No independent ID.
	TypePlan
	// TypeTasks identifies a Tasks artifact (tasks.md). No independent ID.
	TypeTasks
	// TypeValidation identifies a Validation artifact (validation.md).
	// No independent ID.
	TypeValidation
	// TypeKnowledge identifies a Knowledge artifact (KNOW-*.md).
	TypeKnowledge
	// TypeLearning identifies a Learning artifact (LRN-*.md).
	TypeLearning
	// TypeConstitution identifies the project Constitution
	// (constitution.md). No independent ID; exactly one per project.
	TypeConstitution
)

// String returns the human-readable name of t (e.g. "spec"), or "unknown"
// for a value outside the defined ArtifactType range.
func (t ArtifactType) String() string {
	switch t {
	case TypeProgram:
		return "program"
	case TypeFeature:
		return "feature"
	case TypeSpec:
		return "spec"
	case TypePlan:
		return "plan"
	case TypeTasks:
		return "tasks"
	case TypeValidation:
		return "validation"
	case TypeKnowledge:
		return "knowledge"
	case TypeLearning:
		return "learning"
	case TypeConstitution:
		return "constitution"
	default:
		return "unknown"
	}
}

// HasEntityID reports whether artifacts of type t carry an independent
// ids.EntityID (Program, Feature, Spec, Knowledge, Learning) as opposed to
// being addressed only via the Spec/project they belong to (Plan, Tasks,
// Validation, Constitution).
func (t ArtifactType) HasEntityID() bool {
	switch t {
	case TypeProgram, TypeFeature, TypeSpec, TypeKnowledge, TypeLearning:
		return true
	default:
		return false
	}
}

// ClassifyPath determines the ArtifactType of the artifact at path (given
// as absolute, or relative to root), based on its canonical location
// alone — it never reads the file's contents (FR-006). It rejects
// (ErrPathOutsideProject) a path outside root before classifying it
// (FR-008), the same guarantee ResolvePath makes in the other direction.
func ClassifyPath(root string, cfg project.Configuration, path string) (ArtifactType, error) {
	rel, err := relativeWithinRoot(root, path)
	if err != nil {
		return 0, err
	}
	rel = filepath.ToSlash(rel)

	constitutionPath := filepath.ToSlash(cfg.ConstitutionPath)
	knowledgeDir := filepath.ToSlash(cfg.KnowledgeDir)
	learningsDir := filepath.ToSlash(cfg.LearningsDir)
	programsRoot := filepath.ToSlash(cfg.ProgramsRoot)

	switch {
	case rel == constitutionPath:
		return TypeConstitution, nil
	case strings.HasPrefix(rel, knowledgeDir+"/") && strings.HasPrefix(filepath.Base(rel), "KNOW-"):
		return TypeKnowledge, nil
	case strings.HasPrefix(rel, learningsDir+"/") && strings.HasPrefix(filepath.Base(rel), "LRN-"):
		return TypeLearning, nil
	case strings.HasPrefix(rel, programsRoot+"/") && strings.HasSuffix(rel, "/program.md"):
		return TypeProgram, nil
	case strings.HasPrefix(rel, programsRoot+"/") && strings.HasSuffix(rel, "/feature.md"):
		return TypeFeature, nil
	case strings.HasPrefix(rel, programsRoot+"/") && strings.HasSuffix(rel, "/spec.md"):
		return TypeSpec, nil
	case strings.HasPrefix(rel, programsRoot+"/") && strings.HasSuffix(rel, "/plan.md"):
		return TypePlan, nil
	case strings.HasPrefix(rel, programsRoot+"/") && strings.HasSuffix(rel, "/tasks.md"):
		return TypeTasks, nil
	case strings.HasPrefix(rel, programsRoot+"/") && strings.HasSuffix(rel, "/validation.md"):
		return TypeValidation, nil
	default:
		return 0, fmt.Errorf("artifacts: cannot classify %s: no canonical location matches", rel)
	}
}
