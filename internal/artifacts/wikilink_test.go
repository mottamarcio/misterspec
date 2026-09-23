package artifacts_test

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
)

func TestExtractWikiLinks_SimpleLink(t *testing.T) {
	links, err := artifacts.ExtractWikiLinks([]byte("See [[SPEC-014]] for details.\n"))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("len(links) = %d, want 1: %+v", len(links), links)
	}
	if links[0].Target != "SPEC-014" || links[0].Alias != "" {
		t.Errorf("links[0] = %+v, want Target=SPEC-014, Alias=\"\"", links[0])
	}
	if links[0].Line != 1 {
		t.Errorf("links[0].Line = %d, want 1", links[0].Line)
	}
}

func TestExtractWikiLinks_AliasedLink(t *testing.T) {
	links, err := artifacts.ExtractWikiLinks([]byte("[[SPEC-014|Refresh Token Rotation]]\n"))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("len(links) = %d, want 1: %+v", len(links), links)
	}
	if links[0].Target != "SPEC-014" || links[0].Alias != "Refresh Token Rotation" {
		t.Errorf("links[0] = %+v, unexpected", links[0])
	}
}

func TestExtractWikiLinks_MultipleLinksOneLine(t *testing.T) {
	links, err := artifacts.ExtractWikiLinks([]byte("See [[SPEC-011]] and [[SPEC-014]].\n"))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("len(links) = %d, want 2: %+v", len(links), links)
	}
	if links[0].Target != "SPEC-011" || links[1].Target != "SPEC-014" {
		t.Errorf("links = %+v, want [SPEC-011, SPEC-014] in order", links)
	}
}

func TestExtractWikiLinks_MultipleLinesInOrderWithCorrectLines(t *testing.T) {
	body := "" +
		"## Relevant Knowledge\n" +
		"\n" +
		"- [[KNOW-003|Authentication Model]]\n" +
		"- [[KNOW-008]]\n" +
		"\n" +
		"## Related Specs\n" +
		"\n" +
		"- [[SPEC-011]]\n"

	links, err := artifacts.ExtractWikiLinks([]byte(body))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 3 {
		t.Fatalf("len(links) = %d, want 3: %+v", len(links), links)
	}

	want := []struct {
		target string
		line   int
	}{
		{"KNOW-003", 3},
		{"KNOW-008", 4},
		{"SPEC-011", 8},
	}
	for i, w := range want {
		if links[i].Target != w.target || links[i].Line != w.line {
			t.Errorf("links[%d] = %+v, want Target=%s Line=%d", i, links[i], w.target, w.line)
		}
	}
}

func TestExtractWikiLinks_StandardMarkdownLinkIsNotExtracted(t *testing.T) {
	links, err := artifacts.ExtractWikiLinks([]byte("See [the spec](https://example.com/SPEC-014) for details.\n"))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("links = %+v, want none — a standard Markdown link is not a wikilink", links)
	}
}

func TestExtractWikiLinks_InlineCodeSpanIsExcluded(t *testing.T) {
	links, err := artifacts.ExtractWikiLinks([]byte("Example syntax: `[[SPEC-014]]` is how you link.\n"))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("links = %+v, want none — text inside an inline code span is not a wikilink", links)
	}
}

func TestExtractWikiLinks_FencedCodeBlockIsExcluded(t *testing.T) {
	body := "" +
		"Some prose.\n" +
		"\n" +
		"```markdown\n" +
		"[[SPEC-014]]\n" +
		"```\n" +
		"\n" +
		"[[SPEC-011]]\n"

	links, err := artifacts.ExtractWikiLinks([]byte(body))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 1 || links[0].Target != "SPEC-011" {
		t.Errorf("links = %+v, want exactly [SPEC-011] — the fenced block's content must be excluded", links)
	}
}

// TestExtractWikiLinks_TildeFencedCodeBlockIsExcluded pins down current
// behavior before 013-document-model-chunking's planned fencedLines
// extraction: "~~~" fences work exactly like "```" ones.
func TestExtractWikiLinks_TildeFencedCodeBlockIsExcluded(t *testing.T) {
	body := "" +
		"~~~markdown\n" +
		"[[SPEC-014]]\n" +
		"~~~\n" +
		"[[SPEC-011]]\n"

	links, err := artifacts.ExtractWikiLinks([]byte(body))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 1 || links[0].Target != "SPEC-011" {
		t.Errorf("links = %+v, want exactly [SPEC-011]", links)
	}
}

