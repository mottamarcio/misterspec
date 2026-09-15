// Package templates renders the fixed initial content — frontmatter and
// body — for each artifact/entity type misterspec's atomic creation layer
// writes, per docs/architecture-specification.md §22-31's schemas.
//
// Template content itself (the 8 .tmpl files for Program, Feature, Spec,
// Knowledge, Learning, Plan, Tasks, Validation) is no longer embedded
// here — it lives in package kit (kit.TemplatesFS), the shared embedded
// resource root, so template content has exactly one source of truth
// shared with internal/installer. This package only owns the rendering
// logic (Kind, the per-kind *Data structs, and Render). See
// specs/005-embedded-kit/research.md.
package templates
