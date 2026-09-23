package eval

import (
	"fmt"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	contextengine "github.com/mottamarcio/misterspec/internal/context"
	"github.com/mottamarcio/misterspec/internal/context/index"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// RunRetrievalCase runs one EvaluationCase's own Context Pack request
// against contextengine.Collect/Rank/ApplyBudget (in-process — plan.md
// Summary) and reports whether every required identifier is present and
// every forbidden identifier absent (data-model.md "CaseResult").
func RunRetrievalCase(root string, cfg project.Configuration, store index.Store, c EvaluationCase) (CaseResult, error) {
	req := contextengine.Request{Target: c.Target}
	if c.Query != "" {
		req.Query = c.Query
	}
	if c.Task != "" {
		req.Task = c.Task
	}
	if c.Intent != "" {
		req.Intent = contextengine.Intent(c.Intent)
	}
	if c.Budget != nil {
		req.Budget = c.Budget
	}

	candidates, err := contextengine.Collect(root, cfg, store, req)
	if err != nil {
		return CaseResult{}, fmt.Errorf("case %q: %w", c.ID, err)
	}
	ranked := contextengine.Rank(candidates, req)
	estimator := artifacts.DefaultEstimator{}
	result := contextengine.ApplyBudget(ranked, req, estimator)

	// First occurrence (1-based) of each distinct path, in the
	// already-ranked/budgeted item order — the "position" a required
	// identifier's own file was returned at (User Story 1, Acceptance
	// Scenario 3).
	positionByPath := make(map[string]int, len(result.Items))
	for i, item := range result.Items {
		if _, ok := positionByPath[item.Path]; !ok {
			positionByPath[item.Path] = i + 1
		}
	}

	positions := map[string]int{}
	var missing []string
	for _, id := range c.Required {
		loc, err := operations.Inspect(root, cfg, id)
		if err != nil {
			return CaseResult{}, fmt.Errorf("case %q: required identifier %q: %w", c.ID, id, err)
		}
		if pos, ok := positionByPath[loc.Location.Path]; ok {
			positions[id] = pos
		} else {
			missing = append(missing, id)
		}
	}

	var unexpected []string
	for _, id := range c.Forbidden {
		loc, err := operations.Inspect(root, cfg, id)
		if err != nil {
			return CaseResult{}, fmt.Errorf("case %q: forbidden identifier %q: %w", c.ID, id, err)
		}
		if _, ok := positionByPath[loc.Location.Path]; ok {
			unexpected = append(unexpected, id)
		}
	}

	return CaseResult{
		CaseID:     c.ID,
		Passed:     len(missing) == 0 && len(unexpected) == 0,
		Missing:    missing,
		Unexpected: unexpected,
		Positions:  positions,
	}, nil
}

// Summary is the aggregate pass/fail count over one retrieval
// RunRecord's own CaseResults (contract §1's "summary" field).
type Summary struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// SummarizeCaseResults computes Summary from results.
func SummarizeCaseResults(results []CaseResult) Summary {
	s := Summary{Total: len(results)}
	for _, r := range results {
		if r.Passed {
			s.Passed++
		} else {
			s.Failed++
		}
	}
	return s
}
