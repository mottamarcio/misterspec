package validation_test

import (
	"strings"
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

// TestValidateProject_SameTaskNumberAcrossDifferentSpecsIsNotADuplicate
// covers 031-canonical-task-identity spec.md User Story 2: two different
// Specs each legitimately numbering their first task TASK-001 is valid
// per-Spec numbering, not a project-wide duplicate. (Prior to
// 031-canonical-task-identity, this scenario was incorrectly flagged as
// CodeDuplicateID — see that feature's research.md Decision 2.)
func TestValidateProject_SameTaskNumberAcrossDifferentSpecsIsNotADuplicate(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — First in Spec 1\n\n- [ ] Complete\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/tasks.md",
		"---\ntype: tasks\nfor: SPEC-002\n---\n# Tasks\n\n## TASK-001 — First in Spec 2\n\n- [ ] Complete\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if containsCode(findings, validation.CodeDuplicateID) {
		t.Errorf("ValidateProject() findings = %v, want no %q — different Specs sharing a Task number is valid", findingCodes(findings), validation.CodeDuplicateID)
	}
}

// TestValidateProject_DuplicateTaskIDWithinSameSpec covers spec.md User
// Story 2: two "## TASK-001" headings inside the same Spec's tasks.md
// must still be flagged, scoped to that Spec.
func TestValidateProject_DuplicateTaskIDWithinSameSpec(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — First\n\n- [ ] Complete\n\n## TASK-001 — Also first, oops\n\n- [ ] Complete\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	var dupFindings []validation.Finding
	for _, f := range findings {
		if f.Code == validation.CodeDuplicateID {
			dupFindings = append(dupFindings, f)
		}
	}
	if len(dupFindings) != 1 {
		t.Fatalf("ValidateProject() CodeDuplicateID findings = %+v, want exactly 1", dupFindings)
	}
	if !strings.Contains(dupFindings[0].Message, "SPEC-001") {
		t.Errorf("ValidateProject() duplicate finding message = %q, want it to name SPEC-001", dupFindings[0].Message)
	}
}

// TestValidateEntity_SpecScopedTaskDuplicate covers spec.md FR-011:
// validating a single Spec must surface that Spec's own Task duplicate,
// the same way project-wide validation would.
func TestValidateEntity_SpecScopedTaskDuplicate(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — First\n\n- [ ] Complete\n\n## TASK-001 — Also first, oops\n\n- [ ] Complete\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeDuplicateID) {
		t.Errorf("ValidateEntity() findings = %v, want %q present", findingCodes(findings), validation.CodeDuplicateID)
	}
}

// TestValidateEntity_SpecScopedTaskDuplicate_DifferentSpecUnaffected
// covers FR-011's negative case: a duplicate Task in another Spec must
// not leak into a different Spec's own validation.
func TestValidateEntity_SpecScopedTaskDuplicate_DifferentSpecUnaffected(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Only one\n\n- [ ] Complete\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/tasks.md",
		"---\ntype: tasks\nfor: SPEC-002\n---\n# Tasks\n\n## TASK-001 — First\n\n- [ ] Complete\n\n## TASK-001 — Also first, oops\n\n- [ ] Complete\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if containsCode(findings, validation.CodeDuplicateID) {
		t.Errorf("ValidateEntity(SPEC-001) findings = %v, want no %q — the duplicate belongs to SPEC-002", findingCodes(findings), validation.CodeDuplicateID)
	}
}

// TestValidateProject_UncoveredRequirement covers
// 032-requirement-coverage-dependency-validation spec.md Acceptance
// Scenario 1: a requirement with no covering Task is reported; a
// covered one is not.
func TestValidateProject_UncoveredRequirement(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n# SPEC-001\n\n## Requirements\n\n### R1 — First\n\n### R2 — Second\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Cover R1\n\n- [ ] Complete\n\nServes: SPEC-001:R1\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeUncoveredRequirement) {
		t.Errorf("ValidateProject() findings = %v, want %q present for R2", findingCodes(findings), validation.CodeUncoveredRequirement)
	}
	for _, f := range findings {
		if f.Code == validation.CodeUncoveredRequirement && !strings.Contains(f.Message, "R2") {
			t.Errorf("uncovered_requirement finding = %q, want it to name R2", f.Message)
		}
	}
}

// TestValidateProject_UnknownRequirementReference covers Acceptance
// Scenario 2: a Serves: reference to a nonexistent requirement number.
func TestValidateProject_UnknownRequirementReference(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n# SPEC-001\n\n## Requirements\n\n### R1 — First\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Cover R1\n\n- [ ] Complete\n\nServes: SPEC-001:R1\n\n## TASK-002 — Bad reference\n\n- [ ] Complete\n\nServes: SPEC-001:R9\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeUnknownRequirementReference) {
		t.Errorf("ValidateProject() findings = %v, want %q present", findingCodes(findings), validation.CodeUnknownRequirementReference)
	}
}

