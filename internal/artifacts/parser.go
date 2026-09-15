package artifacts

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mottamarcio/misterspec/internal/ids"
)

// frontmatterYAML is the loosely typed raw shape of an artifact's
// frontmatter block, decoded before field-level validation.
type frontmatterYAML struct {
	ID         string   `yaml:"id"`
	Type       string   `yaml:"type"`
	Status     string   `yaml:"status"`
	Parent     string   `yaml:"parent"`
	DependsOn  []string `yaml:"depends_on"`
	Supersedes []string `yaml:"supersedes"`
	For        string   `yaml:"for"`
}

// idBearingTypes are the declared frontmatter `type` values that require
// an `id` field (FR-009). Plan/Tasks/Validation/Constitution are
// addressed via the Spec or project they belong to instead of carrying an
// independent ID (docs/architecture-specification.md §22-31).
var idBearingTypes = map[string]bool{
	"program":   true,
	"feature":   true,
	"spec":      true,
	"knowledge": true,
	"learning":  true,
}

// ParseMetadata reads path and parses its frontmatter block into a
// Metadata value. It returns exactly one of a populated Metadata, or one
// of ErrArtifactNotFound, ErrFrontmatterMalformed, or
// ErrRequiredFieldMissing — never a partially populated Metadata alongside
// a non-nil error (FR-009).
func ParseMetadata(path string) (Metadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Metadata{}, fmt.Errorf("%w: %s", ErrArtifactNotFound, path)
		}
		return Metadata{}, fmt.Errorf("artifacts: reading %s: %w", path, err)
	}

	block, _, err := splitFrontmatter(data)
	if err != nil {
		return Metadata{}, fmt.Errorf("%w: %s: %v", ErrFrontmatterMalformed, path, err)
	}

	var raw frontmatterYAML
	dec := yaml.NewDecoder(bytes.NewReader(block))
	if err := dec.Decode(&raw); err != nil {
		return Metadata{}, fmt.Errorf("%w: %s: %v", ErrFrontmatterMalformed, path, err)
	}

	meta := Metadata{
		Type:   raw.Type,
		Status: raw.Status,
	}

	switch {
	case raw.ID != "":
		id, err := parseFieldID(raw.ID)
		if err != nil {
			return Metadata{}, fmt.Errorf("%w: %s: field \"id\": %v", ErrFrontmatterMalformed, path, err)
		}
		meta.ID = id
	case idBearingTypes[raw.Type]:
		return Metadata{}, fmt.Errorf("%w: %s: field \"id\" is required for type %q", ErrRequiredFieldMissing, path, raw.Type)
	}

	if raw.Parent != "" {
		id, err := parseFieldID(raw.Parent)
		if err != nil {
			return Metadata{}, fmt.Errorf("%w: %s: field \"parent\": %v", ErrFrontmatterMalformed, path, err)
		}
		meta.Parent = id
	}

	if raw.For != "" {
		id, err := parseFieldID(raw.For)
		if err != nil {
			return Metadata{}, fmt.Errorf("%w: %s: field \"for\": %v", ErrFrontmatterMalformed, path, err)
		}
		meta.For = id
	}

	if meta.DependsOn, err = parseFieldIDList(raw.DependsOn, "depends_on", path); err != nil {
		return Metadata{}, err
	}
	if meta.Supersedes, err = parseFieldIDList(raw.Supersedes, "supersedes", path); err != nil {
		return Metadata{}, err
	}

	return meta, nil
}

func parseFieldIDList(raw []string, field, path string) ([]ids.EntityID, error) {
	var out []ids.EntityID
	for _, s := range raw {
		id, err := parseFieldID(s)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: field %q: %v", ErrFrontmatterMalformed, path, field, err)
		}
		if id != nil {
			out = append(out, *id)
		}
	}
	return out, nil
}

// splitFrontmatter splits data into its two halves: the YAML content
// between the first two "---" delimiter lines at the very start of data
// (docs/architecture-specification.md §22-31), and the Markdown body —
// everything after the closing delimiter's own line. ParseMetadata uses
// only the frontmatter half, exactly as before this function gained the
// body half; internal/artifacts.ReadBody (011-wikilink-foundation) is
// the body half's own caller. Widened rather than duplicated, since the
// delimiter-scanning logic is identical either way
// (specs/011-wikilink-foundation/research.md) — extractFrontmatter's
// original name and single-purpose behavior predate this widening.
func splitFrontmatter(data []byte) (frontmatter, body []byte, err error) {
	lines := strings.Split(string(data), "\n")

	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != "---" {
		return nil, nil, fmt.Errorf("does not start with a %q frontmatter delimiter", "---")
	}

	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == "---" {
			frontmatter = []byte(strings.Join(lines[1:i], "\n"))
			if i+1 < len(lines) {
				body = []byte(strings.Join(lines[i+1:], "\n"))
			}
			return frontmatter, body, nil
		}
	}

	return nil, nil, fmt.Errorf("no closing %q frontmatter delimiter found", "---")
}

// parseFieldID parses a frontmatter field's raw ID string (e.g.
// "SPEC-014") into an *ids.EntityID via ids.ParseAny, which derives its
// zero-padding width from the string itself — an already-written
// artifact's ID is authoritative as written, independent of the
// project's currently configured width. Delegating to ids.ParseAny
// (rather than duplicating its type/width-inference logic here) is a
// deliberate DRY fix made in 002-read-operations, once that package
// needed the identical parsing for a bare ID string with no known type in
// advance.
func parseFieldID(raw string) (*ids.EntityID, error) {
	id, err := ids.ParseAny(raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
