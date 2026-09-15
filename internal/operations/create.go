package operations

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/lock"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/templates"
	"github.com/mottamarcio/misterspec/internal/vcs"
)

// ErrInvalidParent is returned when a Create/CreateArtifact request's
// declared parent doesn't exist, is ambiguous, or is not the correct type
// (FR-004, FR-011).
var ErrInvalidParent = errors.New("operations: invalid parent")

// ErrAlreadyExists is returned when a creation's target canonical path
// already has an artifact (FR-008, FR-012).
var ErrAlreadyExists = errors.New("operations: target already exists")

// ErrUnsupportedType is returned when a Create/CreateArtifact request
// names a type this feature does not create.
var ErrUnsupportedType = errors.New("operations: unsupported type")

// ErrInvalidSlug is returned when a Knowledge/Learning creation's Slug is
// empty or not filesystem-safe.
var ErrInvalidSlug = errors.New("operations: invalid slug")

// CreateRequest is the input to Create (FR-001, FR-009).
type CreateRequest struct {
	// Type must be one of Program, Feature, Spec, Knowledge, Learning.
	Type ids.EntityType
	// Parent is the raw parent ID; required for Feature (a Program) and
	// Spec (a Feature), ignored otherwise.
	Parent string
	// Slug is required for Knowledge/Learning (used verbatim in the
	// filename), ignored otherwise.
	Slug string
}

// CreateResult is the outcome of a successful Create.
type CreateResult struct {
	ID   ids.EntityID
	Path string // relative to root

	// GitBranch is the relevant Git branch name (022-feature-branch-
	// automation) — for a Feature, the branch just ensured/created; for
	// a Spec, its parent Feature's own branch. Empty when automation
	// did not run at all (see GitSkippedReason) or for entity types
	// this feature never touches.
	GitBranch string
	// GitBranchCreated is true only when GitBranch was newly created by
	// this call (Feature creation only) — false when an existing branch
	// was resumed instead, and always false for Spec creation.
	GitBranchCreated bool
	// GitSkippedReason is "not_a_git_repo" or "disabled" when Git
	// automation would otherwise apply but did not run; empty
	// otherwise.
	GitSkippedReason string
	// GitWarning is set only for Spec creation, only when the
	// currently checked-out branch does not match the parent Feature's
	// own branch — never blocks the creation.
	GitWarning string
}

// createableTypes are the entity types Create supports (FR-009). Task is
// deliberately excluded: it has no independent file/directory of its
// own — its "creation" is an agent editing tasks.md content, not this
// deterministic layer's job.
var createableTypes = map[ids.EntityType]bool{
	ids.Program:   true,
	ids.Feature:   true,
	ids.Spec:      true,
	ids.Knowledge: true,
	ids.Learning:  true,
}

// Create atomically allocates the next ID for req.Type, creates its
// canonical directory, and writes its initial artifact from that type's
// template — or does none of that, returning an error (FR-001..FR-009).
func Create(root string, cfg project.Configuration, req CreateRequest) (CreateResult, error) {
	if !createableTypes[req.Type] {
		return CreateResult{}, fmt.Errorf("%w: %v", ErrUnsupportedType, req.Type)
	}

	l, err := lock.Acquire(root, lock.DefaultStaleAfter, lock.DefaultTimeout)
	if err != nil {
		return CreateResult{}, err
	}
	defer l.Release()

	// parents is ordered outermost to innermost, matching
	// artifacts.ResolvePath's variadic parents parameter.
	var parents []ids.EntityID
	// parentFeatureID is set only for Spec creation (022-feature-
	// branch-automation) — the parent Feature's own ID, used to derive
	// the branch a Spec is expected to land on.
	var parentFeatureID string

	switch req.Type {
	case ids.Feature:
		prg, err := validateParent(root, cfg, req.Parent, ids.Program)
		if err != nil {
			return CreateResult{}, err
		}
		parents = []ids.EntityID{prg.ID}

	case ids.Spec:
		feat, err := validateParent(root, cfg, req.Parent, ids.Feature)
		if err != nil {
			return CreateResult{}, err
		}
		featParent, err := Parent(root, cfg, feat.ID.String())
		if err != nil {
			return CreateResult{}, err
		}
		if !featParent.HasParent {
			return CreateResult{}, fmt.Errorf("%w: %s has no program parent", ErrInvalidParent, feat.ID)
		}
		parents = []ids.EntityID{featParent.Parent.ID, feat.ID}
		parentFeatureID = feat.ID.String()

	case ids.Knowledge, ids.Learning:
		if err := validateSlug(req.Slug); err != nil {
			return CreateResult{}, err
		}
	}

	scanResult, err := ids.Scan(root, cfg, req.Type)
	if err != nil {
		return CreateResult{}, err
	}
	newID := ids.NextID(scanResult.IDs, req.Type, cfg.IDWidth)

	canonical, err := artifacts.ResolvePath(root, cfg, req.Type, newID, parents...)
	if err != nil {
		return CreateResult{}, err
	}

	targetRel := canonical.File
	if targetRel == "" {
		// Knowledge/Learning: ResolvePath only gives the directory (its
		// filename carries an agent-chosen slug it isn't given) — see
		// 002-read-operations's ResolvePath contract.
		targetRel = canonical.Directory + "/" + newID.String() + "-" + req.Slug + ".md"
	}

	targetAbs := filepath.Join(root, targetRel)
	if _, statErr := os.Stat(targetAbs); statErr == nil {
		return CreateResult{}, fmt.Errorf("%w: %s", ErrAlreadyExists, targetRel)
	} else if !os.IsNotExist(statErr) {
		return CreateResult{}, statErr
	}

	result := CreateResult{ID: newID, Path: targetRel}

	switch req.Type {
	case ids.Feature:
		branch, created, skippedReason, err := ensureFeatureBranch(root, cfg, newID.String(), req.Slug)
		if err != nil {
			return CreateResult{}, err
		}
		result.GitBranch, result.GitBranchCreated, result.GitSkippedReason = branch, created, skippedReason

	case ids.Spec:
		branch, warning, skippedReason, err := checkSpecBranch(root, cfg, parentFeatureID)
		if err != nil {
			return CreateResult{}, err
		}
		result.GitBranch, result.GitWarning, result.GitSkippedReason = branch, warning, skippedReason
	}

	content, err := renderCreate(req.Type, newID, parents)
	if err != nil {
		return CreateResult{}, err
	}

	if err := writeAtomic(targetAbs, content); err != nil {
		return CreateResult{}, err
	}

	return result, nil
}

