package installer

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/kit"
)

// Resource is one embedded item available for installation (FR-002).
type Resource struct {
	// Name is the resource's path relative to sourceDir, using fs.FS's
	// own forward-slash convention (e.g. "program.md.tmpl" for a flat
	// source, "create-plan/SKILL.md" for a nested one — a bare filename
	// is simply the zero-nesting case; 009-canonical-skills-content's
	// research.md).
	Name string
	// Kind categorizes this resource — "template" for everything
	// 005-embedded-kit embeds; a later feature's source (e.g. Skills)
	// supplies its own Kind via ListFS/InstallFS. A string, not a closed
	// enum, so new kinds never require a breaking change.
	Kind string
	// ArtifactType is, for a template resource, the entity/artifact type
	// it belongs to (e.g. "program"), derived from its filename by the
	// "<type>.md.tmpl" convention. For a non-template resource this is
	// simply its filename unchanged (the convention doesn't match, so
	// nothing is trimmed) — harmless, since only template callers rely
	// on it.
	ArtifactType string
}

const templatesDir = "templates"

// List enumerates every resource in the embedded kit, derived directly
// from kit.TemplatesFS's own contents — never a separately maintained
// list that could drift from what's actually embedded (FR-002, FR-009).
// It performs no filesystem access outside the embedded FS itself, and
// returns an identical result on every call.
//
// List is ListFS(kit.TemplatesFS, "templates", "template") — kept as its
// own function so every existing caller's signature is unchanged
// (006-agent-adapter's research.md).
func List() []Resource {
	return ListFS(kit.TemplatesFS, templatesDir, "template")
}

// ListFS is List's generalized form: source and sourceDir replace the
// hardcoded kit.TemplatesFS/"templates", and kind sets every returned
// Resource.Kind — so a caller installing a different resource type (e.g.
// Skills) doesn't inherit "template" as a mislabel.
//
// ListFS walks sourceDir recursively (009-canonical-skills-content's
// research.md) — not only its immediate children — so a resource set
// with real subdirectories (a Skill's own "<name>/SKILL.md" layout, for
// example) is discovered in full. Resource.Name is each file's path
// relative to sourceDir; for a flat source (every kit template, every
// fixture in this package's own tests) that is simply the bare
// filename, unchanged from before this generalization.
func ListFS(source fs.FS, sourceDir, kind string) []Resource {
	var resources []Resource

	err := fs.WalkDir(source, sourceDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		name := p
		if sourceDir != "." {
			name = strings.TrimPrefix(p, sourceDir+"/")
		}

		resources = append(resources, Resource{
			Name:         name,
			Kind:         kind,
			ArtifactType: artifactTypeFromFilename(path.Base(name)),
		})
		return nil
	})
	if err != nil {
		// The embedded kit is compiled into the binary; a walk failure
		// here means the embed itself is broken, not a runtime
		// condition callers can meaningfully recover from. A
		// caller-supplied fixture fs.FS in a test is expected to be
		// well-formed for the same reason.
		panic("installer: walking " + sourceDir + ": " + err.Error())
	}

	sort.Slice(resources, func(i, j int) bool { return resources[i].Name < resources[j].Name })
	return resources
}

// artifactTypeFromFilename derives a template's artifact type from its
// filename by convention: "<type>.md.tmpl" → "<type>".
func artifactTypeFromFilename(name string) string {
	return strings.TrimSuffix(name, ".md.tmpl")
}

// OutcomeStatus is the per-resource result of one Install call (FR-005).
type OutcomeStatus int

const (
	// Installed means the resource was written, byte-for-byte identical
	// to its embedded source.
	Installed OutcomeStatus = iota
	// Skipped means the target already existed and overwrite was not
	// requested — the existing file was left untouched.
	Skipped
	// Failed means the resource could not be installed; Err names why.
	Failed
)

// Outcome is the result of installing one Resource.
type Outcome struct {
	Resource Resource
	Status   OutcomeStatus
	// Path is the resource's destination, relative to targetDir.
	Path string
	// Err is set only when Status == Failed.
	Err error
}

// Install materializes every resource List() reports into targetDir, one
// atomic write per resource (FR-003, FR-007). A resource whose target
// already exists is Skipped unless overwrite is true (FR-004). A
// resource whose destination would resolve outside targetDir is Failed
// before anything is written (FR-006), reusing
// artifacts.RelativeWithinRoot rather than a new containment check.
//
// Install returns exactly one Outcome per resource — never a single
// pass/fail flag for the whole call (FR-005). Its own returned error is
// reserved for something that prevents the operation from running at
// all; a single resource's own problem is always an Outcome, never this
// error.
//
// Install is InstallFS(kit.TemplatesFS, "templates", "template",
// targetDir, overwrite) — kept as its own function so every existing
// caller's signature is unchanged (006-agent-adapter's research.md).
func Install(targetDir string, overwrite bool) ([]Outcome, error) {
	return InstallFS(kit.TemplatesFS, templatesDir, "template", targetDir, overwrite)
}

// InstallFS is Install's generalized form: source, sourceDir, and kind
// replace the hardcoded kit.TemplatesFS/"templates"/"template", so a
// second consumer (e.g. internal/agents/claude installing Skills) reuses
// this exact atomic-write, no-silent-overwrite, containment-checked
// mechanism instead of a second implementation.
func InstallFS(source fs.FS, sourceDir, kind, targetDir string, overwrite bool) ([]Outcome, error) {
	resources := ListFS(source, sourceDir, kind)
	outcomes := make([]Outcome, 0, len(resources))
	for _, r := range resources {
		outcomes = append(outcomes, installOneFS(source, sourceDir, targetDir, r, overwrite))
	}
	return outcomes, nil
}

// installOneFS installs a single resource, checking containment before
// any write (FR-006). Exported as a package-internal function so its
// containment guarantee can be unit-tested directly (containment_test.go)
// independent of whether a given source's resource set can ever trigger
// it.
func installOneFS(source fs.FS, sourceDir, targetDir string, r Resource, overwrite bool) Outcome {
	// r.Name uses fs.FS's forward-slash convention (possibly multi-
	// segment for a nested resource); FromSlash converts it to the
	// host's own separator before joining into an OS path (research.md
	// — correct beyond Unix, not only coincidentally correct there).
	relDest := filepath.Join(sourceDir, filepath.FromSlash(r.Name))
	destRel, err := artifacts.RelativeWithinRoot(targetDir, relDest)
	if err != nil {
		return Outcome{Resource: r, Status: Failed, Err: err}
	}
	destAbs := filepath.Join(targetDir, destRel)

	if !overwrite {
		if _, statErr := os.Stat(destAbs); statErr == nil {
			return Outcome{Resource: r, Status: Skipped, Path: destRel}
		} else if !os.IsNotExist(statErr) {
			return Outcome{Resource: r, Status: Failed, Path: destRel, Err: statErr}
		}
	}

	content, err := fs.ReadFile(source, path.Join(sourceDir, r.Name))
	if err != nil {
		return Outcome{Resource: r, Status: Failed, Path: destRel, Err: err}
	}

	if err := WriteAtomicFile(destAbs, content); err != nil {
		return Outcome{Resource: r, Status: Failed, Path: destRel, Err: err}
	}

	return Outcome{Resource: r, Status: Installed, Path: destRel}
}
