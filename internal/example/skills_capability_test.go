// This file implements 039-lean-skills-integration-contracts's User
// Story 2 machine checks, extending skills_content_test.go's existing
// per-Skill conformance pass (assertSkillConformant) with two more:
//
//   - assertExampleSyntax (FR-005): every `internal ...` invocation
//     example quoted in a Skill's body is parsed against the real
//     *cobra.Command tree (internal/cli.NewRootCmd()) — an example
//     naming a nonexistent subcommand or flag fails the test.
//   - assertVerificationCapabilities (FR-006): every verification/
//     validation promise a Skill's own text makes (skillVerificationAllowlist)
//     is backed by exactly one real capability — an internal/validation
//     Code* constant, or a knownInternalCommands entry.
package example

import (
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/cli"
	"github.com/mottamarcio/misterspec/internal/validation"
)

// errorer is the minimal subset of *testing.T assertExampleSyntax and
// assertVerificationCapabilities need. *testing.T satisfies it
// directly; fakeErrorer (below) lets the "rejects bad input" tests
// observe a failure being reported without marking the real test run
// as failed, which passing a real *testing.T subtest into t.Run would
// otherwise unavoidably do.
type errorer interface {
	Helper()
	Errorf(format string, args ...any)
}

// fakeErrorer records whether Errorf was called, without failing the
// enclosing *testing.T.
type fakeErrorer struct {
	called bool
}

func (f *fakeErrorer) Helper()                           {}
func (f *fakeErrorer) Errorf(format string, args ...any) { f.called = true }

// internalExampleRe matches a whole inline-code-formatted "internal
// ..." invocation example, e.g. "`internal context SPEC-### --intent
// planning`" — the same convention every Skill's Deterministic
// Operations/Procedure sections already use.
var internalExampleRe = regexp.MustCompile("`(internal [^`]*)`")

// assertExampleSyntax implements FR-005: every such example, split into
// tokens, must resolve via the real command tree's Find(), and every
// "--flag"-shaped token among the remaining args must be a real flag
// on the resolved command or one of its ancestors.
func assertExampleSyntax(t errorer, name, content string) {
	t.Helper()

	for _, m := range internalExampleRe.FindAllStringSubmatch(content, -1) {
		example := m[1]
		tokens := strings.Fields(example)
		if len(tokens) == 0 {
			continue
		}

		root := cli.NewRootCmd()
		resolved, remaining, err := root.Find(tokens)
		if err != nil {
			t.Errorf("%s: example `%s` does not resolve against the real command tree: %v", name, example, err)
			continue
		}
		if resolved.HasSubCommands() {
			t.Errorf("%s: example `%s` does not name a real leaf command (stopped at `%s`, which is not a real registered command)", name, example, resolved.CommandPath())
			continue
		}

		for _, tok := range remaining {
			if !strings.HasPrefix(tok, "--") {
				continue
			}
			flagName := strings.TrimPrefix(tok, "--")
			if idx := strings.Index(flagName, "="); idx != -1 {
				flagName = flagName[:idx]
			}
			if !flagExistsOnChain(resolved, flagName) {
				t.Errorf("%s: example `%s` uses flag --%s, which is not registered on `%s` or any of its ancestors", name, example, flagName, resolved.CommandPath())
			}
		}
	}
}

// flagExistsOnChain checks cmd's own local and persistent flags, then
// walks up through cmd.Parent() to the root — a flag defined higher up
// is still real (e.g. a shared PersistentFlags on a parent command).
func flagExistsOnChain(cmd *cobra.Command, flagName string) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Flags().Lookup(flagName) != nil || c.PersistentFlags().Lookup(flagName) != nil {
			return true
		}
	}
	return false
}

// CapabilityClaim maps one Skill's textual verification/validation
// promise to the real capability backing it
// (data-model.md "CapabilityClaim"). Exactly one of BackingCode/
// BackingOp must be set.
type CapabilityClaim struct {
	Claim       string
	BackingCode string
	BackingOp   string
}

