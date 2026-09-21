package index

import (
	"database/sql"
	"fmt"
	"strings"
)

// searchSQL is shared by Search and SearchAdvanced — only how the
// MATCH argument is built differs between them (036-text-search-
// ranking research.md #1, #2).
const searchSQL = `
	SELECT d.path, c.heading, c.content, c.start_line, c.end_line, bm25(chunks_fts) AS rank
	FROM chunks_fts
	JOIN chunks c ON c.id = chunks_fts.chunk_id
	JOIN documents d ON d.id = c.document_id
	WHERE chunks_fts MATCH ?
	ORDER BY rank
	LIMIT ?`

// Search runs a free-text query over indexed chunk content (FR-003),
// ranked by SQLite's own bm25() — lower is more relevant, per FTS5's
// own convention — joined back to chunks/documents for full provenance
// (data-model.md). query is tokenized and every token is treated as
// literal text (036-text-search-ranking spec FR-001, FR-003): no
// character or bareword in query is ever interpreted as FTS5 query
// syntax, so a syntax error here is impossible by construction
// (research.md #1). Use SearchAdvanced when FTS5's own query grammar
// (phrase/prefix/boolean/NEAR) is wanted deliberately.
func (s *sqliteStore) Search(query string, limit int) ([]SearchResult, error) {
	expr := freeTextMatchExpr(query)
	if expr == "" {
		// No searchable token in query (empty, or punctuation/short
		// fragments only) — an empty result, never an error
		// (036-text-search-ranking spec Edge Cases).
		return []SearchResult{}, nil
	}
	rows, err := s.db.Query(searchSQL, expr, limit)
	if err != nil {
		// Impossible in practice (expr is always a well-formed, quoted
		// FTS5 literal expression — research.md #1), but if the driver
		// ever errors here it is a real index/query failure, never a
		// syntax problem attributable to the caller's own query text.
		return nil, fmt.Errorf("index: searching for %q: %w", query, err)
	}
	defer rows.Close()
	results, err := scanSearchResults(rows)
	if err != nil {
		return nil, fmt.Errorf("index: reading results for %q: %w", query, err)
	}
	return results, nil
}

// SearchAdvanced runs query against FTS5's own MATCH grammar unmodified
// (036-text-search-ranking spec FR-002) — phrase, prefix, boolean, and
// NEAR syntax are honored, only when a caller opts into this method
// explicitly (research.md #2: query mode is never inferred from the
// query text). FTS5 parses/validates a MATCH expression while
// preparing the query, so a malformed expression surfaces as an error
// from db.Query itself — that error, and only that one, is wrapped as
// ErrQuerySyntax (spec FR-004); a failure while reading already-
// executing rows (Scan/rows.Err) is a distinct index/read failure, not
// a syntax problem, and is never mislabeled as one (contracts §2: "MUST
// [be] treat[ed] as distinct from internal_error/index-failure codes").
func (s *sqliteStore) SearchAdvanced(query string, limit int) ([]SearchResult, error) {
	rows, err := s.db.Query(searchSQL, query, limit)
	if err != nil {
		return nil, fmt.Errorf("%w: %q: %v", ErrQuerySyntax, query, err)
	}
	defer rows.Close()
	results, err := scanSearchResults(rows)
	if err != nil {
		return nil, fmt.Errorf("index: reading results for %q: %w", query, err)
	}
	return results, nil
}

// scanSearchResults reads every row of an already-successful search
// query. Errors here (Scan/rows.Err) are read/decode failures, never a
// query-syntax problem — the query already parsed and started
// executing by the time these can occur.
func scanSearchResults(rows *sql.Rows) ([]SearchResult, error) {
	results := make([]SearchResult, 0)
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.Path, &r.Heading, &r.Content, &r.StartLine, &r.EndLine, &r.Rank); err != nil {
			return nil, fmt.Errorf("reading search result: %w", err)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading search results: %w", err)
	}
	return results, nil
}

// freeTextMatchExpr builds a safe FTS5 MATCH expression from query,
// treating every whitespace-separated token as a literal string
// (036-text-search-ranking research.md #1): FTS5 gives a double-quoted
// string no operator interpretation, so no reserved character or
// bareword keyword (AND/OR/NOT/NEAR) can ever produce a syntax error.
// Tokens are joined with a space, FTS5's own implicit AND. Returns ""
// when query has no token to search for.
func freeTextMatchExpr(query string) string {
	fields := strings.Fields(query)
	if len(fields) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(fields))
	for _, f := range fields {
		quoted = append(quoted, quoteFTS5Literal(f))
	}
	return strings.Join(quoted, " ")
}

// quoteFTS5Literal wraps token as one FTS5 double-quoted string
// literal, doubling any internal double quote per SQLite's own
// string-literal escaping convention.
func quoteFTS5Literal(token string) string {
	return `"` + strings.ReplaceAll(token, `"`, `""`) + `"`
}
