package kit

import "embed"

// TemplatesFS embeds every artifact template
// (docs/architecture-specification.md §22-31's schemas). It is the
// single source of truth for template content: internal/templates
// (rendering, for internal/operations's Create/CreateArtifact) and
// internal/installer (raw materialization) both read from this same
// embed.FS — see specs/005-embedded-kit/research.md.
//
//go:embed templates/*.tmpl
var TemplatesFS embed.FS
