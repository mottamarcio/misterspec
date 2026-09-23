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
	constCandidates, err := chunkArtifact(root, cfg.ConstitutionPath, Reason{Tier: TierMandatory, Relation: "constitution"})
	switch {
	case err == nil:
		raw = append(raw, constCandidates...)
	case errors.Is(err, artifacts.ErrArtifactNotFound):
		// No Constitution yet — not a failure (FR-004).
	default:
		return CandidateSet{}, err
	}

	// Tier 1: the target itself (FR-003).
	targetCandidates, err := chunkArtifact(root, target.Location.Path, Reason{Tier: TierMandatory, Relation: "target"})
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
		func(e operations.ReferenceEntry) Reason { return referenceEntryReason(TierStructural, e) })
	if err != nil {
		return CandidateSet{}, err
	}
	raw = append(raw, structural...)

	semantic, err := connectedCandidates(root, cfg, refs.Semantic, func(e operations.ReferenceEntry) string { return e.Target.String() },
		func(e operations.ReferenceEntry) Reason { return referenceEntryReason(TierSemantic, e) })
	if err != nil {
		return CandidateSet{}, err
	}
	raw = append(raw, semantic...)

	backlinks, err := operations.Backlinks(root, cfg, req.Target)
	if err != nil {
		return CandidateSet{}, err
	}
	incomingFormal, err := connectedCandidates(root, cfg, backlinks.Formal, func(e operations.BacklinkEntry) string { return e.Source.String() },
		func(e operations.BacklinkEntry) Reason { return backlinkEntryReason(e) })
	if err != nil {
		return CandidateSet{}, err
	}
	raw = append(raw, incomingFormal...)

	incomingSemantic, err := connectedCandidates(root, cfg, backlinks.Semantic, func(e operations.BacklinkEntry) string { return e.Source.String() },
		func(e operations.BacklinkEntry) Reason { return backlinkEntryReason(e) })
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
		var results []index.SearchResult
		var err error
		if req.QueryMode == QueryModeAdvanced {
			results, err = store.SearchAdvanced(query, textSearchLimit)
		} else {
			results, err = store.Search(query, textSearchLimit)
		}
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
				TextRank:  r.Rank,
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
			func(e operations.ReferenceEntry) Reason { return referenceEntryReason(TierSecondHop, e) })
		if err != nil {
			return nil, err
		}
		out = append(out, formal...)

		semantic, err := connectedCandidates(root, cfg, refs.Semantic, func(e operations.ReferenceEntry) string { return e.Target.String() },
			func(e operations.ReferenceEntry) Reason { return referenceEntryReason(TierSecondHop, e) })
		if err != nil {
			return nil, err
		}
		out = append(out, semantic...)
	}
	return out, nil
}

// connectedCandidates resolves and chunks every entry's own referenced
// artifact (via idOf) directly from the filesystem, labeling every
// resulting Candidate with one Reason built by reasonOf — which already
// carries any wikilink occurrence data the entry itself has (038-
// wikilink-chunk-provenance).
func connectedCandidates[E any](root string, cfg project.Configuration, entries []E, idOf func(E) string, reasonOf func(E) Reason) ([]Candidate, error) {
	var out []Candidate
	for _, e := range entries {
		loc, err := operations.Inspect(root, cfg, idOf(e))
		if err != nil {
			return nil, err
		}
		reason := reasonOf(e)
		// An anchor-qualified wikilink contributes exactly one Candidate
		// (that Section alone) instead of chunkArtifact's whole-artifact
		// expansion (040-stable-section-anchors contracts §5) — every
		// other relation/entry is completely unaffected (spec FR-008).
		if reason.Relation == "wikilink" && reason.TargetAnchor != "" {
			candidate, err := chunkArtifactAnchor(root, loc.Location.Path, reason.TargetAnchor, reason)
			if errors.Is(err, errAnchorNotFound) {
				// The anchor doesn't exist on the target artifact —
				// operations.References/Backlinks populate TargetAnchor
				// straight from the wikilink's own raw text with no
				// existence check (that is internal/validation's job,
				// CodeUnknownAnchor), so this is reachable in normal use,
				// not just theoretically. Retrieval must not fabricate a
				// contentless, nonzero-scored Candidate for a reference
				// that resolves to nothing — silently omit it here; the
				// diagnostic surfaces separately via `internal validate`
				// (code review finding, corrects the contract's earlier,
				// wrong "never reached here in practice" assumption).
				continue
			}
			if err != nil {
				return nil, err
			}
			out = append(out, candidate)
			continue
		}
		candidates, err := chunkArtifact(root, loc.Location.Path, reason)
		if err != nil {
			return nil, err
		}
		out = append(out, candidates...)
	}
	return out, nil
}

