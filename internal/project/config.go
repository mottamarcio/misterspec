package project

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Configuration is the resolved set of settings governing where artifacts,
// knowledge, memory, and programs live for one project, plus ID formatting
// rules. It mirrors the schema at .misterspec/config.yaml (see
// docs/architecture-specification.md §21).
//
// This type is data-only here. Loading a Configuration from disk and
// validating it (returning ErrInvalidConfiguration when required fields are
// missing or malformed) is implemented in root.go's sibling loader,
// Load — see that function's documentation for the field-by-field
// validation rules.
type Configuration struct {
	// SchemaVersion is the config schema version. Required; an
	// unrecognized value is an invalid-configuration condition.
	SchemaVersion int `yaml:"schema_version"`

	// AgentID identifies the coding agent this project was initialized
	// for (e.g. "claude-code").
	AgentID string `yaml:"agent_id"`

	// ArtifactsDir is the root of all semantic project state, relative
	// to the project root. Required.
	ArtifactsDir string `yaml:"artifacts_dir"`

	// RawDir is where unprocessed source documents live, relative to
	// the project root. Required.
	RawDir string `yaml:"raw_dir"`

	// KnowledgeDir is where Knowledge artifacts live, relative to the
	// project root. Required.
	KnowledgeDir string `yaml:"knowledge_dir"`

	// ConstitutionPath is the path to the project's constitution,
	// relative to the project root. Required.
	ConstitutionPath string `yaml:"constitution_path"`

	// LearningsDir is where Learning artifacts live, relative to the
	// project root. Required.
	LearningsDir string `yaml:"learnings_dir"`

	// ProgramsRoot is the root directory under which Program/Feature/
	// Spec artifacts live, relative to the project root. Required.
	ProgramsRoot string `yaml:"programs_root"`

	// IDWidth is the zero-padding width for numeric ID suffixes (e.g.
	// 3 for SPEC-014). Must be a positive integer.
	IDWidth int `yaml:"id_width"`

	// GitBranchAutomation controls whether creating a Feature/Spec
	// automatically manages a dedicated Git branch for that Feature
	// (022-feature-branch-automation). Defaults to true.
	GitBranchAutomation bool `yaml:"git_branch_automation"`
}

// Default configuration values, used by Load to fill in fields the project
// configuration file legitimately leaves unset. Per the project
// constitution (Principle IV) and FR-005, these defaults apply only to
// genuinely optional fields — never to a field Load treats as required.
const (
	DefaultArtifactsDir        = "ai"
	DefaultRawDir              = "ai/raw"
	DefaultKnowledgeDir        = "ai/knowledge"
	DefaultConstitutionPath    = "ai/memory/constitution.md"
	DefaultLearningsDir        = "ai/memory/learnings"
	DefaultProgramsRoot        = "ai/programs"
	DefaultIDWidth             = 3
	DefaultGitBranchAutomation = true
)

// configFilePath is the fixed location of a project's configuration file,
// relative to its root, per docs/architecture-specification.md §21.
const configFilePath = ".misterspec/config.yaml"

// ErrInvalidConfiguration is returned (wrapped inside a *ConfigError) when
// a project's configuration file exists but is malformed, not valid YAML,
// or missing/empty for a field that does not fall back to a default
// (FR-005). Use errors.Is(err, ErrInvalidConfiguration) to detect the
// condition, and errors.As(err, &configErr) to inspect which field failed
// and why.
var ErrInvalidConfiguration = errors.New("project: invalid configuration")

// ConfigError names the specific configuration field that failed
// validation, and why. It always wraps ErrInvalidConfiguration.
type ConfigError struct {
	Field  string
	Reason string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("project: invalid configuration: field %q: %s", e.Field, e.Reason)
}

func (e *ConfigError) Unwrap() error {
	return ErrInvalidConfiguration
}

// configYAML mirrors Configuration but with pointer fields, so Load can
// distinguish "the key was absent from the YAML" (nil — a default may
// apply) from "the key was present with an empty/zero value" (non-nil —
// never silently defaulted; FR-005).
type configYAML struct {
	SchemaVersion       *int    `yaml:"schema_version"`
	AgentID             *string `yaml:"agent_id"`
	ArtifactsDir        *string `yaml:"artifacts_dir"`
	RawDir              *string `yaml:"raw_dir"`
	KnowledgeDir        *string `yaml:"knowledge_dir"`
	ConstitutionPath    *string `yaml:"constitution_path"`
	LearningsDir        *string `yaml:"learnings_dir"`
	ProgramsRoot        *string `yaml:"programs_root"`
	IDWidth             *int    `yaml:"id_width"`
	GitBranchAutomation *bool   `yaml:"git_branch_automation"`
}

