package contextengine

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/context/index"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// Collect resolves req.Target (must be one of the five entity types
// operations.References already covers — research.md #3), gathers its
// mandatory (Tier 0/1), structural/semantic (Tier 2/3), text (Tier 4),
// and bounded second-hop (Tier 5) context, deduplicates it (FR-008),
// and returns it in deterministic order. Strictly read-only (FR-012).
func Collect(root string, cfg project.Configuration, store index.Store, req Request) (CandidateSet, error) {
	if err := validateIntent(req.Intent); err != nil {
		return CandidateSet{}, err
	}

	target, err := operations.Inspect(root, cfg, req.Target)
	if err != nil {
		return CandidateSet{}, err
	}
	if !target.Location.Type.HasEntityID() {
		return CandidateSet{}, fmt.Errorf("%w: %v is not a standalone-referenceable entity type", operations.ErrInvalidTarget, target.Location.Type)
	}

	var raw []Candidate

	// Tier 0: the project's Constitution, if one exists (FR-003, FR-004).
	constCandidates, err := chunkArtifact(root, cfg.ConstitutionPath, TierMandatory, "constitution")
	switch {
	case err == nil:
		raw = append(raw, constCandidates...)
	case errors.Is(err, artifacts.ErrArtifactNotFound):
		// No Constitution yet — not a failure (FR-004).
	default:
		return CandidateSet{}, err
	}

	// Tier 1: the target itself (FR-003).
	targetCandidates, err := chunkArtifact(root, target.Location.Path, TierMandatory, "target")
	if err != nil {
		return CandidateSet{}, err
	}
	raw = append(raw, targetCandidates...)

	// Tier 2/3: formal relationships and semantic connections, both
	// outgoing (operations.References) and incoming (operations.
	// Backlinks) — computed directly from project files, never through
	// the disposable index (research.md #5).
	refs, err := operations.References(root, cfg, req.Target)
	if err != nil {
		return CandidateSet{}, err
	}
	structural, err := connectedCandidates(root, cfg, refs.Formal, func(e operations.ReferenceEntry) string { return e.Target.String() },
		func(e operations.ReferenceEntry) (Tier, string) { return TierStructural, e.Relation })
	if err != nil {
		return CandidateSet{}, err
	}
	raw = append(raw, structural...)

	semantic, err := connectedCandidates(root, cfg, refs.Semantic, func(e operations.ReferenceEntry) string { return e.Target.String() },
		func(e operations.ReferenceEntry) (Tier, string) { return TierSemantic, e.Relation })
	if err != nil {
		return CandidateSet{}, err
	}
	raw = append(raw, semantic...)

	backlinks, err := operations.Backlinks(root, cfg, req.Target)
	if err != nil {
		return CandidateSet{}, err
	}
	incomingFormal, err := connectedCandidates(root, cfg, backlinks.Formal, func(e operations.BacklinkEntry) string { return e.Source.String() },
		func(e operations.BacklinkEntry) (Tier, string) { return TierSemantic, "backlink" })
	if err != nil {
		return CandidateSet{}, err
	}
	raw = append(raw, incomingFormal...)

	incomingSemantic, err := connectedCandidates(root, cfg, backlinks.Semantic, func(e operations.BacklinkEntry) string { return e.Source.String() },
		func(e operations.BacklinkEntry) (Tier, string) { return TierSemantic, "backlink" })
	if err != nil {
		return CandidateSet{}, err
	}
	raw = append(raw, incomingSemantic...)

	// Tier 5: one bounded, outgoing-only hop from each first-hop (Tier
	// 2/3) artifact — never that artifact's own backlinks, never a
	// third hop, always attempted (research.md #7). A first-hop
	// artifact outside the five referenceable types simply contributes
	// no second-hop expansion of its own.
	var firstHopIDs []string
	for _, e := range refs.Formal {
		firstHopIDs = append(firstHopIDs, e.Target.String())
	}
	for _, e := range refs.Semantic {
		firstHopIDs = append(firstHopIDs, e.Target.String())
	}
	for _, e := range backlinks.Formal {
		firstHopIDs = append(firstHopIDs, e.Source.String())
	}
	for _, e := range backlinks.Semantic {
		firstHopIDs = append(firstHopIDs, e.Source.String())
	}
	secondHop, err := secondHopCandidates(root, cfg, firstHopIDs)
	if err != nil {
		return CandidateSet{}, err
	}
	raw = append(raw, secondHop...)

	// Tier 4: free-text matches, using Query verbatim, or Task verbatim
	// when Query is empty — never a fabricated question (FR-007,
	// research.md #6). The only tier reading from the disposable index
	// (research.md #5).
	query := req.Query
	if query == "" {
		query = req.Task
	}
	if query != "" {
		results, err := store.Search(query, textSearchLimit)
		if err != nil {
			return CandidateSet{}, err
		}
		for _, r := range results {
			raw = append(raw, Candidate{
				Path:      r.Path,
				Heading:   r.Heading,
				Content:   r.Content,
				StartLine: r.StartLine,
				EndLine:   r.EndLine,
				Reasons:   []Reason{{Tier: TierText, Relation: "text_match"}},
			})
		}
	}

	return mergeAndSort(raw), nil
}

