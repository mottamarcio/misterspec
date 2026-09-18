// This file compiles and runs specs/011-wikilink-foundation/quickstart.md's
// end-to-end flow (author explicit wikilinks in a Spec's body, extract
// them, then confirm structural validation catches a broken link and a
// malformed one while distinguishing them, and leaves a link-free
// artifact exactly as clean as before this feature existed) against
// internal/artifacts and internal/validation, on top of the fixture
// project the other quickstart tests in this package use (plan.md Phase
// 6, T015).
package example

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
	"github.com/mottamarcio/misterspec/internal/validation"
)

func TestWikilinkQuickstart_EndToEnd(t *testing.T) {
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

	// target is a real, resolvable Spec other artifacts can link to.
	target, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{
		Type:   ids.Spec,
		Parent: feat.ID.String(),
	})
	if err != nil {
		t.Fatalf("operations.Create(Spec) target unexpected error: %v", err)
	}
	// subject is the Spec whose body carries the wikilinks (quickstart.md §1).
	subject, err := operations.Create(proj.Root, proj.Config, operations.CreateRequest{
		Type:   ids.Spec,
		Parent: feat.ID.String(),
	})
	if err != nil {
		t.Fatalf("operations.Create(Spec) subject unexpected error: %v", err)
	}

	brokenTarget := ids.EntityID{Type: ids.Spec, Prefix: ids.Spec.Prefix(), Number: 999, Width: proj.Config.IDWidth}.String()

	appendBody(t, root, subject.Path, ""+
		"\n## Relevant Knowledge\n\n"+
		"- [["+target.ID.String()+"|Alias]]\n"+
		"- [["+brokenTarget+"]]\n"+
		"- [[not-a-real-id]]\n")

	// 2. Extraction (Go-level, quickstart.md §2).
	body, err := artifacts.ReadBody(filepath.Join(root, subject.Path))
	if err != nil {
		t.Fatalf("artifacts.ReadBody() unexpected error: %v", err)
	}
	links, err := artifacts.ExtractWikiLinks(body)
	if err != nil {
		t.Fatalf("artifacts.ExtractWikiLinks() unexpected error: %v", err)
	}
	if len(links) != 3 {
		t.Fatalf("len(links) = %d, want 3: %+v", len(links), links)
	}
	if links[0].Target != target.ID.String() || links[0].Alias != "Alias" {
		t.Errorf("links[0] = %+v, want Target=%s Alias=Alias", links[0], target.ID.String())
	}
	if links[1].Target != brokenTarget {
		t.Errorf("links[1].Target = %q, want %q", links[1].Target, brokenTarget)
	}
	if links[2].Target != "not-a-real-id" {
		t.Errorf("links[2].Target = %q, want %q", links[2].Target, "not-a-real-id")
	}

	// 3 & 4. Validation catches the broken link and distinguishes it from
	// the malformed one (quickstart.md §3-4).
	findings, err := validation.ValidateEntity(proj.Root, proj.Config, subject.ID.String())
	if err != nil {
		t.Fatalf("validation.ValidateEntity(%v) unexpected error: %v", subject.ID, err)
	}
	var brokenCount, invalidCount, ambiguousCount int
	for _, f := range findings {
		switch f.Code {
		case validation.CodeBrokenWikilink:
			brokenCount++
		case validation.CodeInvalidWikilink:
			invalidCount++
		case validation.CodeAmbiguousWikilink:
			ambiguousCount++
		}
	}
	if brokenCount != 1 {
		t.Errorf("brokenCount = %d, want 1: %+v", brokenCount, findings)
	}
	if invalidCount != 1 {
		t.Errorf("invalidCount = %d, want 1: %+v", invalidCount, findings)
	}
	if ambiguousCount != 0 {
		t.Errorf("ambiguousCount = %d, want 0: %+v", ambiguousCount, findings)
	}

	// 5. An artifact with no links carries no wikilink-related findings
	// (quickstart.md §5, US3/SC-003) — scoped to wikilink codes
	// specifically, not overall emptiness: the freshly scaffolded
	// template's own placeholder "### R1"/"### R2" headings legitimately
	// trigger 032-requirement-coverage-dependency-validation's own
	// uncovered_requirement findings (no tasks.md exists yet), which is
	// an unrelated validation dimension this test does not exercise.
	clean, err := validation.ValidateEntity(proj.Root, proj.Config, target.ID.String())
	if err != nil {
		t.Fatalf("validation.ValidateEntity(%v) unexpected error: %v", target.ID, err)
	}
	for _, f := range clean {
		switch f.Code {
		case validation.CodeBrokenWikilink, validation.CodeInvalidWikilink, validation.CodeAmbiguousWikilink:
			t.Errorf("ValidateEntity(%v) unexpected wikilink finding: %+v", target.ID, f)
		}
	}
}

// appendBody appends extra to the already-created artifact at
// root/relPath — simulating an author adding prose (with wikilinks) to a
// freshly created artifact's body, on top of whatever renderCreate wrote.
func appendBody(t *testing.T, root, relPath, extra string) {
	t.Helper()
	abs := filepath.Join(root, relPath)
	data, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("appendBody: reading %s: %v", abs, err)
	}
	if err := os.WriteFile(abs, append(data, []byte(extra)...), 0o644); err != nil {
		t.Fatalf("appendBody: writing %s: %v", abs, err)
	}
}
