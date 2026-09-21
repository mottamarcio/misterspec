package internalcmd_test

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/eval"
)

func writeRunRecordFixture(t *testing.T, path, runID, kind string, config map[string]string, results any) {
	t.Helper()
	raw, err := json.Marshal(results)
	if err != nil {
		t.Fatal(err)
	}
	r := eval.RunRecord{
		RunID:     runID,
		Kind:      kind,
		CreatedAt: time.Now().UTC(),
		Config:    config,
		Results:   raw,
	}
	if err := eval.SaveRunRecord(path, r); err != nil {
		t.Fatal(err)
	}
}

func writeBaselineFixture(t *testing.T, root, name string, runIDs []string) {
	t.Helper()
	b := eval.Baseline{Name: name, RunIDs: runIDs, RecordedAt: time.Now().UTC()}
	if err := eval.SaveBaseline(filepath.Join(root, "eval", "baselines", name+".json"), b); err != nil {
		t.Fatal(err)
	}
}

type evalCompareResponse struct {
	OK            bool `json:"ok"`
	SchemaVersion int  `json:"schema_version"`
	Result        struct {
		Comparison struct {
			BaselineName   string   `json:"baseline_name"`
			CandidateRunID string   `json:"candidate_run_id"`
			DimensionDiff  []string `json:"dimension_diff"`
			ChangedCases   []struct {
				ID        string `json:"id"`
				Direction string `json:"direction"`
			} `json:"changed_cases"`
			Aggregate struct {
				SuccessRate float64 `json:"success_rate"`
			} `json:"aggregate"`
			Stale bool `json:"stale"`
		} `json:"comparison"`
	} `json:"result"`
}

func TestEvalCompareCmd_Success(t *testing.T) {
	root := t.TempDir()
	config := map[string]string{"variant": "default"}
	writeRunRecordFixture(t, filepath.Join(root, "eval", "runs", "base-1.json"), "base-1", eval.KindRetrieval, config,
		[]eval.CaseResult{{CaseID: "c1", Passed: false}, {CaseID: "c2", Passed: true}})
	writeBaselineFixture(t, root, "retrieval-2026-09", []string{"base-1"})

	candidatePath := filepath.Join(t.TempDir(), "candidate.json")
	writeRunRecordFixture(t, candidatePath, "cand-1", eval.KindRetrieval, config,
		[]eval.CaseResult{{CaseID: "c1", Passed: true}, {CaseID: "c2", Passed: true}})

	cmd := internalcmd.NewEvalCompareCmd()
	cmd.SetArgs([]string{"--dir", root, "--baseline", "retrieval-2026-09", "--candidate", candidatePath})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded evalCompareResponse
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output not valid JSON: %v (%s)", err, output)
	}
	if !decoded.OK {
		t.Fatalf("ok = false, want true (output: %s)", output)
	}
	if decoded.Result.Comparison.BaselineName != "retrieval-2026-09" {
		t.Errorf("baseline_name = %q, want retrieval-2026-09", decoded.Result.Comparison.BaselineName)
	}
	if len(decoded.Result.Comparison.ChangedCases) != 1 || decoded.Result.Comparison.ChangedCases[0].ID != "c1" {
		t.Fatalf("changed_cases = %+v, want one entry for c1", decoded.Result.Comparison.ChangedCases)
	}
	if decoded.Result.Comparison.ChangedCases[0].Direction != "improved" {
		t.Errorf("direction = %q, want improved", decoded.Result.Comparison.ChangedCases[0].Direction)
	}
}

func TestEvalCompareCmd_NonexistentBaselineIsInvalidArgument(t *testing.T) {
	root := t.TempDir()
	candidatePath := filepath.Join(t.TempDir(), "candidate.json")
	writeRunRecordFixture(t, candidatePath, "cand-1", eval.KindRetrieval, nil, []eval.CaseResult{})

	cmd := internalcmd.NewEvalCompareCmd()
	cmd.SetArgs([]string{"--dir", root, "--baseline", "does-not-exist", "--candidate", candidatePath})
	output, exitCode := runCmd(cmd)
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_argument")
}

