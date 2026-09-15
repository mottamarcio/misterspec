package operations

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/lock"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/templates"
)

// CreateArtifactRequest is the input to CreateArtifact (FR-010).
type CreateArtifactRequest struct {
	// Kind must be TypePlan, TypeTasks, or TypeValidation.
	Kind artifacts.ArtifactType
	// For is the raw Spec ID this artifact belongs to.
	For string
}

// CreateArtifactResult is the outcome of a successful CreateArtifact.
type CreateArtifactResult struct {
	Path string // relative to root
}

var subordinateFilenames = map[artifacts.ArtifactType]string{
	artifacts.TypePlan:       "plan.md",
	artifacts.TypeTasks:      "tasks.md",
	artifacts.TypeValidation: "validation.md",
}

// CreateArtifact atomically writes req.Kind's artifact at its fixed
// location under the Spec named by req.For — no independent ID is
// allocated, since Plan/Tasks/Validation have none (FR-010..FR-012).
//
// Unlike Create, CreateArtifact does not call Create — both depend only
// on the Foundational lock/templates packages and on 002-read-operations's
// Resolve, so the two stories are independent of each other
// (specs/003-entity-creation/tasks.md's Dependencies section).
func CreateArtifact(root string, cfg project.Configuration, req CreateArtifactRequest) (CreateArtifactResult, error) {
	filename, ok := subordinateFilenames[req.Kind]
	if !ok {
		return CreateArtifactResult{}, fmt.Errorf("%w: %v", ErrUnsupportedType, req.Kind)
	}

	l, err := lock.Acquire(root, lock.DefaultStaleAfter, lock.DefaultTimeout)
	if err != nil {
		return CreateArtifactResult{}, err
	}
	defer l.Release()

	spec, err := Resolve(root, cfg, req.For)
	if err != nil {
		return CreateArtifactResult{}, fmt.Errorf("%w: %v", ErrInvalidParent, err)
	}
	if spec.ID.Type != ids.Spec {
		return CreateArtifactResult{}, fmt.Errorf("%w: %s is a %v, want a spec", ErrInvalidParent, req.For, spec.ID.Type)
	}

	specDir := filepath.Dir(spec.Path)
	targetRel := specDir + "/" + filename
	targetAbs := filepath.Join(root, targetRel)

	if _, statErr := os.Stat(targetAbs); statErr == nil {
		return CreateArtifactResult{}, fmt.Errorf("%w: %s", ErrAlreadyExists, targetRel)
	} else if !os.IsNotExist(statErr) {
		return CreateArtifactResult{}, statErr
	}

	content, err := renderCreateArtifact(req.Kind, req.For)
	if err != nil {
		return CreateArtifactResult{}, err
	}

	if err := writeAtomic(targetAbs, content); err != nil {
		return CreateArtifactResult{}, err
	}

	return CreateArtifactResult{Path: targetRel}, nil
}

func renderCreateArtifact(kind artifacts.ArtifactType, forSpecID string) (string, error) {
	switch kind {
	case artifacts.TypePlan:
		return templates.Render(templates.Plan, templates.PlanData{For: forSpecID})
	case artifacts.TypeTasks:
		return templates.Render(templates.Tasks, templates.TasksData{For: forSpecID})
	case artifacts.TypeValidation:
		return templates.Render(templates.Validation, templates.ValidationData{For: forSpecID})
	default:
		return "", fmt.Errorf("%w: %v", ErrUnsupportedType, kind)
	}
}
