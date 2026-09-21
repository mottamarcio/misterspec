// This file guards specs/039-lean-skills-integration-contracts's
// canonical-source invariant (US1): for every internal/skillgen
// SkillManifest, regenerating it against skillgen.KnownFragments MUST
// byte-equal the committed kit/skills/<name>/SKILL.md content. A
// mismatch means someone hand-edited a generated file without updating
// its fragment/manifest source and regenerating (or a fragment/manifest
// edit was made without regenerating) — the same golden-file discipline
// Constitution Principle V already requires for adapter/template output.
package example

import (
	"io/fs"
	"testing"

	"github.com/mottamarcio/misterspec/internal/skillgen"
	"github.com/mottamarcio/misterspec/kit"
)

func TestSkillgenDrift(t *testing.T) {
	if len(skillgen.Manifests) == 0 {
		t.Fatal("skillgen.Manifests is empty")
	}

	for _, m := range skillgen.Manifests {
		m := m
		t.Run(m.SkillName, func(t *testing.T) {
			got, err := skillgen.Generate(m, skillgen.KnownFragments)
			if err != nil {
				t.Fatalf("skillgen.Generate(%s) unexpected error: %v", m.SkillName, err)
			}

			want, err := fs.ReadFile(kit.SkillsFS, m.SkillName+"/SKILL.md")
			if err != nil {
				t.Fatalf("reading committed kit.SkillsFS %s/SKILL.md: %v", m.SkillName, err)
			}

			if string(got) != string(want) {
				t.Errorf("%s: skillgen.Generate() output does not byte-match committed kit/skills/%s/SKILL.md — regenerate via `go generate ./internal/skillgen/...` and commit the result", m.SkillName, m.SkillName)
			}
		})
	}
}
