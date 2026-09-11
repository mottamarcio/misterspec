package validation_test

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
	"github.com/mottamarcio/misterspec/internal/validation"
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

func findingCodes(findings []validation.Finding) []string {
	var codes []string
	for _, f := range findings {
		codes = append(codes, f.Code)
	}
	return codes
}

func containsCode(findings []validation.Finding, code string) bool {
	for _, f := range findings {
		if f.Code == code {
			return true
		}
	}
	return false
}

func TestValidateEntity_WellFormedSpec(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md", "---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("ValidateEntity() findings = %v, want empty", findings)
	}
}

func TestValidateEntity_MissingParent(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-999\n---\n")

	findings, err := validation.ValidateEntity(root, cfg, "FEAT-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeMissingParent) {
		t.Errorf("ValidateEntity() findings = %v, want %q", findingCodes(findings), validation.CodeMissingParent)
	}
}

func TestValidateEntity_WrongParentType(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	// FEAT-001's own parent field wrongly names another Feature-shaped ID.
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-002/feature.md", "---\nid: FEAT-002\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: FEAT-002\n---\n")

	findings, err := validation.ValidateEntity(root, cfg, "FEAT-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeInvalidParentType) {
		t.Errorf("ValidateEntity() findings = %v, want %q", findingCodes(findings), validation.CodeInvalidParentType)
	}
}

func TestValidateEntity_InvalidStatus(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: not-a-real-status\n---\n")

	findings, err := validation.ValidateEntity(root, cfg, "PRG-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeInvalidStatus) {
		t.Errorf("ValidateEntity() findings = %v, want %q", findingCodes(findings), validation.CodeInvalidStatus)
	}
}

func TestValidateEntity_UnresolvedDependency(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md", "---\nid: SPEC-001\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on:\n  - SPEC-999\n---\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeUnresolvedDependency) {
		t.Errorf("ValidateEntity() findings = %v, want %q", findingCodes(findings), validation.CodeUnresolvedDependency)
	}
}

func TestValidateEntity_MultipleSimultaneousProblems(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: not-a-real-status\nparent: PRG-999\n---\n")

	findings, err := validation.ValidateEntity(root, cfg, "FEAT-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeInvalidStatus) || !containsCode(findings, validation.CodeMissingParent) {
		t.Errorf("ValidateEntity() findings = %v, want both %q and %q", findingCodes(findings), validation.CodeInvalidStatus, validation.CodeMissingParent)
	}
}

func TestValidateEntity_NotFound(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-999")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeNotFound) {
		t.Errorf("ValidateEntity() findings = %v, want %q", findingCodes(findings), validation.CodeNotFound)
	}
}

func TestValidateEntity_DuplicateID(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-002/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: active\nparent: PRG-002\n---\n")

	findings, err := validation.ValidateEntity(root, cfg, "FEAT-004")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeDuplicateID) {
		t.Errorf("ValidateEntity() findings = %v, want %q", findingCodes(findings), validation.CodeDuplicateID)
	}
}

func TestValidateEntity_MalformedRawIDIsError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	_, err := validation.ValidateEntity(root, cfg, "not-an-id")
	if err == nil {
		t.Fatal("ValidateEntity() expected an error for a malformed rawID, got nil")
	}
}

func TestValidateEntity_UnsupportedTargetIsError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	_, err := validation.ValidateEntity(root, cfg, "TASK-001")
	if err == nil {
		t.Fatal("ValidateEntity() expected an error for an unsupported target type (Task), got nil")
	}
}

func TestValidateProject_AllClean(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md", "---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("ValidateProject() findings = %v, want empty", findings)
	}
}

func TestValidateProject_AggregatesIndividualProblems(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	// One clean Program.
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	// One Feature with an invalid status.
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: bogus\nparent: PRG-001\n---\n")
	// One Spec with a missing parent.
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md", "---\nid: SPEC-001\ntype: spec\nstatus: draft\nparent: FEAT-999\n---\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeInvalidStatus) {
		t.Errorf("ValidateProject() findings = %v, want %q present", findingCodes(findings), validation.CodeInvalidStatus)
	}
	if !containsCode(findings, validation.CodeMissingParent) {
		t.Errorf("ValidateProject() findings = %v, want %q present", findingCodes(findings), validation.CodeMissingParent)
	}
}

func TestValidateProject_DuplicateIDAcrossPrograms(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-002/features/FEAT-004/feature.md", "---\nid: FEAT-004\ntype: feature\nstatus: active\nparent: PRG-002\n---\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeDuplicateID) {
		t.Errorf("ValidateProject() findings = %v, want %q present", findingCodes(findings), validation.CodeDuplicateID)
	}
}

func TestValidateProject_DuplicateTaskID(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — First\n\n- [ ] Complete\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/tasks.md",
		"---\ntype: tasks\nfor: SPEC-002\n---\n# Tasks\n\n## TASK-001 — Also first, oops\n\n- [ ] Complete\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeDuplicateID) {
		t.Errorf("ValidateProject() findings = %v, want %q present for duplicate Task IDs", findingCodes(findings), validation.CodeDuplicateID)
	}
}

func TestValidateProject_EmptyProject(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("ValidateProject() findings = %v, want empty for a fresh project", findings)
	}
}
