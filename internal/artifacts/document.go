package artifacts

import (
	"regexp"
	"strings"
)

// Section is one heading-bounded region of a Document, in document
// order (docs/context-engine-implementation.md §8). Content between one
// heading (at any level) and the next heading (at any level) belongs
// exclusively to the first heading's own Section — a parent heading with
// child headings never duplicates its children's own content into its
// own Body (013-document-model-chunking/research.md #2).
type Section struct {
	// Heading is the heading text, without the leading "#"s and without
	// any trailing explicit anchor suffix (see Anchor). Empty for the
	// untitled preamble section (content before the first heading, or
	// the whole body when there is no heading at all).
	Heading string
	// Anchor is the explicit, stable section identifier declared as a
	// trailing "{#slug}" suffix on the heading's own line (040-stable-
	// section-anchors research.md #1), e.g. "## Retry Policy
	// {#retry-policy}" yields Heading "Retry Policy" and Anchor
	// "retry-policy". Empty when no such suffix is present — the
	// overwhelming majority of headings. An unterminated/malformed
	// suffix (no closing "}") is never stripped: it stays literal
	// Heading text and Anchor remains empty (ParseDocument's own
	// always-succeeds contract, data-model.md "Section").
	Anchor string
	// Level is 1-6 for a real ATX heading; 0 for the untitled preamble
	// section.
	Level int
	// Body is the raw content strictly between this heading's own line
	// and the next heading (at any level), never trimmed
	// (research.md #8). Empty when a heading is immediately followed by
	// another heading.
	Body string
	// StartLine is 1-indexed: the heading's own line, or 1 for the
	// preamble section.
	StartLine int
	// EndLine is 1-indexed: the last line included in Body; equals
	// StartLine when Body is empty.
	EndLine int
}

// Document is body's structured representation — an ordered, flat list
// of Sections. Distinct from Metadata (frontmatter), which already has
// its own separate, unaffected representation; see ParseMetadata.
type Document struct {
	Sections []Section
}

// headingPattern matches an ATX Markdown heading line ("#" through
// "######", followed by at least one space) — the only heading syntax
// this project's own content uses (research.md #4). A line starting
// with "#" but no following space (e.g. a prose "#123" reference) is
// correctly not matched, per CommonMark's own ATX rule.
var headingPattern = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)

// anchorSuffixPattern matches a trailing explicit-anchor suffix on an
// already-trimmed heading text — "{#slug}" preceded by optional
// whitespace, anchored to the end of the string (040-stable-section-
// anchors research.md #1). The slug itself ([^}]+) is opaque author
// text: internal/validation, not this parser, judges its uniqueness
// (data-model.md "Section"). An unterminated "{#" with no closing "}"
// simply does not match, leaving the heading's literal text untouched —
// ParseDocument never errors on malformed input.
var anchorSuffixPattern = regexp.MustCompile(`\s*\{#([^}]+)\}$`)

// splitHeadingAnchor separates heading's own trailing "{#slug}" suffix
// (if any) from its displayed title text.
func splitHeadingAnchor(heading string) (title, anchor string) {
	if m := anchorSuffixPattern.FindStringSubmatch(heading); m != nil {
		return strings.TrimSpace(heading[:len(heading)-len(m[0])]), m[1]
	}
	return heading, ""
}

// ParseDocument splits body into an ordered, flat list of Sections using
// ATX headings as boundaries (FR-001, FR-002). Heading-like text inside
// a fenced code block is never treated as a boundary — reusing
// fencedLines, the same fence-tracking 011's own ExtractWikiLinks
// already established — though the fenced line's own text still
// contributes to whichever Section's Body is currently open (FR-003). A
// body with no headings at all still produces exactly one Section
// covering the entire body (FR-004); a completely empty body produces
// zero Sections (FR-011). Pure function of body — no I/O, no path,
// deterministic output for the same input every time (FR-008, FR-010).
func ParseDocument(body []byte) Document {
	if len(body) == 0 {
		return Document{}
	}

	lines := strings.Split(string(body), "\n")
	// strings.Split on a body ending in "\n" (every artifact body does)
	// yields one trailing empty element that isn't a real line — drop
	// it so line numbers match how a person would count this artifact's
	// own lines.
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}
	if len(lines) == 0 {
		return Document{}
	}

	fenced := fencedLines(lines)

	var sections []Section
	var cur *Section
	var bodyLines []string

	flush := func(endLine int) {
		if cur == nil {
			return
		}
		cur.Body = strings.Join(bodyLines, "\n")
		cur.EndLine = endLine
		sections = append(sections, *cur)
		cur = nil
		bodyLines = nil
	}

	for i, line := range lines {
		lineNum := i + 1

		if !fenced[i] {
			if m := headingPattern.FindStringSubmatch(strings.TrimRight(line, "\r")); m != nil {
				flush(lineNum - 1)
				title, anchor := splitHeadingAnchor(strings.TrimSpace(m[2]))
				cur = &Section{
					Heading:   title,
					Anchor:    anchor,
					Level:     len(m[1]),
					StartLine: lineNum,
				}
				bodyLines = nil
				continue
			}
		}

		if cur == nil {
			cur = &Section{Heading: "", Level: 0, StartLine: 1}
		}
		bodyLines = append(bodyLines, line)
	}
	flush(len(lines))

	return Document{Sections: sections}
}
