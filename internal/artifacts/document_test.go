package artifacts_test

import (
	"reflect"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
)

func TestParseDocument_HeadingsAtDifferentLevelsInOrder(t *testing.T) {
	body := "# Title\n\n## Intro\nSome intro text.\n## Details\nMore detail line 1.\nMore detail line 2.\n"

	doc := artifacts.ParseDocument([]byte(body))

	if len(doc.Sections) != 3 {
		t.Fatalf("len(Sections) = %d, want 3: %+v", len(doc.Sections), doc.Sections)
	}
	want := []struct {
		heading   string
		level     int
		body      string
		startLine int
		endLine   int
	}{
		{"Title", 1, "", 1, 2}, // the blank line before "## Intro" belongs to Title's own (empty-rendering) body
		{"Intro", 2, "Some intro text.", 3, 4},
		{"Details", 2, "More detail line 1.\nMore detail line 2.", 5, 7},
	}
	for i, w := range want {
		s := doc.Sections[i]
		if s.Heading != w.heading || s.Level != w.level || s.Body != w.body || s.StartLine != w.startLine || s.EndLine != w.endLine {
			t.Errorf("Sections[%d] = %+v, want %+v", i, s, w)
		}
	}
}

func TestParseDocument_NestedHeadingWithNoOwnContentIsEmpty(t *testing.T) {
	body := "## Requirements\n\n### R1\nReq 1 text.\n### R2\nReq 2 text.\n"

	doc := artifacts.ParseDocument([]byte(body))

	if len(doc.Sections) != 3 {
		t.Fatalf("len(Sections) = %d, want 3: %+v", len(doc.Sections), doc.Sections)
	}
	if doc.Sections[0].Heading != "Requirements" || doc.Sections[0].Body != "" {
		t.Errorf("Sections[0] = %+v, want Heading=Requirements Body=\"\" (all its real content belongs to its children)", doc.Sections[0])
	}
	if doc.Sections[1].Heading != "R1" || doc.Sections[1].Level != 3 || doc.Sections[1].Body != "Req 1 text." {
		t.Errorf("Sections[1] = %+v, want Heading=R1 Level=3 Body=\"Req 1 text.\"", doc.Sections[1])
	}
	if doc.Sections[2].Heading != "R2" || doc.Sections[2].Body != "Req 2 text." {
		t.Errorf("Sections[2] = %+v, want Heading=R2 Body=\"Req 2 text.\"", doc.Sections[2])
	}
}

func TestParseDocument_EmptySectionBetweenTwoHeadings(t *testing.T) {
	body := "## A\n## B\ncontent\n"

	doc := artifacts.ParseDocument([]byte(body))

	if len(doc.Sections) != 2 {
		t.Fatalf("len(Sections) = %d, want 2: %+v", len(doc.Sections), doc.Sections)
	}
	a := doc.Sections[0]
	if a.Heading != "A" || a.Body != "" || a.StartLine != 1 || a.EndLine != 1 {
		t.Errorf("Sections[0] = %+v, want Heading=A Body=\"\" StartLine=1 EndLine=1", a)
	}
	b := doc.Sections[1]
	if b.Heading != "B" || b.Body != "content" || b.StartLine != 2 || b.EndLine != 3 {
		t.Errorf("Sections[1] = %+v, want Heading=B Body=content StartLine=2 EndLine=3", b)
	}
}

func TestParseDocument_NoHeadingsProducesOneSection(t *testing.T) {
	body := "Just prose.\nMore prose.\n"

	doc := artifacts.ParseDocument([]byte(body))

	if len(doc.Sections) != 1 {
		t.Fatalf("len(Sections) = %d, want 1: %+v", len(doc.Sections), doc.Sections)
	}
	s := doc.Sections[0]
	if s.Heading != "" || s.Level != 0 || s.Body != "Just prose.\nMore prose." || s.StartLine != 1 || s.EndLine != 2 {
		t.Errorf("Sections[0] = %+v, unexpected", s)
	}
}

func TestParseDocument_EmptyBodyProducesNoSections(t *testing.T) {
	doc := artifacts.ParseDocument([]byte(""))

	if len(doc.Sections) != 0 {
		t.Errorf("len(Sections) = %d, want 0: %+v", len(doc.Sections), doc.Sections)
	}
}

