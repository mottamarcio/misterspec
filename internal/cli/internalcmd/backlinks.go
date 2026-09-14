package internalcmd

import (
	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// NewBacklinksCmd builds "misterspec internal backlinks <id>"
// (docs/context-engine-implementation.md §7.2,
// specs/012-references-backlinks/contracts/references-backlinks.md). It
// calls operations.Backlinks directly — no logic of its own beyond
// argument parsing and JSON shaping, matching NewReferencesCmd's own
// shape.
func NewBacklinksCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:           "backlinks <id>",
		Short:         "Return every artifact that formally or semantically references the given artifact",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			result, err := operations.Backlinks(proj.Root, proj.Config, args[0])
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"target": args[0],
				"backlinks": map[string]any{
					"formal":   renderBacklinkEntries(result.Formal),
					"semantic": renderBacklinkEntries(result.Semantic),
				},
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	return cmd
}

// renderBacklinkEntries renders a []operations.BacklinkEntry as JSON-
// ready maps, always a non-nil slice (empty, never null, when entries is
// empty) — matching renderReferenceEntries' own convention.
func renderBacklinkEntries(entries []operations.BacklinkEntry) []map[string]any {
	out := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		out = append(out, map[string]any{
			"relation": e.Relation,
			"source":   e.Source.String(),
		})
	}
	return out
}
