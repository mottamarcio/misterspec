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
	// For is populated from a `for:` frontmatter key — the Plan/Tasks/
	// Validation schemas' way of naming the Spec they belong to
	// (docs/architecture-specification.md §28-30), since those artifact
	// types have no independent ID or `parent:` field of their own.
	For *ids.EntityID
	// SchemaVersion is populated from a `schema_version:` frontmatter
	// key — required for the Constitution (docs/architecture-
	// specification.md §24), which has no `id` field of its own to
	// validate against. Zero means the field was absent.
	SchemaVersion int
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
