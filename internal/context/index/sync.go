package index

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// clearStatements empties every derived table — used by Rebuild before
// reindexing everything from scratch (FR-006).
var clearStatements = []string{
	`DELETE FROM chunks_fts`,
	`DELETE FROM chunks`,
	`DELETE FROM links`,
	`DELETE FROM documents`,
}

// Sync performs this package's own indexing walk inside one transaction
// (research.md #9) — see indexAllFound for the per-artifact logic this
// task (T011) and later tasks (T016, T018) build up.
func (s *sqliteStore) Sync(root string, cfg project.Configuration) (SyncReport, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return SyncReport{}, fmt.Errorf("index: beginning sync transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op once Commit has succeeded

	report, err := s.indexAllFound(tx, root, cfg)
	if err != nil {
		return SyncReport{}, err
	}

	if err := tx.Commit(); err != nil {
		return SyncReport{}, fmt.Errorf("index: committing sync: %w", err)
	}
	return report, nil
}

// Rebuild clears every derived row, then runs the exact same per-
// artifact indexing walk Sync does — all inside one transaction, so a
// failure partway through never leaves the database emptied without
// also being fully reindexed (FR-006, FR-007, research.md #9).
func (s *sqliteStore) Rebuild(root string, cfg project.Configuration) (SyncReport, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return SyncReport{}, fmt.Errorf("index: beginning rebuild transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	for _, stmt := range clearStatements {
		if _, err := tx.Exec(stmt); err != nil {
			return SyncReport{}, fmt.Errorf("index: clearing prior state: %w", err)
		}
	}

	report, err := s.indexAllFound(tx, root, cfg)
	if err != nil {
		return SyncReport{}, err
	}

	if err := tx.Commit(); err != nil {
		return SyncReport{}, fmt.Errorf("index: committing rebuild: %w", err)
	}
	return report, nil
}

// indexAllFound walks root's eligible artifacts and reconciles them
// against whatever is already in tx: unchanged (matching fingerprint,
// research.md #5/FR-005) artifacts are skipped entirely; changed ones
// have their stale derived rows replaced; new ones are indexed for the
// first time; and whatever remains in existing afterward — no longer
// found on disk — is removed (FR-004). Rebuild's own prior DELETEs
// already guarantee every artifact looks "new" to this same logic, so
// no separate rebuild-specific path is needed.
func (s *sqliteStore) indexAllFound(tx *sql.Tx, root string, cfg project.Configuration) (SyncReport, error) {
	found, err := walkEligibleArtifacts(root, cfg)
	if err != nil {
		return SyncReport{}, fmt.Errorf("index: walking %s: %w", cfg.ArtifactsDir, err)
	}

	existing, err := loadExistingDocuments(tx)
	if err != nil {
		return SyncReport{}, err
	}

	var report SyncReport
	for relPath, t := range found {
		fp, err := operations.Fingerprint(root, filepath.Join(root, relPath))
		if err != nil {
			report.Errors = append(report.Errors, SyncError{Path: relPath, Err: err})
			continue
		}
		fpStr := fp.String()

		prior, wasIndexed := existing[relPath]
		delete(existing, relPath) // whatever remains after this loop no longer exists on disk

		if wasIndexed && prior.fingerprint == fpStr {
			report.Skipped++
			continue
		}
		if wasIndexed {
			if err := deleteDocument(tx, prior.id, prior.artifactID); err != nil {
				return SyncReport{}, err
			}
		}
		if err := indexOneArtifact(tx, root, cfg, relPath, t, fpStr); err != nil {
			report.Errors = append(report.Errors, SyncError{Path: relPath, Err: err})
			continue
		}
		if wasIndexed {
			report.Updated++
		} else {
			report.Indexed++
		}
	}

	for _, prior := range existing {
		if err := deleteDocument(tx, prior.id, prior.artifactID); err != nil {
			return SyncReport{}, err
		}
		report.Removed++
	}

	return report, nil
}

// existingDocument is loadExistingDocuments' own per-path summary of
// what's already indexed — just enough to drive change detection and
// clean up a stale artifact's own links rows.
type existingDocument struct {
	id          int64
	fingerprint string
	artifactID  sql.NullString
}

// loadExistingDocuments reads every already-indexed path's id,
// fingerprint, and artifact_id from tx, keyed by path.
func loadExistingDocuments(tx *sql.Tx) (map[string]existingDocument, error) {
	rows, err := tx.Query(`SELECT id, path, fingerprint, artifact_id FROM documents`)
	if err != nil {
		return nil, fmt.Errorf("index: reading existing documents: %w", err)
	}
	defer rows.Close()

	out := map[string]existingDocument{}
	for rows.Next() {
		var d existingDocument
		var path string
		if err := rows.Scan(&d.id, &path, &d.fingerprint, &d.artifactID); err != nil {
			return nil, fmt.Errorf("index: reading existing document row: %w", err)
		}
		out[path] = d
	}
	return out, rows.Err()
}

// deleteDocument removes documentID's own chunks_fts, chunks, links
// (where it was the source — research.md #4), and documents rows —
// used both when an artifact has changed (delete before reindexing) and
// when it has been removed from disk entirely (delete with nothing to
// replace it).
func deleteDocument(tx *sql.Tx, documentID int64, artifactID sql.NullString) error {
	if _, err := tx.Exec(`DELETE FROM chunks_fts WHERE chunk_id IN (SELECT id FROM chunks WHERE document_id = ?)`, documentID); err != nil {
		return fmt.Errorf("index: deleting stale chunks_fts rows: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM chunks WHERE document_id = ?`, documentID); err != nil {
		return fmt.Errorf("index: deleting stale chunks rows: %w", err)
	}
	if artifactID.Valid {
		if _, err := tx.Exec(`DELETE FROM links WHERE source_artifact_id = ?`, artifactID.String); err != nil {
			return fmt.Errorf("index: deleting stale links rows: %w", err)
		}
	}
	if _, err := tx.Exec(`DELETE FROM documents WHERE id = ?`, documentID); err != nil {
		return fmt.Errorf("index: deleting stale documents row: %w", err)
	}
	return nil
}

// walkEligibleArtifacts finds every ".md" file under root/cfg.ArtifactsDir
// that artifacts.ClassifyPath recognizes as a real artifact — reusing
// ClassifyPath's own existing 9-type vocabulary rather than a new
// "eligible types" list (research.md #3). A file ClassifyPath cannot
// classify is silently skipped, not an error. A missing
// cfg.ArtifactsDir yields an empty result, not an error.
func walkEligibleArtifacts(root string, cfg project.Configuration) (map[string]artifacts.ArtifactType, error) {
	base := filepath.Join(root, cfg.ArtifactsDir)
	found := map[string]artifacts.ArtifactType{}

	if _, err := os.Stat(base); err != nil {
		if os.IsNotExist(err) {
			return found, nil
		}
		return nil, err
	}

	walkErr := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		t, classifyErr := artifacts.ClassifyPath(root, cfg, path)
		if classifyErr != nil {
			return nil // not a recognized artifact — skip silently
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		found[filepath.ToSlash(rel)] = t
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return found, nil
}

// indexOneArtifact reads, structures, and chunks the artifact at
// root/relPath (reusing 011's ReadBody/013's ParseDocument+Chunks and
// ParseMetadata), inserting its documents/chunks/chunks_fts rows into
// tx, tagged with its already-computed fingerprint. For one of the five
// ID-bearing types, also populates its own links rows directly from
// operations.References — consistent with 012's own direct computation
// by construction (FR-009, research.md #4). A Plan/Tasks/Validation/
// Constitution artifact contributes no links rows at all.
func indexOneArtifact(tx *sql.Tx, root string, cfg project.Configuration, relPath string, t artifacts.ArtifactType, fingerprint string) error {
	full := filepath.Join(root, relPath)

	meta, err := artifacts.ParseMetadata(full)
	if err != nil {
		return err
	}
	body, bodyStartLine, err := artifacts.ReadBodyWithOffset(full)
	if err != nil {
		return err
	}
	offset := bodyStartLine - 1
	// File-absolute, not body-relative — the same fix
	// internal/context/collector.go's chunkArtifact applies, needed
	// here too so a chunk found both directly and via this index's own
	// text search reports the identical (Path, StartLine, EndLine)
	// identity and deduplicates correctly (033-context-pack-output-
	// contract research.md Decision 1).
	chunks := artifacts.ChunksWithOffset(relPath, body, offset)
	doc := artifacts.ParseDocument(body)

	var artifactID sql.NullString
	if meta.ID != nil {
		artifactID = sql.NullString{String: meta.ID.String(), Valid: true}
	}

	res, err := tx.Exec(
		`INSERT INTO documents (path, artifact_id, artifact_type, title, fingerprint, indexed_at) VALUES (?, ?, ?, ?, ?, ?)`,
		relPath, artifactID, t.String(), titleFor(doc), fingerprint, time.Now().Unix(),
	)
	if err != nil {
		return fmt.Errorf("index: inserting document row for %s: %w", relPath, err)
	}
	documentID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	for _, c := range chunks {
		if err := insertChunk(tx, documentID, c); err != nil {
			return err
		}
	}

	if t.HasEntityID() && meta.ID != nil {
		if err := indexLinks(tx, root, cfg, meta.ID.String()); err != nil {
			return err
		}
	}
	return nil
}

// indexLinks populates artifactID's own outgoing links rows directly
// from operations.References — both formal and semantic — the same
// function 012's own References command uses, guaranteeing consistency
// (FR-009) by construction rather than by a second, parallel
// computation (research.md #4).
func indexLinks(tx *sql.Tx, root string, cfg project.Configuration, artifactID string) error {
	refs, err := operations.References(root, cfg, artifactID)
	if err != nil {
		return fmt.Errorf("index: computing references for %s: %w", artifactID, err)
	}
	for _, r := range refs.Formal {
		if err := insertLink(tx, r.Relation, artifactID, r.Target.String(), r.SourceSection, r.SourceLine, r.TargetAnchor); err != nil {
			return err
		}
	}
	for _, r := range refs.Semantic {
		if err := insertLink(tx, r.Relation, artifactID, r.Target.String(), r.SourceSection, r.SourceLine, r.TargetAnchor); err != nil {
			return err
		}
	}
	return nil
}

// insertLink inserts one links row. sourceSection/sourceLine/
// targetAnchor are stored NULL for a formal relation, or for a
// semantic one with no anchor (038-wikilink-chunk-provenance data-
// model.md; 040-stable-section-anchors data-model.md "Index Schema
// (extended)") — an empty sourceSection/zero sourceLine/empty
// targetAnchor (as ReferenceEntry/BacklinkEntry already report) becomes
// SQL NULL via sql.NullString/sql.NullInt64's own IsZero-style Valid
// flag.
func insertLink(tx *sql.Tx, relation, source, target, sourceSection string, sourceLine int, targetAnchor string) error {
	section := sql.NullString{String: sourceSection, Valid: sourceSection != ""}
	line := sql.NullInt64{Int64: int64(sourceLine), Valid: sourceLine != 0}
	anchor := sql.NullString{String: targetAnchor, Valid: targetAnchor != ""}
	if _, err := tx.Exec(
		`INSERT INTO links (source_artifact_id, target_artifact_id, relation, source_section, source_line, target_anchor) VALUES (?, ?, ?, ?, ?, ?)`,
		source, target, relation, section, line, anchor,
	); err != nil {
		return fmt.Errorf("index: inserting link row: %w", err)
	}
	return nil
}

// insertChunk inserts one Chunk's own chunks row plus its matching
// chunks_fts row (research.md #8 — explicit Go-driven synchronization,
// no SQL triggers).
func insertChunk(tx *sql.Tx, documentID int64, c artifacts.Chunk) error {
	anchor := sql.NullString{String: c.Anchor, Valid: c.Anchor != ""}
	res, err := tx.Exec(
		`INSERT INTO chunks (document_id, heading, anchor, content, start_line, end_line, token_estimate) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		documentID, c.Heading, anchor, c.Content, c.StartLine, c.EndLine, artifacts.EstimateTokens(c.Content),
	)
	if err != nil {
		return fmt.Errorf("index: inserting chunk row: %w", err)
	}
	chunkID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(
		`INSERT INTO chunks_fts (heading, content, chunk_id) VALUES (?, ?, ?)`,
		c.Heading, c.Content, chunkID,
	); err != nil {
		return fmt.Errorf("index: inserting chunks_fts row: %w", err)
	}
	return nil
}

// titleFor derives documents.title from doc's first Section's own
// Heading — empty if doc has no Sections at all (an empty body).
func titleFor(doc artifacts.Document) string {
	if len(doc.Sections) == 0 {
		return ""
	}
	return doc.Sections[0].Heading
}
