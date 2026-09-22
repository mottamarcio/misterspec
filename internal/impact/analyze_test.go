package impact

import "testing"

// TestAnalyzeImpact_RequirementChangeInvalidatesCoveringTask is
// 042-impact-analysis-review T015 (US1, spec Acceptance Scenario 1): a
// changed Requirement's text is reported against the Task that serves
// it, with a reason naming the changed Requirement.
func TestAnalyzeImpact_RequirementChangeInvalidatesCoveringTask(t *testing.T) {
	dir := initFixtureRepo(t)
	cfg := fixtureConfig()
	specDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014"

	writeFixtureFile(t, dir, specDir+"/spec.md", specFrontmatter+"### R1 — First\n\nOriginal body.\n")
	writeFixtureFile(t, dir, specDir+"/tasks.md", "---\nfor: SPEC-014\ntype: tasks\n---\n## TASK-001 — Do the thing\n\n- [ ] Complete\n\nServes: SPEC-014:R1\n")
	from := commitFixture(t, dir, "base")

	writeFixtureFile(t, dir, specDir+"/spec.md", specFrontmatter+"### R1 — First\n\nCHANGED body.\n")
	to := commitFixture(t, dir, "change requirement")

	report, err := AnalyzeImpact(dir, cfg, AnalyzeImpactRequest{From: from, To: to})
	if err != nil {
		t.Fatalf("AnalyzeImpact() unexpected error: %v", err)
	}

	item := findItem(t, report.AffectedItems, "SPEC-014:TASK-001")
	if item.Classification != DeterministicInvalidation {
		t.Errorf("Classification = %q, want %q", item.Classification, DeterministicInvalidation)
	}
	if item.Reason == "" {
		t.Error("Reason is empty, want a specific explanation")
	}
	if item.ReverificationCandidate == nil || *item.ReverificationCandidate != "SPEC-014:TASK-001" {
		t.Errorf("ReverificationCandidate = %v, want SPEC-014:TASK-001", item.ReverificationCandidate)
	}
}

// TestAnalyzeImpact_DependsOnChangeInvalidatesDependentSpec is
// 042-impact-analysis-review T015 (US1, spec Acceptance Scenario 2): a
// changed Spec's dependent (via depends_on) is reported invalidated.
func TestAnalyzeImpact_DependsOnChangeInvalidatesDependentSpec(t *testing.T) {
	dir := initFixtureRepo(t)
	cfg := fixtureConfig()
	baseSpecDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014"
	dependentSpecDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-020"

	writeFixtureFile(t, dir, baseSpecDir+"/spec.md", specFrontmatter+"### R1 — First\n\nBody.\n")
	writeFixtureFile(t, dir, dependentSpecDir+"/spec.md", "---\nid: SPEC-020\ntype: spec\nstatus: draft\ndepends_on: [SPEC-014]\n---\n### R1 — Depends\n\nBody.\n")
	from := commitFixture(t, dir, "base")

	writeFixtureFile(t, dir, baseSpecDir+"/spec.md", specFrontmatter+"### R1 — First\n\nCHANGED.\n")
	to := commitFixture(t, dir, "change SPEC-014")

	report, err := AnalyzeImpact(dir, cfg, AnalyzeImpactRequest{From: from, To: to})
	if err != nil {
		t.Fatalf("AnalyzeImpact() unexpected error: %v", err)
	}

	item := findItem(t, report.AffectedItems, "SPEC-020")
	if item.Classification != DeterministicInvalidation {
		t.Errorf("Classification = %q, want %q", item.Classification, DeterministicInvalidation)
	}
}

// TestAnalyzeImpact_UnrelatedArtifactNotReported is 042-impact-
// analysis-review T015 (US1, spec Acceptance Scenario 3).
func TestAnalyzeImpact_UnrelatedArtifactNotReported(t *testing.T) {
	dir := initFixtureRepo(t)
	cfg := fixtureConfig()
	changedSpecDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014"
	unrelatedSpecDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-099"

	writeFixtureFile(t, dir, changedSpecDir+"/spec.md", specFrontmatter+"### R1 — First\n\nBody.\n")
	writeFixtureFile(t, dir, unrelatedSpecDir+"/spec.md", "---\nid: SPEC-099\ntype: spec\nstatus: draft\n---\n### R1 — Unrelated\n\nBody.\n")
	from := commitFixture(t, dir, "base")

	writeFixtureFile(t, dir, changedSpecDir+"/spec.md", specFrontmatter+"### R1 — First\n\nCHANGED.\n")
	to := commitFixture(t, dir, "change SPEC-014")

	report, err := AnalyzeImpact(dir, cfg, AnalyzeImpactRequest{From: from, To: to})
	if err != nil {
		t.Fatalf("AnalyzeImpact() unexpected error: %v", err)
	}
	for _, item := range report.AffectedItems {
		if item.ID == "SPEC-099" {
			t.Errorf("SPEC-099 unexpectedly reported affected: %+v", item)
		}
	}
}

