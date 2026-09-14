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
//
// It also exposes an artifact's Markdown body (ReadBody) and a pure
// lexical scan for author-written "[[TARGET]]" wikilinks within it
// (ExtractWikiLinks) — see specs/011-wikilink-foundation/contracts/
// wikilinks.md. Extraction never resolves a target against real project
// state; internal/validation is where that happens.
package artifacts
