// This file compiles and runs specs/015-context-collector/
// quickstart.md's end-to-end flow (mandatory baseline, structural/
// semantic connections, a text match, cross-tier deduplication, bounded
// second-hop expansion, and an out-of-scope target rejection) against
// internal/context, on top of the fixture project the other quickstart
// tests in this package use (plan.md Phase 8, T017).
package example

import (
	"errors"
	"path/filepath"
	"testing"

	contextengine "github.com/mottamarcio/misterspec/internal/context"
	"github.com/mottamarcio/misterspec/internal/context/index"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestContextCollectorQuickstart_EndToEnd(t *testing.T) {
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
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-009/spec.md",
		"---\nid: SPEC-009\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nSecond-hop-only content.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-011/spec.md",
		"---\nid: SPEC-011\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on:\n  - SPEC-009\nsupersedes: []\n---\n## Intent\n\nRefresh token rotation details.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-auth.md",
		"---\nid: KNOW-003\ntype: knowledge\nstatus: active\n---\n## Summary\n\nAuthentication model facts.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on:\n  - SPEC-011\nsupersedes: []\n---\n## Intent\n\nSee [[KNOW-003]] for background.\n")

	store, err := index.Open(filepath.Join(root, ".misterspec/cache/context.db"))
	if err != nil {
		t.Fatalf("index.Open() unexpected error: %v", err)
	}
	defer store.Close()
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	// 1. Ask for a target with nothing else specified (quickstart.md §1).
	set, err := contextengine.Collect(root, cfg, store, contextengine.Request{Target: "SPEC-014"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}
	want := map[string]bool{"target": false, "constitution": false, "depends_on": false, "wikilink": false}
	for _, c := range set.Candidates {
		for _, r := range c.Reasons {
			if _, ok := want[r.Relation]; ok {
				want[r.Relation] = true
			}
		}
	}
	for relation, seen := range want {
		if !seen {
			t.Errorf("Collect() Candidates missing a %q reason: %+v", relation, set.Candidates)
		}
	}

	// 2. Ask with a free-text question (quickstart.md §2).
	withQuery, err := contextengine.Collect(root, cfg, store, contextengine.Request{
		Target: "SPEC-014",
		Query:  "refresh token rotation",
	})
	if err != nil {
		t.Fatalf("Collect() with Query unexpected error: %v", err)
	}
	var sawTextMatch bool
	for _, c := range withQuery.Candidates {
		for _, r := range c.Reasons {
			if r.Tier == contextengine.TierText && r.Relation == "text_match" {
				sawTextMatch = true
			}
		}
	}
	if !sawTextMatch {
		t.Errorf("Collect() with Query = %+v, want a text_match entry", withQuery.Candidates)
	}

	// 3. The same chunk found two ways appears once (quickstart.md §3):
	// SPEC-011 is both a direct dependency and a text match for
	// "rotation".
	var spec011Count int
	var spec011Reasons []contextengine.Reason
	for _, c := range withQuery.Candidates {
		if c.Path == "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-011/spec.md" {
			spec011Count++
			spec011Reasons = c.Reasons
		}
	}
	if spec011Count != 1 {
		t.Fatalf("SPEC-011 appears %d times, want exactly 1: %+v", spec011Count, withQuery.Candidates)
	}
	var hasStructural, hasText bool
	for _, r := range spec011Reasons {
		if r.Relation == "depends_on" {
			hasStructural = true
		}
		if r.Relation == "text_match" {
			hasText = true
		}
	}
	if !hasStructural || !hasText {
		t.Errorf("SPEC-011 reasons = %+v, want both depends_on and text_match", spec011Reasons)
	}

	// 4. Second-hop expansion (quickstart.md §4): SPEC-011 depends on
	// SPEC-009, one hop beyond SPEC-014's own direct dependency.
	var sawSecondHop bool
	for _, c := range set.Candidates {
		if c.Path == "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-009/spec.md" {
			for _, r := range c.Reasons {
				if r.Tier == contextengine.TierSecondHop {
					sawSecondHop = true
				}
			}
		}
	}
	if !sawSecondHop {
		t.Errorf("Collect() Candidates = %+v, want a TierSecondHop entry for SPEC-009", set.Candidates)
	}

	// 5. An out-of-scope target is rejected consistently (quickstart.md §5).
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/tasks.md",
		"---\ntype: tasks\nfor: SPEC-014\n---\n# Tasks\n\n## TASK-001 — Do it\n\n- [ ] Complete\n")
	_, err = contextengine.Collect(root, cfg, store, contextengine.Request{Target: "TASK-001"})
	if !errors.Is(err, operations.ErrInvalidTarget) {
		t.Errorf("Collect(TASK-001) error = %v, want errors.Is(err, ErrInvalidTarget)", err)
	}
}
