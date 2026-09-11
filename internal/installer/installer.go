package installer

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/kit"
)

// Resource is one embedded item available for installation (FR-002).
type Resource struct {
	// Name is the embedded file's own name (e.g. "program.md.tmpl").
	Name string
	// Kind categorizes this resource — "template" for everything this
	// feature embeds. A string, not a closed enum, so a later feature
	// can add "skill"/"integration" kinds without a breaking change.
	Kind string
	// ArtifactType is, for a template resource, the entity/artifact type
	// it belongs to (e.g. "program"), derived from its filename.
	ArtifactType string
}

const templatesDir = "templates"

// List enumerates every resource in the embedded kit, derived directly
// from kit.TemplatesFS's own contents — never a separately maintained
// list that could drift from what's actually embedded (FR-002, FR-009).
// It performs no filesystem access outside the embedded FS itself, and
// returns an identical result on every call.
func List() []Resource {
	entries, err := fs.ReadDir(kit.TemplatesFS, templatesDir)
	if err != nil {
		// kit.TemplatesFS is compiled into the binary; a read failure
		// here means the embed itself is broken, not a runtime
		// condition callers can meaningfully recover from.
		panic("installer: reading embedded kit templates: " + err.Error())
	}

	resources := make([]Resource, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		resources = append(resources, Resource{
			Name:         entry.Name(),
			Kind:         "template",
			ArtifactType: artifactTypeFromFilename(entry.Name()),
		})
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
func Install(targetDir string, overwrite bool) ([]Outcome, error) {
	resources := List()
	outcomes := make([]Outcome, 0, len(resources))
	for _, r := range resources {
		outcomes = append(outcomes, installOne(targetDir, r, overwrite))
	}
	return outcomes, nil
}

// installOne installs a single resource, checking containment before any
// write (FR-006). Exported as a package-internal function so its
// containment guarantee can be unit-tested directly (containment_test.go)
// independent of whether List()'s current resource set can ever trigger
// it.
func installOne(targetDir string, r Resource, overwrite bool) Outcome {
	relDest := filepath.Join(templatesDir, r.Name)
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

	content, err := kit.TemplatesFS.ReadFile(templatesDir + "/" + r.Name)
	if err != nil {
		return Outcome{Resource: r, Status: Failed, Path: destRel, Err: err}
	}

	if err := writeAtomicFile(destAbs, content); err != nil {
		return Outcome{Resource: r, Status: Failed, Path: destRel, Err: err}
	}

	return Outcome{Resource: r, Status: Installed, Path: destRel}
}
