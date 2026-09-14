package validation

import (
	"fmt"
	"path/filepath"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// checkWikilinks reads the artifact at filePath's body, extracts every
// wikilink from it (internal/artifacts.ExtractWikiLinks), and classifies
// each target into exactly one of CodeInvalidWikilink, CodeBrokenWikilink,
// CodeAmbiguousWikilink, or no finding at all (data-model.md's
// classification table). An alias never affects classification — only
// Target is ever inspected (§6.1). An artifact whose body contains no
// links contributes nothing (FR-009, SC-003).
func checkWikilinks(root string, cfg project.Configuration, filePath string) []Finding {
	body, err := artifacts.ReadBody(filepath.Join(root, filePath))
	if err != nil {
		// checkEntity's own artifacts.ParseMetadata call already reports a
		// malformed/missing artifact via its own Finding before this is
		// ever reached in practice; there is nothing further to say here.
		return nil
	}

	links, err := artifacts.ExtractWikiLinks(body)
	if err != nil || len(links) == 0 {
		return nil
	}

	var findings []Finding
	for _, link := range links {
		if f, ok := classifyWikilink(root, cfg, filePath, link); ok {
			findings = append(findings, f)
		}
	}
	return findings
}

// classifyWikilink resolves one link's Target against the project's real
// artifacts, returning the single Finding that applies (if any) — never
// more than one per link (FR-006, FR-007, FR-008).
func classifyWikilink(root string, cfg project.Configuration, filePath string, link artifacts.WikiLink) (Finding, bool) {
	id, err := ids.ParseAny(link.Target)
	if err != nil {
		return Finding{
			Code:     CodeInvalidWikilink,
			Severity: SeverityError,
			Path:     filePath,
			Message:  fmt.Sprintf("wikilink target %q is not a well-formed entity ID: %v", link.Target, err),
		}, true
	}

	result, err := ids.Scan(root, cfg, id.Type)
	if err != nil {
		return Finding{}, false
	}

	switch matches := len(result.Paths[id.Number]); {
	case matches == 0:
		return Finding{
			Code:     CodeBrokenWikilink,
			Severity: SeverityError,
			Path:     filePath,
			Message:  fmt.Sprintf("wikilink target %v does not exist", id),
		}, true
	case matches > 1:
		return Finding{
			Code:     CodeAmbiguousWikilink,
			Severity: SeverityError,
			Path:     filePath,
			Message:  fmt.Sprintf("wikilink target %v resolves to %d artifacts: %v", id, matches, result.Paths[id.Number]),
		}, true
	default:
		return Finding{}, false
	}
}
