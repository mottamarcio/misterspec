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

// ResolveTarget parses raw via ParseAny (width-tolerant, type-inferring —
// the same tolerance ParseMetadata's frontmatter fields and
// internal/validation's wikilink classification already rely on) and
// then Scans for it, returning every matching artifact path: zero
// (unresolved), one (resolved), or more than one (ambiguous — a
// duplicate ID). Only a genuinely malformed raw string is an error; zero
// or many matches are not (012-references-backlinks/research.md #2 — a
// small, shared composition of ParseAny+Scan, factored out of
// internal/validation/wikilinks.go's classifyWikilink, which used to
// repeat this same pair of calls inline).
func ResolveTarget(root string, cfg project.Configuration, raw string) (EntityID, []string, error) {
	id, err := ParseAny(raw)
	if err != nil {
		return EntityID{}, nil, err
	}

	result, err := Scan(root, cfg, id.Type)
	if err != nil {
		return EntityID{}, nil, err
	}

	return id, result.Paths[id.Number], nil
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

// taskClaim is one "## TASK-NNN" heading found while scanning tasks.md
// files, tagged with the Spec number of the tasks.md file it came from —
// the single shared low-level scan every Task-identity view in this
// package (scanTaskHeadings's project-wide map, and ScanTasks's
// per-Spec/cross-Spec views) is built from
// (031-canonical-task-identity/research.md Decision 2, FR-005).
type taskClaim struct {
	Spec int
	Task int
	Path string // "<tasks.md path>#TASK-NNN"
}

// scanTaskClaims globs pattern (relative to root) for tasks.md files and
// parses their "## TASK-NNN — ..." headings (docs/architecture-specification.md
// §29), tagging each with the owning Spec's number (parsed from the
// tasks.md file's own "SPEC-###" parent directory, the same convention
// scanDirs already uses via idDirPattern). A tasks.md whose parent
// directory does not match the canonical "SPEC-###" shape is skipped —
// it cannot be attributed to an owning Spec, mirroring how scanDirs
// silently skips a malformed entity directory name.
func scanTaskClaims(root, pattern string) ([]taskClaim, error) {
	matches, err := filepath.Glob(filepath.Join(root, pattern))
	if err != nil {
		return nil, err
	}

	var claims []taskClaim
	for _, m := range matches {
		specNumber, ok := specNumberFromTasksPath(m)
		if !ok {
			continue
		}

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
			claims = append(claims, taskClaim{
				Spec: specNumber,
				Task: number,
				Path: fmt.Sprintf("%s#%s", rel, sub[1]),
			})
		}
		scanErr := scanner.Err()
		f.Close()
		if scanErr != nil {
			return nil, fmt.Errorf("ids: scanning %s: %w", rel, scanErr)
		}
	}
	return claims, nil
}

// specNumberFromTasksPath extracts the Spec number from a tasks.md
// absolute path's parent directory (e.g.
// ".../specs/SPEC-014/tasks.md" -> 14, true), or (0, false) if that
// directory does not match the canonical "SPEC-###" shape.
func specNumberFromTasksPath(taskFilePath string) (int, bool) {
	base := filepath.Base(filepath.Dir(taskFilePath))
	sub := idDirPattern.FindStringSubmatch(base)
	if sub == nil || sub[1] != Spec.Prefix() {
		return 0, false
	}
	number, err := strconv.Atoi(sub[2])
	if err != nil || number < 1 {
		return 0, false
	}
	return number, true
}

// scanTaskHeadings returns a map of number -> claiming "path#TASK-NNN"
// references, project-wide across every Spec — the same shape and
// cross-Spec-merging behavior Scan(..., Task) has always had, preserved
// here for callers (operations.Status, generic Scan) that only need a
// raw count/enumeration of every Task heading and do not need per-Spec
// scoping (031-canonical-task-identity/research.md Decision 2 — this is
// a view built on scanTaskClaims, not a second parser). Callers that
// need per-Spec duplicate detection or cross-Spec collision detection
// use ScanTasks instead.
func scanTaskHeadings(root, pattern string) (map[int][]string, error) {
	claims, err := scanTaskClaims(root, pattern)
	if err != nil {
		return nil, err
	}

	byNumber := map[int][]string{}
	for _, c := range claims {
		byNumber[c.Task] = append(byNumber[c.Task], c.Path)
	}
	return byNumber, nil
}

