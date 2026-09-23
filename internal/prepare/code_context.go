package prepare

import (
	"sort"
	"strings"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	contextengine "github.com/mottamarcio/misterspec/internal/context"
	"github.com/mottamarcio/misterspec/internal/context/index"
)

// maxInlineCodeBodySize is the fixed, documented threshold (in
// estimated tokens) under which a file's own declarations are
// rendered at full Body instead of Signature-only
// (044-architecture-code-context-rules research.md #7) — internal
// prepare has no budget mechanism to reuse (034 returns everything
// unconditionally), so this is a plain constant, not a per-request
// negotiable budget.
const maxInlineCodeBodySize = 800

// ResolveCodeContext parses scope into file paths, resolves each
// against store's own code index (DeclarationsForFiles — which also
// returns each file's own associated _test.go declarations, spec
// FR-007), includes every declaration at signature tier always, and
// promotes a file's own declarations to full-body tier when that
// file's own total estimated size is at or under
// maxInlineCodeBodySize. Any scope-named path resolving to no indexed
// declaration is reported in notFound rather than dropped silently
// (spec Edge Case, contracts §4).
func ResolveCodeContext(store index.Store, scope string, estimator artifacts.Estimator) (items []contextengine.PackageItem, notFound []string, err error) {
	paths := parseScopePaths(scope)
	if len(paths) == 0 {
		return nil, nil, nil
	}

	decls, err := store.DeclarationsForFiles(paths)
	if err != nil {
		return nil, nil, err
	}

	byPath := map[string][]index.CodeDeclaration{}
	seen := map[string]bool{}
	for _, d := range decls {
		byPath[d.Path] = append(byPath[d.Path], d)
		seen[d.Path] = true
	}

	for _, p := range paths {
		if !seen[p] {
			notFound = append(notFound, p)
		}
	}

	filePaths := make([]string, 0, len(byPath))
	for p := range byPath {
		filePaths = append(filePaths, p)
	}
	sort.Strings(filePaths)

	for _, path := range filePaths {
		group := byPath[path]
		total := 0
		for _, d := range group {
			total += estimator.Estimate(d.Body)
		}
		useFullBody := total <= maxInlineCodeBodySize

		sort.Slice(group, func(i, j int) bool { return group[i].StartLine < group[j].StartLine })
		for _, d := range group {
			content := d.Signature
			if useFullBody && d.Body != "" {
				content = d.Body
			}
			items = append(items, contextengine.PackageItem{
				Path:    d.Path,
				Heading: d.Name,
				Tier:    contextengine.TierMandatory,
				Reasons: []string{"scope"},
				Content: content,
				Location: contextengine.ItemLocation{
					Path:      d.Path,
					StartLine: d.StartLine,
					EndLine:   d.EndLine,
				},
				Fingerprint: contextengine.Fingerprint(content),
				Tokens:      estimator.Estimate(content),
			})
		}
	}

	return items, notFound, nil
}

// parseScopePaths splits scope on commas/whitespace, keeping only
// entries that look like a .go file — a Task's Scope: field may name
// non-code paths too (docs, etc.), which this feature has no reason to
// resolve.
func parseScopePaths(scope string) []string {
	fields := strings.FieldsFunc(scope, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\t' || r == ' '
	})

	var out []string
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if strings.HasSuffix(f, ".go") {
			out = append(out, f)
		}
	}
	return out
}