// TestAnalyzeImpact_WikilinkOnlyIsSuggestedReview is 042-impact-
// analysis-review T021 (US2, spec Acceptance Scenario 1): an artifact
// that only wikilinks a changed artifact is reported
// suggested_review; a Task with a formal/coverage relation to a
// different changed element in the same run is still
// deterministic_invalidation (Scenario 2).
func TestAnalyzeImpact_WikilinkOnlyIsSuggestedReview(t *testing.T) {
	dir := initFixtureRepo(t)
	cfg := fixtureConfig()
	specDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014"
	knowDir := "ai/knowledge"

	writeFixtureFile(t, dir, specDir+"/spec.md", specFrontmatter+"### R1 — First\n\nOriginal.\n")
	writeFixtureFile(t, dir, specDir+"/tasks.md", "---\nfor: SPEC-014\ntype: tasks\n---\n## TASK-001 — Do the thing\n\n- [ ] Complete\n\nServes: SPEC-014:R1\n")
	writeFixtureFile(t, dir, knowDir+"/KNOW-001-mentions.md", "---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\nSee [[SPEC-014]].\n")
	from := commitFixture(t, dir, "base")

	writeFixtureFile(t, dir, specDir+"/spec.md", specFrontmatter+"### R1 — First\n\nCHANGED.\n")
	to := commitFixture(t, dir, "change requirement")

	report, err := AnalyzeImpact(dir, cfg, AnalyzeImpactRequest{From: from, To: to})
	if err != nil {
		t.Fatalf("AnalyzeImpact() unexpected error: %v", err)
	}

	know := findItem(t, report.AffectedItems, "KNOW-001")
	if know.Classification != SuggestedReview {
		t.Errorf("KNOW-001 Classification = %q, want %q", know.Classification, SuggestedReview)
	}

	task := findItem(t, report.AffectedItems, "SPEC-014:TASK-001")
	if task.Classification != DeterministicInvalidation {
		t.Errorf("SPEC-014:TASK-001 Classification = %q, want %q", task.Classification, DeterministicInvalidation)
	}
}

// TestAnalyzeImpact_CycleTerminates is 042-impact-analysis-review T024
// (US3, spec FR-006, Edge Case "ciclo de referências"): two artifacts
// wikilinking each other; changing one and analyzing terminates
// promptly with the other appearing at most once in AffectedItems.
func TestAnalyzeImpact_CycleTerminates(t *testing.T) {
	dir := initFixtureRepo(t)
	cfg := fixtureConfig()
	specDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014"
	knowDir := "ai/knowledge"

	writeFixtureFile(t, dir, specDir+"/spec.md", specFrontmatter+"### R1 — First\n\nOriginal. See [[KNOW-001]].\n")
	writeFixtureFile(t, dir, knowDir+"/KNOW-001-a.md", "---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\nSee [[SPEC-014]] back.\n")
	from := commitFixture(t, dir, "base with mutual wikilink cycle")

	writeFixtureFile(t, dir, specDir+"/spec.md", specFrontmatter+"### R1 — First\n\nCHANGED. See [[KNOW-001]].\n")
	to := commitFixture(t, dir, "change SPEC-014")

	done := make(chan struct{})
	var report ImpactReport
	var err error
	go func() {
		report, err = AnalyzeImpact(dir, cfg, AnalyzeImpactRequest{From: from, To: to})
		close(done)
	}()
	select {
	case <-done:
	case <-timeoutAfterSeconds(5):
		t.Fatal("AnalyzeImpact() did not terminate within 5s")
	}
	if err != nil {
		t.Fatalf("AnalyzeImpact() unexpected error: %v", err)
	}

	count := 0
	for _, item := range report.AffectedItems {
		if item.ID == "KNOW-001" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("KNOW-001 appears %d times in AffectedItems, want exactly 1", count)
	}
}