// referenceEntryReason builds the Reason for an outgoing
// operations.ReferenceEntry at tier — Relation and every occurrence
// field (SourcePath/SourceSection/SourceLine) copied straight from e,
// which already carries them correctly shaped (empty/zero for a formal
// entry, populated for a semantic one) — no relation-based branching
// needed here (data-model.md "Reason (extended)").
func referenceEntryReason(tier Tier, e operations.ReferenceEntry) Reason {
	return Reason{
		Tier:          tier,
		Relation:      e.Relation,
		SourcePath:    e.SourcePath,
		SourceSection: e.SourceSection,
		SourceLine:    e.SourceLine,
		TargetAnchor:  e.TargetAnchor,
	}
}

// backlinkEntryReason builds the Reason for an incoming
// operations.BacklinkEntry — always TierSemantic/"backlink" (unchanged
// from this codebase's existing behavior, regardless of whether the
// underlying entry was itself formal or semantic), with occurrence
// fields copied straight from e.
func backlinkEntryReason(e operations.BacklinkEntry) Reason {
	return Reason{
		Tier:          TierSemantic,
		Relation:      "backlink",
		SourcePath:    e.SourcePath,
		SourceSection: e.SourceSection,
		SourceLine:    e.SourceLine,
		// TargetAnchor is informational only here — it names which
		// anchor of the current (queried) target the Source artifact
		// referenced, but this Candidate is chunkArtifact-ing the
		// Source artifact itself (idOf == e.Source), never the target,
		// so it intentionally never routes through chunkArtifactAnchor
		// (contracts §5: only Relation == "wikilink" does).
		TargetAnchor: e.TargetAnchor,
	}
}

// chunkArtifact reads, structures, and chunks the artifact at
// root/relPath (011's ReadBody, 013's ParseDocument/Chunks via
// artifacts.ChunksWithOffset), attaching reason (already fully
// populated by the caller) to every resulting Chunk. Every Candidate's
// StartLine/EndLine is file-absolute, not relative to the
// post-frontmatter body ParseDocument itself operates on —
// ChunksWithOffset applies artifacts.ReadBodyWithOffset's own reported
// body-start line (minus 1) to each Chunk's body-relative numbers
// (033-context-pack-output-contract research.md Decision 1 — this was
// the actual bug behind every Context Pack item's location, not only
// wikilinks).
func chunkArtifact(root, relPath string, reason Reason) ([]Candidate, error) {
	body, bodyStartLine, err := artifacts.ReadBodyWithOffset(filepath.Join(root, relPath))
	if err != nil {
		return nil, err
	}
	offset := bodyStartLine - 1

	chunks := artifacts.ChunksWithOffset(relPath, body, offset)

	out := make([]Candidate, 0, len(chunks))
	for _, c := range chunks {
		out = append(out, Candidate{
			Path:      c.Path,
			Heading:   c.Heading,
			Content:   c.Content,
			StartLine: c.StartLine,
			EndLine:   c.EndLine,
			Reasons:   []Reason{reason},
		})
	}
	return out, nil
}

