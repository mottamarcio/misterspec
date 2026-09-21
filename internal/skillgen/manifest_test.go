package skillgen

import "testing"

func TestValidateHeadingOrder_ValidPasses(t *testing.T) {
	if err := validateHeadingOrder(requiredSkillHeadings); err != nil {
		t.Errorf("validateHeadingOrder(requiredSkillHeadings) = %v, want nil", err)
	}
}

func TestValidateHeadingOrder_ToleratesExtraHeading(t *testing.T) {
	withExtra := append([]string{}, requiredSkillHeadings[:5]...)
	withExtra = append(withExtra, "Quality Requirements")
	withExtra = append(withExtra, requiredSkillHeadings[5:]...)

	if err := validateHeadingOrder(withExtra); err != nil {
		t.Errorf("validateHeadingOrder() with a tolerated extra heading = %v, want nil", err)
	}
}

func TestValidateHeadingOrder_MissingHeadingFails(t *testing.T) {
	missing := requiredSkillHeadings[:len(requiredSkillHeadings)-1] // drop "Related Skills"

	if err := validateHeadingOrder(missing); err == nil {
		t.Error("validateHeadingOrder() with a missing required heading = nil, want an error")
	}
}

func TestValidateHeadingOrder_OutOfOrderFails(t *testing.T) {
	outOfOrder := append([]string{}, requiredSkillHeadings...)
	outOfOrder[0], outOfOrder[1] = outOfOrder[1], outOfOrder[0]

	if err := validateHeadingOrder(outOfOrder); err == nil {
		t.Error("validateHeadingOrder() with swapped headings = nil, want an error")
	}
}

func TestFragmentIndex_DuplicateNameFails(t *testing.T) {
	_, err := fragmentIndex([]Fragment{
		{Name: "a", Section: "Purpose", Body: "x"},
		{Name: "a", Section: "Purpose", Body: "y"},
	})
	if err == nil {
		t.Error("fragmentIndex() with duplicate names = nil error, want an error")
	}
}

func TestKnownFragments_UniqueNames(t *testing.T) {
	if _, err := fragmentIndex(KnownFragments); err != nil {
		t.Errorf("fragmentIndex(KnownFragments) = %v, want nil (names must be unique)", err)
	}
}

func TestManifests_AllGenerateWithoutError(t *testing.T) {
	for _, m := range Manifests {
		if _, err := Generate(m, KnownFragments); err != nil {
			t.Errorf("Generate(%s) unexpected error: %v", m.SkillName, err)
		}
	}
}
