package internalcmd

import (
	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// NewReferencesCmd builds "misterspec internal references <id>"
// (docs/context-engine-implementation.md §7.1,
// specs/012-references-backlinks/contracts/references-backlinks.md). It
// calls operations.References directly — no logic of its own beyond
// argument parsing and JSON shaping (spec.md FR-009-style contract,
// matching every existing internal command).
func NewReferencesCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:           "references <id>",
		Short:         "Return an artifact's outgoing formal and semantic relationships",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			result, err := operations.References(proj.Root, proj.Config, args[0])
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"target": args[0],
				"references": map[string]any{
					"formal":   renderReferenceEntries(result.Formal),
					"semantic": renderReferenceEntries(result.Semantic),
				},
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	return cmd
}

// renderReferenceEntries renders a []operations.ReferenceEntry as JSON-
// ready maps, always a non-nil slice (empty, never null, when entries is
// empty) so the JSON array is never null — matching inspect.go's own
// idStrings convention. Each entry additionally carries its own
// source_path/source_section/source_line (038-wikilink-chunk-
// provenance contracts §3) — always present, empty/zero for a formal
// relation.
func renderReferenceEntries(entries []operations.ReferenceEntry) []map[string]any {
	out := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		out = append(out, map[string]any{
			"relation":       e.Relation,
			"target":         e.Target.String(),
			"source_path":    e.SourcePath,
			"source_section": e.SourceSection,
			"source_line":    e.SourceLine,
		})
	}
	return out
}
