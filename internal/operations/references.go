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
	// SourcePath is the queried (referencing) artifact's own path —
	// always populated, regardless of Relation, since it is already
	// known to this function (038-wikilink-chunk-provenance data-
	// model.md "ReferenceEntry / BacklinkEntry (extended)").
	SourcePath string
	// SourceSection/SourceLine are the wikilink occurrence's own
	// enclosing section/line (artifacts.OccurrenceFor) — populated only
	// when Relation == "wikilink"; empty/zero for a formal entry, which
	// has no line-level wikilink origin (spec FR-009).
	SourceSection string
	SourceLine    int
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

	formal := formalReferences(result.Metadata, result.Location.Path)

	body, bodyStartLine, err := artifacts.ReadBodyWithOffset(filepath.Join(root, result.Location.Path))
	if err != nil {
		return ReferencesResult{}, err
	}
	offset := bodyStartLine - 1
	links, err := artifacts.ExtractWikiLinks(body)
	if err != nil {
		return ReferencesResult{}, err
	}
	chunks := artifacts.ChunksWithOffset(result.Location.Path, body, offset)

	semantic := make([]ReferenceEntry, 0, len(links))
	for _, link := range links {
		id, paths, err := ids.ResolveTarget(root, cfg, link.Target)
		if err != nil || len(paths) != 1 {
			continue
		}
		link.Line += offset
		occ := artifacts.OccurrenceFor(link, chunks)
		semantic = append(semantic, ReferenceEntry{
			Relation:      "wikilink",
			Target:        id,
			SourcePath:    result.Location.Path,
			SourceSection: occ.SourceSection,
			SourceLine:    occ.SourceLine,
		})
	}

	return ReferencesResult{Formal: formal, Semantic: semantic}, nil
}

// formalReferences builds meta's outgoing formal relationships in
// declaration order: parent (if any), then each depends_on entry, then
// each supersedes entry, in file order — already stable, since YAML
// list order round-trips through ParseMetadata unchanged. sourcePath is
// the queried artifact's own path, attached to every entry regardless
// of relation kind (data-model.md: "SourcePath is always populated").
func formalReferences(meta artifacts.Metadata, sourcePath string) []ReferenceEntry {
	formal := make([]ReferenceEntry, 0, 1+len(meta.DependsOn)+len(meta.Supersedes))
	if meta.Parent != nil {
		formal = append(formal, ReferenceEntry{Relation: "parent", Target: *meta.Parent, SourcePath: sourcePath})
	}
	for _, d := range meta.DependsOn {
		formal = append(formal, ReferenceEntry{Relation: "depends_on", Target: d, SourcePath: sourcePath})
	}
	for _, s := range meta.Supersedes {
		formal = append(formal, ReferenceEntry{Relation: "supersedes", Target: s, SourcePath: sourcePath})
	}
	return formal
}
