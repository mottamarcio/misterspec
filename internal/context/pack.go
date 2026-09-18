package contextengine

import (
	"crypto/sha256"
	"fmt"

	"github.com/mottamarcio/misterspec/internal/artifacts"
)

// PackageItem is one "package"-mode item: every field a "manifest"-mode
// item already carries, plus the full selected content, its
// file-absolute location, and a content fingerprint
// (033-context-pack-output-contract data-model.md "PackageItem").
type PackageItem struct {
	Path        string
	Heading     string
	Tier        Tier
	Reasons     []string
	Score       int
	Tokens      int
	Content     string
	Location    ItemLocation
	Fingerprint string
}

// ItemLocation is one item's file-absolute source position
// (data-model.md "ItemLocation"). StartLine/EndLine are already
// file-absolute on ResultItem/Candidate — see collector.go's
// chunkArtifact (research.md Decision 1) — this type simply names the
// (Path, StartLine, EndLine) triple as its own value for the package
// item's own "location" field.
type ItemLocation struct {
	Path      string
	StartLine int
	EndLine   int
}

// Fingerprint computes content's own SHA-256 digest, rendered
// "sha256:<hex>" — matching operations.FileFingerprint's existing
// string format for consistency, without importing internal/operations
// for it (research.md Decision 4). A pure function of content alone:
// identical content always produces the identical fingerprint; any
// difference, however small, changes it.
func Fingerprint(content string) string {
	sum := sha256.Sum256([]byte(content))
	return fmt.Sprintf("sha256:%x", sum)
}

// BuildPackageItems converts items into their "package"-mode shape
// (data-model.md "PackageItem") — reusing each ResultItem's own
// already-in-memory Content, never re-reading the source file.
func BuildPackageItems(items []ResultItem) []PackageItem {
	out := make([]PackageItem, 0, len(items))
	for _, item := range items {
		reasons := make([]string, 0, len(item.Reasons))
		tier := item.Reasons[0].Tier
		for _, r := range item.Reasons {
			reasons = append(reasons, r.Relation)
			if r.Tier < tier {
				tier = r.Tier
			}
		}
		out = append(out, PackageItem{
			Path:    item.Path,
			Heading: item.Heading,
			Tier:    tier,
			Reasons: reasons,
			Score:   item.Score,
			Tokens:  item.Tokens,
			Content: item.Content,
			Location: ItemLocation{
				Path:      item.Path,
				StartLine: item.StartLine,
				EndLine:   item.EndLine,
			},
			Fingerprint: Fingerprint(item.Content),
		})
	}
	return out
}

// PayloadTokens estimates the actual serialized size, in tokens, of the
// response body for the requested mode (033-context-pack-output-contract
// research.md Decision 6) — always >= the content-only
// Diagnostics.TokensSelected, since it additionally accounts for
// per-item metadata (path, heading, reasons, location, fingerprint) or
// the rendered Markdown's own formatting overhead. contentTokens is the
// caller's already-computed Diagnostics.TokensSelected, reused rather
// than recomputed.
func PayloadTokens(contentTokens int, packageItems []PackageItem, rendered string, estimator artifacts.Estimator) int {
	if rendered != "" {
		return estimator.Estimate(rendered)
	}
	overhead := 0
	for _, item := range packageItems {
		// A rough per-item accounting of the metadata fields a package
		// item carries beyond its own content: path, heading, reasons,
		// location, fingerprint — estimated the same way content
		// itself is, so the unit stays consistent.
		overhead += estimator.Estimate(fmt.Sprintf("%s%s%v%d%d%s", item.Path, item.Heading, item.Reasons, item.Location.StartLine, item.Location.EndLine, item.Fingerprint))
	}
	return contentTokens + overhead
}
