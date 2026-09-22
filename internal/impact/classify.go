package impact

// RelationKind names the kind of relation a PropagationHop walks
// (data-model.md "PropagationHop"). "depends_on"/"parent"/"supersedes"
// mirror operations.ReferenceEntry/BacklinkEntry's own Relation values
// exactly (012-references-backlinks); "coverage" is a Requirement's
// reverse Serves: edge (032-requirement-coverage-dependency-
// validation); "evidence" marks a Task already non-Verified per its
// own 041-task-evidence-fingerprint state; "wikilink" mirrors
// operations.ReferenceEntry/BacklinkEntry's own semantic relation.
type RelationKind string

const (
	RelationDependsOn  RelationKind = "depends_on"
	RelationParent     RelationKind = "parent"
	RelationSupersedes RelationKind = "supersedes"
	RelationCoverage   RelationKind = "coverage"
	RelationEvidence   RelationKind = "evidence"
	RelationWikilink   RelationKind = "wikilink"
)

// Classification is an AffectedItem's own impact category (data-
// model.md "Classification", spec FR-003/FR-004).
type Classification string

const (
	DeterministicInvalidation Classification = "deterministic_invalidation"
	SuggestedReview           Classification = "suggested_review"
)

// Severity is an AffectedItem's own priority signal (data-model.md
// "Severity", spec FR-010).
type Severity string

const (
	SeverityHigh   Severity = "high"
	SeverityMedium Severity = "medium"
	SeverityLow    Severity = "low"
)

// classificationTable is the fixed, unconditional mapping from
// RelationKind to Classification (research.md #7). RelationWikilink
// maps to SuggestedReview and nothing in this package ever overrides
// that — the one guarantee spec FR-003/FR-004 require to hold
// structurally, not just by convention.
var classificationTable = map[RelationKind]Classification{
	RelationDependsOn:  DeterministicInvalidation,
	RelationParent:     DeterministicInvalidation,
	RelationSupersedes: DeterministicInvalidation,
	RelationCoverage:   DeterministicInvalidation,
	RelationEvidence:   DeterministicInvalidation,
	RelationWikilink:   SuggestedReview,
}

// severityTable is the fixed mapping from (Classification, RelationKind)
// to Severity (research.md #7): an already-broken Task's own evidence
// is the most concrete, actionable signal (SeverityHigh); a formal
// relation is high; Requirement coverage is medium (it names exactly
// which Task to re-verify, but the Task itself may still be fine); a
// suggestion is always low, regardless of relation — there is only one
// suggestion-producing relation (RelationWikilink) today, but the key
// is the full pair for clarity and future-proofing against a second
// one.
var severityTable = map[Classification]map[RelationKind]Severity{
	DeterministicInvalidation: {
		RelationEvidence:   SeverityHigh,
		RelationDependsOn:  SeverityHigh,
		RelationParent:     SeverityHigh,
		RelationSupersedes: SeverityHigh,
		RelationCoverage:   SeverityMedium,
	},
	SuggestedReview: {
		RelationWikilink: SeverityLow,
	},
}

// ClassifyRelation returns relation's fixed Classification (research.md
// #7). An unrecognized RelationKind (never produced by this package's
// own propagation walk) classifies as SuggestedReview — the safer of
// the two values, never silently promoted to an invalidation.
func ClassifyRelation(relation RelationKind) Classification {
	if c, ok := classificationTable[relation]; ok {
		return c
	}
	return SuggestedReview
}

// SeverityFor returns the fixed Severity for (classification, relation)
// (research.md #7). An unrecognized pair defaults to SeverityLow — the
// most conservative choice for an as-yet-unclassified combination.
func SeverityFor(classification Classification, relation RelationKind) Severity {
	if byRelation, ok := severityTable[classification]; ok {
		if s, ok := byRelation[relation]; ok {
			return s
		}
	}
	return SeverityLow
}
