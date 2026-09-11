package templates_test

import (
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/templates"
)

func TestRender_Program(t *testing.T) {
	got, err := templates.Render(templates.Program, templates.ProgramData{ID: "PRG-001"})
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}
	want := "---\n" +
		"id: PRG-001\n" +
		"type: program\n" +
		"status: draft\n" +
		"---\n" +
		"\n" +
		"# PRG-001\n" +
		"\n" +
		"## Problem\n" +
		"\n" +
		"## Users and Stakeholders\n" +
		"\n" +
		"## Desired Outcome\n" +
		"\n" +
		"## Scope\n" +
		"\n" +
		"## Non-Goals\n" +
		"\n" +
		"## Constraints\n" +
		"\n" +
		"## Success Criteria\n" +
		"\n" +
		"## Relevant Knowledge\n" +
		"\n" +
		"## Open Questions\n"
	if got != want {
		t.Errorf("Render(Program) mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestRender_Feature(t *testing.T) {
	got, err := templates.Render(templates.Feature, templates.FeatureData{ID: "FEAT-001", Parent: "PRG-001"})
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}
	wantFrontmatter := "---\n" +
		"id: FEAT-001\n" +
		"type: feature\n" +
		"status: draft\n" +
		"parent: PRG-001\n" +
		"---\n"
	if !strings.HasPrefix(got, wantFrontmatter) {
		t.Errorf("Render(Feature) frontmatter mismatch:\n--- got ---\n%s\n--- want prefix ---\n%s", got, wantFrontmatter)
	}
	for _, heading := range []string{"## Capability", "## User Value", "## Scope", "## Non-Goals", "## Constraints", "## Relevant Knowledge", "## Open Questions"} {
		if !strings.Contains(got, heading) {
			t.Errorf("Render(Feature) missing heading %q", heading)
		}
	}
}

func TestRender_Spec(t *testing.T) {
	got, err := templates.Render(templates.Spec, templates.SpecData{ID: "SPEC-001", Parent: "FEAT-001"})
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}
	wantFrontmatter := "---\n" +
		"id: SPEC-001\n" +
		"type: spec\n" +
		"status: draft\n" +
		"parent: FEAT-001\n" +
		"depends_on: []\n" +
		"supersedes: []\n" +
		"---\n"
	if !strings.HasPrefix(got, wantFrontmatter) {
		t.Errorf("Render(Spec) frontmatter mismatch:\n--- got ---\n%s\n--- want prefix ---\n%s", got, wantFrontmatter)
	}
	for _, heading := range []string{"## Intent", "## Requirements", "### R1", "## Acceptance Scenarios", "## Edge Cases", "## Constraints", "## Non-Goals", "## Unresolved Questions", "## Sources"} {
		if !strings.Contains(got, heading) {
			t.Errorf("Render(Spec) missing heading %q", heading)
		}
	}
}

func TestRender_Knowledge(t *testing.T) {
	got, err := templates.Render(templates.Knowledge, templates.KnowledgeData{ID: "KNOW-001"})
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}
	wantFrontmatter := "---\n" +
		"id: KNOW-001\n" +
		"type: knowledge\n" +
		"status: active\n" +
		"sources: []\n" +
		"---\n"
	if !strings.HasPrefix(got, wantFrontmatter) {
		t.Errorf("Render(Knowledge) frontmatter mismatch:\n--- got ---\n%s\n--- want prefix ---\n%s", got, wantFrontmatter)
	}
	for _, heading := range []string{"## Summary", "## Known Facts", "## Constraints", "## Unknowns", "## Conflicts", "## Provenance"} {
		if !strings.Contains(got, heading) {
			t.Errorf("Render(Knowledge) missing heading %q", heading)
		}
	}
}

func TestRender_Learning(t *testing.T) {
	got, err := templates.Render(templates.Learning, templates.LearningData{ID: "LRN-001"})
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}
	wantFrontmatter := "---\n" +
		"id: LRN-001\n" +
		"type: learning\n" +
		"status: candidate\n" +
		"---\n"
	if !strings.HasPrefix(got, wantFrontmatter) {
		t.Errorf("Render(Learning) frontmatter mismatch:\n--- got ---\n%s\n--- want prefix ---\n%s", got, wantFrontmatter)
	}
	for _, heading := range []string{"## Observation", "## Evidence", "## Why It May Be Reusable", "## Potential Destination"} {
		if !strings.Contains(got, heading) {
			t.Errorf("Render(Learning) missing heading %q", heading)
		}
	}
}

func TestRender_Plan(t *testing.T) {
	got, err := templates.Render(templates.Plan, templates.PlanData{For: "SPEC-001"})
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}
	wantFrontmatter := "---\n" +
		"type: plan\n" +
		"for: SPEC-001\n" +
		"status: draft\n" +
		"---\n"
	if !strings.HasPrefix(got, wantFrontmatter) {
		t.Errorf("Render(Plan) frontmatter mismatch:\n--- got ---\n%s\n--- want prefix ---\n%s", got, wantFrontmatter)
	}
	for _, heading := range []string{"## Summary", "## Repository Context", "## Requirement Coverage", "## Architecture", "## Components Affected", "## Data Changes", "## API Changes", "## Integration Changes", "## Implementation Sequence", "## Test Strategy", "## Risks", "## Assumptions"} {
		if !strings.Contains(got, heading) {
			t.Errorf("Render(Plan) missing heading %q", heading)
		}
	}
}

func TestRender_Tasks(t *testing.T) {
	got, err := templates.Render(templates.Tasks, templates.TasksData{For: "SPEC-001"})
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}
	want := "---\n" +
		"type: tasks\n" +
		"for: SPEC-001\n" +
		"---\n" +
		"\n" +
		"# Tasks\n"
	if got != want {
		t.Errorf("Render(Tasks) mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestRender_Validation(t *testing.T) {
	got, err := templates.Render(templates.Validation, templates.ValidationData{For: "SPEC-001"})
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}
	wantFrontmatter := "---\n" +
		"type: validation\n" +
		"for: SPEC-001\n" +
		"result: pending\n" +
		"---\n"
	if !strings.HasPrefix(got, wantFrontmatter) {
		t.Errorf("Render(Validation) frontmatter mismatch:\n--- got ---\n%s\n--- want prefix ---\n%s", got, wantFrontmatter)
	}
	for _, heading := range []string{"## Summary", "## Requirement Validation", "## Unplanned Implementation", "## Findings", "## Recommended Corrections"} {
		if !strings.Contains(got, heading) {
			t.Errorf("Render(Validation) missing heading %q", heading)
		}
	}
}

func TestRender_MismatchedDataType(t *testing.T) {
	_, err := templates.Render(templates.Program, templates.SpecData{ID: "SPEC-001"})
	if err == nil {
		t.Fatal("Render() with mismatched data type expected an error, got nil")
	}
}