// TaskScanEntry is one Task heading found while scanning the project,
// tagged with the Spec that owns it
// (031-canonical-task-identity/data-model.md "TaskScanEntry").
type TaskScanEntry struct {
	Spec int
	Task int
	Path string
}

// TaskDuplicate is more than one "## TASK-NNN" heading with the same
// number inside the *same* Spec's tasks.md — a genuine identity
// collision (spec FR-004), distinct from TaskCollision below.
type TaskDuplicate struct {
	Spec  int
	Task  int
	Paths []string // sorted, length > 1
}

// TaskCollision is a Task number claimed by more than one Spec
// project-wide — expected, valid state under the per-Spec identity model
// (spec FR-001, User Story 1), never a duplicate. It is what a bare
// "TASK-NNN" reference with no Spec context collides against (spec
// FR-003) and what the migration diagnostic reports (spec FR-008).
type TaskCollision struct {
	Task  int
	Specs []int    // sorted, length > 1
	Paths []string // one entry per (Spec, first claiming path), aligned with Specs
}

// TaskScanResult is ScanTasks's result: every Task heading found, plus
// the two derived views every Task-identity consumer in this project
// needs (031-canonical-task-identity/research.md Decision 2).
type TaskScanResult struct {
	Entries    []TaskScanEntry
	Duplicates []TaskDuplicate
	Collisions []TaskCollision
	// BySpec maps Spec number -> Task number -> claiming paths, for
	// resolving a Task within a known owning Spec.
	BySpec map[int]map[int][]string
	// ByNumber maps Task number -> Spec number -> claiming paths, for
	// resolving a bare Task reference with no Spec context.
	ByNumber map[int]map[int][]string
}

// ScanTasks scans every tasks.md in the project (the same underlying
// scanTaskClaims pass scanTaskHeadings uses) and returns the per-Spec and
// cross-Spec views needed to resolve a Task unambiguously and to detect
// duplicates scoped to the correct Spec
// (031-canonical-task-identity/data-model.md "TaskScanEntry").
func ScanTasks(root string, cfg project.Configuration) (TaskScanResult, error) {
	pattern := filepath.Join(cfg.ProgramsRoot, "*", "features", "*", "specs", "*", "tasks.md")
	claims, err := scanTaskClaims(root, pattern)
	if err != nil {
		return TaskScanResult{}, err
	}

	result := TaskScanResult{
		BySpec:   map[int]map[int][]string{},
		ByNumber: map[int]map[int][]string{},
	}
	for _, c := range claims {
		result.Entries = append(result.Entries, TaskScanEntry(c))

		if result.BySpec[c.Spec] == nil {
			result.BySpec[c.Spec] = map[int][]string{}
		}
		result.BySpec[c.Spec][c.Task] = append(result.BySpec[c.Spec][c.Task], c.Path)

		if result.ByNumber[c.Task] == nil {
			result.ByNumber[c.Task] = map[int][]string{}
		}
		result.ByNumber[c.Task][c.Spec] = append(result.ByNumber[c.Task][c.Spec], c.Path)
	}

	for specNum, byTask := range result.BySpec {
		for taskNum, paths := range byTask {
			if len(paths) < 2 {
				continue
			}
			sorted := append([]string(nil), paths...)
			sort.Strings(sorted)
			result.Duplicates = append(result.Duplicates, TaskDuplicate{Spec: specNum, Task: taskNum, Paths: sorted})
		}
	}
	sort.Slice(result.Duplicates, func(i, j int) bool {
		if result.Duplicates[i].Spec != result.Duplicates[j].Spec {
			return result.Duplicates[i].Spec < result.Duplicates[j].Spec
		}
		return result.Duplicates[i].Task < result.Duplicates[j].Task
	})

	for taskNum, bySpec := range result.ByNumber {
		if len(bySpec) < 2 {
			continue
		}
		specs := make([]int, 0, len(bySpec))
		for s := range bySpec {
			specs = append(specs, s)
		}
		sort.Ints(specs)
		var paths []string
		for _, s := range specs {
			paths = append(paths, bySpec[s]...)
		}
		result.Collisions = append(result.Collisions, TaskCollision{Task: taskNum, Specs: specs, Paths: paths})
	}
	sort.Slice(result.Collisions, func(i, j int) bool { return result.Collisions[i].Task < result.Collisions[j].Task })

	return result, nil
}
