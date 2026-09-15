// This file compiles and runs specs/013-document-model-chunking/
// quickstart.md's end-to-end flow (read an artifact's body, structure
// it into Sections, derive traceable Chunks, confirm the empty-section/
// no-chunk case, confirm a fenced example heading is excluded, estimate
// a chunk's token cost) against internal/artifacts, on top of the
// fixture project the other quickstart tests in this package use
// (plan.md Phase 6, T012).
package example

import (
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestDocumentChunkingQuickstart_EndToEnd(t *testing.T) {
	root := testutil.Project(t)
	path := testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md", ""+
		"---\n"+
		"id: KNOW-001\n"+
		"type: knowledge\n"+
		"status: active\n"+
		"---\n"+
		"## Intent\n\n"+
		"Rotate refresh tokens on every use.\n\n"+
		"## Requirements\n\n"+
		"### R1\n\n"+
		"Refresh tokens MUST be single-use.\n\n"+
		"### R2\n\n"+
		"A reused refresh token MUST revoke the session.\n\n"+
		"## Relevant Knowledge\n\n"+
		"See [[SPEC-011]] for the original design.\n\n"+
		"Example wikilink syntax:\n\n"+
		"```markdown\n"+
		"## Not a real heading\n"+
		"```\n")

	// 1. Read the body and structure it (quickstart.md §1).
	body, err := artifacts.ReadBody(path)
	if err != nil {
		t.Fatalf("artifacts.ReadBody() unexpected error: %v", err)
	}
	doc := artifacts.ParseDocument(body)

	var headings []string
	for _, s := range doc.Sections {
		headings = append(headings, s.Heading)
	}
	wantHeadings := []string{"Intent", "Requirements", "R1", "R2", "Relevant Knowledge"}
	if len(headings) != len(wantHeadings) {
		t.Fatalf("headings = %v, want %v", headings, wantHeadings)
	}
	for i, w := range wantHeadings {
		if headings[i] != w {
			t.Errorf("headings[%d] = %q, want %q", i, headings[i], w)
		}
	}

	// "Requirements" has no content of its own before its first child —
	// confirm it produced an empty Body (no duplication into R1/R2).
	var requirements artifacts.Section
	for _, s := range doc.Sections {
		if s.Heading == "Requirements" {
			requirements = s
		}
	}
	if requirements.Body != "" {
		t.Errorf("Requirements.Body = %q, want empty — its content belongs to R1/R2", requirements.Body)
	}

	// 2. Break it into traceable pieces (quickstart.md §2).
	chunks := artifacts.Chunks(path, doc)

	// "Requirements" produced no Chunk (FR-007); everything else did.
	if len(chunks) != len(wantHeadings)-1 {
		t.Fatalf("len(chunks) = %d, want %d (Requirements' own empty section produces none)", len(chunks), len(wantHeadings)-1)
	}
	for _, c := range chunks {
		if c.Heading == "Requirements" {
			t.Errorf("chunks unexpectedly contains an entry for Requirements: %+v", c)
		}
		if c.Path != path {
			t.Errorf("chunk %+v Path = %q, want %q", c, c.Path, path)
		}
	}

	// The fenced example heading never split "Relevant Knowledge" into
	// two sections (quickstart.md §5).
	var relevantKnowledge artifacts.Chunk
	for _, c := range chunks {
		if c.Heading == "Relevant Knowledge" {
			relevantKnowledge = c
		}
	}
	if relevantKnowledge.Heading == "" {
		t.Fatal("no chunk found for \"Relevant Knowledge\"")
	}

	// The wikilink inside it is preserved verbatim, unresolved
	// (quickstart.md §2, FR-012).
	wantSnippet := "[[SPEC-011]]"
	if !strings.Contains(relevantKnowledge.Content, wantSnippet) {
		t.Errorf("Relevant Knowledge chunk Content = %q, want it to contain %q", relevantKnowledge.Content, wantSnippet)
	}
	// The fenced example's own "## Not a real heading" text is still
	// present as ordinary body content, never treated as a boundary.
	if !strings.Contains(relevantKnowledge.Content, "## Not a real heading") {
		t.Errorf("Relevant Knowledge chunk Content = %q, want it to still contain the fenced example text", relevantKnowledge.Content)
	}

	// 3. Estimate a chunk's token cost (quickstart.md §3).
	n := artifacts.EstimateTokens(relevantKnowledge.Content)
	if n <= 0 {
		t.Errorf("EstimateTokens() = %d, want > 0 for non-empty content", n)
	}
	var estimator artifacts.Estimator = artifacts.DefaultEstimator{}
	if got := estimator.Estimate(relevantKnowledge.Content); got != n {
		t.Errorf("DefaultEstimator{}.Estimate() = %d, want %d (EstimateTokens)", got, n)
	}

	// 4. A heading-free body still works (quickstart.md §4).
	noHeadings := artifacts.ParseDocument([]byte("Just prose, no headings at all.\n"))
	if len(noHeadings.Sections) != 1 || noHeadings.Sections[0].Heading != "" {
		t.Errorf("ParseDocument(no headings) = %+v, want exactly one untitled section", noHeadings.Sections)
	}

	// An entirely empty body produces zero sections, not an error.
	empty := artifacts.ParseDocument([]byte(""))
	if len(empty.Sections) != 0 {
		t.Errorf("ParseDocument(\"\") = %+v, want zero sections", empty.Sections)
	}
}
