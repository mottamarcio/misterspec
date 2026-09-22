package internalcmd

import (
	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/impact"
	"github.com/mottamarcio/misterspec/internal/project"
)

// analyzeImpactSchemaVersion is "analyze-impact"'s own response
// contract version (042-impact-analysis-review contracts/
// analyze-impact-contract.md §4).
const analyzeImpactSchemaVersion = 1

// NewAnalyzeImpactCmd builds "misterspec internal analyze-impact"
// (042-impact-analysis-review contracts/analyze-impact-contract.md
// §4). Diffs --from/--to (default: the working tree), optionally
// scoped to --path, and reports every artifact or Task with a known
// relation to what changed — read-only, never mutates the filesystem.
func NewAnalyzeImpactCmd() *cobra.Command {
	var dir, from, to, path string

	cmd := &cobra.Command{
		Use:           "analyze-impact",
		Short:         "Report every artifact/Task with a known relation to a change between two revisions",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if from == "" {
				return WriteError(cmd.OutOrStdout(), ErrInvalidArgument)
			}

			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			report, err := impact.AnalyzeImpact(proj.Root, proj.Config, impact.AnalyzeImpactRequest{
				From: from,
				To:   to,
				Path: path,
			})
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"analyze_impact_schema_version": analyzeImpactSchemaVersion,
				"change_set":                    renderChangeSet(report.ChangeSet),
				"affected_items":                renderAffectedItems(report.AffectedItems),
				"no_known_relation_elements":    renderEntityIDs(report.NoKnownRelationElements),
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	cmd.Flags().StringVar(&from, "from", "", "revision to compare from (required)")
	cmd.Flags().StringVar(&to, "to", "", "revision to compare to (default: working tree)")
	cmd.Flags().StringVar(&path, "path", "", "scope the diff to one artifact path (default: whole project)")
	return cmd
}

func renderChangeSet(cs impact.ChangeSet) map[string]any {
	elements := make([]map[string]any, 0, len(cs.Elements))
	for _, e := range cs.Elements {
		var reqNumber any
		if e.RequirementNumber != nil {
			reqNumber = *e.RequirementNumber
		}
		elements = append(elements, map[string]any{
			"id":                 e.ID.String(),
			"path":               e.Path,
			"status":             string(e.Status),
			"requirement_number": reqNumber,
		})
	}
	return map[string]any{
		"from":                cs.From,
		"to":                  cs.To,
		"elements":            elements,
		"unmapped_code_paths": cs.UnmappedCodePaths,
	}
}

func renderAffectedItems(items []impact.AffectedItem) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		var reverif any
		if item.ReverificationCandidate != nil {
			reverif = *item.ReverificationCandidate
		}
		out = append(out, map[string]any{
			"id":                       item.ID,
			"path":                     item.Path,
			"classification":           string(item.Classification),
			"severity":                 string(item.Severity),
			"reason":                   item.Reason,
			"paths":                    renderPropagationPaths(item.Paths),
			"reverification_candidate": reverif,
		})
	}
	return out
}

func renderPropagationPaths(paths []impact.PropagationPath) []map[string]any {
	out := make([]map[string]any, 0, len(paths))
	for _, p := range paths {
		hops := make([]map[string]any, 0, len(p.Hops))
		for _, h := range p.Hops {
			hops = append(hops, map[string]any{
				"relation":       string(h.Relation),
				"from":           h.FromID,
				"to":             h.ToID,
				"source_path":    h.SourcePath,
				"source_section": h.SourceSection,
				"source_line":    h.SourceLine,
			})
		}
		out = append(out, map[string]any{"hops": hops})
	}
	return out
}

func renderEntityIDs(entityIDs []ids.EntityID) []string {
	out := make([]string, 0, len(entityIDs))
	for _, id := range entityIDs {
		out = append(out, id.String())
	}
	return out
}
