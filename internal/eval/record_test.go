package eval

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestRunRecord_RoundTripsThroughJSON(t *testing.T) {
	results := []CaseResult{{
		CaseID:    "c1",
		Passed:    true,
		Positions: map[string]int{"SPEC-001": 1},
	}}
	raw, err := json.Marshal(results)
	if err != nil {
		t.Fatal(err)
	}
	r := RunRecord{
		RunID:     "retrieval-1",
		Kind:      KindRetrieval,
		CreatedAt: time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC),
		Config:    map[string]string{"variant": "default"},
		Results:   raw,
	}

	path := filepath.Join(t.TempDir(), "run.json")
	if err := SaveRunRecord(path, r); err != nil {
		t.Fatalf("SaveRunRecord() error = %v", err)
	}
	got, err := LoadRunRecord(path)
	if err != nil {
		t.Fatalf("LoadRunRecord() error = %v", err)
	}
	if got.RunID != r.RunID || got.Kind != r.Kind || got.Config["variant"] != "default" {
		t.Fatalf("LoadRunRecord() = %+v, want round-trip of %+v", got, r)
	}
	gotResults, err := got.CaseResults()
	if err != nil {
		t.Fatalf("CaseResults() error = %v", err)
	}
	if len(gotResults) != 1 || gotResults[0].CaseID != "c1" || !gotResults[0].Passed {
		t.Fatalf("CaseResults() = %+v, want round-trip of %+v", gotResults, results)
	}
}

func TestRunRecord_Validate_RejectsUnknownKind(t *testing.T) {
	r := RunRecord{RunID: "x", Kind: "not-a-real-kind"}
	if err := r.Validate(); !errors.Is(err, ErrInvalidRunRecord) {
		t.Fatalf("Validate() error = %v, want ErrInvalidRunRecord", err)
	}
}

func TestMetrics_Validate_RejectsTokenWithoutEstimatedFlag(t *testing.T) {
	n := 100
	m := Metrics{InputTokens: &n}
	if err := m.Validate(); !errors.Is(err, ErrInvalidRunRecord) {
		t.Fatalf("Validate() error = %v, want ErrInvalidRunRecord (missing input_tokens_estimated)", err)
	}
}

func TestMetrics_Validate_AcceptsTokenWithEstimatedFlag(t *testing.T) {
	n := 100
	estimated := true
	m := Metrics{InputTokens: &n, InputTokensEstimated: &estimated}
	if err := m.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestTaskResult_Validate_RejectsUnknownOutcome(t *testing.T) {
	tr := TaskResult{TaskID: "t1", Outcome: "maybe"}
	if err := tr.Validate(); !errors.Is(err, ErrInvalidRunRecord) {
		t.Fatalf("Validate() error = %v, want ErrInvalidRunRecord", err)
	}
}

func TestRunRecord_TaskResults_ValidatesEachResult(t *testing.T) {
	raw, err := json.Marshal([]TaskResult{{TaskID: "t1", Outcome: "pass"}})
	if err != nil {
		t.Fatal(err)
	}
	r := RunRecord{RunID: "r1", Kind: KindTaskExecution, Results: raw}
	out, err := r.TaskResults()
	if err != nil {
		t.Fatalf("TaskResults() error = %v, want nil", err)
	}
	if len(out) != 1 || out[0].TaskID != "t1" {
		t.Fatalf("TaskResults() = %+v", out)
	}
}

func TestBaseline_RoundTripsThroughJSON(t *testing.T) {
	b := Baseline{
		Name:       "retrieval-2026-09",
		RunIDs:     []string{"retrieval-1"},
		RecordedAt: time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC),
	}
	path := filepath.Join(t.TempDir(), "baseline.json")
	if err := SaveBaseline(path, b); err != nil {
		t.Fatalf("SaveBaseline() error = %v", err)
	}
	got, err := LoadBaseline(path)
	if err != nil {
		t.Fatalf("LoadBaseline() error = %v", err)
	}
	if got.Name != b.Name || len(got.RunIDs) != 1 || got.RunIDs[0] != "retrieval-1" {
		t.Fatalf("LoadBaseline() = %+v, want round-trip of %+v", got, b)
	}
}

func TestLoadBaseline_RejectsMissingNameOrRunIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "baseline.json")
	if err := SaveBaseline(path, Baseline{Name: "", RunIDs: nil}); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadBaseline(path); !errors.Is(err, ErrInvalidRunRecord) {
		t.Fatalf("LoadBaseline() error = %v, want ErrInvalidRunRecord", err)
	}
}
