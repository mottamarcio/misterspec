package contextengine

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/context/index"
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
	// HeadingPath/ScoreComponents/SourceReasons carry everything a
	// caller-requested rendering (heading_path/anchor,
	// score_components, provenance) needs, self-contained on the item
	// itself — never requiring a parallel, index-aligned ResultItem
	// slice to look them up (043-incremental-context-reuse code review
	// finding: a diff entry's own "added"/"modified" item has no such
	// alignment once items are appended/reordered by DiffAgainstBase,
	// which silently dropped heading_path/anchor/score_components/
	// provenance from a --base response relative to a direct call —
	// exactly the data FR-003 requires the two to agree on).
	HeadingPath     []string
	ScoreComponents ScoreComponents
	SourceReasons   []Reason
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
			Fingerprint:     Fingerprint(item.Content),
			HeadingPath:     item.HeadingPath,
			ScoreComponents: item.Components,
			SourceReasons:   item.Reasons,
		})
	}
	return out
}

// ConfigHash / PackID are both "sha256:<hex>" strings, computed with
// the same hashing convention Fingerprint already establishes — never
// a second hashing scheme (043-incremental-context-reuse data-model.md
// "ConfigHash / PackID", research.md #2/#3).
type ConfigHash string
type PackID string

// ConfigIdentity is every PackIdentity input except the selected
// items — hashed on its own (ComputeConfigHash) so a live call's own
// freshly computed hash can be compared directly against a stored
// row's config_hash, without needing to invert a hash
// (043-incremental-context-reuse data-model.md, research.md #3's
// correction). Budget/HardLimit are the *resolved* values (after
// DefaultBudget/DefaultHardLimit substitution), never the raw nilable
// Request fields, so two calls that resolve to the same effective
// budget hash identically regardless of whether one passed --budget
// explicitly.
type ConfigIdentity struct {
	Target               string
	Intent               string
	Task                 string
	Query                string
	QueryMode            string
	Budget               int
	HardLimit            int
	PreferSection        bool
	RankingVersion       int
	ContextSchemaVersion int
	Estimator            string
}

// PackItemIdentity is one selected item's own identity+fingerprint
// pair — reuses Candidate's existing (Path, StartLine, EndLine)
// identity convention (043-incremental-context-reuse data-model.md,
// research.md #5).
type PackItemIdentity struct {
	Path        string
	StartLine   int
	EndLine     int
	Fingerprint string
}

// PackIdentity is ConfigIdentity plus the ordered selected-item list
// — the full set of inputs PackID is derived from
// (043-incremental-context-reuse data-model.md).
type PackIdentity struct {
	Config ConfigIdentity
	Items  []PackItemIdentity
}

// ComputeConfigHash derives cfg's own deterministic ConfigHash, over a
// canonical, field-order-fixed serialization — two ConfigIdentity
// values with identical fields always hash identically; any single
// differing field changes the result (043-incremental-context-reuse
// data-model.md "ConfigHash / PackID").
func ComputeConfigHash(cfg ConfigIdentity) ConfigHash {
	var b strings.Builder
	fmt.Fprintf(&b, "target=%s\x1f", cfg.Target)
	fmt.Fprintf(&b, "intent=%s\x1f", cfg.Intent)
	fmt.Fprintf(&b, "task=%s\x1f", cfg.Task)
	fmt.Fprintf(&b, "query=%s\x1f", cfg.Query)
	fmt.Fprintf(&b, "query_mode=%s\x1f", cfg.QueryMode)
	fmt.Fprintf(&b, "budget=%d\x1f", cfg.Budget)
	fmt.Fprintf(&b, "hard_limit=%d\x1f", cfg.HardLimit)
	fmt.Fprintf(&b, "prefer_section=%t\x1f", cfg.PreferSection)
	fmt.Fprintf(&b, "ranking_version=%d\x1f", cfg.RankingVersion)
	fmt.Fprintf(&b, "context_schema_version=%d\x1f", cfg.ContextSchemaVersion)
	fmt.Fprintf(&b, "estimator=%s\x1f", cfg.Estimator)
	return ConfigHash(Fingerprint(b.String()))
}

