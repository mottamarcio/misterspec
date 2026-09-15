package operations_test

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestCreateArtifact_PlanTasksValidationForExistingSpec(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\n")

	plan, err := operations.CreateArtifact(root, cfg, operations.CreateArtifactRequest{Kind: artifacts.TypePlan, For: "SPEC-001"})
	if err != nil {
		t.Fatalf("CreateArtifact(Plan) unexpected error: %v", err)
	}
	wantPlanPath := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/plan.md"
	if plan.Path != wantPlanPath {
		t.Errorf("Plan.Path = %q, want %q", plan.Path, wantPlanPath)
	}

	tasks, err := operations.CreateArtifact(root, cfg, operations.CreateArtifactRequest{Kind: artifacts.TypeTasks, For: "SPEC-001"})
	if err != nil {
		t.Fatalf("CreateArtifact(Tasks) unexpected error: %v", err)
	}
	wantTasksPath := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md"
	if tasks.Path != wantTasksPath {
		t.Errorf("Tasks.Path = %q, want %q", tasks.Path, wantTasksPath)
	}

	validation, err := operations.CreateArtifact(root, cfg, operations.CreateArtifactRequest{Kind: artifacts.TypeValidation, For: "SPEC-001"})
	if err != nil {
		t.Fatalf("CreateArtifact(Validation) unexpected error: %v", err)
	}
	wantValidationPath := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/validation.md"
	if validation.Path != wantValidationPath {
		t.Errorf("Validation.Path = %q, want %q", validation.Path, wantValidationPath)
	}
}

func TestCreateArtifact_NonexistentSpecRejected(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	_, err := operations.CreateArtifact(root, cfg, operations.CreateArtifactRequest{Kind: artifacts.TypeTasks, For: "SPEC-999"})
	if !errors.Is(err, operations.ErrInvalidParent) {
		t.Fatalf("CreateArtifact() error = %v, want errors.Is(err, ErrInvalidParent)", err)
	}
}

func TestCreateArtifact_AlreadyExistsRejected(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\n")

	if _, err := operations.CreateArtifact(root, cfg, operations.CreateArtifactRequest{Kind: artifacts.TypeValidation, For: "SPEC-001"}); err != nil {
		t.Fatalf("first CreateArtifact(Validation) unexpected error: %v", err)
	}

	_, err := operations.CreateArtifact(root, cfg, operations.CreateArtifactRequest{Kind: artifacts.TypeValidation, For: "SPEC-001"})
	if !errors.Is(err, operations.ErrAlreadyExists) {
		t.Fatalf("second CreateArtifact(Validation) error = %v, want errors.Is(err, ErrAlreadyExists)", err)
	}
}

func TestCreateArtifact_UnsupportedKindRejected(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\n")

	_, err := operations.CreateArtifact(root, cfg, operations.CreateArtifactRequest{Kind: artifacts.TypeSpec, For: "SPEC-001"})
	if !errors.Is(err, operations.ErrUnsupportedType) {
		t.Fatalf("CreateArtifact() error = %v, want errors.Is(err, ErrUnsupportedType)", err)
	}
}
