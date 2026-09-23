package eval

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Kind values a RunRecord.Kind MUST be one of (data-model.md
// "RunRecord") — no other value is valid.
const (
	KindRetrieval     = "retrieval"
	KindTaskExecution = "task_execution"
)

// ErrInvalidRunRecord is the sentinel wrapped by every RunRecord/
// Metrics/TaskResult/Baseline validation failure.
var ErrInvalidRunRecord = errors.New("eval: invalid run record")

// RunRecord is one execution of either evaluation set under a
// specific, recorded configuration (data-model.md "RunRecord"). Results
// is deferred/raw JSON: RunRecord itself does not know whether it holds
// []CaseResult (Kind == KindRetrieval) or []TaskResult (Kind ==
// KindTaskExecution) — callers use CaseResults()/TaskResults()
// accordingly.
type RunRecord struct {
	RunID     string            `json:"run_id"`
	Kind      string            `json:"kind"`
	CreatedAt time.Time         `json:"created_at"`
	Config    map[string]string `json:"config"`
	Results   json.RawMessage   `json:"results"`
}

// Validate reports whether r.Kind is one of the two recognized values
// (data-model.md: "A RunRecord MUST declare exactly one kind").
func (r RunRecord) Validate() error {
	if r.Kind != KindRetrieval && r.Kind != KindTaskExecution {
		return fmt.Errorf("%w: run %q: kind %q must be %q or %q", ErrInvalidRunRecord, r.RunID, r.Kind, KindRetrieval, KindTaskExecution)
	}
	return nil
}

// CaseResults unmarshals r.Results as []CaseResult. r.Kind MUST be
// KindRetrieval.
func (r RunRecord) CaseResults() ([]CaseResult, error) {
	if r.Kind != KindRetrieval {
		return nil, fmt.Errorf("%w: run %q: kind %q is not %q", ErrInvalidRunRecord, r.RunID, r.Kind, KindRetrieval)
	}
	var out []CaseResult
	if err := json.Unmarshal(r.Results, &out); err != nil {
		return nil, fmt.Errorf("%w: run %q: decoding case results: %v", ErrInvalidRunRecord, r.RunID, err)
	}
	return out, nil
}

// TaskResults unmarshals r.Results as []TaskResult. r.Kind MUST be
// KindTaskExecution.
func (r RunRecord) TaskResults() ([]TaskResult, error) {
	if r.Kind != KindTaskExecution {
		return nil, fmt.Errorf("%w: run %q: kind %q is not %q", ErrInvalidRunRecord, r.RunID, r.Kind, KindTaskExecution)
	}
	var out []TaskResult
	if err := json.Unmarshal(r.Results, &out); err != nil {
		return nil, fmt.Errorf("%w: run %q: decoding task results: %v", ErrInvalidRunRecord, r.RunID, err)
	}
	for _, tr := range out {
		if err := tr.Validate(); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// LoadRunRecord reads and validates a RunRecord from path.
func LoadRunRecord(path string) (RunRecord, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return RunRecord{}, fmt.Errorf("%w: reading %s: %v", ErrInvalidRunRecord, path, err)
	}
	var r RunRecord
	if err := json.Unmarshal(raw, &r); err != nil {
		return RunRecord{}, fmt.Errorf("%w: parsing %s: %v", ErrInvalidRunRecord, path, err)
	}
	if err := r.Validate(); err != nil {
		return RunRecord{}, err
	}
	return r, nil
}

// SaveRunRecord writes r as indented JSON to path, creating parent
// directories as needed.
func SaveRunRecord(path string, r RunRecord) error {
	raw, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

// CaseResult is one retrieval case's outcome, nested in a
// KindRetrieval RunRecord's Results (data-model.md "CaseResult").
type CaseResult struct {
	CaseID     string         `json:"case_id"`
	Passed     bool           `json:"passed"`
	Missing    []string       `json:"missing"`
	Unexpected []string       `json:"unexpected"`
	Positions  map[string]int `json:"positions"`
	Score      *float64       `json:"score,omitempty"`
}

// validOutcomes is TaskResult.Outcome's fixed vocabulary (data-model.md
// "TaskResult").
var validOutcomes = map[string]bool{
	"pass":         true,
	"fail":         true,
	"inconclusive": true,
}

// TaskResult is one agent task-execution outcome, nested in a
// KindTaskExecution RunRecord's Results (data-model.md "TaskResult").
type TaskResult struct {
	TaskID  string  `json:"task_id"`
	Outcome string  `json:"outcome"`
	Metrics Metrics `json:"metrics"`
}

// Validate checks TaskResult.Outcome's vocabulary and delegates to
// Metrics.Validate.
func (t TaskResult) Validate() error {
	if !validOutcomes[t.Outcome] {
		return fmt.Errorf("%w: task %q: outcome %q must be \"pass\", \"fail\", or \"inconclusive\"", ErrInvalidRunRecord, t.TaskID, t.Outcome)
	}
	return t.Metrics.Validate()
}

// Metrics is the set of measured/estimated effort figures for one
// TaskResult (data-model.md "Metrics", spec FR-007/FR-008). Every
// numeric field that is present MUST have its sibling "*_estimated"
// boolean present and explicit (research.md #6) — Validate enforces
// this structurally.
type Metrics struct {
	InputTokens           *int  `json:"input_tokens,omitempty"`
	InputTokensEstimated  *bool `json:"input_tokens_estimated,omitempty"`
	OutputTokens          *int  `json:"output_tokens,omitempty"`
	OutputTokensEstimated *bool `json:"output_tokens_estimated,omitempty"`
	CachedTokens          *int  `json:"cached_tokens,omitempty"`
	CachedTokensEstimated *bool `json:"cached_tokens_estimated,omitempty"`
	Calls                 int   `json:"calls"`
	ExtraReads            int   `json:"extra_reads"`
	// ContextFallbacks counts times, during this run, a Context Pack or
	// prepared context was insufficient and the agent fell back to
	// further reads (039-lean-skills-integration-contracts FR-008,
	// data-model.md "Metrics (extended)") — a plain count, like
	// ExtraReads/Rework, with no *_estimated sibling.
	ContextFallbacks int      `json:"context_fallbacks"`
	Rework           int      `json:"rework"`
	LatencySeconds   float64  `json:"latency_seconds"`
	Cost             *float64 `json:"cost,omitempty"`
	CostEstimated    *bool    `json:"cost_estimated,omitempty"`
}

// Validate enforces the "every present numeric figure has an explicit
// sibling *_estimated flag" rule (data-model.md Validation Rules
// Summary; spec FR-008, SC-004).
func (m Metrics) Validate() error {
	type pair struct {
		name string
		set  bool
		est  *bool
	}
	pairs := []pair{
		{"input_tokens", m.InputTokens != nil, m.InputTokensEstimated},
		{"output_tokens", m.OutputTokens != nil, m.OutputTokensEstimated},
		{"cached_tokens", m.CachedTokens != nil, m.CachedTokensEstimated},
		{"cost", m.Cost != nil, m.CostEstimated},
	}
	for _, p := range pairs {
		if p.set && p.est == nil {
			return fmt.Errorf("%w: %q is set but %q is not — every present effort figure must explicitly label whether it is estimated", ErrInvalidRunRecord, p.name, p.name+"_estimated")
		}
	}
	if m.ContextFallbacks < 0 {
		return fmt.Errorf("%w: \"context_fallbacks\" must be >= 0, got %d", ErrInvalidRunRecord, m.ContextFallbacks)
	}
	return nil
}

// Baseline is a named, recorded RunRecord (or aggregate of repetition
// RunRecords sharing the same Config except repetition_index)
// designated as the comparison reference (data-model.md "Baseline").
type Baseline struct {
	Name       string    `json:"name"`
	RunIDs     []string  `json:"run_ids"`
	RecordedAt time.Time `json:"recorded_at"`
}

// LoadBaseline reads and lightly validates a Baseline from path.
func LoadBaseline(path string) (Baseline, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Baseline{}, fmt.Errorf("%w: reading baseline %s: %v", ErrInvalidRunRecord, path, err)
	}
	var b Baseline
	if err := json.Unmarshal(raw, &b); err != nil {
		return Baseline{}, fmt.Errorf("%w: parsing baseline %s: %v", ErrInvalidRunRecord, path, err)
	}
	if b.Name == "" || len(b.RunIDs) == 0 {
		return Baseline{}, fmt.Errorf("%w: baseline %s: \"name\" and at least one \"run_ids\" entry are required", ErrInvalidRunRecord, path)
	}
	return b, nil
}

// SaveBaseline writes b as indented JSON to path, creating parent
// directories as needed.
func SaveBaseline(path string, b Baseline) error {
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}
