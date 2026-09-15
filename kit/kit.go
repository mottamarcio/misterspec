package kit

import (
	"embed"
	"io/fs"
)

// TemplatesFS embeds every artifact template
// (docs/architecture-specification.md §22-31's schemas). It is the
// single source of truth for template content: internal/templates
// (rendering, for internal/operations's Create/CreateArtifact) and
// internal/installer (raw materialization) both read from this same
// embed.FS — see specs/005-embedded-kit/research.md.
//
//go:embed templates/*.tmpl
var TemplatesFS embed.FS

// skillsFS is the raw embed, its paths still "skills/..."-prefixed.
// SkillsFS (below) strips that prefix, matching what
// agents.InstallRequest.Skills / Adapter.Install already expect: a
// caller-supplied fs.FS whose root is the Skills content itself (the
// same shape a test's fstest.MapFS fixture has), per
// 006-agent-adapter's own convention.
//
//go:embed skills
var skillsFS embed.FS

// SkillsFS embeds the framework's canonical Skills for installation by
// an Adapter (docs/architecture-specification.md §33, §38-39). Its
// content is a single placeholder file today (kit/skills/README.md) —
// canonical Skill authoring is Phase 6's work, not
// 008-cli-cobra's — but the embed itself is real, not a fixture, so
// misterspec's public init command passes actual data
// (specs/008-cli-cobra/research.md).
var SkillsFS = mustSub(skillsFS, "skills")

// mustSub panics on error — skillsFS is compiled into the binary, so a
// failure here would mean the embed itself is broken, not a runtime
// condition a caller could meaningfully recover from (the same
// reasoning internal/installer.ListFS's own panic uses for a broken
// embedded read).
func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic("kit: " + err.Error())
	}
	return sub
}