// Load reads and validates root's .misterspec/config.yaml, returning a
// fully populated Configuration (every field present — required fields as
// given, optional fields defaulted where the key was absent) or a
// *ConfigError wrapping ErrInvalidConfiguration naming the specific
// problem (FR-004, FR-005).
//
// Load assumes the configuration file exists; Detect is responsible for
// distinguishing "no project here at all" (ErrNotInitialized) from "a
// project is here but its configuration is broken" (ErrInvalidConfiguration)
// — see root.go.
func Load(root string) (Configuration, error) {
	path := filepath.Join(root, configFilePath)

	data, err := os.ReadFile(path)
	if err != nil {
		return Configuration{}, &ConfigError{
			Field:  "(file)",
			Reason: fmt.Sprintf("could not read %s: %v", configFilePath, err),
		}
	}

	var raw configYAML
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&raw); err != nil {
		return Configuration{}, &ConfigError{
			Field:  "(file)",
			Reason: fmt.Sprintf("%s is not well-formed YAML: %v", configFilePath, err),
		}
	}

	return validateConfig(raw)
}

func validateConfig(raw configYAML) (Configuration, error) {
	if raw.SchemaVersion == nil {
		return Configuration{}, &ConfigError{Field: "schema_version", Reason: "required field is missing"}
	}
	if *raw.SchemaVersion != 1 {
		return Configuration{}, &ConfigError{
			Field:  "schema_version",
			Reason: fmt.Sprintf("unsupported schema_version %d", *raw.SchemaVersion),
		}
	}

	if raw.AgentID == nil || *raw.AgentID == "" {
		return Configuration{}, &ConfigError{Field: "agent_id", Reason: "required field is missing or empty"}
	}

	cfg := Configuration{
		SchemaVersion: *raw.SchemaVersion,
		AgentID:       *raw.AgentID,
	}

	var err error
	if cfg.ArtifactsDir, err = stringOrDefault(raw.ArtifactsDir, "artifacts_dir", DefaultArtifactsDir); err != nil {
		return Configuration{}, err
	}
	if cfg.RawDir, err = stringOrDefault(raw.RawDir, "raw_dir", DefaultRawDir); err != nil {
		return Configuration{}, err
	}
	if cfg.KnowledgeDir, err = stringOrDefault(raw.KnowledgeDir, "knowledge_dir", DefaultKnowledgeDir); err != nil {
		return Configuration{}, err
	}
	if cfg.ConstitutionPath, err = stringOrDefault(raw.ConstitutionPath, "constitution_path", DefaultConstitutionPath); err != nil {
		return Configuration{}, err
	}
	if cfg.LearningsDir, err = stringOrDefault(raw.LearningsDir, "learnings_dir", DefaultLearningsDir); err != nil {
		return Configuration{}, err
	}
	if cfg.ProgramsRoot, err = stringOrDefault(raw.ProgramsRoot, "programs_root", DefaultProgramsRoot); err != nil {
		return Configuration{}, err
	}

	switch {
	case raw.IDWidth == nil:
		cfg.IDWidth = DefaultIDWidth
	case *raw.IDWidth <= 0:
		return Configuration{}, &ConfigError{Field: "id_width", Reason: "must be a positive integer"}
	default:
		cfg.IDWidth = *raw.IDWidth
	}

	if raw.GitBranchAutomation == nil {
		cfg.GitBranchAutomation = DefaultGitBranchAutomation
	} else {
		cfg.GitBranchAutomation = *raw.GitBranchAutomation
	}

	return cfg, nil
}

// stringOrDefault returns def when p is nil (the key was absent from the
// YAML). When p is non-nil but points to an empty string, that is a
// validation failure for field — an explicit empty override is never
// silently replaced by the default (FR-005).
func stringOrDefault(p *string, field, def string) (string, error) {
	if p == nil {
		return def, nil
	}
	if *p == "" {
		return "", &ConfigError{Field: field, Reason: "must not be empty when present"}
	}
	return *p, nil
}
