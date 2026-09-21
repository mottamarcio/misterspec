package skillgen

import (
	"strings"
	"testing"
)

func TestGenerate_FragmentResolvesUnderMatchingSection(t *testing.T) {
	fragments := []Fragment{{Name: "greet", Section: "Purpose", Body: "Hello."}}
	manifest := minimalManifest("test-skill", "Purpose", []Part{{FragmentRef: "greet"}})

	got, err := Generate(manifest, fragments)
	if err != nil {
		t.Fatalf("Generate() unexpected error: %v", err)
	}
	if !strings.Contains(string(got), "## Purpose\nHello.") {
		t.Errorf("Generate() = %q, want it to contain %q", got, "## Purpose\nHello.")
	}
	if !strings.HasPrefix(string(got), manifest.Frontmatter) {
		t.Errorf("Generate() does not start with the manifest's own Frontmatter")
	}
}

func TestGenerate_UnresolvedFragmentRef(t *testing.T) {
	manifest := minimalManifest("test-skill", "Purpose", []Part{{FragmentRef: "does-not-exist"}})

	_, err := Generate(manifest, nil)
	if err == nil {
		t.Fatal("Generate() expected error for unresolved FragmentRef, got nil")
	}
	if ge, ok := err.(*GenerateError); !ok || ge.SkillName != "test-skill" {
		t.Errorf("Generate() error = %v, want *GenerateError naming skill %q", err, "test-skill")
	}
}

func TestGenerate_FragmentSectionMismatch(t *testing.T) {
	fragments := []Fragment{{Name: "greet", Section: "Invocation", Body: "Hello."}}
	manifest := minimalManifest("test-skill", "Purpose", []Part{{FragmentRef: "greet"}})

	_, err := Generate(manifest, fragments)
	if err == nil {
		t.Fatal("Generate() expected error for section mismatch, got nil")
	}
}

func TestGenerate_MissingRequiredHeading(t *testing.T) {
	manifest := SkillManifest{
		SkillName:   "test-skill",
		Frontmatter: "---\nname: test-skill\n---\n",
		Sections: []SectionEntry{
			{Heading: "Purpose", Parts: []Part{{Bespoke: "x"}}},
			// every other required heading omitted
		},
	}

	_, err := Generate(manifest, nil)
	if err == nil {
		t.Fatal("Generate() expected error for missing required headings, got nil")
	}
}

func TestGenerate_OutOfOrderRequiredHeading(t *testing.T) {
	manifest := SkillManifest{
		SkillName:   "test-skill",
		Frontmatter: "---\nname: test-skill\n---\n",
		Sections:    []SectionEntry{},
	}
	for i := len(requiredSkillHeadings) - 1; i >= 0; i-- {
		manifest.Sections = append(manifest.Sections, SectionEntry{
			Heading: requiredSkillHeadings[i],
			Parts:   []Part{{Bespoke: "x"}},
		})
	}

	_, err := Generate(manifest, nil)
	if err == nil {
		t.Fatal("Generate() expected error for out-of-order headings, got nil")
	}
}

func TestGenerate_DuplicateFragmentName(t *testing.T) {
	fragments := []Fragment{
		{Name: "greet", Section: "Purpose", Body: "Hi."},
		{Name: "greet", Section: "Purpose", Body: "Hello."},
	}
	manifest := minimalManifest("test-skill", "Purpose", []Part{{FragmentRef: "greet"}})

	_, err := Generate(manifest, fragments)
	if err == nil {
		t.Fatal("Generate() expected error for duplicate fragment name, got nil")
	}
}

// minimalManifest builds a SkillManifest covering every required
// heading with a trivial Bespoke body, except purposeHeading which
// gets purposeParts instead — enough to exercise Generate()'s
// fragment-resolution logic without repeating all 26 headings inline
// per test.
func minimalManifest(skillName, purposeHeading string, purposeParts []Part) SkillManifest {
	m := SkillManifest{
		SkillName:   skillName,
		Frontmatter: "---\nname: " + skillName + "\ndescription: test\n---\n",
	}
	for _, h := range requiredSkillHeadings {
		if h == purposeHeading {
			m.Sections = append(m.Sections, SectionEntry{Heading: h, Parts: purposeParts})
			continue
		}
		m.Sections = append(m.Sections, SectionEntry{Heading: h, Parts: []Part{{Bespoke: "x"}}})
	}
	return m
}
