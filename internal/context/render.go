package contextengine

import (
	"fmt"
	"strings"
)

// Render converts result into a well-formed, readable Markdown context
// pack (017-internal-context-command/spec.md FR-010): one "## <Tier
// label>" heading per Tier present among result.Items, in ascending
// Tier order — matching result's own already-established ordering;
// Render performs no re-sorting of its own — followed by one
// subsection per item naming its Path/Heading, its Content, and its own
// Reasons/Tokens. Contains every item present in result — Render never
// adds, drops, reorders, or re-scores anything (FR-011); it is a pure,
// read-only presentation of an already-computed Result. req is used
// only for a short header line naming the Target and Intent — it plays
// no role in item selection.
func Render(req Request, result Result) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# Context for %s", req.Target)
	if req.Intent != "" {
		fmt.Fprintf(&b, " (%s)", req.Intent)
	}
	b.WriteString("\n\n")

	currentTier := Tier(-1)
	for _, item := range result.Items {
		tier := minTier(item.Reasons)
		if tier != currentTier {
			fmt.Fprintf(&b, "## %s\n\n", tier.String())
			currentTier = tier
		}

		heading := item.Heading
		if heading == "" {
			fmt.Fprintf(&b, "### %s\n\n", item.Path)
		} else {
			fmt.Fprintf(&b, "### %s — %s\n\n", item.Path, heading)
		}
		b.WriteString(item.Content)
		b.WriteString("\n\n")

		reasons := make([]string, 0, len(item.Reasons))
		for _, r := range item.Reasons {
			reasons = append(reasons, r.Relation)
		}
		fmt.Fprintf(&b, "_Reasons: %s — Tokens: %d_\n\n", strings.Join(reasons, ", "), item.Tokens)
	}

	return b.String()
}