// TestAnalyzeImpact_ConsolidatesMultipleChangesInOneRun is
// 042-impact-analysis-review T024 (US3, Edge Case "duas mudanças na
// mesma execução"): a ChangeSet with two simultaneously changed
// elements (a Requirement and its owning Spec's own depends_on
// backlink target) produces one consolidated AffectedItems list, with
// a single row per affected item even when reached from more than one
// origin.
func TestAnalyzeImpact_ConsolidatesMultipleChangesInOneRun(t *testing.T) {
	dir := initFixtureRepo(t)
	cfg := fixtureConfig()
	specDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014"
	dependentSpecDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-020"

	writeFixtureFile(t, dir, specDir+"/spec.md", specFrontmatter+"### R1 — First\n\nOriginal.\n")
	writeFixtureFile(t, dir, specDir+"/tasks.md", "---\nfor: SPEC-014\ntype: tasks\n---\n## TASK-001 — Do the thing\n\n- [ ] Complete\n\nServes: SPEC-014:R1\n")
	writeFixtureFile(t, dir, dependentSpecDir+"/spec.md", "---\nid: SPEC-020\ntype: spec\nstatus: draft\ndepends_on: [SPEC-014]\n---\n### R1 — Depends\n\nBody.\n")
	from := commitFixture(t, dir, "base")

	// Two simultaneous changes: SPEC-014's own R1 text, and SPEC-014
	// itself (which SPEC-020 depends_on) — both changes originate from
	// the same spec.md edit here, exercising the "more than one
	// ChangedElement reaches the same/different items in one run" path
	// without needing a second Spec to hand-edit separately.
	writeFixtureFile(t, dir, specDir+"/spec.md", specFrontmatter+"### R1 — First\n\nCHANGED.\n")
	to := commitFixture(t, dir, "change SPEC-014")

	report, err := AnalyzeImpact(dir, cfg, AnalyzeImpactRequest{From: from, To: to})
	if err != nil {
		t.Fatalf("AnalyzeImpact() unexpected error: %v", err)
	}

	seen := map[string]int{}
	for _, item := range report.AffectedItems {
		seen[item.ID]++
	}
	for id, count := range seen {
		if count != 1 {
			t.Errorf("AffectedItem %q appears %d times, want exactly 1 (consolidated row)", id, count)
		}
	}
	if seen["SPEC-014:TASK-001"] != 1 {
		t.Error("SPEC-014:TASK-001 missing from consolidated report")
	}
	if seen["SPEC-020"] != 1 {
		t.Error("SPEC-020 missing from consolidated report")
	}
}

// TestAnalyzeImpact_UnmappedCodePathDeclaredNotDropped is 042-impact-
// analysis-review T027 (Polish, spec FR-007, research.md #9): a diff
// that includes a .go source file alongside a Spec change reports
// UnmappedCodePaths >= 1 via the full AnalyzeImpact path, and still
// reports the Spec-side impact normally alongside it.
func TestAnalyzeImpact_UnmappedCodePathDeclaredNotDropped(t *testing.T) {
	dir := initFixtureRepo(t)
	cfg := fixtureConfig()
	specDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014"
	dependentSpecDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-020"

	writeFixtureFile(t, dir, specDir+"/spec.md", specFrontmatter+"### R1 — First\n\nOriginal.\n")
	writeFixtureFile(t, dir, dependentSpecDir+"/spec.md", "---\nid: SPEC-020\ntype: spec\nstatus: draft\ndepends_on: [SPEC-014]\n---\n### R1 — Depends\n\nBody.\n")
	writeFixtureFile(t, dir, "internal/foo/foo.go", "package foo\n")
	from := commitFixture(t, dir, "base")

	writeFixtureFile(t, dir, specDir+"/spec.md", specFrontmatter+"### R1 — First\n\nCHANGED.\n")
	writeFixtureFile(t, dir, "internal/foo/foo.go", "package foo\n\nfunc Bar() {}\n")
	to := commitFixture(t, dir, "change spec and code together")

	report, err := AnalyzeImpact(dir, cfg, AnalyzeImpactRequest{From: from, To: to})
	if err != nil {
		t.Fatalf("AnalyzeImpact() unexpected error: %v", err)
	}
	if report.ChangeSet.UnmappedCodePaths < 1 {
		t.Errorf("UnmappedCodePaths = %d, want >= 1", report.ChangeSet.UnmappedCodePaths)
	}

	dependent := findItem(t, report.AffectedItems, "SPEC-020")
	if dependent.Classification != DeterministicInvalidation {
		t.Errorf("SPEC-020 Classification = %q, want %q — the Spec-side impact must still be reported alongside the declared code gap", dependent.Classification, DeterministicInvalidation)
	}
}

// TestAnalyzeImpact_NotARepository is 042-impact-analysis-review T016.
func TestAnalyzeImpact_NotARepository(t *testing.T) {
	dir := t.TempDir()
	if _, err := AnalyzeImpact(dir, fixtureConfig(), AnalyzeImpactRequest{From: "HEAD"}); err == nil {
		t.Error("AnalyzeImpact() error = nil outside a Git repository, want non-nil")
	}
}

func findItem(t *testing.T, items []AffectedItem, id string) AffectedItem {
	t.Helper()
	for _, item := range items {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("no AffectedItem with ID %q found in %+v", id, items)
	return AffectedItem{}
}
