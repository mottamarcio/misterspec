package artifacts

import (
	"fmt"
	"os"
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
