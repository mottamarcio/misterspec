package skillgen

// KnownFragments is the fixed set of canonical fragments this feature
// introduces — the actual literal duplication found by
// specs/039-lean-skills-integration-contracts/research.md #1, not a
// speculative general-purpose library.
var KnownFragments = []Fragment{
	{
		Name:    "resolve-preamble",
		Section: "Deterministic Operations",
		Body:    "- `internal resolve SPEC-###` — confirm the Spec exists and locate its\n  canonical file.",
	},
	{
		Name:    "mechanical-steps-note",
		Section: "Deterministic Operations",
		Body:    "Use misterspec operations for every mechanical repository step.",
	},
	{
		// Body extends the sentence found identically today in
		// mister-plan/mister-tasks/mister-analyze's own "Deterministic
		// Operations" sections with an explicit instruction to record
		// the fallback occurrence (spec FR-008), rather than silently
		// absorbing the extra read — research.md #5.
		Name:    "fallback-tolerance-note",
		Section: "Deterministic Operations",
		Body:    "  Context Pack before broader exploration. If this fails, proceed\n  using this Skill's own Optional Context above instead — it is never\n  a Failure Condition. Record the occurrence as\n  `context_fallbacks` in this task's own RunRecord once one exists,\n  rather than silently absorbing the extra read\n  (037-eval-quality-efficiency).",
	},
}
