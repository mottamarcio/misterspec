package index

import (
	"sort"
	"testing"

	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

// setupRelationshipFixture writes a small graph: SPEC-002 formally
// depends on SPEC-001 and semantically links to KNOW-001. It also
// includes one Plan (a non-ID-bearing type) to prove such artifacts
// contribute no links rows (research.md #4).
func setupRelationshipFixture(t *testing.T, root string) {
	t.Helper()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nThe target spec.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/plan.md",
		"---\ntype: plan\nfor: SPEC-001\n---\n## Summary\n\nSee [[SPEC-001]] for context.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-auth.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nAuthentication facts.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-001\nsupersedes: []\n---\n## Intent\n\nSee [[KNOW-001]] for background.\n")
}

func sortLinks(links []Link) {
	sort.Slice(links, func(i, j int) bool {
		if links[i].Relation != links[j].Relation {
			return links[i].Relation < links[j].Relation
		}
		if links[i].Source != links[j].Source {
			return links[i].Source < links[j].Source
		}
		return links[i].Target < links[j].Target
	})
}

func TestOutgoing_MatchesOperationsReferences(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupRelationshipFixture(t, root)
	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	want, err := operations.References(root, cfg, "SPEC-002")
	if err != nil {
		t.Fatalf("operations.References() unexpected error: %v", err)
	}
	var wantLinks []Link
	for _, r := range want.Formal {
		wantLinks = append(wantLinks, Link{Relation: r.Relation, Source: "SPEC-002", Target: r.Target.String()})
	}
	for _, r := range want.Semantic {
		wantLinks = append(wantLinks, Link{Relation: r.Relation, Source: "SPEC-002", Target: r.Target.String()})
	}
	sortLinks(wantLinks)

	got, err := store.Outgoing("SPEC-002")
	if err != nil {
		t.Fatalf("Outgoing() unexpected error: %v", err)
	}
	sortLinks(got)

	if len(got) != len(wantLinks) {
		t.Fatalf("Outgoing(SPEC-002) = %+v, want %+v", got, wantLinks)
	}
	for i := range wantLinks {
		if got[i] != wantLinks[i] {
			t.Errorf("Outgoing(SPEC-002)[%d] = %+v, want %+v", i, got[i], wantLinks[i])
		}
	}
}

func TestIncoming_MatchesOperationsBacklinks(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupRelationshipFixture(t, root)
	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	want, err := operations.Backlinks(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("operations.Backlinks() unexpected error: %v", err)
	}
	var wantLinks []Link
	for _, b := range want.Formal {
		wantLinks = append(wantLinks, Link{Relation: b.Relation, Source: b.Source.String(), Target: "SPEC-001"})
	}
	for _, b := range want.Semantic {
		wantLinks = append(wantLinks, Link{Relation: b.Relation, Source: b.Source.String(), Target: "SPEC-001"})
	}
	sortLinks(wantLinks)

	got, err := store.Incoming("SPEC-001")
	if err != nil {
		t.Fatalf("Incoming() unexpected error: %v", err)
	}
	sortLinks(got)

	if len(got) != len(wantLinks) {
		t.Fatalf("Incoming(SPEC-001) = %+v, want %+v", got, wantLinks)
	}
	for i := range wantLinks {
		if got[i] != wantLinks[i] {
			t.Errorf("Incoming(SPEC-001)[%d] = %+v, want %+v", i, got[i], wantLinks[i])
		}
	}
}

func TestOutgoing_NoRelationshipsIsEmptyNotError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupRelationshipFixture(t, root)
	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	got, err := store.Outgoing("KNOW-001")
	if err != nil {
		t.Fatalf("Outgoing() unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Outgoing(KNOW-001) = %+v, want none", got)
	}
}

func TestLinks_NonIDBearingArtifactContributesNoLinksRows(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupRelationshipFixture(t, root)
	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	// The total links count must equal exactly the sum of
	// operations.References' own output across every ID-bearing
	// artifact in the fixture (their own "parent" edges included) — no
	// more. The Plan's own wikilink to SPEC-001 must not have added
	// anything beyond that: Plan is not one of the five ID-bearing
	// types operations.References covers (research.md #4), even though
	// it was indexed for full-text search.
	var want int
	for _, id := range []string{"PRG-001", "FEAT-001", "SPEC-001", "SPEC-002", "KNOW-001"} {
		refs, err := operations.References(root, cfg, id)
		if err != nil {
			t.Fatalf("operations.References(%s) unexpected error: %v", id, err)
		}
		want += len(refs.Formal) + len(refs.Semantic)
	}

	var count int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM links`).Scan(&count); err != nil {
		t.Fatalf("counting links: %v", err)
	}
	if count != want {
		t.Errorf("links total = %d, want exactly %d (the sum of every ID-bearing artifact's own References — nothing contributed by the Plan)", count, want)
	}
}
