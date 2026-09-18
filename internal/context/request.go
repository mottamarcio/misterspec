package contextengine

import (
	"errors"
	"fmt"
)

// Intent names what kind of work a Request's context is for. The zero
// value ("") means no particular intent (docs/context-engine-
// implementation.md §13, research.md #4).
type Intent string

// Recognized Intent values — kept verbatim from the source document's
// own illustrative constants (research.md #4).
const (
	IntentPlanning       Intent = "planning"
	IntentTasks          Intent = "tasks"
	IntentImplementation Intent = "implementation"
	IntentValidation     Intent = "validation"
	IntentAnalysis       Intent = "analysis"
)

// Request is one caller's context question (FR-001). No Budget field —
// deliberately deferred to Phase 7, which is the first feature to
// actually enforce one (research.md #2).
type Request struct {
	// Target is a raw entity ID, required — must resolve to one of the
	// five entity types operations.References already covers
	// (research.md #3).
	Target string
	// Task is the current task's own text, if any. Also used as Tier 4's
	// own search query when Query is empty (research.md #6).
	Task string
	// Intent is optional; "" means no particular intent.
	Intent Intent
	// Query is an optional free-text question.
	Query string
	// Budget optionally overrides DefaultBudget (FR-004, 016-ranking-
	// budgeting/research.md #6). nil means "not specified" —
	// DefaultBudget is used. A non-nil pointer, even to zero or a
	// negative number, is a real, explicit budget request — never
	// silently promoted to the default. This is the "soft limit" —
	// only ever shapes how much optional content is included; never
	// used to flag mandatory content as excessive
	// (035-context-budget-accuracy FR-005).
	Budget *int
	// HardLimit optionally overrides DefaultHardLimit — the ceiling
	// above which mandatory content is reported as exceeded (FR-005,
	// FR-007). nil means "not specified" — DefaultHardLimit is used,
	// never treated as "no limit" (FR-006). Independent of Budget: a
	// HardLimit smaller than the resolved Budget is accepted, not
	// rejected — it simply means a tight mandatory-content ceiling
	// alongside a generous optional-content target (data-model.md
	// Request validation rule).
	HardLimit *int
}

// recognizedIntents is the set validateIntent checks Intent against.
var recognizedIntents = map[Intent]bool{
	IntentPlanning:       true,
	IntentTasks:          true,
	IntentImplementation: true,
	IntentValidation:     true,
	IntentAnalysis:       true,
}

// ErrUnsupportedIntent is returned (wrapped) by validateIntent when a
// Request's Intent is neither "" nor one of the recognized values
// (017-internal-context-command/spec.md FR-002, research.md #4) — an
// errors.Is-matchable sentinel the internal "context" command's own
// classify function maps onto its own "unsupported_intent" JSON error
// code.
var ErrUnsupportedIntent = errors.New("contextengine: unsupported intent")

// validateIntent reports whether i is "" (no preference) or one of the
// recognized Intent values (FR-002).
func validateIntent(i Intent) error {
	if i == "" || recognizedIntents[i] {
		return nil
	}
	return fmt.Errorf("%w: %v", ErrUnsupportedIntent, i)
}
