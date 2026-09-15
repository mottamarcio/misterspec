package index

import "fmt"

// Search runs an FTS5 full-text query over indexed chunk content
// (FR-003), ranked by SQLite's own bm25() — lower is more relevant,
// per FTS5's own convention — joined back to chunks/documents for full
// provenance (data-model.md).
func (s *sqliteStore) Search(query string, limit int) ([]SearchResult, error) {
	rows, err := s.db.Query(`
		SELECT d.path, c.heading, c.content, c.start_line, c.end_line, bm25(chunks_fts) AS rank
		FROM chunks_fts
		JOIN chunks c ON c.id = chunks_fts.chunk_id
		JOIN documents d ON d.id = c.document_id
		WHERE chunks_fts MATCH ?
		ORDER BY rank
		LIMIT ?`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("index: searching for %q: %w", query, err)
	}
	defer rows.Close()

	results := make([]SearchResult, 0)
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.Path, &r.Heading, &r.Content, &r.StartLine, &r.EndLine, &r.Rank); err != nil {
			return nil, fmt.Errorf("index: reading search result: %w", err)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("index: reading search results: %w", err)
	}
	return results, nil
}
