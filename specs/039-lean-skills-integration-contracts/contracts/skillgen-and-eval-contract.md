# Contract: Skill Generation & Evaluation Metrics Extension

This documents the machine-readable/machine-checked contracts this
feature adds or changes (Constitution Principle IX). None of this is a
new `misterspec` end-user command — the frozen "Public command surface"
constraint is unaffected.

## 1. `internal/skillgen` (new package)

```go
package skillgen

// Fragment is a named, canonical block of Skill instruction text
// (data-model.md "Fragment"). Body MUST NOT reference another
// Fragment's Name — composition is a single flat pass.
type Fragment struct {
	Name    string
	Section string // one of the 26 §39 headings
	Body    string
}

// Part is exactly one of FragmentRef or Bespoke (data-model.md "Part").
type Part struct {
	FragmentRef string // Fragment.Name, or ""
	Bespoke     string // Skill-specific Markdown, or ""
}

// SectionEntry is one §39 section's ordered composition.
type SectionEntry struct {
	Heading string
	Parts   []Part
}

// SkillManifest is one canonical Skill's full composition
// (data-model.md "SkillManifest"). Sections MUST cover exactly the 26
// required headings, in the required order.
type SkillManifest struct {
	SkillName string
	Sections  []SectionEntry
}

// Generate composes manifest against fragments into the exact bytes a
// SKILL.md file should contain (frontmatter + composed body). Returns
// an error if a Part.FragmentRef does not resolve, resolves to a
// Fragment whose Section disagrees with the enclosing SectionEntry, or
// if Sections does not cover exactly the 26 required headings in order.
func Generate(manifest SkillManifest, fragments []Fragment) ([]byte, error)

// KnownFragments is the fixed set of canonical fragments this feature
// introduces (resolve-preamble, mechanical-steps-note,
// fallback-tolerance-note, …) — the actual duplication found by
// research.md #1, not a speculative general-purpose library.
var KnownFragments []Fragment

// Manifests is the fixed set of all 10 canonical Skills' manifests,
// one per kit/skills/<name>/ directory.
var Manifests []SkillManifest
```

**Drift-check contract** (`internal/example/skillgen_drift_test.go`):
for every `m := range skillgen.Manifests`, `skillgen.Generate(m,
skillgen.KnownFragments)` MUST byte-equal the committed content of
`kit/skills/<m.SkillName>/SKILL.md`. A mismatch fails the test naming
the Skill and a diff — the same shape of failure Constitution Principle
V already requires for golden adapter/template fixtures.

## 2. `internal/example/skills_content_test.go` (extended checks)

No new exported Go API — two additional in-test assertions per Skill,
run alongside the existing heading/`internal <op>`-allowlist checks:

- **Example-syntax check (FR-005)**: every fenced `misterspec …`
  invocation found in a Skill's body is parsed against the real
  `*cobra.Command` tree built by `internal/cli`'s existing constructor
  (`Find()` + flag lookup); an example naming a nonexistent subcommand
  or flag fails the test, naming the Skill and the offending example.
- **Capability-allowlist check (FR-006)**: a new
  `skillVerificationAllowlist map[string][]CapabilityClaim` (shape per
  data-model.md "CapabilityClaim") is checked so that every
  verification/validation promise a Skill's text makes has exactly one
  of `BackingCode` (an `internal/validation.Code*` constant) or
  `BackingOp` (an entry in the existing `knownInternalCommands`) set; an
  unbacked claim fails the test.
- **Size assertion (SC-001)**: total non-frontmatter line count across
  `kit/skills/*/SKILL.md` after this feature MUST be at most 80% of the
  recorded pre-feature baseline (2378 lines), asserted as a fixed
  constant in-test alongside a comment recording the baseline.

## 3. `internal/example/skill_smoke_test.go` (new)

No new exported Go API — a filesystem-integration test table, one row
per `SmokeTestScenario` (data-model.md), for the fixed representative
set `{claude-code, cursor-agent, copilot}`. Each row: installs the
(generated) canonical Skills into a temp project via that adapter's own
`Install()`, then invokes the real `misterspec internal
{resolve,context or prepare,validate}` chain `mister-implement`'s own
text documents, against a small fixture Spec/Task, asserting: (a) each
step exits 0 and returns the documented JSON envelope shape, (b) the
adapter's own `TargetPath()` is where the Skill content actually landed.
A failing row names the adapter and the failing step (spec FR-007).

## 4. `internal/eval` (extended)

```go
package eval

// Metrics gains one field (data-model.md "Metrics (extended)").
type Metrics struct {
	// ... existing fields unchanged ...
	ContextFallbacks int `json:"context_fallbacks"`
}

// Validate gains one check: ContextFallbacks MUST be >= 0. No
// *_estimated sibling is required (it is a plain count, like the
// existing Calls/ExtraReads/Rework fields).
func (m Metrics) Validate() error
```

No change to `RunRecord`, `TaskResult`, `Baseline`, or `compare.go`'s
exported shape — `compare.go`'s existing generic per-`Metrics`-field
diffing already covers the new field with no additional logic.
