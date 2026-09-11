// Package templates renders the fixed initial content — frontmatter and
// body — for each artifact/entity type misterspec's atomic creation layer
// writes, per docs/architecture-specification.md §22-31's schemas.
//
// This package embeds exactly the 8 template files this feature needs
// (Program, Feature, Spec, Knowledge, Learning, Plan, Tasks, Validation)
// — it is not the broader "Embedded Kit" system (canonical Skills, agent
// integration templates) that a future misterspec init feature will
// build; see specs/003-entity-creation/research.md.
package templates