// TestValidateProject_CrossSpecRequirementReference covers Acceptance
// Scenario 3: a Serves: reference naming a different Spec than the
// Task's own owning Spec is invalid, and the target requirement still
// shows as uncovered from its own Spec's perspective.
func TestValidateProject_CrossSpecRequirementReference(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n# SPEC-001\n\n## Requirements\n\n### R1 — First\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — References the wrong Spec\n\n- [ ] Complete\n\nServes: SPEC-002:R1\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n# SPEC-002\n\n## Requirements\n\n### R1 — First\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeCrossSpecRequirementReference) {
		t.Errorf("ValidateProject() findings = %v, want %q present", findingCodes(findings), validation.CodeCrossSpecRequirementReference)
	}
	if !containsCode(findings, validation.CodeUncoveredRequirement) {
		t.Errorf("ValidateProject() findings = %v, want %q present for SPEC-002's own R1 (cross-Spec reference doesn't count as coverage)", findingCodes(findings), validation.CodeUncoveredRequirement)
	}
}

// TestValidateProject_TaskWithoutRequirement covers Acceptance Scenario 4.
func TestValidateProject_TaskWithoutRequirement(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n# SPEC-001\n\n## Requirements\n\n### R1 — First\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — No coverage line\n\n- [ ] Complete\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeTaskWithoutRequirement) {
		t.Errorf("ValidateProject() findings = %v, want %q present", findingCodes(findings), validation.CodeTaskWithoutRequirement)
	}
}

// TestValidateProject_FullyCoveredSpecHasNoCoverageFindings covers
// Acceptance Scenario 5.
func TestValidateProject_FullyCoveredSpecHasNoCoverageFindings(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n# SPEC-001\n\n## Requirements\n\n### R1 — First\n\n### R2 — Second\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Cover R1\n\n- [ ] Complete\n\nServes: SPEC-001:R1\n\n## TASK-002 — Cover R2\n\n- [ ] Complete\n\nServes: SPEC-001:R2\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	coverageCodes := map[string]bool{
		validation.CodeUncoveredRequirement:          true,
		validation.CodeUnknownRequirementReference:   true,
		validation.CodeCrossSpecRequirementReference: true,
		validation.CodeTaskWithoutRequirement:        true,
		validation.CodeDuplicateRequirementID:        true,
	}
	for _, f := range findings {
		if coverageCodes[f.Code] {
			t.Errorf("ValidateProject() unexpected coverage finding: %+v", f)
		}
	}
}

// TestValidateProject_DuplicateRequirementID covers spec.md FR-002,
// Edge Cases.
func TestValidateProject_DuplicateRequirementID(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n# SPEC-001\n\n## Requirements\n\n### R1 — First\n\n### R1 — Duplicated by mistake\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeDuplicateRequirementID) {
		t.Errorf("ValidateProject() findings = %v, want %q present", findingCodes(findings), validation.CodeDuplicateRequirementID)
	}
}

// TestValidateEntity_RequirementCoverageParity covers spec.md FR-013:
// validating a single Spec surfaces the same coverage findings
// ValidateProject would.
func TestValidateEntity_RequirementCoverageParity(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n# SPEC-001\n\n## Requirements\n\n### R1 — First\n\n### R2 — Second\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Cover R1\n\n- [ ] Complete\n\nServes: SPEC-001:R1\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeUncoveredRequirement) {
		t.Errorf("ValidateEntity() findings = %v, want %q present", findingCodes(findings), validation.CodeUncoveredRequirement)
	}
}

// TestValidateProject_DependencyCycle covers
// 032-requirement-coverage-dependency-validation spec.md Acceptance
// Scenario 1: three Specs forming a cycle produce one finding naming
// the full path, and an unrelated Spec still validates normally.
func TestValidateProject_DependencyCycle(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on:\n  - SPEC-002\nsupersedes: []\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on:\n  - SPEC-003\nsupersedes: []\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-003/spec.md",
		"---\nid: SPEC-003\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on:\n  - SPEC-001\nsupersedes: []\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-004/spec.md",
		"---\nid: SPEC-004\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	var cycleFindings []validation.Finding
	for _, f := range findings {
		if f.Code == validation.CodeDependencyCycle {
			cycleFindings = append(cycleFindings, f)
		}
	}
	if len(cycleFindings) != 1 {
		t.Fatalf("ValidateProject() dependency_cycle findings = %+v, want exactly 1", cycleFindings)
	}
	for _, spec := range []string{"SPEC-001", "SPEC-002", "SPEC-003"} {
		if !strings.Contains(cycleFindings[0].Message, spec) {
			t.Errorf("dependency_cycle message = %q, want it to name %s", cycleFindings[0].Message, spec)
		}
	}
}

// TestValidateProject_SelfReferencingDependencyCycle covers Acceptance
// Scenario 2.
func TestValidateProject_SelfReferencingDependencyCycle(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on:\n  - SPEC-001\nsupersedes: []\n---\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeDependencyCycle) {
		t.Errorf("ValidateProject() findings = %v, want %q present", findingCodes(findings), validation.CodeDependencyCycle)
	}
}