func TestParseDocument_FencedHeadingLikeTextIsNotABoundaryButStaysInBody(t *testing.T) {
	body := "Prose.\n\n```markdown\n## Not real\n```\n\n## Real\nContent.\n"

	doc := artifacts.ParseDocument([]byte(body))

	if len(doc.Sections) != 2 {
		t.Fatalf("len(Sections) = %d, want 2 (preamble + Real): %+v", len(doc.Sections), doc.Sections)
	}
	preamble := doc.Sections[0]
	if preamble.Heading != "" {
		t.Errorf("Sections[0].Heading = %q, want \"\" (a preamble section)", preamble.Heading)
	}
	wantPreambleBody := "Prose.\n\n```markdown\n## Not real\n```\n"
	if preamble.Body != wantPreambleBody {
		t.Errorf("Sections[0].Body = %q, want %q — the fenced block's content must still be preserved as real body content", preamble.Body, wantPreambleBody)
	}
	real := doc.Sections[1]
	if real.Heading != "Real" || real.Body != "Content." {
		t.Errorf("Sections[1] = %+v, want Heading=Real Body=Content.", real)
	}
}

// TestParseDocument_HeadingAnchorSuffixIsStrippedIntoAnchorField directly
// proves spec 040's User Story 1: an explicit "{#slug}" suffix on a
// heading line is captured into Section.Anchor and removed from
// Section.Heading itself (data-model.md "Section (extended)").
func TestParseDocument_HeadingAnchorSuffixIsStrippedIntoAnchorField(t *testing.T) {
	body := "## Retry Policy {#retry-policy}\nBackoff details.\n"

	doc := artifacts.ParseDocument([]byte(body))

	if len(doc.Sections) != 1 {
		t.Fatalf("len(Sections) = %d, want 1: %+v", len(doc.Sections), doc.Sections)
	}
	s := doc.Sections[0]
	if s.Heading != "Retry Policy" {
		t.Errorf("Heading = %q, want %q (anchor suffix stripped)", s.Heading, "Retry Policy")
	}
	if s.Anchor != "retry-policy" {
		t.Errorf("Anchor = %q, want %q", s.Anchor, "retry-policy")
	}
}

// TestParseDocument_HeadingWithNoAnchorSuffixIsUnaffected proves spec
// 040 FR-008: a heading without "{#...}" is completely unaffected —
// Anchor is empty, Heading unchanged from before this feature.
func TestParseDocument_HeadingWithNoAnchorSuffixIsUnaffected(t *testing.T) {
	body := "## Retry Policy\nBackoff details.\n"

	doc := artifacts.ParseDocument([]byte(body))

	if len(doc.Sections) != 1 {
		t.Fatalf("len(Sections) = %d, want 1: %+v", len(doc.Sections), doc.Sections)
	}
	s := doc.Sections[0]
	if s.Heading != "Retry Policy" {
		t.Errorf("Heading = %q, want %q", s.Heading, "Retry Policy")
	}
	if s.Anchor != "" {
		t.Errorf("Anchor = %q, want \"\" for a heading with no anchor suffix", s.Anchor)
	}
}

// TestParseDocument_MalformedAnchorSuffixStaysLiteralHeadingText proves
// spec 040's data-model.md "Section" validation rule: an unterminated
// "{#" (no closing "}") is never an error — ParseDocument stays a pure,
// always-succeeding function — and is left as literal heading text with
// Anchor empty.
func TestParseDocument_MalformedAnchorSuffixStaysLiteralHeadingText(t *testing.T) {
	body := "## Retry Policy {#\nBackoff details.\n"

	doc := artifacts.ParseDocument([]byte(body))

	if len(doc.Sections) != 1 {
		t.Fatalf("len(Sections) = %d, want 1: %+v", len(doc.Sections), doc.Sections)
	}
	s := doc.Sections[0]
	if s.Anchor != "" {
		t.Errorf("Anchor = %q, want \"\" for a malformed/unterminated anchor suffix", s.Anchor)
	}
	if s.Heading != "Retry Policy {#" {
		t.Errorf("Heading = %q, want the literal unterminated text preserved: %q", s.Heading, "Retry Policy {#")
	}
}

func TestParseDocument_Deterministic(t *testing.T) {
	body := "# T\n\n## A\ntext a\n## B\ntext b\n"

	first := artifacts.ParseDocument([]byte(body))
	second := artifacts.ParseDocument([]byte(body))

	if !reflect.DeepEqual(first, second) {
		t.Errorf("ParseDocument() is not deterministic: %+v != %+v", first, second)
	}
}