// ensureFeatureBranch creates-and-checks-out (or resumes) the dedicated
// Git branch for a newly allocated Feature ID, using slug (if given) to
// make the branch name readable at a glance (022-feature-branch-
// automation, FR-001/FR-002/FR-007), or reports why it did not
// (FR-003/FR-006) — never an error for either of those two cases.
func ensureFeatureBranch(root string, cfg project.Configuration, featureID, slug string) (branch string, created bool, skippedReason string, err error) {
	if !cfg.GitBranchAutomation {
		return "", false, "disabled", nil
	}
	if !vcs.IsRepo(root) {
		return "", false, "not_a_git_repo", nil
	}

	branch, created, err = vcs.EnsureFeatureBranch(root, featureID, slug)
	if err != nil {
		return "", false, "", err
	}
	return branch, created, "", nil
}

// checkSpecBranch reports the parent Feature's own branch and, if the
// currently checked-out branch differs from it, a non-blocking warning
// (022-feature-branch-automation, FR-004/FR-005) — a Spec never creates
// or switches a branch itself. Returns an empty branch (no warning
// possible) when the parent Feature has no dedicated branch at all yet
// — e.g. it was created before automation was enabled.
func checkSpecBranch(root string, cfg project.Configuration, parentFeatureID string) (branch, warning, skippedReason string, err error) {
	if !cfg.GitBranchAutomation {
		return "", "", "disabled", nil
	}
	if !vcs.IsRepo(root) {
		return "", "", "not_a_git_repo", nil
	}

	branch, err = vcs.BranchForID(root, parentFeatureID)
	if err != nil {
		return "", "", "", err
	}
	if branch == "" {
		return "", "", "", nil
	}

	current, err := vcs.CurrentBranch(root)
	if err != nil {
		return "", "", "", err
	}
	if current != "" && current != branch {
		warning = fmt.Sprintf("current branch %q does not match parent Feature %s's own branch %q", current, parentFeatureID, branch)
	}
	return branch, warning, "", nil
}

// validateParent resolves rawParent and confirms it exists, unambiguously,
// as wantType — reusing Resolve (002-read-operations) rather than
// reimplementing existence/ambiguity checking (Constitution Principle VI).
func validateParent(root string, cfg project.Configuration, rawParent string, wantType ids.EntityType) (ResolvedLocation, error) {
	if rawParent == "" {
		return ResolvedLocation{}, fmt.Errorf("%w: a %v parent is required", ErrInvalidParent, wantType)
	}
	loc, err := Resolve(root, cfg, rawParent)
	if err != nil {
		return ResolvedLocation{}, fmt.Errorf("%w: %v", ErrInvalidParent, err)
	}
	if loc.ID.Type != wantType {
		return ResolvedLocation{}, fmt.Errorf("%w: %s is a %v, want a %v", ErrInvalidParent, rawParent, loc.ID.Type, wantType)
	}
	return loc, nil
}

// validateSlug rejects an empty or filesystem-unsafe slug (FR-009's
// Knowledge/Learning filenames use it verbatim).
func validateSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("%w: slug is required", ErrInvalidSlug)
	}
	if strings.ContainsAny(slug, "/\\") || strings.Contains(slug, "..") {
		return fmt.Errorf("%w: %q is not filesystem-safe", ErrInvalidSlug, slug)
	}
	return nil
}

// renderCreate renders req.Type's initial content via internal/templates.
func renderCreate(t ids.EntityType, id ids.EntityID, parents []ids.EntityID) (string, error) {
	switch t {
	case ids.Program:
		return templates.Render(templates.Program, templates.ProgramData{ID: id.String()})
	case ids.Feature:
		return templates.Render(templates.Feature, templates.FeatureData{ID: id.String(), Parent: parents[0].String()})
	case ids.Spec:
		return templates.Render(templates.Spec, templates.SpecData{ID: id.String(), Parent: parents[len(parents)-1].String()})
	case ids.Knowledge:
		return templates.Render(templates.Knowledge, templates.KnowledgeData{ID: id.String()})
	case ids.Learning:
		return templates.Render(templates.Learning, templates.LearningData{ID: id.String()})
	default:
		return "", fmt.Errorf("%w: %v", ErrUnsupportedType, t)
	}
}
