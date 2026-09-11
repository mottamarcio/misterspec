// Package artifacts implements canonical path resolution, artifact-type
// classification, and frontmatter metadata parsing for misterspec's
// Markdown-with-YAML-frontmatter artifacts (Program, Feature, Spec, Plan,
// Tasks, Validation, Knowledge, Learning, Constitution).
//
// It guarantees that a given (type, ID, parent) always resolves to exactly
// one canonical filesystem location, that no resolved or supplied path can
// ever escape the project root, and that a malformed or incomplete artifact
// produces a distinct, reportable error rather than a silent partial
// result.
//
// See specs/001-core-foundation/contracts/packages.md for the package's
// exported contract and specs/001-core-foundation/data-model.md for the
// Artifact/Metadata/Canonical Path entity definitions.
package artifacts
