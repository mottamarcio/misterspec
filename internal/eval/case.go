package eval

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// ErrInvalidCase is the sentinel wrapped by every EvaluationCase
// validation failure (data-model.md "EvaluationCase" Validation
// rules), letting the CLI layer map it onto the `invalid_case` error
// code (contracts/eval-commands-contract.md §3) without string
// matching.
var ErrInvalidCase = errors.New("eval: invalid evaluation case")

// EvaluationCase is one deterministic retrieval check (data-model.md
// "EvaluationCase").
type EvaluationCase struct {
	ID        string   `yaml:"id"`
	Target    string   `yaml:"target"`
	Query     string   `yaml:"query,omitempty"`
	Task      string   `yaml:"task,omitempty"`
	Intent    string   `yaml:"intent,omitempty"`
	Dir       string   `yaml:"dir"`
	Budget    *int     `yaml:"budget,omitempty"`
	Required  []string `yaml:"required"`
	Forbidden []string `yaml:"forbidden,omitempty"`
	Note      string   `yaml:"note,omitempty"`

	// SourceFile is the case file this value was loaded from, used only
	// for error messages.
	SourceFile string `yaml:"-"`
	// ResolvedDir is Dir resolved to an absolute path against
	// SourceFile's own directory, and confirmed to exist by Validate.
	ResolvedDir string `yaml:"-"`
}

// validate checks c against data-model.md's per-case validation rules,
// resolving Dir to ResolvedDir as a side effect. It does not check
// cross-case rules (ID uniqueness) — see validateCaseSet.
func (c *EvaluationCase) validate() error {
	if c.ID == "" {
		return fmt.Errorf("%w: %s: missing required field \"id\"", ErrInvalidCase, c.SourceFile)
	}
	if c.Target == "" {
		return fmt.Errorf("%w: case %q (%s): missing required field \"target\" — contextengine.Request.Target has no zero-value/\"any\" meaning", ErrInvalidCase, c.ID, c.SourceFile)
	}
	if c.Query != "" && c.Task != "" {
		return fmt.Errorf("%w: case %q (%s): \"query\" and \"task\" are mutually exclusive", ErrInvalidCase, c.ID, c.SourceFile)
	}
	if c.Dir == "" {
		return fmt.Errorf("%w: case %q (%s): missing required field \"dir\"", ErrInvalidCase, c.ID, c.SourceFile)
	}
	resolved := c.Dir
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(filepath.Dir(c.SourceFile), resolved)
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("%w: case %q (%s): \"dir\" %q does not resolve to an existing directory", ErrInvalidCase, c.ID, c.SourceFile, c.Dir)
	}
	c.ResolvedDir = resolved
	if len(c.Required) == 0 && c.Note == "" {
		return fmt.Errorf("%w: case %q (%s): \"required\" is empty and no \"note\" documents this as a deliberate negative/empty-result case", ErrInvalidCase, c.ID, c.SourceFile)
	}
	return nil
}

// validateCaseSet checks rules that span the whole loaded set — ID
// uniqueness (data-model.md: "`id` unique within the loaded set").
func validateCaseSet(cases []EvaluationCase) error {
	seen := make(map[string]string, len(cases))
	for _, c := range cases {
		if prior, ok := seen[c.ID]; ok {
			return fmt.Errorf("%w: duplicate case id %q in %s (first seen in %s)", ErrInvalidCase, c.ID, c.SourceFile, prior)
		}
		seen[c.ID] = c.SourceFile
	}
	return nil
}

// LoadCases reads every *.yaml/*.yml file directly under dir (not
// recursively — one case per file, research.md #2), parses each into
// one EvaluationCase, validates it, and returns the full set in
// filename-sorted (i.e. deterministic) order.
func LoadCases(dir string) ([]EvaluationCase, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("%w: reading case directory %q: %v", ErrInvalidCase, dir, err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext == ".yaml" || ext == ".yml" {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("%w: no case files (*.yaml/*.yml) found under %q", ErrInvalidCase, dir)
	}

	cases := make([]EvaluationCase, 0, len(files))
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("%w: reading %s: %v", ErrInvalidCase, f, err)
		}
		var c EvaluationCase
		if err := yaml.Unmarshal(raw, &c); err != nil {
			return nil, fmt.Errorf("%w: parsing %s: %v", ErrInvalidCase, f, err)
		}
		c.SourceFile = f
		if err := c.validate(); err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}

	if err := validateCaseSet(cases); err != nil {
		return nil, err
	}
	return cases, nil
}
