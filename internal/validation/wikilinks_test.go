package validation_test

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/testutil"
	"github.com/mottamarcio/misterspec/internal/validation"
)

// setupProgram writes the shared PRG-001/FEAT-001 scaffold every test in
// this file hangs its subject Spec off of, plus a real SPEC-002 target to
// link to when a test wants one.
func setupWikilinkProgram(t *testing.T, root string) {
	t.Helper()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md", "---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n# A Real Target\n")
}

func writeSubjectSpec(t *testing.T, root, body string) {
	t.Helper()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md", ""+
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n"+body)
}

func TestCheckWikilinks_ValidTargetProducesNoFinding(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupWikilinkProgram(t, root)
	writeSubjectSpec(t, root, "See [[SPEC-002]] for details.\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	for _, code := range []string{validation.CodeInvalidWikilink, validation.CodeBrokenWikilink, validation.CodeAmbiguousWikilink} {
		if containsCode(findings, code) {
			t.Errorf("findings = %v, did not want wikilink code %q for a valid, resolvable target", findingCodes(findings), code)
		}
	}
}

func TestCheckWikilinks_BrokenTargetProducesBrokenWikilink(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupWikilinkProgram(t, root)
	writeSubjectSpec(t, root, "See [[SPEC-999]] for details.\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeBrokenWikilink) {
		t.Errorf("findings = %v, want %q for a target with no matching artifact", findingCodes(findings), validation.CodeBrokenWikilink)
	}
	if containsCode(findings, validation.CodeInvalidWikilink) || containsCode(findings, validation.CodeAmbiguousWikilink) {
		t.Errorf("findings = %v, a broken target must not also classify as invalid or ambiguous", findingCodes(findings))
	}
}

func TestCheckWikilinks_MalformedTargetSyntaxProducesInvalidWikilink(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupWikilinkProgram(t, root)
	writeSubjectSpec(t, root, "See [[nodash]] for details.\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeInvalidWikilink) {
		t.Errorf("findings = %v, want %q for a target that is not even syntactically an entity ID", findingCodes(findings), validation.CodeInvalidWikilink)
	}
	if containsCode(findings, validation.CodeBrokenWikilink) || containsCode(findings, validation.CodeAmbiguousWikilink) {
		t.Errorf("findings = %v, a syntactically invalid target must not also classify as broken or ambiguous", findingCodes(findings))
	}
}

func TestCheckWikilinks_DifferentZeroPaddingWidthStillResolves(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupWikilinkProgram(t, root)
	// cfg.IDWidth is 3 (SPEC-002), but the link is written with width 1 —
	// exactly how a frontmatter field already tolerates this (parser.go's
	// parseFieldID via ids.ParseAny).
	writeSubjectSpec(t, root, "See [[SPEC-2]] for details.\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	for _, code := range []string{validation.CodeInvalidWikilink, validation.CodeBrokenWikilink, validation.CodeAmbiguousWikilink} {
		if containsCode(findings, code) {
			t.Errorf("findings = %v, a differently zero-padded but otherwise valid target must still resolve cleanly", findingCodes(findings))
		}
	}
}

func TestCheckWikilinks_TargetResolvingToMultipleArtifactsProducesAmbiguousWikilink(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupWikilinkProgram(t, root)
	// Two Knowledge artifacts both claiming KNOW-005 — a duplicate ID, so
	// a link to it resolves to more than one existing artifact.
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-005-a.md", "---\nid: KNOW-005\ntype: knowledge\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-005-b.md", "---\nid: KNOW-005\ntype: knowledge\nstatus: active\n---\n")
	writeSubjectSpec(t, root, "See [[KNOW-005]] for details.\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeAmbiguousWikilink) {
		t.Errorf("findings = %v, want %q for a target resolving to more than one artifact", findingCodes(findings), validation.CodeAmbiguousWikilink)
	}
	if containsCode(findings, validation.CodeInvalidWikilink) || containsCode(findings, validation.CodeBrokenWikilink) {
		t.Errorf("findings = %v, an ambiguous target must not also classify as invalid or broken", findingCodes(findings))
	}
}

