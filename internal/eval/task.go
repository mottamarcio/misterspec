package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// EvaluationTask is one agent task-execution scenario (data-model.md
// "EvaluationTask"). Unlike EvaluationCase, this is never run by the
// binary — it documents a task a maintainer or live agent session
// follows by hand; the binary only consumes the RunRecord produced
// afterward (research.md #1).
type EvaluationTask struct {
	ID             string `yaml:"id"`
	Description    string `yaml:"description"`
	Dir            string `yaml:"dir"`
	AcceptanceTest string `yaml:"acceptance_test"`

	SourceFile  string `yaml:"-"`
	ResolvedDir string `yaml:"-"`
}

func (t *EvaluationTask) validate() error {
	if t.ID == "" {
		return fmt.Errorf("%w: %s: missing required field \"id\"", ErrInvalidCase, t.SourceFile)
	}
	if t.Description == "" {
		return fmt.Errorf("%w: task %q (%s): missing required field \"description\"", ErrInvalidCase, t.ID, t.SourceFile)
	}
	if t.AcceptanceTest == "" {
		return fmt.Errorf("%w: task %q (%s): missing required field \"acceptance_test\"", ErrInvalidCase, t.ID, t.SourceFile)
	}
	if t.Dir == "" {
		return fmt.Errorf("%w: task %q (%s): missing required field \"dir\"", ErrInvalidCase, t.ID, t.SourceFile)
	}
	resolved := t.Dir
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(filepath.Dir(t.SourceFile), resolved)
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("%w: task %q (%s): \"dir\" %q does not resolve to an existing directory", ErrInvalidCase, t.ID, t.SourceFile, t.Dir)
	}
	t.ResolvedDir = resolved
	return nil
}

func validateTaskSet(tasks []EvaluationTask) error {
	seen := make(map[string]string, len(tasks))
	for _, t := range tasks {
		if prior, ok := seen[t.ID]; ok {
			return fmt.Errorf("%w: duplicate task id %q in %s (first seen in %s)", ErrInvalidCase, t.ID, t.SourceFile, prior)
		}
		seen[t.ID] = t.SourceFile
	}
	return nil
}

// LoadTasks reads every *.yaml/*.yml file directly under dir, parses
// each into one EvaluationTask, validates it, and returns the full set
// in filename-sorted order — mirroring LoadCases (case.go).
func LoadTasks(dir string) ([]EvaluationTask, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("%w: reading task directory %q: %v", ErrInvalidCase, dir, err)
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
		return nil, fmt.Errorf("%w: no task files (*.yaml/*.yml) found under %q", ErrInvalidCase, dir)
	}

	tasks := make([]EvaluationTask, 0, len(files))
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("%w: reading %s: %v", ErrInvalidCase, f, err)
		}
		var t EvaluationTask
		if err := yaml.Unmarshal(raw, &t); err != nil {
			return nil, fmt.Errorf("%w: parsing %s: %v", ErrInvalidCase, f, err)
		}
		t.SourceFile = f
		if err := t.validate(); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	if err := validateTaskSet(tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}