func TestEvalCompareCmd_IncompatibleKindIsRejected(t *testing.T) {
	root := t.TempDir()
	writeRunRecordFixture(t, filepath.Join(root, "eval", "runs", "base-1.json"), "base-1", eval.KindRetrieval, nil, []eval.CaseResult{})
	writeBaselineFixture(t, root, "b1", []string{"base-1"})

	candidatePath := filepath.Join(t.TempDir(), "candidate.json")
	writeRunRecordFixture(t, candidatePath, "cand-1", eval.KindTaskExecution, nil, []eval.TaskResult{})

	cmd := internalcmd.NewEvalCompareCmd()
	cmd.SetArgs([]string{"--dir", root, "--baseline", "b1", "--candidate", candidatePath})
	output, exitCode := runCmd(cmd)
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "incompatible_run")
}

func TestEvalCompareCmd_MultiDimensionDiffIsFlagged(t *testing.T) {
	root := t.TempDir()
	writeRunRecordFixture(t, filepath.Join(root, "eval", "runs", "base-1.json"), "base-1", eval.KindRetrieval,
		map[string]string{"variant": "default", "model": "m1"}, []eval.CaseResult{})
	writeBaselineFixture(t, root, "b1", []string{"base-1"})

	candidatePath := filepath.Join(t.TempDir(), "candidate.json")
	writeRunRecordFixture(t, candidatePath, "cand-1", eval.KindRetrieval,
		map[string]string{"variant": "bm25-v2", "model": "m2"}, []eval.CaseResult{})

	cmd := internalcmd.NewEvalCompareCmd()
	cmd.SetArgs([]string{"--dir", root, "--baseline", "b1", "--candidate", candidatePath})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded evalCompareResponse
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if len(decoded.Result.Comparison.DimensionDiff) != 2 {
		t.Fatalf("dimension_diff = %v, want 2 entries", decoded.Result.Comparison.DimensionDiff)
	}
}

func TestEvalCompareCmd_StaleBaselineIsFlaggedNotHardFailure(t *testing.T) {
	root := t.TempDir()
	writeRunRecordFixture(t, filepath.Join(root, "eval", "runs", "base-1.json"), "base-1", eval.KindRetrieval,
		map[string]string{"ranking_version": "1"}, []eval.CaseResult{})
	writeBaselineFixture(t, root, "b1", []string{"base-1"})

	candidatePath := filepath.Join(t.TempDir(), "candidate.json")
	writeRunRecordFixture(t, candidatePath, "cand-1", eval.KindRetrieval,
		map[string]string{"ranking_version": "2"}, []eval.CaseResult{})

	cmd := internalcmd.NewEvalCompareCmd()
	cmd.SetArgs([]string{"--dir", root, "--baseline", "b1", "--candidate", candidatePath})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (stale baseline is a flag, not a failure) (output: %s)", exitCode, output)
	}

	var decoded evalCompareResponse
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if !decoded.Result.Comparison.Stale {
		t.Errorf("stale = false, want true")
	}
}

func TestEvalCompareCmd_BaselineNativeRepetitionsAreValidatedAgainstBaselineNotCandidate(t *testing.T) {
	root := t.TempDir()
	baselineConfig := map[string]string{"variant": "baseline"}
	writeRunRecordFixture(t, filepath.Join(root, "eval", "runs", "base-1.json"), "base-1", eval.KindRetrieval, baselineConfig,
		[]eval.CaseResult{{CaseID: "c1", Passed: true}})
	writeRunRecordFixture(t, filepath.Join(root, "eval", "runs", "base-2.json"), "base-2", eval.KindRetrieval, baselineConfig,
		[]eval.CaseResult{{CaseID: "c1", Passed: true}})
	writeBaselineFixture(t, root, "b1", []string{"base-1", "base-2"})

	// The candidate's own config legitimately differs from the
	// baseline's — that is exactly the A/B comparison eval-compare
	// exists to support, and must not be rejected just because the
	// baseline's own repetitions (base-2) share the baseline's config
	// rather than the candidate's.
	candidatePath := filepath.Join(t.TempDir(), "candidate.json")
	writeRunRecordFixture(t, candidatePath, "cand-1", eval.KindRetrieval,
		map[string]string{"variant": "new-ranker"}, []eval.CaseResult{{CaseID: "c1", Passed: true}})

	cmd := internalcmd.NewEvalCompareCmd()
	cmd.SetArgs([]string{"--dir", root, "--baseline", "b1", "--candidate", candidatePath})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded evalCompareResponse
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output not valid JSON: %v (%s)", err, output)
	}
	if len(decoded.Result.Comparison.DimensionDiff) != 1 || decoded.Result.Comparison.DimensionDiff[0] != "variant" {
		t.Fatalf("dimension_diff = %v, want [variant]", decoded.Result.Comparison.DimensionDiff)
	}
}