// TestValidateProject_AcyclicBranchingGraphNoCycle covers Acceptance
// Scenario 3.
func TestValidateProject_AcyclicBranchingGraphNoCycle(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on:\n  - SPEC-002\n  - SPEC-003\nsupersedes: []\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on:\n  - SPEC-004\nsupersedes: []\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-003/spec.md",
		"---\nid: SPEC-003\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on:\n  - SPEC-004\nsupersedes: []\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-004/spec.md",
		"---\nid: SPEC-004\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if containsCode(findings, validation.CodeDependencyCycle) {
		t.Errorf("ValidateProject() findings = %v, want no %q", findingCodes(findings), validation.CodeDependencyCycle)
	}
}

// TestValidateEntity_DependencyCycleParity covers Acceptance Scenario 4
// / contracts §3: validating a Spec that participates in a cycle
// reports the same cycle project-wide validation would; a Spec not on
// any cycle produces no such finding when validated alone.
func TestValidateEntity_DependencyCycleParity(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on:\n  - SPEC-002\nsupersedes: []\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on:\n  - SPEC-003\nsupersedes: []\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-003/spec.md",
		"---\nid: SPEC-003\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on:\n  - SPEC-001\nsupersedes: []\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-004/spec.md",
		"---\nid: SPEC-004\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-002")
	if err != nil {
		t.Fatalf("ValidateEntity(SPEC-002) unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeDependencyCycle) {
		t.Errorf("ValidateEntity(SPEC-002) findings = %v, want %q present", findingCodes(findings), validation.CodeDependencyCycle)
	}

	unaffected, err := validation.ValidateEntity(root, cfg, "SPEC-004")
	if err != nil {
		t.Fatalf("ValidateEntity(SPEC-004) unexpected error: %v", err)
	}
	if containsCode(unaffected, validation.CodeDependencyCycle) {
		t.Errorf("ValidateEntity(SPEC-004) findings = %v, want no %q — SPEC-004 is not on any cycle", findingCodes(unaffected), validation.CodeDependencyCycle)
	}
}

// TestValidateProject_DraftSpecExemptFromPhaseGate covers
// 032-requirement-coverage-dependency-validation spec.md Acceptance
// Scenario 1: a draft Spec with no tasks.md at all is not blocked, and
// the missing tasks.md does not error.
func TestValidateProject_DraftSpecExemptFromPhaseGate(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: draft\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n# SPEC-001\n\n## Requirements\n\n### R1 — First\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if containsCode(findings, validation.CodePhaseGateBlocked) {
		t.Errorf("ValidateProject() findings = %v, want no %q for a draft Spec", findingCodes(findings), validation.CodePhaseGateBlocked)
	}
}

// TestValidateProject_ReadySpecBlockedOnIncompleteCoverage covers
// Acceptance Scenario 2.
func TestValidateProject_ReadySpecBlockedOnIncompleteCoverage(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n# SPEC-001\n\n## Requirements\n\n### R1 — First\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodePhaseGateBlocked) {
		t.Errorf("ValidateProject() findings = %v, want %q present", findingCodes(findings), validation.CodePhaseGateBlocked)
	}
	if !containsCode(findings, validation.CodeUncoveredRequirement) {
		t.Errorf("ValidateProject() findings = %v, want the underlying %q still present too", findingCodes(findings), validation.CodeUncoveredRequirement)
	}
}

// TestValidateProject_ReadySpecFullyCoveredNoPhaseGate covers
// Acceptance Scenario 3.
func TestValidateProject_ReadySpecFullyCoveredNoPhaseGate(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n# SPEC-001\n\n## Requirements\n\n### R1 — First\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Cover R1\n\n- [ ] Complete\n\nServes: SPEC-001:R1\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if containsCode(findings, validation.CodePhaseGateBlocked) {
		t.Errorf("ValidateProject() findings = %v, want no %q", findingCodes(findings), validation.CodePhaseGateBlocked)
	}
}

func TestValidateProject_ConstitutionAbsent(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	// A project with other clean artifacts but no Constitution at all.
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("ValidateProject() findings = %v, want empty when Constitution does not exist yet", findings)
	}
}

func TestValidateProject_ConstitutionFrontmatterMalformed(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, cfg.ConstitutionPath, "# Project Constitution\n\nNo frontmatter at all.\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeFrontmatterMalformed) {
		t.Errorf("ValidateProject() findings = %v, want %q present", findingCodes(findings), validation.CodeFrontmatterMalformed)
	}
}

func TestValidateProject_ConstitutionSchemaVersionMissing(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, cfg.ConstitutionPath, "---\ntype: constitution\n---\n# Project Constitution\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeRequiredFieldMissing) {
		t.Errorf("ValidateProject() findings = %v, want %q present", findingCodes(findings), validation.CodeRequiredFieldMissing)
	}
}

func TestValidateProject_ConstitutionWellFormed(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, cfg.ConstitutionPath, "---\ntype: constitution\nschema_version: 1\n---\n# Project Constitution\n")

	findings, err := validation.ValidateProject(root, cfg)
	if err != nil {
		t.Fatalf("ValidateProject() unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("ValidateProject() findings = %v, want empty for a well-formed Constitution", findings)
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