// ComputePackID derives id's own deterministic PackID: the hash of
// ComputeConfigHash(id.Config) concatenated with every item's own
// identity+fingerprint, in id.Items's own order — so a difference in
// either the config or any single item (including just its own
// Fingerprint, or the item count itself) changes the result
// (043-incremental-context-reuse data-model.md, research.md #2/#3).
func ComputePackID(id PackIdentity) PackID {
	var b strings.Builder
	fmt.Fprintf(&b, "config=%s\x1f", ComputeConfigHash(id.Config))
	for _, item := range id.Items {
		fmt.Fprintf(&b, "item=%s|%d|%d|%s\x1f", item.Path, item.StartLine, item.EndLine, item.Fingerprint)
	}
	return PackID(Fingerprint(b.String()))
}

// itemIdentity is the (Path, StartLine, EndLine) triple DiffAgainstBase
// keys items by — the same identity Candidate already uses for
// deduplication (043-incremental-context-reuse research.md #5).
type itemIdentity struct {
	Path      string
	StartLine int
	EndLine   int
}

// DiffEntry is one position of a PackDiff's own final, current-order
// item list (043-incremental-context-reuse data-model.md "DiffEntry").
type DiffEntry struct {
	// Kind is "reuse", "added", or "modified".
	Kind string
	// BaseIndex is non-nil only for Kind == "reuse" — the 0-based
	// index into the named base pack's own stored item list to copy
	// content from.
	BaseIndex *int
	// Item is non-nil only for Kind == "added"/"modified" — the
	// item's own full current data.
	Item *PackageItem
}

// RemovedEntry is one base identity absent from the current selection
// (043-incremental-context-reuse data-model.md "RemovedEntry").
type RemovedEntry struct {
	Path      string
	StartLine int
	EndLine   int
}

// PackDiff is DiffAgainstBase's own result (043-incremental-context-
// reuse data-model.md "PackDiff").
type PackDiff struct {
	Entries []DiffEntry
	Removed []RemovedEntry
}

// DiffAgainstBase compares current (the freshly computed package for
// this call) against base (a StoredPack's own deserialized items),
// returning current's own final order as a sequence of reuse-or-
// replace instructions, plus every base identity no longer present
// (043-incremental-context-reuse data-model.md "PackDiff", research.md
// #5). Pure — no I/O. Applying the result against base (substituting
// each "reuse" entry with base[*BaseIndex], keeping every "added"/
// "modified" entry's own Item, in Entries' own order) reproduces
// current exactly (spec FR-003).
func DiffAgainstBase(current []PackageItem, base []index.StoredPackItem) PackDiff {
	baseIndex := make(map[itemIdentity]int, len(base))
	for i, b := range base {
		baseIndex[itemIdentity{Path: b.Path, StartLine: b.StartLine, EndLine: b.EndLine}] = i
	}

	matched := make(map[int]bool, len(base))
	diff := PackDiff{Entries: make([]DiffEntry, 0, len(current))}

	for _, item := range current {
		id := itemIdentity{Path: item.Location.Path, StartLine: item.Location.StartLine, EndLine: item.Location.EndLine}
		if bi, ok := baseIndex[id]; ok {
			matched[bi] = true
			if base[bi].Fingerprint == item.Fingerprint {
				idx := bi
				diff.Entries = append(diff.Entries, DiffEntry{Kind: "reuse", BaseIndex: &idx})
				continue
			}
			itemCopy := item
			diff.Entries = append(diff.Entries, DiffEntry{Kind: "modified", Item: &itemCopy})
			continue
		}
		itemCopy := item
		diff.Entries = append(diff.Entries, DiffEntry{Kind: "added", Item: &itemCopy})
	}

	for i, b := range base {
		if !matched[i] {
			diff.Removed = append(diff.Removed, RemovedEntry{Path: b.Path, StartLine: b.StartLine, EndLine: b.EndLine})
		}
	}

	return diff
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
