package eval

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/testutil"
)

func buildRetrievalFixture(t *testing.T) string {
	t.Helper()
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nReferences [[KNOW-002]] directly.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-002-y.md",
		"---\nid: KNOW-002\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSome unrelated facts.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-z.md",
		"---\nid: KNOW-003\ntype: knowledge\nstatus: active\n---\n## Summary\n\nNever referenced by KNOW-001.\n")
	return root
}

func TestRunRetrievalCase_PassesWhenTargetItselfIsRequired(t *testing.T) {
	root := buildRetrievalFixture(t)
	cfg := testConfig()
	store := openSyncedStore(t, root, cfg)

	c := EvaluationCase{ID: "c1", Target: "KNOW-001", Required: []string{"KNOW-001"}}
	result, err := RunRetrievalCase(root, cfg, store, c)
	if err != nil {
		t.Fatalf("RunRetrievalCase() error = %v", err)
	}
	if !result.Passed {
		t.Fatalf("result = %+v, want Passed", result)
	}
	if len(result.Missing) != 0 || len(result.Unexpected) != 0 {
		t.Errorf("result = %+v, want empty Missing/Unexpected", result)
	}
	if pos, ok := result.Positions["KNOW-001"]; !ok || pos < 1 {
		t.Errorf("Positions[KNOW-001] = %v, ok=%v, want a positive rank", pos, ok)
	}
}

func TestRunRetrievalCase_ReportsMissingRequiredIdentifier(t *testing.T) {
	root := buildRetrievalFixture(t)
	cfg := testConfig()
	store := openSyncedStore(t, root, cfg)

	c := EvaluationCase{ID: "c1", Target: "KNOW-001", Required: []string{"KNOW-001", "KNOW-003"}}
	result, err := RunRetrievalCase(root, cfg, store, c)
	if err != nil {
		t.Fatalf("RunRetrievalCase() error = %v", err)
	}
	if result.Passed {
		t.Fatalf("result = %+v, want not Passed (KNOW-003 is unconnected)", result)
	}
	if len(result.Missing) != 1 || result.Missing[0] != "KNOW-003" {
		t.Fatalf("Missing = %v, want [KNOW-003]", result.Missing)
	}
}

func TestRunRetrievalCase_ReportsUnexpectedForbiddenIdentifier(t *testing.T) {
	root := buildRetrievalFixture(t)
	cfg := testConfig()
	store := openSyncedStore(t, root, cfg)

	c := EvaluationCase{ID: "c1", Target: "KNOW-001", Required: []string{"KNOW-001"}, Forbidden: []string{"KNOW-002"}}
	result, err := RunRetrievalCase(root, cfg, store, c)
	if err != nil {
		t.Fatalf("RunRetrievalCase() error = %v", err)
	}
	if result.Passed {
		t.Fatalf("result = %+v, want not Passed (KNOW-002 is wikilinked, so present)", result)
	}
	if len(result.Unexpected) != 1 || result.Unexpected[0] != "KNOW-002" {
		t.Fatalf("Unexpected = %v, want [KNOW-002]", result.Unexpected)
	}
}

func TestSummarizeCaseResults_CountsPassAndFail(t *testing.T) {
	results := []CaseResult{{Passed: true}, {Passed: false}, {Passed: true}}
	s := SummarizeCaseResults(results)
	if s.Total != 3 || s.Passed != 2 || s.Failed != 1 {
		t.Fatalf("SummarizeCaseResults() = %+v, want {3 2 1}", s)
	}
}

func TestRunRetrievalCase_RepeatedRunsAreIdentical(t *testing.T) {
	root := buildRetrievalFixture(t)
	cfg := testConfig()
	store := openSyncedStore(t, root, cfg)

	c := EvaluationCase{ID: "c1", Target: "KNOW-001", Required: []string{"KNOW-001"}, Forbidden: []string{"KNOW-002"}}
	r1, err := RunRetrievalCase(root, cfg, store, c)
	if err != nil {
		t.Fatalf("RunRetrievalCase() error = %v", err)
	}
	r2, err := RunRetrievalCase(root, cfg, store, c)
	if err != nil {
		t.Fatalf("RunRetrievalCase() error = %v", err)
	}
	if r1.Passed != r2.Passed || len(r1.Missing) != len(r2.Missing) || len(r1.Unexpected) != len(r2.Unexpected) {
		t.Fatalf("repeated runs differ: %+v vs %+v", r1, r2)
	}
	for k, v := range r1.Positions {
		if r2.Positions[k] != v {
			t.Fatalf("Positions differ between runs: %v vs %v", r1.Positions, r2.Positions)
		}
	}
}
