package index

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/mottamarcio/misterspec/internal/gosource"
	"github.com/mottamarcio/misterspec/internal/operations"
)

// SyncCode incrementally reconciles code_files/code_declarations with
// root's current *.go files, mirroring Sync's own new/changed/deleted
// reconciliation shape for Markdown (044-architecture-code-context-
// rules research.md #6). One transaction per call.
func (s *sqliteStore) SyncCode(root string, exclusions []string) (SyncReport, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return SyncReport{}, fmt.Errorf("index: beginning code sync transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op once Commit has succeeded

	report, err := indexAllCodeFound(tx, root, exclusions)
	if err != nil {
		return SyncReport{}, err
	}

	if err := tx.Commit(); err != nil {
		return SyncReport{}, fmt.Errorf("index: committing code sync: %w", err)
	}
	return report, nil
}

// DeclarationsForFiles returns every CodeDeclaration whose own Path is
// in paths, plus every declaration from that path's own associated
// _test.go file(s) in the same directory (spec FR-007).
func (s *sqliteStore) DeclarationsForFiles(paths []string) ([]CodeDeclaration, error) {
	targets := map[string]bool{}
	for _, p := range paths {
		targets[p] = true
		targets[associatedTestPath(p)] = true
	}
	if len(targets) == 0 {
		return nil, nil
	}

	placeholders := make([]string, 0, len(targets))
	args := make([]any, 0, len(targets))
	for p := range targets {
		placeholders = append(placeholders, "?")
		args = append(args, p)
	}

	query := fmt.Sprintf(`
		SELECT cf.path, cd.name, cd.kind, cd.signature, cd.body, cd.start_line, cd.end_line, cd.is_test
		FROM code_declarations cd
		JOIN code_files cf ON cf.id = cd.code_file_id
		WHERE cf.path IN (%s)
	`, strings.Join(placeholders, ", "))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("index: reading code declarations: %w", err)
	}
	defer rows.Close()

	var out []CodeDeclaration
	for rows.Next() {
		var d CodeDeclaration
		var body sql.NullString
		var isTest int
		if err := rows.Scan(&d.Path, &d.Name, &d.Kind, &d.Signature, &body, &d.StartLine, &d.EndLine, &isTest); err != nil {
			return nil, fmt.Errorf("index: reading code declaration row: %w", err)
		}
		d.Body = body.String
		d.IsTest = isTest != 0
		out = append(out, d)
	}
	return out, rows.Err()
}

// associatedTestPath derives foo_test.go from foo.go in the same
// directory — the same file-naming convention Go's own toolchain
// already uses (data-model.md "CodeDeclaration"). For a path already
// ending in _test.go, returns itself (harmless duplicate in the
// targets set, deduplicated by the map).
func associatedTestPath(path string) string {
	if strings.HasSuffix(path, "_test.go") {
		return path
	}
	ext := filepath.Ext(path)
	return strings.TrimSuffix(path, ext) + "_test" + ext
}

func indexAllCodeFound(tx *sql.Tx, root string, exclusions []string) (SyncReport, error) {
	found, err := gosource.WalkGoFiles(root, exclusions)
	if err != nil {
		return SyncReport{}, fmt.Errorf("index: walking %s for *.go files: %w", root, err)
	}

	existing, err := loadExistingCodeFiles(tx)
	if err != nil {
		return SyncReport{}, err
	}

	var report SyncReport
	for _, relPath := range found {
		fp, err := operations.Fingerprint(root, filepath.Join(root, relPath))
		if err != nil {
			report.Errors = append(report.Errors, SyncError{Path: relPath, Err: err})
			continue
		}
		fpStr := fp.String()

		prior, wasIndexed := existing[relPath]
		delete(existing, relPath)

		if wasIndexed && prior.fingerprint == fpStr {
			report.Skipped++
			continue
		}
		if wasIndexed {
			if err := deleteCodeFile(tx, prior.id); err != nil {
				return SyncReport{}, err
			}
		}
		if err := indexOneCodeFile(tx, root, relPath, fpStr); err != nil {
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
		if err := deleteCodeFile(tx, prior.id); err != nil {
			return SyncReport{}, err
		}
		report.Removed++
	}

	return report, nil
}

type existingCodeFile struct {
	id          int64
	fingerprint string
}

func loadExistingCodeFiles(tx *sql.Tx) (map[string]existingCodeFile, error) {
	rows, err := tx.Query(`SELECT id, path, fingerprint FROM code_files`)
	if err != nil {
		return nil, fmt.Errorf("index: reading existing code_files: %w", err)
	}
	defer rows.Close()

	out := map[string]existingCodeFile{}
	for rows.Next() {
		var f existingCodeFile
		var path string
		if err := rows.Scan(&f.id, &path, &f.fingerprint); err != nil {
			return nil, fmt.Errorf("index: reading existing code_files row: %w", err)
		}
		out[path] = f
	}
	return out, rows.Err()
}

func deleteCodeFile(tx *sql.Tx, codeFileID int64) error {
	if _, err := tx.Exec(`DELETE FROM code_declarations WHERE code_file_id = ?`, codeFileID); err != nil {
		return fmt.Errorf("index: deleting stale code_declarations rows: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM code_files WHERE id = ?`, codeFileID); err != nil {
		return fmt.Errorf("index: deleting stale code_files row: %w", err)
	}
	return nil
}

func indexOneCodeFile(tx *sql.Tx, root, relPath, fingerprint string) error {
	full := filepath.Join(root, relPath)
	isTestFile := strings.HasSuffix(relPath, "_test.go")

	decls, err := gosource.Declarations(full)
	if err != nil {
		return err
	}

	res, err := tx.Exec(
		`INSERT INTO code_files (path, fingerprint, indexed_at) VALUES (?, ?, ?)`,
		relPath, fingerprint, time.Now().Unix(),
	)
	if err != nil {
		return fmt.Errorf("index: inserting code_files row for %s: %w", relPath, err)
	}
	codeFileID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	for _, d := range decls {
		body := sql.NullString{String: d.Body, Valid: d.Body != ""}
		if _, err := tx.Exec(
			`INSERT INTO code_declarations (code_file_id, name, kind, signature, body, start_line, end_line, is_test) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			codeFileID, d.Name, d.Kind, d.Signature, body, d.StartLine, d.EndLine, boolToInt(isTestFile),
		); err != nil {
			// Undo the code_files row already inserted for this file
			// so a mid-loop failure never leaves a fingerprint-matching
			// row with only some of its declarations committed (which
			// would make a later SyncCode skip re-indexing it as
			// "unchanged").
			_ = deleteCodeFile(tx, codeFileID)
			return fmt.Errorf("index: inserting code_declarations row for %s: %w", relPath, err)
		}
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
