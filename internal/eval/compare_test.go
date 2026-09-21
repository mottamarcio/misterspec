package eval

import (
	"encoding/json"
	"errors"
	"testing"
)

func caseRunRecord(t *testing.T, id, kind string, config map[string]string, results []CaseResult) RunRecord {
	t.Helper()
	raw, err := json.Marshal(results)
	if err != nil {
		t.Fatal(err)
	}
	return RunRecord{RunID: id, Kind: kind, Config: config, Results: raw}
}

func taskRunRecord(t *testing.T, id, kind string, config map[string]string, results []TaskResult) RunRecord {
	t.Helper()
	raw, err := json.Marshal(results)
	if err != nil {
		t.Fatal(err)
	}
	return RunRecord{RunID: id, Kind: kind, Config: config, Results: raw}
}

func TestDimensionDiff_EmptyWhenConfigsMatch(t *testing.T) {
	cfg := map[string]string{"variant": "default", "repo_revision": "abc"}
	if diff := DimensionDiff(cfg, cfg); len(diff) != 0 {
		t.Fatalf("DimensionDiff() = %v, want empty", diff)
	}
}

func TestDimensionDiff_NamesDifferingField(t *testing.T) {
	base := map[string]string{"variant": "default"}
	cand := map[string]string{"variant": "bm25-v2"}
	diff := DimensionDiff(base, cand)
	if len(diff) != 1 || diff[0] != "variant" {
		t.Fatalf("DimensionDiff() = %v, want [variant]", diff)
	}
}

func TestDimensionDiff_ExcludesRepetitionIndex(t *testing.T) {
	base := map[string]string{"variant": "default", "repetition_index": "1"}
	cand := map[string]string{"variant": "default", "repetition_index": "2"}
	if diff := DimensionDiff(base, cand); len(diff) != 0 {
		t.Fatalf("DimensionDiff() = %v, want empty (repetition_index excluded)", diff)
	}
}

func TestCompare_RejectsMismatchedKind(t *testing.T) {
	baseline := caseRunRecord(t, "b1", KindRetrieval, nil, nil)
	candidate := taskRunRecord(t, "c1", KindTaskExecution, nil, nil)

	_, err := Compare(baseline, candidate, nil)
	if !errors.Is(err, ErrIncompatibleRun) {
		t.Fatalf("Compare() error = %v, want ErrIncompatibleRun", err)
	}
}

func TestCompare_RetrievalReportsImprovedAndRegressed(t *testing.T) {
	baseline := caseRunRecord(t, "b1", KindRetrieval, map[string]string{"variant": "default"}, []CaseResult{
		{CaseID: "c1", Passed: false},
		{CaseID: "c2", Passed: true},
		{CaseID: "c3", Passed: true},
	})
	candidate := caseRunRecord(t, "c1run", KindRetrieval, map[string]string{"variant": "default"}, []CaseResult{
		{CaseID: "c1", Passed: true},
		{CaseID: "c2", Passed: false},
		{CaseID: "c3", Passed: true},
	})

	report, err := Compare(baseline, candidate, nil)
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if len(report.DimensionDiff) != 0 {
		t.Fatalf("DimensionDiff = %v, want empty", report.DimensionDiff)
	}
	if len(report.ChangedCases) != 2 {
		t.Fatalf("ChangedCases = %+v, want 2 entries", report.ChangedCases)
	}
	byID := map[string]ChangedCase{}
	for _, c := range report.ChangedCases {
		byID[c.ID] = c
	}
	if byID["c1"].Direction != DirectionImproved {
		t.Errorf("c1 direction = %q, want improved", byID["c1"].Direction)
	}
	if _, ok := byID["c3"]; ok {
		t.Errorf("unchanged case c3 must not appear in ChangedCases, got %+v", byID["c3"])
	}
	if byID["c2"].Direction != DirectionRegressed {
		t.Errorf("c2 direction = %q, want regressed", byID["c2"].Direction)
	}
	if report.Aggregate.SuccessRate != 2.0/3.0 {
		t.Errorf("SuccessRate = %v, want 2/3", report.Aggregate.SuccessRate)
	}
}

