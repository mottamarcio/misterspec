package agents

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/mottamarcio/misterspec/internal/installer"
)

// SchemaVersion is install.json's schema version
// (docs/architecture-specification.md §22).
const SchemaVersion = 1

// FrameworkVersion is this build's misterspec version — hardcoded to the
// MVP target release named at the top of
// docs/architecture-specification.md; no version-detection machinery
// exists yet to derive it otherwise (research.md).
const FrameworkVersion = "0.1.0"

// InstallRecord mirrors docs/architecture-specification.md §22's
// install.json schema exactly. It is explicitly non-authoritative
// project state (§22, Constitution Principle III) — a report of what
// happened, never re-derived as "the" source of truth for what's
// actually on disk (FR-008).
type InstallRecord struct {
	SchemaVersion     int    `json:"schema_version"`
	MisterspecVersion string `json:"misterspec_version"`
	Agent             struct {
		ID              string `json:"id"`
		IntegrationPath string `json:"integration_path"`
	} `json:"agent"`
}

// installRecordPath is install.json's fixed location, relative to a
// project root, per docs/architecture-specification.md §22.
const installRecordPath = ".misterspec/install.json"

// RecordInstall writes result as install.json, atomically — reusing
// internal/installer's exported WriteAtomicFile rather than a third
// atomic-write implementation (research.md). Exported so every concrete
// Adapter implementation (in its own package) can call it after
// materializing its Skills, sharing one record-writing implementation
// instead of each reimplementing it.
func RecordInstall(projectRoot string, result InstallResult) error {
	record := InstallRecord{
		SchemaVersion:     SchemaVersion,
		MisterspecVersion: FrameworkVersion,
	}
	record.Agent.ID = result.AdapterID
	record.Agent.IntegrationPath = result.IntegrationPath

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}

	return installer.WriteAtomicFile(filepath.Join(projectRoot, installRecordPath), data)
}

// CurrentInstall reads a project's installation record. It returns
// (record, true, nil) if one exists and parses, (zero, false, nil) if
// none exists yet — a distinct "not installed" result, not an error
// (FR-007) — or (zero, false, err) only for a genuine read/parse
// failure distinct from simple absence. It never re-installs anything,
// and never infers its answer from TargetPath()'s contents.
func CurrentInstall(projectRoot string) (InstallRecord, bool, error) {
	data, err := os.ReadFile(filepath.Join(projectRoot, installRecordPath))
	if err != nil {
		if os.IsNotExist(err) {
			return InstallRecord{}, false, nil
		}
		return InstallRecord{}, false, err
	}

	var record InstallRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return InstallRecord{}, false, err
	}
	return record, true, nil
}