// chunkArtifactAnchor resolves exactly one Candidate for the Section
// whose Anchor == anchor within the artifact at relPath (040-stable-
// section-anchors contracts §5) — never the whole artifact's own
// chunkArtifact expansion. It walks artifacts.ParseDocument's own
// Sections directly, not artifacts.Chunks()'s already-filtered output,
// because Chunks() silently skips a Section whose Body is empty (a
// heading immediately followed by a subheading); an anchor on such a
// Section is still valid and must resolve, with empty Content, rather
// than being treated as not-found (data-model.md "Chunk (extended)").
// HeadingPath is that Section's ancestor headings' own titles,
// outermost first, computed from the flat Sections list's own Level
// nesting — the same information chunkArtifact's callers never needed,
// since a full-artifact expansion carries every ancestor's own Chunk
// already. Returns an error only for an I/O failure — anchor-not-found
// is validation's own concern (internal/validation), never reached here
// in practice, since collector.go only calls this for a TargetAnchor a
// prior wikilink-classification step already confirmed exists.
// errAnchorNotFound is returned by chunkArtifactAnchor when relPath
// declares no Section with the given anchor. Reachable in normal use —
// operations.References/Backlinks populate TargetAnchor from the raw
// wikilink text with no existence check (internal/validation's own
// separate job, CodeUnknownAnchor) — so callers MUST check for it with
// errors.Is and omit the reference, never propagate it as a fatal
// Collect error nor fabricate a placeholder Candidate for it (code
// review finding; corrects this function's own earlier assumption that
// the case was unreachable).
var errAnchorNotFound = errors.New("contextengine: anchor not found in target artifact")

func chunkArtifactAnchor(root, relPath, anchor string, reason Reason) (Candidate, error) {
	body, bodyStartLine, err := artifacts.ReadBodyWithOffset(filepath.Join(root, relPath))
	if err != nil {
		return Candidate{}, err
	}
	offset := bodyStartLine - 1
	doc := artifacts.ParseDocument(body)

	idx := -1
	for i, s := range doc.Sections {
		if s.Anchor == anchor {
			idx = i
			break
		}
	}
	if idx == -1 {
		return Candidate{}, fmt.Errorf("%w: %q in %s", errAnchorNotFound, anchor, relPath)
	}

	target := doc.Sections[idx]
	reason.TargetAnchor = anchor

	return Candidate{
		Path:        relPath,
		Heading:     target.Heading,
		Content:     target.Body,
		StartLine:   target.StartLine + offset,
		EndLine:     target.EndLine + offset,
		Reasons:     []Reason{reason},
		HeadingPath: ancestorHeadingPath(doc.Sections, idx),
	}, nil
}

// ancestorHeadingPath returns the titles of every Section preceding
// sections[idx] (in document order) whose own Level is strictly less
// than each ancestor found so far — the same "walk backward, track the
// smallest Level seen" approach any flat, level-tagged outline uses to
// recover nesting without a tree structure of its own. Outermost
// ancestor first.
func ancestorHeadingPath(sections []artifacts.Section, idx int) []string {
	// Always non-nil, even when empty (a top-level anchored Section has
	// no ancestors) — callers distinguish "this Candidate came from
	// chunkArtifactAnchor" from "it didn't" by nil-ness, not length,
	// since a real, valid HeadingPath can legitimately be empty (found
	// during implementation: a manual end-to-end walkthrough of
	// quickstart.md §1, whose own example anchors a top-level heading
	// with no ancestor, surfaced that gating on len()>0 silently
	// dropped heading_path/anchor for exactly that case).
	path := []string{}
	minLevel := sections[idx].Level
	for i := idx - 1; i >= 0 && minLevel > 1; i-- {
		if sections[i].Level > 0 && sections[i].Level < minLevel {
			path = append([]string{sections[i].Heading}, path...)
			minLevel = sections[i].Level
		}
	}
	return path
}
