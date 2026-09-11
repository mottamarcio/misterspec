package ids

import "fmt"

// EntityType identifies which kind of ID-bearing entity an EntityID names.
// These are the entity types the misterspec architecture assigns an
// independent, allocatable ID to — see
// docs/architecture-specification.md §60-61. Plan, Tasks, Validation, and
// the Constitution have no independent ID of their own (they are addressed
// via the Spec/project they belong "for") and so have no EntityType value.
type EntityType int

const (
	// Program identifies a Program entity (PRG-*).
	Program EntityType = iota
	// Feature identifies a Feature entity (FEAT-*).
	Feature
	// Spec identifies a Spec entity (SPEC-*).
	Spec
	// Task identifies a Task entity (TASK-*).
	Task
	// Knowledge identifies a Knowledge entity (KNOW-*).
	Knowledge
	// Learning identifies a Learning entity (LRN-*).
	Learning
)

// prefixes maps each EntityType to its fixed ID prefix, per
// docs/architecture-specification.md §60-61.
var prefixes = map[EntityType]string{
	Program:   "PRG",
	Feature:   "FEAT",
	Spec:      "SPEC",
	Task:      "TASK",
	Knowledge: "KNOW",
	Learning:  "LRN",
}

// String returns the human-readable name of t (e.g. "spec"), or "unknown"
// for a value outside the defined EntityType range.
func (t EntityType) String() string {
	switch t {
	case Program:
		return "program"
	case Feature:
		return "feature"
	case Spec:
		return "spec"
	case Task:
		return "task"
	case Knowledge:
		return "knowledge"
	case Learning:
		return "learning"
	default:
		return "unknown"
	}
}

// Prefix returns t's fixed ID prefix (e.g. "SPEC"), or "" for a value
// outside the defined EntityType range.
func (t EntityType) Prefix() string {
	return prefixes[t]
}

// TypeForPrefix returns the EntityType whose fixed prefix matches prefix
// exactly, and true — or the zero EntityType and false if prefix is not
// one of the defined entity prefixes. It is the inverse of
// EntityType.Prefix, used when an ID string (e.g. "SPEC-014") is already
// in hand and its type must be recovered from its prefix alone.
func TypeForPrefix(prefix string) (EntityType, bool) {
	for t, p := range prefixes {
		if p == prefix {
			return t, true
		}
	}
	return 0, false
}

// EntityID is a typed, prefixed, zero-padded identifier that uniquely
// names one artifact within its EntityType (e.g. SPEC-014).
//
// The zero value is not a valid EntityID. An EntityID is only ever
// produced by Parse, which validates syntax — see ids.go. This type is
// declared here, separately from Parse, because both User Story 2
// (canonical path resolution) and User Story 3 (metadata parsing and ID
// discovery) need the type itself, while only User Story 3 needs the
// parsing/validation/scanning behavior built on top of it.
type EntityID struct {
	// Type is the entity type this ID belongs to.
	Type EntityType
	// Prefix is Type's fixed prefix, recorded alongside Number for
	// convenient formatting without re-deriving it from Type.
	Prefix string
	// Number is the numeric suffix. Always >= 1 for a valid EntityID.
	Number int
	// Width is the zero-padding width in effect when this EntityID was
	// parsed or constructed (from Configuration.IDWidth).
	Width int
}

// String renders id in its canonical textual form (e.g. "SPEC-014").
func (id EntityID) String() string {
	return fmt.Sprintf("%s-%0*d", id.Prefix, id.Width, id.Number)
}
