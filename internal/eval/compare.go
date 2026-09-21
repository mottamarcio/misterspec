package eval

import (
	"errors"
	"fmt"
	"sort"
)

// ErrIncompatibleRun is returned by Compare when baseline.Kind and
// candidate.Kind differ (contract §2 error table "incompatible_run") —
// a comparison across kinds is meaningless and MUST be refused
// (data-model.md: "A ComparisonReport MUST be producible only when
// baseline and candidate share the same kind").
var ErrIncompatibleRun = errors.New("eval: incompatible run kind")

// ChangedCase is one case/task whose outcome differs between a
// baseline and a candidate run (data-model.md "ChangedCase").
type ChangedCase struct {
	ID               string `json:"id"`
	BaselineOutcome  string `json:"baseline_outcome"`
	CandidateOutcome string `json:"candidate_outcome"`
	Direction        string `json:"direction"`
}

const (
	DirectionImproved  = "improved"
	DirectionRegressed = "regressed"
	DirectionNeutral   = "neutral"
)

// Aggregate is the candidate run's own success-rate and effort ratio,
// always reported together (spec FR-009: "never the ratio alone").
type Aggregate struct {
	SuccessRate               float64  `json:"success_rate"`
	TotalTokensPerCorrectTask *float64 `json:"total_tokens_per_correct_task,omitempty"`
}

// ComparisonReport is the output of comparing a candidate RunRecord
// against a Baseline (data-model.md "ComparisonReport").
type ComparisonReport struct {
	BaselineName   string        `json:"baseline_name"`
	CandidateRunID string        `json:"candidate_run_id"`
	DimensionDiff  []string      `json:"dimension_diff"`
	ChangedCases   []ChangedCase `json:"changed_cases"`
	Aggregate      Aggregate     `json:"aggregate"`
	VarianceNote   string        `json:"variance_note,omitempty"`
	Stale          bool          `json:"stale,omitempty"`
}

// staleDimensions are Config keys whose presence in DimensionDiff means
// the baseline was recorded against an earlier contract/ranking version
// than the candidate now reports — comparison still runs but the result
// is flagged (contract §2 "stale_baseline"; spec Edge Cases), never
// silently trusted.
var staleDimensions = map[string]bool{
	"ranking_version":        true,
	"context_schema_version": true,
}

// DimensionDiff names every Config field that differs between baseline
// and candidate, excluding "repetition_index" (data-model.md
// "ComparisonReport": "except repetition_index"), sorted for
// deterministic output.
func DimensionDiff(baseline, candidate map[string]string) []string {
	seen := map[string]bool{}
	var diffs []string
	for k, v := range baseline {
		if k == "repetition_index" {
			continue
		}
		seen[k] = true
		if candidate[k] != v {
			diffs = append(diffs, k)
		}
	}
	for k, v := range candidate {
		if k == "repetition_index" || seen[k] {
			continue
		}
		if baseline[k] != v {
			diffs = append(diffs, k)
		}
	}
	sort.Strings(diffs)
	return diffs
}

// Compare diffs candidate against baseline, optionally using
// repetitions (additional RunRecords sharing candidate's own Config
// minus repetition_index) to distinguish a real change from normal
// run-to-run variance (FR-013).
func Compare(baseline, candidate RunRecord, repetitions []RunRecord) (ComparisonReport, error) {
	if baseline.Kind != candidate.Kind {
		return ComparisonReport{}, fmt.Errorf("%w: baseline kind %q, candidate kind %q", ErrIncompatibleRun, baseline.Kind, candidate.Kind)
	}

	diff := DimensionDiff(baseline.Config, candidate.Config)
	report := ComparisonReport{
		CandidateRunID: candidate.RunID,
		DimensionDiff:  diff,
	}
	for _, d := range diff {
		if staleDimensions[d] {
			report.Stale = true
			break
		}
	}

	switch baseline.Kind {
	case KindRetrieval:
		baseResults, err := baseline.CaseResults()
		if err != nil {
			return ComparisonReport{}, err
		}
		candResults, err := candidate.CaseResults()
		if err != nil {
			return ComparisonReport{}, err
		}
		report.ChangedCases = compareCaseResults(baseResults, candResults)
		report.Aggregate = Aggregate{SuccessRate: successRateCases(candResults)}
		if len(repetitions) > 0 {
			applyCaseVariance(&report, repetitions)
		}
	case KindTaskExecution:
		baseResults, err := baseline.TaskResults()
		if err != nil {
			return ComparisonReport{}, err
		}
		candResults, err := candidate.TaskResults()
		if err != nil {
			return ComparisonReport{}, err
		}
		report.ChangedCases = compareTaskResults(baseResults, candResults)
		report.Aggregate = aggregateTaskResults(candResults)
		if len(repetitions) > 0 {
			applyTaskVariance(&report, repetitions)
		}
	}

	return report, nil
}

func compareCaseResults(baseline, candidate []CaseResult) []ChangedCase {
	baseByID := make(map[string]bool, len(baseline))
	for _, r := range baseline {
		baseByID[r.CaseID] = r.Passed
	}
	candByID := make(map[string]bool, len(candidate))
	for _, r := range candidate {
		candByID[r.CaseID] = r.Passed
	}

	var changed []ChangedCase
	for id, basePassed := range baseByID {
		candPassed, ok := candByID[id]
		if !ok || candPassed == basePassed {
			continue
		}
		changed = append(changed, ChangedCase{
			ID:               id,
			BaselineOutcome:  outcomeLabel(basePassed),
			CandidateOutcome: outcomeLabel(candPassed),
			Direction:        directionFor(basePassed, candPassed),
		})
	}
	sort.Slice(changed, func(i, j int) bool { return changed[i].ID < changed[j].ID })
	return changed
}