func TestEvalCompareCmd_BaselineNativeRepetitionConfigMismatchIsRejected(t *testing.T) {
	root := t.TempDir()
	writeRunRecordFixture(t, filepath.Join(root, "eval", "runs", "base-1.json"), "base-1", eval.KindRetrieval,
		map[string]string{"variant": "baseline"}, []eval.CaseResult{})
	writeRunRecordFixture(t, filepath.Join(root, "eval", "runs", "base-2.json"), "base-2", eval.KindRetrieval,
		map[string]string{"variant": "different-from-base-1"}, []eval.CaseResult{})
	writeBaselineFixture(t, root, "b1", []string{"base-1", "base-2"})

	candidatePath := filepath.Join(t.TempDir(), "candidate.json")
	writeRunRecordFixture(t, candidatePath, "cand-1", eval.KindRetrieval,
		map[string]string{"variant": "baseline"}, []eval.CaseResult{})

	cmd := internalcmd.NewEvalCompareCmd()
	cmd.SetArgs([]string{"--dir", root, "--baseline", "b1", "--candidate", candidatePath})
	output, exitCode := runCmd(cmd)
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_argument")
}

func TestEvalCompareCmd_MissingRequiredFlagsAreInvalidArgument(t *testing.T) {
	cmd := internalcmd.NewEvalCompareCmd()
	cmd.SetArgs([]string{})
	output, exitCode := runCmd(cmd)
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_argument")
}

func TestEvalCompareCmd_RepetitionConfigMismatchIsRejected(t *testing.T) {
	root := t.TempDir()
	config := map[string]string{"variant": "default"}
	writeRunRecordFixture(t, filepath.Join(root, "eval", "runs", "base-1.json"), "base-1", eval.KindRetrieval, config, []eval.CaseResult{})
	writeBaselineFixture(t, root, "b1", []string{"base-1"})

	candidatePath := filepath.Join(t.TempDir(), "candidate.json")
	writeRunRecordFixture(t, candidatePath, "cand-1", eval.KindRetrieval, config, []eval.CaseResult{})

	repDir := t.TempDir()
	writeRunRecordFixture(t, filepath.Join(repDir, "rep1.json"), "rep-1", eval.KindRetrieval,
		map[string]string{"variant": "different-variant"}, []eval.CaseResult{})

	cmd := internalcmd.NewEvalCompareCmd()
	cmd.SetArgs([]string{"--dir", root, "--baseline", "b1", "--candidate", candidatePath, "--repetitions", repDir})
	output, exitCode := runCmd(cmd)
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2 (output: %s)", exitCode, output)
	}
	assertErrorCode(t, output, "invalid_argument")
}

func TestEvalCompareCmd_VarianceFromRepetitionsIsSurfaced(t *testing.T) {
	root := t.TempDir()
	config := map[string]string{"variant": "default"}
	writeRunRecordFixture(t, filepath.Join(root, "eval", "runs", "base-1.json"), "base-1", eval.KindRetrieval, config,
		[]eval.CaseResult{{CaseID: "flaky", Passed: false}})
	writeBaselineFixture(t, root, "b1", []string{"base-1"})

	candidatePath := filepath.Join(t.TempDir(), "candidate.json")
	writeRunRecordFixture(t, candidatePath, "cand-1", eval.KindRetrieval, config,
		[]eval.CaseResult{{CaseID: "flaky", Passed: true}})

	repDir := t.TempDir()
	writeRunRecordFixture(t, filepath.Join(repDir, "rep1.json"), "rep-1", eval.KindRetrieval, config,
		[]eval.CaseResult{{CaseID: "flaky", Passed: true}})
	writeRunRecordFixture(t, filepath.Join(repDir, "rep2.json"), "rep-2", eval.KindRetrieval, config,
		[]eval.CaseResult{{CaseID: "flaky", Passed: false}})

	cmd := internalcmd.NewEvalCompareCmd()
	cmd.SetArgs([]string{"--dir", root, "--baseline", "b1", "--candidate", candidatePath, "--repetitions", repDir})
	output, exitCode := runCmd(cmd)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded evalCompareResponse
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if len(decoded.Result.Comparison.ChangedCases) != 1 || decoded.Result.Comparison.ChangedCases[0].Direction != "neutral" {
		t.Fatalf("changed_cases = %+v, want one neutral entry", decoded.Result.Comparison.ChangedCases)
	}
}
