package impact

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

// AnalyzeImpactRequest names the two revisions to compare and an
// optional path scope (contracts §3).
type AnalyzeImpactRequest struct {
	From string
	To   string
	Path string
}

// AffectedItem is one artifact or Task found reachable from the Change
// Set, with its full explanation (data-model.md "AffectedItem").
type AffectedItem struct {
	ID                      string
	Path                    string
	Classification          Classification
	Severity                Severity
	Reason                  string
	Paths                   []PropagationPath
	ReverificationCandidate *string
}

// ImpactReport is the full response of one AnalyzeImpact call
// (data-model.md "ImpactReport").
type ImpactReport struct {
	ChangeSet               ChangeSet
	AffectedItems           []AffectedItem
	NoKnownRelationElements []ids.EntityID
}

// AnalyzeImpact computes the ChangeSet between req.From/req.To (scoped
// to req.Path when set), walks every changed element's reverse
// relations, and returns the full ImpactReport (contracts §3). Never
// mutates the filesystem.
func AnalyzeImpact(root string, cfg project.Configuration, req AnalyzeImpactRequest) (ImpactReport, error) {
	cs, err := BuildChangeSet(root, cfg, req.From, req.To, req.Path)
	if err != nil {
		return ImpactReport{}, err
	}

	prop, err := walkPropagation(root, cfg, cs.Elements)
	if err != nil {
		return ImpactReport{}, err
	}

	items := buildAffectedItems(cs, prop)

	var noKnown []ids.EntityID
	for id, had := range prop.HadAnyHop {
		if !had {
			noKnown = append(noKnown, id)
		}
	}
	sort.Slice(noKnown, func(i, j int) bool { return noKnown[i].String() < noKnown[j].String() })

	return ImpactReport{ChangeSet: cs, AffectedItems: items, NoKnownRelationElements: noKnown}, nil
}

// buildAffectedItems renders prop's PathsByItem into ordered
// AffectedItems, each classified/severity-scored from its own paths'
// last hop (data-model.md "AffectedItem").
func buildAffectedItems(cs ChangeSet, prop propagationResult) []AffectedItem {
	itemIDs := make([]string, 0, len(prop.PathsByItem))
	for id := range prop.PathsByItem {
		itemIDs = append(itemIDs, id)
	}
	sort.Strings(itemIDs)

	items := make([]AffectedItem, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		paths := prop.PathsByItem[itemID]
		lastHop := paths[0].Hops[len(paths[0].Hops)-1]
		classification := ClassifyRelation(lastHop.Relation)
		severity := SeverityFor(classification, lastHop.Relation)

		item := AffectedItem{
			ID:             itemID,
			Path:           itemPath(itemID, lastHop),
			Classification: classification,
			Severity:       severity,
			Reason:         reasonFor(itemID, lastHop, cs),
			Paths:          paths,
		}
		if classification == DeterministicInvalidation && strings.Contains(itemID, ":TASK-") {
			candidate := itemID
			item.ReverificationCandidate = &candidate
		}
		items = append(items, item)
	}

	sort.SliceStable(items, func(i, j int) bool {
		if severityRank(items[i].Severity) != severityRank(items[j].Severity) {
			return severityRank(items[i].Severity) < severityRank(items[j].Severity)
		}
		return items[i].Path < items[j].Path
	})
	return items
}

// severityRank orders Severity high-to-low for buildAffectedItems'
// own sort (data-model.md "ImpactReport": "Severity (high -> low),
// then Path ascending").
func severityRank(s Severity) int {
	switch s {
	case SeverityHigh:
		return 0
	case SeverityMedium:
		return 1
	default:
		return 2
	}
}

// itemPath renders an AffectedItem's own canonical path: the hop's
// SourcePath directly for an artifact, or "<tasks.md path>#<local Task
// ID>" for a Task — the same "path#TASK-NNN" convention
// validation.Finding.Path already uses.
func itemPath(itemID string, hop PropagationHop) string {
	if _, local, found := strings.Cut(itemID, ":"); found {
		return hop.SourcePath + "#" + local
	}
	return hop.SourcePath
}

// reasonFor renders a specific, human-readable explanation for why
// itemID was included, per hop's own Relation (spec FR-005). For
// RelationEvidence, the base coverage explanation is extended with the
// Task's own already-non-Verified 041 EvidenceState (research.md #10)
// — a distinct reason from "the Requirement it serves changed," kept
// legible rather than blended into one recomputed state (code review
// finding: this clause was previously never reached at all, since
// coverageHops never inspected the Task's own evidence).
func reasonFor(itemID string, hop PropagationHop, cs ChangeSet) string {
	switch hop.Relation {
	case RelationCoverage, RelationEvidence:
		requirement := 0
		if hop.RequirementNumber != nil {
			requirement = *hop.RequirementNumber
		}
		_, local, _ := strings.Cut(itemID, ":")
		reason := fmt.Sprintf("%s serves %s:R%d, whose content changed between %s and %s", local, hop.FromID, requirement, cs.From, cs.To)
		if hop.Relation == RelationEvidence {
			reason += fmt.Sprintf("; its own recorded evidence is already %s against its current content", hop.TaskEvidenceState)
		}
		return reason
	case RelationDependsOn, RelationParent, RelationSupersedes:
		return fmt.Sprintf("%s %s %s, which changed between %s and %s", itemID, hop.Relation, hop.FromID, cs.From, cs.To)
	case RelationWikilink:
		return fmt.Sprintf("%s references %s (from %s), which changed between %s and %s", itemID, hop.FromID, hop.SourcePath, cs.From, cs.To)
	default:
		return fmt.Sprintf("%s is related to %s (%s), which changed between %s and %s", itemID, hop.FromID, hop.Relation, cs.From, cs.To)
	}
}
