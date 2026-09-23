package validation

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"

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
// more than one per link (FR-006, FR-007, FR-008). Delegates the actual
// resolution to ids.ResolveTarget (012-references-backlinks/research.md
// #2) rather than repeating its ParseAny+Scan composition inline —
// behavior here is unchanged: a malformed target (ids.ErrInvalidIDSyntax)
// is still CodeInvalidWikilink; any other resolution error (e.g. a
// filesystem-level Scan failure) is still silently swallowed, exactly as
// before this refactor, since that case is not expected in practice and
// was never reported as a Finding.
func classifyWikilink(root string, cfg project.Configuration, filePath string, link artifacts.WikiLink) (Finding, bool) {
	id, paths, err := ids.ResolveTarget(root, cfg, link.Target)
	if err != nil {
		if errors.Is(err, ids.ErrInvalidIDSyntax) {
			return Finding{
				Code:     CodeInvalidWikilink,
				Severity: SeverityError,
				Path:     filePath,
				Message:  fmt.Sprintf("wikilink target %q is not a well-formed entity ID: %v", link.Target, err),
			}, true
		}
		return Finding{}, false
	}

	switch matches := len(paths); {
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
			Message:  fmt.Sprintf("wikilink target %v resolves to %d artifacts: %v", id, matches, paths),
		}, true
	case link.Anchor != "":
		// The target artifact exists (matches == 1 here, by exhaustion
		// of the switch's own cases) — confirm it declares a Section
		// with a matching Anchor (040-stable-section-anchors contracts
		// §2). Reading the resolved artifact's own body a second time
		// here (rather than plumbing it through from checkWikilinks) is
		// deliberate: checkWikilinks only reads filePath's own body, not
		// every target a link might name, and this branch is reached
		// only for an anchor-qualified link — the common case (no
		// anchor) pays nothing extra.
		if anchorExists(root, paths[0], id.Type, link.Anchor) {
			return Finding{}, false
		}
		return Finding{
			Code:     CodeUnknownAnchor,
			Severity: SeverityError,
			Path:     filePath,
			Message:  fmt.Sprintf("wikilink target %v exists but declares no anchor %q", id, link.Anchor),
		}, true
	default:
		return Finding{}, false
	}
}

// anchorExists reports whether the artifact at root/targetPath declares
// a Section whose own Anchor equals anchor. targetPath, as returned by
// ids.ResolveTarget/Scan, is a directory for a directory-scoped type
// (Program/Feature/Spec) — canonicalFilename resolves it down to the
// real canonical file, exactly as operations.Inspect/Resolve already
// do; a Knowledge/Learning targetPath is already the file itself. A
// read failure is treated as "not found" — the target's own existence
// was already confirmed by classifyWikilink's caller; a read error
// here is not expected in practice and is not itself this function's
// Finding to report.
func anchorExists(root, targetPath string, t ids.EntityType, anchor string) bool {
	full := targetPath
	if fn, ok := canonicalFilename(t); ok {
		full = filepath.Join(targetPath, fn)
	}
	body, err := artifacts.ReadBody(filepath.Join(root, full))
	if err != nil {
		return false
	}
	doc := artifacts.ParseDocument(body)
	for _, s := range doc.Sections {
		if s.Anchor == anchor {
			return true
		}
	}
	return false
}

// checkAnchors scans the artifact at filePath's own Sections for a
// repeated non-empty Anchor declaration (040-stable-section-anchors
// contracts §2) — independent of whether anything currently references
// that anchor (spec Assumptions: an unreferenced anchor is not itself
// an error; a duplicated one is).
func checkAnchors(root, filePath string) []Finding {
	body, err := artifacts.ReadBody(filepath.Join(root, filePath))
	if err != nil {
		return nil
	}
	doc := artifacts.ParseDocument(body)

	seen := map[string]bool{}
	var duplicated map[string]bool
	for _, s := range doc.Sections {
		if s.Anchor == "" {
			continue
		}
		if seen[s.Anchor] {
			if duplicated == nil {
				duplicated = map[string]bool{}
			}
			duplicated[s.Anchor] = true
			continue
		}
		seen[s.Anchor] = true
	}

	var findings []Finding
	for _, anchor := range sortedKeys(duplicated) {
		findings = append(findings, Finding{
			Code:     CodeDuplicateAnchor,
			Severity: SeverityError,
			Path:     filePath,
			Message:  fmt.Sprintf("anchor %q is declared by more than one Section", anchor),
		})
	}
	return findings
}

// sortedKeys returns m's keys in ascending order — deterministic
// Finding order across runs (matching this package's own existing
// determinism discipline), not relying on Go's randomized map
// iteration order.
func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
