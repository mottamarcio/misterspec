package operations

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// backlinkScopeTypes are the five entity types Backlinks scans, in a
// fixed order — the same order internal/validation/validator.go's own
// projectEntityTypes already uses — so BacklinksResult's own ordering is
// deterministic by construction rather than a final sort call
// (research.md #5).
var backlinkScopeTypes = []ids.EntityType{ids.Program, ids.Feature, ids.Spec, ids.Knowledge, ids.Learning}

// BacklinkEntry is one incoming relationship into a queried artifact
// (data-model.md).
type BacklinkEntry struct {
	// Relation is "parent", "depends_on", "supersedes" (formal) or
	// "wikilink" (semantic) — same four values as ReferenceEntry.Relation.
	Relation string
	// Source is the referencing artifact's own declared ID (always
	// non-nil for these five id-bearing types once ParseMetadata
	// succeeds).
	Source ids.EntityID
}

// BacklinksResult groups a queried artifact's incoming relationships,
// discovered by a full scan of the five scoped types, in a fixed
// deterministic type-then-number order.
type BacklinksResult struct {
	Formal   []BacklinkEntry
	Semantic []BacklinkEntry
}

// Backlinks reports every other artifact that formally or semantically
// references rawID (FR-004) — computed directly from filesystem state on
// every call, never depending on any index or cache existing
// (docs/context-engine-implementation.md §7.2). Same target-type scope
// and error behavior as References (FR-008).
func Backlinks(root string, cfg project.Configuration, rawID string) (BacklinksResult, error) {
	target, err := Inspect(root, cfg, rawID)
	if err != nil {
		return BacklinksResult{}, err
	}
	if !referenceableTypes[target.Location.ID.Type] {
		return BacklinksResult{}, fmt.Errorf("%w: %v is not a standalone-referenceable entity type", ErrInvalidTarget, target.Location.ID.Type)
	}
	targetID := target.Location.ID

	var formal, semantic []BacklinkEntry
	for _, t := range backlinkScopeTypes {
		scanResult, err := ids.Scan(root, cfg, t)
		if err != nil {
			return BacklinksResult{}, err
		}

		for _, number := range sortedScanNumbers(scanResult.Paths) {
			for _, scanPath := range scanResult.Paths[number] {
				filePath := scanPath
				if fn, ok := canonicalFilename(t); ok {
					filePath = scanPath + "/" + fn
				}

				meta, err := artifacts.ParseMetadata(filepath.Join(root, filePath))
				if err != nil || meta.ID == nil {
					// A malformed source, or one missing its own required
					// ID, is already 004-structural-validation's problem to
					// report — not this operation's (research.md #4-style
					// separation of concerns, mirrored here).
					continue
				}
				source := *meta.ID

				formal = append(formal, matchingFormalBacklinks(meta, targetID, source)...)

				body, err := artifacts.ReadBody(filepath.Join(root, filePath))
				if err != nil {
					continue
				}
				links, err := artifacts.ExtractWikiLinks(body)
				if err != nil {
					continue
				}
				for _, link := range links {
					id, paths, err := ids.ResolveTarget(root, cfg, link.Target)
					if err != nil || len(paths) != 1 || id != targetID {
						continue
					}
					semantic = append(semantic, BacklinkEntry{Relation: "wikilink", Source: source})
				}
			}
		}
	}

	return BacklinksResult{Formal: formal, Semantic: semantic}, nil
}

// matchingFormalBacklinks reports every formal relationship in meta that
// names target, labeled with source (the artifact meta itself belongs
// to).
func matchingFormalBacklinks(meta artifacts.Metadata, target ids.EntityID, source ids.EntityID) []BacklinkEntry {
	var out []BacklinkEntry
	if meta.Parent != nil && *meta.Parent == target {
		out = append(out, BacklinkEntry{Relation: "parent", Source: source})
	}
	for _, d := range meta.DependsOn {
		if d == target {
			out = append(out, BacklinkEntry{Relation: "depends_on", Source: source})
		}
	}
	for _, s := range meta.Supersedes {
		if s == target {
			out = append(out, BacklinkEntry{Relation: "supersedes", Source: source})
		}
	}
	return out
}

// sortedScanNumbers returns paths's keys in ascending order — the same
// small, package-local pattern internal/validation/validator.go's own
// sortedNumbers already establishes (research.md #5; not promoted to a
// shared package, matching the canonicalFilename precedent for a
// genuinely tiny helper).
func sortedScanNumbers(paths map[int][]string) []int {
	numbers := make([]int, 0, len(paths))
	for n := range paths {
		numbers = append(numbers, n)
	}
	sort.Ints(numbers)
	return numbers
}
