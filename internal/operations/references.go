package operations

import (
	"fmt"
	"path/filepath"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// referenceableTypes are the entity types References/Backlinks operate
// on — exactly the five internal/validation already validates as
// standalone entities (012-references-backlinks/research.md #1). Task,
// Plan, Tasks, Validation, and the Constitution are out of scope.
var referenceableTypes = map[ids.EntityType]bool{
	ids.Program:   true,
	ids.Feature:   true,
	ids.Spec:      true,
	ids.Knowledge: true,
	ids.Learning:  true,
}

// ReferenceEntry is one outgoing relationship from a queried artifact
// (data-model.md).
type ReferenceEntry struct {
	// Relation is "parent", "depends_on", "supersedes" (formal) or
	// "wikilink" (semantic) — never a new fourth kind.
	Relation string
	Target   ids.EntityID
}

// ReferencesResult groups a queried artifact's outgoing relationships,
// each list in its own deterministic order (declaration/document order,
// research.md #5).
type ReferencesResult struct {
	Formal   []ReferenceEntry
	Semantic []ReferenceEntry
}

// References reports every outgoing relationship of rawID: every formal
// one (parent, depends_on, supersedes) exactly as declared in
// frontmatter — no existence check of its own (that remains
// 004-structural-validation's job) — and every semantic one (a
// wikilink) that resolves to exactly one real artifact (FR-001, FR-002).
// rawID must name one of Program, Feature, Spec, Knowledge, or Learning;
// any other type returns ErrInvalidTarget, the same sentinel
// Resolve/Inspect already use for this class of problem (FR-008).
func References(root string, cfg project.Configuration, rawID string) (ReferencesResult, error) {
	result, err := Inspect(root, cfg, rawID)
	if err != nil {
		return ReferencesResult{}, err
	}
	if !referenceableTypes[result.Location.ID.Type] {
		return ReferencesResult{}, fmt.Errorf("%w: %v is not a standalone-referenceable entity type", ErrInvalidTarget, result.Location.ID.Type)
	}

	formal := formalReferences(result.Metadata)

	body, err := artifacts.ReadBody(filepath.Join(root, result.Location.Path))
	if err != nil {
		return ReferencesResult{}, err
	}
	links, err := artifacts.ExtractWikiLinks(body)
	if err != nil {
		return ReferencesResult{}, err
	}

	semantic := make([]ReferenceEntry, 0, len(links))
	for _, link := range links {
		id, paths, err := ids.ResolveTarget(root, cfg, link.Target)
		if err != nil || len(paths) != 1 {
			continue
		}
		semantic = append(semantic, ReferenceEntry{Relation: "wikilink", Target: id})
	}

	return ReferencesResult{Formal: formal, Semantic: semantic}, nil
}

// formalReferences builds meta's outgoing formal relationships in
// declaration order: parent (if any), then each depends_on entry, then
// each supersedes entry, in file order — already stable, since YAML
// list order round-trips through ParseMetadata unchanged.
func formalReferences(meta artifacts.Metadata) []ReferenceEntry {
	formal := make([]ReferenceEntry, 0, 1+len(meta.DependsOn)+len(meta.Supersedes))
	if meta.Parent != nil {
		formal = append(formal, ReferenceEntry{Relation: "parent", Target: *meta.Parent})
	}
	for _, d := range meta.DependsOn {
		formal = append(formal, ReferenceEntry{Relation: "depends_on", Target: d})
	}
	for _, s := range meta.Supersedes {
		formal = append(formal, ReferenceEntry{Relation: "supersedes", Target: s})
	}
	return formal
}