// textSearchLimit bounds Tier 4's own Search call. This feature applies
// no budget of its own (FR-013) — this is simply a generous ceiling on
// how many raw text matches a single Collect call ever considers, not a
// relevance cutoff.
const textSearchLimit = 100

// secondHopCandidates computes each first-hop artifact's own outgoing
// operations.References — never that artifact's own Backlinks, never a
// third hop (research.md #7) — labeling every result TierSecondHop with
// its own relation. A first-hop artifact outside the five
// referenceable types (operations.ErrInvalidTarget) simply contributes
// nothing; any other error still propagates.
func secondHopCandidates(root string, cfg project.Configuration, firstHopIDs []string) ([]Candidate, error) {
	var out []Candidate
	for _, id := range firstHopIDs {
		refs, err := operations.References(root, cfg, id)
		if err != nil {
			if errors.Is(err, operations.ErrInvalidTarget) {
				continue
			}
			return nil, err
		}
		formal, err := connectedCandidates(root, cfg, refs.Formal, func(e operations.ReferenceEntry) string { return e.Target.String() },
			func(e operations.ReferenceEntry) (Tier, string) { return TierSecondHop, e.Relation })
		if err != nil {
			return nil, err
		}
		out = append(out, formal...)

		semantic, err := connectedCandidates(root, cfg, refs.Semantic, func(e operations.ReferenceEntry) string { return e.Target.String() },
			func(e operations.ReferenceEntry) (Tier, string) { return TierSecondHop, e.Relation })
		if err != nil {
			return nil, err
		}
		out = append(out, semantic...)
	}
	return out, nil
}

// connectedCandidates resolves and chunks every entry's own referenced
// artifact (via idOf) directly from the filesystem, labeling every
// resulting Candidate per reasonOf.
func connectedCandidates[E any](root string, cfg project.Configuration, entries []E, idOf func(E) string, reasonOf func(E) (Tier, string)) ([]Candidate, error) {
	var out []Candidate
	for _, e := range entries {
		loc, err := operations.Inspect(root, cfg, idOf(e))
		if err != nil {
			return nil, err
		}
		tier, relation := reasonOf(e)
		candidates, err := chunkArtifact(root, loc.Location.Path, tier, relation)
		if err != nil {
			return nil, err
		}
		out = append(out, candidates...)
	}
	return out, nil
}

// chunkArtifact reads, structures, and chunks the artifact at
// root/relPath (011's ReadBody, 013's ParseDocument/Chunks), labeling
// every resulting Chunk with one Reason{tier, relation}. Every
// Candidate's StartLine/EndLine is file-absolute, not relative to the
// post-frontmatter body ParseDocument itself operates on — achieved by
// adding artifacts.ReadBodyWithOffset's own reported body-start line
// (minus 1) to each Chunk's body-relative numbers
// (033-context-pack-output-contract research.md Decision 1 — this was
// the actual bug behind every Context Pack item's location, not only
// wikilinks).
func chunkArtifact(root, relPath string, tier Tier, relation string) ([]Candidate, error) {
	body, bodyStartLine, err := artifacts.ReadBodyWithOffset(filepath.Join(root, relPath))
	if err != nil {
		return nil, err
	}
	offset := bodyStartLine - 1

	doc := artifacts.ParseDocument(body)
	chunks := artifacts.Chunks(relPath, doc)

	out := make([]Candidate, 0, len(chunks))
	for _, c := range chunks {
		out = append(out, Candidate{
			Path:      c.Path,
			Heading:   c.Heading,
			Content:   c.Content,
			StartLine: c.StartLine + offset,
			EndLine:   c.EndLine + offset,
			Reasons:   []Reason{{Tier: tier, Relation: relation}},
		})
	}
	return out, nil
}
