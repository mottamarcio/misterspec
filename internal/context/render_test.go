package contextengine

import (
	"strings"
	"testing"
)

func TestRender_ContainsEveryItemsOwnPathHeadingAndContent(t *testing.T) {
	result := Result{
		Items: []ResultItem{
			{ScoredCandidate: ScoredCandidate{Candidate: Candidate{
				Path: "ai/memory/constitution.md", Heading: "Principles", Content: "Filesystem is the source of truth.",
				Reasons: []Reason{{Tier: TierMandatory, Relation: "constitution"}},
			}}, Tokens: 10},
			{ScoredCandidate: ScoredCandidate{Candidate: Candidate{
				Path: "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-011/spec.md", Heading: "Intent", Content: "The dependency target.",
				Reasons: []Reason{{Tier: TierStructural, Relation: "depends_on"}},
			}}, Tokens: 8},
		},
	}

	out := Render(Request{Target: "SPEC-014", Intent: IntentImplementation}, result)

	for _, want := range []string{
		"ai/memory/constitution.md", "Principles", "Filesystem is the source of truth.",
		"ai/programs/PRG-001/features/FEAT-001/specs/SPEC-011/spec.md", "Intent", "The dependency target.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Render() output missing %q:\n%s", want, out)
		}
	}
}

func TestRender_GroupsByTierInAscendingOrder(t *testing.T) {
	result := Result{
		Items: []ResultItem{
			{ScoredCandidate: ScoredCandidate{Candidate: Candidate{
				Path: "a.md", Reasons: []Reason{{Tier: TierMandatory, Relation: "target"}},
			}}},
			{ScoredCandidate: ScoredCandidate{Candidate: Candidate{
				Path: "b.md", Reasons: []Reason{{Tier: TierText, Relation: "text_match"}},
			}}},
		},
	}

	out := Render(Request{Target: "SPEC-014"}, result)

	mandatoryIdx := strings.Index(out, "## "+TierMandatory.String())
	textIdx := strings.Index(out, "## "+TierText.String())
	if mandatoryIdx < 0 || textIdx < 0 {
		t.Fatalf("Render() output missing tier headings:\n%s", out)
	}
	if mandatoryIdx >= textIdx {
		t.Errorf("Render() output has Mandatory heading after Text heading, want ascending Tier order:\n%s", out)
	}
}

func TestRender_NoReorderingWithinAGroup(t *testing.T) {
	result := Result{
		Items: []ResultItem{
			{ScoredCandidate: ScoredCandidate{Candidate: Candidate{
				Path: "first.md", Reasons: []Reason{{Tier: TierText, Relation: "text_match"}},
			}}},
			{ScoredCandidate: ScoredCandidate{Candidate: Candidate{
				Path: "second.md", Reasons: []Reason{{Tier: TierText, Relation: "text_match"}},
			}}},
		},
	}

	out := Render(Request{Target: "SPEC-014"}, result)

	firstIdx := strings.Index(out, "first.md")
	secondIdx := strings.Index(out, "second.md")
	if firstIdx < 0 || secondIdx < 0 || firstIdx >= secondIdx {
		t.Errorf("Render() output does not preserve Result.Items order within a Tier group:\n%s", out)
	}
}

func TestRender_EmptyResultDoesNotPanic(t *testing.T) {
	out := Render(Request{Target: "SPEC-014"}, Result{})
	if !strings.Contains(out, "SPEC-014") {
		t.Errorf("Render() of an empty Result should still name the target:\n%s", out)
	}
}
