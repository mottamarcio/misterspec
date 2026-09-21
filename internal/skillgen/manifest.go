package skillgen

// Fragment is a named, canonical block of Skill instruction text that
// recurs verbatim across two or more canonical Skills
// (data-model.md "Fragment"). Body MUST NOT reference another
// Fragment's Name — composition is a single flat substitution pass.
type Fragment struct {
	Name    string
	Section string // one of the required §39 headings
	Body    string
}

// Part is exactly one of FragmentRef or Bespoke (data-model.md "Part").
// FragmentRef names a Fragment.Name to insert verbatim; Bespoke is
// Skill-specific Markdown authored only for this Skill.
type Part struct {
	FragmentRef string
	Bespoke     string
}

// SectionEntry is one §39 section's ordered composition
// (data-model.md "SectionEntry").
type SectionEntry struct {
	Heading string
	Parts   []Part
}

// SkillManifest is one canonical Skill's full composition
// (data-model.md "SkillManifest"). Frontmatter carries the exact
// "---\nname: ...\ndescription: ...\n---\n" block verbatim — not
// itself in data-model.md's field list, but required to reproduce a
// byte-identical SKILL.md file (research.md #2's drift-check
// invariant covers the whole file, not only the body).
type SkillManifest struct {
	SkillName   string
	Frontmatter string
	Sections    []SectionEntry
}

// requiredSkillHeadings mirrors internal/example/skills_content_test.go's
// own requiredSkillHeadings (docs/architecture-specification.md §39) —
// kept in sync manually since internal/example must not import
// internal/skillgen's test-only concerns, nor vice versa.
var requiredSkillHeadings = []string{
	"Purpose", "Invocation", "Responsibility", "Inputs", "Outputs", "Preconditions",
	"Required Context", "Optional Context", "Conditional Context", "Unnecessary Context",
	"Authority", "Allowed Reads", "Allowed Creates", "Allowed Modifications", "Forbidden Mutations",
	"Deterministic Operations", "Procedure", "Decision Rules", "Interaction Rules",
	"Validation Rules", "Failure Conditions", "Stop Conditions", "Success Criteria",
	"Postconditions", "Idempotency", "Resume Behavior",
	"Completion Contract", "Recommended Next Step", "Related Skills",
}

// validateHeadingOrder reports whether every required heading appears
// among headings, in order, as a subsequence — extra headings (e.g.
// mister-constitution's own "Quality Requirements") are tolerated,
// matching skills_content_test.go's existing assertSectionsPresentInOrder
// behavior (an Index-based subsequence check, not an exact-set check).
func validateHeadingOrder(headings []string) error {
	pos := 0
	for _, req := range requiredSkillHeadings {
		found := -1
		for i := pos; i < len(headings); i++ {
			if headings[i] == req {
				found = i
				break
			}
		}
		if found == -1 {
			return &GenerateError{Reason: "missing required heading " + quoteHeading(req) + " in order"}
		}
		pos = found + 1
	}
	return nil
}

func quoteHeading(h string) string {
	return "\"" + h + "\""
}