func TestCompare_FlagsMultiDimensionDiff(t *testing.T) {
	baseline := caseRunRecord(t, "b1", KindRetrieval, map[string]string{"variant": "default", "model": "m1"}, nil)
	candidate := caseRunRecord(t, "c1", KindRetrieval, map[string]string{"variant": "bm25-v2", "model": "m2"}, nil)

	report, err := Compare(baseline, candidate, nil)
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if len(report.DimensionDiff) != 2 {
		t.Fatalf("DimensionDiff = %v, want 2 entries (variant, model)", report.DimensionDiff)
	}
}

func TestCompare_FlagsStaleBaselineOnRankingVersionDiff(t *testing.T) {
	baseline := caseRunRecord(t, "b1", KindRetrieval, map[string]string{"ranking_version": "1"}, nil)
	candidate := caseRunRecord(t, "c1", KindRetrieval, map[string]string{"ranking_version": "2"}, nil)

	report, err := Compare(baseline, candidate, nil)
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if !report.Stale {
		t.Fatalf("report.Stale = false, want true (ranking_version differs)")
	}
}

func TestCompare_VarianceDowngradesInconsistentCaseToNeutral(t *testing.T) {
	baseline := caseRunRecord(t, "b1", KindRetrieval, map[string]string{"variant": "default"}, []CaseResult{
		{CaseID: "flaky", Passed: false},
	})
	candidate := caseRunRecord(t, "c1", KindRetrieval, map[string]string{"variant": "default"}, []CaseResult{
		{CaseID: "flaky", Passed: true},
	})
	repetitions := []RunRecord{
		caseRunRecord(t, "rep1", KindRetrieval, map[string]string{"variant": "default", "repetition_index": "1"}, []CaseResult{{CaseID: "flaky", Passed: true}}),
		caseRunRecord(t, "rep2", KindRetrieval, map[string]string{"variant": "default", "repetition_index": "2"}, []CaseResult{{CaseID: "flaky", Passed: false}}),
	}

	report, err := Compare(baseline, candidate, repetitions)
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if len(report.ChangedCases) != 1 || report.ChangedCases[0].Direction != DirectionNeutral {
		t.Fatalf("ChangedCases = %+v, want one neutral entry", report.ChangedCases)
	}
	if report.VarianceNote == "" {
		t.Errorf("VarianceNote is empty, want it set when repetitions disagree")
	}
}

func TestCompare_TaskExecutionAggregateComputesTokensPerCorrectTask(t *testing.T) {
	in1, out1 := 1000, 500
	estT := true
	baseline := taskRunRecord(t, "b1", KindTaskExecution, nil, []TaskResult{
		{TaskID: "t1", Outcome: "fail"},
	})
	candidate := taskRunRecord(t, "c1", KindTaskExecution, nil, []TaskResult{
		{TaskID: "t1", Outcome: "pass", Metrics: Metrics{InputTokens: &in1, InputTokensEstimated: &estT, OutputTokens: &out1, OutputTokensEstimated: &estT}},
		{TaskID: "t2", Outcome: "inconclusive"},
	})

	report, err := Compare(baseline, candidate, nil)
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if report.Aggregate.TotalTokensPerCorrectTask == nil || *report.Aggregate.TotalTokensPerCorrectTask != 1500 {
		t.Fatalf("TotalTokensPerCorrectTask = %v, want 1500", report.Aggregate.TotalTokensPerCorrectTask)
	}
	if report.Aggregate.SuccessRate != 1.0 {
		t.Errorf("SuccessRate = %v, want 1.0 (inconclusive excluded from denominator)", report.Aggregate.SuccessRate)
	}
}
