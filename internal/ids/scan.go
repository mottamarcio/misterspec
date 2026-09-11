package ids

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mottamarcio/misterspec/internal/project"
)

// DuplicateID records an entity ID number claimed by more than one
// artifact of the same type, found during Scan (FR-012).
type DuplicateID struct {
	ID    EntityID
	Paths []string // artifact paths (relative to root) sharing this number, sorted
}

// ScanResult is the result of scanning a project for existing IDs of one
// entity type (FR-011, FR-012, FR-014).
type ScanResult struct {
	IDs        []EntityID
	Duplicates []DuplicateID
	// Paths maps every found number to every path (relative to root)
	// claiming it — length 1 for an unambiguous ID, length >1 for a
	// duplicated one (the same data Duplicates surfaces). Added for
	// 002-read-operations' Resolve/Children, which need a single match's
	// path, not only a duplicate's.
	Paths map[int][]string
}

// taskHeadingPattern matches a Tasks artifact's per-task Markdown heading,
// e.g. "## TASK-001 — Add session persistence" (docs/architecture-specification.md §29).
var taskHeadingPattern = regexp.MustCompile(`^##\s+(TASK-(\d+))\b`)

// Scan walks the project tree under cfg's configured roots and enumerates
// every syntactically valid, existing ID for entity type t, per
// docs/architecture-specification.md's canonical layout. Duplicate
// numbers are collected into ScanResult.Duplicates without halting
// enumeration of the rest of the project (FR-012); a type with zero
// existing artifacts returns an empty ScanResult, not an error (FR-014).
func Scan(root string, cfg project.Configuration, t EntityType) (ScanResult, error) {
	width := cfg.IDWidth

	var claims map[int][]string
	var err error

	switch t {
	case Program:
		claims, err = scanDirs(root, filepath.Join(cfg.ProgramsRoot, "*"), t)
	case Feature:
		claims, err = scanDirs(root, filepath.Join(cfg.ProgramsRoot, "*", "features", "*"), t)
	case Spec:
		claims, err = scanDirs(root, filepath.Join(cfg.ProgramsRoot, "*", "features", "*", "specs", "*"), t)
	case Knowledge:
		claims, err = scanFlatFiles(root, cfg.KnowledgeDir, t)
	case Learning:
		claims, err = scanFlatFiles(root, cfg.LearningsDir, t)
	case Task:
		claims, err = scanTaskHeadings(root, filepath.Join(cfg.ProgramsRoot, "*", "features", "*", "specs", "*", "tasks.md"))
	default:
		return ScanResult{}, fmt.Errorf("ids: unsupported entity type for Scan: %v", t)
	}
	if err != nil {
		return ScanResult{}, err
	}

	return buildScanResult(t, width, claims), nil
}

// buildScanResult turns a number->claiming-paths map into a ScanResult,
// sorted for deterministic output.
func buildScanResult(t EntityType, width int, claims map[int][]string) ScanResult {
	numbers := make([]int, 0, len(claims))
	for n := range claims {
		numbers = append(numbers, n)
	}
	sort.Ints(numbers)

	result := ScanResult{Paths: map[int][]string{}}
	for _, n := range numbers {
		paths := claims[n]
		sort.Strings(paths)
		id := EntityID{Type: t, Prefix: t.Prefix(), Number: n, Width: width}
		result.IDs = append(result.IDs, id)
		result.Paths[n] = paths
		if len(paths) > 1 {
			result.Duplicates = append(result.Duplicates, DuplicateID{ID: id, Paths: paths})
		}
	}
	return result
}

// idDirPattern matches a canonical entity directory name, e.g. "SPEC-014".
var idDirPattern = regexp.MustCompile(`^([A-Z]+)-(\d+)$`)

// scanDirs globs pattern (relative to root) for directories whose base
// name matches t's prefix and a numeric suffix, returning a map of
// number -> claiming directory paths (relative to root).
func scanDirs(root, pattern string, t EntityType) (map[int][]string, error) {
	matches, err := filepath.Glob(filepath.Join(root, pattern))
	if err != nil {
		return nil, err
	}

	claims := map[int][]string{}
	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil || !info.IsDir() {
			continue
		}
		sub := idDirPattern.FindStringSubmatch(filepath.Base(m))
		if sub == nil || sub[1] != t.Prefix() {
			continue
		}
		number, err := strconv.Atoi(sub[2])
		if err != nil || number < 1 {
			continue
		}
		rel, err := filepath.Rel(root, m)
		if err != nil {
			continue
		}
		claims[number] = append(claims[number], filepath.ToSlash(rel))
	}
	return claims, nil
}

// scanFlatFiles globs dir (relative to root) for files whose name starts
// with t's prefix and a numeric suffix (e.g. "KNOW-001-authentication.md"
// — the slug after the number is not part of the ID), returning a map of
// number -> claiming file paths (relative to root).
func scanFlatFiles(root, dir string, t EntityType) (map[int][]string, error) {
	matches, err := filepath.Glob(filepath.Join(root, dir, t.Prefix()+"-*"))
	if err != nil {
		return nil, err
	}

	filePattern := regexp.MustCompile(`^` + regexp.QuoteMeta(t.Prefix()) + `-(\d+)(?:[-.].*)?$`)

	claims := map[int][]string{}
	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil || info.IsDir() {
			continue
		}
		sub := filePattern.FindStringSubmatch(filepath.Base(m))
		if sub == nil {
			continue
		}
		number, err := strconv.Atoi(sub[1])
		if err != nil || number < 1 {
			continue
		}
		rel, err := filepath.Rel(root, m)
		if err != nil {
			continue
		}
		claims[number] = append(claims[number], filepath.ToSlash(rel))
	}
	return claims, nil
}

// scanTaskHeadings globs pattern (relative to root) for tasks.md files and
// parses their "## TASK-NNN — ..." headings (docs/architecture-specification.md
// §29), returning a map of number -> claiming "path#TASK-NNN" references.
// Unlike other entity types, Task IDs are declared inside a shared file
// rather than named by their own directory or filename.
func scanTaskHeadings(root, pattern string) (map[int][]string, error) {
	matches, err := filepath.Glob(filepath.Join(root, pattern))
	if err != nil {
		return nil, err
	}

	claims := map[int][]string{}
	for _, m := range matches {
		rel, err := filepath.Rel(root, m)
		if err != nil {
			continue
		}
		rel = filepath.ToSlash(rel)

		f, err := os.Open(m)
		if err != nil {
			return nil, fmt.Errorf("ids: reading %s: %w", rel, err)
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimRight(scanner.Text(), " \t")
			sub := taskHeadingPattern.FindStringSubmatch(line)
			if sub == nil {
				continue
			}
			number, err := strconv.Atoi(sub[2])
			if err != nil || number < 1 {
				continue
			}
			claims[number] = append(claims[number], fmt.Sprintf("%s#%s", rel, sub[1]))
		}
		scanErr := scanner.Err()
		f.Close()
		if scanErr != nil {
			return nil, fmt.Errorf("ids: scanning %s: %w", rel, scanErr)
		}
	}
	return claims, nil
}
