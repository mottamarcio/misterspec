// This file compiles and runs specs/003-entity-creation/quickstart.md's
// end-to-end flow (Create a Program, a Feature under it, a Spec under
// that, then a Plan for the Spec via CreateArtifact) against
// internal/operations, on top of the fixture project the other
// quickstart tests in this package use (plan.md Phase 6, T014).
package example

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestCreationQuickstart_EndToEnd(t *testing.T) {
	root := testutil.Project(t)

	proj, err := project.Detect(root)
	if err != nil {
		t.Fatalf("project.Detect() unexpected error: %v", err)
	}

	prg, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("operations.Create(Program) unexpected error: %v", err)
	}
	if prg.ID.String() != "PRG-001" {
		t.Fatalf("Create(Program).ID = %v, want PRG-001", prg.ID)
	}

	feat, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{
		Type:   ids.Feature,
		Parent: prg.ID.String(),
	})
	if err != nil {
		t.Fatalf("operations.Create(Feature) unexpected error: %v", err)
	}

	spec, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{
		Type:   ids.Spec,
		Parent: feat.ID.String(),
	})
	if err != nil {
		t.Fatalf("operations.Create(Spec) unexpected error: %v", err)
	}

	plan, err := operations.CreateArtifact(proj.Root, proj.Config, operations.CreateArtifactRequest{
		Kind: artifacts.TypePlan,
		For:  spec.ID.String(),
	})
	if err != nil {
		t.Fatalf("operations.CreateArtifact(Plan) unexpected error: %v", err)
	}

	// Every newly created artifact must be immediately, correctly
	// inspectable (FR-013).
	inspected, err := operations.Inspect(proj.Root, proj.Config, spec.ID.String())
	if err != nil {
		t.Fatalf("operations.Inspect(%v) unexpected error: %v", spec.ID, err)
	}
	if inspected.Metadata.Parent == nil || inspected.Metadata.Parent.String() != feat.ID.String() {
		t.Fatalf("Inspect(%v).Metadata.Parent = %v, want %v", spec.ID, inspected.Metadata.Parent, feat.ID)
	}
	if plan.Path == "" {
		t.Fatal("CreateArtifact(Plan).Path is empty")
	}
}
