package operations

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mottamarcio/misterspec/internal/artifacts"
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
	taskHeadingLinePattern = regexp.MustCompile(`^##\s+(TASK-\d+)\b`)
	taskCheckboxPattern    = regexp.MustCompile(`^- \[([ xX])\]`)
)

// inspectTask builds the narrowed Metadata for a Task location
// ("<tasks.md path>#TASK-NNN"): its own tasks.md's `for:` field becomes
// Parent, and its own checkbox line becomes Status.
func inspectTask(root string, loc ResolvedLocation) (artifacts.Metadata, error) {
	filePath, heading, ok := strings.Cut(loc.Path, "#")
	if !ok {
		return artifacts.Metadata{}, fmt.Errorf("operations: malformed task location %q", loc.Path)
	}

	fileMeta, err := artifacts.ParseMetadata(filepath.Join(root, filePath))
	if err != nil {
		return artifacts.Metadata{}, err
	}

	status, err := taskCheckboxStatus(filepath.Join(root, filePath), heading)
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

// taskCheckboxStatus scans path for the heading "## <heading>" and
// returns "pending" or "complete" based on the first Markdown checkbox
// line found under it (docs/architecture-specification.md §29's
// "- [ ] Complete" convention), before the next "## " heading or EOF.
func taskCheckboxStatus(path, heading string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%w: %s", artifacts.ErrArtifactNotFound, path)
		}
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inTarget := false
	for scanner.Scan() {
		line := scanner.Text()
		if sub := taskHeadingLinePattern.FindStringSubmatch(line); sub != nil {
			inTarget = sub[1] == heading
			continue
		}
		if inTarget {
			if cb := taskCheckboxPattern.FindStringSubmatch(line); cb != nil {
				if cb[1] == " " {
					return "pending", nil
				}
				return "complete", nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("operations: heading %q not found in %s", heading, path)
}
