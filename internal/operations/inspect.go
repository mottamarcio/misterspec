package operations

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/evidence"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// InspectResult is an entity's structured metadata, plus where it lives
// (FR-004, FR-005).
type InspectResult struct {
	Location ResolvedLocation
	Metadata artifacts.Metadata
}

// Inspect resolves rawID, then retrieves its metadata (FR-004, FR-005,
// FR-016). For a Task — which has no frontmatter of its own — Metadata is
// narrowed to ID, Status (derived from its Markdown checkbox), and Parent
// (derived from its owning tasks.md's own `for:` field); DependsOn and
// Supersedes are left empty, per research.md's Task metadata scope
// decision.
func Inspect(root string, cfg project.Configuration, rawID string) (InspectResult, error) {
	var loc ResolvedLocation
	var err error
	if strings.Contains(rawID, ":") {
		// A composite Task reference ("SPEC-014:TASK-003") — Resolve's
		// generic path never accepts this syntax (ids.ParseAny only
		// parses a single-prefix ID), so it goes straight to ResolveTask
		// (031-canonical-task-identity contracts/task-identity-resolution.md §2).
		loc, err = ResolveTask(root, cfg, rawID, nil)
	} else {
		loc, err = Resolve(root, cfg, rawID)
	}
	if err != nil {
		return InspectResult{}, err
	}

	if loc.ID.Type == ids.Task {
		meta, err := inspectTask(root, loc)
		if err != nil {
			return InspectResult{}, err
		}
		return InspectResult{Location: loc, Metadata: meta}, nil
	}

	meta, err := artifacts.ParseMetadata(filepath.Join(root, loc.Path))
	if err != nil {
		return InspectResult{}, err
	}
	return InspectResult{Location: loc, Metadata: meta}, nil
}

var (
	// taskHeadingTokenPattern matches a Task Section's own Heading text
	// (already stripped of leading "#"s by artifacts.ParseDocument),
	// e.g. "TASK-003" alone or "TASK-003 — Add session persistence" —
	// group 1 is always just "TASK-003", the token compared against the
	// heading fragment a Task's own ResolvedLocation.Path carries
	// (ids.ScanTasks's own taskHeadingPattern precedent).
	taskHeadingTokenPattern = regexp.MustCompile(`^(TASK-\d+)\b`)
	taskCheckboxPattern     = regexp.MustCompile(`^- \[([ xX])\]`)
)

// inspectTask builds the narrowed Metadata for a Task location
// ("<tasks.md path>#TASK-NNN"): its own tasks.md's `for:` field becomes
// Parent, and Status is derived exactly as internal/prepare.
// ScanSpecTasks already derives it — checked AND evidence.Verified,
// never the checkbox alone (evidence.TaskStatus). Code review finding
// (041-task-evidence-fingerprint): this function originally derived
// Status from the checkbox alone via a raw line scan, contradicting
// internal/prepare's own, already-corrected definition for the same
// Task — `internal inspect` and `internal prepare`/`internal validate`
// could report a different state for the identical Task. Rewritten to
// use artifacts.ParseDocument (like ScanSpecTasks) so the Task's own
// full Section.Body is available for evidence parsing and
// fingerprinting, not just its checkbox line.
func inspectTask(root string, loc ResolvedLocation) (artifacts.Metadata, error) {
	filePath, heading, ok := strings.Cut(loc.Path, "#")
	if !ok {
		return artifacts.Metadata{}, fmt.Errorf("operations: malformed task location %q", loc.Path)
	}

	fileMeta, err := artifacts.ParseMetadata(filepath.Join(root, filePath))
	if err != nil {
		return artifacts.Metadata{}, err
	}

	status, err := taskEvidenceAwareStatus(root, filePath, heading, loc.ID)
	if err != nil {
		return artifacts.Metadata{}, err
	}

	id := loc.ID
	return artifacts.Metadata{
		ID:     &id,
		Type:   "task",
		Status: status,
		Parent: fileMeta.For,
	}, nil
}

// taskEvidenceAwareStatus finds heading's own Section in filePath (the
// owning tasks.md), then returns evidence.TaskStatus of its checkbox
// and evidence.DeriveState — the identical definition
// internal/prepare.ScanSpecTasks uses, so the two never diverge again.
func taskEvidenceAwareStatus(root, filePath, heading string, task ids.EntityID) (string, error) {
	body, err := artifacts.ReadBody(filepath.Join(root, filePath))
	if err != nil {
		if errors.Is(err, artifacts.ErrArtifactNotFound) {
			return "", fmt.Errorf("%w: %s", artifacts.ErrArtifactNotFound, filePath)
		}
		return "", err
	}

	doc := artifacts.ParseDocument(body)
	for _, section := range doc.Sections {
		sub := taskHeadingTokenPattern.FindStringSubmatch(section.Heading)
		if sub == nil || sub[1] != heading {
			continue
		}
		sectionBody := []byte(section.Body)
		checked := false
		for _, line := range strings.Split(section.Body, "\n") {
			if cb := taskCheckboxPattern.FindStringSubmatch(line); cb != nil {
				checked = cb[1] != " "
				break
			}
		}
		fields := evidence.ParseEvidenceFields(task, sectionBody)
		currentFingerprint := TaskContentFingerprint(task, sectionBody).String()
		state := evidence.DeriveState(fields, currentFingerprint)
		return evidence.TaskStatus(checked, state), nil
	}
	return "", fmt.Errorf("operations: heading %q not found in %s", heading, filePath)
}
