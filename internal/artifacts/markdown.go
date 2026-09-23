package artifacts

import (
	"fmt"
	"os"
	"strings"
)

// ReadBody reads path and returns its Markdown body — everything after
// the artifact's frontmatter block's closing delimiter. It is
// ParseMetadata's counterpart: ParseMetadata returns only the
// frontmatter half of a file; ReadBody returns only the body half,
// reusing the same splitFrontmatter delimiter scan rather than a second
// one (specs/011-wikilink-foundation/research.md). An artifact with
// nothing after its closing delimiter returns an empty, non-nil-error
// body — not a failure.
//
// ReadBody returns the same ErrArtifactNotFound/ErrFrontmatterMalformed
// vocabulary ParseMetadata already uses for a missing file or a missing/
// malformed frontmatter block, rather than a second, parallel one.
func ReadBody(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrArtifactNotFound, path)
		}
		return nil, fmt.Errorf("artifacts: reading %s: %w", path, err)
	}

	_, body, err := splitFrontmatter(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrFrontmatterMalformed, path, err)
	}

	return body, nil
}

// ReadBodyWithOffset reads path exactly as ReadBody does, and
// additionally reports bodyStartLine — the 1-indexed line number, in
// the whole file, at which the returned body begins. Any caller that
// must report a location *within* the body as absolute to the whole
// file (033-context-pack-output-contract research.md Decision 1) adds
// this offset to a body-relative line number. Shares ReadBody's own
// ErrArtifactNotFound/ErrFrontmatterMalformed vocabulary — every
// canonical misterspec artifact requires frontmatter, so a file without
// it is still an error here, never a zero-offset fallback.
func ReadBodyWithOffset(path string) (body []byte, bodyStartLine int, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, fmt.Errorf("%w: %s", ErrArtifactNotFound, path)
		}
		return nil, 0, fmt.Errorf("artifacts: reading %s: %w", path, err)
	}

	frontmatter, body, err := splitFrontmatter(data)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %s: %v", ErrFrontmatterMalformed, path, err)
	}

	// The body starts after: the opening "---" line, every frontmatter
	// line, and the closing "---" line — three fixed lines plus
	// whatever frontmatter itself contains. strings.Count(.... "\n")+1
	// undercounts to 0 for an empty frontmatter (no content between an
	// immediately-closed delimiter pair), which is exactly correct.
	frontmatterLines := 0
	if len(frontmatter) > 0 {
		frontmatterLines = strings.Count(string(frontmatter), "\n") + 1
	}
	bodyStartLine = frontmatterLines + 3

	return body, bodyStartLine, nil
}
