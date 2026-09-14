package internalcmd

import (
	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// NewInspectCmd builds "misterspec internal inspect <id>"
// (docs/architecture-specification.md §9.3). It calls operations.Inspect
// directly — no logic of its own beyond argument parsing and JSON
// shaping (spec.md FR-009).
func NewInspectCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:           "inspect <id>",
		Short:         "Return deterministic metadata for an artifact",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			result, err := operations.Inspect(proj.Root, proj.Config, args[0])
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			var parent any
			if result.Metadata.Parent != nil {
				parent = result.Metadata.Parent.String()
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"entity": map[string]any{
					"id":         result.Location.ID.String(),
					"type":       result.Location.Type.String(),
					"status":     result.Metadata.Status,
					"parent":     parent,
					"depends_on": idStrings(result.Metadata.DependsOn),
					"supersedes": idStrings(result.Metadata.Supersedes),
				},
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	return cmd
}

// idStrings renders ids as their String() form, always a non-nil slice
// (empty, never null, when ids is empty) so the JSON array is never
// null.
func idStrings(list []ids.EntityID) []string {
	out := make([]string, 0, len(list))
	for _, id := range list {
		out = append(out, id.String())
	}
	return out
}
