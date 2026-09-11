package templates

import (
	"embed"
	"fmt"
	"reflect"
	"strings"
	"text/template"
)

//go:embed files/*.tmpl
var templateFS embed.FS

// Kind identifies which entity/artifact template to render.
type Kind int

const (
	// Program renders program.md (docs/architecture-specification.md §25).
	Program Kind = iota
	// Feature renders feature.md (§26).
	Feature
	// Spec renders spec.md (§27).
	Spec
	// Knowledge renders a KNOW-*.md file (§23).
	Knowledge
	// Learning renders a LRN-*.md file (§31).
	Learning
	// Plan renders plan.md (§28).
	Plan
	// Tasks renders tasks.md (§29).
	Tasks
	// Validation renders validation.md (§30).
	Validation
)

// ProgramData is the render data for Program.
type ProgramData struct{ ID string }

// FeatureData is the render data for Feature.
type FeatureData struct{ ID, Parent string }

// SpecData is the render data for Spec.
type SpecData struct{ ID, Parent string }

// KnowledgeData is the render data for Knowledge.
type KnowledgeData struct{ ID string }

// LearningData is the render data for Learning.
type LearningData struct{ ID string }

// PlanData is the render data for Plan.
type PlanData struct{ For string }

// TasksData is the render data for Tasks.
type TasksData struct{ For string }

// ValidationData is the render data for Validation.
type ValidationData struct{ For string }

var filenames = map[Kind]string{
	Program:    "program.md.tmpl",
	Feature:    "feature.md.tmpl",
	Spec:       "spec.md.tmpl",
	Knowledge:  "knowledge.md.tmpl",
	Learning:   "learning.md.tmpl",
	Plan:       "plan.md.tmpl",
	Tasks:      "tasks.md.tmpl",
	Validation: "validation.md.tmpl",
}

// expectedType maps each Kind to the concrete *Data type Render requires
// for it, so a mismatched call (e.g. Render(Program, SpecData{})) is
// rejected with a precise error rather than executing with the wrong
// fields silently missing.
var expectedType = map[Kind]reflect.Type{
	Program:    reflect.TypeOf(ProgramData{}),
	Feature:    reflect.TypeOf(FeatureData{}),
	Spec:       reflect.TypeOf(SpecData{}),
	Knowledge:  reflect.TypeOf(KnowledgeData{}),
	Learning:   reflect.TypeOf(LearningData{}),
	Plan:       reflect.TypeOf(PlanData{}),
	Tasks:      reflect.TypeOf(TasksData{}),
	Validation: reflect.TypeOf(ValidationData{}),
}

// Render renders kind's embedded template with data (the matching *Data
// struct above), returning the complete initial file content —
// frontmatter and body — per
// docs/architecture-specification.md §22-31's fixed schemas.
func Render(kind Kind, data any) (string, error) {
	want, ok := expectedType[kind]
	if !ok {
		return "", fmt.Errorf("templates: unknown kind %v", kind)
	}
	if got := reflect.TypeOf(data); got != want {
		return "", fmt.Errorf("templates: Render(%v, ...) expects %s, got %s", kind, want, got)
	}

	filename := filenames[kind]
	content, err := templateFS.ReadFile("files/" + filename)
	if err != nil {
		return "", fmt.Errorf("templates: reading %s: %w", filename, err)
	}

	tmpl, err := template.New(filename).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("templates: parsing %s: %w", filename, err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("templates: executing %s: %w", filename, err)
	}
	return buf.String(), nil
}
