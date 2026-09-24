package project_test

import (
	"errors"
	"reflect"
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
		SchemaVersion:       1,
		AgentID:             "claude-code",
		ArtifactsDir:        "ai",
		RawDir:              "ai/raw",
		KnowledgeDir:        "ai/knowledge",
		ConstitutionPath:    "ai/memory/constitution.md",
		LearningsDir:        "ai/memory/learnings",
		ProgramsRoot:        "ai/programs",
		IDWidth:             3,
		GitBranchAutomation: false,
	}
	// reflect.DeepEqual, not !=: Configuration is no longer comparable
	// once ArchitectureRules/CodeExclusions (both slices) were added
	// (044-architecture-code-context-rules) — correction found during
	// implementation.
	if !reflect.DeepEqual(cfg, want) {
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
	if cfg.GitBranchAutomation != project.DefaultGitBranchAutomation {
		t.Errorf("GitBranchAutomation = %v, want default %v", cfg.GitBranchAutomation, project.DefaultGitBranchAutomation)
	}
}

// TestLoad_GitBranchAutomationExplicitFalse (022-feature-branch-
// automation) — an explicit "false" in the YAML must be honored, not
// silently replaced by the default (same "absent means default" rule
// every other optional field already follows).
func TestLoad_GitBranchAutomationExplicitFalse(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteConfig(t, root, testutil.DefaultConfigYAML+"git_branch_automation: false\n")

	cfg, err := project.Load(root)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.GitBranchAutomation {
		t.Error("GitBranchAutomation = true, want false (explicit override in config.yaml)")
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

// TestLoad_ArchitectureRulesAndCodeExclusions is
// 044-architecture-code-context-rules T005 (Foundational): a
// well-formed architecture_rules list and code_exclusions list are
// parsed into Configuration.ArchitectureRules/CodeExclusions
// (data-model.md "ArchitectureRule (config)"/"CodeExclusions
// (config)").
func TestLoad_ArchitectureRulesAndCodeExclusions(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteConfig(t, root, "schema_version: 1\nagent_id: claude-code\n"+
		"architecture_rules:\n  - kind: forbidden_dependency\n    from: internal/artifacts/**\n    to: internal/cli/internalcmd\n"+
		"code_exclusions:\n  - vendor/**\n")

	cfg, err := project.Load(root)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if len(cfg.ArchitectureRules) != 1 {
		t.Fatalf("ArchitectureRules = %+v, want 1 entry", cfg.ArchitectureRules)
	}
	rule := cfg.ArchitectureRules[0]
	if rule.Kind != "forbidden_dependency" || rule.From != "internal/artifacts/**" || rule.To != "internal/cli/internalcmd" {
		t.Errorf("ArchitectureRules[0] = %+v, want {forbidden_dependency, internal/artifacts/**, internal/cli/internalcmd}", rule)
	}
	if len(cfg.CodeExclusions) != 1 || cfg.CodeExclusions[0] != "vendor/**" {
		t.Errorf("CodeExclusions = %+v, want [\"vendor/**\"]", cfg.CodeExclusions)
	}
}

// TestLoad_ArchitectureRulesAndCodeExclusionsAbsentAreEmpty is
// 044-architecture-code-context-rules T005: both keys absent from the
// YAML produce empty/nil slices, never an error.
func TestLoad_ArchitectureRulesAndCodeExclusionsAbsentAreEmpty(t *testing.T) {
	root := testutil.Project(t)

	cfg, err := project.Load(root)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if len(cfg.ArchitectureRules) != 0 {
		t.Errorf("ArchitectureRules = %+v, want none", cfg.ArchitectureRules)
	}
	if len(cfg.CodeExclusions) != 0 {
		t.Errorf("CodeExclusions = %+v, want none", cfg.CodeExclusions)
	}
}

// TestLoad_ArchitectureRuleUnknownKind is
// 044-architecture-code-context-rules T005 (contracts §7): a rule
// with an unrecognized kind is rejected via the same
// ErrInvalidConfiguration convention every other malformed field
// already uses.
func TestLoad_ArchitectureRuleUnknownKind(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteConfig(t, root, "schema_version: 1\nagent_id: claude-code\n"+
		"architecture_rules:\n  - kind: not_a_real_kind\n    from: internal/foo/**\n")

	_, err := project.Load(root)
	assertInvalidField(t, err, "architecture_rules")
}

// TestLoad_ArchitectureRuleEmptyFrom is
// 044-architecture-code-context-rules T005.
func TestLoad_ArchitectureRuleEmptyFrom(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteConfig(t, root, "schema_version: 1\nagent_id: claude-code\n"+
		"architecture_rules:\n  - kind: forbidden_dependency\n    from: \"\"\n    to: internal/foo\n")

	_, err := project.Load(root)
	assertInvalidField(t, err, "architecture_rules")
}

// TestLoad_ArchitectureRuleEmptyTo is a code-review regression: a
// forbidden_dependency/layer_boundary rule with an empty "to" must be
// rejected, not silently accepted and later always report a false
// StatusPass (an empty "to" glob only matches an empty normalized
// import path, which architecture.evaluateDependencyRule essentially
// never produces).
func TestLoad_ArchitectureRuleEmptyTo(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteConfig(t, root, "schema_version: 1\nagent_id: claude-code\n"+
		"architecture_rules:\n  - kind: forbidden_dependency\n    from: internal/foo/**\n")

	_, err := project.Load(root)
	assertInvalidField(t, err, "architecture_rules")
}

// TestLoad_ArchitectureRuleEmptyContract is the required_contract
// counterpart of TestLoad_ArchitectureRuleEmptyTo.
func TestLoad_ArchitectureRuleEmptyContract(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteConfig(t, root, "schema_version: 1\nagent_id: claude-code\n"+
		"architecture_rules:\n  - kind: required_contract\n    from: internal/foo/**\n")

	_, err := project.Load(root)
	assertInvalidField(t, err, "architecture_rules")
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
