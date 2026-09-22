package validation

// Severity classifies a Finding. Only SeverityError is emitted by this
// feature; the type exists so a later feature can add warnings without a
// breaking change to Finding's shape.
type Severity int

// SeverityError is the only severity this feature's checks produce.
const SeverityError Severity = iota

// Finding is one detected structural problem (FR-005). A Finding is only
// ever constructed by the checks in this package — there is no way to
// build an unvalidated one from outside it.
type Finding struct {
	// Code is a stable, machine-matchable identifier (e.g.
	// "missing_parent", "duplicate_id", "invalid_status") — chosen to
	// match docs/architecture-specification.md §7's reserved JSON error
	// codes where one already exists (see
	// specs/004-structural-validation/contracts/validation.md).
	Code string
	// Severity is always SeverityError today.
	Severity Severity
	// Path is the affected artifact's path, relative to the project
	// root.
	Path string
	// Message is a specific, human-readable description — never a
	// generic "something is wrong."
	Message string
}

// Finding codes emitted by this package.
const (
	CodeNotFound             = "not_found"
	CodeDuplicateID          = "duplicate_id"
	CodeMissingParent        = "missing_parent"
	CodeInvalidParentType    = "invalid_parent_type"
	CodeInvalidStatus        = "invalid_status"
	CodeIDLocationMismatch   = "id_location_mismatch"
	CodeUnresolvedDependency = "unresolved_dependency"
	CodeFrontmatterMalformed = "frontmatter_malformed"
	CodeRequiredFieldMissing = "required_field_missing"
	// CodeInvalidWikilink marks a wikilink whose target token is not even
	// syntactically a recognizable entity ID (011-wikilink-foundation).
	CodeInvalidWikilink = "invalid_wikilink"
	// CodeBrokenWikilink marks a wikilink whose target is syntactically
	// valid but resolves to no existing artifact.
	CodeBrokenWikilink = "broken_wikilink"
	// CodeAmbiguousWikilink marks a wikilink whose target resolves to
	// more than one existing artifact.
	CodeAmbiguousWikilink = "ambiguous_wikilink"
	// CodeDuplicateRequirementID marks two "### R<N>" headings with the
	// same number inside one Spec's own body
	// (032-requirement-coverage-dependency-validation FR-002).
	CodeDuplicateRequirementID = "duplicate_requirement_id"
	// CodeUncoveredRequirement marks a declared Requirement with zero
	// same-Spec Serves: reference anywhere in that Spec's own tasks.md
	// (FR-007).
	CodeUncoveredRequirement = "uncovered_requirement"
	// CodeUnknownRequirementReference marks a Serves: reference whose
	// Requirement number does not exist in the named Spec, or whose
	// syntax is malformed (FR-004).
	CodeUnknownRequirementReference = "unknown_requirement_reference"
	// CodeCrossSpecRequirementReference marks a Serves: reference that
	// names a Spec other than the Task's own owning Spec — never valid
	// coverage, even when the referenced requirement exists (FR-005).
	CodeCrossSpecRequirementReference = "cross_spec_requirement_reference"
	// CodeTaskWithoutRequirement marks a Task with no Serves: line at
	// all (FR-006).
	CodeTaskWithoutRequirement = "task_without_requirement"
	// CodeDependencyCycle marks a cycle in the project-wide Spec
	// depends_on graph, including a Spec depending on itself (FR-009).
	CodeDependencyCycle = "dependency_cycle"
	// CodePhaseGateBlocked marks a Spec whose status is not "draft" and
	// which still has an uncovered requirement and/or a task without a
	// requirement — emitted in addition to, never instead of, the
	// underlying coverage Finding(s) it escalates (FR-011, FR-012).
	CodePhaseGateBlocked = "phase_gate_blocked"
	// CodeTaskDependencyCycle marks a cycle in one Spec's Task-to-Task
	// "Depends on:" graph, including a Task depending on itself
	// (034-task-oriented-context-preparation FR-008).
	CodeTaskDependencyCycle = "task_dependency_cycle"
	// CodeInvalidTaskDependency marks a "Depends on:" entry naming a
	// nonexistent Task number, or a Task in a different Spec (FR-009).
	CodeInvalidTaskDependency = "invalid_task_dependency"
	// CodeUnknownAnchor marks a wikilink whose Anchor is non-empty and
	// whose Target resolves to exactly one existing artifact, but that
	// artifact declares no Section with a matching Anchor
	// (040-stable-section-anchors data-model.md "Validation Codes") —
	// distinct from CodeBrokenWikilink, which means the target artifact
	// itself does not exist (spec FR-009).
	CodeUnknownAnchor = "unknown_anchor"
	// CodeDuplicateAnchor marks two or more Sections within the same
	// artifact declaring the same non-empty explicit anchor
	// (040-stable-section-anchors data-model.md "Validation Codes") —
	// raised once per artifact scan, independent of whether anything
	// currently references that anchor.
	CodeDuplicateAnchor = "duplicate_anchor"
	// CodeUnverifiedTask marks a checked Task with no Evidence-Result:
	// line at all — a checkbox alone is never treated as verification
	// (041-task-evidence-fingerprint data-model.md "Validation Codes",
	// spec FR-008/FR-009).
	CodeUnverifiedTask = "unverified_task"
	// CodeStaleTaskEvidence marks a checked Task whose recorded
	// Evidence-Fingerprint: no longer matches its own current content
	// (041-task-evidence-fingerprint data-model.md "Validation Codes",
	// spec FR-007).
	CodeStaleTaskEvidence = "stale_task_evidence"
	// CodeFailedTaskEvidence marks a checked Task whose most recent
	// Evidence-Result: is "fail" (041-task-evidence-fingerprint
	// data-model.md "Validation Codes", spec FR-009).
	CodeFailedTaskEvidence = "failed_task_evidence"
)
