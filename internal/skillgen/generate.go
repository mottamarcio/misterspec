package skillgen

//go:generate go run ./cmd/gen

import "fmt"

// GenerateError names the manifest's SkillName and the specific
// composition problem found (contracts/skillgen-and-eval-contract.md
// §1).
type GenerateError struct {
	SkillName string
	Reason    string
}

func (e *GenerateError) Error() string {
	if e.SkillName == "" {
		return e.Reason
	}
	return fmt.Sprintf("skillgen: %s: %s", e.SkillName, e.Reason)
}

// fragmentIndex builds a lookup by Fragment.Name, erroring on a
// duplicate name (data-model.md "Fragment" validation rule).
func fragmentIndex(fragments []Fragment) (map[string]Fragment, error) {
	idx := make(map[string]Fragment, len(fragments))
	for _, f := range fragments {
		if _, dup := idx[f.Name]; dup {
			return nil, &GenerateError{Reason: fmt.Sprintf("duplicate fragment name %q", f.Name)}
		}
		idx[f.Name] = f
	}
	return idx, nil
}

// Generate composes manifest against fragments into the exact bytes a
// SKILL.md file should contain: frontmatter followed by each
// SectionEntry's "## Heading" line and its concatenated Parts
// (contracts/skillgen-and-eval-contract.md §1). It returns an error
// naming manifest.SkillName and the specific problem when a
// Part.FragmentRef does not resolve, when a resolved Fragment.Section
// disagrees with the enclosing SectionEntry.Heading, or when Sections
// does not cover every required heading in order.
func Generate(manifest SkillManifest, fragments []Fragment) ([]byte, error) {
	idx, err := fragmentIndex(fragments)
	if err != nil {
		return nil, err
	}

	headings := make([]string, len(manifest.Sections))
	for i, s := range manifest.Sections {
		headings[i] = s.Heading
	}
	if err := validateHeadingOrder(headings); err != nil {
		if ge, ok := err.(*GenerateError); ok {
			ge.SkillName = manifest.SkillName
		}
		return nil, err
	}

	out := manifest.Frontmatter
	for _, section := range manifest.Sections {
		out += "\n## " + section.Heading + "\n"
		for _, part := range section.Parts {
			if part.FragmentRef != "" && part.Bespoke != "" {
				return nil, &GenerateError{SkillName: manifest.SkillName, Reason: fmt.Sprintf("section %q has a Part with both FragmentRef and Bespoke set", section.Heading)}
			}
			if part.FragmentRef != "" {
				frag, ok := idx[part.FragmentRef]
				if !ok {
					return nil, &GenerateError{SkillName: manifest.SkillName, Reason: fmt.Sprintf("section %q references unknown fragment %q", section.Heading, part.FragmentRef)}
				}
				if frag.Section != section.Heading {
					return nil, &GenerateError{SkillName: manifest.SkillName, Reason: fmt.Sprintf("fragment %q belongs under section %q, not %q", frag.Name, frag.Section, section.Heading)}
				}
				out += frag.Body
			} else {
				out += part.Bespoke
			}
		}
	}

	return []byte(out), nil
}