// validCodes is every internal/validation Code* constant that actually
// exists — a BackingCode not in this set is itself a test failure
// (guards the allowlist's own data against a typo or a renamed/removed
// code).
var validCodes = map[string]bool{
	validation.CodeNotFound:                      true,
	validation.CodeDuplicateID:                   true,
	validation.CodeMissingParent:                 true,
	validation.CodeInvalidParentType:             true,
	validation.CodeInvalidStatus:                 true,
	validation.CodeIDLocationMismatch:            true,
	validation.CodeUnresolvedDependency:          true,
	validation.CodeFrontmatterMalformed:          true,
	validation.CodeRequiredFieldMissing:          true,
	validation.CodeInvalidWikilink:               true,
	validation.CodeBrokenWikilink:                true,
	validation.CodeAmbiguousWikilink:             true,
	validation.CodeDuplicateRequirementID:        true,
	validation.CodeUncoveredRequirement:          true,
	validation.CodeUnknownRequirementReference:   true,
	validation.CodeCrossSpecRequirementReference: true,
	validation.CodeTaskWithoutRequirement:        true,
	validation.CodeDependencyCycle:               true,
	validation.CodePhaseGateBlocked:              true,
	validation.CodeTaskDependencyCycle:           true,
	validation.CodeInvalidTaskDependency:         true,
}

// skillVerificationAllowlist maps each Skill's own quoted verification
// promises (drawn from its "Validation Rules" section — quoted in the
// comment beside each entry) to the real capability backing it. Only
// specific, checkable promises are listed here; a Skill's own generic
// "internal validate SPEC-### reports zero new findings" restatement is
// backed by the "validate" entry every such Skill gets, since that is
// itself the literal, real capability the sentence names.
var skillVerificationAllowlist = map[string][]CapabilityClaim{
	"mister-features": {
		// "must resolve `internal parent FEAT-###` back to the correct Program"
		{Claim: "parent resolves to the correct Program", BackingOp: "parent"},
		{Claim: "created Feature passes internal validate with zero findings", BackingOp: "validate"},
	},
	"mister-specify": {
		// "must resolve `internal parent SPEC-###` back to the correct Feature"
		{Claim: "parent resolves to the correct Feature", BackingOp: "parent"},
		// "every `depends_on` entry must name a Spec that actually exists"
		{Claim: "depends_on entries name existing Specs", BackingCode: validation.CodeUnresolvedDependency},
		{Claim: "created Spec passes internal validate with zero findings", BackingOp: "validate"},
	},
	"mister-program": {
		// "Every required section must be present and non-empty"
		{Claim: "required sections present and non-empty", BackingCode: validation.CodeRequiredFieldMissing},
		{Claim: "created Program passes internal validate with zero findings", BackingOp: "validate"},
	},
	"mister-constitution": {
		// "must open with the required frontmatter (`type: constitution`, `schema_version`)"
		{Claim: "required frontmatter fields present", BackingCode: validation.CodeFrontmatterMalformed},
	},
	"mister-knowledge-base": {
		// "Every `sources` entry must name a real, fingerprinted raw file"
		{Claim: "sources entries name a real, fingerprinted raw file", BackingOp: "fingerprint"},
		{Claim: "Knowledge artifact passes internal validate with zero new findings", BackingOp: "validate"},
	},
	"mister-plan": {
		{Claim: "Plan passes internal validate with zero new findings", BackingOp: "validate"},
	},
	"mister-tasks": {
		// "no duplicate Task IDs, no Task referencing a nonexistent requirement"
		{Claim: "no duplicate Task IDs", BackingCode: validation.CodeDuplicateID},
		{Claim: "no Task referencing a nonexistent requirement", BackingCode: validation.CodeUnknownRequirementReference},
		// "no invalid or cyclic `Depends on:` declaration"
		{Claim: "no cyclic Depends on: declaration", BackingCode: validation.CodeTaskDependencyCycle},
		{Claim: "no invalid Depends on: declaration", BackingCode: validation.CodeInvalidTaskDependency},
	},
	"mister-implement": {
		{Claim: "project passes internal validate with zero new findings after a Task", BackingOp: "validate"},
	},
	"mister-analyze": {
		{Claim: "project structure passes internal validate", BackingOp: "validate"},
	},
	"mister-wrap-up": {
		// this Skill's own Validation Rules make no internal-validate promise
		// (its output is documentation, not a validated project artifact) —
		// no entry needed.
	},
}

