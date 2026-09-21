package index

import "errors"

// ErrQuerySyntax is returned, wrapped with the offending query and the
// underlying driver error, when SearchAdvanced's own caller-supplied
// FTS5 expression is malformed (036-text-search-ranking spec FR-004).
// Search (free-text mode) never returns this — every token it hands
// FTS5 is a quoted literal, never raw query syntax (research.md #1).
var ErrQuerySyntax = errors.New("index: query syntax error")
