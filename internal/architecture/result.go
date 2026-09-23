package architecture

// Result is one rule's own evaluation outcome (044-architecture-code-
// context-rules data-model.md "Result").
type Result struct {
	// RuleIndex is the rule's own position in the resolved
	// architecture_rules list — stable within one run.
	RuleIndex int
	// Status is "pass", "fail", or "not_evaluated" — never a fourth
	// value, and never conflated with each other (spec FR-004/FR-005).
	Status string
	// Path is the file where a fail was found, or the file/pattern a
	// not_evaluated applies to. Empty for a project-wide pass with
	// nothing more specific to name.
	Path string
	// Line is 1-indexed, file-absolute — the exact import statement's
	// own line for a fail on forbidden_dependency/layer_boundary. 0
	// when not applicable.
	Line int
	// Message is a specific, human-readable explanation — never
	// generic.
	Message string
	// Reason is present only for Status == "not_evaluated":
	// "no_adapter_for_project" or "rule_kind_unsupported".
	Reason string
}

// Status values (data-model.md "Result").
const (
	StatusPass         = "pass"
	StatusFail         = "fail"
	StatusNotEvaluated = "not_evaluated"
)

// not_evaluated reasons (data-model.md "Result").
const (
	ReasonNoAdapterForProject = "no_adapter_for_project"
	ReasonRuleKindUnsupported = "rule_kind_unsupported"
)

// Report is CheckArchitecture's own full result (data-model.md
// "ArchitectureCheckReport").
type Report struct {
	// Adapter is "go", or "" when no adapter applies to this project
	// at all.
	Adapter string
	Results []Result
}
