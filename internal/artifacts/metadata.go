package artifacts

import (
	"errors"

	"github.com/mottamarcio/misterspec/internal/ids"
)

// Metadata is the structured result of parsing one artifact's YAML
// frontmatter block (FR-009). ID is nil for artifact types that have no
// independent ID of their own (Plan, Tasks, Validation, Constitution).
type Metadata struct {
	ID         *ids.EntityID
	Type       string
	Status     string
	Parent     *ids.EntityID
	DependsOn  []ids.EntityID
	Supersedes []ids.EntityID
}

var (
	// ErrArtifactNotFound is returned by ParseMetadata when the file at
	// the given path does not exist.
	ErrArtifactNotFound = errors.New("artifacts: artifact not found")
	// ErrFrontmatterMalformed is returned when the file has no
	// recognizable "---" frontmatter block at its start, or the YAML
	// inside that block does not parse.
	ErrFrontmatterMalformed = errors.New("artifacts: frontmatter not well-formed")
	// ErrRequiredFieldMissing is returned when frontmatter parses
	// successfully but a field required for this artifact's declared
	// type is absent (e.g. "id" on a Spec).
	ErrRequiredFieldMissing = errors.New("artifacts: required frontmatter field missing")
)
