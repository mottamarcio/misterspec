package validation

import "github.com/mottamarcio/misterspec/internal/ids"

// allowedStatesByType is the fixed lifecycle-state vocabulary per entity
// type, hardcoded exactly as frozen in
// docs/architecture-specification.md §25-31 (research.md: this is a
// frozen product fact, not configuration). Knowledge's only ever-named
// state in that document is "active" — treated as its sole allowed state
// for now (spec.md's Assumptions), extendable later without a breaking
// change if a real need appears.
var allowedStatesByType = map[ids.EntityType][]string{
	ids.Program:   {"draft", "active", "done", "cancelled"},
	ids.Feature:   {"draft", "active", "done", "cancelled"},
	ids.Spec:      {"draft", "ready", "in_progress", "validated", "blocked", "superseded", "cancelled"},
	ids.Learning:  {"candidate", "promoted", "dismissed"},
	ids.Knowledge: {"active"},
}

// allowedStates returns t's fixed set of allowed lifecycle status values,
// and whether t has a lifecycle-state concept at all.
func allowedStates(t ids.EntityType) ([]string, bool) {
	states, ok := allowedStatesByType[t]
	return states, ok
}

// requiredParentTypeByType is the fixed structural nesting rule — the
// logical inverse of internal/operations/children.go's unexported
// childTypesFor, duplicated deliberately rather than shared to avoid an
// import cycle (research.md).
var requiredParentTypeByType = map[ids.EntityType]ids.EntityType{
	ids.Feature: ids.Program,
	ids.Spec:    ids.Feature,
}

// requiredParentType returns the entity type t's parent must be, and
// whether t has a parent concept at all (Program/Knowledge/Learning do
// not).
func requiredParentType(t ids.EntityType) (ids.EntityType, bool) {
	parent, ok := requiredParentTypeByType[t]
	return parent, ok
}
