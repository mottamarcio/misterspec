package architecture

import (
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/project"
)

// TestCheckArchitecture_DogfoodsMisterSpecOwnRepository is
// 044-architecture-code-context-rules T026 (Polish, spec FR-006): the
// Go adapter runs cleanly end-to-end against MisterSpec's own
// repository, using a rule reflecting a real, already-true package
// boundary this codebase's own Constitution documents (internal/
// validation MUST NOT import internal/operations, since operations
// already imports validation — a cycle otherwise).
func TestCheckArchitecture_DogfoodsMisterSpecOwnRepository(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}

	rules := []project.ArchitectureRule{
		{Kind: "forbidden_dependency", From: "internal/validation/**", To: "internal/operations"},
	}

	report, err := CheckArchitecture(repoRoot, rules, nil)
	if err != nil {
		t.Fatalf("CheckArchitecture() unexpected error against the real repository: %v", err)
	}
	if report.Adapter != "go" {
		t.Fatalf("Adapter = %q, want \"go\" (this repository has its own go.mod)", report.Adapter)
	}
	if len(report.Results) != 1 || report.Results[0].Status != StatusPass {
		t.Errorf("Results = %+v, want exactly one pass (internal/validation does not import internal/operations today)", report.Results)
	}
}
