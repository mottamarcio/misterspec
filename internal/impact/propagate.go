package impact

import (
	"path/filepath"
	"strings"

	"github.com/mottamarcio/misterspec/internal/evidence"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/prepare"
	"github.com/mottamarcio/misterspec/internal/project"
)

// PropagationHop is one edge in a PropagationPath — one relation,
// walked in the reverse (backlink) direction (data-model.md
// "PropagationHop"). FromID/ToID are rendered canonical ID strings
// ("SPEC-014" or, for a Task, its composite form "SPEC-014:TASK-003",
// 031-canonical-task-identity) rather than ids.EntityID — a Task's own
// identity does not fit that type (it is ids.TaskID instead), and an
// AffectedItem may be either an artifact or a Task, so a single string
// form is used uniformly (implementation-time correction to data-
// model.md's original ids.EntityID-typed draft).
type PropagationHop struct {
	Relation      RelationKind
	FromID        string
	ToID          string
	SourcePath    string
	SourceSection string
	SourceLine    int
	// RequirementNumber is non-nil only for a RelationCoverage hop —
	// which Requirement number in FromID this hop's ToID Task covers.
	// Not part of data-model.md's original draft; added at
	// implementation time so AnalyzeImpact's reason text can name the
	// specific Requirement, per spec FR-005's "specific reason"
	// requirement.
	RequirementNumber *int
	// TaskEvidenceState is non-empty only for a RelationCoverage/
	// RelationEvidence hop — ToID's own 041-task-evidence-fingerprint
	// EvidenceState at the time of this walk, read via
	// prepare.ScanSpecTasks and never recomputed (research.md #10).
	// Code review finding: the original implementation computed
	// coverage hops without ever reading ti.Evidence, so a Task whose
	// own evidence was already Stale/Failed/Unverified for its own 041
	// reasons was indistinguishable from one with perfectly valid
	// evidence — losing exactly the "its own recorded evidence is
	// already stale" reason text data-model.md's AffectedItem.Reason
	// documents, and leaving RelationEvidence/severityTable's own
	// Evidence entry unreachable.
	TaskEvidenceState evidence.EvidenceState
}

// PropagationPath is the full ordered sequence of PropagationHops from
// one ChangedElement to one affected item (data-model.md
// "PropagationPath"). Always has length >= 1.
type PropagationPath struct {
	Hops []PropagationHop
}

// propagationResult is walkPropagation's own return shape: every
// affected item's own set of distinct PropagationPaths, plus which
// ChangedElement IDs produced zero hops at all (spec FR-008).
type propagationResult struct {
	// PathsByItem maps an affected item's own canonical ID string to
	// every distinct PropagationPath reaching it.
	PathsByItem map[string][]PropagationPath
	// HadAnyHop tracks, per ChangedElement.ID, whether any entry sharing
	// that ID (whole-artifact and/or any Requirement-level entry)
	// produced at least one hop — the source NoKnownRelationElements is
	// built from.
	HadAnyHop map[ids.EntityID]bool
}

// walkPropagation is this feature's reverse-relation walk: for each
// Requirement-level ChangedElement, every Task in the same Spec whose
// Serves: covers that Requirement number, via prepare.ScanSpecTasks
// (which already parses coverage — research.md #5's original plan to
// call validation.ParseTaskCoverage directly is refined here to reuse
// ScanSpecTasks instead, since it already wraps that exact parse
// together with each Task's own 041 EvidenceState, needed by
// Decision 10's reason text — Principle VI, DRY). A Task is a leaf: it
// is never further expanded (operations.Backlinks only scans the five
// standalone-referenceable types).
//
// For each whole-artifact ChangedElement, every formal (depends_on/
// parent/supersedes) and semantic (wikilink) backlink via
// operations.Backlinks, walked to a full transitive closure: an
// intermediate artifact reached by one hop is itself expanded in
// turn, so a multi-hop dependency chain is reported in full
// (research.md #6, spec FR-005). Termination is guaranteed by an
// `expanded` set: each artifact's own onward backlinks are computed at
// most once per run, regardless of how many distinct incoming paths
// reach it — a second (or further) path arriving at an
// already-expanded artifact is still recorded (multiple distinct
// PropagationPaths to the same item are always kept, spec FR-005),
// only its own further expansion is skipped, which is what actually
// bounds the walk on a cycle (spec FR-006). The originally changed
// element(s) themselves are never recorded as an affected item, even
// if a cycle loops back to one — they are the cause, not something
// this run found affected.
func walkPropagation(root string, cfg project.Configuration, elements []ChangedElement) (propagationResult, error) {
	result := propagationResult{
		PathsByItem: map[string][]PropagationPath{},
		HadAnyHop:   map[ids.EntityID]bool{},
	}

	changedIDs := map[string]bool{}
	for _, elem := range elements {
		if elem.RequirementNumber == nil {
			changedIDs[elem.ID.String()] = true
		}
	}

	type queueItem struct {
		id   ids.EntityID
		path []PropagationHop
	}
	var queue []queueItem

	for _, elem := range elements {
		if _, seen := result.HadAnyHop[elem.ID]; !seen {
			result.HadAnyHop[elem.ID] = false
		}

		if elem.RequirementNumber == nil {
			queue = append(queue, queueItem{id: elem.ID})
			continue
		}

		hops, err := coverageHops(root, cfg, elem)
		if err != nil {
			return propagationResult{}, err
		}
		for _, hop := range hops {
			result.HadAnyHop[elem.ID] = true
			addPropagationPath(result.PathsByItem, PropagationPath{Hops: []PropagationHop{hop}})
		}
	}

	expanded := map[ids.EntityID]bool{}
	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]
		if expanded[item.id] {
			continue
		}
		expanded[item.id] = true

		hops, err := backlinkHopsForID(root, cfg, item.id)
		if err != nil {
			return propagationResult{}, err
		}
		if len(item.path) == 0 && len(hops) > 0 {
			result.HadAnyHop[item.id] = true
		}

		for _, hop := range hops {
			if changedIDs[hop.ToID] {
				// Never report an originally changed element as its
				// own affected item, even if a cycle loops back to it.
				continue
			}

			newPath := make([]PropagationHop, 0, len(item.path)+1)
			newPath = append(newPath, item.path...)
			newPath = append(newPath, hop)
			addPropagationPath(result.PathsByItem, PropagationPath{Hops: newPath})

			if strings.Contains(hop.ToID, ":") {
				// A Task's composite ID — a leaf, never expanded further.
				continue
			}
			if toEntity, err := ids.ParseAny(hop.ToID); err == nil {
				queue = append(queue, queueItem{id: toEntity, path: newPath})
			}
		}
	}

	return result, nil
}

