// Package architecture evaluates project-declared rules (forbidden
// dependency, layer boundary, required contract) against the current
// Go source tree via a per-language adapter, returning a strict
// pass/fail/not_evaluated Result per rule — never silently promoting
// an unsupported rule or an unsupported project language to "pass"
// (044-architecture-code-context-rules data-model.md "Result", spec
// FR-004/FR-005).
package architecture