// TestAssertExampleSyntax_RejectsBadExamples is 039's US2 Acceptance
// Scenario 3 / T022 regression coverage: a nonexistent subcommand and a
// nonexistent flag must each be caught, using a synthetic Skill body
// rather than breaking real kit/skills content.
func TestAssertExampleSyntax_RejectsBadExamples(t *testing.T) {
	cases := map[string]string{
		"nonexistent subcommand": "See `internal does-not-exist SPEC-###` for details.",
		"nonexistent flag":       "See `internal resolve SPEC-### --nonexistent-flag` for details.",
	}
	for label, content := range cases {
		fake := &fakeErrorer{}
		assertExampleSyntax(fake, "synthetic", content)
		if !fake.called {
			t.Errorf("assertExampleSyntax unexpectedly passed for case %q", label)
		}
	}
}

// TestAssertExampleSyntax_AcceptsRealExample confirms the check does
// not false-positive on a real, valid example.
func TestAssertExampleSyntax_AcceptsRealExample(t *testing.T) {
	fake := &fakeErrorer{}
	assertExampleSyntax(fake, "synthetic", "See `internal context SPEC-### --intent planning` for details.")
	if fake.called {
		t.Error("assertExampleSyntax unexpectedly failed for a real, valid example")
	}
}

// TestAssertVerificationCapabilities_RejectsUnbackedOrDualBackedClaims
// is 039's US2 Acceptance Scenario 2 / T024 regression coverage.
func TestAssertVerificationCapabilities_RejectsUnbackedOrDualBackedClaims(t *testing.T) {
	saved := skillVerificationAllowlist["synthetic-skill"]
	defer func() { skillVerificationAllowlist["synthetic-skill"] = saved }()

	cases := map[string][]CapabilityClaim{
		"neither backing set": {{Claim: "unbacked claim"}},
		"both backings set":   {{Claim: "dual-backed claim", BackingCode: validation.CodeNotFound, BackingOp: "validate"}},
		"unknown code":        {{Claim: "bad code", BackingCode: "not_a_real_code"}},
		"unknown op":          {{Claim: "bad op", BackingOp: "not-a-real-op"}},
	}
	for label, claims := range cases {
		skillVerificationAllowlist["synthetic-skill"] = claims
		fake := &fakeErrorer{}
		assertVerificationCapabilities(fake, "synthetic-skill", "")
		if !fake.called {
			t.Errorf("assertVerificationCapabilities unexpectedly passed for case %q", label)
		}
	}
}

// assertVerificationCapabilities implements FR-006: every claim listed
// for this Skill has exactly one of BackingCode/BackingOp set, and a
// BackingCode names a real validation.Code* constant.
func assertVerificationCapabilities(t errorer, name, content string) {
	t.Helper()

	for _, claim := range skillVerificationAllowlist[name] {
		hasCode := claim.BackingCode != ""
		hasOp := claim.BackingOp != ""
		if hasCode == hasOp {
			t.Errorf("%s: claim %q must have exactly one of BackingCode/BackingOp set (got code=%q op=%q)", name, claim.Claim, claim.BackingCode, claim.BackingOp)
			continue
		}
		if hasCode && !validCodes[claim.BackingCode] {
			t.Errorf("%s: claim %q names BackingCode %q, which is not a real validation.Code* constant", name, claim.Claim, claim.BackingCode)
		}
		if hasOp && !knownInternalCommands[claim.BackingOp] {
			t.Errorf("%s: claim %q names BackingOp %q, which is not a real registered command", name, claim.Claim, claim.BackingOp)
		}
	}
}