// backlinkHops returns one PropagationHop per formal (depends_on/
// parent/supersedes) or semantic (wikilink) backlink into elem.ID —
// operations.Backlinks already computes both lists together in one
// full-project scan, so consuming Formal and Semantic here in the same
// call (rather than a separate scan per relation kind) avoids a second
// scan of the same data (Principle VI, DRY). classify.go's fixed table
// (unchanged by this function) is what actually keeps "wikilink" from
// ever being treated as deterministic invalidation — this function
// only supplies the relation kind, never the classification itself
// (research.md #7; 042-impact-analysis-review User Story 2).
func backlinkHopsForID(root string, cfg project.Configuration, fromID ids.EntityID) ([]PropagationHop, error) {
	back, err := operations.Backlinks(root, cfg, fromID.String())
	if err != nil {
		return nil, err
	}

	hops := make([]PropagationHop, 0, len(back.Formal)+len(back.Semantic))
	for _, f := range back.Formal {
		hops = append(hops, PropagationHop{
			Relation:   RelationKind(f.Relation),
			FromID:     fromID.String(),
			ToID:       f.Source.String(),
			SourcePath: f.SourcePath,
		})
	}
	for _, s := range back.Semantic {
		hops = append(hops, PropagationHop{
			Relation:      RelationWikilink,
			FromID:        fromID.String(),
			ToID:          s.Source.String(),
			SourcePath:    s.SourcePath,
			SourceSection: s.SourceSection,
			SourceLine:    s.SourceLine,
		})
	}
	return hops, nil
}

// coverageHops returns one PropagationHop per Task, in elem's own
// owning Spec, whose Serves: covers elem.RequirementNumber
// (research.md #5). When that Task's own 041 EvidenceState is already
// broken — Stale or Failed, meaning it once had recorded evidence that
// is no longer trustworthy, for reasons unrelated to this change (read
// via prepare.ScanSpecTasks, never recomputed here per research.md
// #10) — the hop's Relation is RelationEvidence rather than
// RelationCoverage, so classify.go's existing severityTable (Evidence
// -> SeverityHigh) applies without any further special-casing. A Task
// that is simply Unverified (never yet attempted — the ordinary state
// of any pending Task) is deliberately NOT promoted to RelationEvidence:
// treating every not-yet-done Task as "evidence already broken" would
// make the common case noisy and indistinguishable from a genuine
// regression (correction found during code review — the original
// implementation used `!= Verified`, which also matched Unverified).
// TaskEvidenceState is always populated on the returned hop either
// way, so reasonFor can report it precisely.
func coverageHops(root string, cfg project.Configuration, elem ChangedElement) ([]PropagationHop, error) {
	specDir := filepath.Dir(elem.Path)
	tasks, found, err := prepare.ScanSpecTasks(root, cfg, elem.ID, specDir)
	if err != nil || !found {
		return nil, err
	}

	tasksPath := specDir + "/tasks.md"

	var hops []PropagationHop
	for _, ti := range tasks {
		for _, ref := range ti.Coverage.References {
			if ref.Spec != elem.ID || ref.Number != *elem.RequirementNumber {
				continue
			}
			taskID := ids.TaskID{Spec: elem.ID, Local: ti.Task}
			number := *elem.RequirementNumber

			relation := RelationCoverage
			if ti.Evidence == evidence.Stale || ti.Evidence == evidence.Failed {
				relation = RelationEvidence
			}

			hops = append(hops, PropagationHop{
				Relation:          relation,
				FromID:            elem.ID.String(),
				ToID:              taskID.String(),
				SourcePath:        tasksPath,
				RequirementNumber: &number,
				TaskEvidenceState: ti.Evidence,
			})
		}
	}
	return hops, nil
}

// addPropagationPath records path against its own last hop's ToID.
func addPropagationPath(pathsByItem map[string][]PropagationPath, path PropagationPath) {
	toID := path.Hops[len(path.Hops)-1].ToID
	pathsByItem[toID] = append(pathsByItem[toID], path)
}
