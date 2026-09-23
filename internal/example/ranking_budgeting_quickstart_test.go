// This file compiles and runs specs/016-ranking-budgeting/
// quickstart.md's end-to-end flow (ranking a candidate set, fitting it
// into an explicit budget, using the default budget when none is
// specified, an oversized-mandatory scenario correctly flagged with
// its overage, and a determinism check) against internal/context, on
// top of the same fixture context_collector_quickstart_test.go (015)
// already builds and Collects from.
package example

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	contextengine "github.com/mottamarcio/misterspec/internal/context"
	"github.com/mottamarcio/misterspec/internal/context/index"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestRankingBudgetingQuickstart_EndToEnd(t *testing.T) {
	root := testutil.Project(t)
	proj, err := project.Detect(root)
	if err != nil {
		t.Fatalf("project.Detect() unexpected error: %v", err)
	}
	cfg := proj.Config

	testutil.WriteFile(t, root, "ai/memory/constitution.md",
		"---\ntype: constitution\n---\n## Principles\n\nFilesystem is the source of truth.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md",
		"---\nid: FEAT-004\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-011/spec.md",
		"---\nid: SPEC-011\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nRefresh token rotation details.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-auth.md",
		"---\nid: KNOW-003\ntype: knowledge\nstatus: active\n---\n## Summary\n\nAuthentication model facts.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on:\n  - SPEC-011\nsupersedes: []\n---\n## Intent\n\nSee [[KNOW-003]] for background.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-002-unrelated.md",
		"---\nid: KNOW-002\ntype: knowledge\nstatus: active\n---\n## Summary\n\nRefresh token rotation mentioned only in passing, unrelated to SPEC-014.\n")

	store, err := index.Open(filepath.Join(root, ".misterspec/cache/context.db"))
	if err != nil {
		t.Fatalf("index.Open() unexpected error: %v", err)
	}
	defer store.Close()
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	req := contextengine.Request{Target: "SPEC-014", Query: "refresh token rotation"}

	// 1. Rank a candidate set (quickstart.md §1): mandatory and
	// structural/semantic content always precede a purely textual
	// match, regardless of its own score.
	set, err := contextengine.Collect(root, cfg, store, req)
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}
	ranked := contextengine.Rank(set, req)
	if len(ranked) == 0 {
		t.Fatal("Rank() returned no candidates")
	}

	lastMandatoryOrHigher := -1
	firstText := -1
	for i, c := range ranked {
		tier := minTierOf(c.Reasons)
		if tier <= contextengine.TierSemantic {
			lastMandatoryOrHigher = i
		}
		if tier == contextengine.TierText && firstText == -1 {
			firstText = i
		}
	}
	if firstText != -1 && firstText < lastMandatoryOrHigher {
		t.Errorf("Rank() placed a text match (index %d) before mandatory/structural/semantic content (last at index %d)", firstText, lastMandatoryOrHigher)
	}

	// 2. Fit the ranked set into an explicit budget (quickstart.md §2).
	budget := 500
	result := contextengine.ApplyBudget(ranked, contextengine.Request{Target: "SPEC-014", Budget: &budget}, artifacts.DefaultEstimator{})
	if result.Diagnostics.TokensSelected > budget && !result.BudgetExceeded {
		t.Errorf("ApplyBudget() selected %d tokens over budget %d without flagging BudgetExceeded", result.Diagnostics.TokensSelected, budget)
	}
	if result.Diagnostics.CandidatesConsidered != len(ranked) {
		t.Errorf("Diagnostics.CandidatesConsidered = %d, want %d", result.Diagnostics.CandidatesConsidered, len(ranked))
	}
	if got := result.Diagnostics.TokensAvailable - result.Diagnostics.TokensExcluded; got != result.Diagnostics.TokensSelected {
		t.Errorf("TokensAvailable - TokensExcluded = %d, want TokensSelected (%d)", got, result.Diagnostics.TokensSelected)
	}

	// 3. No budget specified uses the fixed default (quickstart.md §3).
	defaultResult := contextengine.ApplyBudget(ranked, contextengine.Request{Target: "SPEC-014"}, artifacts.DefaultEstimator{})
	if defaultResult.Diagnostics.TokensAvailable != result.Diagnostics.TokensAvailable {
		t.Errorf("default-budget TokensAvailable = %d, want %d (same candidate set)", defaultResult.Diagnostics.TokensAvailable, result.Diagnostics.TokensAvailable)
	}

	// 4. Mandatory content alone exceeding a deliberately tiny budget
	// (quickstart.md §4): still returned in full, flagged, with the
	// overage reported.
	tiny := 1
	tinyResult := contextengine.ApplyBudget(ranked, contextengine.Request{Target: "SPEC-014", Budget: &tiny, HardLimit: &tiny}, artifacts.DefaultEstimator{})
	if !tinyResult.BudgetExceeded {
		t.Fatal("ApplyBudget() with budget 1: BudgetExceeded = false, want true")
	}
	if tinyResult.Overage <= 0 {
		t.Errorf("ApplyBudget() with budget 1: Overage = %d, want > 0", tinyResult.Overage)
	}
	for _, item := range tinyResult.Items {
		if minTierOf(item.Reasons) != contextengine.TierMandatory {
			t.Errorf("ApplyBudget() with budget 1 kept a non-mandatory item: %+v", item)
		}
	}

	// 5. Repeating the same call is byte-for-byte identical
	// (quickstart.md §5, FR-011, SC-005).
	repeat := contextengine.ApplyBudget(contextengine.Rank(set, req), req, artifacts.DefaultEstimator{})
	if !reflect.DeepEqual(repeat, contextengine.ApplyBudget(contextengine.Rank(set, req), req, artifacts.DefaultEstimator{})) {
		t.Error("ApplyBudget(Rank(...), ...) is not deterministic across repeated calls")
	}
}

func minTierOf(reasons []contextengine.Reason) contextengine.Tier {
	min := reasons[0].Tier
	for _, r := range reasons[1:] {
		if r.Tier < min {
			min = r.Tier
		}
	}
	return min
}
