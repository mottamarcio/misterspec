package artifacts

import "strings"

// WikiLink is one explicit, author-written semantic link from an
// artifact's body to another artifact's ID
// (docs/context-engine-implementation.md §6.2), written as
// "[[TARGET]]" or "[[TARGET|Alias]]".
type WikiLink struct {
	// Target is the raw entity-ID token, exactly as written — not yet
	// validated as a real entity ID; see internal/validation for that
	// separate step (FR-004). Never includes an "#anchor" suffix (see
	// Anchor) or "|Alias" text.
	Target string
	// Anchor is the "#anchor" token written between Target and any
	// "|Alias" (040-stable-section-anchors research.md #2), e.g.
	// "[[KNOW-003#retry-policy]]" or "[[KNOW-003#retry-policy|Alias]]".
	// Empty when absent — the overwhelming majority of wikilinks,
	// including every one written before this feature existed (spec
	// FR-008). Raw, unvalidated text — internal/validation checks its
	// existence separately (data-model.md "WikiLink (extended)").
	Anchor string
	// Alias is the display alias, if written; empty otherwise.
	// Presentation-only — never affects resolution (§6.1).
	Alias string
	// Line is the 1-indexed line number, relative to the body passed
	// to ExtractWikiLinks — not necessarily file-absolute
	// (specs/011-wikilink-foundation/research.md).
	Line int
}

// ExtractWikiLinks parses body for "[[TARGET]]" and "[[TARGET|Alias]]"
// occurrences, in document order (FR-001, FR-002). It is a pure
// lexical scan: it never resolves a target against real project state
// (FR-004) — that is internal/validation's separate job. It never
// treats a standard Markdown link, text inside an inline code span or
// fenced code block, or an unterminated "[[" as a link (FR-003).
//
// A wikilink must be fully contained on one line — ExtractWikiLinks
// never attempts to match a "[[" against a "]]" on a later line
// (research.md).
func ExtractWikiLinks(body []byte) ([]WikiLink, error) {
	var links []WikiLink

	lines := strings.Split(string(body), "\n")
	fenced := fencedLines(lines)

	for i, line := range lines {
		if fenced[i] {
			continue // a fence delimiter line, or fenced content, never contains a link
		}

		for _, link := range extractLineLinks(line) {
			link.Line = i + 1
			links = append(links, link)
		}
	}

	return links, nil
}

// fencedLines reports, for each line of lines, whether that line is
// either a fenced-code-block delimiter itself or falls inside one —
// both cases where the line's own content must never be treated as real
// wikilink or heading syntax. Shared by ExtractWikiLinks and
// internal/artifacts.ParseDocument (013-document-model-chunking/
// research.md #3) — factored out of ExtractWikiLinks's own original
// inline fence-tracking loop, behavior-identical.
func fencedLines(lines []string) []bool {
	fenced := make([]bool, len(lines))

	var inFence bool
	var fenceMarker string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if marker, ok := fenceMarkerOf(trimmed); ok {
			fenced[i] = true
			switch {
			case !inFence:
				inFence, fenceMarker = true, marker
			case marker == fenceMarker:
				inFence, fenceMarker = false, ""
			}
			continue
		}
		fenced[i] = inFence
	}

	return fenced
}

// fenceMarkerOf reports whether trimmed opens or closes a fenced code
// block (a line starting with "```" or "~~~"), and the marker itself —
// this project's own Markdown content never mixes fence styles or
// varies fence length, so a simple prefix check (rather than full
// CommonMark fence-length matching) is sufficient (research.md).
func fenceMarkerOf(trimmed string) (marker string, ok bool) {
	if strings.HasPrefix(trimmed, "```") {
		return "```", true
	}
	if strings.HasPrefix(trimmed, "~~~") {
		return "~~~", true
	}
	return "", false
}

// extractLineLinks finds every complete "[[...]]" pair in line outside
// of any inline code span, splitting each into target/alias. It never
// judges whether the inner content is a well-formed entity ID — an
// empty or malformed target is still extracted as a WikiLink (with an
// empty or malformed Target string); internal/validation is where that
// distinction is made (FR-004).
func extractLineLinks(line string) []WikiLink {
	masked := maskInlineCodeSpans(line)

	var links []WikiLink
	i := 0
	for i < len(line)-1 {
		if masked[i] || line[i] != '[' || line[i+1] != '[' {
			i++
			continue
		}

		closeIdx := -1
		for j := i + 2; j < len(line)-1; j++ {
			if !masked[j] && line[j] == ']' && line[j+1] == ']' {
				closeIdx = j
				break
			}
		}
		if closeIdx == -1 {
			// No closing "]]" anywhere later on this line — not a
			// link, not even a malformed one (spec.md's own Edge
			// Case). Advance past this "[[" and keep scanning.
			i += 2
			continue
		}

		targetAndAnchor, alias := splitTargetAlias(line[i+2 : closeIdx])
		target, anchor := splitTargetAnchor(targetAndAnchor)
		links = append(links, WikiLink{Target: target, Anchor: anchor, Alias: alias})
		i = closeIdx + 2
	}
	return links
}

// splitTargetAlias splits a wikilink's inner content on its first "|",
// if any.
func splitTargetAlias(inner string) (target, alias string) {
	if idx := strings.Index(inner, "|"); idx >= 0 {
		return inner[:idx], inner[idx+1:]
	}
	return inner, ""
}

// splitTargetAnchor splits the pre-alias portion of a wikilink's inner
// content on its first "#", if any (040-stable-section-anchors
// research.md #2) — always applied after splitTargetAlias, so an
// alias's own text is never mistakenly scanned for "#".
func splitTargetAnchor(targetAndAnchor string) (target, anchor string) {
	if idx := strings.Index(targetAndAnchor, "#"); idx >= 0 {
		return targetAndAnchor[:idx], targetAndAnchor[idx+1:]
	}
	return targetAndAnchor, ""
}

// maskInlineCodeSpans returns a same-length boolean slice marking every
// byte of line that falls within a single-backtick-delimited inline
// code span (the only style this project's own Markdown content uses —
// research.md) — those positions are never considered for wikilink
// matching. An unterminated backtick (no matching close on the same
// line) masks nothing, since it never actually forms a code span.
//
// Byte-level indexing is safe here even for a line containing
// multi-byte UTF-8 runes: '`', '[', ']', and '|' are all single-byte
// ASCII values, which a UTF-8 continuation byte can never equal.
func maskInlineCodeSpans(line string) []bool {
	masked := make([]bool, len(line))

	inSpan := false
	start := 0
	for i := 0; i < len(line); i++ {
		if line[i] != '`' {
			continue
		}
		if !inSpan {
			inSpan, start = true, i
			continue
		}
		for k := start; k <= i; k++ {
			masked[k] = true
		}
		inSpan = false
	}

	return masked
}
