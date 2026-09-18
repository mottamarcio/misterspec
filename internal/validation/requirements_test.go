package validation

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

func reqTestConfig() project.Configuration {
	return project.Configuration{IDWidth: 3}
}

func specID(n int) ids.EntityID {
	return ids.EntityID{Type: ids.Spec, Prefix: "SPEC", Number: n, Width: 3}
}

func taskID(n int) ids.EntityID {
	return ids.EntityID{Type: ids.Task, Prefix: "TASK", Number: n, Width: 3}
}

// TestParseSpecRequirements_DistinctNumbers covers data-model.md
// "SpecRequirements": ordinary, distinct "### R<N>" headings.
func TestParseSpecRequirements_DistinctNumbers(t *testing.T) {
	body := []byte("# Requirements\n\n### R1 — First\n\nSome text.\n\n### R2 — Second\n\nMore text.\n")
	got := parseSpecRequirements(specID(1), body)
	if len(got.Numbers) != 2 || got.Numbers[0] != 1 || got.Numbers[1] != 2 {
		t.Fatalf("Numbers = %v, want [1 2]", got.Numbers)
	}
	if len(got.Duplicates) != 0 {
		t.Errorf("Duplicates = %v, want none", got.Duplicates)
	}
}

// TestParseSpecRequirements_DuplicateNumber covers spec.md FR-002.
func TestParseSpecRequirements_DuplicateNumber(t *testing.T) {
	body := []byte("### R1 — First\n\n### R1 — Duplicated by mistake\n")
	got := parseSpecRequirements(specID(1), body)
	if len(got.Duplicates) != 1 || got.Duplicates[0] != 1 {
		t.Fatalf("Duplicates = %v, want [1]", got.Duplicates)
	}
}

// TestParseSpecRequirements_AnyHeadingLevel covers data-model.md: any
// ATX heading level, not just "###", is matched.
func TestParseSpecRequirements_AnyHeadingLevel(t *testing.T) {
	body := []byte("## R3 — Second level\n\n#### R4 — Fourth level\n")
	got := parseSpecRequirements(specID(1), body)
	if len(got.Numbers) != 2 || got.Numbers[0] != 3 || got.Numbers[1] != 4 {
		t.Fatalf("Numbers = %v, want [3 4]", got.Numbers)
	}
}

// TestParseSpecRequirements_NonRequirementHeadingIgnored covers that a
// heading not shaped like "R<digits>" is not treated as a requirement.
func TestParseSpecRequirements_NonRequirementHeadingIgnored(t *testing.T) {
	body := []byte("### Retry policy\n\n### R1 — Real requirement\n\n### Rationale\n")
	got := parseSpecRequirements(specID(1), body)
	if len(got.Numbers) != 1 || got.Numbers[0] != 1 {
		t.Fatalf("Numbers = %v, want [1] — non-R headings must be ignored", got.Numbers)
	}
}

// TestParseRequirementRef_Valid covers data-model.md "RequirementRef".
func TestParseRequirementRef_Valid(t *testing.T) {
	cfg := reqTestConfig()
	got, err := parseRequirementRef("SPEC-014:R3", cfg)
	if err != nil {
		t.Fatalf("parseRequirementRef() unexpected error: %v", err)
	}
	want := RequirementRef{Spec: specID(14), Number: 3}
	if got != want {
		t.Errorf("parseRequirementRef() = %+v, want %+v", got, want)
	}
	if got.String() != "SPEC-014:R3" {
		t.Errorf("String() = %q, want %q", got.String(), "SPEC-014:R3")
	}
}

// TestParseRequirementRef_Invalid covers contracts §1's malformed-input
// table.
func TestParseRequirementRef_Invalid(t *testing.T) {
	cfg := reqTestConfig()
	tests := []struct {
		name string
		raw  string
	}{
		{"missing R prefix", "SPEC-014:3"},
		{"non-numeric requirement number", "SPEC-014:Rabc"},
		{"missing spec half", "R3"},
		{"missing separator", "SPEC-014R3"},
		{"invalid spec ID syntax", "SPEC-abc:R3"},
		{"wrong spec ID width", "SPEC-14:R3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseRequirementRef(tt.raw, cfg)
			if err == nil {
				t.Fatalf("parseRequirementRef(%q) expected an error, got nil", tt.raw)
			}
		})
	}
}

// TestParseTaskCoverage_SingleReference covers data-model.md
// "TaskCoverage" and research.md Decision 3.
func TestParseTaskCoverage_SingleReference(t *testing.T) {
	cfg := reqTestConfig()
	body := []byte("- [ ] Complete\n\nServes: SPEC-014:R1\n")
	got := ParseTaskCoverage(taskID(1), body, cfg)
	if len(got.References) != 1 || got.References[0] != (RequirementRef{Spec: specID(14), Number: 1}) {
		t.Fatalf("References = %+v, want [SPEC-014:R1]", got.References)
	}
	if len(got.MalformedReferences) != 0 {
		t.Errorf("MalformedReferences = %v, want none", got.MalformedReferences)
	}
}

// TestParseTaskCoverage_CommaSeparatedList covers research.md Decision 3.
func TestParseTaskCoverage_CommaSeparatedList(t *testing.T) {
	cfg := reqTestConfig()
	body := []byte("Serves: SPEC-014:R1, SPEC-014:R2\n")
	got := ParseTaskCoverage(taskID(1), body, cfg)
	if len(got.References) != 2 {
		t.Fatalf("References = %+v, want 2 entries", got.References)
	}
}

// TestParseTaskCoverage_RepeatedLinesUnioned covers research.md
// Decision 3: multiple Serves: lines in one Task body are unioned.
func TestParseTaskCoverage_RepeatedLinesUnioned(t *testing.T) {
	cfg := reqTestConfig()
	body := []byte("Serves: SPEC-014:R1\nServes: SPEC-014:R2\n")
	got := ParseTaskCoverage(taskID(1), body, cfg)
	if len(got.References) != 2 {
		t.Fatalf("References = %+v, want 2 entries from two Serves: lines", got.References)
	}
}

// TestParseTaskCoverage_MalformedEntryCaptured covers data-model.md:
// a malformed reference is captured, not silently dropped.
func TestParseTaskCoverage_MalformedEntryCaptured(t *testing.T) {
	cfg := reqTestConfig()
	body := []byte("Serves: not-a-valid-ref\n")
	got := ParseTaskCoverage(taskID(1), body, cfg)
	if len(got.References) != 0 {
		t.Errorf("References = %+v, want none for a malformed entry", got.References)
	}
	if len(got.MalformedReferences) != 1 || got.MalformedReferences[0] != "not-a-valid-ref" {
		t.Fatalf("MalformedReferences = %v, want [\"not-a-valid-ref\"]", got.MalformedReferences)
	}
}

// TestParseTaskCoverage_NoServesLine covers spec.md FR-006.
func TestParseTaskCoverage_NoServesLine(t *testing.T) {
	cfg := reqTestConfig()
	body := []byte("- [ ] Complete\n\nNo coverage line here.\n")
	got := ParseTaskCoverage(taskID(1), body, cfg)
	if len(got.References) != 0 {
		t.Errorf("References = %+v, want none", got.References)
	}
}