func outcomeLabel(passed bool) string {
	if passed {
		return "pass"
	}
	return "fail"
}

func directionFor(basePassed, candPassed bool) string {
	if !basePassed && candPassed {
		return DirectionImproved
	}
	if basePassed && !candPassed {
		return DirectionRegressed
	}
	return DirectionNeutral
}

func successRateCases(results []CaseResult) float64 {
	if len(results) == 0 {
		return 0
	}
	passed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		}
	}
	return float64(passed) / float64(len(results))
}

// applyCaseVariance downgrades a changed case's Direction to
// DirectionNeutral (and sets VarianceNote) when the supplied
// repetitions themselves disagree on that case's own Passed outcome —
// the comparison MUST NOT claim a directional change the baseline's
// own repeated runs don't support (FR-013).
func applyCaseVariance(report *ComparisonReport, repetitions []RunRecord) {
	observed := map[string]map[bool]bool{}
	for _, rep := range repetitions {
		results, err := rep.CaseResults()
		if err != nil {
			continue
		}
		for _, r := range results {
			if observed[r.CaseID] == nil {
				observed[r.CaseID] = map[bool]bool{}
			}
			observed[r.CaseID][r.Passed] = true
		}
	}

	var flagged []string
	for i, c := range report.ChangedCases {
		if len(observed[c.ID]) > 1 {
			report.ChangedCases[i].Direction = DirectionNeutral
			flagged = append(flagged, c.ID)
		}
	}
	if len(flagged) > 0 {
		report.VarianceNote = fmt.Sprintf("cases with outcome inconsistent across %d supplied repetitions: %v", len(repetitions), flagged)
	}
}

func compareTaskResults(baseline, candidate []TaskResult) []ChangedCase {
	baseByID := make(map[string]string, len(baseline))
	for _, r := range baseline {
		baseByID[r.TaskID] = r.Outcome
	}
	candByID := make(map[string]string, len(candidate))
	for _, r := range candidate {
		candByID[r.TaskID] = r.Outcome
	}

	var changed []ChangedCase
	for id, baseOutcome := range baseByID {
		candOutcome, ok := candByID[id]
		if !ok || candOutcome == baseOutcome {
			continue
		}
		changed = append(changed, ChangedCase{
			ID:               id,
			BaselineOutcome:  baseOutcome,
			CandidateOutcome: candOutcome,
			Direction:        taskDirectionFor(baseOutcome, candOutcome),
		})
	}
	sort.Slice(changed, func(i, j int) bool { return changed[i].ID < changed[j].ID })
	return changed
}

func taskDirectionFor(baseOutcome, candOutcome string) string {
	if baseOutcome != "pass" && candOutcome == "pass" {
		return DirectionImproved
	}
	if baseOutcome == "pass" && candOutcome != "pass" {
		return DirectionRegressed
	}
	return DirectionNeutral
}

// aggregateTaskResults computes success rate and total-tokens-per-
// correctly-completed-task together, per spec FR-009 ("never the ratio
// alone"). "inconclusive" outcomes are excluded from both the
// denominator and the token sum, per spec Edge Cases (recorded, not
// miscounted as pass or fail).
func aggregateTaskResults(results []TaskResult) Aggregate {
	counted := 0
	passed := 0
	var totalTokens int
	haveTokens := false
	for _, r := range results {
		if r.Outcome == "inconclusive" {
			continue
		}
		counted++
		if r.Outcome == "pass" {
			passed++
		}
		if r.Metrics.InputTokens != nil {
			totalTokens += *r.Metrics.InputTokens
			haveTokens = true
		}
		if r.Metrics.OutputTokens != nil {
			totalTokens += *r.Metrics.OutputTokens
			haveTokens = true
		}
	}

	agg := Aggregate{}
	if counted > 0 {
		agg.SuccessRate = float64(passed) / float64(counted)
	}
	if haveTokens && passed > 0 {
		ratio := float64(totalTokens) / float64(passed)
		agg.TotalTokensPerCorrectTask = &ratio
	}
	return agg
}

func applyTaskVariance(report *ComparisonReport, repetitions []RunRecord) {
	observed := map[string]map[string]bool{}
	for _, rep := range repetitions {
		results, err := rep.TaskResults()
		if err != nil {
			continue
		}
		for _, r := range results {
			if observed[r.TaskID] == nil {
				observed[r.TaskID] = map[string]bool{}
			}
			observed[r.TaskID][r.Outcome] = true
		}
	}

	var flagged []string
	for i, c := range report.ChangedCases {
		if len(observed[c.ID]) > 1 {
			report.ChangedCases[i].Direction = DirectionNeutral
			flagged = append(flagged, c.ID)
		}
	}
	if len(flagged) > 0 {
		report.VarianceNote = fmt.Sprintf("tasks with outcome inconsistent across %d supplied repetitions: %v", len(repetitions), flagged)
	}
}
