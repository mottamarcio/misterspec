// Package ids implements misterspec's entity ID vocabulary: typed,
// prefixed, zero-padded identifiers (e.g. SPEC-014) for each supported
// entity type, syntax validation for them, project-tree scanning to
// discover existing IDs of a type, pure "next available ID" computation,
// and width-tolerant target resolution (ResolveTarget) for a raw ID
// string of unknown type/width — shared by internal/validation's
// wikilink classification and internal/operations's reference/backlink
// graph (012-references-backlinks).
//
// Per the project constitution (Principle III, "Filesystem Is the Single
// Source of Truth"), the next ID for a type is always derived by scanning
// existing artifacts — this package never reads or writes a persisted
// counter.
//
// See specs/001-core-foundation/contracts/packages.md for the package's
// exported contract and specs/001-core-foundation/data-model.md for the
// Entity ID entity definition.
package ids
