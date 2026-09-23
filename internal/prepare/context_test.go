package prepare

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/validation"
)

func specID(n int) ids.EntityID {
	return ids.EntityID{Type: ids.Spec, Prefix: "SPEC", Number: n, Width: 3}
}

// TestLookupRequirementTexts_ReturnsSectionContentAndFingerprint covers
// data-model.md "RequirementText": each served Requirement's own
// section body, verbatim, plus a content fingerprint.
func TestLookupRequirementTexts_ReturnsSectionContentAndFingerprint(t *testing.T) {
	specBody := []byte("## Requirements\n\n### R1 — First\n\nRefresh tokens must rotate on every use.\n\n### R2 — Second\n\nOther requirement.\n")
	refs := []validation.RequirementRef{{Spec: specID(1), Number: 1}}

	got := LookupRequirementTexts(specBody, refs)
	if len(got) != 1 {
		t.Fatalf("got = %+v, want exactly 1 entry", got)
	}
	if got[0].Ref != refs[0] {
		t.Errorf("Ref = %+v, want %+v", got[0].Ref, refs[0])
	}
	if got[0].Content == "" {
		t.Errorf("Content is empty, want R1's own section body")
	}
	if got[0].Fingerprint == "" {
		t.Errorf("Fingerprint is empty, want a computed digest")
	}
}

// TestAssociatePlanSections_MatchesBareAndCompositeForms covers
// data-model.md "PlanSectionAssociation" (research.md Decision 4): a
// Plan section mentioning a served Requirement, in either the bare
// "R<N>" or composite "SPEC-###:R#" form, is associated.
func TestAssociatePlanSections_MatchesBareAndCompositeForms(t *testing.T) {
	planBody := []byte("## Requirement Coverage\n\nCovers R1 directly.\n\n## Unrelated Section\n\nNothing to do with any requirement.\n\n## Composite Mention\n\nSee SPEC-001:R1 for details.\n")
	served := []validation.RequirementRef{{Spec: specID(1), Number: 1}}

	got := AssociatePlanSections(planBody, served)

	headings := map[string]bool{}
	for _, s := range got {
		headings[s.Heading] = true
	}
	if !headings["Requirement Coverage"] {
		t.Errorf("associations = %+v, want \"Requirement Coverage\" (bare R1 mention)", got)
	}
	if !headings["Composite Mention"] {
		t.Errorf("associations = %+v, want \"Composite Mention\" (SPEC-001:R1 mention)", got)
	}
	if headings["Unrelated Section"] {
		t.Errorf("associations = %+v, want \"Unrelated Section\" excluded", got)
	}
	for _, s := range got {
		if len(s.MatchedRequirements) == 0 {
			t.Errorf("section %q has no MatchedRequirements, want at least one", s.Heading)
		}
	}
}
