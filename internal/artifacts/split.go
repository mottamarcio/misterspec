package artifacts

import "strings"

// SplitIntoCoherentUnits splits content into paragraph-bounded units,
// reusing the same fence tracker ParseDocument/ExtractWikiLinks already
// use so a fenced code block's opening/closing pair, and everything
// between, always stays inside one unit — never split
// (035-context-budget-accuracy FR-011, data-model.md "CoherentUnit",
// research.md Decision 5). Concatenating every returned unit, in
// order, reproduces content exactly — a pure partition; only the
// caller's own choice of which prefix to keep drops anything.
// SplitIntoCoherentUnits("") returns nil.
func SplitIntoCoherentUnits(content string) []string {
	if content == "" {
		return nil
	}

	rawLines := strings.SplitAfter(content, "\n")
	stripped := make([]string, len(rawLines))
	for i, l := range rawLines {
		stripped[i] = strings.TrimSuffix(l, "\n")
	}
	fenced := fencedLines(stripped)

	var units []string
	var buf strings.Builder
	hasContent := false

	flush := func() {
		if buf.Len() > 0 {
			units = append(units, buf.String())
			buf.Reset()
			hasContent = false
		}
	}

	for i, raw := range rawLines {
		buf.WriteString(raw)
		if fenced[i] {
			hasContent = true
			continue
		}
		if strings.TrimSpace(stripped[i]) == "" {
			if hasContent {
				flush()
			}
			continue
		}
		hasContent = true
	}
	flush()

	return units
}
