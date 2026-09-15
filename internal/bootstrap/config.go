package bootstrap

import (
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/mottamarcio/misterspec/internal/installer"
	"github.com/mottamarcio/misterspec/internal/project"
)

// configFilePath is the fixed location of a project's configuration
// file, relative to its root — mirrors project's own unexported
// configFilePath constant (docs/architecture-specification.md §21).
const configFilePath = ".misterspec/config.yaml"

// writeDefaultConfig writes a new project's .misterspec/config.yaml at
// targetDir, atomically, with valid defaults for every field
// project.Configuration governs, plus the given agentID (FR-003).
//
// It marshals a real project.Configuration value — built directly from
// internal/project's own exported Default* constants — rather than a
// separately hand-written YAML template, so drift between "what's
// written here" and "what project.Load actually parses" is structurally
// impossible (research.md).
func writeDefaultConfig(targetDir, agentID string) error {
	cfg := project.Configuration{
		SchemaVersion:       1,
		AgentID:             agentID,
		ArtifactsDir:        project.DefaultArtifactsDir,
		RawDir:              project.DefaultRawDir,
		KnowledgeDir:        project.DefaultKnowledgeDir,
		ConstitutionPath:    project.DefaultConstitutionPath,
		LearningsDir:        project.DefaultLearningsDir,
		ProgramsRoot:        project.DefaultProgramsRoot,
		IDWidth:             project.DefaultIDWidth,
		GitBranchAutomation: project.DefaultGitBranchAutomation,
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return installer.WriteAtomicFile(filepath.Join(targetDir, configFilePath), data)
}
