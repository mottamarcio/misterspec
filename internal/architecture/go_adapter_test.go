package architecture

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/project"
)

func writeFixtureFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", relPath, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", relPath, err)
	}
}

// setupForbiddenDependencyFixture builds a fixture Go module: module
// "fixture.example", with a/a.go importing fixture.example/b.
func setupForbiddenDependencyFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFixtureFile(t, dir, "go.mod", "module fixture.example\n\ngo 1.23\n")
	writeFixtureFile(t, dir, "a/a.go", "package a\n\nimport \"fixture.example/b\"\n\nfunc UseB() { b.Thing() }\n")
	writeFixtureFile(t, dir, "b/b.go", "package b\n\nfunc Thing() {}\n")
	return dir
}

// TestCheckArchitecture_ForbiddenDependencyViolation is
// 044-architecture-code-context-rules T014 (US1): a violated
// forbidden_dependency rule returns a fail Result with the exact file
// and line of the forbidden import, reproducibly.
func TestCheckArchitecture_ForbiddenDependencyViolation(t *testing.T) {
	dir := setupForbiddenDependencyFixture(t)
	rules := []project.ArchitectureRule{
		{Kind: "forbidden_dependency", From: "a/**", To: "b"},
	}

	report, err := CheckArchitecture(dir, rules, nil)
	if err != nil {
		t.Fatalf("CheckArchitecture() unexpected error: %v", err)
	}
	if report.Adapter != "go" {
		t.Errorf("Adapter = %q, want \"go\"", report.Adapter)
	}

	var fails []Result
	for _, r := range report.Results {
		if r.Status == StatusFail {
			fails = append(fails, r)
		}
	}
	if len(fails) != 1 {
		t.Fatalf("fail results = %+v, want exactly 1", fails)
	}
	if fails[0].Path != "a/a.go" || fails[0].Line != 3 {
		t.Errorf("fail = %+v, want Path \"a/a.go\", Line 3", fails[0])
	}

	// Re-running against unchanged fixture code produces byte-identical
	// Path/Line/Message (spec Acceptance Scenario 3).
	report2, err := CheckArchitecture(dir, rules, nil)
	if err != nil {
		t.Fatalf("CheckArchitecture() (second run) unexpected error: %v", err)
	}
	var fails2 []Result
	for _, r := range report2.Results {
		if r.Status == StatusFail {
			fails2 = append(fails2, r)
		}
	}
	if len(fails2) != 1 || fails2[0] != fails[0] {
		t.Errorf("second run fails = %+v, want identical to first run %+v", fails2, fails)
	}
}

// TestCheckArchitecture_ForbiddenDependencyPass is
// 044-architecture-code-context-rules T014 (US1): a rule with no
// violation returns pass.
func TestCheckArchitecture_ForbiddenDependencyPass(t *testing.T) {
	dir := setupForbiddenDependencyFixture(t)
	rules := []project.ArchitectureRule{
		{Kind: "forbidden_dependency", From: "b/**", To: "a"},
	}

	report, err := CheckArchitecture(dir, rules, nil)
	if err != nil {
		t.Fatalf("CheckArchitecture() unexpected error: %v", err)
	}
	if len(report.Results) != 1 || report.Results[0].Status != StatusPass {
		t.Errorf("Results = %+v, want exactly one pass", report.Results)
	}
}

// TestCheckArchitecture_NoGoModIsNotEvaluated is
// 044-architecture-code-context-rules T018 (US2, spec FR-004): a
// project directory with no go.mod returns Report{Adapter: ""} with
// every rule as not_evaluated/"no_adapter_for_project", never "pass".
func TestCheckArchitecture_NoGoModIsNotEvaluated(t *testing.T) {
	dir := t.TempDir() // no go.mod written
	rules := []project.ArchitectureRule{
		{Kind: "forbidden_dependency", From: "a/**", To: "b"},
	}

	report, err := CheckArchitecture(dir, rules, nil)
	if err != nil {
		t.Fatalf("CheckArchitecture() unexpected error: %v", err)
	}
	if report.Adapter != "" {
		t.Errorf("Adapter = %q, want \"\"", report.Adapter)
	}
	if len(report.Results) != 1 {
		t.Fatalf("Results = %+v, want exactly 1", report.Results)
	}
	r := report.Results[0]
	if r.Status != StatusNotEvaluated {
		t.Errorf("Status = %q, want %q", r.Status, StatusNotEvaluated)
	}
	if r.Reason != ReasonNoAdapterForProject {
		t.Errorf("Reason = %q, want %q", r.Reason, ReasonNoAdapterForProject)
	}
	if r.Status == StatusPass {
		t.Error("Status = pass, want never pass when no adapter applies")
	}
}

// TestCheckArchitecture_UnsupportedRuleKindIsNotEvaluated is
// 044-architecture-code-context-rules T020 (US2, spec Acceptance
// Scenario 2): a required_contract rule is not_evaluated, while a
// forbidden_dependency rule in the same run still reports normally.
func TestCheckArchitecture_UnsupportedRuleKindIsNotEvaluated(t *testing.T) {
	dir := setupForbiddenDependencyFixture(t)
	rules := []project.ArchitectureRule{
		{Kind: "required_contract", From: "a/**", Contract: "b"},
		{Kind: "forbidden_dependency", From: "b/**", To: "a"}, // no violation, pass
	}

	report, err := CheckArchitecture(dir, rules, nil)
	if err != nil {
		t.Fatalf("CheckArchitecture() unexpected error: %v", err)
	}
	if len(report.Results) != 2 {
		t.Fatalf("Results = %+v, want 2", report.Results)
	}

	contractResult := report.Results[0]
	if contractResult.Status != StatusNotEvaluated || contractResult.Reason != ReasonRuleKindUnsupported {
		t.Errorf("required_contract result = %+v, want not_evaluated/rule_kind_unsupported", contractResult)
	}

	depResult := report.Results[1]
	if depResult.Status != StatusPass {
		t.Errorf("forbidden_dependency result = %+v, want pass (unaffected by the unsupported rule)", depResult)
	}
}

// TestCheckArchitecture_LayerBoundaryViolation is
// 044-architecture-code-context-rules T015 (US1): a layer_boundary
// rule is evaluated the same way forbidden_dependency is.
func TestCheckArchitecture_LayerBoundaryViolation(t *testing.T) {
	dir := setupForbiddenDependencyFixture(t)
	rules := []project.ArchitectureRule{
		{Kind: "layer_boundary", From: "a/**", To: "b"},
	}

	report, err := CheckArchitecture(dir, rules, nil)
	if err != nil {
		t.Fatalf("CheckArchitecture() unexpected error: %v", err)
	}
	var fails int
	for _, r := range report.Results {
		if r.Status == StatusFail {
			fails++
		}
	}
	if fails != 1 {
		t.Errorf("fail count = %d, want 1", fails)
	}
}
