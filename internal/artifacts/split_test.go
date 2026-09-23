package artifacts_test

import (
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
)

func TestSplitIntoCoherentUnits_ConcatenationReproducesOriginal(t *testing.T) {
	content := "First paragraph, one line.\n\nSecond paragraph,\nspanning two lines.\n\n```go\nfunc f() {\n\treturn 1\n}\n```\n\nTrailing paragraph.\n"

	units := artifacts.SplitIntoCoherentUnits(content)

	var rebuilt strings.Builder
	for _, u := range units {
		rebuilt.WriteString(u)
	}
	if rebuilt.String() != content {
		t.Errorf("concatenated units = %q, want %q (a pure partition — FR-011)", rebuilt.String(), content)
	}
}

func TestSplitIntoCoherentUnits_NeverSplitsInsideAFence(t *testing.T) {
	content := "Intro.\n\n```go\nfunc f() {\n\treturn 1\n}\n```\n\nOutro.\n"

	units := artifacts.SplitIntoCoherentUnits(content)

	for _, u := range units {
		opens := strings.Count(u, "```")
		if opens == 1 {
			t.Errorf("unit %q contains an unmatched fence delimiter — a fence's own opening/closing pair must stay in one unit (FR-011): units = %+v", u, units)
		}
	}

	// The fence's own content must appear together, in one unit.
	found := false
	for _, u := range units {
		if strings.Contains(u, "func f() {") && strings.Contains(u, "```go") && strings.Contains(u, "```\n") {
			found = true
		}
	}
	if !found {
		t.Errorf("no single unit contains the whole fenced block: units = %+v", units)
	}
}

func TestSplitIntoCoherentUnits_EmptyContentIsEmpty(t *testing.T) {
	if got := artifacts.SplitIntoCoherentUnits(""); len(got) != 0 {
		t.Errorf("SplitIntoCoherentUnits(\"\") = %+v, want empty", got)
	}
}

func TestSplitIntoCoherentUnits_MultipleParagraphsProduceMultipleUnits(t *testing.T) {
	content := "Paragraph one.\n\nParagraph two.\n\nParagraph three.\n"

	units := artifacts.SplitIntoCoherentUnits(content)
	if len(units) < 2 {
		t.Fatalf("SplitIntoCoherentUnits() = %+v, want more than one unit for multiple blank-line-separated paragraphs", units)
	}
}
