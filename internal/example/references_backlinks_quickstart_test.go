// This file compiles and runs specs/012-references-backlinks/
// quickstart.md's end-to-end flow (query an artifact's outgoing
// references, query what points back at it, confirm a broken wikilink
// contributes nothing, confirm an unreferenced artifact returns a
// well-formed empty answer, confirm an out-of-scope target is rejected
// consistently) against internal/operations, on top of the fixture
// project the other quickstart tests in this package use (plan.md Phase
// 6, T019).
package example

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestReferencesBacklinksQuickstart_EndToEnd(t *testing.T) {
	root := testutil.Project(t)

	proj, err := project.Detect(root)
	if err != nil {
		t.Fatalf("project.Detect() unexpected error: %v", err)
	}

	prg, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{Type: ids.Program})
	if err != nil {
		t.Fatalf("operations.Create(Program) unexpected error: %v", err)
	}
	feat, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{
		Type:   ids.Feature,
		Parent: prg.ID.String(),
	})
	if err != nil {
		t.Fatalf("operations.Create(Feature) unexpected error: %v", err)
	}
	dependency, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{
		Type:   ids.Spec,
		Parent: feat.ID.String(),
	})
	if err != nil {
		t.Fatalf("operations.Create(Spec) dependency unexpected error: %v", err)
	}
	know, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{
		Type: ids.Knowledge,
		Slug: "authentication-model",
	})
	if err != nil {
		t.Fatalf("operations.Create(Knowledge) unexpected error: %v", err)
	}

	// 1. Create the subject Spec through the same deterministic path as
	// every other artifact here, then overwrite its content to add a
	// formal depends_on (§1), a valid semantic wikilink to the Knowledge
	// artifact (§1-2), and a broken wikilink (§3) — mirroring
	// quickstart.md's own fixture, since operations.Create has no
	// depends_on flag of its own.
	subject, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{
		Type:   ids.Spec,
		Parent: feat.ID.String(),
	})
	if err != nil {
		t.Fatalf("operations.Create(Spec) subject unexpected error: %v", err)
	}
	testutil.WriteFile(t, root, subject.Path, ""+
		"---\nid: "+subject.ID.String()+"\ntype: spec\nstatus: ready\nparent: "+feat.ID.String()+"\n"+
		"depends_on:\n  - "+dependency.ID.String()+"\nsupersedes: []\n---\n"+
		"## Relevant Knowledge\n\n"+
		"- [["+know.ID.String()+"|Authentication Model]]\n"+
		"- [[SPEC-899]]\n") // broken: no such Spec exists

	// 2. Query what the subject points to (quickstart.md §1-2).
	refs, err := operations.References(proj.Root, proj.Config, subject.ID.String())
	if err != nil {
		t.Fatalf("operations.References() unexpected error: %v", err)
	}
	if len(refs.Formal) != 2 {
		t.Fatalf("References().Formal = %+v, want 2 (parent, depends_on)", refs.Formal)
	}
	if refs.Formal[0].Relation != "parent" || refs.Formal[0].Target.String() != feat.ID.String() {
		t.Errorf("Formal[0] = %+v, want relation=parent target=%v", refs.Formal[0], feat.ID)
	}
	if refs.Formal[1].Relation != "depends_on" || refs.Formal[1].Target.String() != dependency.ID.String() {
		t.Errorf("Formal[1] = %+v, want relation=depends_on target=%v", refs.Formal[1], dependency.ID)
	}
	// 3. The broken wikilink contributes nothing; the valid one does
	// (quickstart.md §3).
	if len(refs.Semantic) != 1 || refs.Semantic[0].Target.String() != know.ID.String() {
		t.Fatalf("References().Semantic = %+v, want exactly [{wikilink %v}]", refs.Semantic, know.ID)
	}

	// 4. Query what points at the dependency (quickstart.md §2).
	backlinks, err := operations.Backlinks(proj.Root, proj.Config, dependency.ID.String())
	if err != nil {
		t.Fatalf("operations.Backlinks() unexpected error: %v", err)
	}
	if len(backlinks.Formal) != 1 || backlinks.Formal[0].Relation != "depends_on" || backlinks.Formal[0].Source.String() != subject.ID.String() {
		t.Errorf("Backlinks().Formal = %+v, want [{depends_on %v}]", backlinks.Formal, subject.ID)
	}

	// 5. An artifact nothing points to returns a well-formed empty
	// answer, not an error (quickstart.md §4).
	lonely, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{
		Type:   ids.Spec,
		Parent: feat.ID.String(),
	})
	if err != nil {
		t.Fatalf("operations.Create(Spec) lonely unexpected error: %v", err)
	}
	lonelyBacklinks, err := operations.Backlinks(proj.Root, proj.Config, lonely.ID.String())
	if err != nil {
		t.Fatalf("operations.Backlinks() unexpected error: %v", err)
	}
	if len(lonelyBacklinks.Formal) != 0 || len(lonelyBacklinks.Semantic) != 0 {
		t.Errorf("Backlinks(%v) = %+v, want both empty", lonely.ID, lonelyBacklinks)
	}

	// 6. An out-of-scope target (Task) is rejected consistently
	// (quickstart.md §5) — a real Task must exist first, or Resolve
	// itself reports entity_not_found before this feature's own scope
	// check is ever reached.
	testutil.WriteFile(t, root, "ai/programs/"+prg.ID.String()+"/features/"+feat.ID.String()+"/specs/"+dependency.ID.String()+"/tasks.md",
		"---\ntype: tasks\nfor: "+dependency.ID.String()+"\n---\n# Tasks\n\n## TASK-001 — Do it\n\n- [ ] Complete\n")
	_, err = operations.References(proj.Root, proj.Config, "TASK-001")
	if !errors.Is(err, operations.ErrInvalidTarget) {
		t.Errorf("References(TASK-001) error = %v, want errors.Is(err, ErrInvalidTarget)", err)
	}
}