// TestExtractWikiLinks_MismatchedFenceMarkersDoNotClose pins down
// current behavior before the fencedLines extraction: a "```" fence is
// only closed by another "```" line, never by a "~~~" one — the fence
// stays open through the rest of the body in that case.
func TestExtractWikiLinks_MismatchedFenceMarkersDoNotClose(t *testing.T) {
	body := "" +
		"```markdown\n" +
		"[[SPEC-014]]\n" +
		"~~~\n" +
		"[[SPEC-011]]\n"

	links, err := artifacts.ExtractWikiLinks([]byte(body))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("links = %+v, want none — the \"```\" fence is never closed by a \"~~~\" line", links)
	}
}

func TestExtractWikiLinks_UnterminatedBracketsAreNotALink(t *testing.T) {
	links, err := artifacts.ExtractWikiLinks([]byte("This has a stray [[SPEC-014 with no closing brackets.\n"))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("links = %+v, want none — an unterminated \"[[\" is not a link, not even a malformed one", links)
	}
}

func TestExtractWikiLinks_BareSingleBracketIsNotExtracted(t *testing.T) {
	links, err := artifacts.ExtractWikiLinks([]byte("A footnote reference [1] is not a link.\n"))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("links = %+v, want none", links)
	}
}

func TestExtractWikiLinks_EveryEntityPrefixRoundTrips(t *testing.T) {
	body := "" +
		"[[PRG-001]]\n" +
		"[[FEAT-002]]\n" +
		"[[SPEC-014]]\n" +
		"[[TASK-003]]\n" +
		"[[KNOW-008]]\n" +
		"[[LRN-001]]\n"

	links, err := artifacts.ExtractWikiLinks([]byte(body))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	want := []string{"PRG-001", "FEAT-002", "SPEC-014", "TASK-003", "KNOW-008", "LRN-001"}
	if len(links) != len(want) {
		t.Fatalf("len(links) = %d, want %d: %+v", len(links), len(want), links)
	}
	for i, w := range want {
		if links[i].Target != w {
			t.Errorf("links[%d].Target = %q, want %q", i, links[i].Target, w)
		}
	}
}

// TestExtractWikiLinks_AnchorQualifiedLink proves spec 040 research.md
// #2: "[[ID#anchor]]" splits into Target/Anchor, with Alias empty.
func TestExtractWikiLinks_AnchorQualifiedLink(t *testing.T) {
	links, err := artifacts.ExtractWikiLinks([]byte("See [[KNOW-003#retry-policy]] for details.\n"))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("len(links) = %d, want 1: %+v", len(links), links)
	}
	if links[0].Target != "KNOW-003" || links[0].Anchor != "retry-policy" || links[0].Alias != "" {
		t.Errorf("links[0] = %+v, want Target=KNOW-003 Anchor=retry-policy Alias=\"\"", links[0])
	}
}

// TestExtractWikiLinks_AnchorAndAliasCombined proves the alias split
// happens first, then the anchor split, per research.md #2 — both
// fields resolve correctly when combined.
func TestExtractWikiLinks_AnchorAndAliasCombined(t *testing.T) {
	links, err := artifacts.ExtractWikiLinks([]byte("[[KNOW-003#retry-policy|Política de retries]]\n"))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("len(links) = %d, want 1: %+v", len(links), links)
	}
	want := artifacts.WikiLink{Target: "KNOW-003", Anchor: "retry-policy", Alias: "Política de retries", Line: 1}
	if links[0].Target != want.Target || links[0].Anchor != want.Anchor || links[0].Alias != want.Alias {
		t.Errorf("links[0] = %+v, want %+v", links[0], want)
	}
}

// TestExtractWikiLinks_NonAnchorLinksAreUnaffected proves spec 040
// FR-008: both existing wikilink forms are completely unaffected —
// Anchor stays empty for each.
func TestExtractWikiLinks_NonAnchorLinksAreUnaffected(t *testing.T) {
	links, err := artifacts.ExtractWikiLinks([]byte("[[KNOW-003]] and [[KNOW-003|Alias Text]]\n"))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("len(links) = %d, want 2: %+v", len(links), links)
	}
	if links[0].Target != "KNOW-003" || links[0].Anchor != "" || links[0].Alias != "" {
		t.Errorf("links[0] = %+v, want Target=KNOW-003 Anchor=\"\" Alias=\"\"", links[0])
	}
	if links[1].Target != "KNOW-003" || links[1].Anchor != "" || links[1].Alias != "Alias Text" {
		t.Errorf("links[1] = %+v, want Target=KNOW-003 Anchor=\"\" Alias=\"Alias Text\"", links[1])
	}
}

func TestExtractWikiLinks_EmptyBody(t *testing.T) {
	links, err := artifacts.ExtractWikiLinks([]byte(""))
	if err != nil {
		t.Fatalf("ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("links = %+v, want none for an empty body", links)
	}
}
