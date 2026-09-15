package internalcmd

import (
	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// NewParentCmd builds "misterspec internal parent <id>"
// (docs/architecture-specification.md §9.4). It calls operations.Parent
// directly — no logic of its own beyond argument parsing and JSON
// shaping (spec.md FR-009).
func NewParentCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:           "parent <id>",
		Short:         "Return the structural parent",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			result, err := operations.Parent(proj.Root, proj.Config, args[0])
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			var parent any
			if result.HasParent {
				parent = map[string]any{
					"id":   result.Parent.ID.String(),
					"type": result.Parent.Type.String(),
				}
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{"parent": parent})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	return cmd
}