func TestCheckWikilinks_AliasNeverAffectsResolution(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupWikilinkProgram(t, root)
	writeSubjectSpec(t, root, "See [[SPEC-999|First Alias]] and [[SPEC-999|Second Alias]].\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}

	var brokenCount int
	for _, f := range findings {
		if f.Code == validation.CodeBrokenWikilink {
			brokenCount++
		}
	}
	if brokenCount != 2 {
		t.Errorf("brokenCount = %d, want 2 — same target, different alias, must classify identically for each occurrence: %v", brokenCount, findingCodes(findings))
	}
}

// TestCheckWikilinks_NoLinksProducesNoWikilinkFindings directly proves
// spec.md's User Story 3 acceptance scenario: an artifact whose body
// contains zero links contributes zero wikilink-related findings — its
// validation output is exactly as it would have been before this feature
// existed.
// TestCheckAnchors_DuplicateAnchorInSameArtifactIsDetected proves spec
// 040 data-model.md "Validation Codes": two Sections in one artifact
// declaring the same explicit anchor produce a CodeDuplicateAnchor
// Finding, independent of whether anything references that anchor.
func TestCheckAnchors_DuplicateAnchorInSameArtifactIsDetected(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupWikilinkProgram(t, root)
	writeSubjectSpec(t, root, ""+
		"## Retry Policy {#retry-policy}\n\nFirst declaration.\n\n"+
		"## Another Retry Policy {#retry-policy}\n\nSecond declaration, same anchor.\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeDuplicateAnchor) {
		t.Errorf("findings = %v, want %q for two Sections declaring the same anchor", findingCodes(findings), validation.CodeDuplicateAnchor)
	}
}

// TestCheckAnchors_NoDuplicateAnchorsProducesNoFinding proves an
// artifact whose declared anchors are all distinct contributes no
// CodeDuplicateAnchor Finding.
func TestCheckAnchors_NoDuplicateAnchorsProducesNoFinding(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupWikilinkProgram(t, root)
	writeSubjectSpec(t, root, ""+
		"## Retry Policy {#retry-policy}\n\nDetails.\n\n"+
		"## Timeout Policy {#timeout-policy}\n\nOther details.\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if containsCode(findings, validation.CodeDuplicateAnchor) {
		t.Errorf("findings = %v, want no %q for distinct anchors", findingCodes(findings), validation.CodeDuplicateAnchor)
	}
}

// TestCheckAnchors_UnreferencedAnchorIsNotAnError proves spec 040's
// Assumptions: an anchor declared but never referenced by any wikilink
// is not itself a problem.
func TestCheckAnchors_UnreferencedAnchorIsNotAnError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupWikilinkProgram(t, root)
	writeSubjectSpec(t, root, "## Retry Policy {#retry-policy}\n\nNever referenced anywhere.\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("findings = %v, want none for a declared-but-unreferenced anchor", findingCodes(findings))
	}
}

// TestCheckWikilinks_UnknownAnchorInExistingTargetIsDetected proves
// spec 040 data-model.md "Validation Codes": a wikilink whose Anchor is
// non-empty, whose Target resolves to exactly one existing artifact,
// but which declares no Section with that Anchor, produces
// CodeUnknownAnchor.
func TestCheckWikilinks_UnknownAnchorInExistingTargetIsDetected(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupWikilinkProgram(t, root)
	writeSubjectSpec(t, root, "See [[SPEC-002#nonexistent]] for details.\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeUnknownAnchor) {
		t.Errorf("findings = %v, want %q for a target existing but declaring no such anchor", findingCodes(findings), validation.CodeUnknownAnchor)
	}
	if containsCode(findings, validation.CodeBrokenWikilink) {
		t.Errorf("findings = %v, an existing target with an unknown anchor must not also classify as broken", findingCodes(findings))
	}
}

// TestCheckWikilinks_UnknownAnchorAndBrokenWikilinkAreMutuallyExclusive
// proves spec 040 FR-009/Acceptance Scenario US3.2: a nonexistent
// target artifact (with or without an anchor) is CodeBrokenWikilink,
// never CodeUnknownAnchor — the two diagnoses stay distinguishable by
// which of the two facts actually failed.
func TestCheckWikilinks_UnknownAnchorAndBrokenWikilinkAreMutuallyExclusive(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupWikilinkProgram(t, root)
	writeSubjectSpec(t, root, "See [[SPEC-999#retry-policy]] for details.\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if !containsCode(findings, validation.CodeBrokenWikilink) {
		t.Errorf("findings = %v, want %q for a nonexistent target artifact even with an anchor", findingCodes(findings), validation.CodeBrokenWikilink)
	}
	if containsCode(findings, validation.CodeUnknownAnchor) {
		t.Errorf("findings = %v, a nonexistent target artifact must never also classify as unknown_anchor", findingCodes(findings))
	}
}

// TestCheckWikilinks_KnownAnchorProducesNoFinding proves the positive
// case: a wikilink whose anchor really is declared in the target
// produces neither CodeUnknownAnchor nor CodeBrokenWikilink.
func TestCheckWikilinks_KnownAnchorProducesNoFinding(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupWikilinkProgram(t, root)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Retry Policy {#retry-policy}\n\nDetails.\n")
	writeSubjectSpec(t, root, "See [[SPEC-002#retry-policy]] for details.\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if containsCode(findings, validation.CodeUnknownAnchor) || containsCode(findings, validation.CodeBrokenWikilink) {
		t.Errorf("findings = %v, want neither for a real, declared anchor", findingCodes(findings))
	}
}

// TestCheckWikilinks_KnownAnchorInKnowledgeTargetProducesNoFinding
// proves anchor resolution also works for a non-directory-scoped
// target type (Knowledge, a single file — unlike Spec/Feature/Program,
// which are directories canonicalFilename must resolve into).
func TestCheckWikilinks_KnownAnchorInKnowledgeTargetProducesNoFinding(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupWikilinkProgram(t, root)
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-x.md",
		"---\nid: KNOW-003\ntype: knowledge\nstatus: active\n---\n## Retry Policy {#retry-policy}\n\nDetails.\n")
	writeSubjectSpec(t, root, "See [[KNOW-003#retry-policy]] for details.\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if containsCode(findings, validation.CodeUnknownAnchor) || containsCode(findings, validation.CodeBrokenWikilink) {
		t.Errorf("findings = %v, want neither for a real, declared anchor in a Knowledge target", findingCodes(findings))
	}
}

func TestCheckWikilinks_NoLinksProducesNoWikilinkFindings(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupWikilinkProgram(t, root)
	writeSubjectSpec(t, root, "# Just prose.\n\nNo links anywhere in this body.\n")

	findings, err := validation.ValidateEntity(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("ValidateEntity() unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("findings = %v, want none — a well-formed Spec with a link-free body must be byte-for-byte clean", findingCodes(findings))
	}
}
